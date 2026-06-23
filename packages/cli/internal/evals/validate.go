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

// ValidateRubric validates a rubric YAML document.
func ValidateRubric(path string, data []byte) []ValidationIssue {
	if len(strings.TrimSpace(string(data))) == 0 {
		return []ValidationIssue{{
			File:    path,
			Message: "rubric file is empty",
		}}
	}

	var doc map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return []ValidationIssue{{File: path, Message: fmt.Sprintf("invalid YAML: %v", err)}}
	}
	if doc == nil {
		return []ValidationIssue{{File: path, Message: "document is empty"}}
	}

	var issues []ValidationIssue

	// Required top-level fields
	for _, field := range []string{"id", "title"} {
		if err := requirePresentStringOrObject(path, doc, field); err != nil {
			issues = append(issues, *err)
		}
	}
	if _, ok := doc["criteria"]; !ok || doc["criteria"] == nil {
		issues = append(issues, ValidationIssue{
			File:    path,
			Message: "missing required field: criteria",
		})
	}

	// Validate id is a non-empty string
	if id, ok := doc["id"]; ok {
		if s, ok := id.(string); !ok || strings.TrimSpace(s) == "" {
			issues = append(issues, ValidationIssue{
				File:    path,
				Message: "field 'id' must be a non-empty string",
			})
		}
	}

	// Validate title is a non-empty string
	if title, ok := doc["title"]; ok {
		if s, ok := title.(string); !ok || strings.TrimSpace(s) == "" {
			issues = append(issues, ValidationIssue{
				File:    path,
				Message: "field 'title' must be a non-empty string",
			})
		}
	}

	// Validate criteria is a non-empty array of objects
	if criteria, ok := doc["criteria"]; ok {
		criteriaList, ok := criteria.([]any)
		if !ok {
			issues = append(issues, ValidationIssue{
				File:    path,
				Message: "field 'criteria' must be an array",
			})
		} else if len(criteriaList) == 0 {
			issues = append(issues, ValidationIssue{
				File:    path,
				Message: "field 'criteria' must have at least one entry",
			})
		} else {
			for i, c := range criteriaList {
				criterion, ok := c.(map[string]any)
				if !ok {
					issues = append(issues, ValidationIssue{
						File:    path,
						Message: fmt.Sprintf("criteria[%d] must be a mapping object", i),
					})
					continue
				}
				for _, field := range []string{"id", "label", "weight", "description", "pass", "fail"} {
					if _, ok := criterion[field]; !ok {
						issues = append(issues, ValidationIssue{
							File:    path,
							Message: fmt.Sprintf("criteria[%d] missing required field: %s", i, field),
						})
					}
				}
				if w, ok := criterion["weight"]; ok {
					wf, ok := w.(int)
					if !ok {
						issues = append(issues, ValidationIssue{
							File:    path,
							Message: fmt.Sprintf("criteria[%d].weight must be an integer", i),
						})
					} else if wf < 1 || wf > 100 {
						issues = append(issues, ValidationIssue{
							File:    path,
							Message: fmt.Sprintf("criteria[%d].weight must be between 1 and 100", i),
						})
					}
				}
			}
		}
	}

	// Validate scoring block if present
	if scoring, ok := doc["scoring"]; ok {
		scoringMap, ok := scoring.(map[string]any)
		if !ok {
			issues = append(issues, ValidationIssue{
				File:    path,
				Message: "field 'scoring' must be a mapping object",
			})
		} else {
			if _, ok := scoringMap["scale"]; !ok {
				issues = append(issues, ValidationIssue{
					File:    path,
					Message: "scoring missing required field: scale",
				})
			}
			if _, ok := scoringMap["thresholds"]; !ok {
				issues = append(issues, ValidationIssue{
					File:    path,
					Message: "scoring missing required field: thresholds",
				})
			}
		}
	}

	// Validate human_gate block if present
	if hg, ok := doc["human_gate"]; ok {
		hgMap, ok := hg.(map[string]any)
		if !ok {
			issues = append(issues, ValidationIssue{
				File:    path,
				Message: "field 'human_gate' must be a mapping object",
			})
		} else {
			if _, ok := hgMap["required"]; !ok {
				issues = append(issues, ValidationIssue{
					File:    path,
					Message: "human_gate missing required field: required",
				})
			}
		}
	}

	return issues
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
