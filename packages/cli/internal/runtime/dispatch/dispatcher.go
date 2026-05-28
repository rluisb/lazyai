package dispatch

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// Result contains the outcome of a dispatch
type Result struct {
	Agent      string
	Task       string
	Mode       string
	Output     string
	Error      error
	StartedAt  time.Time
	FinishedAt time.Time
}

// Dispatcher executes agents via AI tools
type Dispatcher interface {
	Dispatch(sessionID string, agent string, task string, mode string) (*Result, error)
}

// OpenCodeDispatcher implements Dispatcher for OpenCode
type OpenCodeDispatcher struct {
	Timeout time.Duration
}

// NewOpenCodeDispatcher creates an OpenCode dispatcher
func NewOpenCodeDispatcher() *OpenCodeDispatcher {
	return &OpenCodeDispatcher{
		Timeout: 5 * time.Minute,
	}
}

// Dispatch executes an agent via opencode
func (d *OpenCodeDispatcher) Dispatch(sessionID string, agent string, task string, mode string) (*Result, error) {
	result := &Result{
		Agent:     agent,
		Task:      task,
		Mode:      mode,
		StartedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout)
	defer cancel()

	// Build opencode command
	args := []string{"run"}
	if agent != "" {
		args = append(args, "--agent", agent)
	}
	if mode != "" {
		args = append(args, "--mode", mode)
	}
	args = append(args, task)

	cmd := exec.CommandContext(ctx, "opencode", args...)
	output, err := cmd.CombinedOutput()

	result.FinishedAt = time.Now()
	result.Output = string(output)

	if err != nil {
		result.Error = fmt.Errorf("opencode dispatch failed: %w\nOutput: %s", err, output)
		return result, result.Error
	}

	return result, nil
}

// MockDispatcher is a test dispatcher that simulates execution
type MockDispatcher struct {
	Responses map[string]*Result
}

// NewMockDispatcher creates a mock dispatcher for testing
func NewMockDispatcher() *MockDispatcher {
	return &MockDispatcher{
		Responses: make(map[string]*Result),
	}
}

// Dispatch simulates agent execution for testing
func (d *MockDispatcher) Dispatch(sessionID string, agent string, task string, mode string) (*Result, error) {
	key := agent + ":" + task
	if resp, ok := d.Responses[key]; ok {
		return resp, resp.Error
	}

	// Default success response
	return &Result{
		Agent:      agent,
		Task:       task,
		Mode:       mode,
		Output:     "Mock execution completed",
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
	}, nil
}
