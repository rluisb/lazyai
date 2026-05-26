package workflow

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/runtime"
	"github.com/rluisb/lazyai/packages/cli/internal/runtime/dispatch"
)

// Manager handles workflow operations
type Manager struct {
	db       *runtime.DB
	session  SessionManager
	dispatch dispatch.Dispatcher
}

// SessionManager handles session operations
type SessionManager interface {
	RecordDispatch(sessionID string, agent string, task string, phase string) error
}

// NewManager creates a workflow manager
func NewManager(db *runtime.DB, sessionMgr SessionManager, disp dispatch.Dispatcher) *Manager {
	return &Manager{
		db:       db,
		session:  sessionMgr,
		dispatch: disp,
	}
}

// Run executes a workflow by name
func (m *Manager) Run(sessionID string, workflowName string, opts RunOptions) (*RunResult, error) {
	// Load workflow definition
	wf, err := m.loadWorkflow(workflowName)
	if err != nil {
		return nil, fmt.Errorf("load workflow: %w", err)
	}

	// Resolve mode
	mode := opts.Mode
	if mode == "" {
		mode = wf.Config.DefaultMode
	}
	modeConfig, ok := wf.Config.Modes[mode]
	if !ok {
		return nil, fmt.Errorf("mode %q not found in workflow %q", mode, workflowName)
	}

	// Create workflow run record
	result := &RunResult{
		WorkflowName: workflowName,
		SessionID:    sessionID,
		Status:       "running",
		CurrentStep:  0,
		TotalSteps:   len(modeConfig.Phases),
		StartedAt:    time.Now(),
	}

	if err := m.createWorkflowRun(result); err != nil {
		return nil, fmt.Errorf("create workflow run: %w", err)
	}

	// Execute phases
	for i, phaseName := range modeConfig.Phases {
		result.CurrentStep = i + 1

		// Find step definition
		step := m.findStep(wf, phaseName)
		if step == nil {
			result.Status = "failed"
			result.ErrorMessage = fmt.Sprintf("phase %q not found in workflow", phaseName)
			_ = m.updateWorkflowRun(result)
			return result, fmt.Errorf("phase %q not found", phaseName)
		}

		// Interpolate variables
		interpolatedTask, err := InterpolateWithTernary(step.Feedforward, opts.Context)
		if err != nil {
			result.Status = "failed"
			result.ErrorMessage = fmt.Sprintf("interpolate phase %q: %v", phaseName, err)
			_ = m.updateWorkflowRun(result)
			return result, err
		}

		// Create step record
		stepID, err := m.createWorkflowStep(result.InstanceID, step, interpolatedTask)
		if err != nil {
			result.Status = "failed"
			result.ErrorMessage = fmt.Sprintf("create step record: %v", err)
			_ = m.updateWorkflowRun(result)
			return result, err
		}

		// Record dispatch to session
		if m.session != nil {
			_ = m.session.RecordDispatch(sessionID, step.Agent, interpolatedTask, phaseName)
		}

		// Dispatch agent (or simulate in dry-run)
		if opts.DryRun {
			fmt.Printf("[DRY RUN] Would dispatch: agent=%s task=%s mode=%s\n", step.Agent, interpolatedTask, step.Mode)
			_ = m.updateWorkflowStep(stepID, "completed", "dry-run")
		} else {
			dispatchResult, err := m.dispatch.Dispatch(sessionID, step.Agent, interpolatedTask, step.Mode)
			if err != nil {
				// Handle failure
				result.Status = "failed"
				result.ErrorMessage = fmt.Sprintf("dispatch phase %q: %v", phaseName, err)
				_ = m.updateWorkflowStep(stepID, "failed", err.Error())
				_ = m.updateWorkflowRun(result)
				return result, err
			}

			// Update step with result
			_ = m.updateWorkflowStep(stepID, "completed", dispatchResult.Output)
		}

		// Evaluate gate
		if step.Gate != nil && !opts.SkipGates && !modeConfig.SkipGates {
			if err := m.evaluateGate(step.Gate); err != nil {
				result.Status = "failed"
				result.ErrorMessage = fmt.Sprintf("gate rejected phase %q: %v", phaseName, err)
				_ = m.updateWorkflowRun(result)
				return result, err
			}
		}
	}

	// Mark complete
	result.Status = "completed"
	now := time.Now()
	result.CompletedAt = &now
	_ = m.updateWorkflowRun(result)

	return result, nil
}

