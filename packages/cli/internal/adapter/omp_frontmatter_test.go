package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/frontmatter"
	"github.com/rluisb/lazyai/packages/cli/internal/library"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// TestRewriteAgentForOMP_ResearcherIsReadOnly verifies that a read-only agent
// (tools: [read, search]) gets only read and search in its OMP tools list, and
// that LazyAI-only fields are not emitted.
func TestRewriteAgentForOMP_ResearcherIsReadOnly(t *testing.T) {
	source := []byte(`---
name: researcher
description: Scout agent — read-only codebase explorer.
role: researcher
mode: all
temperature: 0.2
steps: 10
tools:
  - read
  - search
skills:
  - tdd-planning
---

# System Prompt
You are a research specialist.
`)
	out, err := RewriteAgentForOMP(source, &AdapterContext{})
	if err != nil {
		t.Fatalf("RewriteAgentForOMP: %v", err)
	}
	outStr := string(out)

	// OMP-native tools must contain read and search.
	for _, want := range []string{`"read"`, `"search"`} {
		if !strings.Contains(outStr, want) {
			t.Errorf("tools missing %s:\n%s", want, outStr)
		}
	}
	// Mutable tools must be absent for read-only agents.
	for _, forbidden := range []string{`"bash"`, `"edit"`, `"write"`, `"task"`} {
		if strings.Contains(outStr, forbidden) {
			t.Errorf("tools must not contain %s for read-only agent:\n%s", forbidden, outStr)
		}
	}
	// LazyAI-only fields must not leak.
	fm, _, err := frontmatter.ExtractFrontmatter(out)
	if err != nil {
		t.Fatalf("parse emitted frontmatter: %v", err)
	}
	for _, leaked := range []string{"role", "mode", "temperature", "steps"} {
		if _, ok := fm[leaked]; ok {
			t.Errorf("LazyAI-only field %q must not appear in OMP output", leaked)
		}
	}
	// autoloadSkills from canonical skills:.
	if !strings.Contains(outStr, "tdd-planning") {
		t.Errorf("autoloadSkills must include tdd-planning:\n%s", outStr)
	}
	// thinkingLevel: low for read-only.
	if !strings.Contains(outStr, "thinkingLevel: low") {
		t.Errorf("thinkingLevel must be 'low' for read-only agent:\n%s", outStr)
	}
}

// TestRewriteAgentForOMP_DeployerHasShell verifies that a full-capability
// agent includes bash in its OMP tools list and emits no autoloadSkills when
// the canonical source has no skills: list.
func TestRewriteAgentForOMP_DeployerHasShell(t *testing.T) {
	source := []byte(`---
name: deployer
description: Infrastructure, deployment, and CI/CD operations agent.
role: deployer
mode: all
temperature: 0.2
steps: 14
tools:
  - read
  - edit
  - shell
  - search
  - web
  - mcp
  - spawn
---

# System Prompt
You are a deployment specialist.
`)
	out, err := RewriteAgentForOMP(source, &AdapterContext{})
	if err != nil {
		t.Fatalf("RewriteAgentForOMP: %v", err)
	}
	outStr := string(out)

	if !strings.Contains(outStr, `"bash"`) {
		t.Errorf("deployer tools must include bash:\n%s", outStr)
	}
	// No autoloadSkills when there is no skills: field.
	if strings.Contains(outStr, "autoloadSkills") {
		t.Errorf("autoloadSkills must be absent when no skills: field:\n%s", outStr)
	}
	// thinkingLevel: auto (deployer is not read-only and not planner).
	if !strings.Contains(outStr, "thinkingLevel: auto") {
		t.Errorf("thinkingLevel must be 'auto' for deployer:\n%s", outStr)
	}
}

// TestRewriteAgentForOMP_PlannerHasHighThinking verifies the planner agent
// gets thinkingLevel: high.
func TestRewriteAgentForOMP_PlannerHasHighThinking(t *testing.T) {
	source := []byte(`---
name: planner
description: Specification and planning agent.
role: planner
tools:
  - read
  - edit
  - shell
  - search
  - web
  - mcp
  - spawn
skills:
  - tdd-planning
---

# System Prompt
You are a planning specialist.
`)
	out, err := RewriteAgentForOMP(source, &AdapterContext{})
	if err != nil {
		t.Fatalf("RewriteAgentForOMP: %v", err)
	}
	if !strings.Contains(string(out), "thinkingLevel: high") {
		t.Errorf("thinkingLevel must be 'high' for planner:\n%s", out)
	}
}

// TestRewriteAgentForOMP_ManagedMarker verifies the vibe-lab managed marker
// is present in the emitted output.
func TestRewriteAgentForOMP_ManagedMarker(t *testing.T) {
	source := []byte(`---
name: implementer
description: Universal implementer.
tools:
  - read
  - edit
  - shell
  - search
---

# System Prompt
You are an implementer.
`)
	out, err := RewriteAgentForOMP(source, &AdapterContext{})
	if err != nil {
		t.Fatalf("RewriteAgentForOMP: %v", err)
	}
	want := managedAgentMarker("omp", "implementer")
	if !strings.Contains(string(out), want) {
		t.Errorf("missing managed marker %q in:\n%s", want, out)
	}
}

// TestRewriteAgentForOMP_BodyPreserved verifies the agent body is preserved
// verbatim after the managed marker.
func TestRewriteAgentForOMP_BodyPreserved(t *testing.T) {
	source := []byte(`---
name: researcher
description: Scout agent.
tools:
  - read
  - search
---

# System Prompt
You are a research specialist. Unique marker: abc123xyz.
`)
	out, err := RewriteAgentForOMP(source, &AdapterContext{})
	if err != nil {
		t.Fatalf("RewriteAgentForOMP: %v", err)
	}
	if !strings.Contains(string(out), "Unique marker: abc123xyz") {
		t.Errorf("body not preserved in output:\n%s", out)
	}
}

