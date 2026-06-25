package log

import "testing"

func TestInfoJSON(t *testing.T) {
	Info("test_event", map[string]any{"pid": 1, "stage": "freeze"})
}
