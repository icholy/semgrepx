package main

import (
	"bytes"
	"flag"
	"log"
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
		var output bytes.Buffer
		cmd := exec.Command(flag.Arg(0), flag.Args()[1:]...)
		cmd.Stdin = strings.NewReader(r.Extra.Lines)
		cmd.Stdout = &output
		if err := cmd.Run(); err != nil {
			return nil, err
		}
		return output.Bytes(), nil
	})
	if err != nil {
		log.Fatalf("failed to rewrite: %v", err)
	}
}
