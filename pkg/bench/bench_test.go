package bench

import "testing"

func TestFormatTable(t *testing.T) {
	out := FormatTable(Result{Op: "ping", Count: 10, PerOpUs: 12.5})
	if out == "" {
		t.Fatal("empty table")
	}
}
