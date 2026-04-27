package detect

import (
	"fmt"
	"os/exec"
	"runtime"
)

// IsCodexInstalled checks whether the `codex` binary is available on $PATH.
func IsCodexInstalled() bool {
	_, err := exec.LookPath("codex")
	return err == nil
}

// CodexInstallHint returns platform-specific instructions for installing Codex CLI.
func CodexInstallHint() string {
	switch runtime.GOOS {
	case "darwin":
		return "Install Codex CLI via Homebrew:\n  brew install codex\n\nOr via npm:\n  npm install -g @openai/codex"
	case "linux":
		return "Install Codex CLI via npm:\n  npm install -g @openai/codex\n\nOr download from https://github.com/openai/codex"
	case "windows":
		return "Install Codex CLI via npm:\n  npm install -g @openai/codex\n\nOr download from https://github.com/openai/codex"
	default:
		return "Install Codex CLI via npm:\n  npm install -g @openai/codex"
	}
}

// EnsureCodexOrPrompt checks if Codex is installed. If not, it prints an
// install hint and returns nil (non-fatal). Callers should treat this as
// a warning, not an error.
func EnsureCodexOrPrompt() error {
	if IsCodexInstalled() {
		return nil
	}

	hint := CodexInstallHint()
	fmt.Printf("\n⚠️  Codex CLI is not installed on your system.\n%s\n\n", hint)
	return nil
}
