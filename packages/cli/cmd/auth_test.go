package cmd

import (
	"os"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/db"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestAuthListCmd(t *testing.T) {
	// Change working directory to a temporary directory for the test
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	// Create a dummy store data with some providers
	dbPath := db.DefaultDBPath(tmpDir)
	database, err := db.Open(dbPath)
	assert.NoError(t, err)
	err = db.RunMigrations(database)
	assert.NoError(t, err)

	// Create a dummy cmd
	cmd := &cobra.Command{Use: "test"}

	// We need to temporarily replace os.Stdout or just let runAuthList print
	// Since runAuthList uses fmt.Println directly, we can't easily capture it without os.Stdout redirection,
	// but we can at least ensure the command runs without error.

	err = runAuthList(cmd, []string{})
	assert.NoError(t, err, "runAuthList should not return an error")

	// Set an env var and run again
	os.Setenv("OPENAI_API_KEY", "dummy-key-12345")
	defer os.Unsetenv("OPENAI_API_KEY")

	err = runAuthList(cmd, []string{})
	assert.NoError(t, err, "runAuthList should not return an error when env var is set")
}
