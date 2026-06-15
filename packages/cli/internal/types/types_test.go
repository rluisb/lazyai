package types

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDefaultStoreData_ReturnsValidDefaults(t *testing.T) {
	t.Parallel()

	d := DefaultStoreData()

	if d.Meta.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("SchemaVersion = %d, want %d", d.Meta.SchemaVersion, CurrentSchemaVersion)
	}
	if d.Meta.CLIVersion != "" {
		t.Errorf("CLIVersion = %q, want empty", d.Meta.CLIVersion)
	}
	if d.Meta.InstalledAt == "" {
		t.Error("InstalledAt is empty, want a timestamp")
	}
	if d.Meta.LastUpdatedAt == "" {
		t.Error("LastUpdatedAt is empty, want a timestamp")
	}

	if d.Config.SetupScope != SetupScopeProject {
		t.Errorf("SetupScope = %q, want %q", d.Config.SetupScope, SetupScopeProject)
	}
	if d.Config.ProjectName != "" {
		t.Errorf("ProjectName = %q, want empty", d.Config.ProjectName)
	}

	if d.Files == nil {
		t.Error("Files is nil, want empty slice")
	}
	if len(d.Files) != 0 {
		t.Errorf("Files has %d elements, want 0", len(d.Files))
	}

	if d.Operations == nil {
		t.Error("Operations is nil, want empty slice")
	}
	if len(d.Operations) != 0 {
		t.Errorf("Operations has %d elements, want 0", len(d.Operations))
	}

	if !d.Sync.Dirty {
		t.Error("Sync.Dirty = false, want true")
	}
}

func TestDefaultStoreData_EmptySlicesMarshalAsArray(t *testing.T) {
	t.Parallel()

	d := DefaultStoreData()
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("Marshal StoreData: %v", err)
	}
	s := string(data)

	for _, field := range []string{`"files":[]`, `"operations":[]`, `"tools":[]`} {
		if !strings.Contains(s, field) {
			t.Errorf("JSON does not contain %q", field)
		}
	}
}

func TestDefaultFeatureFlags_ReturnsCorrectValues(t *testing.T) {
	t.Parallel()

	f := DefaultFeatureFlags()

	if !f.ContextEngineering || !f.RPIWorkflow || !f.ChainOfThought ||
		!f.TreeOfThoughts || !f.ADREnforcement || !f.QualityGates ||
		!f.AgentHarness || !f.BugResolution || !f.PivotHandling {
		t.Errorf("DefaultFeatureFlags has some flags disabled: %+v", f)
	}
}

func TestDefaultGitConventions_ReturnsCorrectPatterns(t *testing.T) {
	t.Parallel()

	g := DefaultGitConventions()

	if g.BranchPattern != "{type}/{ticket}-{description}" {
		t.Errorf("BranchPattern = %q, want {type}/{ticket}-{description}", g.BranchPattern)
	}
	if g.CommitPattern != "{type}({scope}): {description}" {
		t.Errorf("CommitPattern = %q, want {type}({scope}): {description}", g.CommitPattern)
	}
	if g.RequireTicket != false {
		t.Error("RequireTicket = true, want false")
	}
	if g.TicketPattern != "[A-Z]+-[0-9]+" {
		t.Errorf("TicketPattern = %q, want [A-Z]+-[0-9]+", g.TicketPattern)
	}

	expectedTypes := []string{"feat", "fix", "docs", "style", "refactor", "perf", "test", "build", "ci", "chore", "revert"}
	if len(g.Types) != len(expectedTypes) {
		t.Fatalf("Types has %d elements, want %d", len(g.Types), len(expectedTypes))
	}
	for i, want := range expectedTypes {
		if g.Types[i] != want {
			t.Errorf("Types[%d] = %q, want %q", i, g.Types[i], want)
		}
	}
}

func TestTypeConstants(t *testing.T) {
	t.Parallel()

	if SetupScopeGlobal != "global" {
		t.Errorf("SetupScopeGlobal = %q, want global", SetupScopeGlobal)
	}
	if SetupScopeWorkspace != "workspace" {
		t.Errorf("SetupScopeWorkspace = %q, want workspace", SetupScopeWorkspace)
	}
	if SetupScopeProject != "project" {
		t.Errorf("SetupScopeProject = %q, want project", SetupScopeProject)
	}

	if ToolIdOpenCode != "opencode" {
		t.Errorf("ToolIdOpenCode = %q, want opencode", ToolIdOpenCode)
	}
	if ToolIdClaudeCode != "claude-code" {
		t.Errorf("ToolIdClaudeCode = %q, want claude-code", ToolIdClaudeCode)
	}
}

