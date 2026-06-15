package handoff

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestWriteReadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "specs", "memory", "handoffs", "2026-06-14-phase4.md")
	doc := Document{
		Goal:            "ship phase four handoff",
		Constraints:     []string{"Repo: lazyai", "Worktree: feature/phase4"},
		Progress:        ProgressInProgress,
		Decisions:       []string{"Keep the handoff writer filesystem-only.", "Store handoff metadata in the V2 runtime table."},
		CriticalContext: "Phase 3 is complete; Phase 4 is wiring session-close handoff output.",
		NextSteps:       []string{"Wire session-close integration.", "Record gate evidence in the phase checklist."},
		Risks:           []string{"Do not start Phase 5 before Phase 4 approval."},
		Owner:           "Ricardo Conceicao",
		SessionID:       "ses_1234567890",
		OpenQuestions:   []string{"Should the handoff include ledger references?"},
		Items: ProgressItems{
			Done:       []string{"Migrate runtime schema."},
			InProgress: []string{"Implement handoff writer."},
			Pending:    []string{"Add session-close regression coverage."},
		},
	}

	if err := Write(path, doc); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	content := string(raw)
	for _, needle := range []string{
		"goal:",
		"constraints:",
		"progress:",
		"decisions:",
		"critical_context:",
		"next_steps:",
		"## Goal",
		"## Constraints & Preferences",
		"## Progress",
		"## Key Decisions",
		"## Critical Context",
		"## Next Steps",
		"## Open Assumptions/Questions",
		"## Risks/Watchouts",
	} {
		if !strings.Contains(content, needle) {
			t.Fatalf("handoff content missing %q:\n%s", needle, content)
		}
	}

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if got.Goal != doc.Goal {
		t.Fatalf("Goal = %q, want %q", got.Goal, doc.Goal)
	}
	if got.Progress != doc.Progress {
		t.Fatalf("Progress = %q, want %q", got.Progress, doc.Progress)
	}
	if got.CriticalContext != doc.CriticalContext {
		t.Fatalf("CriticalContext = %q, want %q", got.CriticalContext, doc.CriticalContext)
	}
	if got.Owner != doc.Owner {
		t.Fatalf("Owner = %q, want %q", got.Owner, doc.Owner)
	}
	if got.SessionID != doc.SessionID {
		t.Fatalf("SessionID = %q, want %q", got.SessionID, doc.SessionID)
	}
	if !reflect.DeepEqual(got.Constraints, doc.Constraints) {
		t.Fatalf("Constraints = %#v, want %#v", got.Constraints, doc.Constraints)
	}
	if !reflect.DeepEqual(got.Decisions, doc.Decisions) {
		t.Fatalf("Decisions = %#v, want %#v", got.Decisions, doc.Decisions)
	}
	if !reflect.DeepEqual(got.NextSteps, doc.NextSteps) {
		t.Fatalf("NextSteps = %#v, want %#v", got.NextSteps, doc.NextSteps)
	}
	if !reflect.DeepEqual(got.Risks, doc.Risks) {
		t.Fatalf("Risks = %#v, want %#v", got.Risks, doc.Risks)
	}
	if !reflect.DeepEqual(got.OpenQuestions, doc.OpenQuestions) {
		t.Fatalf("OpenQuestions = %#v, want %#v", got.OpenQuestions, doc.OpenQuestions)
	}
	if !reflect.DeepEqual(got.Items, doc.Items) {
		t.Fatalf("Items = %#v, want %#v", got.Items, doc.Items)
	}
}

func TestWriteReplacesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "specs", "memory", "handoffs", "2026-06-14-phase4.md")
	first := Document{
		Goal:            "first pass",
		Constraints:     []string{"Repo: lazyai"},
		Progress:        ProgressPending,
		Decisions:       []string{"First decision."},
		CriticalContext: "First context.",
		NextSteps:       []string{"First next step."},
		Items:           ProgressItems{Pending: []string{"First pending item."}},
	}
	second := Document{
		Goal:            "second pass",
		Constraints:     []string{"Repo: lazyai", "Worktree: feature/phase4"},
		Progress:        ProgressDone,
		Decisions:       []string{"Second decision."},
		CriticalContext: "Second context.",
		NextSteps:       []string{"Second next step."},
		Risks:           []string{"Second risk."},
		Items:           ProgressItems{Done: []string{"Second done item."}},
	}

	if err := Write(path, first); err != nil {
		t.Fatalf("first Write failed: %v", err)
	}
	if err := Write(path, second); err != nil {
		t.Fatalf("second Write failed: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	content := string(raw)
	if strings.Contains(content, "First decision.") || strings.Contains(content, "First context.") || strings.Contains(content, "First pending item.") {
		t.Fatalf("replaced handoff kept stale content:\n%s", content)
	}
	for _, heading := range []string{"## Goal", "## Constraints & Preferences", "## Progress", "## Key Decisions", "## Critical Context", "## Next Steps"} {
		if strings.Count(content, heading) != 1 {
			t.Fatalf("heading %q count = %d, want 1\n%s", heading, strings.Count(content, heading), content)
		}
	}

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if got.Goal != second.Goal {
		t.Fatalf("Goal = %q, want %q", got.Goal, second.Goal)
	}
	if !reflect.DeepEqual(got.Decisions, second.Decisions) {
		t.Fatalf("Decisions = %#v, want %#v", got.Decisions, second.Decisions)
	}
	if !reflect.DeepEqual(got.Items.Done, second.Items.Done) {
		t.Fatalf("Done items = %#v, want %#v", got.Items.Done, second.Items.Done)
	}
}
