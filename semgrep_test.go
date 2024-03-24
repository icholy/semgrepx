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
