package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

type Tier struct {
	Dir string
}

func New(dir string) (*Tier, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Tier{Dir: dir}, nil
}

func (t *Tier) Spill(name string, data []byte) (int64, error) {
	p := filepath.Join(t.Dir, name)
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()
	n, err := f.Write(data)
	return int64(n), err
}

func (t *Tier) MmapRead(name string) ([]byte, error) {
	p := filepath.Join(t.Dir, name)
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	size := st.Size()
	if size == 0 {
		return nil, nil
	}
	data, err := syscall.Mmap(int(f.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("mmap: %w", err)
	}
	out := make([]byte, size)
	copy(out, data)
	_ = syscall.Munmap(data)
	return out, nil
}

func (t *Tier) DevicePath() string {
	return filepath.Join(t.Dir, "device.bin")
}

func (t *Tier) ManifestPath() string {
	return filepath.Join(t.Dir, "manifest.bin")
}
