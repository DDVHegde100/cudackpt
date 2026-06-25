package image

import (
	"os"
	"path/filepath"
)

type ProcessOpts struct {
	Compress bool
	Sparse   bool
	Dedup    bool
	Parent   string
}

func ProcessImage(dir string, opts ProcessOpts) error {
	if opts.Sparse {
		if err := ApplySparse(dir); err != nil {
			return err
		}
	}
	if opts.Dedup {
		if err := DedupDevice(dir); err != nil {
			return err
		}
	}
	if opts.Compress {
		if err := CompressDevice(dir); err != nil {
			return err
		}
	}
	if opts.Parent != "" {
		if err := WriteDelta(dir, opts.Parent); err != nil {
			return err
		}
	}
	return nil
}

func ParentFromEnv() string {
	return os.Getenv("CUDACKPT_PARENT_IMAGE")
}

func OptsFromEnv() ProcessOpts {
	return ProcessOpts{
		Compress: os.Getenv("CUDACKPT_COMPRESS") == "1",
		Sparse:   os.Getenv("CUDACKPT_SPARSE") == "1",
		Dedup:    os.Getenv("CUDACKPT_DEDUP") == "1",
		Parent:   ParentFromEnv(),
	}
}

func StagingSnapshot(dir string, snapshotFn func(staging string) error, opts ProcessOpts) error {
	return WriteStaging(dir, func(staging string) error {
		if err := snapshotFn(staging); err != nil {
			return err
		}
		return ProcessImage(staging, opts)
	})
}

func DevicePath(dir string) string {
	return filepath.Join(dir, "device.bin")
}
