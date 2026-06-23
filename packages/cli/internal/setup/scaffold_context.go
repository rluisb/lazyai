package setup

import (
	"fmt"
	"io/fs"

	"github.com/rluisb/lazyai/packages/cli/internal/preset"
	"github.com/rluisb/lazyai/packages/cli/internal/scaffold"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// Library carries the library adapter dependencies needed to build a scaffold context.
type Library struct {
	Dir string
	FS  fs.FS
}

// AddSelections contains artifacts requested by the add command after CLI/UI parsing.
type AddSelections struct {
	Tools  []types.ToolId
	Agents []string
	Skills []string
}

// UpdateOptions contains update command behavior after CLI/UI parsing.
type UpdateOptions struct {
	Force  bool
	DryRun bool
}

// BuildAddScaffoldContext merges requested artifacts into stored setup state and
// returns the scaffold context needed to re-run the scaffold pipeline.
func BuildAddScaffoldContext(targetDir string, library Library, storeData *types.StoreData, selections AddSelections) (*scaffold.ScaffoldContext, types.PresetLevel, error) {
	if storeData == nil {
		return nil, "", fmt.Errorf("store data is required")
	}

	merged := *storeData
	merged.Config.Tools = mergeToolIDs(storeData.Config.Tools, selections.Tools)
	merged.Config.CLITools = mergeStrings(storeData.Config.CLITools, toolIDsToStrings(selections.Tools))
	merged.Selections.Agents = mergeAgentIDs(storeData.Selections.Agents, selections.Agents)
	merged.Selections.Skills = mergeSkillIDs(storeData.Selections.Skills, selections.Skills)

	presetLevel := preset.DefaultPresetForScope(merged.Config.SetupScope)
	ctx := buildStoredScaffoldContext(targetDir, library, &merged, presetLevel, types.ConflictStrategyAlign, false, false)
	return ctx, presetLevel, nil
}

// BuildUpdateScaffoldContext returns the scaffold context needed to update an
// existing setup from stored state.
func BuildUpdateScaffoldContext(targetDir string, library Library, storeData *types.StoreData, options UpdateOptions) (*scaffold.ScaffoldContext, types.PresetLevel, error) {
	if storeData == nil {
		return nil, "", fmt.Errorf("store data is required")
	}

	strategy := types.ConflictStrategyAlign
	if options.Force {
		strategy = types.ConflictStrategyBackupAndReplace
	}

	presetLevel := preset.DefaultPresetForScope(storeData.Config.SetupScope)
	ctx := buildStoredScaffoldContext(targetDir, library, storeData, presetLevel, strategy, options.Force, options.DryRun)
	ctx.StoreData = storeData
	return ctx, presetLevel, nil
}

func buildStoredScaffoldContext(targetDir string, library Library, storeData *types.StoreData, presetLevel types.PresetLevel, strategy types.ConflictStrategy, force, dryRun bool) *scaffold.ScaffoldContext {
	return &scaffold.ScaffoldContext{
		TargetDir:         targetDir,
		LibraryDir:        library.Dir,
		LibraryFS:         library.FS,
		Tools:             storeData.Config.Tools,
		CLITools:          storeData.Config.CLITools,
		EnableServers:     storeData.Config.EnableServers,
		ProjectName:       storeData.Config.ProjectName,
		PlanningDir:       storeData.Config.PlanningDir,
		SetupScope:        storeData.Config.SetupScope,
		Features:          storeData.Selections.Features,
		GitConventions:    storeData.Selections.GitConventions,
		Strategy:          strategy,
		Force:             force,
		DryRun:            dryRun,
		Agents:            storeData.Selections.Agents,
		Skills:            storeData.Selections.Skills,
		Prompts:           storeData.Selections.Prompts,
		ChatModes:         storeData.Selections.ChatModes,
		OpenCodeCommands:  storeData.Selections.OpenCodeCommands,
		OpenCodeModes:     storeData.Selections.OpenCodeModes,
		OpenCodeProviders: storeData.Selections.OpenCodeProviders,
		Templates:         storeData.Selections.Templates,
		Rules:             storeData.Selections.Rules,
		Infra:             storeData.Selections.Infra,
		SpecsDirs:         preset.SpecsDirsForPreset(presetLevel),
		Housekeeping:      storeData.Config.Housekeeping,
	}
}

func mergeToolIDs(existing, incoming []types.ToolId) []types.ToolId {
	seen := make(map[types.ToolId]bool, len(existing)+len(incoming))
	result := make([]types.ToolId, 0, len(existing)+len(incoming))
	for _, tool := range existing {
		if !seen[tool] {
			result = append(result, tool)
			seen[tool] = true
		}
	}
	for _, tool := range incoming {
		if !seen[tool] {
			result = append(result, tool)
			seen[tool] = true
		}
	}
	return result
}

func mergeAgentIDs(existing []types.AgentId, incoming []string) []types.AgentId {
	seen := make(map[types.AgentId]bool, len(existing)+len(incoming))
	result := make([]types.AgentId, 0, len(existing)+len(incoming))
	for _, agent := range existing {
		if !seen[agent] {
			result = append(result, agent)
			seen[agent] = true
		}
	}
	for _, agent := range incoming {
		id := types.AgentId(agent)
		if !seen[id] {
			result = append(result, id)
			seen[id] = true
		}
	}
	return result
}

func mergeSkillIDs(existing []types.SkillId, incoming []string) []types.SkillId {
	seen := make(map[types.SkillId]bool, len(existing)+len(incoming))
	result := make([]types.SkillId, 0, len(existing)+len(incoming))
	for _, skill := range existing {
		if !seen[skill] {
			result = append(result, skill)
			seen[skill] = true
		}
	}
	for _, skill := range incoming {
		id := types.SkillId(skill)
		if !seen[id] {
			result = append(result, id)
			seen[id] = true
		}
	}
	return result
}

func mergeStrings(existing, incoming []string) []string {
	seen := make(map[string]bool, len(existing)+len(incoming))
	result := make([]string, 0, len(existing)+len(incoming))
	for _, item := range existing {
		if !seen[item] {
			result = append(result, item)
			seen[item] = true
		}
	}
	for _, item := range incoming {
		if !seen[item] {
			result = append(result, item)
			seen[item] = true
		}
	}
	return result
}

func toolIDsToStrings(ids []types.ToolId) []string {
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = string(id)
	}
	return result
}
