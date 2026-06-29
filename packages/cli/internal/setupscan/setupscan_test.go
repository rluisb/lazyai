package setupscan

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestScanDetectsObservedTargetsAndSharedPaths(t *testing.T) {
	homeDir := t.TempDir()
	targetDir := t.TempDir()

	mustWriteFile(t, filepath.Join(homeDir, ".config", "opencode", "opencode.json"), `{"version":"1.2.3"}`)
	mustWriteFile(t, filepath.Join(targetDir, ".claude", "settings.json"), `{"theme":"dark"}`)
	mustMkdir(t, filepath.Join(homeDir, ".ai-setup"))
	mustMkdir(t, filepath.Join(targetDir, ".ai"))

	inventory, err := Scan(Options{HomeDir: homeDir, TargetDir: targetDir})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	if len(inventory.CurrentState.SharedPaths) != 2 {
		t.Fatalf("shared paths len = %d, want 2", len(inventory.CurrentState.SharedPaths))
	}
	if !inventory.CurrentState.SharedPaths[0].Exists {
		t.Fatalf("expected %s to exist", inventory.CurrentState.SharedPaths[0].ID)
	}
	if !inventory.CurrentState.SharedPaths[1].Exists {
		t.Fatalf("expected %s to exist", inventory.CurrentState.SharedPaths[1].ID)
	}

	opencode := findTarget(t, inventory.CurrentState.Targets, "opencode")
	assertDetectionStatus(t, opencode, "global", "detected")
	if version := detectionForScope(t, opencode, "global").Version; version != "1.2.3" {
		t.Fatalf("opencode global version = %q, want 1.2.3", version)
	}

	claude := findTarget(t, inventory.CurrentState.Targets, "claude-code")
	assertDetectionStatus(t, claude, "project", "detected")

	if _, err := json.Marshal(inventory); err != nil {
		t.Fatalf("inventory marshal: %v", err)
	}
}

func TestScanDetectsReusableWorkspaceAgentsAndScopedMCPMetadata(t *testing.T) {
	homeDir := t.TempDir()
	targetDir := t.TempDir()
	mustWriteFile(t, filepath.Join(targetDir, ".ai", "agents", "test-agent", "AGENT.md"), `---
title: Test Agent
description: Handles setup inventory work.
tools:
  - bash
  - read
---

# Test Agent

Focused prompt body.
`)
	mustWriteFile(t, filepath.Join(targetDir, ".ai", "agents", "test-agent", "mcp.json"), `{
  "mcpServers": {
    "filesystem": {"command": "filesystem"},
    "memory": {"command": "memory"}
  }
}`)

	inventory, err := Scan(Options{HomeDir: homeDir, TargetDir: targetDir})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	if len(inventory.CurrentState.Agents) != 1 {
		t.Fatalf("agent count = %d, want 1", len(inventory.CurrentState.Agents))
	}
	agent := inventory.CurrentState.Agents[0]
	if agent.ID != "test-agent" {
		t.Fatalf("agent id = %q, want test-agent", agent.ID)
	}
	if agent.Status != "detected" {
		t.Fatalf("agent status = %q, want detected", agent.Status)
	}
	if agent.Title != "Test Agent" {
		t.Fatalf("agent title = %q, want Test Agent", agent.Title)
	}
	if agent.Description != "Handles setup inventory work." {
		t.Fatalf("agent description = %q", agent.Description)
	}
	if strings.Join(agent.Tools, ",") != "bash,read" {
		t.Fatalf("agent tools = %v, want [bash read]", agent.Tools)
	}
	if agent.MCP == nil {
		t.Fatal("expected scoped MCP metadata")
	}
	if !agent.MCP.Scoped {
		t.Fatal("expected MCP metadata to be scoped")
	}
	if strings.Join(agent.MCP.ServerNames, ",") != "filesystem,memory" {
		t.Fatalf("agent MCP servers = %v, want [filesystem memory]", agent.MCP.ServerNames)
	}
	if agent.MCP.ServerCount != 2 {
		t.Fatalf("agent MCP server count = %d, want 2", agent.MCP.ServerCount)
	}
}

