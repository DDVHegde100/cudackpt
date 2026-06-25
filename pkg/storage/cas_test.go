package storage

import (
	"bytes"
	"testing"
)

func TestCASPutGet(t *testing.T) {
	dir := t.TempDir()
	c, err := NewCAS(dir)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("chunk-bytes")
	h1, err := c.Put(data)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := c.Put(data)
	if err != nil || h1 != h2 {
		t.Fatal("dedup put")
	}
	got, err := c.Get(h1)
	if err != nil || !bytes.Equal(got, data) {
		t.Fatalf("get %q err=%v", got, err)
	}
}
