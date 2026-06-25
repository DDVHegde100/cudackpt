package storage

import (
	"path/filepath"
	"testing"
)

func TestTierSpillAndPaths(t *testing.T) {
	dir := t.TempDir()
	tier, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	n, err := tier.Spill("probe.bin", []byte("tier"))
	if err != nil || n != 4 {
		t.Fatalf("spill n=%d err=%v", n, err)
	}
	if tier.ManifestPath() != filepath.Join(dir, "manifest.bin") {
		t.Fatal("manifest path")
	}
	if tier.DevicePath() != filepath.Join(dir, "device.bin") {
		t.Fatal("device path")
	}
	data, err := tier.MmapRead("probe.bin")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "tier" {
		t.Fatalf("got %q", data)
	}
}
