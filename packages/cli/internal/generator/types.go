// Package generator provides artifact generators for the ai-setup project.
// Each generator produces a markdown or JSON file for a specific artifact type
// (agent, skill, workflow, domain, mode, prompt, command, template).
// Ported from the TypeScript generators in src/generators/.
package generator

import (
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// PromptQuestion represents a question to ask during interactive generation.
type PromptQuestion struct {
	Key      string
	Label    string
	Type     string // "text", "select", "multiselect"
	Options  []PromptOption
	Required bool
	Default  string
}

// PromptOption represents an option in a select prompt.
type PromptOption struct {
	Value string
	Label string
}

// GeneratorConfig holds the configuration for a generation run.
type GeneratorConfig struct {
	Name        string
	Description string
	TargetDir   string
	Force       bool
	Answers     map[string]string
}

// GeneratedFile represents a single file produced by a generator.
type GeneratedFile struct {
	Path    string
	Content string
}

// Generator is the interface each artifact generator must implement.
type Generator interface {
	Type() types.ArtifactType
	Generate(config GeneratorConfig) ([]GeneratedFile, error)
	GetPromptQuestions() []PromptQuestion
}

// ToSlug converts a string to a URL-safe slug.
func ToSlug(value string) string {
	result := make([]byte, 0, len(value))
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			result = append(result, byte(r))
		} else if r >= 'A' && r <= 'Z' {
			result = append(result, byte(r+32)) // lowercase
		} else if r == ' ' || r == '_' || r == '-' {
			result = append(result, '-')
		}
	}

	// Trim leading and trailing hyphens.
	start := 0
	end := len(result)
	for start < end && result[start] == '-' {
		start++
	}
	for end > start && result[end-1] == '-' {
		end--
	}

	return string(result[start:end])
}

// ToTitleCase converts a slug or kebab-case string to title case.
func ToTitleCase(value string) string {
	if value == "" {
		return ""
	}

	result := make([]byte, 0, len(value))
	capitalize := true

	for _, r := range value {
		if r == '-' || r == '_' || r == ' ' {
			result = append(result, ' ')
			capitalize = true
		} else if capitalize {
			if r >= 'a' && r <= 'z' {
				result = append(result, byte(r-32))
			} else {
				result = append(result, byte(r))
			}
			capitalize = false
		} else {
			result = append(result, byte(r))
		}
	}

	return string(result)
}

// getAnswer returns the answer for the given key, or the default value.
func getAnswer(answers map[string]string, key, defaultVal string) string {
	if v, ok := answers[key]; ok && v != "" {
		return v
	}
	return defaultVal
}
