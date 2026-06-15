package library

import (
	"strings"
	"testing"
)

func TestAutoRecoveryPolicyDefinesConservativeAllowlist(t *testing.T) {
	t.Parallel()

	content := readLibraryFile(t, "rules/auto-recovery.md")

	required := []string{
		"## Safe Auto-Recovery Allowlist",
		"re-run deterministic checks",
		"retry transient provider/tool failures within existing retry limits",
		"regenerate malformed report JSON from the same inputs",
		"create handoff when blocked",
		"idempotency",
		"same inputs",
		"failure cause/evidence",
		"retry limit",
	}

	assertContainsAll(t, "rules/auto-recovery.md", content, required)
}

func TestAutoRecoveryPolicyRequiresHumanConfirmationForUnsafeRecovery(t *testing.T) {
	t.Parallel()

	content := readLibraryFile(t, "rules/auto-recovery.md")

	required := []string{
		"## Human Confirmation Required",
		"code edits",
		"dependency changes",
		"destructive commands",
		"migration changes",
		"secrets/config changes",
		"ambiguous failures",
		"No destructive recovery without explicit human approval",
	}

	assertContainsAll(t, "rules/auto-recovery.md", content, required)
}

func TestAutoRecoveryGuidanceDoesNotClaimRuntimeAutonomy(t *testing.T) {
	t.Parallel()

	paths := []string{
		"rules/auto-recovery.md",
	}
	for _, path := range paths {
		content := strings.ToLower(readLibraryFile(t, path))
		forbidden := []string{
			"runtime autonomous recovery is enabled",
			"autonomous runtime recovery is enabled",
			"automatically edit files after failures",
			"automatically edits files after failures",
			"auto-apply fixes",
			"bypass human approval",
		}
		for _, phrase := range forbidden {
			if strings.Contains(content, phrase) {
				t.Errorf("%s should not claim runtime autonomous recovery support with phrase %q", path, phrase)
			}
		}
	}
}
