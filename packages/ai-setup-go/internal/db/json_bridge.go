package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
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

// AutoImportJSON imports from .ai-setup.json if it exists and the SQLite DB
// doesn't yet exist. If both exist, it prefers SQLite. If neither exists, it
// does nothing.
//
// Returns true if an import was performed, false otherwise.
func AutoImportJSON(targetDir string, db *DB) (bool, error) {
	jsonExists := HasExistingJSON(targetDir)
	dbExists := HasExistingDB(targetDir)

	if !jsonExists {
		// No JSON to import from.
		return false, nil
	}

	if dbExists {
		// Both exist; prefer SQLite. The JSON file is left in place
		// for backward compatibility / manual inspection.
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
