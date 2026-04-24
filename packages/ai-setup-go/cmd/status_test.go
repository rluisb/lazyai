package cmd

import (
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

func TestStatusReadStorePrefersSQLite(t *testing.T) {
	dir := t.TempDir()
	seedStoreData(t, dir, func(data *types.StoreData) {
		data.Config.ProjectName = "sqlite-project"
	})
	writeManifestStoreData(t, dir, func(data *types.StoreData) {
		data.Config.ProjectName = "manifest-project"
	})

	storeData, err := readStore(dir)
	if err != nil {
		t.Fatalf("readStore: %v", err)
	}
	if storeData.Config.ProjectName != "sqlite-project" {
		t.Fatalf("ProjectName = %q, want sqlite-project", storeData.Config.ProjectName)
	}
}

func TestStatusReadStoreFallsBackToManifest(t *testing.T) {
	dir := t.TempDir()
	writeManifestStoreData(t, dir, func(data *types.StoreData) {
		data.Config.ProjectName = "manifest-project"
		data.Config.Tools = []types.ToolId{types.ToolIdOpenCode}
	})

	storeData, err := readStore(dir)
	if err != nil {
		t.Fatalf("readStore: %v", err)
	}
	if storeData.Config.ProjectName != "manifest-project" {
		t.Fatalf("ProjectName = %q, want manifest-project", storeData.Config.ProjectName)
	}
	if len(storeData.Config.Tools) != 1 || storeData.Config.Tools[0] != types.ToolIdOpenCode {
		t.Fatalf("Tools = %v, want [%s]", storeData.Config.Tools, types.ToolIdOpenCode)
	}
}