// evaluateGate handles human and auto gates
func (m *Manager) evaluateGate(gate *GateConfig) error {
	switch gate.Type {
	case "human":
		fmt.Printf("\n⛔ Gate: %s\n", gate.Description)
		if gate.Prompt != "" {
			fmt.Printf("%s\n", gate.Prompt)
		}
		fmt.Print("Continue? (y/n): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			return fmt.Errorf("human gate rejected")
		}
	case "auto":
		// Auto gates always pass for now
		fmt.Printf("✅ Auto-gate: %s\n", gate.Description)
	default:
		return fmt.Errorf("unknown gate type: %s", gate.Type)
	}
	return nil
}

// findStep locates a step by phase name
func (m *Manager) findStep(wf *Workflow, phaseName string) *WorkflowStep {
	for i := range wf.Steps {
		if wf.Steps[i].Name == phaseName {
			return &wf.Steps[i]
		}
	}
	return nil
}

// loadWorkflow retrieves a workflow from the database
func (m *Manager) loadWorkflow(name string) (*Workflow, error) {
	var id, wfName, description, trigger, configJSON, team string
	var version int
	var createdAt string
	var updatedAt sql.NullString

	err := m.db.QueryRow(
		"SELECT id, name, description, trigger_cmd, config_json, version, created_at, updated_at, team FROM workflows WHERE name = ?",
		name,
	).Scan(&id, &wfName, &description, &trigger, &configJSON, &version, &createdAt, &updatedAt, &team)
	if err == sql.ErrNoRows {
		// Fall back to YAML file
		return ParseYAML(".opencode/workflows/" + name + ".yaml")
	}
	if err != nil {
		return nil, fmt.Errorf("query workflow: %w", err)
	}

	// TODO: Parse config_json and steps
	return nil, fmt.Errorf("DB workflow loading not yet implemented")
}

// createWorkflowRun inserts a workflow run record
func (m *Manager) createWorkflowRun(result *RunResult) error {
	res, err := m.db.Exec(
		"INSERT INTO workflow_instances (workflow_name, session_id, status, current_step, total_steps, started_at) VALUES (?, ?, ?, ?, ?, ?)",
		result.WorkflowName, result.SessionID, result.Status, result.CurrentStep, result.TotalSteps, result.StartedAt.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("insert workflow instance: %w", err)
	}
	id, _ := res.LastInsertId()
	result.InstanceID = int(id)
	return nil
}

// updateWorkflowRun updates a workflow run record
func (m *Manager) updateWorkflowRun(result *RunResult) error {
	var completedAt interface{}
	if result.CompletedAt != nil {
		completedAt = result.CompletedAt.Format(time.RFC3339)
	}
	_, err := m.db.Exec(
		"UPDATE workflow_instances SET status = ?, current_step = ?, result = ?, error_message = ?, completed_at = ? WHERE id = ?",
		result.Status, result.CurrentStep, result.Result, result.ErrorMessage, completedAt, result.InstanceID,
	)
	return err
}

// createWorkflowStep inserts a workflow step record
func (m *Manager) createWorkflowStep(instanceID int, step *WorkflowStep, task string) (int, error) {
	res, err := m.db.Exec(
		"INSERT INTO workflow_steps (instance_id, step_order, agent, task, mode, status, started_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		instanceID, 0, step.Agent, task, step.Mode, "running", time.Now().Format(time.RFC3339),
	)
	if err != nil {
		return 0, fmt.Errorf("insert workflow step: %w", err)
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

// updateWorkflowStep updates a workflow step record
func (m *Manager) updateWorkflowStep(stepID int, status string, result string) error {
	_, err := m.db.Exec(
		"UPDATE workflow_steps SET status = ?, result = ?, completed_at = ? WHERE id = ?",
		status, result, time.Now().Format(time.RFC3339), stepID,
	)
	return err
}