func TestAllSlices_ContainExpectedElements(t *testing.T) {
	t.Parallel()

	if len(ALL_AGENTS) != 5 {
		t.Errorf("ALL_AGENTS has %d elements, want 5 active canonical agents", len(ALL_AGENTS))
	}
	if len(ALL_SKILLS) != 4 {
		t.Errorf("ALL_SKILLS has %d elements, want 4 active canonical skills", len(ALL_SKILLS))
	}
	if len(ALL_PROMPTS) != 5 {
		t.Errorf("ALL_PROMPTS has %d elements, want 5", len(ALL_PROMPTS))
	}
	if len(ALL_TEMPLATES) != 10 {
		t.Errorf("ALL_TEMPLATES has %d elements, want 10", len(ALL_TEMPLATES))
	}
	if len(ALL_RULES) != 9 {
		t.Errorf("ALL_RULES has %d elements, want 9", len(ALL_RULES))
	}
	if len(ALL_INFRA) != 4 {
		t.Errorf("ALL_INFRA has %d elements, want 4", len(ALL_INFRA))
	}
	if len(ALL_SPECS_DIRS) != 10 {
		t.Errorf("ALL_SPECS_DIRS has %d elements, want 10", len(ALL_SPECS_DIRS))
	}
	if !containsAgentID(ALL_AGENTS, AgentIdPrimaryAgent) {
		t.Errorf("ALL_AGENTS missing %q", AgentIdPrimaryAgent)
	}
	if containsAgentID(ALL_AGENTS, AgentId("orchestrator")) {
		t.Errorf("ALL_AGENTS should not contain legacy orchestrator entry: %v", ALL_AGENTS)
	}
}

func TestIsValidSetupScope(t *testing.T) {
	t.Parallel()

	for _, s := range []SetupScope{SetupScopeGlobal, SetupScopeWorkspace, SetupScopeProject} {
		if !IsValidSetupScope(s) {
			t.Errorf("IsValidSetupScope(%q) = false, want true", s)
		}
	}
	if IsValidSetupScope("invalid") {
		t.Error("IsValidSetupScope(\"invalid\") = true, want false")
	}
}

func TestIsValidExistingSetupPolicy(t *testing.T) {
	t.Parallel()

	for _, policy := range []SetupPolicy{SetupPolicyAbsorb, SetupPolicyAdapt, SetupPolicyBackupOnly} {
		if !IsValidSetupPolicy(policy) {
			t.Errorf("IsValidSetupPolicy(%q) = false, want true", policy)
		}
	}
	if IsValidSetupPolicy("invalid") {
		t.Error("IsValidSetupPolicy(\"invalid\") = true, want false")
	}
}

func TestIsValidToolId(t *testing.T) {
	t.Parallel()

	for _, id := range []ToolId{ToolIdOpenCode, ToolIdClaudeCode, ToolIdCopilot} {
		if !IsValidToolId(id) {
			t.Errorf("IsValidToolId(%q) = false, want true", id)
		}
	}
	for _, id := range []ToolId{"gemini", "codex", "pi"} {
		if IsValidToolId(id) {
			t.Errorf("IsValidToolId(%q) = true, want false", id)
		}
	}
	if IsValidToolId("invalid") {
		t.Error("IsValidToolId(\"invalid\") = true, want false")
	}
}

func TestParseOperationResult(t *testing.T) {
	t.Parallel()

	for _, s := range []string{"success", "partial", "failure"} {
		r, err := ParseOperationResult(s)
		if err != nil {
			t.Errorf("ParseOperationResult(%q) returned error: %v", s, err)
		}
		if string(r) != s {
			t.Errorf("ParseOperationResult(%q) = %q, want %q", s, r, s)
		}
	}

	_, err := ParseOperationResult("invalid")
	if err == nil {
		t.Error("ParseOperationResult(\"invalid\") should return error")
	}
}

func TestParseFileOwner(t *testing.T) {
	t.Parallel()

	for _, s := range []string{"library", "user", "migrated"} {
		o, err := ParseFileOwner(s)
		if err != nil {
			t.Errorf("ParseFileOwner(%q) returned error: %v", s, err)
		}
		if string(o) != s {
			t.Errorf("ParseFileOwner(%q) = %q, want %q", s, o, s)
		}
	}

	_, err := ParseFileOwner("invalid")
	if err == nil {
		t.Error("ParseFileOwner(\"invalid\") should return error")
	}
}

func TestParseFileStatus(t *testing.T) {
	t.Parallel()

	for _, s := range []string{"installed", "modified", "missing", "conflict"} {
		st, err := ParseFileStatus(s)
		if err != nil {
			t.Errorf("ParseFileStatus(%q) returned error: %v", s, err)
		}
		if string(st) != s {
			t.Errorf("ParseFileStatus(%q) = %q, want %q", s, st, s)
		}
	}

	_, err := ParseFileStatus("invalid")
	if err == nil {
		t.Error("ParseFileStatus(\"invalid\") should return error")
	}
}

func containsAgentID(items []AgentId, want AgentId) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
