package scaffold

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/rluisb/lazyai/packages/cli/internal/compiler"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// TestScaffoldCompiledRoot_GlobalRequiresHomeDir verifies that passing an empty
// HomeDir at global scope returns an error rather than falling through to the
// real os.UserHomeDir() (R-3 mitigation from spec 008 risks).
func TestScaffoldCompiledRoot_GlobalRequiresHomeDir(t *testing.T) {
	err := ScaffoldCompiledRoot(ScaffoldCompiledRootOptions{
		TargetDir:  t.TempDir(),
		HomeDir:    "", // intentionally empty
		SetupScope: types.SetupScopeGlobal,
		Tools:      []types.ToolId{types.ToolIdClaudeCode},
	})
	if err == nil {
		t.Fatal("expected error for empty HomeDir at global scope, got nil")
	}
	if !strings.Contains(err.Error(), "HomeDir must be set") {
		t.Errorf("expected 'HomeDir must be set' message, got: %v", err)
	}
}

func TestMemoryDocDestPath(t *testing.T) {
	target := t.TempDir()
	workspaceRoot := t.TempDir()
	home := t.TempDir()

	cases := []struct {
		name  string
		tool  types.ToolId
		scope types.SetupScope
		want  string
		unsup bool
	}{
		// claude-code
		{"claude_project", types.ToolIdClaudeCode, types.SetupScopeProject, filepath.Join(target, "AGENTS.md"), false},
		{"claude_workspace", types.ToolIdClaudeCode, types.SetupScopeWorkspace, filepath.Join(workspaceRoot, "AGENTS.md"), false},
		{"claude_global", types.ToolIdClaudeCode, types.SetupScopeGlobal, filepath.Join(home, ".claude", "AGENTS.md"), false},
		// opencode
		{"opencode_project", types.ToolIdOpenCode, types.SetupScopeProject, filepath.Join(target, "AGENTS.md"), false},
		{"opencode_global", types.ToolIdOpenCode, types.SetupScopeGlobal, filepath.Join(home, ".config", "opencode", "AGENTS.md"), false},
		// copilot
		{"copilot_project", types.ToolIdCopilot, types.SetupScopeProject, filepath.Join(target, ".github", "copilot-instructions.md"), false},
		{"copilot_workspace", types.ToolIdCopilot, types.SetupScopeWorkspace, filepath.Join(workspaceRoot, ".github", "copilot-instructions.md"), false},
		{"copilot_global", types.ToolIdCopilot, types.SetupScopeGlobal, "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			outputFile := RootFileByTool[c.tool]
			got, err := memoryDocDestPath(c.tool, c.scope, target, workspaceRoot, home, outputFile)
			if c.unsup {
				if !errors.Is(err, errMemoryDocScopeUnsupported) {
					t.Fatalf("want errMemoryDocScopeUnsupported, got err=%v path=%q", err, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestScaffoldCompiledRootClaudeGeneratesClaudeMd(t *testing.T) {
	targetDir := t.TempDir()
	records := []types.TrackedFile{}

	if err := ScaffoldCompiledRoot(ScaffoldCompiledRootOptions{
		TargetDir:        targetDir,
		LibraryFS:        rootTestFS(),
		Tools:            []types.ToolId{types.ToolIdClaudeCode},
		ProjectName:      "test-project",
		FileRecords:      &records,
		Strategy:         types.ConflictStrategySkip,
		PerFileOverrides: map[string]types.ConflictStrategy{},
		SetupScope:       types.SetupScopeProject,
	}); err != nil {
		t.Fatalf("ScaffoldCompiledRoot: %v", err)
	}

	if _, err := os.Stat(filepath.Join(targetDir, "AGENTS.md")); err != nil {
		t.Fatalf("expected AGENTS.md: %v", err)
	}
	// FR-012: Claude Code target must get a native CLAUDE.md (AGENTS.md alone
	// is not enough). It single-sources via the @AGENTS.md import.
	claudeBytes, err := os.ReadFile(filepath.Join(targetDir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("expected generated CLAUDE.md: %v", err)
	}
	if !strings.Contains(string(claudeBytes), "@AGENTS.md") {
		t.Fatalf("CLAUDE.md must import AGENTS.md; got:\n%s", claudeBytes)
	}
	tracked := false
	for _, f := range records {
		if f.Path == "CLAUDE.md" {
			tracked = true
		}
	}
	if !tracked {
		t.Fatalf("generated CLAUDE.md must be tracked; records=%v", records)
	}
}

func TestScaffoldCompiledRootAntigravityGeneratesGeminiMd(t *testing.T) {
	targetDir := t.TempDir()
	records := []types.TrackedFile{}

	if err := ScaffoldCompiledRoot(ScaffoldCompiledRootOptions{
		TargetDir:        targetDir,
		LibraryFS:        rootTestFS(),
		Tools:            []types.ToolId{types.ToolIdAntigravity},
		ProjectName:      "test-project",
		FileRecords:      &records,
		Strategy:         types.ConflictStrategySkip,
		PerFileOverrides: map[string]types.ConflictStrategy{},
		SetupScope:       types.SetupScopeProject,
	}); err != nil {
		t.Fatalf("ScaffoldCompiledRoot: %v", err)
	}

	if _, err := os.Stat(filepath.Join(targetDir, "AGENTS.md")); err != nil {
		t.Fatalf("expected AGENTS.md: %v", err)
	}
	// #486 gap 2: Antigravity/Gemini target must get a native GEMINI.md (Gemini
	// CLI does not discover a bare root AGENTS.md). It single-sources via the
	// @./AGENTS.md import.
	geminiBytes, err := os.ReadFile(filepath.Join(targetDir, "GEMINI.md"))
	if err != nil {
		t.Fatalf("expected generated GEMINI.md: %v", err)
	}
	if !strings.Contains(string(geminiBytes), "@./AGENTS.md") {
		t.Fatalf("GEMINI.md must import AGENTS.md; got:\n%s", geminiBytes)
	}
	tracked := false
	for _, f := range records {
		if f.Path == "GEMINI.md" {
			tracked = true
		}
	}
	if !tracked {
		t.Fatalf("generated GEMINI.md must be tracked; records=%v", records)
	}
}

func TestScaffoldCompiledRootAppendsClaudeAgentsReferenceOnce(t *testing.T) {
	targetDir := t.TempDir()
	claudePath := filepath.Join(targetDir, "CLAUDE.md")
	if err := os.WriteFile(claudePath, []byte("# Existing Claude\n"), 0o644); err != nil {
		t.Fatalf("seed CLAUDE.md: %v", err)
	}

	for i := 0; i < 2; i++ {
		records := []types.TrackedFile{}
		if err := ScaffoldCompiledRoot(ScaffoldCompiledRootOptions{
			TargetDir:        targetDir,
			LibraryFS:        rootTestFS(),
			Tools:            []types.ToolId{types.ToolIdClaudeCode},
			ProjectName:      "test-project",
			FileRecords:      &records,
			Strategy:         types.ConflictStrategySkip,
			PerFileOverrides: map[string]types.ConflictStrategy{},
			SetupScope:       types.SetupScopeProject,
		}); err != nil {
			t.Fatalf("ScaffoldCompiledRoot pass %d: %v", i, err)
		}
	}

	got, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	content := string(got)
	if !strings.Contains(content, "# Existing Claude") {
		t.Fatalf("existing content was not preserved: %q", content)
	}
	// The append path must inject the FUNCTIONAL @import (not a markdown
	// comment), and it must appear exactly once after two compiles (#496).
	if count := strings.Count(content, claudeImportToken); count != 1 {
		t.Fatalf("functional import count = %d, want 1\n%s", count, content)
	}
}

func TestScaffoldCompiledRootAppendsGeminiAgentsReferenceOnce(t *testing.T) {
	targetDir := t.TempDir()
	geminiPath := filepath.Join(targetDir, "GEMINI.md")
	if err := os.WriteFile(geminiPath, []byte("# Existing Gemini\n"), 0o644); err != nil {
		t.Fatalf("seed GEMINI.md: %v", err)
	}

	for i := 0; i < 2; i++ {
		records := []types.TrackedFile{}
		if err := ScaffoldCompiledRoot(ScaffoldCompiledRootOptions{
			TargetDir:        targetDir,
			LibraryFS:        rootTestFS(),
			Tools:            []types.ToolId{types.ToolIdAntigravity},
			ProjectName:      "test-project",
			FileRecords:      &records,
			Strategy:         types.ConflictStrategySkip,
			PerFileOverrides: map[string]types.ConflictStrategy{},
			SetupScope:       types.SetupScopeProject,
		}); err != nil {
			t.Fatalf("ScaffoldCompiledRoot pass %d: %v", i, err)
		}
	}

	got, err := os.ReadFile(geminiPath)
	if err != nil {
		t.Fatalf("read GEMINI.md: %v", err)
	}
	content := string(got)
	if !strings.Contains(content, "# Existing Gemini") {
		t.Fatalf("existing content was not preserved: %q", content)
	}
	// The append path must inject the FUNCTIONAL @import (not a markdown
	// comment), and it must appear exactly once after two compiles (#496).
	if count := strings.Count(content, geminiImportToken); count != 1 {
		t.Fatalf("functional import count = %d, want 1\n%s", count, content)
	}
}

// TestScaffoldCompiledRootContextDocIdempotent recompiles a freshly-generated
// context doc and asserts (a) byte-identical content after the second compile,
// (b) a single functional import line, and (c) no spurious extra TrackedFile
// churn for that file (#496).
func TestScaffoldCompiledRootContextDocIdempotent(t *testing.T) {
	cases := []struct {
		name     string
		tool     types.ToolId
		filename string
		token    string
	}{
		{"claude", types.ToolIdClaudeCode, "CLAUDE.md", claudeImportToken},
		{"gemini", types.ToolIdAntigravity, "GEMINI.md", geminiImportToken},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			targetDir := t.TempDir()
			docPath := filepath.Join(targetDir, c.filename)

			// First compile: generates the fresh context doc.
			records1 := []types.TrackedFile{}
			if err := ScaffoldCompiledRoot(ScaffoldCompiledRootOptions{
				TargetDir:        targetDir,
				LibraryFS:        rootTestFS(),
				Tools:            []types.ToolId{c.tool},
				ProjectName:      "test-project",
				FileRecords:      &records1,
				Strategy:         types.ConflictStrategySkip,
				PerFileOverrides: map[string]types.ConflictStrategy{},
				SetupScope:       types.SetupScopeProject,
			}); err != nil {
				t.Fatalf("first compile: %v", err)
			}
			firstBytes, err := os.ReadFile(docPath)
			if err != nil {
				t.Fatalf("read after first compile: %v", err)
			}
			if count := strings.Count(string(firstBytes), c.token); count != 1 {
				t.Fatalf("first compile: import count = %d, want 1\n%s", count, firstBytes)
			}
			firstHashes := contextDocTrackedHashes(records1, c.filename)
			if len(firstHashes) != 1 {
				t.Fatalf("first compile: expected exactly 1 tracked record for %s, got %d (%v)", c.filename, len(firstHashes), firstHashes)
			}

			// Second compile: must be a true no-op on the context doc.
			records2 := []types.TrackedFile{}
			if err := ScaffoldCompiledRoot(ScaffoldCompiledRootOptions{
				TargetDir:        targetDir,
				LibraryFS:        rootTestFS(),
				Tools:            []types.ToolId{c.tool},
				ProjectName:      "test-project",
				FileRecords:      &records2,
				Strategy:         types.ConflictStrategySkip,
				PerFileOverrides: map[string]types.ConflictStrategy{},
				SetupScope:       types.SetupScopeProject,
			}); err != nil {
				t.Fatalf("second compile: %v", err)
			}
			secondBytes, err := os.ReadFile(docPath)
			if err != nil {
				t.Fatalf("read after second compile: %v", err)
			}
			if !bytes.Equal(firstBytes, secondBytes) {
				t.Fatalf("second compile changed %s (must be byte-identical)\n--- first ---\n%s\n--- second ---\n%s", c.filename, firstBytes, secondBytes)
			}
			if count := strings.Count(string(secondBytes), c.token); count != 1 {
				t.Fatalf("second compile: import count = %d, want 1 (no duplicate)\n%s", count, secondBytes)
			}
			secondHashes := contextDocTrackedHashes(records2, c.filename)
			if len(secondHashes) != 1 {
				t.Fatalf("second compile: expected exactly 1 tracked record for %s, got %d (%v)", c.filename, len(secondHashes), secondHashes)
			}
			// The recorded hash must not churn: the same file content yields the
			// same hash on both compiles.
			if firstHashes[0] != secondHashes[0] {
				t.Fatalf("tracked hash churn for %s: first=%s second=%s", c.filename, firstHashes[0], secondHashes[0])
			}
		})
	}
}

// contextDocTrackedHashes returns the hashes of all TrackedFile records whose
// path ends with the given context-doc filename (e.g. CLAUDE.md / GEMINI.md).
func contextDocTrackedHashes(records []types.TrackedFile, filename string) []string {
	var hashes []string
	for _, r := range records {
		if filepath.Base(r.Path) == filename {
			hashes = append(hashes, r.Hash)
		}
	}
	return hashes
}

func TestScaffoldCompiledRootCopilotUsesRootTemplate(t *testing.T) {
	targetDir := t.TempDir()
	records := []types.TrackedFile{}

	if err := ScaffoldCompiledRoot(ScaffoldCompiledRootOptions{
		TargetDir:        targetDir,
		LibraryFS:        rootTestFS(),
		Tools:            []types.ToolId{types.ToolIdCopilot},
		ProjectName:      "test-project",
		FileRecords:      &records,
		Strategy:         types.ConflictStrategySkip,
		PerFileOverrides: map[string]types.ConflictStrategy{},
		SetupScope:       types.SetupScopeProject,
	}); err != nil {
		t.Fatalf("ScaffoldCompiledRoot: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(targetDir, ".github", "copilot-instructions.md"))
	if err != nil {
		t.Fatalf("read copilot instructions: %v", err)
	}
	if !strings.Contains(string(got), "Copilot root template") {
		t.Fatalf("copilot instructions did not use root template: %q", string(got))
	}
}

func TestBuildTargetedAgentsUpdatePatchPreservesHandEditedContent(t *testing.T) {
	coverageThreshold := 87
	existing := strings.Join([]string{
		"# Hand Edited Agents",
		"",
		"## Custom Runbook",
		"Do not remove this paragraph.",
		"",
		"## Conventions",
		"",
		"### Naming",
		"- [YOUR_NAMING_CONVENTION]",
		"",
		"### Imports",
		"- [YOUR_IMPORT_ORDER]",
		"",
		"## Project Overview",
		"",
		"[YOUR_PROJECT_OVERVIEW]",
		"",
		"**Stack:**",
		"- Language: [YOUR_LANGUAGE]",
		"- Framework: [YOUR_FRAMEWORK]",
		"- Testing: [YOUR_TEST_FRAMEWORK]",
		"- Package manager: [YOUR_PACKAGE_MANAGER]",
		"",
		"## Do NOT",
		"",
		"- Never push directly to `[YOUR_PROTECTED_BRANCH]`",
		"- Keep this user-authored rule.",
		"",
		"## Testing",
		"",
		"- Minimum coverage: `[YOUR_COVERAGE_THRESHOLD]`%",
		"",
	}, "\n")

	updated, patch := BuildTargetedAgentsUpdatePatch("AGENTS.md", existing, compiler.FragmentContext{
		Constitution: &compiler.ConstitutionContext{
			ProjectOverview: "Acme checkout orchestrates creator payments.",
			Stack: compiler.ConstitutionStack{
				Language:       "TypeScript",
				Framework:      "Next.js",
				Testing:        "Vitest",
				PackageManager: "yarn",
			},
			Conventions: compiler.ConstitutionConventions{
				Naming:      "camelCase for values and PascalCase for React components",
				ImportOrder: "External packages, internal aliases, relative imports",
			},
			ProtectedBranch:   "main",
			CoverageThreshold: &coverageThreshold,
		},
	})

	mustContain(t, updated, "## Custom Runbook\nDo not remove this paragraph.")
	mustContain(t, updated, "- Keep this user-authored rule.")
	if strings.Contains(updated, "## Decision Tree") {
		t.Fatalf("targeted update full-re-emitted AGENTS.md instead of preserving existing structure:\n%s", updated)
	}
	for _, want := range []string{
		"Acme checkout orchestrates creator payments.",
		"- Language: TypeScript",
		"- Framework: Next.js",
		"- Testing: Vitest",
		"- Package manager: yarn",
		"- camelCase for values and PascalCase for React components",
		"- External packages, internal aliases, relative imports",
		"- Never push directly to `main`",
		"- Minimum coverage: `87`%",
	} {
		mustContain(t, updated, want)
	}
	if !patch.PreservedUnrecognizedContent {
		t.Fatal("patch should report preservedUnrecognizedContent=true")
	}
	if patch.File != "AGENTS.md" {
		t.Fatalf("patch file = %q, want AGENTS.md", patch.File)
	}
	if len(patch.Replacements) < 8 {
		t.Fatalf("expected targeted replacements, got %#v", patch.Replacements)
	}
}

func TestTargetedUpdatePatchJSONShape(t *testing.T) {
	updated, patch := BuildTargetedAgentsUpdatePatch("AGENTS.md", "## Project Overview\n\n[YOUR_PROJECT_OVERVIEW]\n", compiler.FragmentContext{
		Constitution: &compiler.ConstitutionContext{ProjectOverview: "Updated overview"},
	})
	if !strings.Contains(updated, "Updated overview") {
		t.Fatalf("expected updated text, got %q", updated)
	}

	data, err := json.Marshal(patch)
	if err != nil {
		t.Fatalf("marshal patch: %v", err)
	}
	var shape map[string]any
	if err := json.Unmarshal(data, &shape); err != nil {
		t.Fatalf("unmarshal patch: %v", err)
	}
	for _, key := range []string{"file", "replacements", "warnings", "preservedUnrecognizedContent"} {
		if _, ok := shape[key]; !ok {
			t.Fatalf("patch JSON missing key %q: %s", key, data)
		}
	}
	replacements := shape["replacements"].([]any)
	if len(replacements) != 1 {
		t.Fatalf("expected one replacement, got %s", data)
	}
	replacement := replacements[0].(map[string]any)
	for _, key := range []string{"field", "oldText", "newText", "location"} {
		if _, ok := replacement[key]; !ok {
			t.Fatalf("replacement JSON missing key %q: %s", key, data)
		}
	}
	location := replacement["location"].(map[string]any)
	for _, key := range []string{"section", "lineStart", "lineEnd"} {
		if _, ok := location[key]; !ok {
			t.Fatalf("location JSON missing key %q: %s", key, data)
		}
	}
}

func TestBuildTargetedAgentsUpdatePatchTargetedReplacementDoesNotReprocessInsertedPlaceholders(t *testing.T) {
	existing := "## Project Overview\n\n[YOUR_PROJECT_OVERVIEW]\n[YOUR_PROJECT_OVERVIEW]\n"
	updated, patch := BuildTargetedAgentsUpdatePatch("AGENTS.md", existing, compiler.FragmentContext{
		Constitution: &compiler.ConstitutionContext{ProjectOverview: "Project [YOUR_PROJECT_OVERVIEW]"},
	})

	want := "## Project Overview\n\nProject [YOUR_PROJECT_OVERVIEW]\nProject [YOUR_PROJECT_OVERVIEW]\n"
	if updated != want {
		t.Fatalf("targeted replacement should replace each original placeholder exactly once\nwant:%q\n got:%q", want, updated)
	}
	if len(patch.Replacements) != 2 {
		t.Fatalf("replacement count = %d, want 2: %#v", len(patch.Replacements), patch.Replacements)
	}
	if strings.Contains(updated, "Project Project") {
		t.Fatalf("replacement reprocessed inserted placeholder text: %q", updated)
	}
}

func TestBuildTargetedAgentsUpdatePatchWarnsAndLeavesUnsafeSlots(t *testing.T) {
	existing := "## Project Overview\n\nA human-authored overview.\n\n**Stack:**\n- Language: Ruby maintained by platform\n"
	updated, patch := BuildTargetedAgentsUpdatePatch("AGENTS.md", existing, compiler.FragmentContext{
		Constitution: &compiler.ConstitutionContext{
			ProjectOverview: "Generated overview",
			Stack:           compiler.ConstitutionStack{Language: "Go"},
		},
	})

	if updated != existing {
		t.Fatalf("unsafe parse should leave content unchanged\nwant:%q\n got:%q", existing, updated)
	}
	if len(patch.Replacements) != 0 {
		t.Fatalf("unsafe parse should not report replacements: %#v", patch.Replacements)
	}
	if len(patch.Warnings) < 2 {
		t.Fatalf("expected warnings for unsafe overview and language slots, got %#v", patch.Warnings)
	}
}

func TestScaffoldCompiledRootPatchesExistingAgentsWithoutFullReemit(t *testing.T) {
	targetDir := t.TempDir()
	agentsPath := filepath.Join(targetDir, "AGENTS.md")
	existing := "# Existing\n\n## Custom Section\nKeep me.\n\n## Project Overview\n\n[YOUR_PROJECT_OVERVIEW]\n"
	if err := os.WriteFile(agentsPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("seed AGENTS.md: %v", err)
	}
	records := []types.TrackedFile{}

	if err := ScaffoldCompiledRoot(ScaffoldCompiledRootOptions{
		TargetDir:        targetDir,
		LibraryFS:        rootTestFS(),
		Tools:            []types.ToolId{types.ToolIdOpenCode},
		ProjectName:      "test-project",
		ProjectOverview:  "Updated overview",
		FileRecords:      &records,
		Strategy:         types.ConflictStrategyAlign,
		PerFileOverrides: map[string]types.ConflictStrategy{},
		SetupScope:       types.SetupScopeProject,
	}); err != nil {
		t.Fatalf("ScaffoldCompiledRoot: %v", err)
	}

	got, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	content := string(got)
	mustContain(t, content, "## Custom Section\nKeep me.")
	mustContain(t, content, "Updated overview")
	if strings.Contains(content, "# AGENTS") {
		t.Fatalf("existing AGENTS.md was full-re-emitted:\n%s", content)
	}
}

func TestTestingRuleTemplateIncludesCoverageThresholdSubstitution(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "library", "rules", "testing.md"))
	if err != nil {
		t.Fatalf("read testing rule template: %v", err)
	}
	mustContain(t, string(data), "{{COVERAGE_THRESHOLD}}")
}

func TestW1AFieldParitySnapshot(t *testing.T) {
	coverageThreshold := 87
	ctx := compiler.FragmentContext{
		ProjectName: "creator-checkout",
		Constitution: &compiler.ConstitutionContext{
			ProjectOverview: "Creator Checkout processes creator payments.",
			Stack: compiler.ConstitutionStack{
				Language:       "TypeScript",
				Framework:      "Next.js",
				Database:       "PostgreSQL",
				ORM:            "Prisma",
				Testing:        "Vitest",
				PackageManager: "yarn",
			},
			Conventions: compiler.ConstitutionConventions{
				Naming:        "camelCase values; PascalCase React components",
				ErrorHandling: "Return typed Result values at service boundaries",
				APIResponses:  "JSON envelopes include data or error",
				ImportOrder:   "External, internal aliases, relative imports",
			},
			Commands: compiler.ConstitutionCommands{
				Test:  "yarn test",
				Lint:  "yarn lint",
				Build: "yarn build",
			},
			ProtectedBranch:   "main",
			CoverageThreshold: &coverageThreshold,
			CodebaseMap: []compiler.CodebaseMapEntry{
				{Path: "src"},
				{Path: "packages/api", Responsibility: "API package"},
			},
		},
	}
	template := strings.Join([]string{
		"{{PROJECT_NAME}}",
		"{{PROJECT_OVERVIEW}}",
		"{{LANGUAGE}}|{{FRAMEWORK}}|{{DATABASE}}|{{ORM}}|{{TEST_FRAMEWORK}}|{{PACKAGE_MANAGER}}",
		"{{NAMING_CONVENTIONS}}|{{ERROR_HANDLING}}|{{API_CONVENTIONS}}|{{IMPORT_ORDER}}",
		"{{TEST_COMMAND}}|{{LINT_COMMAND}}|{{BUILD_COMMAND}}|{{PROTECTED_BRANCH}}|{{COVERAGE_THRESHOLD}}",
		"{{CODEBASE_MAP}}",
	}, "\n")

	got := compiler.NewFragmentResolver("").Resolve(template, ctx)
	want := strings.Join([]string{
		"creator-checkout",
		"Creator Checkout processes creator payments.",
		"TypeScript|Next.js|PostgreSQL|Prisma|Vitest|yarn",
		"camelCase values; PascalCase React components|Return typed Result values at service boundaries|JSON envelopes include data or error|External, internal aliases, relative imports",
		"yarn test|yarn lint|yarn build|main|87",
		"| src | [WHAT_IT_DOES] |",
		"| packages/api | API package |",
	}, "\n")
	if got != want {
		t.Fatalf("W1.A field snapshot mismatch\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestW1ACompiledRootAnsweredProfileHasNoCollectedFallbackMarkers(t *testing.T) {
	// AC-N8-001, AC-N8-004, AC-N8-005, AC-N4-002: answered W1.A
	// profile fields render into AGENTS.md, generated codebase rows keep
	// human-owned responsibility placeholders, and collected fallback markers
	// disappear from the shared Go/TS root surface.
	targetDir := t.TempDir()
	for _, dir := range []string{"src", "node_modules", "dist", ".git", "vendor"} {
		if err := os.MkdirAll(filepath.Join(targetDir, dir), 0o755); err != nil {
			t.Fatalf("create target dir %s: %v", dir, err)
		}
	}
	coverageThreshold := 91
	records := []types.TrackedFile{}

	if err := ScaffoldCompiledRoot(ScaffoldCompiledRootOptions{
		TargetDir:         targetDir,
		LibraryFS:         os.DirFS(filepath.Join("..", "..", "library")),
		Tools:             []types.ToolId{types.ToolIdOpenCode},
		ProjectName:       "creator-checkout",
		PlanningDir:       "specs",
		ProjectOverview:   "Creator Checkout processes creator payments.",
		PrimaryLanguage:   "TypeScript",
		Framework:         "Next.js",
		Database:          "PostgreSQL",
		ORM:               "Prisma",
		TestFramework:     "Vitest",
		PackageManager:    "yarn",
		NamingConventions: "camelCase values; PascalCase React components",
		ErrorHandling:     "Return typed Result values at service boundaries",
		APIConventions:    "JSON envelopes include data or error",
		ImportOrder:       "External packages, internal aliases, relative imports",
		ProtectedBranch:   "main",
		MigrationsPath:    "prisma/migrations",
		TestPath:          "src/**/*.test.ts",
		StrictMode:        "TypeScript strict",
		InstallCommand:    "yarn install",
		TestCommand:       "yarn test",
		LintCommand:       "yarn lint",
		BuildCommand:      "yarn build",
		CoverageThreshold: coverageThreshold,
		FileRecords:       &records,
		Strategy:          types.ConflictStrategySkip,
		PerFileOverrides:  map[string]types.ConflictStrategy{},
		SetupScope:        types.SetupScopeProject,
	}); err != nil {
		t.Fatalf("ScaffoldCompiledRoot: %v", err)
	}

	contentBytes, err := os.ReadFile(filepath.Join(targetDir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	content := string(contentBytes)
	for _, want := range []string{
		"Creator Checkout processes creator payments.",
		"- Language: TypeScript",
		"- Framework: Next.js",
		"- Database: PostgreSQL",
		"- ORM/Query: Prisma",
		"- Testing: Vitest",
		"- Package manager: yarn",
		"- camelCase values; PascalCase React components",
		"- Return typed Result values at service boundaries",
		"- JSON envelopes include data or error",
		"- External packages, internal aliases, relative imports",
		"- Never modify `prisma/migrations` without explicit human approval",
		"- Never bypass `TypeScript strict`",
		"- Never push directly to `main`",
		"- Test location: `src/**/*.test.ts`",
		"- Minimum coverage: `91`%",
		"yarn install     # Install dependencies",
		"| src | [WHAT_IT_DOES] |",
	} {
		mustContain(t, content, want)
	}
	for _, marker := range []string{
		"[YOUR_PROJECT_OVERVIEW]",
		"[YOUR_LANGUAGE]",
		"[YOUR_FRAMEWORK]",
		"[YOUR_DATABASE]",
		"[YOUR_ORM]",
		"[YOUR_TEST_FRAMEWORK]",
		"[YOUR_PACKAGE_MANAGER]",
		"[YOUR_NAMING_CONVENTION]",
		"[YOUR_ERROR_PATTERN]",
		"[YOUR_API_CONVENTION]",
		"[YOUR_IMPORT_ORDER]",
		"[YOUR_PROTECTED_BRANCH]",
	} {
		if strings.Contains(content, marker) {
			t.Fatalf("collected fallback marker %s remained in AGENTS.md:\n%s", marker, content)
		}
	}
	for _, ignored := range []string{"node_modules", "dist", ".git", "vendor"} {
		if strings.Contains(content, "| "+ignored+" |") {
			t.Fatalf("ignored codebase map path %q rendered in AGENTS.md:\n%s", ignored, content)
		}
	}
}

func TestW1ACompiledRootSkippedProfilePreservesDocumentedFallbacks(t *testing.T) {
	// AC-N8-002, AC-N4-003: skipped/non-interactive fields preserve the
	// documented literal fallbacks while coverage resolves to the safe default.
	targetDir := t.TempDir()
	records := []types.TrackedFile{}

	if err := ScaffoldCompiledRoot(ScaffoldCompiledRootOptions{
		TargetDir:        targetDir,
		LibraryFS:        os.DirFS(filepath.Join("..", "..", "library")),
		Tools:            []types.ToolId{types.ToolIdOpenCode},
		ProjectName:      "skipped-profile",
		PlanningDir:      "specs",
		FileRecords:      &records,
		Strategy:         types.ConflictStrategySkip,
		PerFileOverrides: map[string]types.ConflictStrategy{},
		SetupScope:       types.SetupScopeProject,
	}); err != nil {
		t.Fatalf("ScaffoldCompiledRoot: %v", err)
	}

	contentBytes, err := os.ReadFile(filepath.Join(targetDir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	content := string(contentBytes)
	for _, fallback := range []string{
		"<!-- fill-in: project overview -->",
		"- Language: <!-- fill-in: language -->",
		"- Framework: <!-- fill-in: framework -->",
		"- Database: <!-- fill-in: database -->",
		"- ORM/Query: <!-- fill-in: ORM or query layer -->",
		"- Testing: <!-- fill-in: test framework -->",
		"- Package manager: <!-- fill-in: package manager -->",
		"<!-- fill-in: naming convention -->",
		"<!-- fill-in: error handling pattern -->",
		"<!-- fill-in: API response convention -->",
		"<!-- fill-in: import order -->",
		"<!-- fill-in: protected branch -->",
		"<!-- fill-in: migrations path -->",
		"<!-- fill-in: strict mode -->",
		"<!-- fill-in: shared path -->",
		"<!-- fill-in: test path -->",
		"<!-- fill-in: install command -->",
		"<!-- fill-in: lint command -->",
		"<!-- fill-in: test command -->",
		"<!-- fill-in: build command -->",
		"- Minimum coverage: `80`%",
	} {
		mustContain(t, content, fallback)
	}
	if strings.Contains(content, "[YOUR_") {
		t.Fatalf("skipped profile should not leave raw [YOUR_*] placeholders in generated AGENTS.md:\n%s", content)
	}
	for _, bad := range []string{"null", "undefined"} {
		if strings.Contains(content, bad) {
			t.Fatalf("skipped profile rendered %q in AGENTS.md:\n%s", bad, content)
		}
	}
}

func TestW1ATargetedUpdatePreservesCustomSectionsByteForByte(t *testing.T) {
	// AC-N8-003: update may patch recognized placeholders, but custom
	// AGENTS.md sections must survive byte-for-byte.
	customBlock := strings.Join([]string{
		"## Custom Operations Runbook",
		"",
		"  Keep indentation, spacing, and punctuation exactly.",
		"- User-owned checklist item",
		"",
	}, "\n")
	existing := strings.Join([]string{
		"# Existing Agents",
		"",
		customBlock,
		"## Project Overview",
		"",
		"[YOUR_PROJECT_OVERVIEW]",
		"",
	}, "\n")

	updated, _ := BuildTargetedAgentsUpdatePatch("AGENTS.md", existing, compiler.FragmentContext{
		Constitution: &compiler.ConstitutionContext{ProjectOverview: "Generated overview"},
	})

	if !strings.Contains(updated, customBlock) {
		t.Fatalf("custom AGENTS.md block was not preserved byte-for-byte\nwant block:%q\nupdated:%q", customBlock, updated)
	}
	mustContain(t, updated, "Generated overview")
}

func rootTestFS() fstest.MapFS {
	return fstest.MapFS{
		"root/AGENTS.template.md":               &fstest.MapFile{Data: []byte("# AGENTS {{projectName}}")},
		"root/copilot-instructions.template.md": &fstest.MapFile{Data: []byte("# Copilot root template")},
	}
}
