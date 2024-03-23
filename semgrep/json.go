package semgrep

import (
	"encoding/json"
	"os"
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

// Replace the the buf contents between start and end with data
func Replace(buf []byte, start, end Pos, data []byte) []byte {
	return slices.Replace(buf, start.Offset, end.Offset, data...)
}

type RewriteFn = func(r Result) (string, error)

func Rewrite(dir string, results []Result, rewrite RewriteFn) error {
	files := map[string][]Result{}
	for _, r := range results {
		files[r.Path] = append(files[r.Path], r)
	}
	for _, rr := range files {
		slices.SortFunc(rr, func(a, b Result) int {
			return a.Start.Offset - b.Start.Offset
		})
	}
	return nil
}
