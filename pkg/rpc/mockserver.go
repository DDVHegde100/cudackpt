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
	StatusSeq       []uint32
}

type mockState struct {
	secret          string
	failFreezeUntil int32
	statusSeq       []uint32
	statusIdx       int32
}

func ServeMockAt(path string) (func(), error) {
	return ServeMockWithOptions(path, MockOptions{})
}

func ServeMockWithOptions(path string, opts MockOptions) (func(), error) {
	st := &mockState{
		secret:          opts.Secret,
		failFreezeUntil: opts.FailFreezeUntil,
		statusSeq:       opts.StatusSeq,
	}
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

func mockStatus(st *mockState) uint32 {
	if len(st.statusSeq) == 0 {
		return 4
	}
	idx := atomic.AddInt32(&st.statusIdx, 1) - 1
	if int(idx) >= len(st.statusSeq) {
		return st.statusSeq[len(st.statusSeq)-1]
	}
	return st.statusSeq[idx]
}

func replyU32(c net.Conn, v uint32) bool {
	return writeU32(c, v) == nil
}

func serveMockConn(c net.Conn, st *mockState) {
	defer func() { _ = c.Close() }()
	if st.secret != "" {
		var op [4]byte
		if _, err := io.ReadFull(c, op[:]); err != nil {
			return
		}
		if binary.BigEndian.Uint32(op[:]) != OpAuth {
			replyU32(c, 2)
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
			replyU32(c, 1)
			return
		}
		if !replyU32(c, 0) {
			return
		}
	}
	for {
		var op [4]byte
		if _, err := io.ReadFull(c, op[:]); err != nil {
			return
		}
		switch binary.BigEndian.Uint32(op[:]) {
		case OpPing, OpResume:
			if !replyU32(c, 0) {
				return
			}
		case OpFreeze:
			if st.failFreezeUntil > 0 && atomic.AddInt32(&st.failFreezeUntil, -1) >= 0 {
				if !replyU32(c, 1) {
					return
				}
				break
			}
			if !replyU32(c, 0) {
				return
			}
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
				if err := os.MkdirAll(dir, 0o755); err != nil {
					replyU32(c, 1)
					return
				}
				payload := []byte{1, 2, 3, 4}
				if err := os.WriteFile(filepath.Join(dir, "device.bin"), payload, 0o644); err != nil {
					replyU32(c, 1)
					return
				}
				_ = os.WriteFile(filepath.Join(dir, "dev.bin"), []byte{0, 0, 0, 0}, 0o644)
				e := image.Entry{Ptr: 0x1000, Size: uint64(len(payload)), Offset: 0, CRC32C: image.CRC32C(payload), Seq: 1}
				if err := image.WriteManifest(filepath.Join(dir, "manifest.bin"), []image.Entry{e}); err != nil {
					replyU32(c, 1)
					return
				}
			}
			if !replyU32(c, 0) {
				return
			}
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
			if !replyU32(c, 0) {
				return
			}
		case OpStatus:
			if !replyU32(c, mockStatus(st)) {
				return
			}
		case OpStats:
			if !replyU32(c, 0) {
				return
			}
			for _, n := range []uint32{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4} {
				if !replyU32(c, n) {
					return
				}
			}
		case OpQuit:
			replyU32(c, 0)
			return
		default:
			replyU32(c, 1)
		}
	}
}
