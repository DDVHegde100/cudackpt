package health

import (
	"path/filepath"
	"testing"
)

func TestProbe(t *testing.T) {
	s := Probe()
	if len(s.Checks) < 3 {
		t.Fatalf("checks=%d", len(s.Checks))
	}
	out := Format(s)
	if out == "" {
		t.Fatal("empty format")
	}
}

func TestProbeCustomRunDir(t *testing.T) {
	root := t.TempDir()
	runDir := filepath.Join(root, "custom-run")
	s := ProbeWith(runDir)
	var found bool
	for _, c := range s.Checks {
		if c.Name == "run_dir" {
			found = true
			if !c.OK || c.Detail != runDir {
				t.Fatalf("check=%+v", c)
			}
		}
	}
	if !found {
		t.Fatal("run_dir check missing")
	}
}
