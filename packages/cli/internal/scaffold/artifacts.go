package scaffold

import (
	"github.com/rluisb/lazyai/packages/cli/internal/adapter"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// ScaffoldArtifacts installs agents, skills, and prompts for all selected tools
// using the adapter registry. Ported from src/scaffold/agents-skills-prompts.ts.
func (ctx *ScaffoldContext) ScaffoldArtifacts() ([]types.TrackedFile, error) {
	registry := adapter.NewRegistry()
	adapters, err := registry.GetForTools(ctx.Tools)
	if err != nil {
		return nil, err
	}

	var allRecords []types.TrackedFile

	for _, a := range adapters {
		adapterCtx := &adapter.AdapterContext{
			TargetDir:           ctx.TargetDir,
			WorkspaceRoot:       ctx.WorkspaceRoot,
			SetupScope:          ctx.SetupScope,
			HomeDir:             ctx.HomeDir,
			LibraryDir:          ctx.LibraryDir,
			LibraryFS:           ctx.LibraryFS,
			FileRecords:         []types.TrackedFile{},
			EnableServers:       ctx.EnableServers,
			Force:               ctx.Force,
			DryRun:              ctx.DryRun,
			DriveCLI:            ctx.DriveCLI,
			LocalSecrets:        ctx.LocalSecrets,
			FortniteMode:        ctx.FortniteMode,
			Strategy:            ctx.Strategy,
			PerFileOverrides:    ctx.PerFileOverrides,
			ConfiguredProviders: ctx.OpenCodeProviders,
			Selections: adapter.AdapterSelections{
				Agents:           ctx.Agents,
				Skills:           ctx.Skills,
				Prompts:          ctx.Prompts,
				ChatModes:        ctx.ChatModes,
				OpenCodeCommands: ctx.OpenCodeCommands,
				OpenCodeModes:    ctx.OpenCodeModes,
				OpenCodePlugins:  ctx.OpenCodePlugins,
			},
		}

		records, err := a.Install(adapterCtx)
		if err != nil {
			scaffoldLog.Warn("adapter install failed", "adapter", a.ID(), "error", err)
			continue
		}
		allRecords = append(allRecords, records...)
	}

	return allRecords, nil
}
