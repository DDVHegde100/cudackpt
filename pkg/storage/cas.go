package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

type CAS struct {
	Root string
}

func NewCAS(root string) (*CAS, error) {
	dir := filepath.Join(root, "cas")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &CAS{Root: dir}, nil
}

func (c *CAS) Put(data []byte) ([32]byte, error) {
	sum := sha256.Sum256(data)
	path := c.path(sum)
	if _, err := os.Stat(path); err == nil {
		return sum, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return [32]byte{}, err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return [32]byte{}, err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return [32]byte{}, err
	}
	return sum, nil
}

func (c *CAS) Get(hash [32]byte) ([]byte, error) {
	return os.ReadFile(c.path(hash))
}

func (c *CAS) Has(hash [32]byte) bool {
	_, err := os.Stat(c.path(hash))
	return err == nil
}

func (c *CAS) path(hash [32]byte) string {
	hexHash := hex.EncodeToString(hash[:])
	return filepath.Join(c.Root, hexHash[:2], hexHash[2:])
}

func HashString(h [32]byte) string {
	return hex.EncodeToString(h[:])
}

func ParseHash(s string) ([32]byte, error) {
	var out [32]byte
	b, err := hex.DecodeString(s)
	if err != nil {
		return out, fmt.Errorf("hash: %w", err)
	}
	if len(b) != 32 {
		return out, fmt.Errorf("hash length %d", len(b))
	}
	copy(out[:], b)
	return out, nil
}
