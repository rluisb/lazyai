package adapter

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// CmdRunner is an injectable command executor for testing.
type CmdRunner func(name string, args ...string) ([]byte, error)

// ValidationWarning describes a single non-fatal issue found during post-install validation.
type ValidationWarning struct {
	Scope  string
	Item   string
	Reason string
}

func (w ValidationWarning) String() string {
	return fmt.Sprintf("[opencode validate] %s / %s: %s", w.Scope, w.Item, w.Reason)
}

// ValidateOpenCodeInstall runs post-install sanity checks via the opencode CLI.
// It is a no-op (returns nil, nil) when the binary is not on PATH.
// All failures are returned as ValidationWarning entries — the function never
// propagates CLI errors as hard errors so init/update continue to exit 0.
func ValidateOpenCodeInstall(ctx *AdapterContext) ([]ValidationWarning, error) {
	return validateOpenCodeInstall(ctx, DefaultCmdRunner)
}

// ValidateOpenCodeInstallWith runs validation with a custom command runner.
// Pass nil to use DefaultCmdRunner; callers that need to suppress external CLI
// probes (e.g. tests) should pass a no-op CmdRunner.
func ValidateOpenCodeInstallWith(ctx *AdapterContext, run CmdRunner) ([]ValidationWarning, error) {
	if run == nil {
		run = DefaultCmdRunner
	}
	return validateOpenCodeInstall(ctx, run)
}

func validateOpenCodeInstall(ctx *AdapterContext, run CmdRunner) ([]ValidationWarning, error) {
	if _, err := exec.LookPath("opencode"); err != nil {
		return nil, nil
	}

	ocDir, err := ResolveToolRoot(types.ToolIdOpenCode, ctx.SetupScope, ctx)
	if err != nil {
		return nil, nil
	}
	configRoot, err := openCodeConfigRoot(ctx)
	if err != nil {
		return nil, nil
	}

	var warnings []ValidationWarning

	// Validate config.
	if out, err := run("opencode", "debug", "config"); err != nil {
		warnings = append(warnings, ValidationWarning{
			Scope:  "config",
			Item:   "opencode.json",
			Reason: fmt.Sprintf("opencode debug config failed: %v", err),
		})
	} else if !strings.Contains(string(out), "mcp") && files.FileExists(filepath.Join(configRoot, OpenCodeConfigFilename)) {
		warnings = append(warnings, ValidationWarning{
			Scope:  "config",
			Item:   "opencode.json",
			Reason: "opencode debug config output does not mention mcp — MCP entries may not have been picked up",
		})
	}

	// Validate each installed agent.
	agentsDir := filepath.Join(ocDir, "agents")
	entries, err := filepath.Glob(filepath.Join(agentsDir, "*.md"))
	if err != nil {
		return warnings, nil
	}
	for _, agentPath := range entries {
		name := fileID(filepath.Base(agentPath))
		if _, err := run("opencode", "debug", "agent", name); err != nil {
			warnings = append(warnings, ValidationWarning{
				Scope:  "agent",
				Item:   name,
				Reason: fmt.Sprintf("opencode debug agent failed: %v", err),
			})
		}
	}

	return warnings, nil
}

// DefaultCmdRunner is the production command executor used when no override is
// supplied. It runs the command and returns its combined stdout/stderr.
var DefaultCmdRunner CmdRunner = func(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}
