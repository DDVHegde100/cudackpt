package control

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dhruvhegde/cudackpt/pkg/config"
	"github.com/dhruvhegde/cudackpt/pkg/image"
	"github.com/dhruvhegde/cudackpt/pkg/rpc"
	"github.com/dhruvhegde/cudackpt/third_party/criu"
)

func TestRestoreHermeticWithFakeCRIU(t *testing.T) {
	root := t.TempDir()
	runDir := filepath.Join(root, "run")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatal(err)
	}
	img := filepath.Join(root, "image")
	writeTestImage(t, img)

	const pid = 9001
	sock := filepath.Join(runDir, "9001.sock")
	stop, err := rpc.ServeMockAt(sock)
	if err != nil {
		t.Skip(err)
	}
	defer stop()

	cfg := config.Default()
	cfg.RunDir = runDir
	cfg.RestoreTimeout = 3 * time.Second
	cfg.RetryBackoff = 20 * time.Millisecond

	fake := &criu.Fake{RestorePID: pid}
	orc := NewWithCRIU(cfg, fake)
	got, err := orc.Restore(img)
	if err != nil {
		t.Fatal(err)
	}
	if got != pid {
		t.Fatalf("pid got=%d want=%d", got, pid)
	}
	if readRestoredPID(img) != pid {
		t.Fatal("restored.pid not updated")
	}
	events, err := os.ReadFile(filepath.Join(img, restoreEventsName))
	if err != nil || len(events) == 0 {
		t.Fatal("restore events missing")
	}
}

func TestCheckpointWithRetryHermetic(t *testing.T) {
	root := t.TempDir()
	runDir := filepath.Join(root, "run")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatal(err)
	}
	const pid = 7001
	sock := filepath.Join(runDir, "7001.sock")
	stop, err := rpc.ServeMockWithOptions(sock, rpc.MockOptions{FailFreezeUntil: 2})
	if err != nil {
		t.Skip(err)
	}
	defer stop()

	out := filepath.Join(root, "ckpt")
	cfg := config.Default()
	cfg.RunDir = runDir
	cfg.MaxRetries = 3
	cfg.RetryBackoff = 5 * time.Millisecond
	t.Setenv("LD_PRELOAD", "/usr/lib/libcudackpt.so")

	orc := NewWithCRIU(cfg, &criu.Fake{})
	policy := RetryPolicy{MaxAttempts: 3, Backoff: 5 * time.Millisecond}
	if err := orc.CheckpointWithRetry(pid, out, policy); err != nil {
		t.Fatal(err)
	}
	if !image.IsComplete(out) {
		t.Fatal("checkpoint not finalized")
	}
}

func TestRestoreHermeticFakeCRIUError(t *testing.T) {
	root := t.TempDir()
	img := filepath.Join(root, "image")
	writeTestImage(t, img)
	cfg := config.Default()
	cfg.RunDir = filepath.Join(root, "run")
	fake := &criu.Fake{RestoreErr: os.ErrPermission}
	orc := NewWithCRIU(cfg, fake)
	if _, err := orc.Restore(img); err == nil {
		t.Fatal("expected criu error")
	}
}
