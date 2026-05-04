package main

import (
	"context"
	"os"

	"github.com/rluisb/lazyai/packages/cli/cmd"
	"github.com/rluisb/lazyai/packages/cli/internal/library"
	libraryembed "github.com/rluisb/lazyai/packages/cli/library"
)

func init() {
	library.SetEmbeddedFS(libraryembed.FS)
}

func main() {
	if err := cmd.Execute(context.Background()); err != nil {
		os.Exit(1)
	}
}
