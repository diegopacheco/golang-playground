package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/diegopacheco/golang-playground/pocs/log-tailer-grep/internal/ansihtml"
)

func main() {
	title := flag.String("title", "log-tailer-grep", "html page title")
	flag.Parse()
	if err := ansihtml.Convert(os.Stdin, os.Stdout, *title); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}
