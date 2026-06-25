package control

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dhruvhegde/cudackpt/pkg/config"
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
