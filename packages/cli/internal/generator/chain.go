package generator

import (
	"encoding/json"
	"fmt"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// ChainGenerator generates chain JSON definition files.
type ChainGenerator struct{}

func (g *ChainGenerator) Type() types.ArtifactType { return types.ArtifactTypeChain }

func (g *ChainGenerator) GetPromptQuestions() []PromptQuestion {
	return []PromptQuestion{
		{
			Key:      "description",
			Label:    "Chain description",
			Type:     "text",
			Required: false,
		},
	}
}

func (g *ChainGenerator) Generate(config GeneratorConfig) ([]GeneratedFile, error) {
	slug := ToSlug(config.Name)
	if slug == "" {
		slug = "new-chain"
	}

	description := config.Description
	if description == "" {
		description = fmt.Sprintf("%s orchestration chain", ToTitleCase(slug))
	}

	chain := map[string]any{
		"kind":        "chain",
		"name":        slug,
		"description": description,
		"version":     "1.0.0",
		"entry":       "step-1",
		"steps": []map[string]any{
			{
				"id":      "step-1",
				"agent":   "planner",
				"prompt":  "",
				"timeout": "5m",
				"transitions": map[string]string{
					"success": "done",
				},
			},
		},
	}

	content, err := json.MarshalIndent(chain, "", "  ")
	if err != nil {
		return nil, err
	}

	return []GeneratedFile{
		{
			Path:    fmt.Sprintf(".ai/orchestration/chains/%s.json", slug),
			Content: string(content) + "\n",
		},
	}, nil
}
