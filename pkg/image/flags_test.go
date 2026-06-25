package image

import (
	"path/filepath"
	"testing"
)

func TestFlags(t *testing.T) {
	f := FlagCompressed | FlagSparse
	if !HasFlag(f, FlagCompressed) || !HasFlag(f, FlagSparse) {
		t.Fatal("flags")
	}
	if HasFlag(f, FlagDedup) {
		t.Fatal("dedup unset")
	}
}

func TestManifestV2Roundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.bin")
	var hash [32]byte
	hash[0] = 0xab
	entries := []Entry{
		{Ptr: 0x1000, Size: 4096, Offset: 0, CRC32C: 1, Seq: 1, ContentHash: hash},
	}
	if err := WriteManifestFlags(path, entries, FlagCompressed, VersionV2); err != nil {
		t.Fatal(err)
	}
	got, hdr, err := ReadManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if hdr.Version != VersionV2 || hdr.Flags != FlagCompressed {
		t.Fatalf("hdr %+v", hdr)
	}
	if got[0].ContentHash != hash {
		t.Fatalf("hash %+v", got[0].ContentHash)
	}
}
