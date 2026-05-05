package cmd

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Skip headless init in tests — avoids shelling out to real CLI binaries
	// (claude, opencode, copilot) which are slow and may not be installed.
	os.Setenv("AI_SETUP_SKIP_HEADLESS_INIT", "1")
	os.Exit(m.Run())
}