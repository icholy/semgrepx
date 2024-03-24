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
	err = RewriteAll(dir, output.Results, func(data []byte, r Result) ([]byte, error) {
		if lines {
			r = ExtendLines(r, data)
		}
		fmt.Printf("--- before: %s\n%s\n",
			r.Path,
			FormatLines(string(data), r.Start.Line, 5),
		)
		var stdout bytes.Buffer
		cmd := exec.Command(flag.Arg(0), flag.Args()[1:]...)
		cmd.Stdin = bytes.NewReader(data)
		cmd.Stdout = &stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return nil, err
		}
		output := stdout.Bytes()
		if trim {
			output = bytes.TrimSpace(output)
		}
		fmt.Printf("--- after: %s\n%s\n",
			r.Path,
			FormatLines(string(output), r.Start.Line, 5),
		)
		return output, nil
	})
	if err != nil {
		log.Fatalf("failed to rewrite: %v", err)
	}
}

// ExtendLines returns r with the Start and End extended to include
// the full ine content
func ExtendLines(r Result, data []byte) Result {
	isNewline := func(b byte) bool { return b == '\n' || b == '\r' }
	for r.Start.Offset > 0 && !isNewline(data[r.Start.Offset]) {
		r.Start.Offset--
		r.Start.Col--
	}
	for r.End.Offset < len(data) && !isNewline(data[r.End.Offset]) {
		r.End.Offset++
		r.End.Col++
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
