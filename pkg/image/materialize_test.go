package image

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDeviceMaterializedFromCompressed(t *testing.T) {
	dir := t.TempDir()
	payload := []byte{0xde, 0xad, 0xbe, 0xef}
	writePipelineFixture(t, dir, payload)
	if err := CompressDevice(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, "device.bin")); err != nil {
		t.Fatal(err)
	}
	if err := EnsureDeviceMaterialized(dir); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "device.bin"))
	if err != nil || string(got) != string(payload) {
		t.Fatalf("payload mismatch got=%v err=%v", got, err)
	}
}

func TestEnsureDeviceMaterializedFromDelta(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	child := filepath.Join(root, "child")
	base := make([]byte, 4096)
	changed := append([]byte(nil), base...)
	changed[0] = 0x42
	writePipelineFixture(t, parent, base)
	writePipelineFixture(t, child, changed)
	if err := WriteDelta(child, parent); err != nil {
		t.Fatal(err)
	}
	_ = os.Remove(filepath.Join(child, "device.bin"))
	if err := EnsureDeviceMaterialized(child); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(child, "device.bin"))
	if err != nil || got[0] != 0x42 {
		t.Fatalf("delta materialize failed got=%v err=%v", got, err)
	}
}
