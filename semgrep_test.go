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
		func(r Result, data []byte) (Result, []byte, error) {
			return r, []byte("Good()"), nil
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
		result     Result
		data, want []byte
	}{
		{
			data: []byte(""),
			want: []byte(""),
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
		},
		{
			data: []byte(" a "),
			want: []byte(" a "),
			result: Result{
				Start: Pos{
					Line:   1,
					Offset: 1,
					Col:    1,
				},
				End: Pos{
					Line:   1,
					Col:    1,
					Offset: 1,
				},
			},
		},
		{
			data: []byte("foo();\nbar();\nbaz++"),
			want: []byte("bar();"),
			result: Result{
				Start: Pos{
					Line:   2,
					Col:    9,
					Offset: 9,
				},
				End: Pos{
					Line:   1,
					Col:    9,
					Offset: 9,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			r := ExtendLines(tt.result, tt.data)
			got := tt.data[r.Start.Offset:r.End.Offset]
			if !bytes.Equal(got, tt.want) {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
