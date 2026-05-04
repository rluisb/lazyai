package dispatch

import (
	"context"
	"fmt"

	"github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/types"
)

// InvocationRequest captures the host-native agent invocation that the MCP
// invoke_agent tool currently composes. Phase 1 intentionally keeps this as a
// prompt/spec handoff rather than executing any agent process directly.
type InvocationRequest struct {
	Agent string
	Task  string
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
