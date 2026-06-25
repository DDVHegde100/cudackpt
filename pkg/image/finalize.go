package image

import (
	"os"
	"path/filepath"
)

const completeName = "COMPLETE"

func StageDir(dest string) string {
	return dest + ".staging"
}

func Finalize(staging, dest string) error {
	if err := os.WriteFile(filepath.Join(staging, completeName), []byte("1\n"), 0o644); err != nil {
		return err
	}
	if err := os.RemoveAll(dest); err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.Rename(staging, dest)
}

func IsComplete(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, completeName))
	return err == nil
}

func WriteStaging(dir string, fn func(staging string) error) error {
	staging := StageDir(dir)
	_ = os.RemoveAll(staging)
	if err := os.MkdirAll(staging, 0o755); err != nil {
		return err
	}
	if err := fn(staging); err != nil {
		_ = os.RemoveAll(staging)
		return err
	}
	return Finalize(staging, dir)
}
