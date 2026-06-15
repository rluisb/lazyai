package domain

import "github.com/rluisb/lazyai/packages/orchestrator/internal/types"

// InvocationRequest captures a host-native or remote agent invocation request.
type InvocationRequest struct {
	Agent   string
	Task    string
	CliTool string
}

// InvocationResult is the normalized result returned by agent invocation ports.
type InvocationResult struct {
	Spec          *types.ComposedAgentSpec
	ExecutionMode types.ExecutionMode
}
