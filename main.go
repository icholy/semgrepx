package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	// parse flags
	var dir string
	var trim, lines bool
	flag.StringVar(&dir, "dir", ".", "directory to run in")
	flag.BoolVar(&trim, "tri,", false, "trim whitespace")
	flag.BoolVar(&lines, "lines", false, "expand matches to full lines")
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatalf("expecting a command to run")
	}
	// read semgrep json
	output, err := ReadOutput(os.Stdin)
	if err != nil {
		log.Fatalf("failed to read semgrep json: %v", err)
	}
	err = RewriteAll(dir, output.Results, func(r Result, data []byte) (Result, []byte, error) {
		if lines {
			r = ExtendLines(r, data)
		}
		match := data[r.Start.Offset:r.End.Offset]
		fmt.Printf("--- before: %s\n%s\n",
			r.Path,
			FormatLines(string(match), r.Start.Line, 5),
		)
		var stdout bytes.Buffer
		cmd := exec.Command(flag.Arg(0), flag.Args()[1:]...)
		cmd.Stdin = bytes.NewReader(match)
		cmd.Stdout = &stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return r, nil, err
		}
		output := stdout.Bytes()
		if trim {
			output = bytes.TrimSpace(output)
		}
		fmt.Printf("--- after: %s\n%s\n",
			r.Path,
			FormatLines(string(output), r.Start.Line, 5),
		)
		return r, output, nil
	})
	if err != nil {
		log.Fatalf("failed to rewrite: %v", err)
	}
}

// ExtendLines returns r with the Start and End extended to include
// the full line content
func ExtendLines(r Result, data []byte) Result {
	if len(data) == 0 {
		return r
	}
	isNL := func(b byte) bool { return b == '\n' || b == '\r' }
	if !isNL(data[r.Start.Offset]) {
		for r.Start.Offset > 0 && !isNL(data[r.Start.Offset-1]) {
			r.Start.Offset--
			r.Start.Col--
		}
	}
	if !isNL(data[r.End.Offset]) {
		for r.End.Offset < len(data) && !isNL(data[r.End.Offset]) {
			r.End.Offset++
			r.End.Col++
		}
	}
	return r
}

func FormatLines(code string, lineno, indent int) string {
	var b strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(code))
	for scanner.Scan() {
		num := strconv.Itoa(lineno)
		b.WriteString(strings.Repeat(" ", indent-len(num)))
		b.WriteString(num)
		b.WriteString("| ")
		b.WriteString(scanner.Text())
		b.WriteByte('\n')
		lineno++
	}
	if scanner.Err() != nil {
		panic("unreachable")
	}
	return b.String()
}
