package control

import (
	"sort"

	"os"

	"github.com/dhruvhegde/cudackpt/internal/ckpterr"
	"github.com/dhruvhegde/cudackpt/pkg/image"
	jlog "github.com/dhruvhegde/cudackpt/pkg/log"
	"github.com/dhruvhegde/cudackpt/pkg/rpc"
)

func (o *Orchestrator) dial(pid int) (*rpc.Client, error) {
	secret := o.cfg.RpcSecret
	if secret == "" {
		secret = os.Getenv("CUDACKPT_RPC_SECRET")
	}
	cli, err := rpc.DialPathWithSecret(shimSocketPath(o.cfg.RunDir, pid), secret)
	if err != nil {
		return nil, ckpterr.Wrap(ckpterr.RPC, "dial", err)
	}
	return cli, nil
}

func (o *Orchestrator) Ping(pid int) error {
	cli, err := o.dial(pid)
	if err != nil {
		return err
	}
	defer func() { _ = cli.Close() }()
	if err := cli.Ping(); err != nil {
		return ckpterr.Wrap(ckpterr.RPC, "ping", err)
	}
	return nil
}

func (o *Orchestrator) Status(pid int) (uint32, error) {
	cli, err := o.dial(pid)
	if err != nil {
		return 0, err
	}
	defer func() { _ = cli.Close() }()
	st, err := cli.Status()
	if err != nil {
		return 0, ckpterr.Wrap(ckpterr.RPC, "status", err)
	}
	return st, nil
}

func (o *Orchestrator) Stats(pid int) (rpc.Stats, error) {
	cli, err := o.dial(pid)
	if err != nil {
		return rpc.Stats{}, err
	}
	defer func() { _ = cli.Close() }()
	st, err := cli.Stats()
	if err != nil {
		return rpc.Stats{}, ckpterr.Wrap(ckpterr.RPC, "stats", err)
	}
	return st, nil
}

func (o *Orchestrator) Freeze(pid int) error {
	cli, err := o.dial(pid)
	if err != nil {
		return err
	}
	defer func() { _ = cli.Close() }()
	if err := cli.Freeze(); err != nil {
		return ckpterr.Wrap(ckpterr.RPC, "freeze", err)
	}
	return nil
}

func (o *Orchestrator) Snapshot(pid int, dir string) error {
	cli, err := o.dial(pid)
	if err != nil {
		return err
	}
	defer func() { _ = cli.Close() }()
	if err := cli.Snapshot(dir); err != nil {
		return ckpterr.Wrap(ckpterr.CUDA, "snapshot", err)
	}
	if err := o.verifyImage(dir); err != nil {
		return err
	}
	jlog.Info("snapshot_ok", map[string]any{"pid": pid, "dir": dir})
	return nil
}

func (o *Orchestrator) GpuRestore(pid int, dir string) error {
	if err := image.EnsureDeviceMaterialized(dir); err != nil {
		return ckpterr.Wrap(ckpterr.IO, "materialize", err)
	}
	cli, err := o.dial(pid)
	if err != nil {
		return err
	}
	defer func() { _ = cli.Close() }()
	if err := cli.Restore(dir); err != nil {
		return ckpterr.Wrap(ckpterr.CUDA, "gpu restore", err)
	}
	return nil
}

func (o *Orchestrator) Resume(pid int) error {
	cli, err := o.dial(pid)
	if err != nil {
		return err
	}
	defer func() { _ = cli.Close() }()
	if err := cli.Resume(); err != nil {
		return ckpterr.Wrap(ckpterr.RPC, "resume", err)
	}
	return nil
}

func (o *Orchestrator) tryShimRestore(imagePath string, pid int) (int, error) {
	cli, err := o.dial(pid)
	if err != nil {
		return 0, err
	}
	defer func() { _ = cli.Close() }()
	if err := cli.Ping(); err != nil {
		return 0, err
	}
	if err := cli.Restore(imagePath); err != nil {
		return 0, ckpterr.Wrap(ckpterr.CUDA, "gpu restore", err)
	}
	logRestorePhase(imagePath, "gpu", pid, nil)
	if err := cli.Resume(); err != nil {
		return 0, ckpterr.Wrap(ckpterr.RPC, "resume", err)
	}
	logRestorePhase(imagePath, "resume", pid, nil)
	return pid, nil
}

func sortedShims(runDir string) ([]int, error) {
	pids, err := ListShims(runDir)
	if err != nil {
		return nil, err
	}
	sort.Ints(pids)
	return pids, nil
}

func (o *Orchestrator) sortedShims() ([]int, error) {
	return sortedShims(o.cfg.RunDir)
}
