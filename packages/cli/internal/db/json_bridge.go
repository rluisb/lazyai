package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// ImportFromJSON reads a .ai-setup.json file and returns StoreData.
// Since the JSON format is identical to the Go struct serialization,
// this is mostly just json.Unmarshal.
func ImportFromJSON(jsonPath string) (*types.StoreData, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("read JSON file %s: %w", jsonPath, err)
	}

	var storeData types.StoreData
	if err := json.Unmarshal(data, &storeData); err != nil {
		return nil, fmt.Errorf("parse JSON %s: %w", jsonPath, err)
	}

	return &storeData, nil
}

// HasExistingJSON checks if a .ai-setup.json exists at the target directory.
func HasExistingJSON(targetDir string) bool {
	_, err := os.Stat(filepath.Join(targetDir, ".ai-setup.json"))
	return err == nil
}

// HasExistingDB checks if a .ai-setup.db exists at the target directory.
func HasExistingDB(targetDir string) bool {
	_, err := os.Stat(filepath.Join(targetDir, ".ai-setup.db"))
	return err == nil
}

// AutoImportJSON imports from .ai-setup.json when it exists and the SQLite DB
// did not exist before this run. Callers MUST capture DB existence BEFORE
// opening the database (dbPreexisted), because db.Open creates the DB file —
// re-checking existence here would always report the freshly-created file.
//
// Returns true if an import was performed, false otherwise.
func AutoImportJSON(targetDir string, db *DB, dbPreexisted bool) (bool, error) {
	if !HasExistingJSON(targetDir) {
		// No JSON to import from.
		return false, nil
	}

	if dbPreexisted {
		// DB already existed before this run; prefer SQLite. The JSON file is
		// left in place for backward compatibility / manual inspection.
		return false, nil
	}

	// JSON exists but DB doesn't: perform import.
	jsonPath := filepath.Join(targetDir, ".ai-setup.json")
	storeData, err := ImportFromJSON(jsonPath)
	if err != nil {
		return false, fmt.Errorf("auto-import JSON: %w", err)
	}

	store := NewStore(db)
	if err := store.WriteStoreData(storeData); err != nil {
		return false, fmt.Errorf("write imported data to SQLite: %w", err)
	}

	return true, nil
}
