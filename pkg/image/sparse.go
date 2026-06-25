package image

import (
	"encoding/binary"
	"os"
	"path/filepath"
)

const sparseMagic = 0x53505253
const minZeroRun = 4096

type SparseEntry struct {
	Offset uint64
	Size   uint64
}

func ApplySparse(dir string) error {
	manifestPath := filepath.Join(dir, "manifest.bin")
	entries, hdr, err := ReadManifest(manifestPath)
	if err != nil {
		return err
	}
	devicePath := filepath.Join(dir, "device.bin")
	f, err := os.Open(devicePath)
	if err != nil {
		return err
	}
	defer f.Close()
	var sparse []SparseEntry
	for _, e := range entries {
		if e.Size < minZeroRun {
			continue
		}
		buf := make([]byte, e.Size)
		if _, err := f.ReadAt(buf, int64(e.Offset)); err != nil {
			return err
		}
		if isZero(buf) {
			sparse = append(sparse, SparseEntry{Offset: e.Offset, Size: e.Size})
		}
	}
	if len(sparse) == 0 {
		return nil
	}
	spPath := filepath.Join(dir, "sparse.bin")
	sf, err := os.Create(spPath)
	if err != nil {
		return err
	}
	defer sf.Close()
	if err := binary.Write(sf, binary.LittleEndian, uint32(sparseMagic)); err != nil {
		return err
	}
	if err := binary.Write(sf, binary.LittleEndian, uint32(len(sparse))); err != nil {
		return err
	}
	for _, s := range sparse {
		if err := binary.Write(sf, binary.LittleEndian, s); err != nil {
			return err
		}
	}
	hdr.Flags |= FlagSparse
	return WriteManifestFlags(manifestPath, entries, hdr.Flags, hdr.Version)
}

func isZero(buf []byte) bool {
	for _, b := range buf {
		if b != 0 {
			return false
		}
	}
	return true
}

func ReadSparse(path string) ([]SparseEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var magic uint32
	if err := binary.Read(f, binary.LittleEndian, &magic); err != nil {
		return nil, err
	}
	if magic != sparseMagic {
		return nil, os.ErrInvalid
	}
	var count uint32
	if err := binary.Read(f, binary.LittleEndian, &count); err != nil {
		return nil, err
	}
	out := make([]SparseEntry, count)
	for i := range out {
		if err := binary.Read(f, binary.LittleEndian, &out[i]); err != nil {
			return nil, err
		}
	}
	return out, nil
}
