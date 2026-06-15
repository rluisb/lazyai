package adapter

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/frontmatter"
	"github.com/rluisb/lazyai/packages/cli/internal/jsonc"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// --- Test: OpenCode adapter with fs.FS ---

func TestOpenCodeAdapter_Install_FromFS(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.Selections = AdapterSelections{
		Agents: types.ALL_AGENTS[:],
		Skills: types.ALL_SKILLS[:],
	}

	adapter := &OpenCodeAdapter{}
	records, err := adapter.Install(ctx)
	if err != nil {
		t.Fatalf("OpenCode Install failed: %v", err)
	}

	if len(records) == 0 {
		t.Fatal("expected at least one tracked file record")
	}

	// Check that key files were created. Adapter install preserves actual agent
	// and skill definitions, but must not create reserved context docs inside the
	// tool directory; root context docs are handled elsewhere.
	keyFiles := []string{
		".opencode/opencode.jsonc",
		".opencode/package.json",
		".opencode/agents/primary-agent.md",
		".opencode/agents/builder.md",
		".opencode/skills/diagnose/SKILL.md",
	}
	// The pre-unification .json variant must never be produced on a fresh install.
	if _, err := os.Stat(filepath.Join(targetDir, ".opencode", "opencode.json")); err == nil {
		t.Error("opencode.json should not exist; install must target opencode.jsonc only")
	}
	for _, f := range keyFiles {
		path := filepath.Join(targetDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s was not created", f)
		}
	}
	if _, err := os.Stat(filepath.Join(targetDir, ".opencode", "agents", "orchestrator.md")); err == nil {
		t.Error("orchestrator.md should not be installed")
	}
	for _, f := range []string{
		".opencode/AGENTS.md",
		".opencode/agents/AGENTS.md",
		".opencode/skills/AGENTS.md",
	} {
		if _, err := os.Stat(filepath.Join(targetDir, f)); err == nil {
			t.Errorf("reserved context doc %s should not be created by adapter install", f)
		}
	}

	// Every installed agent file must carry opencode-schema-valid
	// frontmatter: at minimum `description` and `mode`. This closes the gap
	// where the old shared stripper emitted HTML comments that opencode
	// could not parse.
	agentsDir := filepath.Join(targetDir, ".opencode", "agents")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		t.Fatalf("read agents dir: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || e.Name() == "AGENTS.md" {
			continue
		}
		data, _ := os.ReadFile(filepath.Join(agentsDir, e.Name()))
		fm, _, err := frontmatter.ExtractFrontmatter(data)
		if err != nil {
			t.Errorf("%s: frontmatter does not parse: %v", e.Name(), err)
			continue
		}
		if fm["description"] == nil || fm["description"] == "" {
			t.Errorf("%s: missing description key", e.Name())
		}
		if fm["mode"] == nil || fm["mode"] == "" {
			t.Errorf("%s: missing mode key", e.Name())
		}
	}
}

// --- Test: OpenCode adapter global scope path resolution ---

func TestOpenCodeAdapter_GlobalScope_UsesGlobalPath(t *testing.T) {
	targetDir := t.TempDir()
	libFS := createTestFS()

	// Create a temp dir that mimics ~/.config/opencode structure.
	homeDir := t.TempDir()
	expectedDir := filepath.Join(homeDir, ".config", "opencode")

	ctx := &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeGlobal,
		HomeDir:    homeDir,
		LibraryDir: "",
		LibraryFS:  libFS,
		Strategy:   types.ConflictStrategyAlign,
		Selections: AdapterSelections{
			Agents: []types.AgentId{"builder"},
			Skills: []types.SkillId{types.SkillIdDiagnose},
		},
	}

	adapter := &OpenCodeAdapter{}
	records, err := adapter.Install(ctx)
	if err != nil {
		t.Fatalf("OpenCode Install in global scope failed: %v", err)
	}

	if len(records) == 0 {
		t.Fatal("expected at least one tracked file record")
	}

	// Verify the expected directory structure was created in the global dir.
	keyDirs := []string{
		filepath.Join(expectedDir, "agents"),
		filepath.Join(expectedDir, "skills"),
		filepath.Join(expectedDir, "commands"),
	}
	for _, d := range keyDirs {
		if _, err := os.Stat(d); os.IsNotExist(err) {
			t.Errorf("expected directory %s was not created", d)
		}
	}

	// Verify no .opencode directory was created in the project target dir.
	projectOpencode := filepath.Join(targetDir, ".opencode")
	if _, err := os.Stat(projectOpencode); !os.IsNotExist(err) {
		t.Error("global scope should not create .opencode in project target dir")
	}
}

