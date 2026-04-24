package main

import (
	"context"
	"embed"
	"io/fs"
	"os"

	"github.com/ricardoborges-teachable/ai-setup/cmd"
	"github.com/ricardoborges-teachable/ai-setup/internal/library"
)

//go:embed all:library
var libraryEmbed embed.FS

func init() {
	// Strip the "library" prefix so files are accessed as "agents/builder.md" etc.
	fsys, err := fs.Sub(libraryEmbed, "library")
	if err != nil {
		// If embed fails, fallback to filesystem resolution in GetLibraryFS.
		return
	}
	library.SetEmbeddedFS(fsys)
}

func main() {
	if err := cmd.Execute(context.Background()); err != nil {
		os.Exit(1)
	}
}
