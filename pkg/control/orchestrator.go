package control

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dhruvhegde/cudackpt/internal/ckpterr"
	"github.com/dhruvhegde/cudackpt/pkg/config"
	"github.com/dhruvhegde/cudackpt/pkg/image"
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
	if err := os.MkdirAll(out, 0o755); err != nil {
		return ckpterr.Wrap(ckpterr.IO, "mkdir", err)
	}
	if err := o.Freeze(pid); err != nil {
		return err
	}
	if err := o.Snapshot(pid, out); err != nil {
		return err
	}
	if err := o.criu.Dump(pid, out); err != nil {
		return ckpterr.Wrap(ckpterr.CRIU, "dump", err)
	}
	dev, _ := image.ReadDev(filepath.Join(out, "dev.bin"))
	meta := image.Meta{
		Pid:     uint32(pid),
		Dev:     dev,
		Preload: os.Getenv("LD_PRELOAD"),
		Visible: os.Getenv("CUDA_VISIBLE_DEVICES"),
	}
	if err := image.WriteMeta(filepath.Join(out, "meta.bin"), meta); err != nil {
		return ckpterr.Wrap(ckpterr.IO, "meta", err)
	}
	return nil
}

func (o *Orchestrator) verifyImage(dir string) error {
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
		return 0, ckpterr.Wrap(ckpterr.CRIU, "restore", err)
	}
	if pid == 0 {
		pids, lerr := ListShims(o.cfg.RunDir)
		if lerr == nil && len(pids) > 0 {
			pid = pids[0]
		}
	}
	deadline := time.Now().Add(o.cfg.RestoreTimeout)
	for time.Now().Before(deadline) {
		var candidates []int
		if pid > 0 {
			candidates = append(candidates, pid)
		}
		if pids, lerr := o.sortedShims(); lerr == nil {
			for _, p := range pids {
				if p != pid {
					candidates = append(candidates, p)
				}
			}
		}
		for _, try := range candidates {
			if got, rerr := o.tryShimRestore(imagePath, try); rerr == nil {
				return got, nil
			}
		}
		time.Sleep(o.cfg.ShimPoll)
	}
	return 0, ckpterr.E(ckpterr.RPC, "shim not ready after criu restore")
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
