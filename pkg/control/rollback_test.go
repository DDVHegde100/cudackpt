package control

import (
	"testing"

	"github.com/dhruvhegde/cudackpt/pkg/config"
)

func TestStopProcessInvalid(t *testing.T) {
	if err := stopProcess(0); err != nil {
		t.Fatal(err)
	}
}

func TestRollbackValidateOnly(t *testing.T) {
	dir := t.TempDir()
	orc := New(testConfig(t))
	if _, err := orc.Rollback(dir, 0); err == nil {
		t.Fatal("expected validate failure")
	}
}

func testConfig(t *testing.T) config.Config {
	t.Helper()
	cfg := config.Default()
	cfg.ImageRoot = t.TempDir()
	return cfg
}
