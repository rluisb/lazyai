package adapter

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

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
		"opencode.json",
		".opencode/package.json",
		".opencode/agents/guide.md",
		".opencode/agents/implementer.md",
		".opencode/agents/researcher.md",
		".opencode/skills/diagnose/SKILL.md",
	}
	// The default config must be written to root opencode.json; .opencode is reserved for runtime assets.
	if _, err := os.Stat(filepath.Join(targetDir, "opencode.json")); err != nil {
		t.Fatal("root opencode.json was not created during install")
	}
	if _, err := os.Stat(filepath.Join(targetDir, ".opencode", "opencode.json")); err == nil {
		t.Error(".opencode/opencode.json should not exist on default install")
	}
	if _, err := os.Stat(filepath.Join(targetDir, ".opencode", "opencode.jsonc")); err == nil {
		t.Error(".opencode/opencode.jsonc should not exist on default install")
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
		if _, ok := fm["name"]; ok {
			t.Errorf("%s: name key should not be emitted in baseline OpenCode output", e.Name())
		}
		if _, ok := fm["mode"]; ok {
			t.Errorf("%s: mode key should not be emitted in baseline OpenCode output", e.Name())
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
			Agents: []types.AgentId{"researcher"},
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
			Agents: []types.AgentId{"researcher"},
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

// --- Test: OpenCode adapter preserves pre-existing root opencode.json ---

func TestOpenCodeAdapter_Install_PreservesRootJson(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.Selections = AdapterSelections{
		Agents: []types.AgentId{"researcher"},
		Skills: []types.SkillId{types.SkillIdDiagnose},
	}

	jsonPath := filepath.Join(targetDir, "opencode.json")
	original := []byte(`{
  "$schema": "https://opencode.ai/config.json",
  "permission": { "edit": "allow" },
  "user_key": "preserved"
}
`)
	if err := os.WriteFile(jsonPath, original, 0o644); err != nil {
		t.Fatalf("seed opencode.json: %v", err)
	}

	if _, err := (&OpenCodeAdapter{}).Install(ctx); err != nil {
		t.Fatalf("OpenCode Install failed: %v", err)
	}

	contents, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read root opencode.json: %v", err)
	}
	if string(contents) != string(original) {
		t.Errorf("root opencode.json should be preserved unchanged.\nwant: %s\ngot:  %s", original, contents)
	}
	if _, err := os.Stat(filepath.Join(targetDir, ".opencode", "opencode.json")); err == nil {
		t.Error(".opencode/opencode.json should not be created when root opencode.json exists")
	}
	if _, err := os.Stat(filepath.Join(targetDir, ".opencode", "opencode.jsonc")); err == nil {
		t.Error(".opencode/opencode.jsonc should not be created when root opencode.json exists")
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
					Agents: []types.AgentId{"researcher"},
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
			Agents:           []types.AgentId{"researcher"},
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
// opencode resolves entries in root opencode.json's `instructions` array
// relative to the config file's directory. Project/workspace configs live at
// the project/workspace root; global configs live alongside the global
// AGENTS.md and can use `AGENTS.md` directly.

func TestOpenCodeAdapter_InstructionsKeyResolves(t *testing.T) {
	type scopeCase struct {
		name  string
		scope types.SetupScope
		// rootFn returns the expected config root given a fresh (targetDir, homeDir) pair.
		rootFn func(targetDir, homeDir string) string
	}

	cases := []scopeCase{
		{
			name:  "project",
			scope: types.SetupScopeProject,
			rootFn: func(targetDir, _ string) string {
				return targetDir
			},
		},
		{
			name:  "workspace",
			scope: types.SetupScopeWorkspace,
			rootFn: func(targetDir, _ string) string {
				return targetDir
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
					Agents: []types.AgentId{"researcher"},
					Skills: []types.SkillId{types.SkillIdDiagnose},
				},
			}

			if _, err := (&OpenCodeAdapter{}).Install(ctx); err != nil {
				t.Fatalf("Install (%s): %v", tc.name, err)
			}

			root := tc.rootFn(targetDir, homeDir)

			// Read back opencode.json and assert instructions shape.
			cfgPath := filepath.Join(root, OpenCodeConfigFilename)
			cfg, err := jsonc.ReadJSONCFile(cfgPath)
			if err != nil {
				t.Fatalf("read %s: %v", cfgPath, err)
			}
			rawInstr, ok := cfg["instructions"]
			if !ok {
				t.Fatalf("opencode.json missing `instructions` key at %s", tc.name)
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

// --- Test: OpenCode adapter default agent install ---

func testRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

func TestOpenCodeAdapter_Install_DefaultAgentMode(t *testing.T) {
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

	cfgPath := filepath.Join(targetDir, OpenCodeConfigFilename)
	cfg, err := jsonc.ReadJSONCFile(cfgPath)
	if err != nil {
		t.Fatalf("read opencode.json: %v", err)
	}
	if _, ok := cfg["default_agent"]; ok {
		t.Errorf("default_agent should not be emitted in baseline OpenCode config")
	}
	if _, ok := cfg["mcp"]; ok {
		t.Errorf("mcp should not be emitted in baseline OpenCode config")
	}
	if _, err := os.Stat(filepath.Join(targetDir, ".opencode", "agents", "guide.md")); os.IsNotExist(err) {
		t.Error("guide.md was not installed")
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

	cfgPath := filepath.Join(targetDir, OpenCodeConfigFilename)
	cfg, err := jsonc.ReadJSONCFile(cfgPath)
	if err != nil {
		t.Fatalf("read opencode.json: %v", err)
	}
	if got, _ := cfg["$schema"].(string); got != "https://opencode.ai/config.json" {
		t.Fatalf("$schema = %q, want OpenCode schema", got)
	}
	if got, _ := cfg["share"].(string); got != "disabled" {
		t.Fatalf("share = %q, want disabled", got)
	}
	instructions, ok := cfg["instructions"].([]any)
	if !ok || len(instructions) != 1 || instructions[0] != "AGENTS.md" {
		t.Fatalf("instructions = %v, want [AGENTS.md]", cfg["instructions"])
	}
	skills, ok := cfg["skills"].(map[string]any)
	if !ok {
		t.Fatalf("skills = %T, want object", cfg["skills"])
	}
	paths, ok := skills["paths"].([]any)
	if !ok || len(paths) != 1 || paths[0] != ".opencode/skills" {
		t.Fatalf("skills.paths = %v, want [.opencode/skills]", skills["paths"])
	}
	permission, ok := cfg["permission"].(map[string]any)
	if !ok {
		t.Fatalf("permission = %T, want object", cfg["permission"])
	}
	if _, ok := permission["bash"]; ok {
		t.Fatalf("top-level permission.bash should not be emitted")
	}
	if _, ok := permission["edit"]; ok {
		t.Fatalf("top-level permission.edit should not be emitted")
	}
	skillPerm, ok := permission["skill"].(map[string]any)
	if !ok {
		t.Fatalf("permission.skill = %T, want object", permission["skill"])
	}
	if got, _ := skillPerm["*"].(string); got != "allow" {
		t.Fatalf("permission.skill[*] = %q, want allow", got)
	}
	if _, ok := cfg["default_agent"]; ok {
		t.Fatalf("default_agent should not be emitted")
	}
	if _, ok := cfg["mcp"]; ok {
		t.Fatalf("mcp should not be emitted in default opencode.json")
	}
	agents, ok := cfg["agent"].(map[string]any)
	if !ok {
		t.Fatalf("agent = %T, want object", cfg["agent"])
	}
	assertAgentPermission := func(name, wantEdit, wantBash string) {
		t.Helper()
		rawAgent, ok := agents[name].(map[string]any)
		if !ok {
			t.Fatalf("agent.%s = %T, want object", name, agents[name])
		}
		perm, ok := rawAgent["permission"].(map[string]any)
		if !ok {
			t.Fatalf("agent.%s.permission = %T, want object", name, rawAgent["permission"])
		}
		if got, _ := perm["edit"].(string); got != wantEdit {
			t.Fatalf("agent.%s.permission.edit = %q, want %q", name, got, wantEdit)
		}
		if got, _ := perm["bash"].(string); got != wantBash {
			t.Fatalf("agent.%s.permission.bash = %q, want %q", name, got, wantBash)
		}
		agentSkillPerm, ok := perm["skill"].(map[string]any)
		if !ok {
			t.Fatalf("agent.%s.permission.skill = %T, want object", name, perm["skill"])
		}
		if got, _ := agentSkillPerm["*"].(string); got != "allow" {
			t.Fatalf("agent.%s.permission.skill[*] = %q, want allow", name, got)
		}
	}
	assertAgentPermission("plan", "deny", "ask")
	assertAgentPermission("build", "ask", "ask")
	assertAgentPermission("explore", "deny", "deny")
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
