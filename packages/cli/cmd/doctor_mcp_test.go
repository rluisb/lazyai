package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeClaudeJSON(t *testing.T, home, body string) {
	t.Helper()
	path := filepath.Join(home, ".claude.json")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestFindStaleClaudeMcpEntries_DetectsLegacyCommand(t *testing.T) {
	home := t.TempDir()
	writeClaudeJSON(t, home, `{
	  "mcpServers": {
	    "orchestrator": {
	      "command": "ai-setup-orchestrator",
	      "args": ["connect"]
	    }
	  }
	}`)

	entries, err := findStaleClaudeMcpEntries(home)
	if err != nil {
		t.Fatalf("findStaleClaudeMcpEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 stale entry, got %d: %+v", len(entries), entries)
	}
	e := entries[0]
	if e.Name != "orchestrator" {
		t.Errorf("Name = %q, want orchestrator", e.Name)
	}
	if e.Scope != "user" {
		t.Errorf("Scope = %q, want user", e.Scope)
	}
	if e.Remediation != "claude mcp remove orchestrator -s user" {
		t.Errorf("Remediation = %q", e.Remediation)
	}
	if !strings.Contains(e.Reason, "ai-setup-orchestrator") {
		t.Errorf("Reason does not name the legacy command: %q", e.Reason)
	}
}

func TestFindStaleClaudeMcpEntries_DetectsLegacyPackageInArgs(t *testing.T) {
	home := t.TempDir()
	writeClaudeJSON(t, home, `{
	  "mcpServers": {
	    "orchestrator": {
	      "command": "npx",
	      "args": ["-y", "@ai-setup/orchestrator", "connect"]
	    }
	  }
	}`)

	entries, err := findStaleClaudeMcpEntries(home)
	if err != nil {
		t.Fatalf("findStaleClaudeMcpEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 stale entry, got %d", len(entries))
	}
	if !strings.Contains(entries[0].Reason, "@ai-setup/orchestrator") {
		t.Errorf("Reason does not surface the legacy package: %q", entries[0].Reason)
	}
}

func TestFindStaleClaudeMcpEntries_IgnoresHealthyLazyAIEntry(t *testing.T) {
	home := t.TempDir()
	writeClaudeJSON(t, home, `{
	  "mcpServers": {
	    "orchestrator": {
	      "command": "lazyai-orchestrator",
	      "args": ["connect"]
	    },
	    "memory": {
	      "command": "npx",
	      "args": ["-y", "@modelcontextprotocol/server-memory"]
	    }
	  }
	}`)

	entries, err := findStaleClaudeMcpEntries(home)
	if err != nil {
		t.Fatalf("findStaleClaudeMcpEntries: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 stale entries, got %d: %+v", len(entries), entries)
	}
}

func TestFindStaleClaudeMcpEntries_MissingConfigFileIsHealthy(t *testing.T) {
	home := t.TempDir() // no .claude.json written
	entries, err := findStaleClaudeMcpEntries(home)
	if err != nil {
		t.Fatalf("findStaleClaudeMcpEntries: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries for missing config, got %d", len(entries))
	}
}

func TestFindStaleClaudeMcpEntries_MalformedJSONIsNotAnError(t *testing.T) {
	home := t.TempDir()
	writeClaudeJSON(t, home, `{not valid json`)
	entries, err := findStaleClaudeMcpEntries(home)
	if err != nil {
		t.Fatalf("malformed JSON must not propagate as error, got: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("malformed JSON must yield 0 entries, got %d", len(entries))
	}
}

func TestFindStaleClaudeMcpEntries_EmptyHomeIsHealthy(t *testing.T) {
	// Defensive: caller passes "" (os.UserHomeDir failed).
	entries, err := findStaleClaudeMcpEntries("")
	if err != nil {
		t.Fatalf("findStaleClaudeMcpEntries(\"\"): %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries for empty home, got %d", len(entries))
	}
}

func TestFindStaleClaudeMcpEntries_StableOrderingByName(t *testing.T) {
	home := t.TempDir()
	writeClaudeJSON(t, home, `{
	  "mcpServers": {
	    "zorch": {"command": "ai-setup-orchestrator", "args": []},
	    "alpha": {"command": "ai-setup-orchestrator", "args": []}
	  }
	}`)
	entries, err := findStaleClaudeMcpEntries(home)
	if err != nil {
		t.Fatalf("findStaleClaudeMcpEntries: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 stale entries, got %d", len(entries))
	}
	if entries[0].Name != "alpha" || entries[1].Name != "zorch" {
		t.Errorf("entries not sorted by name: %+v", entries)
	}
}
