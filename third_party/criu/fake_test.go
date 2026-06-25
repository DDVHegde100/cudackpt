package criu

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFakeRestoreWritesPID(t *testing.T) {
	dir := t.TempDir()
	f := &Fake{RestorePID: 7777}
	pid, err := f.Restore(dir, filepath.Join(dir, "restore.log"), nil)
	if err != nil || pid != 7777 {
		t.Fatalf("pid=%d err=%v", pid, err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "restored.pid"))
	if err != nil || string(b) != "7777\n" {
		t.Fatalf("pidfile=%q err=%v", b, err)
	}
}
