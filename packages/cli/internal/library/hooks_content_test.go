package library

import (
	"io/fs"
	"strings"
	"testing"
)

// knownLifecycleEvents is the canonical lifecycle vocabulary from the hook catalog.
// These are the only valid lifecycle event names for hook policy files.
var knownLifecycleEvents = []string{
	"before_agent", "before_model", "before_tool",
	"after_tool", "after_model", "after_agent",
	"on_error", "on_compaction", "on_handoff", "on_human_gate",
}

// hookPolicyFiles returns the relative paths of all markdown hook policy files
// under hooks/ and canonical/hooks/, excluding the catalog itself.
func hookPolicyFiles(t *testing.T) []string {
	t.Helper()

	libFS := GetLibraryFS()
	var paths []string

	for _, dir := range []string{"hooks", "canonical/hooks"} {
		entries, err := fs.ReadDir(libFS, dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			// Skip the catalog itself — it's metadata, not a policy.
			if entry.Name() == "catalog.md" {
				continue
			}
			paths = append(paths, dir+"/"+entry.Name())
		}
	}
	return paths
}

// TestHookPolicyFilesHavePurpose checks that every hook policy markdown file
// declares a purpose or responsibilities section (FR-009 hook completeness).
func TestHookPolicyFilesHavePurpose(t *testing.T) {
	t.Parallel()

	for _, path := range hookPolicyFiles(t) {
		content := readLibraryFile(t, path)
		if !strings.Contains(content, "## Purpose") && !strings.Contains(content, "## Responsibilities") {
			t.Errorf("%s: missing '## Purpose' or '## Responsibilities' section", path)
		}
	}
}

// TestHookPolicyFilesUseKnownLifecycleEvents checks that any backtick-quoted
// lifecycle event references in hook policy files match the canonical vocabulary
// from the catalog. Unknown event names are flagged.
func TestHookPolicyFilesUseKnownLifecycleEvents(t *testing.T) {
	t.Parallel()

	for _, path := range hookPolicyFiles(t) {
		content := readLibraryFile(t, path)

		// Only check files that have an ## Events section.
		eventsIdx := strings.Index(content, "## Events")
		if eventsIdx < 0 {
			continue
		}
		eventsSection := content[eventsIdx:]

		// Scan for backtick-quoted identifiers that look like lifecycle events.
		rest := eventsSection
		for {
			open := strings.Index(rest, "`")
			if open < 0 {
				break
			}
			rest = rest[open+1:]
			close := strings.Index(rest, "`")
			if close < 0 {
				break
			}
			token := rest[:close]
			rest = rest[close+1:]

			// Only flag tokens that look like lifecycle event names.
			if !isLifecycleLike(token) {
				continue
			}
			if !isKnownLifecycle(token) {
				t.Errorf("%s: unknown lifecycle event %q in Events section", path, token)
			}
		}
	}
}

// TestBlockingHookPoliciesDocumentSafetyBehavior checks that hook policies with
// deny or block semantics document their safety behavior (fail-closed semantics
// or safety guardrails section).
func TestBlockingHookPoliciesDocumentSafetyBehavior(t *testing.T) {
	t.Parallel()

	for _, path := range hookPolicyFiles(t) {
		content := readLibraryFile(t, path)

		// Only check hooks that mention deny or block semantics.
		hasDeny := strings.Contains(content, "## Denied Commands") ||
			strings.Contains(content, "- Deny") ||
			strings.Contains(content, "- Deny when")
		hasBlock := strings.Contains(content, "## Decision") &&
			(strings.Contains(content, "Deny") || strings.Contains(content, "Block"))

		if !hasDeny && !hasBlock {
			continue
		}

		hasSafety := strings.Contains(content, "## Fail-Closed Semantics") ||
			strings.Contains(content, "## Safety Guardrails")
		if !hasSafety {
			t.Errorf("%s: hook with deny/block semantics missing '## Fail-Closed Semantics' or '## Safety Guardrails' section", path)
		}
	}
}

// isLifecycleLike returns true if the token looks like a lifecycle event name.
func isLifecycleLike(token string) bool {
	return strings.HasPrefix(token, "before_") ||
		strings.HasPrefix(token, "after_") ||
		strings.HasPrefix(token, "on_")
}

// isKnownLifecycle returns true if the token is in the known lifecycle vocabulary.
func isKnownLifecycle(token string) bool {
	for _, known := range knownLifecycleEvents {
		if token == known {
			return true
		}
	}
	return false
}
