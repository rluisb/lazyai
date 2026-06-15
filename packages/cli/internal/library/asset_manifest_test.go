package library

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateProjectAssetManifestsCurrentLibrary(t *testing.T) {
	projectRoot := projectRootForAssetManifestTest(t)
	if err := ValidateProjectAssetManifests(projectRoot); err != nil {
		t.Fatalf("ValidateProjectAssetManifests: %v", err)
	}
}

func TestHistoricalLibraryMaterialArchivedOutOfActiveRoots(t *testing.T) {
	projectRoot := projectRootForAssetManifestTest(t)

	activePaths := []string{
		"packages/cli/library/agents/orchestrator.md",
		"packages/cli/library/skills/orchestrate.md",
		"packages/cli/library/copilot/agents/orchestrator.agent.yaml",
		"packages/cli/library/standards/starter/orchestration-patterns.md",
	}
	for _, relPath := range activePaths {
		if _, err := os.Stat(filepath.Join(projectRoot, filepath.FromSlash(relPath))); err == nil {
			t.Fatalf("historical asset %s remains in an active library root", relPath)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat active historical asset %s: %v", relPath, err)
		}
	}

	archivedPaths := []string{
		"archive/issue-244-historical-library/packages/cli/library/agents/orchestrator.md",
		"archive/issue-244-historical-library/packages/cli/library/skills/orchestrate.md",
		"archive/issue-244-historical-library/packages/cli/library/copilot/agents/orchestrator.agent.yaml",
		"archive/issue-244-historical-library/packages/cli/library/standards/starter/orchestration-patterns.md",
	}
	for _, relPath := range archivedPaths {
		if _, err := os.Stat(filepath.Join(projectRoot, filepath.FromSlash(relPath))); err != nil {
			t.Fatalf("archived historical asset %s missing: %v", relPath, err)
		}
	}
}

func TestValidateProvenanceManifestFailsMissingCoverage(t *testing.T) {
	projectRoot := t.TempDir()
	writeAssetManifestTestFile(t, projectRoot, "packages/cli/library/canonical/agents/new.md", "# New Agent\n")
	writeAssetManifestTestFile(t, projectRoot, ProvenanceManifestRelPath, "version: 1\ncanonical_root: packages/cli/library/canonical\nentries: []\nexclusions: []\n")

	err := ValidateProvenanceManifest(projectRoot)
	assertManifestErrorContains(t, err, "missing provenance manifest coverage: packages/cli/library/canonical/agents/new.md")
}

func TestValidateProvenanceManifestFailsStaleHash(t *testing.T) {
	projectRoot := t.TempDir()
	assetPath := "packages/cli/library/canonical/skills/demo.md"
	writeAssetManifestTestFile(t, projectRoot, assetPath, "# Demo Skill\n")
	writeAssetManifestTestFile(t, projectRoot, ProvenanceManifestRelPath, `version: 1
canonical_root: packages/cli/library/canonical
entries:
  - path: packages/cli/library/canonical/skills/demo.md
    kind: skill
    source_repo: lazyai-local
    source_ref: test
    source_path: packages/cli/library/canonical/skills/demo.md
    local_sha256: 0000000000000000000000000000000000000000000000000000000000000000
    mode: LazyAI-authored
    notes: test fixture
exclusions: []
`)

	err := ValidateProvenanceManifest(projectRoot)
	assertManifestErrorContains(t, err, "stale local_sha256 for "+assetPath)
}

func TestValidateCurationManifestFailsMissingCoverage(t *testing.T) {
	projectRoot := t.TempDir()
	writeAssetManifestTestFile(t, projectRoot, "packages/cli/library/rules/custom.md", "# Custom Rule\n")
	writeAssetManifestTestFile(t, projectRoot, CurationManifestRelPath, "version: 1\nlibrary_root: packages/cli/library\nentries: []\nexclusions: []\n")

	err := ValidateCurationManifest(projectRoot)
	assertManifestErrorContains(t, err, "missing curation manifest coverage: packages/cli/library/rules/custom.md")
}

func TestValidateCurationManifestRequiresAdapterTargets(t *testing.T) {
	projectRoot := t.TempDir()
	writeAssetManifestTestFile(t, projectRoot, "packages/cli/library/tool-templates/shared/root.template.md", "# Root\n")
	writeAssetManifestTestFile(t, projectRoot, CurationManifestRelPath, `version: 1
library_root: packages/cli/library
entries:
  - path: packages/cli/library/tool-templates/shared/root.template.md
    kind: template
    category: adapter-support
    adapter_targets: []
    reason_kept: shared tool root template
    reason_compressed_or_changed: test fixture
    token_rent_relevant: false
exclusions: []
`)

	err := ValidateCurationManifest(projectRoot)
	assertManifestErrorContains(t, err, "adapter_targets must name tools or none")
}

func projectRootForAssetManifestTest(t *testing.T) string {
	t.Helper()
	libraryDir := projectLibraryDir()
	if libraryDir == "" {
		t.Fatal("could not locate packages/cli/library for manifest validation")
	}
	return filepath.Clean(filepath.Join(libraryDir, "..", "..", ".."))
}

func writeAssetManifestTestFile(t *testing.T, root, relPath, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", relPath, err)
	}
}

func assertManifestErrorContains(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected manifest validation error containing %q", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("manifest error = %q, want substring %q", err.Error(), want)
	}
}
