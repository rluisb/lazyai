package compiler

import (
	"testing"
	"testing/fstest"
)

func TestLoadSkillContractsParsesAndFiltersMarkdownFrontmatter(t *testing.T) {
	libFS := fstest.MapFS{
		"skills/foo.md": &fstest.MapFile{Data: []byte(`---
name: foo
output: foo-output
consumes: [bar-output]
produces_for:
  - downstream
mcp_tools: filesystem
---

# Foo
`)},
		"skills/AGENTS.md":   &fstest.MapFile{Data: []byte("---\nname: ignored-agents\n---\n")},
		"skills/_partial.md": &fstest.MapFile{Data: []byte("---\nname: ignored-partial\n---\n")},
		"skills/no-name.md":  &fstest.MapFile{Data: []byte("---\noutput: ignored\n---\n")},
		"agents/reviewer.md": &fstest.MapFile{Data: []byte("---\nname: reviewer\n---\n")},
	}

	contracts, err := LoadSkillContracts(libFS)
	if err != nil {
		t.Fatalf("LoadSkillContracts: %v", err)
	}
	if got, want := len(contracts), 2; got != want {
		t.Fatalf("len(contracts) = %d, want %d: %+v", got, want, contracts)
	}

	foo := findContract(t, contracts, "foo")
	if foo.Source != "skills/foo.md" {
		t.Errorf("Source = %q, want skills/foo.md", foo.Source)
	}
	if foo.Output != "foo-output" {
		t.Errorf("Output = %q, want foo-output", foo.Output)
	}
	if got := foo.Consumes; len(got) != 1 || got[0] != "bar-output" {
		t.Errorf("Consumes = %#v, want [bar-output]", got)
	}
	if got := foo.ProducesFor; len(got) != 1 || got[0] != "downstream" {
		t.Errorf("ProducesFor = %#v, want [downstream]", got)
	}
	if got := foo.MCPTools; len(got) != 1 || got[0] != "filesystem" {
		t.Errorf("MCPTools = %#v, want [filesystem]", got)
	}
}

func TestLoadSkillContractsRecoversContractFieldsFromMalformedFrontmatter(t *testing.T) {
	libFS := fstest.MapFS{
		"skills/upstream.md": &fstest.MapFile{Data: []byte(`
---
name: upstream
produces_for:
  - downstream
workspace:
  writes: [specs/{NNN}/tasks.md, per-task harness files in tasks/ subdirectory]
---

# Upstream
`)},
		"skills/downstream.md": &fstest.MapFile{Data: []byte(`---
name: downstream
mcp_tools: [filesystem, qmd]
workspace:
  scope: [project, workspace]
  cross_repo: true (workspace memory shared)
---

# Downstream
`)},
	}

	contracts, err := LoadSkillContracts(libFS)
	if err != nil {
		t.Fatalf("LoadSkillContracts: %v", err)
	}
	if got, want := len(contracts), 2; got != want {
		t.Fatalf("len(contracts) = %d, want %d: %+v", got, want, contracts)
	}
	upstream := findContract(t, contracts, "upstream")
	if got := upstream.ProducesFor; len(got) != 1 || got[0] != "downstream" {
		t.Fatalf("ProducesFor = %#v, want [downstream]", got)
	}

	issues := ValidateChain(contracts)
	for _, issue := range issues {
		if issue.Code == "missing-downstream" {
			t.Fatalf("recovered downstream contract should not be reported missing: %+v", issue)
		}
	}
}

