package control

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAppendRestorePhaseEvent(t *testing.T) {
	dir := t.TempDir()
	if err := appendRestorePhaseEvent(dir, "preflight", 0, map[string]any{"ok": true}); err != nil {
		t.Fatal(err)
	}
	if err := appendRestorePhaseEvent(dir, "criu", 42, nil); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, restoreEventsName))
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("empty events file")
	}
	lines := 0
	for _, c := range b {
		if c == '\n' {
			lines++
		}
	}
	if lines != 2 {
		t.Fatalf("lines=%d", lines)
	}
}

func TestRecordRestorePID(t *testing.T) {
	dir := t.TempDir()
	if err := recordRestorePID(dir, 1234); err != nil {
		t.Fatal(err)
	}
	if readRestoredPID(dir) != 1234 {
		t.Fatal("pid not written")
	}
	if _, err := os.Stat(filepath.Join(dir, restoreEventsName)); err != nil {
		t.Fatal("events missing")
	}
}
