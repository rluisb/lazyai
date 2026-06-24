package scaffold

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/conflict"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// RepoPermissions defines the default permissions for workspace repos.
type RepoPermissions struct {
	Write          bool
	RunCommands    bool
	RunDestructive bool
	GitOperations  bool
}

// DefaultRepoPermissions returns the default repo permissions.
func DefaultRepoPermissions() RepoPermissions {
	return RepoPermissions{
		Write:          true,
		RunCommands:    true,
		RunDestructive: false,
		GitOperations:  false,
	}
}

// ScaffoldRepoRoots generates lightweight root files in each referenced repo.
// Ported from src/scaffold/repo-roots.ts.
func ScaffoldRepoRoots(repos []types.RepoInfo, planningRepoPath string, tools []types.ToolId, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) map[string][]types.TrackedFile {
	results := make(map[string][]types.TrackedFile)

	for _, repo := range repos {
		repoAbsPath := filepath.Join(planningRepoPath, repo.Path)
		var records []types.TrackedFile

		if !files.FileExists(repoAbsPath) || !files.IsDirectory(repoAbsPath) {
			results[repo.Name] = records
			continue
		}

		// Generate repo root content (basic version without stack detection).
		content := generateRepoRootContent(repo.Name, planningRepoPath)

		writtenFiles := make(map[string]bool)
		for _, tool := range tools {
			outputFile, ok := RootFileByTool[tool]
			if !ok {
				continue
			}
			if writtenFiles[outputFile] {
				continue
			}
			writtenFiles[outputFile] = true

			destPath := filepath.Join(repoAbsPath, outputFile)
			if strings.Contains(outputFile, "/") {
				_ = files.EnsureDir(filepath.Dir(destPath))
			}

			action, err := conflict.ApplyStrategy(destPath, strategy, perFileOverrides, repoAbsPath)
			if err != nil || action == "skip" {
				continue
			}

			if err := files.WriteFile(destPath, []byte(content), 0o644); err != nil {
				scaffoldLog.Warn("could not write repo root", "path", destPath, "error", err)
				continue
			}

			hash, _ := files.FileHash(destPath)
			records = append(records, types.TrackedFile{
				Path:   repo.Name + "/" + outputFile,
				Hash:   hash,
				Source: "workspace:repo-root",
				Owner:  types.FileOwnerLibrary,
			})
		}

		// Check if claude-code is in the tools list.
		hasClaudeCode := false
		for _, t := range tools {
			if t == types.ToolIdClaudeCode {
				hasClaudeCode = true
				break
			}
		}

		if hasClaudeCode {
			settings := GenerateClaudeSettings(DefaultRepoPermissions())
			claudeDir := filepath.Join(repoAbsPath, ".claude")
			_ = files.EnsureDir(claudeDir)

			settingsPath := filepath.Join(claudeDir, "settings.json")
			action, err := conflict.ApplyStrategy(settingsPath, strategy, perFileOverrides, repoAbsPath)
			if err == nil && action != "skip" {
				content, _ := json.MarshalIndent(settings, "", "  ")
				if err := files.WriteFile(settingsPath, content, 0o644); err == nil {
					hash, _ := files.FileHash(settingsPath)
					records = append(records, types.TrackedFile{
						Path:   repo.Name + "/.claude/settings.json",
						Hash:   hash,
						Source: "workspace:permissions",
						Owner:  types.FileOwnerLibrary,
					})
				}
			}
		}

		results[repo.Name] = records
	}

	return results
}

// GenerateClaudeSettings generates Claude Code settings for a repo.
func GenerateClaudeSettings(permissions RepoPermissions) map[string]any {
	var allow []string
	allow = append(allow, "Read")

	if permissions.Write {
		allow = append(allow, "Edit")
	}

	var deny []string
	if !permissions.RunDestructive {
		deny = append(deny, "Bash(rm -rf *)")
	}
	if !permissions.GitOperations {
		deny = append(deny, "Bash(git push*)", "Bash(git push --force*)")
	}

	return map[string]any{
		"permissions": map[string]any{
			"allow": allow,
			"deny":  deny,
		},
	}
}

