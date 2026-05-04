package mcp

import (
	"fmt"
	"strings"

	"github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/dispatch"
	"github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/types"
)

type RuntimeConfig struct {
	ExecutionMode types.ExecutionMode
}

func DefaultRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{ExecutionMode: types.ExecutionNative}
}

func NewRuntimeConfig(mode string) (RuntimeConfig, error) {
	trimmed := strings.TrimSpace(mode)
	if trimmed == "" {
		return DefaultRuntimeConfig(), nil
	}

	switch types.ExecutionMode(trimmed) {
	case types.ExecutionNative, types.ExecutionA2A, types.ExecutionHybrid:
		return RuntimeConfig{ExecutionMode: types.ExecutionMode(trimmed)}, nil
	default:
		return RuntimeConfig{}, fmt.Errorf("invalid execution mode %q (expected native, a2a, or hybrid)", mode)
	}
}

type OrchestratorOption func(*Orchestrator)

func WithRuntimeConfig(config RuntimeConfig) OrchestratorOption {
	return func(o *Orchestrator) {
		o.Runtime = config
		o.Dispatcher = defaultDispatcherFor(config)
	}
}

func WithDispatcher(dispatcher dispatch.Dispatcher) OrchestratorOption {
	return func(o *Orchestrator) {
		if dispatcher != nil {
			o.Dispatcher = dispatcher
		}
	}
}

func defaultDispatcherFor(config RuntimeConfig) dispatch.Dispatcher {
	switch config.ExecutionMode {
	case types.ExecutionA2A:
		return dispatch.NewDisabledA2ADispatcher()
	case types.ExecutionHybrid:
		// Phase 1 has no configured A2A agents/client. Hybrid remains opt-in and
		// safe by falling back to the existing host-native dispatcher.
		return dispatch.NewNativeDispatcher()
	case types.ExecutionNative:
		fallthrough
	default:
		return dispatch.NewNativeDispatcher()
	}
}
