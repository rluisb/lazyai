package generator

import (
	"fmt"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// Registry holds all registered artifact generators indexed by ArtifactType.
// Ported from src/generators/registry.ts.
type Registry struct {
	generators map[types.ArtifactType]Generator
}

// NewRegistry creates a Registry with all built-in generators registered.
func NewRegistry() *Registry {
	r := &Registry{
		generators: make(map[types.ArtifactType]Generator),
	}

	r.register(&AgentGenerator{})
	r.register(&SkillGenerator{})
	r.register(&CommandGenerator{})
	r.register(&PromptGenerator{})
	r.register(&TemplateGenerator{})

	return r
}

func (r *Registry) register(g Generator) {
	r.generators[g.Type()] = g
}

// Get returns the generator for the given artifact type, or an error if not found.
func (r *Registry) Get(artifactType types.ArtifactType) (Generator, error) {
	g, ok := r.generators[artifactType]
	if !ok {
		return nil, fmt.Errorf("no generator registered for artifact type %q", artifactType)
	}
	return g, nil
}

// ListTypes returns all registered artifact types.
func (r *Registry) ListTypes() []types.ArtifactType {
	types_ := make([]types.ArtifactType, 0, len(r.generators))
	for t := range r.generators {
		types_ = append(types_, t)
	}
	return types_
}
