package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dhruvhegde/cudackpt/pkg/image"
)

func TestRenderImage(t *testing.T) {
	dir := t.TempDir()
	entries := []image.Entry{
		{Ptr: 0x100, Size: 16, Offset: 0, CRC32C: 1, Seq: 1},
	}
	if err := image.WriteManifest(filepath.Join(dir, "manifest.bin"), entries); err != nil {
		t.Fatal(err)
	}
	if err := image.WriteMeta(filepath.Join(dir, "meta.bin"), image.Meta{Pid: 7, Dev: 0}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "device.bin"), make([]byte, 16), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := RenderImage(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, needle := range []string{"manifest_version", "source_pid", "7", "chunks"} {
		if !strings.Contains(out, needle) {
			t.Fatalf("missing %q in:\n%s", needle, out)
		}
	}
}
