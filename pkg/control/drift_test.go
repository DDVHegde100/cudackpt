package control

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dhruvhegde/cudackpt/pkg/image"
)

func TestCompareImagesNoDrift(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a")
	b := filepath.Join(dir, "b")
	for _, d := range []string{a, b} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		payload := []byte{1, 2, 3, 4}
		if err := os.WriteFile(filepath.Join(d, "device.bin"), payload, 0o644); err != nil {
			t.Fatal(err)
		}
		e := image.Entry{Ptr: 0x100, Size: 4, Offset: 0, CRC32C: image.CRC32C(payload), Seq: 1}
		if err := image.WriteManifest(filepath.Join(d, "manifest.bin"), []image.Entry{e}); err != nil {
			t.Fatal(err)
		}
		if err := image.WriteMeta(filepath.Join(d, "meta.bin"), image.Meta{Pid: 1, Dev: 0, Preload: "/lib/x.so"}); err != nil {
			t.Fatal(err)
		}
	}
	rep, err := CompareImages(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Manifest)+len(rep.MetaDrift)+len(rep.FlagDrift) != 0 {
		t.Fatalf("drift=%s", FormatDrift(rep))
	}
}

func TestPreflightRestore(t *testing.T) {
	dir := t.TempDir()
	if err := PreflightRestore(dir); err == nil {
		t.Fatal("expected incomplete checkpoint failure")
	}
}
