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

func TestValidateChainMatchesTypeScriptIssueSemantics(t *testing.T) {
	contracts := []SkillContract{
		{Source: "skills/dup-a.md", Name: "dup"},
		{Source: "skills/dup-b.md", Name: "dup"},
		{Source: "skills/source.md", Name: "source", Output: "source", ProducesFor: []string{"missing-target"}},
		{Source: "skills/consumer.md", Name: "consumer", Consumes: []string{"missing-artifact", ".scaffolded", "path/to/file.md"}},
		{Source: "skills/orphan.md", Name: "orphan"},
		{Source: "agents/agent.md", Name: "agent"},
		{Source: "skills/root.md", Name: "rpi"},
	}

	issues := ValidateChain(contracts)
	assertIssue(t, issues, "duplicate-name", ContractSeverityError, "skills/dup-b.md")
	assertIssue(t, issues, "missing-downstream", ContractSeverityWarn, "skills/source.md")
	assertIssue(t, issues, "missing-producer", ContractSeverityWarn, "skills/consumer.md")
	assertIssue(t, issues, "self-output", ContractSeverityWarn, "skills/source.md")
	assertIssue(t, issues, "orphan-skill", ContractSeverityWarn, "skills/orphan.md")
	assertNoIssue(t, issues, "agents/agent.md", "orphan-skill")
	assertNoIssue(t, issues, "skills/root.md", "orphan-skill")

	for _, issue := range issues {
		if issue.Code == "missing-producer" && (issue.Message == ".scaffolded" || issue.Message == "path/to/file.md") {
			t.Fatalf("path-like/scaffold refs should not produce missing-producer issue: %+v", issue)
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
