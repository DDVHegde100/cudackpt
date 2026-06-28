package control

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dhruvhegde/cudackpt/pkg/config"
	"github.com/dhruvhegde/cudackpt/pkg/image"
	"github.com/dhruvhegde/cudackpt/pkg/metrics"
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

	before, _ := metrics.Default.Snapshot()
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
	after, _ := metrics.Default.Snapshot()
	if after[metrics.RestoresTotal]-before[metrics.RestoresTotal] != 1 {
		t.Fatalf("restores before=%v after=%v", before, after)
	}
	if after[metrics.RestoreFailuresTotal]-before[metrics.RestoreFailuresTotal] != 0 {
		t.Fatalf("failures before=%v after=%v", before, after)
	}
}

func TestRestoreHermeticWithAuth(t *testing.T) {
	root := t.TempDir()
	runDir := filepath.Join(root, "run")
	img := filepath.Join(root, "image")
	writeTestImage(t, img)
	const pid = 9100
	const secret = "restore-secret"
	sock := filepath.Join(runDir, "9100.sock")
	stop, err := rpc.ServeMockWithOptions(sock, rpc.MockOptions{Secret: secret})
	if err != nil {
		t.Skip(err)
	}
	defer stop()
	t.Setenv("CUDACKPT_RPC_SECRET", secret)

	cfg := config.Default()
	cfg.RunDir = runDir
	cfg.RestoreTimeout = 3 * time.Second
	orc := NewWithCRIU(cfg, &criu.Fake{RestorePID: pid})
	got, err := orc.Restore(img)
	if err != nil {
		t.Fatal(err)
	}
	if got != pid {
		t.Fatalf("pid=%d", got)
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

	before, _ := metrics.Default.Snapshot()
	orc := NewWithCRIU(cfg, &criu.Fake{})
	policy := RetryPolicy{MaxAttempts: 3, Backoff: 5 * time.Millisecond}
	if err := orc.CheckpointWithRetry(pid, out, policy); err != nil {
		t.Fatal(err)
	}
	if !image.IsComplete(out) {
		t.Fatal("checkpoint not finalized")
	}
	after, _ := metrics.Default.Snapshot()
	if after[metrics.CheckpointsTotal]-before[metrics.CheckpointsTotal] != 1 {
		t.Fatalf("checkpoints before=%v after=%v", before, after)
	}
	if after[metrics.CheckpointFailures]-before[metrics.CheckpointFailures] != 0 {
		t.Fatalf("failures before=%v after=%v", before, after)
	}
}

func TestCheckpointWithRetryExhaustionMetrics(t *testing.T) {
	root := t.TempDir()
	runDir := filepath.Join(root, "run")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatal(err)
	}
	const pid = 7050
	sock := filepath.Join(runDir, "7050.sock")
	stop, err := rpc.ServeMockWithOptions(sock, rpc.MockOptions{FailFreezeUntil: 100})
	if err != nil {
		t.Skip(err)
	}
	defer stop()

	cfg := config.Default()
	cfg.RunDir = runDir
	out := filepath.Join(root, "fail")
	orc := NewWithCRIU(cfg, &criu.Fake{})
	policy := RetryPolicy{MaxAttempts: 3, Backoff: time.Millisecond}

	before, _ := metrics.Default.Snapshot()
	err = orc.CheckpointWithRetry(pid, out, policy)
	if err == nil {
		t.Fatal("expected checkpoint failure")
	}
	after, _ := metrics.Default.Snapshot()
	if after[metrics.CheckpointsTotal]-before[metrics.CheckpointsTotal] != 0 {
		t.Fatalf("checkpoints before=%v after=%v", before, after)
	}
	if after[metrics.CheckpointFailures]-before[metrics.CheckpointFailures] != 1 {
		t.Fatalf("failures before=%v after=%v", before, after)
	}
}

func TestCheckpointWithRetryAuthHermetic(t *testing.T) {
	root := t.TempDir()
	runDir := filepath.Join(root, "run")
	const pid = 7100
	const secret = "ckpt-secret"
	sock := filepath.Join(runDir, "7100.sock")
	stop, err := rpc.ServeMockWithOptions(sock, rpc.MockOptions{Secret: secret})
	if err != nil {
		t.Skip(err)
	}
	defer stop()
	t.Setenv("CUDACKPT_RPC_SECRET", secret)
	t.Setenv("LD_PRELOAD", "/usr/lib/libcudackpt.so")

	cfg := config.Default()
	cfg.RunDir = runDir
	out := filepath.Join(root, "ckpt-auth")
	orc := NewWithCRIU(cfg, &criu.Fake{})
	if err := orc.CheckpointWithRetry(pid, out, RetryPolicy{MaxAttempts: 1, Backoff: time.Millisecond}); err != nil {
		t.Fatal(err)
	}
}

func TestRollbackHermetic(t *testing.T) {
	root := t.TempDir()
	runDir := filepath.Join(root, "run")
	img := filepath.Join(root, "image")
	writeTestImage(t, img)
	const pid = 9200
	sock := filepath.Join(runDir, "9200.sock")
	stop, err := rpc.ServeMockAt(sock)
	if err != nil {
		t.Skip(err)
	}
	defer stop()

	before, _ := metrics.Default.Snapshot()
	cfg := config.Default()
	cfg.RunDir = runDir
	cfg.RestoreTimeout = 3 * time.Second
	orc := NewWithCRIU(cfg, &criu.Fake{RestorePID: pid})
	got, err := orc.Rollback(img, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got != pid {
		t.Fatalf("pid=%d", got)
	}
	after, _ := metrics.Default.Snapshot()
	if after[metrics.RollbacksTotal]-before[metrics.RollbacksTotal] != 1 {
		t.Fatalf("rollbacks before=%v after=%v", before, after)
	}
}

func TestRestoreFailureMetrics(t *testing.T) {
	root := t.TempDir()
	img := filepath.Join(root, "image")
	writeTestImage(t, img)
	cfg := config.Default()
	cfg.RunDir = filepath.Join(root, "run")
	before, _ := metrics.Default.Snapshot()
	orc := NewWithCRIU(cfg, &criu.Fake{RestoreErr: os.ErrPermission})
	if _, err := orc.Restore(img); err == nil {
		t.Fatal("expected criu error")
	}
	after, _ := metrics.Default.Snapshot()
	if after[metrics.RestoresTotal]-before[metrics.RestoresTotal] != 0 {
		t.Fatalf("restores before=%v after=%v", before, after)
	}
	if after[metrics.RestoreFailuresTotal]-before[metrics.RestoreFailuresTotal] != 1 {
		t.Fatalf("failures before=%v after=%v", before, after)
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
