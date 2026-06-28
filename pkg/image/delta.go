package image

import (
	"encoding/binary"
	"os"
	"path/filepath"
)

const deltaMagic = 0x44454C54

type DeltaRecord struct {
	Ptr         uint64
	Size        uint64
	DeviceOff   uint64
	PatchOffset uint64
}

func WriteDelta(dir, parentDir string) error {
	if err := EnsureDeviceMaterialized(parentDir); err != nil {
		return err
	}
	if err := EnsureDeviceMaterialized(dir); err != nil {
		return err
	}
	entries, hdr, err := ReadManifest(filepath.Join(dir, "manifest.bin"))
	if err != nil {
		return err
	}
	parentEntries, _, err := ReadManifest(filepath.Join(parentDir, "manifest.bin"))
	if err != nil {
		return err
	}
	parentByPtr := make(map[uint64]Entry, len(parentEntries))
	for _, pe := range parentEntries {
		parentByPtr[pe.Ptr] = pe
	}
	curDev, err := os.ReadFile(filepath.Join(dir, "device.bin"))
	if err != nil {
		return err
	}
	var delta []DeltaRecord
	var patch []byte
	for _, e := range entries {
		pe, ok := parentByPtr[e.Ptr]
		if ok && pe.Size == e.Size && pe.CRC32C == e.CRC32C {
			continue
		}
		rec := DeltaRecord{
			Ptr: e.Ptr, Size: e.Size, DeviceOff: e.Offset,
			PatchOffset: uint64(len(patch)),
		}
		patch = append(patch, curDev[e.Offset:e.Offset+e.Size]...)
		delta = append(delta, rec)
	}
	if len(delta) == 0 {
		return nil
	}
	if err := os.WriteFile(filepath.Join(dir, "device.patch"), patch, 0o644); err != nil {
		return err
	}
	df, err := os.Create(filepath.Join(dir, "delta.bin"))
	if err != nil {
		return err
	}
	defer func() { _ = df.Close() }()
	parent := []byte(parentDir)
	if err := binary.Write(df, binary.LittleEndian, uint32(deltaMagic)); err != nil {
		return err
	}
	if err := binary.Write(df, binary.LittleEndian, uint32(len(delta))); err != nil {
		return err
	}
	if err := binary.Write(df, binary.LittleEndian, uint32(len(parent))); err != nil {
		return err
	}
	if _, err := df.Write(parent); err != nil {
		return err
	}
	for _, d := range delta {
		if err := binary.Write(df, binary.LittleEndian, d); err != nil {
			return err
		}
	}
	hdr.Flags |= FlagDelta
	return WriteManifestFlags(filepath.Join(dir, "manifest.bin"), entries, hdr.Flags, hdr.Version)
}

func ApplyDelta(dir string) error {
	path := filepath.Join(dir, "delta.bin")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer func() { _ = f.Close() }()
	var magic, count, plen uint32
	if err := binary.Read(f, binary.LittleEndian, &magic); err != nil {
		return err
	}
	if magic != deltaMagic {
		return os.ErrInvalid
	}
	if err := binary.Read(f, binary.LittleEndian, &count); err != nil {
		return err
	}
	if err := binary.Read(f, binary.LittleEndian, &plen); err != nil {
		return err
	}
	parent := make([]byte, plen)
	if _, err := f.Read(parent); err != nil {
		return err
	}
	parentDir := string(parent)
	if err := EnsureDeviceMaterialized(parentDir); err != nil {
		return err
	}
	base, err := os.ReadFile(filepath.Join(parentDir, "device.bin"))
	if err != nil {
		return err
	}
	entries, _, err := ReadManifest(filepath.Join(dir, "manifest.bin"))
	if err != nil {
		return err
	}
	var total uint64
	for _, e := range entries {
		end := e.Offset + e.Size
		if end > total {
			total = end
		}
	}
	if uint64(len(base)) < total {
		grow := make([]byte, total)
		copy(grow, base)
		base = grow
	} else if uint64(len(base)) > total {
		base = base[:total]
	}
	patch, err := os.ReadFile(filepath.Join(dir, "device.patch"))
	if err != nil {
		return err
	}
	for i := uint32(0); i < count; i++ {
		var rec DeltaRecord
		if err := binary.Read(f, binary.LittleEndian, &rec); err != nil {
			return err
		}
		copy(base[rec.DeviceOff:rec.DeviceOff+rec.Size], patch[rec.PatchOffset:rec.PatchOffset+rec.Size])
	}
	return os.WriteFile(filepath.Join(dir, "device.bin"), base, 0o644)
}
