package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadFileAndEnvOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cudackpt.conf")
	body := "image_root=/data/images\nrun_dir=/run/test\nmax_retries=5\nretry_backoff=2s\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CUDACKPT_CONFIG", path)
	t.Setenv("CUDACKPT_IMAGE_ROOT", "/override")
	cfg := Load()
	if cfg.ImageRoot != "/override" {
		t.Fatalf("env override got %q", cfg.ImageRoot)
	}
	if cfg.RunDir != "/run/test" || cfg.MaxRetries != 5 || cfg.RetryBackoff != 2*time.Second {
		t.Fatalf("file load got %+v", cfg)
	}
}

func TestLoadWithWarningsUnknownKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cudackpt.conf")
	body := "image-root=/bad\nrestore_timeout=not-a-duration\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CUDACKPT_CONFIG", path)
	r := LoadWithWarnings()
	if len(r.Warnings) < 2 {
		t.Fatalf("warnings=%v", r.Warnings)
	}
	found := false
	for _, w := range r.Warnings {
		if strings.Contains(w, "unknown config key") {
			found = true
		}
	}
	if !found {
		t.Fatalf("warnings=%v", r.Warnings)
	}
}