func TestScanRejectsReusableAgentWithInvalidDirectoryName(t *testing.T) {
	homeDir := t.TempDir()
	targetDir := t.TempDir()
	mustWriteFile(t, filepath.Join(targetDir, ".ai", "agents", "Bad Agent", "AGENT.md"), "# Bad Agent\n\nPrompt body.\n")

	inventory, err := Scan(Options{HomeDir: homeDir, TargetDir: targetDir})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	agent := findAgent(t, inventory.CurrentState.Agents, "Bad Agent")
	if agent.Status != "invalid" {
		t.Fatalf("agent status = %q, want invalid", agent.Status)
	}
	if !containsString(agent.Reasons, "invalid-agent-id") {
		t.Fatalf("agent reasons = %v, want invalid-agent-id", agent.Reasons)
	}
}

func TestScanRejectsReusableAgentWithInvalidMcpSchema(t *testing.T) {
	homeDir := t.TempDir()
	targetDir := t.TempDir()
	mustWriteFile(t, filepath.Join(targetDir, ".ai", "agents", "test-agent", "AGENT.md"), "# Test Agent\n\nPrompt body.\n")
	mustWriteFile(t, filepath.Join(targetDir, ".ai", "agents", "test-agent", "mcp.json"), `{"servers":{"memory":{"command":"memory"}}}`)

	inventory, err := Scan(Options{HomeDir: homeDir, TargetDir: targetDir})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	agent := findAgent(t, inventory.CurrentState.Agents, "test-agent")
	if agent.Status != "invalid" {
		t.Fatalf("agent status = %q, want invalid", agent.Status)
	}
	if !containsStringPrefix(agent.Reasons, "invalid-agent-mcp-schema") {
		t.Fatalf("agent reasons = %v, want invalid-agent-mcp-schema", agent.Reasons)
	}
}

