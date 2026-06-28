package control

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	jlog "github.com/dhruvhegde/cudackpt/pkg/log"
)

const restoreEventsName = "restore.events.jsonl"

type restorePhaseEvent struct {
	Time  string         `json:"time"`
	Phase string         `json:"phase"`
	Dir   string         `json:"dir"`
	PID   int            `json:"pid,omitempty"`
	Extra map[string]any `json:"extra,omitempty"`
}

func logRestorePhase(dir, phase string, pid int, extra map[string]any) {
	fields := map[string]any{
		"phase": phase,
		"dir":   dir,
	}
	if pid > 0 {
		fields["pid"] = pid
	}
	for k, v := range extra {
		fields[k] = v
	}
	jlog.Info("restore_phase", fields)
	_ = appendRestorePhaseEvent(dir, phase, pid, extra)
}

func appendRestorePhaseEvent(dir, phase string, pid int, extra map[string]any) error {
	ev := restorePhaseEvent{
		Time:  time.Now().UTC().Format(time.RFC3339Nano),
		Phase: phase,
		Dir:   dir,
		PID:   pid,
		Extra: extra,
	}
	b, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(dir, restoreEventsName), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.Write(append(b, '\n'))
	return err
}

func recordRestorePID(dir string, pid int) error {
	if err := writeRestoreHandoff(dir, pid); err != nil {
		return err
	}
	logRestorePhase(dir, "pidfile", pid, nil)
	return nil
}
