package bench

import (
	"path/filepath"
	"testing"

	"github.com/dhruvhegde/cudackpt/pkg/rpc"
)

func TestFormatTable(t *testing.T) {
	out := FormatTable(Result{Op: "ping", Count: 10, PerOpUs: 12.5})
	if out == "" {
		t.Fatal("empty table")
	}
}

func TestPingCustomRunDir(t *testing.T) {
	root := t.TempDir()
	runDir := filepath.Join(root, "run")
	const pid = 4242
	sock := filepath.Join(runDir, "4242.sock")
	stop, err := rpc.ServeMockAt(sock)
	if err != nil {
		t.Skip(err)
	}
	defer stop()
	r := Ping(runDir, pid, 3)
	if r.Errors != 0 {
		t.Fatalf("errors=%d", r.Errors)
	}
}
