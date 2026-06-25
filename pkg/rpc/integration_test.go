package rpc

import (
	"encoding/binary"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestMockShimRPC(t *testing.T) {
	if os.Getuid() != 0 && os.Getenv("CUDACKPT_RPC_TEST") == "" {
		t.Skip("unix socket bind may require permissions")
	}
	dir := t.TempDir()
	sock := filepath.Join(dir, "mock.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Skip(err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go handleMock(c)
		}
	}()
	cli, err := DialPath(sock)
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	if err := cli.Ping(); err != nil {
		t.Fatal(err)
	}
	st, err := cli.Status()
	if err != nil || st != 4 {
		t.Fatalf("status=%d err=%v", st, err)
	}
	stats, err := cli.Stats()
	if err != nil || stats.AllocCount != 2 {
		t.Fatalf("stats=%+v err=%v", stats, err)
	}
}

func handleMock(c net.Conn) {
	defer c.Close()
	var op [4]byte
	if _, err := c.Read(op[:]); err != nil {
		return
	}
	v := binary.BigEndian.Uint32(op[:])
	switch v {
	case OpPing, OpFreeze, OpResume:
		_ = writeU32(c, 0)
	case OpStatus:
		_ = writeU32(c, 4)
	case OpStats:
		_ = writeU32(c, 0)
		for _, n := range []uint32{2, 4096, 0, 0, 1, 0, 0, 0, 0, 0, 4} {
			_ = writeU32(c, n)
		}
	default:
		_ = writeU32(c, 1)
	}
}
