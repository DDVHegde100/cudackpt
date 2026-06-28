package image

import (
	"os"
	"path/filepath"
	"testing"
)

func writePipelineFixture(t *testing.T, dir string, payload []byte) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "device.bin"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	e := Entry{Ptr: 0x1000, Size: uint64(len(payload)), Offset: 0, CRC32C: CRC32C(payload), Seq: 1}
	if err := WriteManifest(filepath.Join(dir, "manifest.bin"), []Entry{e}); err != nil {
		t.Fatal(err)
	}
}

func TestProcessImageSparseThenCompress(t *testing.T) {
	dir := t.TempDir()
	payload := make([]byte, 4096)
	writePipelineFixture(t, dir, payload)
	if err := ProcessImage(dir, ProcessOpts{Sparse: true, Compress: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "device.zst")); err != nil {
		t.Fatal("expected compressed artifact")
	}
}

func TestProcessImageDedupOnly(t *testing.T) {
	dir := t.TempDir()
	writePipelineFixture(t, dir, []byte{9, 8, 7, 6})
	if err := ProcessImage(dir, ProcessOpts{Dedup: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "cas")); err != nil {
		t.Fatal("expected cas dir")
	}
}

func TestProcessImageDeltaRequiresParent(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	child := filepath.Join(root, "child")
	writePipelineFixture(t, parent, make([]byte, 4096))
	writePipelineFixture(t, child, append(make([]byte, 4096), 0xff))
	if err := ProcessImage(child, ProcessOpts{Parent: parent}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(child, "delta.bin")); err != nil {
		t.Fatal("expected delta")
	}
}

func TestOptsFromEnv(t *testing.T) {
	t.Setenv("CUDACKPT_COMPRESS", "1")
	t.Setenv("CUDACKPT_SPARSE", "1")
	t.Setenv("CUDACKPT_DEDUP", "1")
	t.Setenv("CUDACKPT_PARENT_IMAGE", "/tmp/parent")
	opts := OptsFromEnv()
	if !opts.Compress || !opts.Sparse || !opts.Dedup || opts.Parent != "/tmp/parent" {
		t.Fatalf("opts=%+v", opts)
	}
}
