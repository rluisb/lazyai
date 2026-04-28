// Package adapter — workspace_compile.go propagates compile-time tool
// configuration from the workspace root out to each repo registered under
// `Config.Repos`. Spec 022 / E2.3 added it so `ai-setup compile` at
// workspace scope produces per-repo MCP wiring without each adapter
// learning about workspaces individually.
//
// The function is conservative on purpose:
//   - No-ops outside workspace scope.
//   - Skips repos whose physical directory is missing (workspace ledger
//     creation is responsible for repo presence; compile shouldn't fail
//     just because a repo wasn't checked out).
//   - Reuses each adapter's existing CompileMCP function with TargetDir
//     pointed at the repo, so all per-tool quirks (Claude settings.json
//     merge, OpenCode jsonc, Codex TOML, etc.) come along for free.
package adapter

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// PropagateMcpToRepos compiles MCP config for every repo registered on the
// workspace, in addition to the workspace-root compile already performed by
// the caller. Returns the list of newly tracked files (records are
// per-repo, with repo-relative paths) and the first error encountered;
// callers may choose to log non-fatal failures and continue.
//
// Behavior matrix:
//
//   - SetupScope != workspace                → returns (nil, nil) immediately.
//   - WorkspaceRoot == "" or Repos == nil    → returns (nil, nil).
//   - repo dir missing or not a directory    → skip with warning (no error).
//   - per-repo compile fails                 → returns the error, but earlier
//     repos remain written (best-effort).
//
// The caller should pass the same Registry it used for the root-level
// compile so the same adapter instances are reused.
func PropagateMcpToRepos(reg *Registry, parent CompileContext) ([]types.TrackedFile, error) {
	if parent.SetupScope != types.SetupScopeWorkspace {
		return nil, nil
	}
	if parent.WorkspaceRoot == "" || len(parent.Repos) == 0 {
		return nil, nil
	}
	if reg == nil {
		return nil, errors.New("PropagateMcpToRepos: nil registry")
	}

	var aggregated []types.TrackedFile
	for _, repo := range parent.Repos {
		repoPath := filepath.Join(parent.WorkspaceRoot, repo.Path)
		if !files.IsDirectory(repoPath) {
			// Repo isn't checked out / present on disk; skip rather than
			// fail. ScaffoldRepoLedgers handles the case where repos are
			// declared but missing.
			continue
		}

		// Each repo gets its own CompileContext so adapters write to
		// <repo>/.claude, <repo>/.opencode, etc. SetupScope=project
		// because per-repo writes use the project layout (.ai/mcp.json
		// is also expected per-repo, falling back to the workspace
		// canonical when missing).
		repoCtx := CompileContext{
			TargetDir:     repoPath,
			HomeDir:       parent.HomeDir,
			SetupScope:    types.SetupScopeProject,
			LocalSecrets:  parent.LocalSecrets,
			WorkspaceRoot: parent.WorkspaceRoot,
		}

		tools := reg.List()
		for _, tool := range tools {
			adapt, err := reg.Get(tool)
			if err != nil {
				continue
			}
			records, err := adapt.CompileMCP(repoCtx)
			if err != nil {
				return aggregated, fmt.Errorf("compile %s for repo %s: %w",
					tool, repo.Name, err)
			}
			// Tag each record with the repo prefix so the caller can
			// distinguish workspace-root writes from per-repo ones.
			for i := range records {
				records[i].Source = fmt.Sprintf("workspace:%s/%s", repo.Name, records[i].Source)
			}
			aggregated = append(aggregated, records...)
		}
	}
	return aggregated, nil
}

// WorkspaceCompileSummary aggregates root and propagated counts so the
// `compile` command can render a single-line workspace status.
type WorkspaceCompileSummary struct {
	RootCount int
	RepoCount int
	Repos     []string
}

// SummarizeWorkspaceCompile produces a WorkspaceCompileSummary from the
// counts the caller already has. Kept tiny on purpose — its only job is
// to keep formatting consistent across `compile` and future workspace
// commands.
func SummarizeWorkspaceCompile(rootRecords, propagated []types.TrackedFile, repos []types.RepoInfo) WorkspaceCompileSummary {
	names := make([]string, 0, len(repos))
	for _, r := range repos {
		names = append(names, r.Name)
	}
	return WorkspaceCompileSummary{
		RootCount: len(rootRecords),
		RepoCount: len(propagated),
		Repos:     names,
	}
}
