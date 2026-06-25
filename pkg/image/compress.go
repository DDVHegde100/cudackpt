package image

import (
	"encoding/binary"
	"io"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
)

const deviceZst = "device.zst"

func CompressDevice(dir string) error {
	src := filepath.Join(dir, "device.bin")
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	enc, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedDefault))
	if err != nil {
		return err
	}
	defer enc.Close()
	compressed := enc.EncodeAll(data, nil)
	if err := os.WriteFile(filepath.Join(dir, deviceZst), compressed, 0o644); err != nil {
		return err
	}
	entries, hdr, err := ReadManifest(filepath.Join(dir, "manifest.bin"))
	if err != nil {
		return err
	}
	hdr.Flags |= FlagCompressed
	return WriteManifestFlags(filepath.Join(dir, "manifest.bin"), entries, hdr.Flags, hdr.Version)
}

func DecompressDevice(dir string) error {
	entries, hdr, err := ReadManifest(filepath.Join(dir, "manifest.bin"))
	if err != nil {
		return err
	}
	if !HasFlag(hdr.Flags, FlagCompressed) {
		return nil
	}
	f, err := os.Open(filepath.Join(dir, deviceZst))
	if err != nil {
		return err
	}
	defer f.Close()
	dec, err := zstd.NewReader(f)
	if err != nil {
		return err
	}
	defer dec.Close()
	out, err := io.ReadAll(dec)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "device.bin"), out, 0o644); err != nil {
		return err
	}
	hdr.Flags &^= FlagCompressed
	return WriteManifestFlags(filepath.Join(dir, "manifest.bin"), entries, hdr.Flags, hdr.Version)
}

func EnsureDeviceMaterialized(dir string) error {
	hdr, err := readManifestHeader(filepath.Join(dir, "manifest.bin"))
	if err != nil {
		return err
	}
	if HasFlag(hdr.Flags, FlagDelta) {
		if err := ApplyDelta(dir); err != nil {
			return err
		}
	}
	if HasFlag(hdr.Flags, FlagCompressed) {
		return DecompressDevice(dir)
	}
	if _, err := os.Stat(filepath.Join(dir, "device.bin")); os.IsNotExist(err) {
		return err
	}
	return nil
}

func readManifestHeader(path string) (Header, error) {
	f, err := os.Open(path)
	if err != nil {
		return Header{}, err
	}
	defer f.Close()
	var h Header
	if err := binary.Read(f, binary.LittleEndian, &h); err != nil {
		return Header{}, err
	}
	if h.Magic != Magic {
		return Header{}, os.ErrInvalid
	}
	return h, nil
}
