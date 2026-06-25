package image

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMetaRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "meta.bin")
	in := Meta{Pid: 42, Dev: 1, Preload: "/lib/libcudackpt.so", Visible: "0"}
	if err := WriteMeta(path, in); err != nil {
		t.Fatal(err)
	}
	out, err := ReadMeta(path)
	if err != nil {
		t.Fatal(err)
	}
	if out.Pid != in.Pid || out.Dev != in.Dev || out.Preload != in.Preload || out.Visible != in.Visible {
		t.Fatalf("got %+v want %+v", out, in)
	}
}

func TestReadDev(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dev.bin")
	b := []byte{1, 0, 0, 0}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	d, err := ReadDev(path)
	if err != nil || d != 1 {
		t.Fatalf("dev=%d err=%v", d, err)
	}
}
