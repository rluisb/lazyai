package workflow

import (
	"fmt"
	"time"
)

// FallbackHandler manages failure recovery
type FallbackHandler struct {
	maxRetries    int
	retryDelay    time.Duration
	fallbackAgent string
}

// NewFallbackHandler creates a fallback handler
func NewFallbackHandler(maxRetries int, fallbackAgent string) *FallbackHandler {
	return &FallbackHandler{
		maxRetries:    maxRetries,
		retryDelay:    2 * time.Second,
		fallbackAgent: fallbackAgent,
	}
}

// OnAgentFailure handles agent execution failure
func (f *FallbackHandler) OnAgentFailure(attempt int, agent string, err error) (string, error) {
	if attempt >= f.maxRetries {
		if f.fallbackAgent != "" && f.fallbackAgent != agent {
			return f.fallbackAgent, fmt.Errorf("agent %s failed after %d attempts, falling back to %s: %v",
				agent, attempt, f.fallbackAgent, err)
		}
		return "", fmt.Errorf("agent %s failed after %d attempts, no fallback available: %v",
			agent, attempt, err)
	}

	// Retry with exponential backoff
	if f.retryDelay > 0 {
		time.Sleep(f.retryDelay * time.Duration(attempt))
	}

	return agent, nil // Retry with same agent
}

// OnTestFailure handles test failure during workflow
func (f *FallbackHandler) OnTestFailure(phase string, testOutput string) (string, error) {
	// Default: abort workflow
	return "", fmt.Errorf("phase %s failed tests: %s", phase, testOutput)
}

// OnTimeout handles timeout during workflow execution
func (f *FallbackHandler) OnTimeout(phase string, timeout time.Duration) (string, error) {
	// Default: abort workflow
	return "", fmt.Errorf("phase %s timed out after %v", phase, timeout)
}
