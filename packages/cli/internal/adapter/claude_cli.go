package adapter

import (
	"bytes"
	"context"
	"os/exec"
)

// ClaudeCLIRunner abstracts running `claude` commands for testability.
type ClaudeCLIRunner interface {
	Run(ctx context.Context, workingDir string, args ...string) (stdout, stderr []byte, err error)
}

// DefaultClaudeCLIRunner is the production implementation of ClaudeCLIRunner.
type DefaultClaudeCLIRunner struct{}

// Run executes a `claude` command with the given arguments, capturing output.
// If workingDir is not empty, the command runs in that directory.
// Returns (stdout, stderr, err). If the command exits non-zero, err is set.
func (d *DefaultClaudeCLIRunner) Run(ctx context.Context, workingDir string, args ...string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, "claude", args...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	return outBuf.Bytes(), errBuf.Bytes(), err
}

// LookupClaudeBinary checks whether the `claude` binary is available on PATH.
// Returns (path, true) if found, ("", false) if not.
func LookupClaudeBinary() (string, bool) {
	path, err := exec.LookPath("claude")
	if err != nil {
		return "", false
	}
	return path, true
}