func TestOpenCodeAdapter_GlobalScope_FallbackHomeDir(t *testing.T) {
	targetDir := t.TempDir()
	libFS := createTestFS()

	ctx := &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeGlobal,
		HomeDir:    "", // empty — should fall back to os.UserHomeDir()
		LibraryDir: "",
		LibraryFS:  libFS,
		Strategy:   types.ConflictStrategyAlign,
		Selections: AdapterSelections{
			Agents: []types.AgentId{"builder"},
			Skills: []types.SkillId{types.SkillIdDiagnose},
		},
	}

	adapter := &OpenCodeAdapter{}
	_, err := adapter.Install(ctx)
	if err != nil {
		t.Fatalf("OpenCode Install with empty HomeDir failed: %v", err)
	}

	// Verify files were written to the real home directory.
	realHome, _ := os.UserHomeDir()
	expectedDir := filepath.Join(realHome, ".config", "opencode")
	if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
		t.Errorf("expected global dir %s was not created", expectedDir)
	}
}

// --- Test: OpenCode adapter migrates pre-existing opencode.json to .jsonc ---

func TestOpenCodeAdapter_Install_MigratesJsonToJsonc(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.Selections = AdapterSelections{
		Agents: []types.AgentId{"builder"},
		Skills: []types.SkillId{types.SkillIdDiagnose},
	}

	// Seed a pre-existing opencode.json with a user-authored key that must
	// survive the migration unchanged.
	ocDir := filepath.Join(targetDir, ".opencode")
	if err := files.EnsureDir(ocDir); err != nil {
		t.Fatalf("EnsureDir: %v", err)
	}
	jsonPath := filepath.Join(ocDir, "opencode.json")
	jsoncPath := filepath.Join(ocDir, "opencode.jsonc")
	original := []byte(`{
  "$schema": "https://opencode.ai/config.json",
  "permission": { "edit": "allow" },
  "user_key": "preserved"
}
`)
	if err := os.WriteFile(jsonPath, original, 0o644); err != nil {
		t.Fatalf("seed opencode.json: %v", err)
	}

	adapter := &OpenCodeAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("OpenCode Install failed: %v", err)
	}

	// .jsonc must exist; .json must be gone; .json.bak must preserve original.
	if _, err := os.Stat(jsoncPath); os.IsNotExist(err) {
		t.Fatal("opencode.jsonc was not created by migration")
	}
	if _, err := os.Stat(jsonPath); err == nil {
		t.Error("opencode.json should have been removed after migration")
	}
	bakPath := jsonPath + ".bak"
	bakContents, err := os.ReadFile(bakPath)
	if err != nil {
		t.Fatalf("opencode.json.bak was not created: %v", err)
	}
	if string(bakContents) != string(original) {
		t.Errorf(".bak sidecar content mismatch.\nwant: %q\ngot:  %q", original, bakContents)
	}

	// The migrated .jsonc must carry the user-authored key forward verbatim
	// (no default-config merge on top of the migrated file — that would
	// silently clobber customizations like permission.edit == "allow").
	jsoncContents, err := os.ReadFile(jsoncPath)
	if err != nil {
		t.Fatalf("read migrated .jsonc: %v", err)
	}
	if !strings.Contains(string(jsoncContents), `"user_key": "preserved"`) {
		t.Errorf("migrated .jsonc dropped user_key:\n%s", jsoncContents)
	}
	if !strings.Contains(string(jsoncContents), `"edit": "allow"`) {
		t.Errorf("migrated .jsonc did not preserve user-authored permission.edit:\n%s", jsoncContents)
	}
}

