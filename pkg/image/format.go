package image

import (
	"encoding/binary"
	"hash/crc32"
	"io"
	"os"
)

const Magic = 0x434B5054

type Header struct {
	Magic      uint32
	Version    uint16
	Flags      uint16
	Count      uint32
	TotalBytes uint64
}

type Entry struct {
	Ptr         uint64
	Size        uint64
	Offset      uint64
	CRC32C      uint32
	Seq         uint32
	ContentHash [32]byte
}

const entryV1Size = 40
const entryV2Size = 72

func WriteManifest(path string, entries []Entry) error {
	return WriteManifestFlags(path, entries, FlagNone, VersionV2)
}

func WriteManifestFlags(path string, entries []Entry, flags uint16, version uint16) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	var total uint64
	for _, e := range entries {
		total += e.Size
	}
	h := Header{Magic: Magic, Version: version, Flags: flags, Count: uint32(len(entries)), TotalBytes: total}
	if err := binary.Write(f, binary.LittleEndian, &h); err != nil {
		return err
	}
	for _, e := range entries {
		if err := writeEntry(f, e, version); err != nil {
			return err
		}
	}
	return nil
}

func writeEntry(w io.Writer, e Entry, version uint16) error {
	if err := binary.Write(w, binary.LittleEndian, e.Ptr); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, e.Size); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, e.Offset); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, e.CRC32C); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, e.Seq); err != nil {
		return err
	}
	if version >= VersionV2 {
		if _, err := w.Write(e.ContentHash[:]); err != nil {
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
		e, err := readEntry(f, h.Version)
		if err != nil {
			return nil, Header{}, err
		}
		out[i] = e
	}
	return out, h, nil
}

func readEntry(r io.Reader, version uint16) (Entry, error) {
	var e Entry
	if err := binary.Read(r, binary.LittleEndian, &e.Ptr); err != nil {
		return Entry{}, err
	}
	if err := binary.Read(r, binary.LittleEndian, &e.Size); err != nil {
		return Entry{}, err
	}
	if err := binary.Read(r, binary.LittleEndian, &e.Offset); err != nil {
		return Entry{}, err
	}
	if err := binary.Read(r, binary.LittleEndian, &e.CRC32C); err != nil {
		return Entry{}, err
	}
	if err := binary.Read(r, binary.LittleEndian, &e.Seq); err != nil {
		return Entry{}, err
	}
	if version >= VersionV2 {
		if _, err := io.ReadFull(r, e.ContentHash[:]); err != nil {
			return Entry{}, err
		}
	}
	return e, nil
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

func EntrySize(version uint16) int {
	if version >= VersionV2 {
		return entryV2Size
	}
	return entryV1Size
}
