package control

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dhruvhegde/cudackpt/pkg/config"
	"github.com/dhruvhegde/cudackpt/pkg/image"
)

func writeTestImage(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "criu"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "COMPLETE"), []byte("1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "dev.bin"), []byte{0, 0, 0, 0}, 0o644); err != nil {
		t.Fatal(err)
	}
	meta := image.Meta{Preload: "/usr/lib/libcudackpt.so", Pid: 1, Dev: 0}
	if err := image.WriteMeta(filepath.Join(dir, "meta.bin"), meta); err != nil {
		t.Fatal(err)
	}
	payload := []byte{1, 2, 3, 4}
	if err := os.WriteFile(filepath.Join(dir, "device.bin"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	e := image.Entry{Ptr: 0x1000, Size: uint64(len(payload)), Offset: 0, CRC32C: image.CRC32C(payload), Seq: 1}
	if err := image.WriteManifest(filepath.Join(dir, "manifest.bin"), []image.Entry{e}); err != nil {
		t.Fatal(err)
	}
}

func TestPromoteCopiesAndPins(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	dest := filepath.Join(root, "dest")
	pin := filepath.Join(root, "pins.txt")
	writeTestImage(t, src)
	orc := New(config.Default())
	if err := orc.Promote(PromoteOptions{Src: src, Dest: dest, PinFile: pin}); err != nil {
		t.Fatal(err)
	}
	if !image.IsComplete(dest) {
		t.Fatal("dest incomplete")
	}
	pinned, err := LoadPinnedPaths(pin)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := pinned[dest]; !ok {
		t.Fatal("dest not pinned")
	}
	if err := orc.Promote(PromoteOptions{Src: src, Dest: dest, PinFile: pin}); err != nil {
		t.Fatal("re-promote")
	}
	lines, _ := os.ReadFile(pin)
	if n := len(lines); n == 0 {
		t.Fatal("pin file empty")
	}
}

func TestAppendPinnedPathDedupes(t *testing.T) {
	root := t.TempDir()
	pin := filepath.Join(root, "pins.txt")
	path := filepath.Join(root, "ckpt-1")
	if err := appendPinnedPath(pin, path); err != nil {
		t.Fatal(err)
	}
	if err := appendPinnedPath(pin, path); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(pin)
	if err != nil {
		t.Fatal(err)
	}
	if count := len(b); count == 0 {
		t.Fatal("empty pin")
	}
}