// --- Test: opencode commands + modes install at every scope ---

func TestOpenCodeAdapter_InstallsCommandsAndModes(t *testing.T) {
	type scopeCase struct {
		name   string
		scope  types.SetupScope
		rootFn func(targetDir, homeDir string) string
	}
	cases := []scopeCase{
		{"project", types.SetupScopeProject, func(t, _ string) string { return filepath.Join(t, ".opencode") }},
		{"workspace", types.SetupScopeWorkspace, func(t, _ string) string { return filepath.Join(t, ".opencode") }},
		{"global", types.SetupScopeGlobal, func(_, h string) string { return filepath.Join(h, ".config", "opencode") }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			targetDir := t.TempDir()
			homeDir := t.TempDir()
			ctx := &AdapterContext{
				TargetDir:  targetDir,
				SetupScope: tc.scope,
				HomeDir:    homeDir,
				LibraryFS:  createTestFS(),
				Strategy:   types.ConflictStrategyAlign,
				Selections: AdapterSelections{
					Agents: []types.AgentId{"builder"},
					Skills: []types.SkillId{types.SkillIdDiagnose},
					// Leaving OpenCodeCommands / OpenCodeModes unset means
					// "install all" — the wizard will populate these later.
				},
			}

			if _, err := (&OpenCodeAdapter{}).Install(ctx); err != nil {
				t.Fatalf("Install (%s): %v", tc.name, err)
			}

			root := tc.rootFn(targetDir, homeDir)
			for _, want := range []string{
				"commands/review.md",
				"commands/test.md",
				"commands/commit.md",
				"modes/plan.md",
				"modes/audit.md",
			} {
				path := filepath.Join(root, want)
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Errorf("%s: missing %s after install", tc.name, want)
				}
			}
		})
	}
}

// TestOpenCodeAdapter_SelectionFiltersCommandsAndModes verifies that
// ctx.Selections.OpenCodeCommands narrows the install set. An explicit
// selection of ["review"] must leave test.md and commit.md uninstalled.
func TestOpenCodeAdapter_SelectionFiltersCommandsAndModes(t *testing.T) {
	targetDir := t.TempDir()
	ctx := &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
		LibraryFS:  createTestFS(),
		Strategy:   types.ConflictStrategyAlign,
		Selections: AdapterSelections{
			Agents:           []types.AgentId{"builder"},
			Skills:           []types.SkillId{types.SkillIdDiagnose},
			OpenCodeCommands: []types.OpenCodeCommandId{types.OpenCodeCommandIdReview},
			OpenCodeModes:    []types.OpenCodeModeId{types.OpenCodeModeIdPlan},
		},
	}
	if _, err := (&OpenCodeAdapter{}).Install(ctx); err != nil {
		t.Fatalf("Install: %v", err)
	}

	root := filepath.Join(targetDir, ".opencode")
	mustExist := []string{"commands/review.md", "modes/plan.md"}
	mustNotExist := []string{"commands/test.md", "commands/commit.md", "modes/audit.md"}
	for _, p := range mustExist {
		if _, err := os.Stat(filepath.Join(root, p)); os.IsNotExist(err) {
			t.Errorf("selection did not include %s: missing", p)
		}
	}
	for _, p := range mustNotExist {
		if _, err := os.Stat(filepath.Join(root, p)); err == nil {
			t.Errorf("selection leaked: %s should not exist", p)
		}
	}
}

// --- Test: instructions key resolves to a real file at every scope ---
//
// opencode resolves entries in opencode.jsonc's `instructions` array
// relative to the config file's directory. Project/workspace configs live in
// `.opencode/`, so they must point up to the root AGENTS.md. Global configs
// live alongside the global AGENTS.md and can use `AGENTS.md` directly.

