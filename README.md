# SEMGREPX

> A tool for rewriting semgrep matches using externals tools

Example:

``` sh
semgrep -l go --pattern 'log.$A(...)' --json | semgrepx llm 'update this go to use log.Printf'
```
