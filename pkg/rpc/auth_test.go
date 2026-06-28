package rpc

import (
	"encoding/binary"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestAuthenticateRequired(t *testing.T) {
	dir := t.TempDir()
	sock := filepath.Join(dir, "auth.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Skip(err)
	}
	defer func() { _ = ln.Close() }()
	const secret = "test-secret"
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go handleMockConn(c, secret)
		}
	}()
	t.Setenv("CUDACKPT_RPC_SECRET", secret)
	cli, err := DialPath(sock)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cli.Close() }()
	if err := cli.Ping(); err != nil {
		t.Fatal(err)
	}
}

func TestAuthenticateRejectsBadSecret(t *testing.T) {
	dir := t.TempDir()
	sock := filepath.Join(dir, "auth-bad.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Skip(err)
	}
	defer func() { _ = ln.Close() }()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go handleMockConn(c, "good")
		}
	}()
	t.Setenv("CUDACKPT_RPC_SECRET", "bad")
	if _, err := DialPath(sock); err == nil {
		t.Fatal("expected auth failure")
	}
}

func handleMockConn(c net.Conn, secret string) {
	defer func() { _ = c.Close() }()
	if secret != "" {
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
		if string(buf) != secret {
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
		case OpPing, OpFreeze, OpResume:
			_ = writeU32(c, 0)
		case OpStatus:
			_ = writeU32(c, 4)
		case OpStats:
			_ = writeU32(c, 0)
			for _, n := range []uint32{2, 4096, 0, 0, 1, 0, 0, 0, 0, 0, 4} {
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

func TestAuthenticateSkippedWithoutEnv(t *testing.T) {
	if os.Getenv("CUDACKPT_RPC_SECRET") != "" {
		t.Setenv("CUDACKPT_RPC_SECRET", "")
	}
	dir := t.TempDir()
	sock := filepath.Join(dir, "open.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Skip(err)
	}
	defer func() { _ = ln.Close() }()
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
	defer func() { _ = cli.Close() }()
	if err := cli.Ping(); err != nil {
		t.Fatal(err)
	}
}
