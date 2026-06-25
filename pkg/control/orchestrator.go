package control

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dhruvhegde/cudackpt/internal/ckpterr"
	"github.com/dhruvhegde/cudackpt/pkg/config"
	"github.com/dhruvhegde/cudackpt/pkg/image"
	jlog "github.com/dhruvhegde/cudackpt/pkg/log"
	"github.com/dhruvhegde/cudackpt/pkg/storage"
	"github.com/dhruvhegde/cudackpt/third_party/criu"
)

type Orchestrator struct {
	criu *criu.Wrapper
	cfg  config.Config
}

func New(cfg config.Config) *Orchestrator {
	return &Orchestrator{criu: criu.New(), cfg: cfg}
}

func imageDir(base string, pid int) string {
	return filepath.Join(base, fmt.Sprintf("ckpt-%d", pid))
}

func (o *Orchestrator) Checkpoint(pid int, out string) error {
	if out == "" {
		out = imageDir(o.cfg.ImageRoot, pid)
	}
	jlog.Info("checkpoint_start", map[string]any{"pid": pid, "dir": out})
	opts := image.OptsFromEnv()
	err := image.WriteStaging(out, func(staging string) error {
		if err := o.Freeze(pid); err != nil {
			return err
		}
		if err := o.Snapshot(pid, staging); err != nil {
			return err
		}
		if err := image.ProcessImage(staging, opts); err != nil {
			return ckpterr.Wrap(ckpterr.IO, "process", err)
		}
		if err := o.criu.Dump(pid, staging); err != nil {
			return ckpterr.Wrap(ckpterr.CRIU, "dump", err)
		}
		dev, _ := image.ReadDev(filepath.Join(staging, "dev.bin"))
		meta := image.Meta{
			Pid:     uint32(pid),
			Dev:     dev,
			Preload: os.Getenv("LD_PRELOAD"),
			Visible: os.Getenv("CUDA_VISIBLE_DEVICES"),
		}
		if err := image.WriteMeta(filepath.Join(staging, "meta.bin"), meta); err != nil {
			return ckpterr.Wrap(ckpterr.IO, "meta", err)
		}
		return nil
	})
	if err != nil {
		jlog.Error("checkpoint_fail", err, map[string]any{"pid": pid})
		return err
	}
	jlog.Info("checkpoint_ok", map[string]any{"pid": pid, "dir": out})
	return nil
}

func (o *Orchestrator) verifyImage(dir string) error {
	if err := image.EnsureDeviceMaterialized(dir); err != nil {
		return ckpterr.Wrap(ckpterr.IO, "materialize", err)
	}
	tier, err := storage.New(dir)
	if err != nil {
		return ckpterr.Wrap(ckpterr.IO, "tier", err)
	}
	entries, hdr, err := image.ReadManifest(tier.ManifestPath())
	if err != nil {
		return ckpterr.Wrap(ckpterr.IO, "manifest", err)
	}
	if hdr.Magic != image.Magic {
		return ckpterr.E(ckpterr.Invalid, "bad magic")
	}
	f, err := os.Open(tier.DevicePath())
	if err != nil {
		return ckpterr.Wrap(ckpterr.IO, "device.bin", err)
	}
	defer f.Close()
	for _, e := range entries {
		ok, verr := image.VerifyChunk(f, int64(e.Offset), int64(e.Size), e.CRC32C)
		if verr != nil {
			return ckpterr.Wrap(ckpterr.IO, "verify", verr)
		}
		if !ok {
			return ckpterr.E(ckpterr.Invalid, "crc mismatch")
		}
	}
	return nil
}

func (o *Orchestrator) Restore(imagePath string) (int, error) {
	jlog.Info("restore_start", map[string]any{"dir": imagePath})
	if err := PreflightRestore(imagePath); err != nil {
		jlog.Error("restore_preflight", err, map[string]any{"dir": imagePath})
		return 0, err
	}
	if err := image.EnsureDeviceMaterialized(imagePath); err != nil {
		jlog.Error("restore_materialize", err, map[string]any{"dir": imagePath})
		return 0, ckpterr.Wrap(ckpterr.IO, "materialize", err)
	}
	logPath := filepath.Join(imagePath, "restore.log")
	var env []string
	if m, err := image.ReadMeta(filepath.Join(imagePath, "meta.bin")); err == nil {
		if m.Preload != "" {
			env = append(env, "LD_PRELOAD="+m.Preload)
		}
		if m.Visible != "" {
			env = append(env, "CUDA_VISIBLE_DEVICES="+m.Visible)
		}
	}
	pid, err := o.criu.Restore(imagePath, logPath, env)
	if err != nil {
		jlog.Error("restore_criu", err, map[string]any{"dir": imagePath})
		return 0, ckpterr.Wrap(ckpterr.CRIU, "restore", err)
	}
	if pid > 0 {
		_ = writeRestoreHandoff(imagePath, pid)
	} else if filePID := readRestoredPID(imagePath); filePID > 0 {
		pid = filePID
	}
	jlog.Info("restore_criu_ok", map[string]any{"dir": imagePath, "pid": pid})
	deadline := time.Now().Add(o.cfg.RestoreTimeout)
	attempt := 0
	for time.Now().Before(deadline) {
		attempt++
		pids, lerr := o.sortedShims()
		if lerr != nil {
			time.Sleep(restorePollDelay(attempt, o.cfg.RetryBackoff))
			continue
		}
		target := resolveRestorePID(imagePath, pid, pids)
		if target > 0 && !shimSocketReady(o.cfg.RunDir, target) {
			jlog.Info("restore_wait_shim", map[string]any{"pid": target, "attempt": attempt})
			time.Sleep(restorePollDelay(attempt, o.cfg.RetryBackoff))
			continue
		}
		for _, try := range restoreCandidates(target, pids) {
			if !shimSocketReady(o.cfg.RunDir, try) {
				continue
			}
			if got, rerr := o.tryShimRestore(imagePath, try); rerr == nil {
				_ = writeRestoreHandoff(imagePath, got)
				jlog.Info("restore_ok", map[string]any{"pid": got, "dir": imagePath, "attempt": attempt})
				return got, nil
			}
		}
		time.Sleep(restorePollDelay(attempt, o.cfg.RetryBackoff))
	}
	err = ckpterr.E(ckpterr.RPC, "shim not ready after criu restore")
	jlog.Error("restore_fail", err, map[string]any{"dir": imagePath})
	return 0, err
}

func ListShims(runDir string) ([]int, error) {
	if runDir == "" {
		runDir = "/run/cudackpt"
	}
	ents, err := os.ReadDir(runDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var pids []int
	for _, e := range ents {
		var pid int
		if _, err := fmt.Sscanf(e.Name(), "%d.sock", &pid); err == nil {
			pids = append(pids, pid)
		}
	}
	return pids, nil
}

func ListImages(root string) ([]string, error) {
	ents, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []string
	for _, e := range ents {
		if e.IsDir() {
			out = append(out, filepath.Join(root, e.Name()))
		}
	}
	return out, nil
}
