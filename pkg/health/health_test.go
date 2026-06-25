package health

import "testing"

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
