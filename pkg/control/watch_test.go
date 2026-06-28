package control

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dhruvhegde/cudackpt/pkg/config"
	"github.com/dhruvhegde/cudackpt/pkg/rpc"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	fn()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = old
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestWatchShimUntilRunning(t *testing.T) {
	root := t.TempDir()
	runDir := filepath.Join(root, "run")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatal(err)
	}
	const pid = 7000
	sock := filepath.Join(runDir, "7000.sock")
	stop, err := rpc.ServeMockWithOptions(sock, rpc.MockOptions{StatusSeq: []uint32{1, 4}})
	if err != nil {
		t.Skip(err)
	}
	defer stop()

	cfg := config.Default()
	cfg.RunDir = runDir
	orc := New(cfg)
	out := captureStdout(t, func() {
		if err := WatchShimWith(orc, pid, 10*time.Millisecond, WatchOptions{UntilRunning: true}, nil); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "frozen") || !strings.Contains(out, "running") {
		t.Fatalf("output=%q", out)
	}
}

func TestWatchShimTimeout(t *testing.T) {
	root := t.TempDir()
	runDir := filepath.Join(root, "run")
	const pid = 8000
	sock := filepath.Join(runDir, "8000.sock")
	stop, err := rpc.ServeMockWithOptions(sock, rpc.MockOptions{StatusSeq: []uint32{1}})
	if err != nil {
		t.Skip(err)
	}
	defer stop()

	cfg := config.Default()
	cfg.RunDir = runDir
	orc := New(cfg)
	err = WatchShimWith(orc, pid, 5*time.Millisecond, WatchOptions{Timeout: 25 * time.Millisecond}, nil)
	if err == nil {
		t.Fatal("expected timeout")
	}
}

func TestWatchShimStop(t *testing.T) {
	root := t.TempDir()
	runDir := filepath.Join(root, "run")
	const pid = 9000
	sock := filepath.Join(runDir, "9000.sock")
	mockStop, err := rpc.ServeMockWithOptions(sock, rpc.MockOptions{StatusSeq: []uint32{1}})
	if err != nil {
		t.Skip(err)
	}
	defer mockStop()

	cfg := config.Default()
	cfg.RunDir = runDir
	orc := New(cfg)
	done := make(chan struct{})
	go func() {
		time.Sleep(30 * time.Millisecond)
		close(done)
	}()
	if err := WatchShimWith(orc, pid, 10*time.Millisecond, WatchOptions{}, done); err != nil {
		t.Fatal(err)
	}
}
