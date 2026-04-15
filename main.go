package main

import (
	"context"
	"os"

	"github.com/ricardoborges-teachable/ai-setup/cmd"
)

func main() {
	if err := cmd.Execute(context.Background()); err != nil {
		os.Exit(1)
	}
}
