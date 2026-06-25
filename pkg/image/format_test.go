package image

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestManifestRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.bin")
	entries := []Entry{
		{Ptr: 0x1000, Size: 4096, Offset: 0, CRC32C: 0xdeadbeef, Seq: 1},
		{Ptr: 0x2000, Size: 8192, Offset: 4096, CRC32C: 0xcafebabe, Seq: 2},
	}
	if err := WriteManifest(path, entries); err != nil {
		t.Fatal(err)
	}
	got, hdr, err := ReadManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if hdr.Magic != Magic || hdr.Count != 2 || hdr.TotalBytes != 12288 {
		t.Fatalf("hdr %+v", hdr)
	}
	if len(got) != 2 || got[0].Seq != 1 || got[1].Ptr != 0x2000 {
		t.Fatalf("entries %+v", got)
	}
}

func TestCRC32CAndVerify(t *testing.T) {
	data := []byte("cudackpt chunk payload")
	want := CRC32C(data)
	if want == 0 {
		t.Fatal("zero crc")
	}
	r := bytes.NewReader(data)
	ok, err := VerifyChunk(r, 0, int64(len(data)), want)
	if err != nil || !ok {
		t.Fatalf("verify ok=%v err=%v", ok, err)
	}
	ok, err = VerifyChunk(r, 0, int64(len(data)), want^1)
	if err != nil || ok {
		t.Fatalf("mismatch ok=%v err=%v", ok, err)
	}
}

func TestVerifyChunkFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "device.bin")
	payload := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ok, err := VerifyChunk(f, 0, int64(len(payload)), CRC32C(payload))
	if err != nil || !ok {
		t.Fatalf("file verify ok=%v err=%v", ok, err)
	}
}
