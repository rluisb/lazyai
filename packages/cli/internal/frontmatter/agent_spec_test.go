package frontmatter

import (
	"strings"
	"testing"
)

func TestParseAgentSpec_LibraryAgentShape(t *testing.T) {
	src := []byte(`---
name: Planner
tier: frontier
temperature: 0.5
thinking: high
risk: 5
tools: memory qmd
---

# Planner Agent
`)
	got, err := ParseAgentSpec(src)
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if got.Name != "Planner" || got.Tier != "frontier" || got.Thinking != "high" {
		t.Errorf("strings: %+v", got)
	}
	if got.Temperature != 0.5 || got.Risk != 5 {
		t.Errorf("numbers: temp=%v risk=%v", got.Temperature, got.Risk)
	}
	if got.Multimodal {
		t.Error("multimodal should default false")
	}
}

func TestParseAgentSpec_IntegerTemperature(t *testing.T) {
	src := []byte(`---
name: T
tier: balanced
temperature: 0
thinking: none
risk: 1
---
`)
	got, err := ParseAgentSpec(src)
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if got.Temperature != 0.0 {
		t.Errorf("expected 0.0, got %v", got.Temperature)
	}
}

func TestParseAgentSpec_MultimodalFlag(t *testing.T) {
	src := []byte(`---
name: Vision
tier: frontier
temperature: 0.1
thinking: medium
risk: 3
multimodal: true
---
`)
	got, _ := ParseAgentSpec(src)
	if !got.Multimodal {
		t.Error("multimodal: true should parse as true")
	}
}

func TestParseAgentSpec_MissingTierIsError(t *testing.T) {
	src := []byte(`---
name: Stale
model: opus
---
`)
	_, err := ParseAgentSpec(src)
	if err == nil || !strings.Contains(err.Error(), "tier") {
		t.Errorf("expected tier-missing error, got %v", err)
	}
}

func TestParseAgentSpec_NoFrontmatterIsError(t *testing.T) {
	_, err := ParseAgentSpec([]byte("just a body, no frontmatter"))
	if err == nil {
		t.Error("expected error for missing frontmatter")
	}
}
