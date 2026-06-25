package config

import "testing"

func TestDefault(t *testing.T) {
	c := Default()
	if c.ImageRoot == "" || c.RunDir == "" {
		t.Fatalf("empty paths %+v", c)
	}
	if c.RestoreTimeout <= 0 || c.ShimPoll <= 0 {
		t.Fatalf("bad durations %+v", c)
	}
}
