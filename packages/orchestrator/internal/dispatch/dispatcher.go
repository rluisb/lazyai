package dispatch

import (
	"context"
	"fmt"
	"strings"

	oconfig "github.com/rluisb/lazyai/packages/orchestrator/internal/config"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

// InvocationRequest captures the host-native agent invocation that the MCP
// invoke_agent tool currently composes. Phase 1 intentionally keeps this as a
// prompt/spec handoff rather than executing any agent process directly.
type InvocationRequest struct {
	Agent   string
	Task    string
	CliTool string
}

// InvocationResult is the dispatcher seam's normalized response. Native mode
// returns Spec with the existing invoke_agent JSON shape; disabled A2A paths
// return an error before this result is emitted.
type InvocationResult struct {
	Spec          *types.ComposedAgentSpec
	ExecutionMode types.ExecutionMode
}

// Dispatcher resolves an invocation request into the execution handoff for the
// selected runtime mode.
type Dispatcher interface {
	InvokeAgent(ctx context.Context, req InvocationRequest) (*InvocationResult, error)
}

// NativeDispatcher preserves the existing MCP invoke_agent behavior: produce a
// composed agent spec for the host CLI to execute natively.
type NativeDispatcher struct{}

func NewNativeDispatcher() *NativeDispatcher {
	return &NativeDispatcher{}
}

func (d *NativeDispatcher) InvokeAgent(ctx context.Context, req InvocationRequest) (*InvocationResult, error) {
	return &InvocationResult{
		ExecutionMode: types.ExecutionNative,
		Spec: &types.ComposedAgentSpec{
			ID: req.Agent, Base: req.Agent, Model: "sonnet",
			Prompt: fmt.Sprintf("Task: %s\nExecute as the %s agent.", req.Task, req.Agent),
		},
	}, nil
}

// DisabledA2ADispatcher is the Phase 1 safety stub. It makes opt-in A2A mode
// explicit without adding a client, network transport, or provider dependency.
type DisabledA2ADispatcher struct{}

func NewDisabledA2ADispatcher() *DisabledA2ADispatcher {
	return &DisabledA2ADispatcher{}
}

func (d *DisabledA2ADispatcher) InvokeAgent(ctx context.Context, req InvocationRequest) (*InvocationResult, error) {
	return nil, fmt.Errorf("A2A execution mode is not configured: Phase 1 does not include an A2A client; use execution-mode native or hybrid without A2A config for native fallback")
}

// ConfiguredDispatcher applies Phase 2 runtime selection from orchestrator
// config. It intentionally does not execute A2A network calls until Phase 3.
type ConfiguredDispatcher struct {
	mode   types.ExecutionMode
	config *oconfig.Config
	native *NativeDispatcher
}

func NewConfiguredDispatcher(mode types.ExecutionMode, config *oconfig.Config) *ConfiguredDispatcher {
	return &ConfiguredDispatcher{mode: mode, config: config, native: NewNativeDispatcher()}
}

func (d *ConfiguredDispatcher) InvokeAgent(ctx context.Context, req InvocationRequest) (*InvocationResult, error) {
	switch d.mode {
	case types.ExecutionA2A:
		provider, err := d.resolveA2ATarget(req)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("A2A provider %q for agent %q is configured, but A2A network execution is not implemented until Phase 3; use execution-mode native or hybrid for native fallback", provider, req.Agent)
	case types.ExecutionHybrid:
		// Phase 2 selects providers explicitly but keeps hybrid safe by preserving
		// the existing host-native response shape until the Phase 3 client exists.
		return d.native.InvokeAgent(ctx, req)
	case types.ExecutionNative:
		fallthrough
	default:
		return d.native.InvokeAgent(ctx, req)
	}
}

func (d *ConfiguredDispatcher) resolveA2ATarget(req InvocationRequest) (string, error) {
	if d.config == nil {
		return "", fmt.Errorf("A2A execution mode is not configured: no orchestrator config loaded; use execution-mode native or provide .ai/orchestrator.json")
	}
	agent, ok := d.config.Agents[req.Agent]
	if !ok {
		return "", fmt.Errorf("A2A agent %q is not configured", req.Agent)
	}
	if !agent.IsEnabled() {
		return "", fmt.Errorf("A2A agent %q is disabled", req.Agent)
	}
	if strings.TrimSpace(req.CliTool) != "" && len(agent.Tools) > 0 && !containsTool(agent.Tools, req.CliTool) {
		return "", fmt.Errorf("A2A agent %q is not enabled for cliTool %q", req.Agent, req.CliTool)
	}
	if _, ok := d.config.Providers[agent.Provider]; !ok {
		return "", fmt.Errorf("A2A agent %q references unknown provider %q", req.Agent, agent.Provider)
	}
	return agent.Provider, nil
}

func containsTool(tools []string, target string) bool {
	for _, tool := range tools {
		if tool == target {
			return true
		}
	}
	return false
}
