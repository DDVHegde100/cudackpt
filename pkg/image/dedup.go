package image

import (
	"os"
	"path/filepath"

	"github.com/dhruvhegde/cudackpt/pkg/storage"
)

func DedupDevice(dir string) error {
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
	defer func() { _ = f.Close() }()
	cas, err := storage.NewCAS(dir)
	if err != nil {
		return err
	}
	seen := make(map[[32]byte]struct{})
	unique := 0
	for i := range entries {
		buf := make([]byte, entries[i].Size)
		if _, err := f.ReadAt(buf, int64(entries[i].Offset)); err != nil {
			return err
		}
		hash, err := cas.Put(buf)
		if err != nil {
			return err
		}
		entries[i].ContentHash = hash
		if _, ok := seen[hash]; !ok {
			seen[hash] = struct{}{}
			unique++
		}
	}
	mapPath := filepath.Join(dir, "device.map")
	mf, err := os.Create(mapPath)
	if err != nil {
		return err
	}
	for _, e := range entries {
		line := storage.HashString(e.ContentHash) + " " + itoa64(e.Size) + " " + itoa64(e.Ptr) + "\n"
		if _, err := mf.WriteString(line); err != nil {
			_ = mf.Close()
			return err
		}
	}
	_ = mf.Close()
	hdr.Flags |= FlagDedup
	if err := WriteManifestFlags(manifestPath, entries, hdr.Flags, hdr.Version); err != nil {
		return err
	}
	return nil
}

func itoa64(v uint64) string {
	if v == 0 {
		return "0"
	}
	var buf [32]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}