func TestScanProducesDeterministicTargetOrdering(t *testing.T) {
	homeDir := t.TempDir()
	targetDir := t.TempDir()

	inventory, err := Scan(Options{HomeDir: homeDir, TargetDir: targetDir})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	got := make([]string, 0, len(inventory.CurrentState.Targets))
	for _, target := range inventory.CurrentState.Targets {
		got = append(got, target.ID)
	}
	want := []string{"antigravity", "claude-code", "codex", "copilot", "kiro", "omp", "opencode", "pi"}
	if len(got) != len(want) {
		t.Fatalf("target count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("target[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestScanDoesNotTreatBareGithubDirectoryAsCopilotDetection(t *testing.T) {
	homeDir := t.TempDir()
	targetDir := t.TempDir()
	mustMkdir(t, filepath.Join(targetDir, ".github"))

	inventory, err := Scan(Options{HomeDir: homeDir, TargetDir: targetDir})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	copilot := findTarget(t, inventory.CurrentState.Targets, "copilot")
	if got := detectionForScope(t, copilot, "project").Status; got != "missing" {
		t.Fatalf("copilot project status = %q, want missing", got)
	}
}

func TestScanClassifiesAdoptableAndMCPEntries(t *testing.T) {
	homeDir := t.TempDir()
	targetDir := t.TempDir()
	mustWriteFile(t, filepath.Join(targetDir, ".claude", "settings.json"), `{"mcpServers":{"memory":{"command":"memory"}}}`)

	inventory, err := Scan(Options{HomeDir: homeDir, TargetDir: targetDir})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	claude := findTarget(t, inventory.CurrentState.Targets, "claude-code")
	detection := detectionForScope(t, claude, "project")
	if detection.State != resourceStateAdoptable {
		t.Fatalf("state = %q, want %q", detection.State, resourceStateAdoptable)
	}
	if len(detection.MCPEntries) != 1 {
		t.Fatalf("mcp entry count = %d, want 1", len(detection.MCPEntries))
	}
	if detection.MCPEntries[0].Name != "memory" || detection.MCPEntries[0].State != resourceStateAdoptable {
		t.Fatalf("mcp entry = %+v, want adoptable memory entry", detection.MCPEntries[0])
	}
}

func TestRunAdoptMarksAdoptableConfigsManaged(t *testing.T) {
	homeDir := t.TempDir()
	targetDir := t.TempDir()
	settingsPath := filepath.Join(targetDir, ".claude", "settings.json")
	mustWriteFile(t, settingsPath, `{"mcpServers":{"memory":{"command":"memory"}}}`)

	inventory, err := Run(Options{HomeDir: homeDir, TargetDir: targetDir, Adopt: true})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	claude := findTarget(t, inventory.CurrentState.Targets, "claude-code")
	detection := detectionForScope(t, claude, "project")
	if detection.State != resourceStateManaged {
		t.Fatalf("state = %q, want %q", detection.State, resourceStateManaged)
	}
	if inventory.Operation == nil || len(inventory.Operation.Adopted) != 1 {
		t.Fatalf("operation = %+v, want one adopted record", inventory.Operation)
	}

	registryBytes, err := os.ReadFile(filepath.Join(homeDir, ".ai-setup", "config", "setup-scan-registry.json"))
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}
	if !strings.Contains(string(registryBytes), `"state": "managed"`) {
		t.Fatalf("registry missing managed state:\n%s", registryBytes)
	}
}

func TestScanMarksManagedConfigConflictWhenUserChangesIt(t *testing.T) {
	homeDir := t.TempDir()
	targetDir := t.TempDir()
	settingsPath := filepath.Join(targetDir, ".claude", "settings.json")
	mustWriteFile(t, settingsPath, `{"mcpServers":{"memory":{"command":"memory"}}}`)
	if _, err := Run(Options{HomeDir: homeDir, TargetDir: targetDir, Adopt: true}); err != nil {
		t.Fatalf("initial adopt: %v", err)
	}
	mustWriteFile(t, settingsPath, `{"mcpServers":{"memory":{"command":"memory-v2"}}}`)

	inventory, err := Scan(Options{HomeDir: homeDir, TargetDir: targetDir})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	claude := findTarget(t, inventory.CurrentState.Targets, "claude-code")
	detection := detectionForScope(t, claude, "project")
	if detection.State != resourceStateConflict {
		t.Fatalf("state = %q, want %q", detection.State, resourceStateConflict)
	}
	if len(detection.Reasons) == 0 {
		t.Fatal("expected conflict reasons")
	}
	if len(detection.MCPEntries) != 1 || detection.MCPEntries[0].State != resourceStateConflict {
		t.Fatalf("mcp entries = %+v, want one conflicting entry", detection.MCPEntries)
	}
}

func TestRunAdoptMarksExistingOrchestratorEntryManaged(t *testing.T) {
	homeDir := t.TempDir()
	targetDir := t.TempDir()
	configPath := filepath.Join(targetDir, ".opencode", "opencode.json")
	mustWriteFile(t, configPath, `{"mcp":{"orchestrator":{"type":"local","command":["/tmp/node","/tmp/orchestrator/dist/index.js"]}}}`)

	inventory, err := Run(Options{HomeDir: homeDir, TargetDir: targetDir, Adopt: true})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	opencode := findTarget(t, inventory.CurrentState.Targets, "opencode")
	detection := detectionForScope(t, opencode, "project")
	if detection.State != resourceStateManaged {
		t.Fatalf("state = %q, want %q", detection.State, resourceStateManaged)
	}
	if len(detection.MCPEntries) != 1 {
		t.Fatalf("mcp entries = %+v, want one orchestrator entry", detection.MCPEntries)
	}
	entry := detection.MCPEntries[0]
	if entry.Name != "orchestrator" || entry.State != resourceStateManaged {
		t.Fatalf("entry = %+v, want managed orchestrator", entry)
	}

	registryBytes, err := os.ReadFile(filepath.Join(homeDir, ".ai-setup", "config", "setup-scan-registry.json"))
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}
	if !strings.Contains(string(registryBytes), `"name": "orchestrator"`) {
		t.Fatalf("registry missing orchestrator entry:\n%s", registryBytes)
	}
}

func TestScanMarksManagedOrchestratorEntryConflictWhenChanged(t *testing.T) {
	homeDir := t.TempDir()
	targetDir := t.TempDir()
	configPath := filepath.Join(targetDir, ".opencode", "opencode.json")
	mustWriteFile(t, configPath, `{"mcp":{"orchestrator":{"type":"local","command":["/tmp/node","/tmp/orchestrator/dist/index.js"]}}}`)
	if _, err := Run(Options{HomeDir: homeDir, TargetDir: targetDir, Adopt: true}); err != nil {
		t.Fatalf("initial adopt: %v", err)
	}
	mustWriteFile(t, configPath, `{"mcp":{"orchestrator":{"type":"local","command":["/tmp/node","/tmp/orchestrator/dist/changed.js"]}}}`)

	inventory, err := Scan(Options{HomeDir: homeDir, TargetDir: targetDir})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	opencode := findTarget(t, inventory.CurrentState.Targets, "opencode")
	detection := detectionForScope(t, opencode, "project")
	if detection.State != resourceStateConflict {
		t.Fatalf("state = %q, want %q", detection.State, resourceStateConflict)
	}
	if len(detection.MCPEntries) != 1 || detection.MCPEntries[0].State != resourceStateConflict {
		t.Fatalf("mcp entries = %+v, want conflicting orchestrator entry", detection.MCPEntries)
	}
}

func TestScanClassifiesUserOwnedFromRegistry(t *testing.T) {
	homeDir := t.TempDir()
	targetDir := t.TempDir()
	rootPath := filepath.Join(targetDir, ".claude")
	mustWriteFile(t, filepath.Join(rootPath, "settings.json"), `{"mcpServers":{"memory":{"command":"memory"}}}`)
	mustWriteRegistry(t, filepath.Join(homeDir, ".ai-setup", "config", "setup-scan-registry.json"), scanRegistry{
		Version: scanRegistryVersion,
		Resources: []managedResource{{
			TargetID:  "claude-code",
			Scope:     "project",
			Origin:    "project",
			RootPath:  rootPath,
			State:     resourceStateUserOwned,
			UpdatedAt: "2026-04-26T00:00:00Z",
		}},
	})

	inventory, err := Scan(Options{HomeDir: homeDir, TargetDir: targetDir})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	claude := findTarget(t, inventory.CurrentState.Targets, "claude-code")
	detection := detectionForScope(t, claude, "project")
	if detection.State != resourceStateUserOwned {
		t.Fatalf("state = %q, want %q", detection.State, resourceStateUserOwned)
	}
}

func TestRunImportCopiesConfigsIntoAiSetupImportsAndBacksUpManagedDestination(t *testing.T) {
	homeDir := t.TempDir()
	targetDir := t.TempDir()
	settingsPath := filepath.Join(targetDir, ".claude", "settings.json")
	mustWriteFile(t, settingsPath, `{"mcpServers":{"memory":{"command":"memory"}}}`)

	first, err := Run(Options{HomeDir: homeDir, TargetDir: targetDir, Import: true})
	if err != nil {
		t.Fatalf("first import: %v", err)
	}
	if first.Operation == nil || len(first.Operation.Imported) == 0 {
		t.Fatalf("operation = %+v, want imported resources", first.Operation)
	}
	importedPath := first.Operation.Imported[0].DestinationPath
	if !filesExist(t, importedPath) {
		t.Fatalf("expected imported path %q", importedPath)
	}
	mustWriteFile(t, settingsPath, `{"mcpServers":{"memory":{"command":"memory-v2"}}}`)

	second, err := Run(Options{HomeDir: homeDir, TargetDir: targetDir, Import: true})
	if err != nil {
		t.Fatalf("second import: %v", err)
	}
	if len(second.Operation.Backups) == 0 {
		t.Fatalf("expected backup paths, got %+v", second.Operation)
	}
	backupFound := false
	for _, backupPath := range second.Operation.Backups {
		if strings.Contains(backupPath, filepath.Base(importedPath)) {
			backupFound = true
		}
	}
	if !backupFound {
		t.Fatalf("expected backup for imported file, got %+v", second.Operation.Backups)
	}
	content, err := os.ReadFile(importedPath)
	if err != nil {
		t.Fatalf("read imported path: %v", err)
	}
	if !strings.Contains(string(content), "memory-v2") {
		t.Fatalf("imported content = %s, want updated source content", content)
	}
}

func TestRunImportCopiesAgentDirectoryButSkipsNestedReservedContextDocs(t *testing.T) {
	homeDir := t.TempDir()
	targetDir := t.TempDir()
	mustWriteFile(t, filepath.Join(targetDir, ".opencode", "opencode.json"), `{"version":"1.0.0"}`)
	mustWriteFile(t, filepath.Join(targetDir, ".opencode", "agents", "implementer.md"), "# Implementer\n\nBuild things.")
	mustWriteFile(t, filepath.Join(targetDir, ".opencode", "agents", "AGENTS.md"), "# Nested Context\n\nDo not import.")

	inventory, err := Run(Options{HomeDir: homeDir, TargetDir: targetDir, Import: true})
	if err != nil {
		t.Fatalf("import: %v", err)
	}

	opencode := findTarget(t, inventory.CurrentState.Targets, "opencode")
	detection := detectionForScope(t, opencode, "project")
	importedAgentsDir := filepath.Join(inventory.Operation.ImportRoot, "opencode", importDirectoryName(detection.Scope, detection.Origin, detection.RootPath), "agents")
	if !filesExist(t, filepath.Join(importedAgentsDir, "implementer.md")) {
		t.Fatalf("expected implementer agent to be imported")
	}
	if filesExist(t, filepath.Join(importedAgentsDir, "AGENTS.md")) {
		t.Fatalf("nested reserved context doc should not be imported")
	}
}

func findTarget(t *testing.T, targets []ObservedTarget, id string) ObservedTarget {
	t.Helper()
	for _, target := range targets {
		if target.ID == id {
			return target
		}
	}
	t.Fatalf("target %q not found", id)
	return ObservedTarget{}
}

func findAgent(t *testing.T, agents []ObservedAgent, id string) ObservedAgent {
	t.Helper()
	for _, agent := range agents {
		if agent.ID == id {
			return agent
		}
	}
	t.Fatalf("agent %q not found", id)
	return ObservedAgent{}
}

func detectionForScope(t *testing.T, target ObservedTarget, scope string) TargetDetection {
	t.Helper()
	for _, detection := range target.Detections {
		if detection.Scope == scope {
			return detection
		}
	}
	t.Fatalf("scope %q not found for target %q", scope, target.ID)
	return TargetDetection{}
}

func assertDetectionStatus(t *testing.T, target ObservedTarget, scope, want string) {
	t.Helper()
	if got := detectionForScope(t, target, scope).Status; got != want {
		t.Fatalf("target %q scope %q status = %q, want %q", target.ID, scope, got, want)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %q: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %q: %v", path, err)
	}
}

func mustWriteRegistry(t *testing.T, path string, registry scanRegistry) {
	t.Helper()
	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		t.Fatalf("marshal registry: %v", err)
	}
	mustWriteFile(t, path, string(data))
}

func filesExist(t *testing.T, path string) bool {
	t.Helper()
	_, err := os.Stat(path)
	return err == nil
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsStringPrefix(values []string, wantPrefix string) bool {
	for _, value := range values {
		if strings.HasPrefix(value, wantPrefix) {
			return true
		}
	}
	return false
}

func TestSupportedScopesIncludesGlobalForAllTools(t *testing.T) {
	for _, tool := range []types.ToolId{types.ToolIdOpenCode, types.ToolIdClaudeCode, types.ToolIdCopilot, types.ToolIdOmp, types.ToolIdKiro, types.ToolIdPi, types.ToolIdAntigravity} {
		scopes := supportedScopesForTool(tool)
		hasGlobal := false
		for _, s := range scopes {
			if s == types.SetupScopeGlobal {
				hasGlobal = true
			}
		}
		if !hasGlobal {
			t.Errorf("tool %q expected global scope advertised, got %v", tool, scopes)
		}
		if len(scopes) != 3 {
			t.Errorf("tool %q expected 3 supported scopes (global, project, workspace), got %v", tool, scopes)
		}
	}
}

func TestScanResolvesWorkspaceRootForWorkspaceScope(t *testing.T) {
	homeDir := t.TempDir()
	workspaceRoot := t.TempDir()
	nestedDir := filepath.Join(workspaceRoot, "app")
	mustMkdir(t, nestedDir)

	// A workspace install writes tool configs to the workspace root, not the
	// nested planning-repo directory the user invokes commands from.
	mustWriteFile(t, filepath.Join(workspaceRoot, ".kiro", "skills", "diagnose", "SKILL.md"), "# diagnose\n")

	// Scanning from the nested dir without the workspace root misses the install.
	bare, err := Scan(Options{HomeDir: homeDir, TargetDir: nestedDir})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	assertDetectionStatus(t, findTarget(t, bare.CurrentState.Targets, "kiro"), "workspace", "missing")

	// Supplying the workspace root makes the workspace-scope detection resolve
	// against it, rediscovering the install.
	aware, err := Scan(Options{HomeDir: homeDir, TargetDir: nestedDir, WorkspaceRoot: workspaceRoot})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	kiro := findTarget(t, aware.CurrentState.Targets, "kiro")
	assertDetectionStatus(t, kiro, "workspace", "detected")
	if rootPath := detectionForScope(t, kiro, "workspace").RootPath; !strings.HasPrefix(rootPath, workspaceRoot) {
		t.Fatalf("workspace rootPath = %q, want under %q", rootPath, workspaceRoot)
	}
}
