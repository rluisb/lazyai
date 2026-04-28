package adapter

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/ricardoborges-teachable/ai-setup/internal/library"
)

func TestLoadSkillContractsParsesFrontmatter(t *testing.T) {
	libFS := fstest.MapFS{
		"skills/foo.md": &fstest.MapFile{Data: []byte(`---
name: foo
output: specs/{NNN}/foo.md
consumes:
  - bar
  - baz
produces_for: [downstream]
mcp_tools: [filesystem]
---

# Foo skill
`)},
	}

	contracts, err := LoadSkillContracts(libFS)
	if err != nil {
		t.Fatalf("LoadSkillContracts: %v", err)
	}
	if len(contracts) != 1 {
		t.Fatalf("expected 1 contract, got %d", len(contracts))
	}
	c := contracts[0]
	if c.Name != "foo" {
		t.Errorf("Name=%q, want foo", c.Name)
	}
	if c.Output != "specs/{NNN}/foo.md" {
		t.Errorf("Output=%q", c.Output)
	}
	if got, want := len(c.Consumes), 2; got != want {
		t.Errorf("Consumes len=%d, want %d", got, want)
	}
	if c.Consumes[0] != "bar" || c.Consumes[1] != "baz" {
		t.Errorf("Consumes=%v", c.Consumes)
	}
	if len(c.ProducesFor) != 1 || c.ProducesFor[0] != "downstream" {
		t.Errorf("ProducesFor=%v", c.ProducesFor)
	}
	if len(c.MCPTools) != 1 || c.MCPTools[0] != "filesystem" {
		t.Errorf("MCPTools=%v", c.MCPTools)
	}
}

func TestValidateContractsDetectsMissingDownstream(t *testing.T) {
	contracts := []SkillContract{
		{Source: "skills/a.md", Name: "alpha", ProducesFor: []string{"ghost"}},
		{Source: "skills/b.md", Name: "beta"},
	}
	issues := ValidateContracts(contracts)
	if len(issues) == 0 {
		t.Fatal("expected at least one issue")
	}
	found := false
	for _, i := range issues {
		if i.Code == "missing-downstream" && i.Severity == ContractSeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected missing-downstream error, got %+v", issues)
	}
	if !HasContractErrors(issues) {
		t.Error("HasContractErrors should be true")
	}
}

func TestValidateContractsDetectsMissingProducer(t *testing.T) {
	contracts := []SkillContract{
		{Source: "skills/a.md", Name: "alpha", Consumes: []string{"missing-skill"}},
	}
	issues := ValidateContracts(contracts)
	if len(issues) == 0 {
		t.Fatal("expected at least one issue")
	}
	for _, i := range issues {
		if i.Code != "missing-producer" {
			continue
		}
		if i.Severity != ContractSeverityWarn {
			t.Errorf("missing-producer should be warn, got %s", i.Severity)
		}
		return
	}
	t.Errorf("expected missing-producer warning, got %+v", issues)
}

func TestValidateContractsAllowsPathLikeRefs(t *testing.T) {
	// Path-like consumes (containing / or .) are skipped because they're
	// real file paths, not skill names.
	contracts := []SkillContract{
		{
			Source:   "skills/a.md",
			Name:     "alpha",
			Consumes: []string{"specs/{NNN}/spec.md", ".specify/memory/constitution.md"},
		},
	}
	issues := ValidateContracts(contracts)
	for _, i := range issues {
		if i.Code == "missing-producer" {
			t.Errorf("path-like consumes should not produce missing-producer, got %+v", i)
		}
	}
}

func TestValidateContractsAcceptsKnownChain(t *testing.T) {
	contracts := []SkillContract{
		{Source: "skills/a.md", Name: "specify", ProducesFor: []string{"plan"}, Output: "spec.md"},
		{Source: "skills/b.md", Name: "plan", Consumes: []string{"specify"}, ProducesFor: []string{"tasks"}, Output: "plan.md"},
		{Source: "skills/c.md", Name: "tasks", Consumes: []string{"plan"}, Output: "tasks.md"},
	}
	issues := ValidateContracts(contracts)
	if HasContractErrors(issues) {
		t.Errorf("known chain should pass validation, got: %v", issues)
	}
}

func TestValidateContractsMissingNameIsWarn(t *testing.T) {
	contracts := []SkillContract{
		{Source: "skills/anonymous.md"}, // no name
	}
	issues := ValidateContracts(contracts)
	for _, i := range issues {
		if i.Code == "missing-name" && i.Severity == ContractSeverityWarn {
			return
		}
	}
	t.Errorf("expected missing-name warning, got %+v", issues)
}

func TestFormatContractIssuesEmptyReturnsEmpty(t *testing.T) {
	if got := FormatContractIssues(nil); got != "" {
		t.Errorf("empty issues should format empty, got %q", got)
	}
}

func TestFormatContractIssuesGroupsBySeverity(t *testing.T) {
	issues := []ContractIssue{
		{Source: "a", Severity: ContractSeverityError, Code: "e1", Message: "msg-e"},
		{Source: "b", Severity: ContractSeverityWarn, Code: "w1", Message: "msg-w"},
	}
	got := FormatContractIssues(issues)
	if got == "" {
		t.Fatal("FormatContractIssues should produce output")
	}
	if !contains(got, "contract errors: 1") {
		t.Errorf("missing error count: %s", got)
	}
	if !contains(got, "contract warnings: 1") {
		t.Errorf("missing warning count: %s", got)
	}
}

func TestLoadSkillContractsLibraryIntegration(t *testing.T) {
	// Smoke: parse the real library/skills directory and ensure the
	// production skills don't break the parser. The validator may still
	// report issues; we just want to ensure parsing succeeds.
	contracts, err := LoadSkillContracts(realLibFS(t))
	if err != nil {
		t.Fatalf("LoadSkillContracts on real lib: %v", err)
	}
	if len(contracts) == 0 {
		t.Fatal("expected at least one contract from real library")
	}
}

// realLibFS returns the on-disk library FS used by the production binary.
// Tests that exercise contract parsing against real library files use this.
func realLibFS(t *testing.T) fs.FS {
	t.Helper()
	libFS := library.GetLibraryFS()
	if libFS == nil {
		t.Fatal("library.GetLibraryFS returned nil")
	}
	return libFS
}

// contains is a local helper to avoid importing strings into the test file.
func contains(haystack, needle string) bool {
	if len(needle) == 0 {
		return true
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
