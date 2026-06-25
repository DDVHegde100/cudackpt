package rpc

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestWriteReadU32(t *testing.T) {
	var buf bytes.Buffer
	if err := writeU32(&buf, 0x01020304); err != nil {
		t.Fatal(err)
	}
	v, err := readU32(&buf)
	if err != nil || v != 0x01020304 {
		t.Fatalf("v=%x err=%v", v, err)
	}
}

func TestOpConstants(t *testing.T) {
	if OpPing != 1 || OpResume != 6 || OpAuth != 9 {
		t.Fatalf("ops ping=%d resume=%d auth=%d", OpPing, OpResume, OpAuth)
	}
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], OpSnapshot)
	if binary.BigEndian.Uint32(b[:]) != OpSnapshot {
		t.Fatal("endian mismatch")
	}
}
