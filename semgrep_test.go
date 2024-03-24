package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRewrite(t *testing.T) {
	good, err := os.ReadFile(filepath.FromSlash("./testdata/good.go"))
	if err != nil {
		t.Fatalf("failed to read good file: %v", err)
	}
	bad, err := os.ReadFile(filepath.FromSlash("./testdata/bad.go"))
	if err != nil {
		t.Fatalf("failed to read bad file: %v", err)
	}
	output, err := ReadOutputFile(filepath.FromSlash("./testdata/semgrep.json"))
	if err != nil {
		t.Fatalf("failed to read semgrep json: %v", err)
	}
	rewritten, err := Rewrite(
		bad,
		output.Results,
		func(data []byte, r Result) ([]byte, error) {
			return []byte("Good()"), nil
		},
	)
	if err != nil {
		t.Fatalf("failed to rewrite: %v", err)
	}
	if !bytes.Equal(rewritten, good) {
		t.Fatalf("rewritten file does not match good file")
	}
}

func TestExtendLines(t *testing.T) {
	tests := []struct {
		data         []byte
		result, want Result
	}{
		{
			data: []byte(""),
			result: Result{
				Start: Pos{
					Line:   1,
					Col:    0,
					Offset: 0,
				},
				End: Pos{
					Line:   1,
					Col:    0,
					Offset: 0,
				},
			},
			want: Result{
				Start: Pos{
					Line:   1,
					Col:    0,
					Offset: 0,
				},
				End: Pos{
					Line:   1,
					Col:    0,
					Offset: 0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := ExtendLines(tt.result, tt.data)
			if got.Start != tt.want.Start {
				t.Fatalf("bad start position: want %v, got %v", tt.want.Start, got.Start)
			}
			if got.End != tt.want.End {
				t.Fatalf("bad end position: want %v, got %v", tt.want.End, got.End)
			}
		})
	}
}
