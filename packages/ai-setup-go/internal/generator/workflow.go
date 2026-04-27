package generator

import (
	"encoding/json"
	"fmt"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// WorkflowGenerator generates workflow JSON definition files.
// Ported from src/generators/workflow.ts.
type WorkflowGenerator struct{}

func (g *WorkflowGenerator) Type() types.ArtifactType { return types.ArtifactTypeWorkflow }

func (g *WorkflowGenerator) GetPromptQuestions() []PromptQuestion {
	return []PromptQuestion{
		{
			Key:      "chain",
			Label:    "Primary chain reference",
			Type:     "text",
			Required: false,
			Default:  "feature",
		},
		{
			Key:      "team",
			Label:    "Optional review/synthesis team reference",
			Type:     "text",
			Required: false,
		},
	}
}

func (g *WorkflowGenerator) Generate(config GeneratorConfig) ([]GeneratedFile, error) {
	slug := ToSlug(config.Name)
	if slug == "" {
		slug = "new-workflow"
	}

	chainRef := getAnswer(config.Answers, "chain", "feature")
	if chainRef == "" {
		chainRef = "feature"
	}
	teamRef := getAnswer(config.Answers, "team", "")

	chainPhaseID := "run-" + ToSlug(chainRef)
	if chainPhaseID == "run-" {
		chainPhaseID = "run-chain"
	}

	teamPhaseID := ""
	if teamRef != "" {
		teamPhaseID = "run-" + ToSlug(teamRef)
		if teamPhaseID == "run-" {
			teamPhaseID = "run-team"
		}
	}

	type PhaseOn struct {
		Success string `json:"success"`
		Failure string `json:"failure"`
	}

	type Phase struct {
		ID   string  `json:"id"`
		Kind string  `json:"kind"`
		Ref  string  `json:"ref,omitempty"`
		On   PhaseOn `json:"on,omitempty"`
	}

	phases := []any{}

	successTarget := "complete"
	if teamPhaseID != "" {
		successTarget = teamPhaseID
	}

	chainPhase := map[string]any{
		"id":   chainPhaseID,
		"kind": "chain",
		"ref":  chainRef,
		"on": map[string]string{
			"success": successTarget,
			"failure": "handoff",
		},
	}
	phases = append(phases, chainPhase)

	if teamPhaseID != "" {
		teamPhase := map[string]any{
			"id":   teamPhaseID,
			"kind": "team",
			"ref":  teamRef,
			"on": map[string]string{
				"success": "complete",
				"failure": "handoff",
			},
		}
		phases = append(phases, teamPhase)
	}

	phases = append(phases,
		map[string]any{"id": "handoff", "kind": "terminal"},
		map[string]any{"id": "complete", "kind": "terminal"},
	)

	description := config.Description
	if description == "" {
		description = fmt.Sprintf("Workflow scaffold for %s.", slug)
	}

	workflow := map[string]any{
		"kind":        "workflow",
		"name":        slug,
		"description": description,
		"version":     "1.0.0",
		"entry":       chainPhaseID,
		"phases":      phases,
	}

	content, err := json.MarshalIndent(workflow, "", "  ")
	if err != nil {
		return nil, err
	}

	return []GeneratedFile{
		{
			Path:    fmt.Sprintf(".ai/orchestration/workflows/%s.json", slug),
			Content: string(content) + "\n",
		},
	}, nil
}
