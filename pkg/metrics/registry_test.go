package metrics

import (
	"bytes"
	"strings"
	"testing"
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

func TestDefaultRegistry(t *testing.T) {
	Default.Inc(CheckpointsTotal)
	c, _ := Default.Snapshot()
	if c[CheckpointsTotal] == 0 {
		t.Fatal("expected counter")
	}
}
