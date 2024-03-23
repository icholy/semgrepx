package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// parse flags
	var file, dir string
	flag.StringVar(&file, "f", "semgrep.json", "semgrep json file")
	flag.StringVar(&dir, "d", ".", "directory to run in")
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatalf("expecting a command to run")
	}
	// read semgrep file
	output, err := ReadFile(file)
	if err != nil {
		log.Fatalf("failed to read semgrep json: %v", err)
	}
	err = RewriteAll(dir, output.Results, func(r Result) ([]byte, error) {
		fmt.Printf("file: %s:%d\n--- before ---\n%s\n", r.Path, r.Start.Line, r.Extra.Lines)
		var output bytes.Buffer
		cmd := exec.Command(flag.Arg(0), flag.Args()[1:]...)
		cmd.Stdin = strings.NewReader(strings.TrimSpace(r.Extra.Lines))
		cmd.Stdout = &output
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return nil, err
		}
		fmt.Printf("--- after ---\n%s\n---\n\n", output.String())
		return output.Bytes(), nil
	})
	if err != nil {
		log.Fatalf("failed to rewrite: %v", err)
	}
}
