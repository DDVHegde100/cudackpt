package control

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dhruvhegde/cudackpt/internal/ckpterr"
)

func TestPreflightMissingComplete(t *testing.T) {
	dir := t.TempDir()
	err := PreflightRestore(dir)
	if err == nil {
		t.Fatal("expected error")
	}
	if e, ok := err.(*ckpterr.Error); !ok || e.Code != ckpterr.Invalid {
		t.Fatalf("err=%v", err)
	}
}

func TestPreflightBadManifest(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "COMPLETE"), []byte("1\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "meta.bin"), []byte("bad"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "manifest.bin"), []byte("bad"), 0o644)
	_ = os.Mkdir(filepath.Join(dir, "criu"), 0o755)
	err := PreflightRestore(dir)
	if err == nil {
		t.Fatal("expected manifest error")
	}
}

func TestPreflightMissingCRIU(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "COMPLETE"), []byte("1\n"), 0o644)
	err := PreflightRestore(dir)
	if err == nil {
		t.Fatal("expected criu dir error")
	}
}
