package control

import (
	"os"
	"path/filepath"

	"github.com/dhruvhegde/cudackpt/internal/ckpterr"
	"github.com/dhruvhegde/cudackpt/pkg/image"
)

func PreflightRestore(dir string) error {
	if !image.IsComplete(dir) {
		return ckpterr.E(ckpterr.Invalid, "checkpoint not finalized")
	}
	for _, name := range []string{"manifest.bin", "meta.bin", "criu"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err != nil {
			return ckpterr.Wrap(ckpterr.Invalid, name, err)
		}
	}
	meta, err := image.ReadMeta(filepath.Join(dir, "meta.bin"))
	if err != nil {
		return ckpterr.Wrap(ckpterr.Invalid, "meta", err)
	}
	if meta.Preload == "" {
		return ckpterr.E(ckpterr.Invalid, "missing LD_PRELOAD in meta")
	}
	entries, hdr, err := image.ReadManifest(filepath.Join(dir, "manifest.bin"))
	if err != nil {
		return ckpterr.Wrap(ckpterr.Invalid, "manifest", err)
	}
	if hdr.Magic != image.Magic || hdr.Count == 0 {
		return ckpterr.E(ckpterr.Invalid, "empty manifest")
	}
	if image.HasFlag(hdr.Flags, image.FlagCompressed) {
		if _, err := os.Stat(filepath.Join(dir, "device.zst")); err != nil {
			return ckpterr.Wrap(ckpterr.Invalid, "device.zst", err)
		}
	} else if _, err := os.Stat(filepath.Join(dir, "device.bin")); err != nil {
		if !image.HasFlag(hdr.Flags, image.FlagDelta) {
			return ckpterr.Wrap(ckpterr.Invalid, "device.bin", err)
		}
	}
	_ = entries
	return nil
}
