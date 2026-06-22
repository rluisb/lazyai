package migration

import (
	"path/filepath"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
)

func TestDetectSetupCoversPhaseDSources(t *testing.T) {
	dir := t.TempDir()
	mustWrite := func(rel, content string) {
		if err := files.WriteFile(filepath.Join(dir, rel), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite(".pi/agents/researcher.md", "# Researcher\n")
	mustWrite(".omp/skills/diagnose/SKILL.md", "# Diagnose\n")
	mustWrite(".kiro/skills/review/SKILL.md", "# Review\n")
	mustWrite(".agents/skills/triage/SKILL.md", "# Triage\n")
	mustWrite(".github/instructions/team.instructions.md", "# Team instructions\n")

	got, err := DetectSetup(dir)
	if err != nil {
		t.Fatalf("DetectSetup: %v", err)
	}

	seen := map[string]bool{}
	for _, detection := range got {
		seen[detection.AdapterID] = true
	}
	for _, adapter := range []string{"pi", "omp", "kiro", "antigravity", "copilot"} {
		if !seen[adapter] {
			t.Fatalf("expected adapter %q in detections, got %+v", adapter, got)
		}
	}
}
