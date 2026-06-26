package control

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dhruvhegde/cudackpt/internal/ckpterr"
	"github.com/dhruvhegde/cudackpt/pkg/config"
	"github.com/dhruvhegde/cudackpt/pkg/image"
)

func TestValidateImageOK(t *testing.T) {
	dir := t.TempDir()
	writeTestImage(t, dir)
	orc := New(config.Default())
	if err := orc.ValidateImage(dir); err != nil {
		t.Fatal(err)
	}
}

func TestValidateImageMissingDevBin(t *testing.T) {
	dir := t.TempDir()
	writeTestImage(t, dir)
	_ = os.Remove(filepath.Join(dir, "dev.bin"))
	orc := New(config.Default())
	err := orc.ValidateImage(dir)
	if err == nil {
		t.Fatal("expected error")
	}
	if e, ok := err.(*ckpterr.Error); !ok || e.Code != ckpterr.Invalid {
		t.Fatalf("err=%v", err)
	}
}

func TestValidateImageNotFinalized(t *testing.T) {
	dir := t.TempDir()
	writeTestImage(t, dir)
	_ = os.Remove(filepath.Join(dir, "COMPLETE"))
	orc := New(config.Default())
	err := orc.ValidateImage(dir)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateImageMissingPreload(t *testing.T) {
	dir := t.TempDir()
	writeTestImage(t, dir)
	meta := image.Meta{Preload: "", Pid: 1, Dev: 0}
	if err := image.WriteMeta(filepath.Join(dir, "meta.bin"), meta); err != nil {
		t.Fatal(err)
	}
	orc := New(config.Default())
	err := orc.ValidateImage(dir)
	if err == nil {
		t.Fatal("expected error")
	}
}
