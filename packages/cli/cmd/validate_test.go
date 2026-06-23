package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateAllFailsOnBadSkill(t *testing.T) {
	dir := t.TempDir()
	skillPath := filepath.Join(dir, ".ai", "skills", "diagnose.md")
	if err := os.MkdirAll(filepath.Dir(skillPath), 0o755); err != nil {
		t.Fatal(err)
	}
	// Valid name but no description -> error.
	if err := os.WriteFile(skillPath, []byte("---\nname: diagnose\n---\n# Diagnose\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	withWorkingDir(t, dir)
	t.Cleanup(func() { validateAllFlag = false; validateProfileFlag = "" })
	validateAllFlag = true

	var err error
	_, _ = captureOutput(t, func() { err = runValidateAll(nil) })
	if err == nil {
		t.Fatal("expected validate --all to fail on skill missing description")
	}
}

func TestValidateAllPassesOnCleanTree(t *testing.T) {
	dir := t.TempDir()
	skillPath := filepath.Join(dir, ".ai", "skills", "plan.md")
	if err := os.MkdirAll(filepath.Dir(skillPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skillPath, []byte("---\nname: plan\ndescription: Plan before building\n---\n# Plan\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	withWorkingDir(t, dir)
	t.Cleanup(func() { validateAllFlag = false; validateProfileFlag = "" })
	validateAllFlag = true

	var err error
	_, _ = captureOutput(t, func() { err = runValidateAll(nil) })
	if err != nil {
		t.Fatalf("expected validate --all to pass on clean tree, got: %v", err)
	}
}

func TestValidateAgentsPassesOnCanonicalShape(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, ".ai", "agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatalf("mkdir agents dir: %v", err)
	}

	agentPath := filepath.Join(agentsDir, "canonical-agent.md")
	agentContent := `---
name: canonical-agent
description: test agent
---
# System Prompt

## Rules
Use stable, minimal changes.
`
	if err := os.WriteFile(agentPath, []byte(agentContent), 0o644); err != nil {
		t.Fatalf("write agent file: %v", err)
	}

	withWorkingDir(t, dir)

	var err error
	_, _ = captureOutput(t, func() {
		err = runValidateAgents(nil, nil)
	})
	if err != nil {
		t.Fatalf("expected validation to pass, got: %v", err)
	}
}

func TestValidateAgentsFailsOnMissingFrontmatter(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, ".ai", "agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatalf("mkdir agents dir: %v", err)
	}

	agentPath := filepath.Join(agentsDir, "bad-agent.md")
	if err := os.WriteFile(agentPath, []byte(`# System Prompt
## Rules
`), 0o644); err != nil {
		t.Fatalf("write agent file: %v", err)
	}

	withWorkingDir(t, dir)

	var err error
	_, _ = captureOutput(t, func() {
		err = runValidateAgents(nil, nil)
	})
	if err == nil {
		t.Fatalf("expected validation failure for missing frontmatter")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Fatalf("expected validation failed error, got: %v", err)
	}
}

func TestValidateAgentsFailsOnMissingName(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, ".ai", "agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatalf("mkdir agents dir: %v", err)
	}

	agentPath := filepath.Join(agentsDir, "no-name.md")
	if err := os.WriteFile(agentPath, []byte(`---
description: test
---
# System Prompt
## Rules
`), 0o644); err != nil {
		t.Fatalf("write agent file: %v", err)
	}

	withWorkingDir(t, dir)

	var err error
	_, _ = captureOutput(t, func() {
		err = runValidateAgents(nil, nil)
	})
	if err == nil {
		t.Fatalf("expected validation failure for missing name")
	}
}

func TestValidateSkillsPassesOnCanonicalShape(t *testing.T) {
	dir := t.TempDir()
	skillsDir := filepath.Join(dir, ".ai", "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatalf("mkdir skills dir: %v", err)
	}

	skillContent := `---
name: review
description: Conduct rigorous code review.
---
# Review

Do the review.
`
	if err := os.WriteFile(filepath.Join(skillsDir, "review.md"), []byte(skillContent), 0o644); err != nil {
		t.Fatalf("write skill file: %v", err)
	}

	withWorkingDir(t, dir)

	var err error
	_, _ = captureOutput(t, func() {
		err = runValidateSkills(nil, nil)
	})
	if err != nil {
		t.Fatalf("expected skill validation to pass, got: %v", err)
	}
}

func TestValidateSkillsFailsOnMissingFrontmatter(t *testing.T) {
	dir := t.TempDir()
	skillsDir := filepath.Join(dir, ".ai", "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatalf("mkdir skills dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "bad.md"), []byte("# No frontmatter\n"), 0o644); err != nil {
		t.Fatalf("write skill file: %v", err)
	}

	withWorkingDir(t, dir)

	var err error
	_, _ = captureOutput(t, func() {
		err = runValidateSkills(nil, nil)
	})
	if err == nil {
		t.Fatalf("expected validation failure for missing frontmatter")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Fatalf("expected validation failed error, got: %v", err)
	}
}

func TestValidateSkillsFailsOnMissingNameField(t *testing.T) {
	dir := t.TempDir()
	skillsDir := filepath.Join(dir, ".ai", "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatalf("mkdir skills dir: %v", err)
	}
	skillContent := `---
description: missing the name field
---
# Body
`
	if err := os.WriteFile(filepath.Join(skillsDir, "noname.md"), []byte(skillContent), 0o644); err != nil {
		t.Fatalf("write skill file: %v", err)
	}

	withWorkingDir(t, dir)

	var err error
	_, _ = captureOutput(t, func() {
		err = runValidateSkills(nil, nil)
	})
	if err == nil {
		t.Fatalf("expected validation failure for missing name field")
	}
}

func TestValidateSkillsSupportsDirSkillMdLayout(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, ".ai", "skills", "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir skill dir: %v", err)
	}
	skillContent := `---
name: my-skill
description: A directory-form skill.
---
# My Skill
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	withWorkingDir(t, dir)

	var err error
	_, _ = captureOutput(t, func() {
		err = runValidateSkills(nil, nil)
	})
	if err != nil {
		t.Fatalf("expected dir-form skill validation to pass, got: %v", err)
	}
}

func TestValidateEvalsPassesOnValidCase(t *testing.T) {
	dir := t.TempDir()
	casesDir := filepath.Join(dir, ".ai", "evals", "cases")
	holdoutsDir := filepath.Join(dir, ".ai", "evals", "holdouts")
	if err := os.MkdirAll(casesDir, 0o755); err != nil {
		t.Fatalf("mkdir eval cases dir: %v", err)
	}
	if err := os.MkdirAll(holdoutsDir, 0o755); err != nil {
		t.Fatalf("mkdir eval holdouts dir: %v", err)
	}

	caseContent := `id: skill-pr-review-trigger-001
