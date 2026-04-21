package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCheckGitignoreGuidance_LocalSecrets_AppendsWhenMissing verifies that when
// localSecrets=true and .gitignore exists without the settings.local.json line,
// the line is appended automatically.
func TestCheckGitignoreGuidance_LocalSecrets_AppendsWhenMissing(t *testing.T) {
	targetDir := t.TempDir()
	gitignorePath := filepath.Join(targetDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("node_modules/\n.env\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	CheckGitignoreGuidance(targetDir, true)

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(data), LocalSecretsGitignoreEntry) {
		t.Errorf("expected %q appended to .gitignore, got:\n%s", LocalSecretsGitignoreEntry, string(data))
	}
}

// TestCheckGitignoreGuidance_LocalSecrets_IdempotentAppend verifies that a
// second run does not duplicate the line.
func TestCheckGitignoreGuidance_LocalSecrets_IdempotentAppend(t *testing.T) {
	targetDir := t.TempDir()
	gitignorePath := filepath.Join(targetDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("node_modules/\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	CheckGitignoreGuidance(targetDir, true)
	CheckGitignoreGuidance(targetDir, true)

	data, _ := os.ReadFile(gitignorePath)
	count := strings.Count(string(data), LocalSecretsGitignoreEntry)
	if count != 1 {
		t.Errorf("expected exactly 1 occurrence of %q, got %d:\n%s", LocalSecretsGitignoreEntry, count, string(data))
	}
}

// TestCheckGitignoreGuidance_WithoutLocalSecrets_NoAutoAppend verifies that
// when the flag is off, .gitignore is never mutated.
func TestCheckGitignoreGuidance_WithoutLocalSecrets_NoAutoAppend(t *testing.T) {
	targetDir := t.TempDir()
	gitignorePath := filepath.Join(targetDir, ".gitignore")
	original := "node_modules/\n"
	if err := os.WriteFile(gitignorePath, []byte(original), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	CheckGitignoreGuidance(targetDir, false)

	data, _ := os.ReadFile(gitignorePath)
	if string(data) != original {
		t.Errorf(".gitignore mutated without --local-secrets:\ngot:  %q\nwant: %q", string(data), original)
	}
}
