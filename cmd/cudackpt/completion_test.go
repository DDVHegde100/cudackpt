package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestEmitBashCompletion(t *testing.T) {
	var buf bytes.Buffer
	if err := emitCompletion("bash", &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "complete -F _cudackpt_completions cudackpt") {
		t.Fatalf("missing bash complete directive: %q", out)
	}
	if !strings.Contains(out, "checkpoint restore rollback") {
		t.Fatal("missing command list")
	}
}

func TestEmitZshCompletion(t *testing.T) {
	var buf bytes.Buffer
	if err := emitCompletion("zsh", &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "#compdef cudackpt") {
		t.Fatalf("missing zsh compdef: %q", out)
	}
}

func TestEmitCompletionUnsupported(t *testing.T) {
	err := emitCompletion("fish", &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for unsupported shell")
	}
}
