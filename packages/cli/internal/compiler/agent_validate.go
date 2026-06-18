package compiler

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/auth"
	"github.com/rluisb/lazyai/packages/cli/internal/frontmatter"
	"github.com/rluisb/lazyai/packages/cli/internal/models"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// AgentValidationIssue describes one (agent, tool) pair where models.Resolve
// could not produce a valid model field. Surfaced by ValidateAgentResolutions
// during compile so users see broken tier/auth combinations before any
// per-tool agent file is written.
type AgentValidationIssue struct {
	Agent string
	Tool  types.ToolId
	Err   error
}

func (i AgentValidationIssue) String() string {
	return fmt.Sprintf("agent %q on %s: %v", i.Agent, i.Tool, i.Err)
}

// ValidateAgentResolutions walks every library/agents/*.md and runs
// models.Resolve against each selected tool's catalog. Returns a slice of
// issues; an empty slice means every agent resolves on every tool.
//
// configuredProviders is the user-authenticated provider set used by
// OpenCode's RequireConfigured filter. Pass nil to fall back to a live
// auth.DetectAll probe (matches the behaviour of adapter installs that
// run before the wizard populates the store).
func ValidateAgentResolutions(libFS fs.FS, tools []types.ToolId, configuredProviders []string) ([]AgentValidationIssue, error) {
	if libFS == nil {
		return nil, nil
	}
	entries, err := fs.ReadDir(libFS, "agents")
	if err != nil {
		return nil, nil
	}

	if configuredProviders == nil {
		probeCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		for _, p := range auth.DetectAll(probeCtx) {
			configuredProviders = append(configuredProviders, string(p))
		}
	}

	var issues []AgentValidationIssue
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".md") {
			continue
		}
		data, err := fs.ReadFile(libFS, path.Join("agents", ent.Name()))
		if err != nil {
			continue
		}
		raw, err := frontmatter.ParseAgentSpec(data)
		if err != nil {
			issues = append(issues, AgentValidationIssue{
				Agent: ent.Name(),
				Err:   err,
			})
			continue
		}
		spec := models.AgentSpec{
			Name:        raw.Name,
			Tier:        models.Tier(raw.Tier),
			Temperature: raw.Temperature,
			Thinking:    models.Thinking(raw.Thinking),
			Risk:        raw.Risk,
			Multimodal:  raw.Multimodal,
		}
		for _, tool := range tools {
			cat := models.CatalogFor(tool)
			rc := models.ResolveCtx{
				Catalog:             cat,
				ConfiguredProviders: configuredProviders,
			}
			if _, err := models.Resolve(spec, rc); err != nil {
				issues = append(issues, AgentValidationIssue{
					Agent: raw.Name,
					Tool:  tool,
					Err:   err,
				})
			}
		}
	}
	return issues, nil
}

// FormatAgentValidationIssues returns a multi-line, human-readable rendering
// suitable for cmd-level error/warn output. Empty input yields empty string.
func FormatAgentValidationIssues(issues []AgentValidationIssue) string {
	if len(issues) == 0 {
		return ""
	}
	var b strings.Builder
	for _, i := range issues {
		b.WriteString("  - ")
		b.WriteString(i.String())
		b.WriteString("\n")
	}
	return b.String()
}
