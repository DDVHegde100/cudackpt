package control

import (
	"sort"

	"github.com/dhruvhegde/cudackpt/internal/ckpterr"
	"github.com/dhruvhegde/cudackpt/pkg/rpc"
)

func dial(pid int) (*rpc.Client, error) {
	cli, err := rpc.Dial(pid)
	if err != nil {
		return nil, ckpterr.Wrap(ckpterr.RPC, "dial", err)
	}
	return cli, nil
}

func (o *Orchestrator) Ping(pid int) error {
	cli, err := dial(pid)
	if err != nil {
		return err
	}
	defer cli.Close()
	if err := cli.Ping(); err != nil {
		return ckpterr.Wrap(ckpterr.RPC, "ping", err)
	}
	return nil
}

func (o *Orchestrator) Status(pid int) (uint32, error) {
	cli, err := dial(pid)
	if err != nil {
		return 0, err
	}
	defer cli.Close()
	st, err := cli.Status()
	if err != nil {
		return 0, ckpterr.Wrap(ckpterr.RPC, "status", err)
	}
	return st, nil
}

func (o *Orchestrator) Freeze(pid int) error {
	cli, err := dial(pid)
	if err != nil {
		return err
	}
	defer cli.Close()
	if err := cli.Freeze(); err != nil {
		return ckpterr.Wrap(ckpterr.RPC, "freeze", err)
	}
	return nil
}

func (o *Orchestrator) Snapshot(pid int, dir string) error {
	cli, err := dial(pid)
	if err != nil {
		return err
	}
	defer cli.Close()
	if err := cli.Snapshot(dir); err != nil {
		return ckpterr.Wrap(ckpterr.CUDA, "snapshot", err)
	}
	return o.verifyImage(dir)
}

func (o *Orchestrator) GpuRestore(pid int, dir string) error {
	cli, err := dial(pid)
	if err != nil {
		return err
	}
	defer cli.Close()
	if err := cli.Restore(dir); err != nil {
		return ckpterr.Wrap(ckpterr.CUDA, "gpu restore", err)
	}
	return nil
}

func (o *Orchestrator) Resume(pid int) error {
	cli, err := dial(pid)
	if err != nil {
		return err
	}
	defer cli.Close()
	if err := cli.Resume(); err != nil {
		return ckpterr.Wrap(ckpterr.RPC, "resume", err)
	}
	return nil
}

func (o *Orchestrator) tryShimRestore(imagePath string, pid int) (int, error) {
	cli, err := rpc.Dial(pid)
	if err != nil {
		return 0, err
	}
	defer cli.Close()
	if err := cli.Ping(); err != nil {
		return 0, err
	}
	if err := cli.Restore(imagePath); err != nil {
		return 0, ckpterr.Wrap(ckpterr.CUDA, "gpu restore", err)
	}
	if err := cli.Resume(); err != nil {
		return 0, ckpterr.Wrap(ckpterr.RPC, "resume", err)
	}
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
