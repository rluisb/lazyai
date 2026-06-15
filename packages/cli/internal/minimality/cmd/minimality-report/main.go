package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rluisb/lazyai/packages/cli/internal/minimality"
)

func main() {
	root := flag.String("root", ".", "repository root to inspect")
	flag.Parse()

	if err := minimality.Run(*root, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
