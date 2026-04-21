package library

import (
	"io/fs"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/ricardoborges-teachable/ai-setup/internal/frontmatter"
)

// TestCopilotAgentsSchema verifies that every .agent.yaml file has required fields.
func TestCopilotAgentsSchema(t *testing.T) {
	libFS := GetLibraryFS()
	agentsDir := "copilot/agents"

	entries, err := fs.ReadDir(libFS, agentsDir)
	if err != nil {
		t.Fatalf("read agents dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".agent.yaml") {
			continue
		}

		path := strings.TrimPrefix(agentsDir+"/"+entry.Name(), "copilot/agents/")
		content, err := fs.ReadFile(libFS, agentsDir+"/"+entry.Name())
		if err != nil {
			t.Fatalf("read %s: %v", entry.Name(), err)
		}

		var agent map[string]any
		if err := yaml.Unmarshal(content, &agent); err != nil {
			t.Errorf("parse %s: %v", entry.Name(), err)
			continue
		}

		// Validate required fields
		if _, ok := agent["name"]; !ok {
			t.Errorf("%s: missing 'name'", path)
		}
		if name, ok := agent["name"].(string); ok && name == "" {
			t.Errorf("%s: 'name' is empty", path)
		}
		if name, ok := agent["name"].(string); ok {
			// name must be lowercase
			if name != strings.ToLower(name) {
				t.Errorf("%s: 'name' must be lowercase, got %q", path, name)
			}
		}

		if _, ok := agent["description"]; !ok {
			t.Errorf("%s: missing 'description'", path)
		}
		if desc, ok := agent["description"].(string); ok && desc == "" {
			t.Errorf("%s: 'description' is empty", path)
		}

		if _, ok := agent["prompt"]; !ok {
			t.Errorf("%s: missing 'prompt'", path)
		}
		if prompt, ok := agent["prompt"].(string); ok && strings.TrimSpace(prompt) == "" {
			t.Errorf("%s: 'prompt' is empty", path)
		}
	}
}

// TestCopilotInstructionsSchema verifies that every .instructions.md file has required frontmatter.
func TestCopilotInstructionsSchema(t *testing.T) {
	libFS := GetLibraryFS()
	instructionsDir := "copilot/instructions"

	entries, err := fs.ReadDir(libFS, instructionsDir)
	if err != nil {
		t.Fatalf("read instructions dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".instructions.md") {
			continue
		}

		path := strings.TrimPrefix(instructionsDir+"/"+entry.Name(), "copilot/instructions/")
		content, err := fs.ReadFile(libFS, instructionsDir+"/"+entry.Name())
		if err != nil {
			t.Fatalf("read %s: %v", entry.Name(), err)
		}

		fm, body := frontmatter.SplitYamlFrontmatter(string(content))
		if fm == "" {
			t.Errorf("%s: no frontmatter found", path)
			continue
		}

		var data map[string]any
		if err := yaml.Unmarshal([]byte(fm), &data); err != nil {
			t.Errorf("%s: parse frontmatter: %v", path, err)
			continue
		}

		// Validate required fields
		applyTo, ok := data["applyTo"].(string)
		if !ok || applyTo == "" {
			t.Errorf("%s: missing or empty 'applyTo'", path)
		}

		// Validate body is non-empty
		if strings.TrimSpace(body) == "" {
			t.Errorf("%s: instruction body is empty", path)
		}
	}
}
