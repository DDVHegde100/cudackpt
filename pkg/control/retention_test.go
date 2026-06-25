package control

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dhruvhegde/cudackpt/pkg/image"
)

func TestPlanImageGCRemovesStaging(t *testing.T) {
	root := t.TempDir()
	staging := filepath.Join(root, "ckpt-1.staging")
	if err := os.MkdirAll(staging, 0o755); err != nil {
		t.Fatal(err)
	}
	plan, err := PlanImageGC(GCOptions{Root: root, Now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Remove) != 1 || plan.Remove[0] != staging {
		t.Fatalf("remove=%v", plan.Remove)
	}
}

func TestPlanImageGCSkipsIncomplete(t *testing.T) {
	root := t.TempDir()
	incomplete := filepath.Join(root, "ckpt-2")
	if err := os.MkdirAll(incomplete, 0o755); err != nil {
		t.Fatal(err)
	}
	plan, err := PlanImageGC(GCOptions{Root: root, MaxAge: time.Hour, Now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Remove) != 0 {
		t.Fatalf("remove=%v", plan.Remove)
	}
}

func TestPlanImageGCAgeAndPin(t *testing.T) {
	root := t.TempDir()
	old := filepath.Join(root, "ckpt-old")
	if err := os.MkdirAll(old, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(old, "COMPLETE"), []byte("1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(old, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}
	pin := filepath.Join(root, "pins.txt")
	if err := os.WriteFile(pin, []byte(old+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, err := PlanImageGC(GCOptions{
		Root:    root,
		MaxAge:  24 * time.Hour,
		PinFile: pin,
		Now:     time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Remove) != 0 {
		t.Fatalf("pinned remove=%v", plan.Remove)
	}
	_ = os.WriteFile(pin, []byte("# empty\n"), 0o644)
	plan, err = PlanImageGC(GCOptions{Root: root, MaxAge: 24 * time.Hour, Now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Remove) != 1 {
		t.Fatalf("remove=%v", plan.Remove)
	}
}

func TestApplyGCPlanDryRun(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "ckpt-x.staging")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	removed, err := ApplyGCPlan(GCPlan{Remove: []string{target}}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(removed) != 1 {
		t.Fatal("dry run should list target")
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatal("dry run should not delete")
	}
}

func TestIsCompleteUsed(t *testing.T) {
	dir := t.TempDir()
	if image.IsComplete(dir) {
		t.Fatal("incomplete")
	}
	if err := os.WriteFile(filepath.Join(dir, "COMPLETE"), []byte("1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !image.IsComplete(dir) {
		t.Fatal("complete")
	}
}
