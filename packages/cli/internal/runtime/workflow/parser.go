package workflow

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Workflow represents a workflow definition
type Workflow struct {
	ID          string
	Name        string
	Description string
	Trigger     string
	Config      WorkflowConfig
	Version     int
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	Team        string
	Steps       []WorkflowStep
}

// WorkflowConfig contains mode configurations
type WorkflowConfig struct {
	Modes       map[string]ModeConfig
	DefaultMode string
}

// ModeConfig defines a workflow mode
type ModeConfig struct {
	Phases    []string
	SkipGates bool
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	Name        string
	Agent       string
	Skill       string
	Mode        string
	Feedforward string
	Gate        *GateConfig
	Metrics     []string
}

// GateConfig defines a human or automatic gate
type GateConfig struct {
	Type        string // "human" or "auto"
	Description string
	Prompt      string
}

// RunResult contains the outcome of a workflow run
type RunResult struct {
	InstanceID   int
	WorkflowName string
	SessionID    string
	Status       string
	CurrentStep  int
	TotalSteps   int
	Result       string
	ErrorMessage string
	StartedAt    time.Time
	CompletedAt  *time.Time
}

// RunOptions configures workflow execution
type RunOptions struct {
	Mode      string
	SkipGates bool
	DryRun    bool
	Context   map[string]string // Variable interpolation context
}

// ParseYAML parses a workflow YAML file into a Workflow struct
func ParseYAML(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read workflow YAML: %w", err)
	}

	var raw rawWorkflow
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse workflow YAML: %w", err)
	}

	wf := &Workflow{
		Name:        raw.Name,
		Description: raw.Description,
		Trigger:     raw.Trigger,
		Version:     1,
		CreatedAt:   time.Now(),
	}

	// Parse modes
	if raw.Modes != nil {
		wf.Config.Modes = make(map[string]ModeConfig)
		for name, mode := range raw.Modes {
			wf.Config.Modes[name] = ModeConfig{
				Phases:    mode.Phases,
				SkipGates: mode.SkipGates,
			}
		}
	}
	wf.Config.DefaultMode = raw.DefaultMode

	// Parse phases/steps
	for _, phase := range raw.Phases {
		step := WorkflowStep{
			Name:        phase.Name,
			Agent:       phase.Agent,
			Skill:       phase.Skill,
			Mode:        phase.Mode,
			Feedforward: phase.Feedforward,
		}

		// Parse metrics map to slice
		for metricName, enabled := range phase.Metrics {
			if enabled {
				step.Metrics = append(step.Metrics, metricName)
			}
		}

		// Parse gate - can be string, null, or object
		if phase.Gate != nil {
			switch v := phase.Gate.(type) {
			case string:
				// Gate is a string like "human_confirms"
				step.Gate = &GateConfig{
					Type:        "human",
					Description: v,
					Prompt:      phase.GatePrompt,
				}
			case map[string]interface{}:
				// Gate is an object
				gateType, _ := v["type"].(string)
				gateDesc, _ := v["description"].(string)
				gatePrompt, _ := v["prompt"].(string)
				step.Gate = &GateConfig{
					Type:        gateType,
					Description: gateDesc,
					Prompt:      gatePrompt,
				}
			}
		}

		wf.Steps = append(wf.Steps, step)
	}

	// Validate
	if err := wf.Validate(); err != nil {
		return nil, fmt.Errorf("validate workflow: %w", err)
	}

	return wf, nil
}

// Validate checks that the workflow is well-formed
func (w *Workflow) Validate() error {
	if w.Name == "" {
		return fmt.Errorf("workflow name is required")
	}
	if len(w.Steps) == 0 {
		return fmt.Errorf("workflow must have at least one phase")
	}
	for i, step := range w.Steps {
		if step.Name == "" {
			return fmt.Errorf("phase %d: name is required", i)
		}
		if step.Agent == "" {
			return fmt.Errorf("phase %s: agent is required", step.Name)
		}
	}

	// If no modes defined, create a default mode that runs all phases
	if len(w.Config.Modes) == 0 {
		allPhases := make([]string, len(w.Steps))
		for i, step := range w.Steps {
			allPhases[i] = step.Name
		}
		w.Config.Modes = map[string]ModeConfig{
			"default": {
				Phases:    allPhases,
				SkipGates: false,
			},
		}
		w.Config.DefaultMode = "default"
	}

	if w.Config.DefaultMode == "" {
		return fmt.Errorf("default_mode is required")
	}
	if _, ok := w.Config.Modes[w.Config.DefaultMode]; !ok {
		return fmt.Errorf("default_mode %q not found in modes", w.Config.DefaultMode)
	}
	return nil
}

// rawWorkflow is the YAML unmarshaling struct
type rawWorkflow struct {
	Name        string             `yaml:"name"`
	Trigger     string             `yaml:"trigger"`
	Description string             `yaml:"description"`
	Modes       map[string]rawMode `yaml:"modes"`
	DefaultMode string             `yaml:"default_mode"`
	Phases      []rawPhase         `yaml:"phases"`
}

type rawMode struct {
	Phases    []string `yaml:"phases"`
	SkipGates bool     `yaml:"skip_gates"`
}

type rawPhase struct {
	Name        string          `yaml:"name"`
	Agent       string          `yaml:"agent"`
	Skill       string          `yaml:"skill"`
	Mode        string          `yaml:"mode"`
	Feedforward string          `yaml:"feedforward"`
	Gate        interface{}     `yaml:"gate"` // Can be string, null, or object
	GatePrompt  string          `yaml:"gate_prompt"`
	Metrics     map[string]bool `yaml:"metrics"` // Map of metric_name: true/false
}

type rawGate struct {
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
	Prompt      string `yaml:"prompt"`
}