func TestValidateChainMatchesTypeScriptIssueSemantics(t *testing.T) {
	contracts := []SkillContract{
		{Source: "skills/dup-a.md", Name: "dup"},
		{Source: "skills/dup-b.md", Name: "Dup"}, // case-variant of "dup" — normalized, catches duplicate
		{Source: "skills/source.md", Name: "source", Output: "source", ProducesFor: []string{"missing-target"}},
		{Source: "skills/sinks.md", Name: "sinks", ProducesFor: []string{"constitution (if new article needed)", "memory (if new discovery)", "github (PR comment / approval)", "plan-gate"}},
		{Source: "skills/consumer.md", Name: "consumer", Consumes: []string{"missing-artifact", ".scaffolded", "path/to/file.md", "plan.md"}},
		{Source: "skills/prose.md", Name: "prose", Consumes: []string{"task files"}, ProducesFor: []string{"human reviewer", "review.md"}},
		{Source: "skills/orphan.md", Name: "orphan"},
		{Source: "agents/agent.md", Name: "Agent"}, // TitleCase agent name
		{Source: "skills/root.md", Name: "rpi"},
		{Source: "skills/downstream.md", Name: "Reviewer", ProducesFor: []string{"agent"}}, // TitleCase produces_for targets lowercase agent
	}

	issues := ValidateChain(contracts)
	assertIssue(t, issues, "duplicate-name", ContractSeverityError, "skills/dup-b.md")
	assertIssue(t, issues, "missing-downstream", ContractSeverityWarn, "skills/source.md")
	assertIssue(t, issues, "missing-producer", ContractSeverityWarn, "skills/consumer.md")
	assertIssue(t, issues, "self-output", ContractSeverityWarn, "skills/source.md")
	assertIssue(t, issues, "orphan-skill", ContractSeverityWarn, "skills/orphan.md")
	assertNoIssue(t, issues, "agents/agent.md", "orphan-skill")
	assertNoIssue(t, issues, "skills/root.md", "orphan-skill")
	assertNoIssue(t, issues, "skills/prose.md", "missing-downstream")
	assertNoIssue(t, issues, "skills/prose.md", "missing-producer")
	assertNoIssue(t, issues, "skills/sinks.md", "missing-downstream")

	for _, issue := range issues {
		if issue.Code == "missing-producer" && (issue.Message == ".scaffolded" || issue.Message == "path/to/file.md" || issue.Message == "plan.md") {
			t.Fatalf("path-like/scaffold/dot-containing refs should not produce missing-producer issue: %+v", issue)
		}
	}

	// Verify case-insensitive produces_for matching: "Reviewer" produces_for "agent" should NOT
	// generate missing-downstream because "Agent" normalizes to "agent".
	for _, issue := range issues {
		if issue.Code == "missing-downstream" && issue.Source == "skills/downstream.md" {
			t.Fatalf("TitleCase produces_for target should match lowercase agent via normalization: %+v", issue)
		}
	}
}

func TestStrictContractFailureTreatsWarningsAsFailures(t *testing.T) {
	warnings := []ContractIssue{{Severity: ContractSeverityWarn, Code: "orphan-skill"}}
	if ContractValidationFails(warnings, false) {
		t.Fatal("warn-only validation should not fail without strict mode")
	}
	if !ContractValidationFails(warnings, true) {
		t.Fatal("strict validation should fail on warnings")
	}
	errors := []ContractIssue{{Severity: ContractSeverityError, Code: "duplicate-name"}}
	if !ContractValidationFails(errors, false) {
		t.Fatal("error validation should fail without strict mode")
	}
}

func findContract(t *testing.T, contracts []SkillContract, name string) SkillContract {
	t.Helper()
	for _, contract := range contracts {
		if contract.Name == name {
			return contract
		}
	}
	t.Fatalf("contract %q not found in %+v", name, contracts)
	return SkillContract{}
}

func assertIssue(t *testing.T, issues []ContractIssue, code string, severity ContractSeverity, source string) {
	t.Helper()
	for _, issue := range issues {
		if issue.Code == code && issue.Severity == severity && issue.Source == source {
			return
		}
	}
	t.Fatalf("issue (%s, %s, %s) not found in %+v", code, severity, source, issues)
}

func assertNoIssue(t *testing.T, issues []ContractIssue, source string, code string) {
	t.Helper()
	for _, issue := range issues {
		if issue.Source == source && issue.Code == code {
			t.Fatalf("unexpected issue (%s, %s) in %+v", source, code, issues)
		}
	}
}