func TestOpenCodeAdapter_InstructionsKeyResolves(t *testing.T) {
	type scopeCase struct {
		name  string
		scope types.SetupScope
		// rootFn returns the expected .opencode/ root given a fresh (targetDir, homeDir) pair.
		rootFn func(targetDir, homeDir string) string
	}

	cases := []scopeCase{
		{
			name:  "project",
			scope: types.SetupScopeProject,
			rootFn: func(targetDir, _ string) string {
				return filepath.Join(targetDir, ".opencode")
			},
		},
		{
			name:  "workspace",
			scope: types.SetupScopeWorkspace,
			rootFn: func(targetDir, _ string) string {
				return filepath.Join(targetDir, ".opencode")
			},
		},
		{
			name:  "global",
			scope: types.SetupScopeGlobal,
			rootFn: func(_, homeDir string) string {
				return filepath.Join(homeDir, ".config", "opencode")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			targetDir := t.TempDir()
			homeDir := t.TempDir()
			rootDoc := filepath.Join(targetDir, "AGENTS.md")
			if tc.scope == types.SetupScopeGlobal {
				rootDoc = filepath.Join(homeDir, ".config", "opencode", "AGENTS.md")
				if err := os.MkdirAll(filepath.Dir(rootDoc), 0o755); err != nil {
					t.Fatalf("create global root doc dir: %v", err)
				}
			}
			if err := os.WriteFile(rootDoc, []byte("# Root instructions\n"), 0o644); err != nil {
				t.Fatalf("write root doc: %v", err)
			}
			ctx := &AdapterContext{
				TargetDir:  targetDir,
				SetupScope: tc.scope,
				HomeDir:    homeDir,
				LibraryFS:  createTestFS(),
				Strategy:   types.ConflictStrategyAlign,
				Selections: AdapterSelections{
					Agents: []types.AgentId{"builder"},
					Skills: []types.SkillId{types.SkillIdDiagnose},
				},
			}

			if _, err := (&OpenCodeAdapter{}).Install(ctx); err != nil {
				t.Fatalf("Install (%s): %v", tc.name, err)
			}

			root := tc.rootFn(targetDir, homeDir)

			// Read back opencode.jsonc and assert instructions shape.
			cfgPath := filepath.Join(root, OpenCodeConfigFilename)
			cfg, err := jsonc.ReadJSONCFile(cfgPath)
			if err != nil {
				t.Fatalf("read %s: %v", cfgPath, err)
			}
			rawInstr, ok := cfg["instructions"]
			if !ok {
				t.Fatalf("opencode.jsonc missing `instructions` key at %s", tc.name)
			}
			instr, ok := rawInstr.([]any)
			if !ok || len(instr) == 0 {
				t.Fatalf("instructions must be a non-empty array, got %T: %v", rawInstr, rawInstr)
			}

			// Each instructions entry, resolved relative to the config dir,
			// must point at an existing, non-empty file.
			cfgDir := filepath.Dir(cfgPath)
			for _, raw := range instr {
				rel, ok := raw.(string)
				if !ok {
					t.Errorf("instructions entry is not a string: %T %v", raw, raw)
					continue
				}
				resolved := rel
				if !filepath.IsAbs(rel) {
					resolved = filepath.Join(cfgDir, rel)
				}
				info, err := os.Stat(resolved)
				if err != nil {
					t.Errorf("instructions entry %q resolves to %q which does not exist: %v", rel, resolved, err)
					continue
				}
				if info.Size() == 0 {
					t.Errorf("instructions entry %q resolves to an empty file at %q", rel, resolved)
				}
			}
		})
	}
}

// --- Test: OpenCode adapter primary-agent install ---

func testRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

