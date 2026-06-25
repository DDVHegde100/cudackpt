package image

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompressRoundtrip(t *testing.T) {
	dir := t.TempDir()
	payload := []byte("cudackpt compress test payload with repetition repetition repetition")
	if err := os.WriteFile(filepath.Join(dir, "device.bin"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	entries := []Entry{{Ptr: 1, Size: uint64(len(payload)), Offset: 0, CRC32C: CRC32C(payload), Seq: 1}}
	if err := WriteManifest(filepath.Join(dir, "manifest.bin"), entries); err != nil {
		t.Fatal(err)
	}
	if err := CompressDevice(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "device.zst")); err != nil {
		t.Fatal(err)
	}
	if err := DecompressDevice(dir); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(filepath.Join(dir, "device.bin"))
	if err != nil || string(out) != string(payload) {
		t.Fatalf("payload mismatch")
	}
}
