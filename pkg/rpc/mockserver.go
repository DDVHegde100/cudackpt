package rpc

import (
	"encoding/binary"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/dhruvhegde/cudackpt/pkg/image"
)

type MockOptions struct {
	Secret          string
	FailFreezeUntil int32
}

type mockState struct {
	secret          string
	failFreezeUntil int32
}

func ServeMockAt(path string) (func(), error) {
	return ServeMockWithOptions(path, MockOptions{})
}

func ServeMockWithOptions(path string, opts MockOptions) (func(), error) {
	st := &mockState{secret: opts.Secret, failFreezeUntil: opts.FailFreezeUntil}
	_ = os.Remove(path)
	ln, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
			}
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go serveMockConn(c, st)
		}
	}()
	return func() {
		close(stop)
		_ = ln.Close()
		_ = os.Remove(path)
	}, nil
}

func serveMockConn(c net.Conn, st *mockState) {
	defer c.Close()
	if st.secret != "" {
		var op [4]byte
		if _, err := io.ReadFull(c, op[:]); err != nil {
			return
		}
		if binary.BigEndian.Uint32(op[:]) != OpAuth {
			_ = writeU32(c, 2)
			return
		}
		n, err := readU32(c)
		if err != nil {
			return
		}
		buf := make([]byte, n)
		if _, err := io.ReadFull(c, buf); err != nil {
			return
		}
		if string(buf) != st.secret {
			_ = writeU32(c, 1)
			return
		}
		_ = writeU32(c, 0)
	}
	for {
		var op [4]byte
		if _, err := io.ReadFull(c, op[:]); err != nil {
			return
		}
		switch binary.BigEndian.Uint32(op[:]) {
		case OpPing, OpResume:
			_ = writeU32(c, 0)
		case OpFreeze:
			if st.failFreezeUntil > 0 && atomic.AddInt32(&st.failFreezeUntil, -1) >= 0 {
				_ = writeU32(c, 1)
				break
			}
			_ = writeU32(c, 0)
		case OpSnapshot:
			n, err := readU32(c)
			if err != nil {
				return
			}
			if n > 0 {
				buf := make([]byte, n)
				if _, err := io.ReadFull(c, buf); err != nil {
					return
				}
				dir := string(buf)
				_ = os.MkdirAll(dir, 0o755)
				payload := []byte{1, 2, 3, 4}
				_ = os.WriteFile(filepath.Join(dir, "device.bin"), payload, 0o644)
				_ = os.WriteFile(filepath.Join(dir, "dev.bin"), []byte{0, 0, 0, 0}, 0o644)
				e := image.Entry{Ptr: 0x1000, Size: uint64(len(payload)), Offset: 0, CRC32C: image.CRC32C(payload), Seq: 1}
				_ = image.WriteManifest(filepath.Join(dir, "manifest.bin"), []image.Entry{e})
			}
			_ = writeU32(c, 0)
		case OpRestore:
			n, err := readU32(c)
			if err != nil {
				return
			}
			if n > 0 {
				buf := make([]byte, n)
				if _, err := io.ReadFull(c, buf); err != nil {
					return
				}
			}
			_ = writeU32(c, 0)
		case OpStatus:
			_ = writeU32(c, 4)
		case OpStats:
			_ = writeU32(c, 0)
			for _, n := range []uint32{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4} {
				_ = writeU32(c, n)
			}
		case OpQuit:
			_ = writeU32(c, 0)
			return
		default:
			_ = writeU32(c, 1)
		}
	}
}
