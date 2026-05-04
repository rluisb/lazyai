// Package manifest reads and writes .ai-setup.json store files.
// Ported from the TypeScript utilities in src/utils/manifest.ts.
package manifest

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	aierror "github.com/rluisb/lazyai/packages/cli/internal/error"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// ManifestFile is the name of the manifest file.
const ManifestFile = ".ai-setup.json"

// ReadManifest reads the .ai-setup.json file in targetDir and unmarshals it
// into a types.StoreData.
func ReadManifest(targetDir string) (*types.StoreData, error) {
	manifestPath := filepath.Join(targetDir, ManifestFile)
	if !files.FileExists(manifestPath) {
		return nil, aierror.ManifestNotFound(targetDir)
	}

	data, err := files.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	var storeData types.StoreData
	if err := json.Unmarshal(data, &storeData); err != nil {
		return nil, aierror.ManifestCorrupt(targetDir, err)
	}

	// Validate schema version.
	if storeData.Meta.SchemaVersion > types.CurrentSchemaVersion {
		return nil, aierror.ManifestVersion(fmt.Sprintf("%d", storeData.Meta.SchemaVersion))
	}

	return &storeData, nil
}

// ManifestExists reports whether a .ai-setup.json file exists in targetDir.
func ManifestExists(targetDir string) bool {
	manifestPath := filepath.Join(targetDir, ManifestFile)
	return files.FileExists(manifestPath)
}

// WriteManifest writes a types.StoreData as JSON to .ai-setup.json in targetDir.
// This supports backward compatibility with projects that have the old manifest format.
func WriteManifest(targetDir string, data *types.StoreData) error {
	manifestPath := filepath.Join(targetDir, ManifestFile)

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return aierror.Unknown("failed to marshal manifest", err)
	}

	return files.WriteFile(manifestPath, jsonData, 0o644)
}

// ReadManifestOptional reads the manifest, returning nil (no error) if it doesn't exist.
func ReadManifestOptional(targetDir string) (*types.StoreData, error) {
	if !ManifestExists(targetDir) {
		return nil, nil
	}
	return ReadManifest(targetDir)
}
