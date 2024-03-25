# SEMGREPX

> A tool for rewriting semgrep matches using externals tools

### CLI:

``` text
Usage: semgrepx [flags] <command> [args...]
flags:
  -dir string
    	directory to run in (default ".")
  -lines
    	expand matches to full lines
  -trim
    	trim whitespace
```

### How it works:

The provided command is executed for every semgrep match.
The matched code is sent to the command's stdin.
The matched code is replaced by the command's stdout.

### Example:

``` sh
# create a file of matches
semgrep -l go --pattern 'log.$A(...)' --json > matches.json

# rewrite all the matches using the llm tool
semgrepx llm 'update this go to use log.Printf' < matches.json
```
