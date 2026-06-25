package rpc

import (
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
			go handleMockConn(c, "")
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
	handleMockConn(c, "")
}
