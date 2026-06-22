// Package adapter provides the OpenCode adapter implementation.
// Ported from the TypeScript OpenCodeAdapter.
package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/configmerge"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// OpenCodeConfigFilename is the canonical root opencode config filename.
const OpenCodeConfigFilename = "opencode.json"

// OpenCodeLegacyMCPFilename keeps compatibility with the old LazyAI MCP-only output
// location (`.opencode/lazyai.mcp.jsonc`) while migration is active.
const OpenCodeLegacyMCPFilename = "lazyai.mcp.jsonc"

// OpenCodeAdapter implements ToolAdapter for OpenCode (opencode CLI).
type OpenCodeAdapter struct{}

func (a *OpenCodeAdapter) ID() types.ToolId  { return types.ToolIdOpenCode }
func (a *OpenCodeAdapter) Name() string      { return "OpenCode" }
func (a *OpenCodeAdapter) ConfigDir() string { return ".opencode" }

func (a *OpenCodeAdapter) Install(ctx *AdapterContext) ([]types.TrackedFile, error) {
	ocDir, err := ResolveToolRoot(types.ToolIdOpenCode, ctx.SetupScope, ctx)
	if err != nil {
		return nil, err
	}

	_ = files.EnsureDir(ocDir)
	_ = files.EnsureDir(filepath.Join(ocDir, "agents"))
	_ = files.EnsureDir(filepath.Join(ocDir, "skills"))
	_ = files.EnsureDir(filepath.Join(ocDir, "commands"))
	_ = files.EnsureDir(filepath.Join(ocDir, "modes"))
	_ = files.EnsureDir(filepath.Join(ocDir, "plugins"))

	adapterLog.Info("installing tools", "adapter", "opencode")

	configRoot, err := openCodeConfigRoot(ctx)
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(configRoot, OpenCodeConfigFilename)

	// Install the default config only when no config is present. Skipping the
	// merge on pre-existing configs preserves user customizations (e.g., a
	// hand-tuned permissions object) across re-runs.
	if !files.FileExists(configPath) {
		instructions := openCodeInstructionsPath(ctx)
		defaultConfig := map[string]any{
			"$schema": "https://opencode.ai/config.json",
			"share":   "disabled",
			"instructions": []any{
				instructions,
			},
			"skills": map[string]any{
				"paths": []any{".opencode/skills"},
			},
			"permission": map[string]any{
				"skill": map[string]any{
					"*": "allow",
				},
			},
			"agent": map[string]any{
				"plan": map[string]any{
					"permission": map[string]any{
						"edit": "deny",
						"bash": "ask",
						"skill": map[string]any{
							"*": "allow",
						},
					},
				},
				"build": map[string]any{
					"permission": map[string]any{
						"edit": "ask",
						"bash": "ask",
						"skill": map[string]any{
							"*": "allow",
						},
					},
				},
				"explore": map[string]any{
					"permission": map[string]any{
						"edit": "deny",
						"bash": "deny",
						"skill": map[string]any{
							"*": "allow",
						},
					},
				},
			},
		}
		if _, err := configmerge.MergeJSONFile(configPath, defaultConfig); err != nil {
			return nil, err
		}
		hash, _ := files.FileHash(configPath)
		ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
			Path:   configPath,
			Hash:   hash,
			Source: "generated",
			Owner:  types.FileOwnerLibrary,
		})
	}

	if err := copyCanonicalDefaultAgent(ctx,
		filepath.Join(ocDir, "agents", defaultAgentID+".md"),
		openCodeDefaultAgentContent,
	); err != nil {
		return nil, err
	}

	// Copy canonical baseline-facing agents. The default guide is sourced above
	// only to guarantee the default file exists even with narrow selections.
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "canonical/agents",
		SelectionKey: "agents",
		ToDestPath: func(file string) string {
			return filepath.Join(ocDir, "agents", file)
		},
		WarnOnSkip: true,
		Transform: func(content []byte) []byte {
			out, err := RewriteAgentForOpenCode(content, ctx, "")
			if err != nil {
				adapterLog.Warn("opencode agent rewrite fell back to baseline frontmatter", "adapter", "opencode", "error", err)
				return BuildOpenCodeAgentFrontmatter(content, OpenCodeAgentOpts{})
			}
			return out
		},
		IncludeFile: func(file string) bool {
			return !isDefaultAgentFile(file) && isCanonicalAgentFile(file)
		},
	}); err != nil {
		return nil, err
	}

	// Copy skills (directory-per-skill layout).
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "skills",
		SelectionKey: "skills",
		ToDestPath: func(file string) string {
			name := fileID(file)
			return filepath.Join(ocDir, "skills", name, "SKILL.md")
		},
		WarnOnSkip: true,
	}); err != nil {
		return nil, err
	}

	// Copy opencode-native slash commands from library/opencode/commands/.
	// SelectionKey "opencodeCommands" honors per-user filtering once the
	// wizard populates ctx.Selections.OpenCodeCommands; with an unset
	// selection, all starter commands install.
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "opencode/commands",
		SelectionKey: "opencodeCommands",
		ToDestPath: func(file string) string {
			return filepath.Join(ocDir, "commands", file)
		},
		WarnOnSkip: true,
	}); err != nil {
		return nil, err
	}

	// Copy opencode chat modes from library/opencode/modes/.

	// Copy the managed OpenCode hook runtime plugin.
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "opencode/plugins",
		ToDestPath: func(file string) string {
			return filepath.Join(ocDir, "plugins", file)
		},
	}); err != nil {
		return nil, err
	}
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "opencode/modes",
		SelectionKey: "opencodeModes",
		ToDestPath: func(file string) string {
			return filepath.Join(ocDir, "modes", file)
		},
		WarnOnSkip: true,
	}); err != nil {
		return nil, err
	}

	if err := normalizeOpenCodePackageJSON(ocDir); err != nil {
		adapterLog.Warn("package.json normalization failed", "adapter", "opencode", "error", err)
	}

	// Install selected plugins via the opencode CLI if the binary is present.
	if err := installOpenCodePlugins(ctx, defaultCmdRunner); err != nil {
		adapterLog.Warn("plugin install failed", "adapter", "opencode", "error", err)
	}
	if err := normalizeOpenCodePackageJSON(ocDir); err != nil {
		adapterLog.Warn("package.json normalization failed", "adapter", "opencode", "error", err)
	}

	return ctx.FileRecords, nil
}
func openCodeConfigRoot(ctx *AdapterContext) (string, error) {
	switch ctx.SetupScope {
	case types.SetupScopeProject:
		return ctx.TargetDir, nil
	case types.SetupScopeWorkspace:
		if ctx.WorkspaceRoot != "" {
			return ctx.WorkspaceRoot, nil
		}
		return ctx.TargetDir, nil
	case types.SetupScopeGlobal:
		return ResolveToolRoot(types.ToolIdOpenCode, ctx.SetupScope, ctx)
	default:
		return "", fmt.Errorf("unsupported opencode config scope %q", ctx.SetupScope)
	}
}

