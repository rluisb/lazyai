package workflow

import (
	"fmt"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/runtime/dispatch"
)

// Manager handles workflow operations
type Manager struct {
	db       DB
	session  SessionManager
	dispatch dispatch.Dispatcher
}

// DB is the database interface for workflows
type DB interface {
	Exec(query string, args ...interface{}) error
	QueryRow(query string, args ...interface{}) Row
	Query(query string, args ...interface{}) (Rows, error)
}

// Row is a single database row
type Row interface {
	Scan(dest ...interface{}) error
}

// Rows is a result set
type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
}

// SessionManager handles session operations
type SessionManager interface {
	RecordDispatch(sessionID string, agent string, task string, phase string) error
}

// NewManager creates a workflow manager
func NewManager(db DB, sessionMgr SessionManager, disp dispatch.Dispatcher) *Manager {
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
	// TODO: Implement DB query
	// For now, parse from YAML file
	return ParseYAML(".opencode/workflows/" + name + ".yaml")
}

// createWorkflowRun inserts a workflow run record
func (m *Manager) createWorkflowRun(result *RunResult) error {
	// TODO: Implement DB insert
	result.InstanceID = int(time.Now().Unix())
	return nil
}

// updateWorkflowRun updates a workflow run record
func (m *Manager) updateWorkflowRun(result *RunResult) error {
	// TODO: Implement DB update
	return nil
}

// createWorkflowStep inserts a workflow step record
func (m *Manager) createWorkflowStep(instanceID int, step *WorkflowStep, task string) (int, error) {
	// TODO: Implement DB insert
	return int(time.Now().Unix()), nil
}

// updateWorkflowStep updates a workflow step record
func (m *Manager) updateWorkflowStep(stepID int, status string, result string) error {
	// TODO: Implement DB update
	return nil
}
