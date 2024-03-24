package main

import (
	"encoding/json"
	"errors"
	"io"
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

func ReadOutput(r io.Reader) (*Output, error) {
	var output Output
	dec := json.NewDecoder(r)
	if err := dec.Decode(&output); err != nil {
		return nil, err
	}
	return &output, nil
}

func ReadOutputFile(filename string) (*Output, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadOutput(f)
}

type RewriteFn = func(data []byte, r Result) ([]byte, error)

var ErrSkip = errors.New("skip")

func Rewrite(data []byte, results []Result, rewrite RewriteFn) ([]byte, error) {
	slices.SortFunc(results, func(a, b Result) int {
		return b.Start.Offset - a.Start.Offset
	})
	for _, r := range results {
		rewritten, err := rewrite(data[r.Start.Offset:r.End.Offset], r)
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
