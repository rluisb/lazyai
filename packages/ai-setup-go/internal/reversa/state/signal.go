package state

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type PopulateSignal struct {
	PlaceholderCount int     `json:"placeholder_count"`
	GeneratedAt      string  `json:"generated_at"`
	LastPopulated    *string `json:"last_populated"`
	Skipped          bool    `json:"skipped"`
}

// ReadPopulateSignal reads the .ai/populate-needed signal file.
// Returns nil if the file doesn't exist.
func ReadPopulateSignal(targetDir string) (*PopulateSignal, error) {
	path := filepath.Join(targetDir, ".ai", "populate-needed")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var s PopulateSignal
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// WritePopulateSignal writes the .ai/populate-needed signal file.
func WritePopulateSignal(targetDir string, s PopulateSignal) error {
	aiDir := filepath.Join(targetDir, ".ai")
	if err := os.MkdirAll(aiDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(aiDir, "populate-needed"), data, 0644)
}

// DeletePopulateSignal removes the signal file (called after successful population).
func DeletePopulateSignal(targetDir string) error {
	path := filepath.Join(targetDir, ".ai", "populate-needed")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
