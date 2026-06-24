package adapter

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// TestAdapters_DoNotEmitWorkflowsDirectory enforces the ADR-007 boundary:
// "LazyAI generates and validates workflow catalog assets; host tools execute."
// Workflows are docs-only (curation adapter_targets: [none]) until a target's
// native/plugin surface is source-verified. No adapter may emit a workflows/
// directory into a target tree. This is the per-target negative guard the ADR
// requires before any workflow helper implementation; it complements the
// AssetKind-level guard in TestOutputMappingDoesNotEmitWorkflowDirectories and
// the Kiro-specific assertion in TestKiroAdapter_Install_*.
func TestAdapters_DoNotEmitWorkflowsDirectory(t *testing.T) {
	adapters := map[string]ToolAdapter{
		"opencode":    &OpenCodeAdapter{},
		"claude":      &ClaudeCodeAdapter{},
		"copilot":     &CopilotAdapter{},
		"pi":          &PiAdapter{},
		"omp":         &OmpAdapter{},
		"kiro":        &KiroAdapter{},
		"antigravity": &AntigravityAdapter{},
	}

	for name, adapter := range adapters {
		t.Run(name, func(t *testing.T) {
			ctx, targetDir := createTestAdapterContext(t)

			if _, err := adapter.Install(ctx); err != nil {
				t.Fatalf("%s Install failed: %v", name, err)
			}

			err := filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() && d.Name() == "workflows" {
					rel, relErr := filepath.Rel(targetDir, path)
					if relErr != nil {
						rel = path
					}
					t.Errorf("%s emitted a workflows/ directory at %s; workflows are docs-only per ADR-007", name, rel)
				}
				return nil
			})
			if err != nil && !os.IsNotExist(err) {
				t.Fatalf("%s walk target tree: %v", name, err)
			}
		})
	}
}
