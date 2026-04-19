// Package adapter provides the Codex adapter implementation.
// Ported from the TypeScript CodexAdapter.
package adapter

import (
	"context"
	"log"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ricardoborges-teachable/ai-setup/internal/configmerge"
	"github.com/ricardoborges-teachable/ai-setup/internal/detect"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// CodexAdapter implements ToolAdapter for OpenAI Codex CLI.
//
// Codex uses two roots per upstream convention:
//   - configRoot (.codex/ or ~/.codex/): holds config.toml with
//     [mcp_servers.*] tables.
//   - skillsRoot (.agents/skills/ or ~/.agents/skills/): holds per-skill
//     <name>/SKILL.md. Skills live outside .codex/ in Codex's model.
//
// The root AGENTS.md is placed by scaffold/root.go (scope-aware).
type CodexAdapter struct{}

func (a *CodexAdapter) ID() types.ToolId  { return types.ToolIdCodex }
func (a *CodexAdapter) Name() string      { return "Codex CLI" }
func (a *CodexAdapter) ConfigDir() string { return ".codex" }

func (a *CodexAdapter) Install(ctx *AdapterContext) ([]types.TrackedFile, error) {
	// Check if Codex is installed; print a non-fatal warning if not.
	_ = detect.EnsureCodexOrPrompt()

	configRoot, skillsRoot, err := ResolveCodexRoots(ctx.SetupScope, ctx)
	if err != nil {
		return nil, err
	}

	_ = files.EnsureDir(configRoot)
	_ = files.EnsureDir(skillsRoot)

	log.Println("Installing Codex tools...")

	// Emit a minimal config.toml with an empty [mcp_servers] table so Codex
	// recognises the project/global config as trusted. User-authored tables
	// survive via configmerge.
	configPath := filepath.Join(configRoot, "config.toml")
	configPatch := map[string]any{
		"mcp_servers": map[string]any{},
	}
	preExisted := files.FileExists(configPath)
	if _, err := configmerge.MergeTOMLFile(configPath, configPatch); err != nil {
		return nil, err
	}
	if !preExisted {
		relPath, _ := filepath.Rel(ctx.TargetDir, configPath)
		if relPath == "" || relPath == "." {
			relPath = configPath
		}
		hash, _ := files.FileHash(configPath)
		ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
			Path: relPath, Hash: hash, Source: "generated", Owner: types.FileOwnerLibrary,
		})
	}

	// Codex uses skills in directory format at the skills root.
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "skills",
		SelectionKey: "skills",
		ToDestPath: func(file string) string {
			name := fileID(file)
			return filepath.Join(skillsRoot, name, "SKILL.md")
		},
	}); err != nil {
		return nil, err
	}

	// Orchestrator as a skill if enabled.
	if IsOrchestratorEnabled(ctx) {
		content := GetOrchestratorSkillContent(ctx)
		if err := WriteContentWithRecord(
			filepath.Join(skillsRoot, "orchestrator", "SKILL.md"),
			content, ctx, "generated:orchestrator-skill", false,
		); err != nil {
			return nil, err
		}
	}

	// Root AGENTS.md placement is handled centrally by scaffold/root.go.

	// CLI-driven MCP registration: when DriveCLI=true, attempt to register MCP
	// servers via `codex mcp add`. Falls back silently to direct-write TOML
	// (already merged above) on any failure.
	if ctx.DriveCLI {
		if ok := installCodexMCPViaCLI(ctx); ok {
			log.Println("[codex] MCP servers registered via CLI")
		}
	}

	// Global scope: emit AGENTS.override.md in the config root so users have a
	// ready-to-edit override file. Only written on first install; never overwrites.
	if ctx.SetupScope == types.SetupScopeGlobal {
		overridePath := filepath.Join(configRoot, "AGENTS.override.md")
		if !files.FileExists(overridePath) {
			const overrideContent = "# AGENTS Override\n\n" +
				"Add custom global instructions here. Codex reads this file at startup\n" +
				"and merges it with the project-level AGENTS.md.\n"
			if err := files.WriteFile(overridePath, []byte(overrideContent), 0o644); err != nil {
				return nil, err
			}
			relPath, _ := filepath.Rel(ctx.TargetDir, overridePath)
			if relPath == "" || relPath == "." {
				relPath = overridePath
			}
			hash, _ := files.FileHash(overridePath)
			ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
				Path: relPath, Hash: hash, Source: "generated:codex-override", Owner: types.FileOwnerLibrary,
			})
		}
	}

	return ctx.FileRecords, nil
}

func (a *CodexAdapter) CompileMCP(ctx CompileContext) ([]types.TrackedFile, error) {
	return CompileMCPForTool(types.ToolIdCodex, ctx)
}

func (a *CodexAdapter) CanRunHeadless() bool { return true }

func (a *CodexAdapter) RunHeadlessValidation(ctx *AdapterContext) error {
	_, err := exec.LookPath("codex")
	if err != nil {
		log.Printf("[codex] codex not on PATH, skipping headless validation")
		return nil
	}

	log.Printf("[codex] running headless validation: codex exec \"check .agents/ structure\"")
	execCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "codex", "exec", "check .agents/ structure")
	cmd.Dir = ctx.TargetDir
	cmd.Stdin = nil // pipe /dev/null equivalent — no interactive input

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[codex] headless validation completed with warning: %v", err)
		if len(output) > 0 {
			log.Printf("[codex] output: %s", string(output))
		}
		return nil // non-fatal
	}

	log.Printf("[codex] headless validation succeeded")
	return nil
}

// installCodexMCPViaCLI calls `codex mcp add <name> [--env K=V]... -- <cmd> <args>...`
// for each enabled MCP server in the canonical .ai/mcp.json. Returns true only
// if at least one server was registered successfully. The direct-write TOML
// merge already happened earlier in Install, so returning false is safe.
//
// Codex's CLI format (verified from the upstream source): positional
// COMMAND [ARGS...] follow a literal `--` separator. Env vars are repeated
// `--env KEY=VALUE` flags.
func installCodexMCPViaCLI(ctx *AdapterContext) bool {
	bin, err := exec.LookPath("codex")
	if err != nil {
		log.Println("[codex] --drive-cli requested but codex binary not found; using direct-write")
		return false
	}

	catalog := readCanonicalMcp(ctx.TargetDir)
	if catalog == nil || len(catalog.Servers) == 0 {
		return false
	}

	success := false
	for name, srv := range catalog.Servers {
		if !srv.isEnabled() {
			continue
		}
		// Remote servers not supported by this helper — skip with a log.
		if srv.URL != "" {
			log.Printf("[codex] skipping remote server %q (use --url flow manually)", name)
			continue
		}
		if srv.Command == "" {
			log.Printf("[codex] skipping server %q: no command", name)
			continue
		}

		args := []string{"mcp", "add", name}
		for k, v := range srv.Env {
			args = append(args, "--env", k+"="+v)
		}
		args = append(args, "--", srv.Command)
		args = append(args, srv.Args...)

		runCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		out, runErr := exec.CommandContext(runCtx, bin, args...).CombinedOutput()
		cancel()
		if runErr != nil {
			log.Printf("[codex] mcp add %q failed: %v\n%s", name, runErr, string(out))
			continue
		}
		log.Printf("[codex] registered MCP server %q via CLI", name)
		success = true
	}
	return success
}
