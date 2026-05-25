// Package adapter provides the OpenCode adapter implementation.
// Ported from the TypeScript OpenCodeAdapter.
package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/configmerge"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// OpenCodeConfigFilename is the canonical opencode config filename. Both the
// install-time default-config writer and the compile-time MCP writer target
// this single file so users never end up with both opencode.json and
// opencode.jsonc side-by-side.
const OpenCodeConfigFilename = "opencode.jsonc"

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

	adapterLog.Info("installing tools", "adapter", "opencode")

	jsonPath := filepath.Join(ocDir, "opencode.json")
	jsoncPath := filepath.Join(ocDir, OpenCodeConfigFilename)

	// One-shot migration: collapse any pre-existing opencode.json onto
	// opencode.jsonc. The original .json is preserved as a .bak sidecar so
	// users can recover if the rename surprises them. If both files exist
	// (rare), .jsonc is authoritative.
	if files.FileExists(jsonPath) {
		bakPath := jsonPath + ".bak"
		if !files.FileExists(bakPath) {
			if err := files.CopyFile(jsonPath, bakPath); err != nil {
				return nil, fmt.Errorf("backup opencode.json: %w", err)
			}
		}
		if !files.FileExists(jsoncPath) {
			if err := files.CopyFile(jsonPath, jsoncPath); err != nil {
				return nil, fmt.Errorf("migrate opencode.json to opencode.jsonc: %w", err)
			}
		}
		if err := os.Remove(jsonPath); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("remove opencode.json after migration: %w", err)
		}
	}

	// Install the default config only when no config is present. Skipping the
	// merge on pre-existing configs preserves user customizations (e.g., a
	// hand-tuned `permission.edit` value) across re-runs.
	if !files.FileExists(jsoncPath) {
		instructions := []any{openCodeInstructionsPath(ctx)}
		defaultAgent := "orchestrator"
		if ctx.FortniteMode {
			instructions = []any{"AGENTS.md", "STARTUP.md"}
			defaultAgent = "loop-driver"
		}
		defaultConfig := map[string]any{
			"$schema":       "https://opencode.ai/config.json",
			"default_agent": defaultAgent,
			"instructions":  instructions,
			"permission": map[string]any{
				"edit": "ask",
				"bash": "ask",
			},
		}
		if _, err := configmerge.MergeJSONFile(jsoncPath, defaultConfig); err != nil {
			return nil, err
		}
		hash, _ := files.FileHash(jsoncPath)
		ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
			Path:   jsoncPath,
			Hash:   hash,
			Source: "generated",
			Owner:  types.FileOwnerLibrary,
		})
	}

	if ctx.FortniteMode {
		// --- Fortnite mode: install Fortnite agents, skills, scripts, workflows ---

		// Merge Fortnite AGENTS.md into project root via managed block.
		// This preserves any existing root AGENTS.md content while ensuring
		// the Fortnite agent registry is present.
		rootAgentsPath := filepath.Join(ctx.TargetDir, "AGENTS.md")
		var fortniteAgentsContent []byte
		if ctx.LibraryFS != nil {
			data, err := fs.ReadFile(ctx.LibraryFS, "fortnite/AGENTS.md")
			if err != nil {
				return nil, fmt.Errorf("read fortnite AGENTS.md: %w", err)
			}
			fortniteAgentsContent = data
		} else {
			data, err := files.ReadFile(filepath.Join(ctx.LibraryDir, "fortnite", "AGENTS.md"))
			if err != nil {
				return nil, fmt.Errorf("read fortnite AGENTS.md: %w", err)
			}
			fortniteAgentsContent = data
		}

		_, modified, err := WriteManagedBlockToFile(
			rootAgentsPath,
			fortniteAgentsContent,
			ManagedBlockStartMarker,
			ManagedBlockEndMarker,
		)
		if err != nil {
			return nil, fmt.Errorf("merge managed block into root AGENTS.md: %w", err)
		}
		if modified {
			relPath, _ := filepath.Rel(ctx.TargetDir, rootAgentsPath)
			if relPath == "" {
				relPath = rootAgentsPath
			}
			hash, _ := files.FileHash(rootAgentsPath)
			ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
				Path:   relPath,
				Hash:   hash,
				Source: "fortnite/AGENTS.md",
				Owner:  types.FileOwnerLibrary,
			})
		}

		// Copy Fortnite STARTUP.md to .opencode/STARTUP.md.
		if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
			Ctx:          ctx,
			SourceSubdir: "fortnite",
			SelectionKey: "fortniteStartup",
			ToDestPath: func(file string) string {
				return filepath.Join(ocDir, file)
			},
			IncludeFile: func(file string) bool {
				return file == "STARTUP.md"
			},
			WarnOnSkip: true,
		}); err != nil {
			return nil, err
		}

		// Recursively copy Fortnite scripts to .opencode/scripts/.
		if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
			Ctx:          ctx,
			SourceSubdir: "fortnite/scripts",
			SelectionKey: "fortniteScripts",
			ToDestPath: func(file string) string {
				return filepath.Join(ocDir, "scripts", file)
			},
			Recursive:   true,
			WarnOnSkip:  true,
		}); err != nil {
			return nil, err
		}
		if err := ChmodScriptsExecutable(filepath.Join(ocDir, "scripts")); err != nil {
			adapterLog.Warn("chmod scripts failed", "adapter", "opencode", "error", err)
		}

		// Recursively copy Fortnite workflows to .opencode/workflows/.
		if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
			Ctx:          ctx,
			SourceSubdir: "fortnite/workflows",
			SelectionKey: "fortniteWorkflows",
			ToDestPath: func(file string) string {
				return filepath.Join(ocDir, "workflows", file)
			},
			Recursive:   true,
			WarnOnSkip:  true,
		}); err != nil {
			return nil, err
		}

		// Copy Fortnite agents (flat).
		if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
			Ctx:          ctx,
			SourceSubdir: "fortnite/agents",
			SelectionKey: "fortniteAgents",
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
		}); err != nil {
			return nil, err
		}

		// Copy Fortnite skills (recursive, nested dirs).
		if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
			Ctx:          ctx,
			SourceSubdir: "fortnite/skills",
			SelectionKey: "fortniteSkills",
			ToDestPath: func(file string) string {
				return filepath.Join(ocDir, "skills", file)
			},
			Recursive:   true,
			WarnOnSkip:  true,
		}); err != nil {
			return nil, err
		}
	} else {
		// --- Legacy mode: copy generic agents and skills ---

		// Copy agents (excluding orchestrator unless enabled). The transform
		// rewrites each source agent with opencode-schema frontmatter so
		// `opencode debug agent <name>` can parse it.
		if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
			Ctx:          ctx,
			SourceSubdir: "agents",
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
				name := fileID(file)
				return name != "orchestrator"
			},
		}); err != nil {
			return nil, err
		}

		// Orchestrator agent if enabled. It is the primary entry point
		// (opencode's default_agent typically points here), so mode=primary.
		if IsOrchestratorEnabled(ctx) {
			raw := GetOrchestratorAgentContent(ctx)
			content, err := RewriteAgentForOpenCode(raw, ctx, "primary")
			if err != nil {
				adapterLog.Warn("opencode orchestrator rewrite fell back to baseline", "adapter", "opencode", "error", err)
				content = BuildOpenCodeAgentFrontmatter(raw, OpenCodeAgentOpts{Mode: "primary"})
			}
			if err := CopyWithRecord("agents/orchestrator.md",
				filepath.Join(ocDir, "agents", "orchestrator.md"),
				ctx, true,
				func([]byte) []byte { return content },
				0o644,
			); err != nil {
				return nil, err
			}
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

	// Install selected plugins via the opencode CLI if the binary is present.
	if err := installOpenCodePlugins(ctx, defaultCmdRunner); err != nil {
		adapterLog.Warn("plugin install failed", "adapter", "opencode", "error", err)
	}

	return ctx.FileRecords, nil
}

func openCodeInstructionsPath(ctx *AdapterContext) string {
	if ctx.SetupScope == types.SetupScopeGlobal {
		return "AGENTS.md"
	}
	return "../AGENTS.md"
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

// MarshalJSON marshals v to indented JSON bytes.
func MarshalJSON(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}
