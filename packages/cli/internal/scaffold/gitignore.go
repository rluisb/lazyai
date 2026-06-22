package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GitignoreEntries lists the LazyAI-owned paths recommended for .gitignore.
var GitignoreEntries = []string{
	".ai/memory/",
}

// LocalSecretsGitignoreEntry is appended to the suggestion list when the
// --local-secrets flag is set (spec 015).
const LocalSecretsGitignoreEntry = ".claude/settings.local.json"

// CheckGitignoreGuidance checks if .gitignore has recommended LazyAI entries and
// prints guidance for missing ones. When localSecrets is true, also checks
// for `.claude/settings.local.json` and appends it to an existing .gitignore
// automatically (non-disruptive: idempotent, only when file already exists).
func CheckGitignoreGuidance(targetDir string, localSecrets bool) {
	gitignorePath := filepath.Join(targetDir, ".gitignore")
	entries := GitignoreEntries
	if localSecrets {
		entries = append(append([]string{}, GitignoreEntries...), LocalSecretsGitignoreEntry)
	}

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("\n💡 Consider creating a .gitignore with:")
			for _, entry := range entries {
				fmt.Printf("   %s\n", entry)
			}
			return
		}
		return
	}

	content := string(data)
	var missing []string

	for _, entry := range entries {
		searchPattern := strings.ReplaceAll(entry, "*", "")
		searchPattern = regexp.QuoteMeta(searchPattern)
		matched, _ := regexp.MatchString(searchPattern, content)
		if !strings.Contains(content, entry) && !matched {
			missing = append(missing, entry)
		}
	}

	// With --local-secrets, auto-append the settings.local.json line if
	// missing (the file is otherwise hidden by its gitignore design). Other
	// entries stay advisory — we don't want to surprise users by mutating
	// their committed .gitignore without explicit opt-in.
	if localSecrets && contains(missing, LocalSecretsGitignoreEntry) {
		line := LocalSecretsGitignoreEntry
		if len(content) > 0 && !strings.HasSuffix(content, "\n") {
			line = "\n" + line
		}
		line += "\n"
		if err := appendFile(gitignorePath, line); err == nil {
			fmt.Printf("\n✓ Added %s to .gitignore (local-secrets)\n", LocalSecretsGitignoreEntry)
			// Remove from missing so we don't repeat it in the suggestion block.
			missing = removeString(missing, LocalSecretsGitignoreEntry)
		}
	}

	if len(missing) > 0 {
		fmt.Println("\n💡 Consider adding to .gitignore:")
		for _, entry := range missing {
			fmt.Printf("   %s\n", entry)
		}
	}
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func removeString(haystack []string, needle string) []string {
	out := haystack[:0]
	for _, s := range haystack {
		if s != needle {
			out = append(out, s)
		}
	}
	return out
}

func appendFile(path, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}
