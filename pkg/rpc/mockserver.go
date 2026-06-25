package rpc

import (
	"encoding/binary"
	"io"
	"net"
	"os"
)

func ServeMockAt(path string) (func(), error) {
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
			go serveMockConn(c)
		}
	}()
	return func() {
		close(stop)
		_ = ln.Close()
		_ = os.Remove(path)
	}, nil
}

func serveMockConn(c net.Conn) {
	defer c.Close()
	for {
		var op [4]byte
		if _, err := io.ReadFull(c, op[:]); err != nil {
			return
		}
		switch binary.BigEndian.Uint32(op[:]) {
		case OpPing, OpFreeze, OpResume:
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
