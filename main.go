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
	var dir, file string
	var trim, lines bool
	var retry int
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: semgrepx [flags] <command> [args...]")
		fmt.Println("flags:")
		flag.PrintDefaults()
	}
	flag.StringVar(&file, "file", "", "semgrep json file")
	flag.StringVar(&dir, "dir", ".", "directory to run in")
	flag.BoolVar(&trim, "trim", false, "trim whitespace")
	flag.BoolVar(&lines, "lines", false, "expand matches to full lines")
	flag.IntVar(&retry, "retry", 0, "number of retries (< 0 is unlimited)")
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	// read semgrep json
	var output *Output
	if file != "" {
		var err error
		output, err = ReadOutputFile(file)
		if err != nil {
			log.Fatalf("failed to open semgrep json file: %v", err)
		}
	} else {
		var err error
		output, err = ReadOutput(os.Stdin)
		if err != nil {
			log.Fatalf("failed to read semgrep json: %v", err)
		}
	}
	err := RewriteAll(dir, output.Results, func(r Result, data []byte) (Result, []byte, error) {
		if lines {
			r = ExtendLines(r, data)
		}
		match := data[r.Start.Offset:r.End.Offset]
		fmt.Printf("--- before: %s\n%s\n",
			r.Path,
			FormatLines(match, r.Start.Line, 5),
		)
		var stdout bytes.Buffer
		for retries := 0; true; retries++ {
			stdout.Reset()
			cmd := exec.Command(flag.Arg(0), flag.Args()[1:]...)
			cmd.Stdin = bytes.NewReader(match)
			cmd.Stdout = &stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err == nil {
				break
			}
			if retry < 0 || retries <= retry {
				log.Printf("retrying: %v\n", err)
				continue
			}
			return r, nil, err
		}
		rewritten := stdout.Bytes()
		if trim {
			rewritten = bytes.TrimSpace(rewritten)
		}
		fmt.Printf("--- after: %s\n%s\n",
			r.Path,
			FormatLines(rewritten, r.Start.Line, 5),
		)
		return r, rewritten, nil
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

func FormatLines(data []byte, lineno, indent int) string {
	var b strings.Builder
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		num := strconv.Itoa(lineno)
		b.WriteString(strings.Repeat(" ", indent-len(num)))
		b.WriteString(num)
		b.WriteString("| ")
		b.Write(scanner.Bytes())
		b.WriteByte('\n')
		lineno++
	}
	if scanner.Err() != nil {
		panic("unreachable")
	}
	return b.String()
}
