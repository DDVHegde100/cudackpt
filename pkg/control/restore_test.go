package control

import (
	"testing"
	"time"
)

func TestRestorePollDelay(t *testing.T) {
	base := 100 * time.Millisecond
	if restorePollDelay(1, base) != base {
		t.Fatal("attempt 1")
	}
	if restorePollDelay(2, base) != 200*time.Millisecond {
		t.Fatal("attempt 2")
	}
	if restorePollDelay(10, base) != maxRestoreBackoff {
		t.Fatal("attempt cap")
	}
}

func TestReadRestoredPID(t *testing.T) {
	dir := t.TempDir()
	if readRestoredPID(dir) != 0 {
		t.Fatal("missing file")
	}
	if err := writeRestoreHandoff(dir, 12345); err != nil {
		t.Fatal(err)
	}
	if readRestoredPID(dir) != 12345 {
		t.Fatal("read pid")
	}
}

func TestResolveRestorePID(t *testing.T) {
	dir := t.TempDir()
	_ = writeRestoreHandoff(dir, 99)
	if got := resolveRestorePID(dir, 42, []int{1, 2}); got != 42 {
		t.Fatalf("criu pid got=%d", got)
	}
	if got := resolveRestorePID(dir, 0, nil); got != 99 {
		t.Fatalf("pidfile got=%d", got)
	}
	shimDir := t.TempDir()
	if got := resolveRestorePID(shimDir, 0, []int{55, 66}); got != 55 {
		t.Fatalf("shim fallback got=%d", got)
	}
}
