package adapter

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

// CopilotCLIRunner abstracts running `copilot` commands for testability.
type CopilotCLIRunner interface {
	Run(ctx context.Context, workingDir string, args ...string) (stdout, stderr []byte, err error)
}

// DefaultCopilotCLIRunner is the production implementation of CopilotCLIRunner.
type DefaultCopilotCLIRunner struct{}

// Run executes a `copilot` command with the given arguments, capturing output.
// If workingDir is not empty, the command runs in that directory.
// Returns (stdout, stderr, err). If the command exits non-zero, err is set.
func (d *DefaultCopilotCLIRunner) Run(ctx context.Context, workingDir string, args ...string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, "copilot", args...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	return outBuf.Bytes(), errBuf.Bytes(), err
}

// LookupCopilotBinary checks whether the `copilot` binary is available on PATH.
// Returns (path, true) if found, ("", false) if not.
func LookupCopilotBinary() (string, bool) {
	path, err := exec.LookPath("copilot")
	if err != nil {
		return "", false
	}
	return path, true
}

// CopilotHomePresent checks whether the ~/.copilot/ directory exists.
// Returns true if present, false otherwise.
func CopilotHomePresent() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	copilotDir := filepath.Join(home, ".copilot")
	info, err := os.Stat(copilotDir)
	return err == nil && info.IsDir()
}
