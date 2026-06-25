package image

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSparseZeroRun(t *testing.T) {
	dir := t.TempDir()
	zeros := make([]byte, 8192)
	entries := []Entry{{Ptr: 0x1000, Size: 8192, Offset: 0, CRC32C: CRC32C(zeros), Seq: 1}}
	if err := os.WriteFile(filepath.Join(dir, "device.bin"), zeros, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteManifest(filepath.Join(dir, "manifest.bin"), entries); err != nil {
		t.Fatal(err)
	}
	if err := ApplySparse(dir); err != nil {
		t.Fatal(err)
	}
	_, hdr, err := ReadManifest(filepath.Join(dir, "manifest.bin"))
	if err != nil || !HasFlag(hdr.Flags, FlagSparse) {
		t.Fatalf("flags=%v err=%v", hdr.Flags, err)
	}
	runs, err := ReadSparse(filepath.Join(dir, "sparse.bin"))
	if err != nil || len(runs) != 1 {
		t.Fatalf("runs=%v err=%v", runs, err)
	}
}

func TestDedupDevice(t *testing.T) {
	dir := t.TempDir()
	a := []byte("duplicate-chunk-data-------------")
	b := append([]byte(nil), a...)
	payload := append(a, b...)
	if err := os.WriteFile(filepath.Join(dir, "device.bin"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	entries := []Entry{
		{Ptr: 1, Size: uint64(len(a)), Offset: 0, CRC32C: CRC32C(a), Seq: 1},
		{Ptr: 2, Size: uint64(len(b)), Offset: uint64(len(a)), CRC32C: CRC32C(b), Seq: 2},
	}
	if err := WriteManifest(filepath.Join(dir, "manifest.bin"), entries); err != nil {
		t.Fatal(err)
	}
	if err := DedupDevice(dir); err != nil {
		t.Fatal(err)
	}
	_, hdr, err := ReadManifest(filepath.Join(dir, "manifest.bin"))
	if err != nil || !HasFlag(hdr.Flags, FlagDedup) {
		t.Fatalf("hdr=%+v err=%v", hdr, err)
	}
}
