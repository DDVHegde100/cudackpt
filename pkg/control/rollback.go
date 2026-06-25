package control

import (
	"os"
	"syscall"
	"time"

	"github.com/dhruvhegde/cudackpt/internal/ckpterr"
	jlog "github.com/dhruvhegde/cudackpt/pkg/log"
)

func stopProcess(pid int) error {
	if pid <= 0 {
		return nil
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return ckpterr.Wrap(ckpterr.Invalid, "find process", err)
	}
	_ = proc.Signal(syscall.SIGTERM)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if syscall.Kill(pid, 0) != nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err := proc.Signal(syscall.SIGKILL); err != nil {
		return ckpterr.Wrap(ckpterr.Invalid, "kill process", err)
	}
	return nil
}

func (o *Orchestrator) Rollback(imagePath string, stopPID int) (int, error) {
	jlog.Info("rollback_start", map[string]any{"dir": imagePath, "stop_pid": stopPID})
	if stopPID > 0 {
		if err := stopProcess(stopPID); err != nil {
			jlog.Error("rollback_stop", err, map[string]any{"pid": stopPID})
			return 0, err
		}
	}
	if err := o.ValidateImage(imagePath); err != nil {
		jlog.Error("rollback_validate", err, map[string]any{"dir": imagePath})
		return 0, err
	}
	pid, err := o.Restore(imagePath)
	if err != nil {
		jlog.Error("rollback_restore", err, map[string]any{"dir": imagePath})
		return 0, err
	}
	jlog.Info("rollback_ok", map[string]any{"dir": imagePath, "pid": pid})
	return pid, nil
}
