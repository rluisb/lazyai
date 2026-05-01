// Package adapter provides the adapter registry that maps tool IDs to their
// adapter implementations. Ported from the TypeScript AdapterRegistry.
package adapter

import (
	"fmt"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// Registry holds all registered tool adapters indexed by their ToolId.
type Registry struct {
	adapters map[types.ToolId]ToolAdapter
}

// NewRegistry creates a Registry with all built-in adapters registered.
func NewRegistry() *Registry {
	r := &Registry{
		adapters: make(map[types.ToolId]ToolAdapter),
	}

	r.register(&OpenCodeAdapter{})
	r.register(&ClaudeCodeAdapter{})
	r.register(&CopilotAdapter{})

	return r
}

func (r *Registry) register(a ToolAdapter) {
	r.adapters[a.ID()] = a
}

// Get returns the adapter for the given tool ID, or an error if not found.
func (r *Registry) Get(id types.ToolId) (ToolAdapter, error) {
	a, ok := r.adapters[id]
	if !ok {
		return nil, fmt.Errorf("unsupported tool %q (supported tools: opencode, claude-code, copilot)", id)
	}
	return a, nil
}

// List returns all registered tool IDs.
func (r *Registry) List() []types.ToolId {
	ids := make([]types.ToolId, 0, len(r.adapters))
	for id := range r.adapters {
		ids = append(ids, id)
	}
	return ids
}

// GetForTools returns adapters for all the given tool IDs. Returns an error
// if any tool ID has no registered adapter.
func (r *Registry) GetForTools(toolIds []types.ToolId) ([]ToolAdapter, error) {
	result := make([]ToolAdapter, 0, len(toolIds))
	for _, id := range toolIds {
		a, err := r.Get(id)
		if err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, nil
}
