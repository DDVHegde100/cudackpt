package main

import (
	"bytes"
	"fmt"
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
	sock := filepath.Join(runDir, fmt.Sprintf("%d.sock", pid))
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
	if !bytes.Contains(buf.Bytes(), []byte(fmt.Sprintf("%d", pid))) {
		t.Fatalf("serve ps output=%q", buf.String())
	}
}

func TestRunServeHealth(t *testing.T) {
	root := t.TempDir()
	runDir := filepath.Join(root, "run")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatal(err)
	}
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

	oldIn := os.Stdin
	rIn, wIn, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdin = rIn
	if _, err := wIn.Write([]byte("health\n")); err != nil {
		t.Fatal(err)
	}
	wIn.Close()
	defer func() { os.Stdin = oldIn }()

	errCh := make(chan error, 1)
	go func() {
		errCh <- runServe(cfg, orc)
	}()
	serveErr := <-errCh
	if err := wOut.Close(); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, rOut); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("run_dir")) {
		t.Fatalf("serve health output=%q", buf.String())
	}
	if serveErr != nil && serveErr.Error() != "health degraded" {
		t.Fatal(serveErr)
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
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	defer func() { os.Stdin = oldIn }()
	if err := runServe(cfg, orc); err == nil {
		t.Fatal("expected error for unknown serve command")
	}
}
