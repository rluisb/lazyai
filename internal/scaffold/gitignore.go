package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GitignoreEntries lists the recommended .gitignore entries for ai-setup.
var GitignoreEntries = []string{
	".ai/memory/",
	".env",
	".env.local",
	".env*.local",
}

// CheckGitignoreGuidance checks if .gitignore has recommended entries and
// prints guidance for missing ones. Ported from src/scaffold/gitignore.ts.
func CheckGitignoreGuidance(targetDir string) {
	gitignorePath := filepath.Join(targetDir, ".gitignore")

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("\n💡 Consider creating a .gitignore with:")
			for _, entry := range GitignoreEntries {
				fmt.Printf("   %s\n", entry)
			}
			return
		}
		return
	}

	content := string(data)
	var missing []string

	for _, entry := range GitignoreEntries {
		searchPattern := strings.ReplaceAll(entry, "*", "")
		searchPattern = regexp.QuoteMeta(searchPattern)
		matched, _ := regexp.MatchString(searchPattern, content)
		if !strings.Contains(content, entry) && !matched {
			missing = append(missing, entry)
		}
	}

	if len(missing) > 0 {
		fmt.Println("\n💡 Consider adding to .gitignore:")
		for _, entry := range missing {
			fmt.Printf("   %s\n", entry)
		}
	}
}
