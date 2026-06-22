package wizard

import (
	"fmt"
	"strings"

	"charm.land/huh/v2"
)

const defaultHoverHint = "Move through options to see what changes."

func selectHoverDescription(field *huh.Select[string], descriptions map[string]string, fallback string) string {
	if field == nil {
		return fallback
	}
	value, ok := field.Hovered()
	if !ok {
		return fallback
	}
	return optionDescription(value, descriptions, fallback)
}

func multiSelectHoverDescription(field *huh.MultiSelect[string], descriptions map[string]string, fallback string) string {
	if field == nil {
		return fallback
	}
	value, ok := field.Hovered()
	if !ok {
		return fallback
	}
	return optionDescription(value, descriptions, fallback)
}

func optionDescription(value string, descriptions map[string]string, fallback string) string {
	if value == "" {
		return fallback
	}
	if description := descriptions[value]; description != "" {
		return description
	}
	if value == phase1BackValue || value == phase2BackValue || value == "__phase5_back__" {
		return "Return to the previous step without applying this screen."
	}
	return fmt.Sprintf("%s: no extra setup beyond selecting this item.", value)
}

func catalogServerDescriptions() map[string]string {
	catalog, err := loadMcpCatalog()
	if err != nil || catalog == nil {
		return nil
	}
	descriptions := make(map[string]string, len(catalog.Servers))
	for id, server := range catalog.Servers {
		description := strings.TrimSpace(server.Description)
		if server.RequiresInstall {
			description = strings.TrimSpace(description + " Requires a local CLI install.")
		}
		descriptions[id] = description
	}
	return descriptions
}

func catalogCliToolDescriptions() map[string]string {
	catalog, err := loadMcpCatalog()
	if err != nil || catalog == nil {
		return nil
	}
	descriptions := make(map[string]string, len(catalog.CliTools))
	for id, tool := range catalog.CliTools {
		description := strings.TrimSpace(tool.Description)
		if tool.InstallHint != "" {
			description = strings.TrimSpace(description + " " + tool.InstallHint)
		}
		descriptions[id] = description
	}
	return descriptions
}

var scopeDescriptions = map[string]string{
	"global":    "Install shared configuration for this user account.",
	"workspace": "Install planning assets for a multi-project workspace.",
	"project":   "Install repository-local assets; safest default for one repo.",
}

var toolDescriptions = map[string]string{
	"opencode":    "Adds OpenCode-native agents, commands, modes, and config.",
	"claude-code": "Adds Claude Code commands, agents, skills, and project config.",
	"copilot":     "Adds GitHub Copilot instructions and chat modes.",
	"pi":          "Adds Pi-compatible skills, agents, and rules.",
	"omp":         "Adds OMP-compatible skills, agents, and rules.",
	"kiro":        "Adds Kiro steering and agent assets.",
	"antigravity": "Adds Antigravity rules and generated configuration.",
}

var skillDescriptions = map[string]string{
	"architecture-review":  "Lightweight ADR-style review before structural changes.",
	"codebase-exploration": "Disciplined search/read flow before editing unfamiliar code.",
	"debugging":            "Hypothesis-driven failure investigation and root-cause fixes.",
	"implementation":       "Scoped implementation workflow with verification.",
	"planning":             "Turns requirements into ordered implementation tasks.",
	"pr-review":            "Evidence-first pull request review workflow.",
	"testing":              "Test-first and focused verification practices.",
}

var agentDescriptions = map[string]string{
	"guide":             "Coordinates workflow and keeps user-facing progress clear.",
	"implementer":       "Applies scoped code changes and focused verification.",
	"researcher":        "Maps code and docs before planning or implementation.",
	"deployer":          "Handles release, deployment, and rollout checks.",
	"responder":         "Drafts user-facing responses from grounded evidence.",
	"planner":           "Converts approved scope into concrete task order.",
	"reviewer":          "Checks code quality, risks, and missed edge cases.",
	"evidence-verifier": "Independently verifies claims against files and commands.",
}

var mcpPresetDescriptions = map[string]string{
	"minimal":     "Only core local setup servers.",
	"recommended": "Balanced default for normal project work.",
	"full":        "Every catalog server; more capability and more setup surface.",
}

var presetDescriptions = map[string]string{
	"minimal":  "Smallest ruleset: quality gates and git conventions.",
	"standard": "Recommended defaults: RPI, reasoning, and bug-resolution workflows.",
	"full":     "Enable every workflow feature.",
	"custom":   "Choose individual features and chat modes.",
}

var featureDescriptions = map[string]string{
	"qualityGates":       "Requires validation before work is considered complete.",
	"rpiWorkflow":        "Research, plan, implement flow with approval gates.",
	"chainOfThought":     "Adds concise reasoning protocol for non-trivial work.",
	"bugResolution":      "Root-cause bugfix process with regression checks.",
	"contextEngineering": "Keeps task context scoped and reusable.",
	"treeOfThoughts":     "Compares alternatives before architectural decisions.",
	"adrEnforcement":     "Prompts ADR capture for durable architecture choices.",
	"agentHarness":       "Adds multi-agent coordination guidance.",
	"pivotHandling":      "Handles requirement changes without losing state.",
	"adversarialDesign":  "Adds explicit challenge/review pass for designs.",
}

var branchPatternDescriptions = map[string]string{
	"{type}/{ticket}-{description}": "Conventional ticket branch, good default for Jira teams.",
	"{type}/{ticket}/{description}": "Nested ticket branch used by some Jira workflows.",
	"{type}/{description}":          "Simple branch when tickets are optional.",
	"{ticket}/{description}":        "Ticket-first branch for strict issue tracking.",
	"{description}":                 "Shortest branch; least policy metadata.",
	"custom":                        "Enter your own branch template next.",
}

var commitPatternDescriptions = map[string]string{
	"{type}({scope}): {description}": "Conventional Commits with optional scope.",
	"{type}: {description}":          "Simple conventional type prefix.",
	"[{ticket}] {description}":       "Ticket-first message for Jira traceability.",
	"{description}":                  "Plain message; least policy metadata.",
	"custom":                         "Enter your own commit template next.",
}

var chatModeDescriptions = map[string]string{
	"architect": "Copilot mode for architecture and design tradeoffs.",
	"reviewer":  "Copilot mode for code review and risk checks.",
}
