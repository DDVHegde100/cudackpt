package image

import (
	"encoding/binary"
	"os"
)

const MetaMagic = 0x4d435054

type Meta struct {
	Magic   uint32
	Version uint16
	Flags   uint16
	Pid     uint32
	Dev     uint32
	Preload string
	Visible string
}

func WriteMeta(path string, m Meta) error {
	m.Magic = MetaMagic
	m.Version = 2
	pl := []byte(m.Preload)
	vl := []byte(m.Visible)
	b := make([]byte, 24+len(pl)+len(vl))
	binary.LittleEndian.PutUint32(b[0:], m.Magic)
	binary.LittleEndian.PutUint16(b[4:], m.Version)
	binary.LittleEndian.PutUint16(b[6:], m.Flags)
	binary.LittleEndian.PutUint32(b[8:], m.Pid)
	binary.LittleEndian.PutUint32(b[12:], m.Dev)
	binary.LittleEndian.PutUint32(b[16:], uint32(len(pl)))
	copy(b[20:], pl)
	o := 20 + len(pl)
	binary.LittleEndian.PutUint32(b[o:], uint32(len(vl)))
	copy(b[o+4:], vl)
	return os.WriteFile(path, b, 0o644)
}

func ReadMeta(path string) (Meta, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Meta{}, err
	}
	if len(b) < 16 {
		return Meta{}, os.ErrInvalid
	}
	m := Meta{
		Magic:   binary.LittleEndian.Uint32(b[0:4]),
		Version: binary.LittleEndian.Uint16(b[4:6]),
		Flags:   binary.LittleEndian.Uint16(b[6:8]),
		Pid:     binary.LittleEndian.Uint32(b[8:12]),
		Dev:     binary.LittleEndian.Uint32(b[12:16]),
	}
	if m.Magic != MetaMagic {
		return Meta{}, os.ErrInvalid
	}
	if m.Version >= 2 && len(b) >= 20 {
		pl := int(binary.LittleEndian.Uint32(b[16:20]))
		if 20+pl <= len(b) {
			m.Preload = string(b[20 : 20+pl])
			o := 20 + pl
			if o+4 <= len(b) {
				vl := int(binary.LittleEndian.Uint32(b[o : o+4]))
				if o+4+vl <= len(b) {
					m.Visible = string(b[o+4 : o+4+vl])
				}
			}
		}
		return m, nil
	}
	if len(b) > 16 {
		m.Preload = string(b[16:])
	}
	return m, nil
}

func ReadDev(path string) (uint32, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	if len(b) < 4 {
		return 0, os.ErrInvalid
	}
	return binary.LittleEndian.Uint32(b[:4]), nil
}
