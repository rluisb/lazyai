package scaffold

import (
	"strings"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/reversa/state"
)

// countPlaceholders counts <!-- fill-in: markers in the compiled content.
func countPlaceholders(content string) int {
	return strings.Count(content, "<!-- fill-in:")
}

// writePopulateSignal writes .ai/populate-needed after checking remaining placeholders.
func writePopulateSignal(targetDir string, content string) error {
	remaining := countPlaceholders(content)
	if remaining == 0 {
		return nil
	}

	return state.WritePopulateSignal(targetDir, state.PopulateSignal{
		PlaceholderCount: remaining,
		GeneratedAt:      time.Now().UTC().Format(time.RFC3339),
		Skipped:          false,
	})
}
