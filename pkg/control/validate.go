package control

import (
	"os"
	"path/filepath"

	"github.com/dhruvhegde/cudackpt/internal/ckpterr"
	"github.com/dhruvhegde/cudackpt/pkg/image"
)

func (o *Orchestrator) ValidateImage(dir string) error {
	if err := o.verifyImage(dir); err != nil {
		return err
	}
	for _, name := range []string{"dev.bin", "meta.bin", "criu"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err != nil {
			return ckpterr.Wrap(ckpterr.Invalid, name, err)
		}
	}
	if !image.IsComplete(dir) {
		return ckpterr.E(ckpterr.Invalid, "checkpoint not finalized")
	}
	if m, err := image.ReadMeta(filepath.Join(dir, "meta.bin")); err != nil {
		return ckpterr.Wrap(ckpterr.Invalid, "meta", err)
	} else if m.Preload == "" {
		return ckpterr.E(ckpterr.Invalid, "missing LD_PRELOAD in meta")
	}
	return nil
}
