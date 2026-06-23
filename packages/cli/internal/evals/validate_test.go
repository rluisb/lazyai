package evals

import (
	"strings"
	"testing"
)

func TestValidateRubric_Valid(t *testing.T) {
	content := `id: test-rubric
title: Test Rubric
description: A test rubric for validation
tags:
  - quality
  - test
criteria:
  - id: correctness
    label: Correctness
    weight: 30
    description: Does it work?
    pass: All criteria met
    fail: A requirement is not met
  - id: minimality
    label: Minimality
    weight: 20
    description: Is it minimal?
    pass: Every line serves a purpose
    fail: Diff includes unnecessary code
scoring:
  scale: 0-100
  thresholds:
    excellent: 90
    good: 75
    acceptable: 60
    needs-improvement: 40
human_gate:
  required: true
  stop_conditions:
    - "A correctness issue is unresolved"
`
	issues := ValidateRubric("test.rubric.yaml", []byte(content))
	if len(issues) != 0 {
		t.Fatalf("expected no issues for valid rubric, got %d: %v", len(issues), issues)
	}
}

func TestValidateRubric_Empty(t *testing.T) {
	issues := ValidateRubric("empty.yaml", []byte{})
	if len(issues) == 0 {
		t.Fatal("expected issues for empty rubric")
	}
	if issues[0].Message != "rubric file is empty" {
		t.Fatalf("expected 'rubric file is empty', got: %s", issues[0].Message)
	}
}

func TestValidateRubric_MissingRequiredFields(t *testing.T) {
	content := `title: Missing id and criteria
`
	issues := ValidateRubric("missing.yaml", []byte(content))
	if len(issues) == 0 {
		t.Fatal("expected issues for rubric missing required fields")
	}
	found := map[string]bool{}
	for _, issue := range issues {
		found[issue.Message] = true
	}
	for _, field := range []string{"id", "criteria"} {
		msg := "missing required field: " + field
		if !found[msg] {
			t.Fatalf("expected issue: %s", msg)
		}
	}
}

func TestValidateRubric_EmptyCriteria(t *testing.T) {
	content := `id: test
title: Test
criteria: []
`
	issues := ValidateRubric("empty-criteria.yaml", []byte(content))
	if len(issues) == 0 {
		t.Fatal("expected issues for empty criteria array")
	}
	hasEmpty := false
	for _, issue := range issues {
		if issue.Message == "field 'criteria' must have at least one entry" {
			hasEmpty = true
			break
		}
	}
	if !hasEmpty {
		t.Fatalf("expected 'criteria must have at least one entry', got: %v", issues)
	}
}

func TestValidateRubric_CriteriaMissingFields(t *testing.T) {
	content := `id: test
title: Test
criteria:
  - id: only-id
    label: Only Label
`
	issues := ValidateRubric("bad-criteria.yaml", []byte(content))
	if len(issues) == 0 {
		t.Fatal("expected issues for criteria missing required fields")
	}
	for _, field := range []string{"weight", "description", "pass", "fail"} {
		msg := "criteria[0] missing required field: " + field
		found := false
		for _, issue := range issues {
			if issue.Message == msg {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected issue: %s", msg)
		}
	}
}

func TestValidateRubric_InvalidWeight(t *testing.T) {
	content := `id: test
title: Test
criteria:
  - id: c1
    label: C1
    weight: 200
    description: Test
    pass: Pass
    fail: Fail
`
	issues := ValidateRubric("bad-weight.yaml", []byte(content))
	if len(issues) == 0 {
		t.Fatal("expected issues for weight > 100")
	}
	hasWeight := false
	for _, issue := range issues {
		if issue.Message == "criteria[0].weight must be between 1 and 100" {
			hasWeight = true
			break
		}
	}
	if !hasWeight {
		t.Fatalf("expected weight range issue, got: %v", issues)
	}
}

func TestValidateRubric_InvalidYAML(t *testing.T) {
	content := `id: test
title: Test
criteria: [invalid
`
	issues := ValidateRubric("bad-yaml.yaml", []byte(content))
	if len(issues) == 0 {
		t.Fatal("expected issues for invalid YAML")
	}
	if !strings.Contains(issues[0].Message, "invalid YAML") {
		t.Fatalf("expected 'invalid YAML' issue, got: %s", issues[0].Message)
	}
}

func TestValidateRubric_ScoringBlock(t *testing.T) {
	content := `id: test
title: Test
criteria:
  - id: c1
    label: C1
    weight: 10
    description: Test
    pass: Pass
    fail: Fail
scoring:
  scale: 0-100
`
	issues := ValidateRubric("missing-thresholds.yaml", []byte(content))
	if len(issues) == 0 {
		t.Fatal("expected issues for scoring missing thresholds")
	}
	hasThresholds := false
	for _, issue := range issues {
		if issue.Message == "scoring missing required field: thresholds" {
			hasThresholds = true
			break
		}
	}
	if !hasThresholds {
		t.Fatalf("expected 'scoring missing required field: thresholds', got: %v", issues)
	}
}

func TestValidateRubric_HumanGateBlock(t *testing.T) {
	content := `id: test
title: Test
criteria:
  - id: c1
    label: C1
    weight: 10
    description: Test
    pass: Pass
    fail: Fail
human_gate:
  required: true
`
	issues := ValidateRubric("valid-human-gate.yaml", []byte(content))
	if len(issues) != 0 {
		t.Fatalf("expected no issues for valid human_gate, got: %v", issues)
	}
}

func TestValidateRubric_HumanGateMissingRequired(t *testing.T) {
	content := `id: test
title: Test
criteria:
  - id: c1
    label: C1
    weight: 10
    description: Test
    pass: Pass
    fail: Fail
human_gate:
  stop_conditions:
    - "Something"
`
	issues := ValidateRubric("missing-required.yaml", []byte(content))
	if len(issues) == 0 {
		t.Fatal("expected issues for human_gate missing required field")
	}
	hasRequired := false
	for _, issue := range issues {
		if issue.Message == "human_gate missing required field: required" {
			hasRequired = true
			break
		}
	}
	if !hasRequired {
		t.Fatalf("expected 'human_gate missing required field: required', got: %v", issues)
	}
}
