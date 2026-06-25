package image

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFinalizeStaging(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "ckpt-1")
	err := WriteStaging(dest, func(staging string) error {
		return os.WriteFile(filepath.Join(staging, "probe"), []byte("ok"), 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	if !IsComplete(dest) {
		t.Fatal("not complete")
	}
	b, err := os.ReadFile(filepath.Join(dest, "probe"))
	if err != nil || string(b) != "ok" {
		t.Fatal("probe missing")
	}
}

func TestDeltaAgainstParent(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	child := filepath.Join(root, "child")
	for _, d := range []string{parent, child} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	base := make([]byte, 4096)
	changed := append([]byte(nil), base...)
	changed[0] = 0xff
	for name, payload := range map[string][]byte{"parent": base, "child": changed} {
		dir := parent
		if name == "child" {
			dir = child
		}
		if err := os.WriteFile(filepath.Join(dir, "device.bin"), payload, 0o644); err != nil {
			t.Fatal(err)
		}
		e := Entry{Ptr: 0x1000, Size: 4096, Offset: 0, CRC32C: CRC32C(payload), Seq: 1}
		if err := WriteManifest(filepath.Join(dir, "manifest.bin"), []Entry{e}); err != nil {
			t.Fatal(err)
		}
	}
	if err := WriteDelta(child, parent); err != nil {
		t.Fatal(err)
	}
	_ = os.Remove(filepath.Join(child, "device.bin"))
	if err := ApplyDelta(child); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(filepath.Join(child, "device.bin"))
	if err != nil || out[0] != 0xff {
		t.Fatalf("delta apply out[0]=%v err=%v", out[0], err)
	}
}
