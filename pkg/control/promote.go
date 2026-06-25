package control

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhruvhegde/cudackpt/internal/ckpterr"
)

type PromoteOptions struct {
	Src     string
	Dest    string
	PinFile string
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func copyDir(src, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return ckpterr.E(ckpterr.Invalid, "source not a directory")
	}
	return filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if fi.IsDir() {
			return os.MkdirAll(target, fi.Mode())
		}
		return copyFile(path, target, fi.Mode())
	})
}

func appendPinnedPath(pinFile, path string) error {
	path = filepath.Clean(path)
	pinned, err := LoadPinnedPaths(pinFile)
	if err != nil {
		return err
	}
	if _, ok := pinned[path]; ok {
		return nil
	}
	f, err := os.OpenFile(pinFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.WriteString(f, path+"\n")
	return err
}

func (o *Orchestrator) Promote(opts PromoteOptions) error {
	src := filepath.Clean(opts.Src)
	dest := filepath.Clean(opts.Dest)
	if src == dest {
		return ckpterr.E(ckpterr.Invalid, "source and dest are the same")
	}
	if err := o.ValidateImage(src); err != nil {
		return err
	}
	if err := os.RemoveAll(dest); err != nil {
		return ckpterr.Wrap(ckpterr.IO, "remove dest", err)
	}
	if err := copyDir(src, dest); err != nil {
		return ckpterr.Wrap(ckpterr.IO, "copy", err)
	}
	if err := o.ValidateImage(dest); err != nil {
		_ = os.RemoveAll(dest)
		return ckpterr.Wrap(ckpterr.Invalid, "dest validate", err)
	}
	if opts.PinFile != "" {
		if err := appendPinnedPath(opts.PinFile, dest); err != nil {
			return ckpterr.Wrap(ckpterr.IO, "pin", err)
		}
	}
	return nil
}

func ParsePromotePin(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return os.Getenv("CUDACKPT_PIN_FILE")
	}
	return path
}
