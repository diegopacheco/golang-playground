package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/diegopacheco/tui-chess/internal/ui"
)

const version = "1.0.0"

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	captureDir := flag.String("capture", "", "render HTML screen captures into this directory and exit")
	flag.Parse()
	var err error
	switch {
	case *showVersion:
		fmt.Println("tchess", version)
	case *captureDir != "":
		err = ui.Capture(*captureDir)
	default:
		err = ui.Run()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "tchess:", err)
		os.Exit(1)
	}
}
