package main

import (
	"os"

	"github.com/ricardoborges-teachable/ai-setup/packages/diffviewer"
)

func main() {
	code := diffviewer.RunCLI(os.Stdin, os.Stdout, os.Stderr, func(views []diffviewer.ConflictView) diffviewer.Reviewer {
		return diffviewer.NewDiffViewer(views)
	})
	os.Exit(code)
}
