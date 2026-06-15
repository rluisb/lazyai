package library

import (
	"errors"
	"io/fs"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/rluisb/lazyai/packages/cli/internal/frontmatter"
)

// TestCopilotAgentsFixturesArchived verifies hand-authored Copilot agent YAML
// fixtures stay out of the active library. The Copilot adapter now generates
// agent YAML from canonical markdown.
func TestCopilotAgentsFixturesArchived(t *testing.T) {
	libFS := GetLibraryFS()
	agentsDir := "copilot/agents"

	entries, err := fs.ReadDir(libFS, agentsDir)
	if err == nil && len(entries) > 0 {
		t.Fatalf("legacy Copilot agent fixtures remain active under %s", agentsDir)
	}
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("read agents dir: %v", err)
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
