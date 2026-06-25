package health

import "testing"

func TestDeepProbe(t *testing.T) {
	s := DeepProbe()
	if len(s.Checks) < 7 {
		t.Fatalf("checks=%d", len(s.Checks))
	}
}
