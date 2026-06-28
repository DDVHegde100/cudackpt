package rpc

import (
	"path/filepath"
	"testing"
)

func TestServeMockWithAuth(t *testing.T) {
	dir := t.TempDir()
	sock := filepath.Join(dir, "auth.sock")
	const secret = "mock-secret"
	stop, err := ServeMockWithOptions(sock, MockOptions{Secret: secret})
	if err != nil {
		t.Skip(err)
	}
	defer stop()
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

func TestServeMockRejectsBadAuth(t *testing.T) {
	dir := t.TempDir()
	sock := filepath.Join(dir, "bad.sock")
	stop, err := ServeMockWithOptions(sock, MockOptions{Secret: "good"})
	if err != nil {
		t.Skip(err)
	}
	defer stop()
	t.Setenv("CUDACKPT_RPC_SECRET", "bad")
	if _, err := DialPath(sock); err == nil {
		t.Fatal("expected auth failure")
	}
}

func TestServeMockFailFreezeUntil(t *testing.T) {
	dir := t.TempDir()
	sock := filepath.Join(dir, "freeze.sock")
	opts := MockOptions{FailFreezeUntil: 2}
	stop, err := ServeMockWithOptions(sock, opts)
	if err != nil {
		t.Skip(err)
	}
	defer stop()
	cli, err := DialPath(sock)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cli.Close() }()
	for i := 0; i < 2; i++ {
		if err := cli.Freeze(); err == nil {
			t.Fatalf("freeze %d should fail", i)
		}
	}
	if err := cli.Freeze(); err != nil {
		t.Fatal(err)
	}
}
