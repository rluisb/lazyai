//go:build windows

package main

import (
	"os"
	"os/exec"
)

func configureDetachedCommand(cmd *exec.Cmd) {
	// Windows does not support Unix process groups. The daemon is still started
	// as a child process and released after Start; failure cleanup uses Kill.
}

func terminateProcess(pid int) {
	if pid <= 0 {
		return
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return
	}
	_ = process.Kill()
}

func processExists(pid int) bool {
	if pid <= 0 {
		return false
	}
	_, err := os.FindProcess(pid)
	return err == nil
}
