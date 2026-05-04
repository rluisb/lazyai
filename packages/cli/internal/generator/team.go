package generator

import (
	"encoding/json"
	"fmt"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// TeamGenerator generates team JSON definition files.
type TeamGenerator struct{}

func (g *TeamGenerator) Type() types.ArtifactType { return types.ArtifactTypeTeam }

func (g *TeamGenerator) GetPromptQuestions() []PromptQuestion {
	return []PromptQuestion{
		{
			Key:      "description",
			Label:    "Team description",
			Type:     "text",
			Required: false,
		},
		{
			Key:      "defaultChain",
			Label:    "Default chain reference",
			Type:     "text",
			Required: false,
			Default:  "feature",
		},
	}
}

func (g *TeamGenerator) Generate(config GeneratorConfig) ([]GeneratedFile, error) {
	slug := ToSlug(config.Name)
	if slug == "" {
		slug = "new-team"
	}

	description := config.Description
	if description == "" {
		description = fmt.Sprintf("%s team", ToTitleCase(slug))
	}

	defaultChain := getAnswer(config.Answers, "defaultChain", "feature")

	team := map[string]any{
		"kind":         "team",
		"name":         slug,
		"description":  description,
		"version":      "1.0.0",
		"members":      []string{"scout", "builder", "reviewer"},
		"defaultChain": defaultChain,
	}

	content, err := json.MarshalIndent(team, "", "  ")
	if err != nil {
		return nil, err
	}

	return []GeneratedFile{
		{
			Path:    fmt.Sprintf(".ai/orchestration/teams/%s.json", slug),
			Content: string(content) + "\n",
		},
	}, nil
}
