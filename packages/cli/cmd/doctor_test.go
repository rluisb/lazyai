package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/manifest"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestDoctorReportsStrayAgentsAndMetadataGapsInJSON(t *testing.T) {
	dir := t.TempDir()
	writeDoctorManifest(t, dir)

	if err := os.MkdirAll(filepath.Join(dir, "specs", "features"), 0o755); err != nil {
		t.Fatalf("mkdir features: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "specs", "006-test"), 0o755); err != nil {
		t.Fatalf("mkdir spec: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "specs", "features", "AGENTS.md"), []byte("# stray\n"), 0o644); err != nil {
		t.Fatalf("write stray AGENTS: %v", err)
	}
	newSpec := `---
created_at: 2026-04-17T10:00:00Z
updated_at: 2026-04-17T10:00:00Z
created_by: planner
updated_by: planner
---
# missing schema metadata
`
	if err := os.WriteFile(filepath.Join(dir, "specs", "006-test", "plan.md"), []byte(newSpec), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}

	cmd := newDoctorCommand(dir, true)
	stdout, _ := captureOutput(t, func() {
		_ = runDoctor(cmd, nil)
	})

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal output: %v\noutput=%s", err, stdout)
	}

	stray, ok := payload["strayAgentsFiles"].([]any)
	if !ok || len(stray) != 1 {
		t.Fatalf("expected one strayAgentsFiles entry, got %#v", payload["strayAgentsFiles"])
	}
	gaps, ok := payload["metadataGaps"].([]any)
	if !ok || len(gaps) == 0 {
		t.Fatalf("expected metadata gaps, got %#v", payload["metadataGaps"])
	}
}

func TestDoctorLegacyMetadataGapIsWarningOnly(t *testing.T) {
	dir := t.TempDir()
	writeDoctorManifest(t, dir)

	if err := os.MkdirAll(filepath.Join(dir, "specs", "legacy"), 0o755); err != nil {
		t.Fatalf("mkdir legacy: %v", err)
	}
	legacy := `---
title: Legacy Spec
---
# legacy
`
	if err := os.WriteFile(filepath.Join(dir, "specs", "legacy", "plan.md"), []byte(legacy), 0o644); err != nil {
		t.Fatalf("write legacy plan: %v", err)
	}

	cmd := newDoctorCommand(dir, true)
	stdout, _ := captureOutput(t, func() {
		if err := runDoctor(cmd, nil); err != nil {
			t.Fatalf("runDoctor returned unexpected error: %v", err)
		}
	})

	var payload struct {
		Healthy      bool `json:"healthy"`
		MetadataGaps []struct {
			Severity string `json:"severity"`
		} `json:"metadataGaps"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if !payload.Healthy {
		t.Fatal("expected legacy warning to keep doctor healthy")
	}
	if len(payload.MetadataGaps) != 1 || payload.MetadataGaps[0].Severity != "warning" {
		t.Fatalf("unexpected metadata gaps payload: %#v", payload.MetadataGaps)
	}
}

func newDoctorCommand(dir string, jsonOutput bool) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("dir", dir, "")
	cmd.Flags().Bool("fix", false, "")
	cmd.Flags().Bool("verbose", false, "")
	cmd.Flags().Bool("json", jsonOutput, "")
	return cmd
}

func writeDoctorManifest(t *testing.T, dir string) {
	t.Helper()
	storeData := types.DefaultStoreData()
	storeData.Config.SetupScope = types.SetupScopeProject
	storeData.Config.ProjectName = filepath.Base(dir)
	storeData.Config.TargetDir = dir
	storeData.Config.PlanningDir = "specs"
	storeData.Meta.CLIVersion = Version
	if err := manifest.WriteManifest(dir, &storeData); err != nil {
		t.Fatalf("WriteManifest: %v", err)
	}
}
