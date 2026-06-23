package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/db"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func newSetupTestCommand(t *testing.T) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("scan", false, "")
	cmd.Flags().Bool("list", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("adopt", false, "")
	cmd.Flags().Bool("import", false, "")
	cmd.Flags().StringSlice("tool", nil, "")
	cmd.Flags().Bool("all", false, "")
	cmd.Flags().Bool("global", false, "")
	return cmd
}

func TestRunSetupScanOutputsInventoryJSON(t *testing.T) {
	dir := t.TempDir()
	withWorkingDir(t, dir)

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	if err := os.MkdirAll(filepath.Join(homeDir, ".config", "opencode"), 0o755); err != nil {
		t.Fatalf("mkdir home config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(homeDir, ".config", "opencode", "opencode.json"), []byte(`{"version":"2.0.0"}`), 0o644); err != nil {
		t.Fatalf("seed opencode config: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0o755); err != nil {
		t.Fatalf("mkdir claude dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".claude", "settings.json"), []byte(`{"ok":true}`), 0o644); err != nil {
		t.Fatalf("seed claude settings: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.Flags().Bool("scan", false, "")
	cmd.Flags().Bool("adopt", false, "")
	cmd.Flags().Bool("import", false, "")
	_ = cmd.Flags().Set("scan", "true")

	stdout, stderr := captureOutput(t, func() {
		if err := runSetup(cmd, nil); err != nil {
			t.Fatalf("runSetup: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	var inventory struct {
		CurrentState struct {
			Targets []struct {
				ID         string `json:"id"`
				Detections []struct {
					Scope   string `json:"scope"`
					Status  string `json:"status"`
					Version string `json:"version"`
				} `json:"detections"`
			} `json:"targets"`
		} `json:"currentState"`
	}
	if err := json.Unmarshal([]byte(stdout), &inventory); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout:\n%s", err, stdout)
	}

	foundOpenCode := false
	foundClaudeProject := false
	for _, target := range inventory.CurrentState.Targets {
		if target.ID == "opencode" {
			for _, detection := range target.Detections {
				if detection.Scope == "global" && detection.Status == "detected" && detection.Version == "2.0.0" {
					foundOpenCode = true
				}
			}
		}
		if target.ID == "claude-code" {
			for _, detection := range target.Detections {
				if detection.Scope == "project" && detection.Status == "detected" {
					foundClaudeProject = true
				}
			}
		}
	}
	if !foundOpenCode {
		t.Fatal("expected detected global opencode target in JSON output")
	}
	if !foundClaudeProject {
		t.Fatal("expected detected project claude-code target in JSON output")
	}
}

func TestRunSetupRejectsAdoptWithoutScan(t *testing.T) {
	cmd := newSetupTestCommand(t)
	_ = cmd.Flags().Set("adopt", "true")

	err := runSetup(cmd, nil)
	if err == nil {
		t.Fatal("expected error when adopt is used without scan")
	}
	if got, want := err.Error(), "--adopt and --import require --scan"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestRunSetupListOutputsDeterministicJSON(t *testing.T) {
	dir := t.TempDir()
	withWorkingDir(t, dir)

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	mustWriteSetupTestFile(t, filepath.Join(dir, ".ai", "agents", "test-agent", "AGENT.md"), "# Test Agent\n\nPrompt body.\n")

	cmd := newSetupTestCommand(t)
	_ = cmd.Flags().Set("list", "true")

	stdout, stderr := captureOutput(t, func() {
		if err := runSetup(cmd, nil); err != nil {
			t.Fatalf("runSetup: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	var result struct {
		Mode    string `json:"mode"`
		Targets []struct {
			ID string `json:"id"`
		} `json:"targets"`
		Agents []struct {
			ID string `json:"id"`
		} `json:"agents"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout:\n%s", err, stdout)
	}
	if result.Mode != "list" {
		t.Fatalf("mode = %q, want list", result.Mode)
	}
	gotTargets := make([]string, 0, len(result.Targets))
	for _, target := range result.Targets {
		gotTargets = append(gotTargets, target.ID)
	}
	if got, want := strings.Join(gotTargets, ","), "antigravity,claude-code,copilot,kiro,omp,opencode,pi"; got != want {
		t.Fatalf("targets = %q, want %q", got, want)
	}
	if len(result.Agents) != 1 || result.Agents[0].ID != "test-agent" {
		t.Fatalf("agents = %+v, want [test-agent]", result.Agents)
	}
}

func TestRunSetupListGlobalFiltersUnsupportedTargets(t *testing.T) {
	dir := t.TempDir()
	withWorkingDir(t, dir)

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	_ = os.MkdirAll(filepath.Join(homeDir, ".omp", "agent"), 0o755)
	_ = os.MkdirAll(filepath.Join(homeDir, ".kiro"), 0o755)

	cmd := newSetupTestCommand(t)
	_ = cmd.Flags().Set("list", "true")
	_ = cmd.Flags().Set("global", "true")

	stdout, stderr := captureOutput(t, func() {
		if err := runSetup(cmd, nil); err != nil {
			t.Fatalf("runSetup: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	var result struct {
		ScopeFilter string `json:"scopeFilter"`
		Targets     []struct {
			ID              string   `json:"id"`
			SupportedScopes []string `json:"supportedScopes"`
			CandidateRoots  []struct {
				Scope string `json:"scope"`
			} `json:"candidateRoots"`
		} `json:"targets"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout:\n%s", err, stdout)
	}
	if result.ScopeFilter != "global" {
		t.Fatalf("scopeFilter = %q, want global", result.ScopeFilter)
	}
	for _, target := range result.Targets {
		if got := strings.Join(target.SupportedScopes, ","); got != "global" {
			t.Fatalf("target %s supported scopes = %q, want global", target.ID, got)
		}
		if len(target.CandidateRoots) != 1 || target.CandidateRoots[0].Scope != "global" {
			t.Fatalf("target %s candidate roots = %+v, want one global root", target.ID, target.CandidateRoots)
		}
	}
}

func TestRunSetupDryRunPlansSelectedToolWithoutWriting(t *testing.T) {
	dir := t.TempDir()
	withWorkingDir(t, dir)

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	mustWriteSetupTestFile(t, filepath.Join(dir, ".claude", "settings.json"), `{"ok":true}`)

	cmd := newSetupTestCommand(t)
	_ = cmd.Flags().Set("dry-run", "true")
	_ = cmd.Flags().Set("tool", "claude-code")

	stdout, stderr := captureOutput(t, func() {
		if err := runSetup(cmd, nil); err != nil {
			t.Fatalf("runSetup: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	var result struct {
		Mode    string `json:"mode"`
		Scope   string `json:"scope"`
		Targets []struct {
			ID             string   `json:"id"`
			Action         string   `json:"action"`
			ExistingStatus string   `json:"existingStatus"`
			ObservedFiles  []string `json:"observedFiles"`
		} `json:"targets"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout:\n%s", err, stdout)
	}
	if result.Mode != "dry-run" || result.Scope != "project" {
		t.Fatalf("result = %+v, want mode=dry-run scope=project", result)
	}
	if len(result.Targets) != 1 || result.Targets[0].ID != "claude-code" {
		t.Fatalf("targets = %+v, want one claude-code target", result.Targets)
	}
	if result.Targets[0].Action != "preserve-existing" {
		t.Fatalf("action = %q, want preserve-existing", result.Targets[0].Action)
	}
	if result.Targets[0].ExistingStatus != "detected" {
		t.Fatalf("existing status = %q, want detected", result.Targets[0].ExistingStatus)
	}
	if got := strings.Join(result.Targets[0].ObservedFiles, ","); got != "settings.json" {
		t.Fatalf("observed files = %q, want settings.json", got)
	}
	if _, err := os.Stat(filepath.Join(homeDir, ".ai-setup")); !os.IsNotExist(err) {
		t.Fatalf("expected no writes to %s, stat err = %v", filepath.Join(homeDir, ".ai-setup"), err)
	}
}

func TestRunSetupDryRunGlobalAllFiltersToSupportedTargets(t *testing.T) {
	dir := t.TempDir()
	withWorkingDir(t, dir)

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	_ = os.MkdirAll(filepath.Join(homeDir, ".omp", "agent"), 0o755)
	_ = os.MkdirAll(filepath.Join(homeDir, ".kiro"), 0o755)

	cmd := newSetupTestCommand(t)
	_ = cmd.Flags().Set("dry-run", "true")
	_ = cmd.Flags().Set("all", "true")
	_ = cmd.Flags().Set("global", "true")

	stdout, stderr := captureOutput(t, func() {
		if err := runSetup(cmd, nil); err != nil {
			t.Fatalf("runSetup: %v", err)
		}
	})
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	var result struct {
		Scope   string `json:"scope"`
		Targets []struct {
			ID string `json:"id"`
		} `json:"targets"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal stdout: %v\nstdout:\n%s", err, stdout)
	}
	if result.Scope != "global" {
		t.Fatalf("scope = %q, want global", result.Scope)
	}
	gotTargets := make([]string, 0, len(result.Targets))
	for _, target := range result.Targets {
		gotTargets = append(gotTargets, target.ID)
	}
	sort.Strings(gotTargets)
	if got, want := strings.Join(gotTargets, ","), "antigravity,claude-code,copilot,kiro,omp,opencode,pi"; got != want {
		t.Fatalf("targets = %q, want %q", got, want)
	}
}

func TestRunSetupRejectsUnknownTool(t *testing.T) {
	cmd := newSetupTestCommand(t)
	_ = cmd.Flags().Set("dry-run", "true")
	_ = cmd.Flags().Set("tool", "nope")

	err := runSetup(cmd, nil)
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
	if got, want := err.Error(), `unknown tool "nope"`; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestRunSetupRejectsMultiplePrimaryActions(t *testing.T) {
	cmd := newSetupTestCommand(t)
	_ = cmd.Flags().Set("scan", "true")
	_ = cmd.Flags().Set("list", "true")

	err := runSetup(cmd, nil)
	if err == nil {
		t.Fatal("expected error for multiple setup actions")
	}
	if got, want := err.Error(), "select exactly one of --scan, --list, or --dry-run"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func mustWriteSetupTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestDiscoverWorkspaceRoot(t *testing.T) {
	workspaceRoot := t.TempDir()
	nested := filepath.Join(workspaceRoot, "app", "src")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	// State DB lives at the planning-repo dir (workspaceRoot/app) and records a
	// workspace-scope install pointing back at the workspace root.
	stateDir := filepath.Join(workspaceRoot, "app")
	database, err := db.Open(db.DefaultDBPath(stateDir))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.RunMigrations(database); err != nil {
		database.Close()
		t.Fatalf("migrate db: %v", err)
	}
	store := db.NewStore(database)
	data := types.DefaultStoreData()
	data.Config.SetupScope = types.SetupScopeWorkspace
	data.Config.WorkspaceRoot = workspaceRoot
	if err := store.WriteStoreData(&data); err != nil {
		database.Close()
		t.Fatalf("write store: %v", err)
	}
	database.Close()

	if got := discoverWorkspaceRoot(nested); got != workspaceRoot {
		t.Fatalf("discoverWorkspaceRoot(nested) = %q, want %q", got, workspaceRoot)
	}

	// A directory with no LazyAI state yields no workspace root.
	if got := discoverWorkspaceRoot(t.TempDir()); got != "" {
		t.Fatalf("discoverWorkspaceRoot(empty) = %q, want \"\"", got)
	}

	// A project-scope install yields no workspace root.
	projectDir := t.TempDir()
	pdb, err := db.Open(db.DefaultDBPath(projectDir))
	if err != nil {
		t.Fatalf("open project db: %v", err)
	}
	if err := db.RunMigrations(pdb); err != nil {
		pdb.Close()
		t.Fatalf("migrate project db: %v", err)
	}
	pdata := types.DefaultStoreData()
	pdata.Config.SetupScope = types.SetupScopeProject
	if err := db.NewStore(pdb).WriteStoreData(&pdata); err != nil {
		pdb.Close()
		t.Fatalf("write project store: %v", err)
	}
	pdb.Close()
	if got := discoverWorkspaceRoot(projectDir); got != "" {
		t.Fatalf("discoverWorkspaceRoot(project) = %q, want \"\"", got)
	}
}
