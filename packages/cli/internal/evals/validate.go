package evals

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// EvalFileType identifies an eval artifact type.
type EvalFileType string

const (
	CaseFileType    EvalFileType = "case"
	HoldoutFileType EvalFileType = "holdout"
	RubricFileType  EvalFileType = "rubric"
)

// ValidationIssue represents one issue found while validating an eval file.
type ValidationIssue struct {
	File    string
	Message string
}

func ValidateCase(path string, data []byte) []ValidationIssue {
	return validateEvalYaml(path, data)
}

// ValidateHoldout validates an eval holdout YAML document.
func ValidateHoldout(path string, data []byte) []ValidationIssue {
	return validateEvalYaml(path, data)
}

// ValidateRubric validates a rubric markdown document.
func ValidateRubric(path string, data []byte) []ValidationIssue {
	if len(strings.TrimSpace(string(data))) == 0 {
		return []ValidationIssue{{
			File:    path,
			Message: "rubric file is empty",
		}}
	}
	return nil
}

func validateEvalYaml(path string, data []byte) []ValidationIssue {
	var doc map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return []ValidationIssue{{File: path, Message: fmt.Sprintf("invalid YAML: %v", err)}}
	}
	if doc == nil {
		return []ValidationIssue{{File: path, Message: "document is empty"}}
	}

	var issues []ValidationIssue
	for _, field := range []string{"id", "title", "input", "expected"} {
		if err := requirePresentStringOrObject(path, doc, field); err != nil {
			issues = append(issues, *err)
		}
	}
	if v, ok := doc["holdout"]; ok {
		if _, ok := v.(bool); !ok {
			issues = append(issues, ValidationIssue{
				File:    path,
				Message: "field 'holdout' must be a boolean",
			})
		}
	}
	return issues
}

func requirePresentStringOrObject(path string, doc map[string]any, field string) *ValidationIssue {
	value, ok := doc[field]
	if !ok || value == nil {
		return &ValidationIssue{File: path, Message: fmt.Sprintf("missing required field: %s", field)}
	}

	switch field {
	case "id", "title":
		s, ok := value.(string)
		if !ok || strings.TrimSpace(s) == "" {
			return &ValidationIssue{File: path, Message: fmt.Sprintf("field %q must be a non-empty string", field)}
		}
	default:
		obj, ok := value.(map[string]any)
		if !ok || obj == nil {
			return &ValidationIssue{File: path, Message: fmt.Sprintf("field %q must be a mapping object", field)}
		}
	}
	return nil
}
