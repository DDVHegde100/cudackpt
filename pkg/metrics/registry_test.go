package metrics

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestWritePrometheus(t *testing.T) {
	r := NewRegistry()
	r.Inc("cudackpt_checkpoints_total")
	r.Set("cudackpt_images", 3)
	var buf bytes.Buffer
	r.WritePrometheus(&buf)
	out := buf.String()
	if !strings.Contains(out, "cudackpt_checkpoints_total 1") {
		t.Fatalf("out=%q", out)
	}
	if !strings.Contains(out, "cudackpt_images 3") {
		t.Fatalf("out=%q", out)
	}
	if !strings.Contains(out, "# TYPE cudackpt_checkpoints_total counter") {
		t.Fatalf("out=%q", out)
	}
}

func TestServeContextShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- ServeContext(ctx, "127.0.0.1:0", NewRegistry())
	}()
	time.Sleep(100 * time.Millisecond)
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for metrics server shutdown")
	}
}

func TestDefaultRegistry(t *testing.T) {
	Default.Inc(CheckpointsTotal)
	c, _ := Default.Snapshot()
	if c[CheckpointsTotal] == 0 {
		t.Fatal("expected counter")
	}
}
