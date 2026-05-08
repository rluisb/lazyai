package frontmatter

import (
	"fmt"
	"strings"
)

// AgentSpecRaw is the per-agent capability declaration parsed from the
// source frontmatter. It is the input shape for models.Resolve, which lives
// in another package — kept here as a plain struct (no models import) to
// avoid an import cycle. Adapters convert this to models.AgentSpec at the
// call site.
//
// Source agents under library/agents/*.md declare:
//
//	tier: frontier | balanced | speed
//	temperature: <float>
//	thinking: high | medium | low | none
//	risk: 1-5
//	multimodal: true   # optional, defaults false
//
// All five fields are required for non-multimodal agents. Adapters that
// emit per-target frontmatter use AgentSpecRaw to drive both model
// selection (via Resolve) and the per-target translation of thinking/
// temperature into target-specific keys.
type AgentSpecRaw struct {
	Name        string
	Tier        string
	Temperature float64
	Thinking    string
	Risk        int
	Multimodal  bool
}

// ParseAgentSpec extracts an AgentSpecRaw from agent source content. Returns
// an error if the frontmatter is missing or any required field is absent —
// callers should treat that as a migration failure (the agent has not been
// updated to the tier-based schema).
func ParseAgentSpec(source []byte) (AgentSpecRaw, error) {
	fm, _, err := ExtractFrontmatter(source)
	if err != nil {
		return AgentSpecRaw{}, fmt.Errorf("parse frontmatter: %w", err)
	}
	if len(fm) == 0 {
		return AgentSpecRaw{}, fmt.Errorf("agent has no frontmatter")
	}

	spec := AgentSpecRaw{
		Name:     ExtractField(fm, "name"),
		Tier:     strings.ToLower(strings.TrimSpace(ExtractField(fm, "tier"))),
		Thinking: strings.ToLower(strings.TrimSpace(ExtractField(fm, "thinking"))),
	}
	if t, ok := coerceFloat(fm["temperature"]); ok {
		spec.Temperature = t
	}
	if r, ok := coerceInt(fm["risk"]); ok {
		spec.Risk = r
	}
	if mm, ok := fm["multimodal"].(bool); ok {
		spec.Multimodal = mm
	}

	if spec.Tier == "" {
		return spec, fmt.Errorf("agent %q missing required field: tier", spec.Name)
	}
	return spec, nil
}

// coerceFloat handles YAML's int-or-float ambiguity for temperature: 0 is
// parsed as int, 0.5 as float64. Both are valid; we accept either.
func coerceFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	}
	return 0, false
}

func coerceInt(v any) (int, bool) {
	switch x := v.(type) {
	case int:
		return x, true
	case int64:
		return int(x), true
	case float64:
		return int(x), true
	}
	return 0, false
}
