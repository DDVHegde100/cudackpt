package agent

import (
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