// TestRewriteAgentForOMP_NilGrantsFallback verifies that an agent without a
// tools: field gets the full unrestricted OMP tools set (legacy behaviour).
func TestRewriteAgentForOMP_NilGrantsFallback(t *testing.T) {
	source := []byte(`---
name: guide
description: Front-door agent.
---

Body.
`)
	out, err := RewriteAgentForOMP(source, &AdapterContext{})
	if err != nil {
		t.Fatalf("RewriteAgentForOMP: %v", err)
	}
	outStr := string(out)
	for _, want := range []string{`"read"`, `"search"`, `"bash"`, `"edit"`, `"write"`} {
		if !strings.Contains(outStr, want) {
			t.Errorf("nil-grants fallback missing %s:\n%s", want, outStr)
		}
	}
}

// TestRewriteAgentForOMP_SkillsAutoloadOmittedWhenEmpty verifies that
// autoloadSkills is omitted when the canonical source has an empty skills list.
func TestRewriteAgentForOMP_SkillsAutoloadOmittedWhenEmpty(t *testing.T) {
	source := []byte(`---
name: responder
description: SRE and incident response agent.
tools:
  - read
  - edit
  - shell
  - search
---

Body.
`)
	out, err := RewriteAgentForOMP(source, &AdapterContext{})
	if err != nil {
		t.Fatalf("RewriteAgentForOMP: %v", err)
	}
	if strings.Contains(string(out), "autoloadSkills") {
		t.Errorf("autoloadSkills must be absent when skills list is empty:\n%s", out)
	}
}

// TestRewriteAgentForOMP_MultipleSkills verifies that multiple skills map to
// multiple autoloadSkills entries.
func TestRewriteAgentForOMP_MultipleSkills(t *testing.T) {
	source := []byte(`---
name: implementer
description: Universal implementer.
tools:
  - read
  - edit
  - shell
  - search
  - web
  - mcp
  - spawn
skills:
  - build-mode
  - tdd-planning
  - refresh-dev-containers
---

Body.
`)
	out, err := RewriteAgentForOMP(source, &AdapterContext{})
	if err != nil {
		t.Fatalf("RewriteAgentForOMP: %v", err)
	}
	outStr := string(out)
	for _, skill := range []string{"build-mode", "tdd-planning", "refresh-dev-containers"} {
		if !strings.Contains(outStr, skill) {
			t.Errorf("autoloadSkills missing %q:\n%s", skill, outStr)
		}
	}
}

// TestOmpAdapter_Install_AgentFrontmatterContent is an integration test that
// installs against the real library and verifies:
//  1. The on-disk researcher.md has OMP-native fields (tools, thinkingLevel, autoloadSkills).
//  2. LazyAI-only fields (role, mode, temperature, steps) are absent.
//  3. read-only agents (researcher) have only read+search in tools.
func TestOmpAdapter_Install_AgentFrontmatterContent(t *testing.T) {
	libDir, err := library.FindLibraryDir()
	if err != nil {
		t.Skipf("library not found (embedded-only build): %v", err)
	}

	targetDir := t.TempDir()
	ctx := &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
		LibraryDir: libDir,
		LibraryFS:  nil,
		Strategy:   types.ConflictStrategyAlign,
		Selections: AdapterSelections{
			Agents: []types.AgentId{types.AgentIdResearcher},
		},
	}

	adapter := &OmpAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("OMP Install failed: %v", err)
	}

	agentPath := filepath.Join(targetDir, ".omp", "agents", "researcher.md")
	content, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("read researcher.md: %v", err)
	}
	outStr := string(content)

	fm, _, err := frontmatter.ExtractFrontmatter(content)
	if err != nil {
		t.Fatalf("parse frontmatter: %v", err)
	}

	// Required OMP-native fields.
	if frontmatter.ExtractField(fm, "name") == "" {
		t.Error("emitted agent missing 'name' field")
	}
	if frontmatter.ExtractField(fm, "description") == "" {
		t.Error("emitted agent missing 'description' field")
	}
	if _, ok := fm["tools"]; !ok {
		t.Fatal("emitted agent missing 'tools' field")
	}
	if !strings.Contains(outStr, "thinkingLevel:") {
		t.Error("emitted agent missing 'thinkingLevel' field")
	}

	// researcher is read-only: must have read and search, no bash/edit/write/task.
	for _, want := range []string{`"read"`, `"search"`} {
		if !strings.Contains(outStr, want) {
			t.Errorf("researcher tools must contain %s:\n%s", want, outStr)
		}
	}
	for _, forbidden := range []string{`"bash"`, `"edit"`, `"write"`, `"task"`} {
		if strings.Contains(outStr, forbidden) {
			t.Errorf("researcher tools must not contain %s:\n%s", forbidden, outStr)
		}
	}

	// LazyAI-only fields must not appear.
	for _, leaked := range []string{"role", "mode", "temperature", "steps"} {
		if _, ok := fm[leaked]; ok {
			t.Errorf("LazyAI-only field %q must not appear in OMP output", leaked)
		}
	}

	// researcher has skills: [tdd-planning] → autoloadSkills must include it.
	if !strings.Contains(outStr, "tdd-planning") {
		t.Error("autoloadSkills must include tdd-planning for researcher")
	}
}
