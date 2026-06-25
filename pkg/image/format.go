package image

import (
	"encoding/binary"
	"hash/crc32"
	"io"
	"os"
)

const Magic = 0x434B5054
const Version = 1

type Header struct {
	Magic      uint32
	Version    uint16
	Flags      uint16
	Count      uint32
	TotalBytes uint64
}

type Entry struct {
	Ptr    uint64
	Size   uint64
	Offset uint64
	CRC32C uint32
	Seq    uint32
}

func WriteManifest(path string, entries []Entry) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	var total uint64
	for _, e := range entries {
		total += e.Size
	}
	h := Header{Magic: Magic, Version: Version, Count: uint32(len(entries)), TotalBytes: total}
	if err := binary.Write(f, binary.LittleEndian, &h); err != nil {
		return err
	}
	for _, e := range entries {
		if err := binary.Write(f, binary.LittleEndian, &e); err != nil {
			return err
		}
	}
	return nil
}

func ReadManifest(path string) ([]Entry, Header, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, Header{}, err
	}
	defer f.Close()
	var h Header
	if err := binary.Read(f, binary.LittleEndian, &h); err != nil {
		return nil, Header{}, err
	}
	if h.Magic != Magic {
		return nil, Header{}, os.ErrInvalid
	}
	out := make([]Entry, h.Count)
	for i := range out {
		if err := binary.Read(f, binary.LittleEndian, &out[i]); err != nil {
			return nil, Header{}, err
		}
	}
	return out, h, nil
}

func CRC32C(data []byte) uint32 {
	return crc32.Checksum(data, crc32.MakeTable(crc32.Castagnoli))
}

func VerifyChunk(r io.ReaderAt, off int64, size int64, want uint32) (bool, error) {
	buf := make([]byte, size)
	if _, err := r.ReadAt(buf, off); err != nil {
		return false, err
	}
	return CRC32C(buf) == want, nil
}
