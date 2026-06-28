package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/dhruvhegde/cudackpt/pkg/config"
	"github.com/dhruvhegde/cudackpt/pkg/control"
	"github.com/dhruvhegde/cudackpt/pkg/rpc"
)

func TestRunServePs(t *testing.T) {
	root := t.TempDir()
	runDir := filepath.Join(root, "run")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatal(err)
	}
	const pid = 4242
	sock := filepath.Join(runDir, "4242.sock")
	stop, err := rpc.ServeMockAt(sock)
	if err != nil {
		t.Skip(err)
	}
	defer stop()

	cfg := config.Default()
	cfg.RunDir = runDir
	orc := control.New(cfg)

	oldOut := os.Stdout
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = wOut
	defer func() { os.Stdout = oldOut }()

	errCh := make(chan error, 1)
	go func() {
		errCh <- runServe(cfg, orc)
	}()
	if err := <-errCh; err != nil {
		t.Fatal(err)
	}
	if err := wOut.Close(); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, rOut); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("4242")) {
		t.Fatalf("serve ps output=%q", buf.String())
	}
}

func TestRunServeUnknownCommand(t *testing.T) {
	cfg := config.Default()
	orc := control.New(cfg)
	oldIn := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdin = r
	if _, err := w.Write([]byte("bogus\n")); err != nil {
		t.Fatal(err)
	}
	w.Close()
	defer func() { os.Stdin = oldIn }()
	if err := runServe(cfg, orc); err == nil {
		t.Fatal("expected error for unknown serve command")
	}
}
