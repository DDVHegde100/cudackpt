package agent

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dhruvhegde/cudackpt/pkg/config"
	"github.com/dhruvhegde/cudackpt/pkg/metrics"
)

func TestRefreshGauges(t *testing.T) {
	root := t.TempDir()
	runDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "ckpt-1"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.ImageRoot = root
	cfg.RunDir = runDir
	metrics.Default.Set(metrics.ImagesGauge, 0)
	metrics.Default.Set(metrics.ShimsGauge, 0)
	RefreshGauges(cfg)
	_, gauges := metrics.Default.Snapshot()
	if gauges[metrics.ImagesGauge] != 1 {
		t.Fatalf("images=%v", gauges)
	}
}

func TestRunGCErrorIncrementsMetric(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root bypasses pin file permission checks")
	}
	metrics.Default.Set(metrics.GCErrorsTotal, 0)
	root := t.TempDir()
	pin := filepath.Join(root, "noperm")
	if err := os.WriteFile(pin, []byte("/var/lib/cudackpt/ckpt-1\n"), 0o000); err != nil {
		t.Fatal(err)
	}
	opts := OptionsFromConfig(config.Config{ImageRoot: root, RunDir: root})
	opts.PinFile = pin
	opts.GCMaxAge = time.Hour
	runGC(opts)
	counters, _ := metrics.Default.Snapshot()
	if counters[metrics.GCErrorsTotal] != 1 {
		t.Fatalf("errors=%v", counters[metrics.GCErrorsTotal])
	}
}

func TestOptionsFromConfig(t *testing.T) {
	t.Setenv("CUDACKPT_METRICS_ADDR", "127.0.0.1:19190")
	t.Setenv("CUDACKPT_AGENT_GC_INTERVAL", "1h")
	opts := OptionsFromConfig(config.Default())
	if opts.Listen != "127.0.0.1:19190" {
		t.Fatalf("listen=%q", opts.Listen)
	}
	if opts.GCInterval != time.Hour {
		t.Fatalf("gc=%v", opts.GCInterval)
	}
}

func TestHealthHandlerDeep(t *testing.T) {
	runDir := t.TempDir()
	cfg := config.Default()
	cfg.RunDir = runDir
	req := httptest.NewRequest(http.MethodGet, "/health?deep=1", nil)
	rec := httptest.NewRecorder()
	healthHandler(cfg)(rec, req)
	if rec.Code != http.StatusOK && rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("code=%d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("run_dir")) {
		t.Fatalf("body=%q", rec.Body.String())
	}
}

func TestHealthHandler(t *testing.T) {
	runDir := t.TempDir()
	cfg := config.Default()
	cfg.RunDir = runDir
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	healthHandler(cfg)(rec, req)
	if rec.Code != http.StatusOK && rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("code=%d", rec.Code)
	}
	if rec.Body.Len() == 0 {
		t.Fatal("empty body")
	}
}
