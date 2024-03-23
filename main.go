package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
)

type Pos struct {
	Col    int `json:"col"`
	Line   int `json:"line"`
	Offset int `json:"offset"`
}

type Result struct {
	CheckID string `json:"check_id"`
	End     Pos    `json:"end"`
	Extra   Extra  `json:"extra"`
	Path    string `json:"path"`
	Start   Pos    `json:"start"`
}

type Extra struct {
	EngineKind  string         `json:"engine_kind"`
	Fingerprint string         `json:"fingerprint"`
	IsIgnored   bool           `json:"is_ignored"`
	Lines       string         `json:"lines"`
	Message     string         `json:"message"`
	Metadata    map[string]any `json:"metadata"`
	Severity    string         `json:"severity"`
}

type Paths struct {
	Comment string   `json:"_comment"`
	Scanned []string `json:"scanned"`
}

type Output struct {
	Errors  []any    `json:"errors"`
	Paths   Paths    `json:"paths"`
	Results []Result `json:"results"`
	Version string   `json:"version"`
}

// ReadFile reads and parses a file containing semgrep json
func ReadFile(name string) (*Output, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	var output Output
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

type RewriteFn = func(r Result) ([]byte, error)

var ErrSkip = errors.New("skip")

func Rewrite(data []byte, results []Result, rewrite RewriteFn) ([]byte, error) {
	slices.SortFunc(results, func(a, b Result) int {
		return a.Start.Offset - b.Start.Offset
	})
	for _, r := range results {
		rewritten, err := rewrite(r)
		if err == ErrSkip {
			continue
		}
		if err != nil {
			return nil, err
		}
		data = slices.Replace(data, r.Start.Offset, r.End.Offset, rewritten...)
	}
	return data, nil
}

func RewriteAll(dir string, results []Result, rewrite RewriteFn) error {
	files := map[string][]Result{}
	for _, r := range results {
		files[r.Path] = append(files[r.Path], r)
	}
	for file, rr := range files {
		path := filepath.Join(dir, file)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		data, err = Rewrite(data, rr, rewrite)
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, data, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}
