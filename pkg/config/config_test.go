package config

import "testing"

func TestDefault(t *testing.T) {
	c := Default()
	if c.ImageRoot == "" || c.RunDir == "" || c.MaxRetries <= 0 {
		t.Fatalf("bad default %+v", c)
	}
}