title: PR review skill triggers on staged diff review request
input:
  user: Review my staged changes before I commit.
expected:
  shouldUseSkill: pr-review
  shouldNotUseSkills:
    - deploy
`
	if err := os.WriteFile(filepath.Join(casesDir, "case-001.yaml"), []byte(caseContent), 0o644); err != nil {
		t.Fatalf("write case file: %v", err)
	}

	holdoutContent := `id: skill-pr-review-trigger-holdout-001
title: PR review holdout
input:
  user: Draft a quick patch.
expected:
  shouldUseSkill: pr-review
`
	if err := os.WriteFile(filepath.Join(holdoutsDir, "holdout-001.yaml"), []byte(holdoutContent), 0o644); err != nil {
		t.Fatalf("write holdout file: %v", err)
	}

	withWorkingDir(t, dir)

	var err error
	_, _ = captureOutput(t, func() {
		err = runValidateEvals(nil, nil)
	})
	if err != nil {
		t.Fatalf("expected eval validation to pass, got: %v", err)
	}
}

func TestValidateEvalsFailsOnInvalidCase(t *testing.T) {
	dir := t.TempDir()
	casesDir := filepath.Join(dir, ".ai", "evals", "cases")
	holdoutsDir := filepath.Join(dir, ".ai", "evals", "holdouts")
	if err := os.MkdirAll(casesDir, 0o755); err != nil {
		t.Fatalf("mkdir eval cases dir: %v", err)
	}
	if err := os.MkdirAll(holdoutsDir, 0o755); err != nil {
		t.Fatalf("mkdir eval holdouts dir: %v", err)
	}

	invalidCaseContent := `title: Missing id
input:
  user: What should I do?
expected:
  shouldUseSkill: pr-review
`
	if err := os.WriteFile(filepath.Join(casesDir, "bad-case.yaml"), []byte(invalidCaseContent), 0o644); err != nil {
		t.Fatalf("write case file: %v", err)
	}

	holdoutContent := `id: skill-pr-review-trigger-holdout-001
title: PR review holdout
input:
  user: Draft a quick patch.
expected:
  shouldUseSkill: pr-review
`
	if err := os.WriteFile(filepath.Join(holdoutsDir, "holdout-001.yaml"), []byte(holdoutContent), 0o644); err != nil {
		t.Fatalf("write holdout file: %v", err)
	}

	withWorkingDir(t, dir)

	var err error
	_, _ = captureOutput(t, func() {
		err = runValidateEvals(nil, nil)
	})
	if err == nil {
		t.Fatalf("expected eval validation failure for invalid case")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Fatalf("expected validation failed error, got: %v", err)
	}
}

func TestValidateEvalsFailsWhenEvalsDirectoryMissing(t *testing.T) {
	dir := t.TempDir()
	withWorkingDir(t, dir)

	var err error
	_, _ = captureOutput(t, func() {
		err = runValidateEvals(nil, nil)
	})
	if err == nil {
		t.Fatalf("expected failure when .ai/evals is missing")
	}
	if !strings.Contains(err.Error(), "evals directory not found") {
		t.Fatalf("expected missing-evals error, got: %v", err)
	}
}
