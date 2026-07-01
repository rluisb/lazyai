package sidecar

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadSidecarAt_MissingFile replaces TestLoadProjectSidecar_MissingFile
// and TestLoadGlobalSidecar_MissingFile: LoadSidecarAt is a single unified
// function regardless of scope root, so both scenarios are subtests of one
// table.
func TestLoadSidecarAt_MissingFile(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: "project-style root"},
		{name: "global-style root"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scopeRoot := t.TempDir()

			cfg, err := LoadSidecarAt(scopeRoot)
			require.NoError(t, err)
			assert.Nil(t, cfg)
		})
	}
}

// TestLoadSidecarAt_Valid replaces TestLoadProjectSidecar_Valid and
// TestLoadGlobalSidecar_Valid.
func TestLoadSidecarAt_Valid(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: "project-style root"},
		{name: "global-style root"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scopeRoot := t.TempDir()
			cfg := &SidecarConfig{
				Path:     "../kb",
				SpecsDir: "specs",
				DocsDir:  "docs",
				PlansDir: "plans",
			}
			require.NoError(t, WriteSidecarAt(scopeRoot, cfg))

			got, err := LoadSidecarAt(scopeRoot)
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, "../kb", got.Path)
			assert.Equal(t, "specs", got.SpecsDir)
			assert.Equal(t, "docs", got.DocsDir)
			assert.Equal(t, "plans", got.PlansDir)
		})
	}
}

// TestLoadSidecarAt_MalformedYAML replaces TestLoadProjectSidecar_MalformedYAML
// and TestLoadGlobalSidecar_MalformedYAML.
func TestLoadSidecarAt_MalformedYAML(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: "project-style root"},
		{name: "global-style root"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scopeRoot := t.TempDir()
			lazyaiDir := filepath.Join(scopeRoot, ".lazyai")
			require.NoError(t, os.MkdirAll(lazyaiDir, 0o755))
			path := filepath.Join(lazyaiDir, "sidecar.yaml")
			require.NoError(t, os.WriteFile(path, []byte("not: valid: yaml: ["), 0o644))

			_, err := LoadSidecarAt(scopeRoot)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "parsing sidecar at")
		})
	}
}
