package control

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRestoreCandidates(t *testing.T) {
	tests := []struct {
		name    string
		primary int
		shims   []int
		want    []int
	}{
		{"criu pid first", 42, []int{10, 42, 7}, []int{42, 10, 7}},
		{"no primary", 0, []int{3, 1}, []int{3, 1}},
		{"single shim", 5, []int{5}, []int{5}},
		{"empty shims", 9, nil, []int{9}},
		{"dedupe primary", 100, []int{100}, []int{100}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := restoreCandidates(tc.primary, tc.shims)
			if len(got) != len(tc.want) {
				t.Fatalf("len got=%v want=%v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("got=%v want=%v", got, tc.want)
				}
			}
		})
	}
}

func TestSortedShimsOrder(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"300.sock", "100.sock", "200.sock"} {
		path := filepath.Join(dir, name)
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
	}
	got, err := sortedShims(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0] != 100 || got[1] != 200 || got[2] != 300 {
		t.Fatalf("got=%v", got)
	}
}
