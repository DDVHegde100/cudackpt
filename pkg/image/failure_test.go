package image

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDeviceMissing(t *testing.T) {
	dir := t.TempDir()
	if err := EnsureDeviceMaterialized(dir); err == nil {
		t.Fatal("expected missing device")
	}
}

func TestReadManifestInvalidMagic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.bin")
	if err := os.WriteFile(path, []byte{0, 0, 0, 0}, 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := ReadManifest(path)
	if err == nil {
		t.Fatal("expected invalid magic")
	}
}

func TestFinalizeRemovesStagingOnError(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "ckpt")
	err := WriteStaging(dest, func(staging string) error {
		return os.ErrInvalid
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if _, err := os.Stat(StageDir(dest)); !os.IsNotExist(err) {
		t.Fatal("staging should be removed")
	}
}