// ScaffoldRepoLedgers creates activity ledger and state files for each workspace repo.
// Ported from src/scaffold/repo-roots.ts scaffoldRepoLedgers.
func ScaffoldRepoLedgers(planningRepoPath string, repos []types.RepoInfo, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	// Detect memory path: .specify/memory/repos/ (speckit) or specs/memory/repos/ (legacy)
	baseMemoryRepos := filepath.Join(planningRepoPath, "specs", "memory", "repos")
	if specifyPath := filepath.Join(planningRepoPath, ".specify", "memory", "repos"); files.IsDirectory(specifyPath) {
		baseMemoryRepos = specifyPath
	}

	for _, repo := range repos {
		repoMemoryDir := filepath.Join(baseMemoryRepos, repo.Name)
		if err := files.EnsureDir(repoMemoryDir); err != nil {
			return err
		}

		// Create ledger.
		ledgerPath := filepath.Join(repoMemoryDir, "ledger.md")
		ledgerContent := fmt.Sprintf(`# %s — Activity Ledger

| Date | Who | What | Plan ref | Verified |
|------|-----|------|----------|----------|

<!-- AI: append a new row after every task completed in this repo -->
`, repo.Name)

		action, err := conflict.ApplyStrategy(ledgerPath, strategy, perFileOverrides, planningRepoPath)
		if err != nil {
			return err
		}
		if action != "skip" {
			if err := files.WriteFile(ledgerPath, []byte(ledgerContent), 0o644); err != nil {
				return err
			}
			hash, _ := files.FileHash(ledgerPath)
			relPath, _ := filepath.Rel(planningRepoPath, ledgerPath)
			*fileRecords = append(*fileRecords, types.TrackedFile{
				Path:   filepath.ToSlash(relPath),
				Hash:   hash,
				Source: "workspace:ledger",
				Owner:  types.FileOwnerLibrary,
			})
		}

		// Create last-known-state.
		statePath := filepath.Join(repoMemoryDir, "last-known-state.md")
		stateLines := []string{
			fmt.Sprintf("# %s — Last Known State", repo.Name),
			"",
			fmt.Sprintf("- **Type**: %s", repo.Type),
			"- **Language**: Unknown",
		}
		if repo.Description != "" {
			stateLines = append(stateLines, fmt.Sprintf("- **Description**: %s", repo.Description))
		}
		stateLines = append(stateLines, "", fmt.Sprintf("*Generated: %s*", time.Now().Format("2006-01-02")), "")

		action2, err := conflict.ApplyStrategy(statePath, strategy, perFileOverrides, planningRepoPath)
		if err != nil {
			return err
		}
		if action2 != "skip" {
			if err := files.WriteFile(statePath, []byte(strings.Join(stateLines, "\n")+"\n"), 0o644); err != nil {
				return err
			}
			hash, _ := files.FileHash(statePath)
			relPath, _ := filepath.Rel(planningRepoPath, statePath)
			*fileRecords = append(*fileRecords, types.TrackedFile{
				Path:   filepath.ToSlash(relPath),
				Hash:   hash,
				Source: "workspace:state",
				Owner:  types.FileOwnerLibrary,
			})
		}
	}

	return nil
}

// generateRepoRootContent creates a root file content for a workspace repo.
func generateRepoRootContent(repoName, planningRepoPath string) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("# %s", repoName), "", "## Project Stack", "")
	lines = append(lines, "- **Language**: Unknown")
	lines = append(lines, "",
		"## Workspace", "",
		"This repo is part of a workspace. Plans, standards, and coordination live in the planning repo.", "",
		fmt.Sprintf("- **Planning repo**: %s", planningRepoPath),
		"- For feature plans, see: specs/features/ in the planning repo",
		"- For coding standards, see: specs/standards/ in the planning repo", "",
		"## Claude Code Permissions", "",
		"- Default Claude Code permissions: read, write, and safe project commands are allowed",
		"- Destructive commands and git push operations are denied by default",
		"- If this repo needs different access, customize `.claude/settings.json` manually", "",
		"## Before Making Changes", "",
		"1. Pull latest — other team members may have pushed",
		"2. Check the plan is still current — read the task file before implementing",
		"3. After completing work — update the ledger in the planning repo", "",
	)
	return strings.Join(lines, "\n") + "\n"
}
