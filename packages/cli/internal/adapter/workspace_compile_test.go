package adapter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// TestPropagateNoOpOutsideWorkspaceScope verifies the function is a no-op
// for project and global scope, regardless of any Repos configuration.
func TestPropagateNoOpOutsideWorkspaceScope(t *testing.T) {
	for _, scope := range []types.SetupScope{types.SetupScopeProject, types.SetupScopeGlobal} {
		records, err := PropagateMcpToRepos(NewRegistry(), CompileContext{
			SetupScope:    scope,
			WorkspaceRoot: "/tmp/whatever",
			Repos: []types.RepoInfo{
				{Name: "api", Path: "api"},
			},
		})
		if err != nil {
			t.Errorf("scope=%s: unexpected err %v", scope, err)
		}
		if len(records) != 0 {
			t.Errorf("scope=%s: expected no records, got %d", scope, len(records))
		}
	}
}

// TestPropagateNoOpWithoutRepos verifies workspace scope without any repos
// does nothing and returns no error.
func TestPropagateNoOpWithoutRepos(t *testing.T) {
	tmp := t.TempDir()
	records, err := PropagateMcpToRepos(NewRegistry(), CompileContext{
		SetupScope:    types.SetupScopeWorkspace,
		WorkspaceRoot: tmp,
		Repos:         nil,
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected no records, got %d", len(records))
	}
}

// TestPropagateSkipsMissingRepoDirs verifies that repos whose physical
// directory is missing are skipped silently.
func TestPropagateSkipsMissingRepoDirs(t *testing.T) {
	tmp := t.TempDir()
	// Repo "ghost" has no directory on disk.
	records, err := PropagateMcpToRepos(NewRegistry(), CompileContext{
		SetupScope:    types.SetupScopeWorkspace,
		WorkspaceRoot: tmp,
		Repos:         []types.RepoInfo{{Name: "ghost", Path: "ghost"}},
	})
	if err != nil {
		t.Fatalf("missing repo should not error: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("missing repo should produce no records, got %d", len(records))
	}
}

// TestPropagateNilRegistryErrors verifies the registry guard.
func TestPropagateNilRegistryErrors(t *testing.T) {
	_, err := PropagateMcpToRepos(nil, CompileContext{
		SetupScope:    types.SetupScopeWorkspace,
		WorkspaceRoot: "/tmp",
		Repos:         []types.RepoInfo{{Name: "api", Path: "api"}},
	})
	if err == nil {
		t.Fatal("expected error for nil registry")
	}
}

// TestPropagateWritesPerRepoMcpConfigs sets up a workspace with two real
// repos, a canonical .ai/mcp.json, and verifies that propagation writes
// per-repo tool configs (here .opencode/opencode.jsonc as the simplest
// signal — Codex/Claude write distinct files at distinct scopes).
func TestPropagateWritesPerRepoMcpConfigs(t *testing.T) {
	tmp := t.TempDir()

	// Set up a workspace with two repos.
	apiDir := filepath.Join(tmp, "api")
	webDir := filepath.Join(tmp, "web")
	for _, d := range []string{apiDir, webDir} {
		if err := files.EnsureDir(d); err != nil {
			t.Fatalf("ensure %s: %v", d, err)
		}
		// Each repo needs a .ai/mcp.json so adapters can read the catalog.
		aiDir := filepath.Join(d, ".ai")
		if err := files.EnsureDir(aiDir); err != nil {
			t.Fatalf("ensure %s: %v", aiDir, err)
		}
		mcp := []byte(`{"servers":{"filesystem":{"command":"npx","args":["-y","@modelcontextprotocol/server-filesystem","."],"enabled":true}}}`)
		if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), mcp, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Workspace canonical mcp.json (PropagateMcpToRepos reads from each
	// repo's TargetDir, but having it at the workspace root is part of
	// the realistic setup — adapters fall back to it via ReadCanonicalMcp).
	if err := files.EnsureDir(filepath.Join(tmp, ".ai")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".ai", "mcp.json"), []byte(`{"servers":{"filesystem":{"command":"npx","args":["-y","@modelcontextprotocol/server-filesystem","."],"enabled":true}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	records, err := PropagateMcpToRepos(NewRegistry(), CompileContext{
		SetupScope:    types.SetupScopeWorkspace,
		WorkspaceRoot: tmp,
		HomeDir:       t.TempDir(), // unused in workspace path but required for safety
		Repos: []types.RepoInfo{
			{Name: "api", Path: "api", Type: "service"},
			{Name: "web", Path: "web", Type: "frontend"},
		},
	})
	if err != nil {
		t.Fatalf("PropagateMcpToRepos: %v", err)
	}

	// We expect at least one record per repo (each adapter's CompileMCP
	// returns one or more records). Pi adapter is the only one that
	// returns no records, but it's fine — the others produce signal.
	if len(records) == 0 {
		t.Fatal("expected propagated records, got none")
	}

	// Verify each repo got at least one tool config touched. We test a
	// representative path: opencode mcp jsonc.
	for _, r := range []string{"api", "web"} {
		expected := filepath.Join(tmp, r, ".opencode", "lazyai.mcp.jsonc")
		if !files.FileExists(expected) {
			t.Errorf("expected per-repo tool config at %s, missing", expected)
		}
	}
}

// TestSummarizeWorkspaceCompile verifies the summary helper formats
// counts and repo names correctly.
func TestSummarizeWorkspaceCompile(t *testing.T) {
	root := []types.TrackedFile{{Path: "a"}, {Path: "b"}}
	prop := []types.TrackedFile{{Path: "c"}}
	repos := []types.RepoInfo{{Name: "api"}, {Name: "web"}}
	s := SummarizeWorkspaceCompile(root, prop, repos)
	if s.RootCount != 2 {
		t.Errorf("RootCount=%d, want 2", s.RootCount)
	}
	if s.RepoCount != 1 {
		t.Errorf("RepoCount=%d, want 1", s.RepoCount)
	}
	if len(s.Repos) != 2 || s.Repos[0] != "api" || s.Repos[1] != "web" {
		t.Errorf("Repos=%v", s.Repos)
	}
}