func TestOpenCodeAdapter_Install_PrimaryAgentMode(t *testing.T) {
	targetDir := t.TempDir()
	repoRoot := testRepoRoot(t)
	libFS := os.DirFS(filepath.Join(repoRoot, "library"))

	ctx := &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
		LibraryFS:  libFS,
		Strategy:   types.ConflictStrategyAlign,
		Selections: AdapterSelections{},
	}

	adapter := &OpenCodeAdapter{}
	_, err := adapter.Install(ctx)
	if err != nil {
		t.Fatalf("OpenCode Install failed: %v", err)
	}

	cfgPath := filepath.Join(targetDir, ".opencode", "opencode.jsonc")
	cfg, err := jsonc.ReadJSONCFile(cfgPath)
	if err != nil {
		t.Fatalf("read opencode.jsonc: %v", err)
	}
	if da, ok := cfg["default_agent"].(string); !ok || da != "primary-agent" {
		t.Errorf("default_agent = %q, want primary-agent", da)
	}
	if _, err := os.Stat(filepath.Join(targetDir, ".opencode", "agents", "primary-agent.md")); os.IsNotExist(err) {
		t.Error("primary-agent.md was not installed")
	}
	if _, err := os.Stat(filepath.Join(targetDir, ".opencode", "agents", "orchestrator.md")); err == nil {
		t.Error("orchestrator.md should not be installed")
	}
	if _, err := os.Stat(filepath.Join(targetDir, ".opencode", "STARTUP.md")); err == nil {
		t.Error("STARTUP.md should not be installed")
	}
}
func TestOpenCodeAdapter_DefaultConfigIncludesSkillSurface(t *testing.T) {
	targetDir := t.TempDir()
	repoRoot := testRepoRoot(t)
	libFS := os.DirFS(filepath.Join(repoRoot, "library"))

	ctx := &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
		LibraryFS:  libFS,
		Strategy:   types.ConflictStrategyAlign,
	}

	adapter := &OpenCodeAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("OpenCode Install failed: %v", err)
	}

	cfgPath := filepath.Join(targetDir, ".opencode", "opencode.jsonc")
	cfg, err := jsonc.ReadJSONCFile(cfgPath)
	if err != nil {
		t.Fatalf("read opencode.jsonc: %v", err)
	}
	skills, ok := cfg["skills"].(map[string]any)
	if !ok {
		t.Fatalf("skills = %T, want object", cfg["skills"])
	}
	paths, ok := skills["paths"].([]any)
	if !ok || len(paths) != 1 || paths[0] != "skills" {
		t.Fatalf("skills.paths = %v, want [skills]", skills["paths"])
	}
	permission, ok := cfg["permission"].(map[string]any)
	if !ok {
		t.Fatalf("permission = %T, want object", cfg["permission"])
	}
	if got, _ := permission["bash"].(string); got != "ask" {
		t.Fatalf("permission.bash = %q, want ask", got)
	}
	if got, _ := permission["edit"].(string); got != "ask" {
		t.Fatalf("permission.edit = %q, want ask", got)
	}
	skillPerm, ok := permission["skill"].(map[string]any)
	if !ok {
		t.Fatalf("permission.skill = %T, want object", permission["skill"])
	}
	if got, _ := skillPerm["*"].(string); got != "allow" {
		t.Fatalf("permission.skill[*] = %q, want allow", got)
	}
}
func TestOpenCodeAdapter_PackageJSONUsesModuleType(t *testing.T) {
	targetDir := t.TempDir()
	repoRoot := testRepoRoot(t)
	libFS := os.DirFS(filepath.Join(repoRoot, "library"))

	ctx := &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
		LibraryFS:  libFS,
		Strategy:   types.ConflictStrategyAlign,
	}

	adapter := &OpenCodeAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("OpenCode Install failed: %v", err)
	}

	packageData, err := os.ReadFile(filepath.Join(targetDir, ".opencode", "package.json"))
	if err != nil {
		t.Fatalf("read package.json: %v", err)
	}
	if !strings.Contains(string(packageData), `"type": "module"`) {
		t.Fatalf("package.json missing type=module:\n%s", packageData)
	}
}
