// Package adapter provides the Codex adapter implementation.
// Ported from the TypeScript CodexAdapter.
package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ricardoborges-teachable/ai-setup/internal/configmerge"
	"github.com/ricardoborges-teachable/ai-setup/internal/detect"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/library"
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

	// Emit AGENTS.override.md in the config root on first install so users
	// have a ready-to-edit override scaffold. Never overwrites existing
	// files (user-authored content always wins). Spec 018 moved the content
	// into library/codex/AGENTS.override.template.md for per-tool parity
	// and to make the starter richer than the previous three-line stub.
	if err := writeCodexAgentsOverride(ctx, configRoot); err != nil {
		return nil, err
	}

	// Post-install summary (non-fatal): log how many MCP servers Codex has
	// registered. Mirrors the Claude Code summary from spec 012. Skipped
	// when codex is not on PATH.
	displayCodexInstallSummary(ctx)

	return ctx.FileRecords, nil
}

// writeCodexAgentsOverride writes <configRoot>/AGENTS.override.md from the
// library template, only when the destination does not already exist.
// Applies at every scope (project/workspace/global), closing a spec-008/010
// asymmetry where only global scope received an override scaffold.
func writeCodexAgentsOverride(ctx *AdapterContext, configRoot string) error {
	overridePath := filepath.Join(configRoot, "AGENTS.override.md")
	if files.FileExists(overridePath) {
		return nil // user-authored content wins
	}

	content, err := readCodexAgentsOverrideTemplate(ctx)
	if err != nil {
		// Fall back to the historical inline stub so we don't hard-fail when
		// the library template is missing (e.g. in minimal test FSes that
		// don't ship library/codex/).
		content = []byte("# AGENTS Override\n\n" +
			"Add custom instructions here. Codex reads this file at startup\n" +
			"and merges it with the project-level AGENTS.md.\n")
	}
	if err := files.WriteFile(overridePath, content, 0o644); err != nil {
		return err
	}
	relPath, _ := filepath.Rel(ctx.TargetDir, overridePath)
	if relPath == "" || relPath == "." {
		relPath = overridePath
	}
	hash, _ := files.FileHash(overridePath)
	ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
		Path: relPath, Hash: hash, Source: "generated:codex-override", Owner: types.FileOwnerLibrary,
	})
	return nil
}

// readCodexAgentsOverrideTemplate reads the AGENTS.override starter template
// from the library FS, falling back to the on-disk path in dev mode.
func readCodexAgentsOverrideTemplate(ctx *AdapterContext) ([]byte, error) {
	if ctx.LibraryFS != nil {
		return fs.ReadFile(ctx.LibraryFS, library.CodexAgentsOverrideTemplate)
	}
	if ctx.LibraryDir != "" {
		return files.ReadFile(filepath.Join(ctx.LibraryDir, library.CodexAgentsOverrideTemplate))
	}
	return nil, fmt.Errorf("no library source available for Codex AGENTS.override template")
}

func (a *CodexAdapter) CompileMCP(ctx CompileContext) ([]types.TrackedFile, error) {
	return CompileMCPForTool(types.ToolIdCodex, ctx)
}

func (a *CodexAdapter) CanRunHeadless() bool { return true }

// codexExecValidationArgs returns the argv used by RunHeadlessValidation.
// Factored for testability — the install flow passes `--skip-git-repo-check`
// so the probe works against non-repo temp dirs (the previous invocation
// failed silently with "Not inside a trusted directory").
func codexExecValidationArgs() []string {
	return []string{"exec", "--skip-git-repo-check", "check .agents/ structure"}
}

func (a *CodexAdapter) RunHeadlessValidation(ctx *AdapterContext) error {
	_, err := exec.LookPath("codex")
	if err != nil {
		log.Printf("[codex] codex not on PATH, skipping headless validation")
		return nil
	}

	args := codexExecValidationArgs()
	log.Printf("[codex] running headless validation: codex %v", args)
	execCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "codex", args...)
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

// displayCodexInstallSummary prints a one-line post-install summary of MCP
// servers registered with Codex. Mirrors the pattern used by
// displayInstallSummary for Claude Code (spec 012). Non-fatal on any
// failure — the install already succeeded before this runs.
func displayCodexInstallSummary(ctx *AdapterContext) {
	bin, err := exec.LookPath("codex")
	if err != nil {
		log.Println("[codex] codex CLI not on PATH; skipping post-install summary")
		return
	}

	count, ok := codexMcpServerCount(bin, ctx.TargetDir)
	if !ok {
		log.Println("[codex] could not determine MCP server count; post-install summary skipped")
		return
	}

	scopeLabel := "project"
	switch ctx.SetupScope {
	case types.SetupScopeGlobal:
		scopeLabel = "user"
	case types.SetupScopeWorkspace:
		scopeLabel = "workspace"
	}

	log.Printf("[codex] Install summary (scope: %s)", scopeLabel)
	log.Printf("  • %d MCP server(s) registered", count)
}

// codexMcpServerCount asks the codex binary for a server count. Tries the
// JSON form first; falls back to parsing plaintext line output if JSON is
// unavailable or fails to parse. Returns (count, true) on success or
// (0, false) when no reliable count can be determined.
func codexMcpServerCount(bin, workingDir string) (int, bool) {
	// Try JSON first for a structured count.
	runCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	jsonCmd := exec.CommandContext(runCtx, bin, "mcp", "list", "--json")
	jsonCmd.Dir = workingDir
	if out, err := jsonCmd.Output(); err == nil {
		if n, ok := parseCodexMcpListJSON(out); ok {
			return n, true
		}
	}

	// Fallback: plaintext. Parse any non-empty, non-header line as a server
	// entry. This is intentionally loose — some Codex versions print a
	// header like "SERVERS:" and some don't. Counting lines that look like
	// server entries (no leading whitespace, non-empty, not the word
	// "SERVERS") is the safest heuristic.
	runCtx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	plainCmd := exec.CommandContext(runCtx2, bin, "mcp", "list")
	plainCmd.Dir = workingDir
	out, err := plainCmd.Output()
	if err != nil {
		return 0, false
	}
	return countCodexMcpPlaintext(out), true
}

// parseCodexMcpListJSON interprets the output of `codex mcp list --json`.
// The upstream format isn't rigidly documented across versions, so we
// accept either a top-level array of server objects or an object with a
// `servers` key carrying the same.
func parseCodexMcpListJSON(data []byte) (int, bool) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return 0, false
	}
	switch trimmed[0] {
	case '[':
		var arr []any
		if err := json.Unmarshal(trimmed, &arr); err != nil {
			return 0, false
		}
		return len(arr), true
	case '{':
		var obj map[string]any
		if err := json.Unmarshal(trimmed, &obj); err != nil {
			return 0, false
		}
		if servers, ok := obj["servers"].([]any); ok {
			return len(servers), true
		}
		if servers, ok := obj["servers"].(map[string]any); ok {
			return len(servers), true
		}
		return 0, false
	default:
		return 0, false
	}
}

// countCodexMcpPlaintext counts server entries in the plaintext output of
// `codex mcp list`. A line counts if it is non-empty, not indented, and
// not the word "SERVERS" (a common header variant).
func countCodexMcpPlaintext(data []byte) int {
	count := 0
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if line != trimmed {
			// Indented continuation line — skip.
			continue
		}
		upper := strings.ToUpper(trimmed)
		if upper == "SERVERS" || upper == "SERVERS:" || strings.HasPrefix(upper, "NAME") {
			continue
		}
		count++
	}
	return count
}
