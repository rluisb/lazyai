package main

import (
	"fmt"
	"os"

	"github.com/rluisb/lazyai/packages/cli/internal/tokenrent"
)

func main() {
	result, err := tokenrent.CheckProject(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Printf("Canonical library budget OK: %d / %d bytes\n", result.TotalBytes, result.BudgetBytes)
}
