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
	var file, dir string
	var trim bool
	flag.StringVar(&file, "f", "semgrep.json", "semgrep json file")
	flag.StringVar(&dir, "d", ".", "directory to run in")
	flag.BoolVar(&trim, "t", false, "trim whitespace")
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatalf("expecting a command to run")
	}
	// read semgrep file
	output, err := ReadFile(file)
	if err != nil {
		log.Fatalf("failed to read semgrep json: %v", err)
	}
	err = RewriteAll(dir, output.Results, func(data []byte, r Result) ([]byte, error) {
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
