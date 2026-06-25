//go:build linux

package health

import "testing"

func TestProbeCapsLinux(t *testing.T) {
	c := probeCaps()
	if c.Name != "caps" {
		t.Fatalf("name=%q", c.Name)
	}
}