func openCodeInstructionsPath(_ *AdapterContext) string {
	return "AGENTS.md"
}

func (a *OpenCodeAdapter) CompileMCP(ctx CompileContext) ([]types.TrackedFile, error) {
	return CompileMCPForTool(types.ToolIdOpenCode, ctx)
}

func (a *OpenCodeAdapter) CanRunHeadless() bool { return true }

func (a *OpenCodeAdapter) RunHeadlessInit(ctx *AdapterContext, prompt string) error {
	_, err := exec.LookPath("opencode")
	if err != nil {
		adapterLog.Info("opencode not on PATH, skipping headless init", "adapter", "opencode")
		return nil
	}

	adapterLog.Info("running headless init", "adapter", "opencode", "command", "opencode run")
	execCtx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "opencode",
		"run", prompt,
		"--dangerously-skip-permissions",
		"--format", "json",
	)
	cmd.Dir = ctx.TargetDir
	cmd.Stdin = nil

	output, err := cmd.CombinedOutput()
	if err != nil {
		adapterLog.Warn("headless init completed with warning", "adapter", "opencode", "error", err)
		if len(output) > 0 {
			adapterLog.Info("headless init output", "adapter", "opencode", "output", truncateOutput(string(output), 200))
		}
		return nil // non-fatal
	}

	adapterLog.Info("headless init completed", "adapter", "opencode", "bytes", len(output))
	return nil
}

func (a *OpenCodeAdapter) RunHeadlessValidation(ctx *AdapterContext) error { return nil }

// fileID extracts the filename without extension.
func fileID(filename string) string {
	ext := filepath.Ext(filename)
	return filename[:len(filename)-len(ext)]
}

// installOpenCodePlugins shells out to `opencode plugin <module>` for each
// selected plugin. Requires the binary to be on PATH; no-ops silently otherwise.
// Global scope passes -g; project/workspace scopes use the target dir as cwd.
// Each failure is logged and skipped — plugin errors do not block the install.
func installOpenCodePlugins(ctx *AdapterContext, run CmdRunner) error {
	if len(ctx.Selections.OpenCodePlugins) == 0 {
		return nil
	}
	if _, err := exec.LookPath("opencode"); err != nil {
		return nil
	}

	for _, module := range ctx.Selections.OpenCodePlugins {
		var args []string
		if ctx.SetupScope == types.SetupScopeGlobal {
			args = []string{"plugin", module, "-g"}
		} else {
			args = []string{"plugin", module}
		}
		if out, err := run("opencode", args...); err != nil {
			adapterLog.Warn("opencode plugin failed", "adapter", "opencode", "plugin", module, "error", err, "output", out)
		}
	}
	return nil
}

func normalizeOpenCodePackageJSON(ocDir string) error {
	packagePath := filepath.Join(ocDir, "package.json")
	data, err := os.ReadFile(packagePath)
	pkg := map[string]any{}
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("read opencode package.json: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, &pkg); err != nil {
			return fmt.Errorf("parse opencode package.json: %w", err)
		}
	}
	if current, ok := pkg["type"].(string); ok && current == "module" {
		return nil
	}
	pkg["type"] = "module"
	out, err := MarshalJSON(pkg)
	if err != nil {
		return fmt.Errorf("marshal opencode package.json: %w", err)
	}
	out = append(out, '\n')
	if err := os.WriteFile(packagePath, out, 0o644); err != nil {
		return fmt.Errorf("write opencode package.json: %w", err)
	}
	return nil
}

// MarshalJSON marshals v to indented JSON bytes.
func MarshalJSON(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}
