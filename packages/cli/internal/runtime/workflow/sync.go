package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Sync reads workflow YAML files from a directory and updates the database
func (m *Manager) Sync(workflowsDir string) error {
	entries, err := os.ReadDir(workflowsDir)
	if err != nil {
		return fmt.Errorf("read workflows directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		path := filepath.Join(workflowsDir, entry.Name())
		wf, err := ParseYAML(path)
		if err != nil {
			return fmt.Errorf("parse workflow %s: %w", entry.Name(), err)
		}

		if err := m.upsertWorkflow(wf); err != nil {
			return fmt.Errorf("upsert workflow %s: %w", wf.Name, err)
		}
	}

	return nil
}

// upsertWorkflow inserts or updates a workflow definition
func (m *Manager) upsertWorkflow(wf *Workflow) error {
	// TODO: Implement DB upsert
	// Check if workflow exists
	// If yes: update version, updated_at
	// If no: insert with version=1
	wf.Version = 1
	now := time.Now()
	wf.UpdatedAt = &now
	return nil
}

// List returns available workflows
func (m *Manager) List() ([]Workflow, error) {
	// TODO: Implement DB query
	return nil, nil
}

// Show returns workflow details
func (m *Manager) Show(workflowName string) (*Workflow, error) {
	// TODO: Implement DB query
	return m.loadWorkflow(workflowName)
}