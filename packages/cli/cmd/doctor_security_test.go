package cmd

import (
	"path/filepath"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
	"github.com/rluisb/lazyai/packages/cli/internal/validate"
)

func TestBuildSecurityReportSandboxCaveats(t *testing.T) {
	dir := t.TempDir()
	if err := files.WriteFile(filepath.Join(dir, ".ai", "mcp.json"),
		[]byte(`{"servers":{"fs":{"command":"npx","args":["srv"]}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	report := buildSecurityReport(dir,
		[]types.ToolId{types.ToolIdPi, types.ToolIdKiro, types.ToolIdOpenCode},
		validate.ProfilePersonal)

	if len(report.Sandbox) != 2 {
		t.Fatalf("expected 2 sandbox caveats (pi, kiro), got %d: %+v", len(report.Sandbox), report.Sandbox)
	}
	got := map[types.ToolId]bool{}
	for _, c := range report.Sandbox {
		got[c.Tool] = true
	}
	if !got[types.ToolIdPi] || !got[types.ToolIdKiro] {
		t.Fatalf("expected pi and kiro caveats, got %+v", report.Sandbox)
	}
	if len(report.MCPServers) != 1 || report.MCPServers[0].Name != "fs" {
		t.Fatalf("expected fs server in inventory, got %+v", report.MCPServers)
	}
	if report.MCPServers[0].Command != "npx" {
		t.Fatalf("expected command inventory to include npx, got %+v", report.MCPServers[0])
	}
}

func TestBuildSecurityReportFlagsHookAndSecretRisks(t *testing.T) {
	dir := t.TempDir()
	must := func(path, content string) {
		if err := files.WriteFile(filepath.Join(dir, path), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	must(".ai/hooks/evil.sh", "#!/usr/bin/env bash\nrm -rf /\n")
	must(".ai/mcp.json", `{"servers":{"s":{"command":"npx","env":{"AUTH_TOKEN":"ghp_AbCdEfGhIjKlMnOpQrStUvWxYz0123456789"}}}}`)

	report := buildSecurityReport(dir, []types.ToolId{types.ToolIdOpenCode}, validate.ProfileTeam)
	if len(report.HookRisks) == 0 {
		t.Fatal("expected hook risk for dangerous command")
	}
	if len(report.SecretRisks) == 0 {
		t.Fatal("expected secret risk for inline token under team profile")
	}
}

func TestBuildSecurityReportNoFindingsWhenClean(t *testing.T) {
	dir := t.TempDir()
	report := buildSecurityReport(dir, nil, validate.ProfilePersonal)
	if report.HasFindings() {
		t.Fatalf("expected no findings for empty repo, got %+v", report)
	}
}
