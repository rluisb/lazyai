package scaffold

import (
	"log"

	"github.com/ricardoborges-teachable/ai-setup/internal/adapter"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
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
			TargetDir:        ctx.TargetDir,
			SetupScope:       ctx.SetupScope,
			HomeDir:          ctx.HomeDir,
			LibraryDir:       ctx.LibraryDir,
			LibraryFS:        ctx.LibraryFS,
			FileRecords:      []types.TrackedFile{},
			EnableServers:    ctx.EnableServers,
			Force:            ctx.Force,
			DryRun:           ctx.DryRun,
			DriveCLI:         ctx.DriveCLI,
			Strategy:         ctx.Strategy,
			PerFileOverrides: ctx.PerFileOverrides,
			Selections: adapter.AdapterSelections{
				Agents:           ctx.Agents,
				Skills:           ctx.Skills,
				Prompts:          ctx.Prompts,
				Commands:         ctx.Commands,
				ChatModes:        ctx.ChatModes,
				OpenCodeCommands: ctx.OpenCodeCommands,
				OpenCodeModes:    ctx.OpenCodeModes,
			},
		}

		records, err := a.Install(adapterCtx)
		if err != nil {
			log.Printf("Warning: adapter %s install failed: %v", a.ID(), err)
			continue
		}
		allRecords = append(allRecords, records...)
	}

	return allRecords, nil
}
