package ports

import (
	"context"

	"github.com/rluisb/lazyai/packages/orchestrator/domain"
)

// AgentInvoker is the outbound port for resolving an agent invocation into an execution handoff.
type AgentInvoker interface {
	InvokeAgent(ctx context.Context, req domain.InvocationRequest) (*domain.InvocationResult, error)
}
