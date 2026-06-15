package adapter

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/library"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestCurationManifestNamesAdapterTargetsForMappedAssets(t *testing.T) {
	projectRoot := projectRootForAdapterManifestTest(t)
	manifest, err := library.ReadCurationManifest(filepath.Join(projectRoot, library.CurationManifestRelPath))
	if err != nil {
		t.Fatalf("ReadCurationManifest: %v", err)
	}

	mappedTargets := mappedAdapterTargetsBySource(t)
	for _, entry := range manifest.Entries {
		libraryRel, ok := strings.CutPrefix(entry.Path, library.LibraryRootRelPath+"/")
		if !ok {
			continue
		}
		for source, tools := range mappedTargets {
			if libraryRel != source && !strings.HasPrefix(libraryRel, source+"/") {
				continue
			}
			for tool := range tools {
				if !containsAdapterTarget(entry.AdapterTargets, string(tool)) {
					t.Errorf("%s adapter_targets=%v, want %q for output mapping source %s", entry.Path, entry.AdapterTargets, tool, source)
				}
			}
		}
	}
}

func mappedAdapterTargetsBySource(t *testing.T) map[string]map[types.ToolId]bool {
	t.Helper()
	mappedKinds := []AssetKind{AssetKindAgents, AssetKindSkills, AssetKindTemplates}
	mappedTargets := map[string]map[types.ToolId]bool{}
	for _, tool := range []types.ToolId{types.ToolIdClaudeCode, types.ToolIdOpenCode, types.ToolIdCopilot} {
		targets, err := OutputTargetsForTool(tool)
		if err != nil {
			t.Fatalf("OutputTargetsForTool(%q): %v", tool, err)
		}
		for _, kind := range mappedKinds {
			target := targets[kind]
			if target.Shape == ShapeNone || target.SourceSubdir == "" {
				continue
			}
			if mappedTargets[target.SourceSubdir] == nil {
				mappedTargets[target.SourceSubdir] = map[types.ToolId]bool{}
			}
			mappedTargets[target.SourceSubdir][tool] = true
		}
	}
	return mappedTargets
}

func projectRootForAdapterManifestTest(t *testing.T) string {
	t.Helper()
	libraryDir, err := library.FindLibraryDir()
	if err != nil {
		t.Fatalf("FindLibraryDir: %v", err)
	}
	return filepath.Clean(filepath.Join(libraryDir, "..", "..", ".."))
}

func containsAdapterTarget(targets []string, want string) bool {
	for _, target := range targets {
		if target == want {
			return true
		}
	}
	return false
}
