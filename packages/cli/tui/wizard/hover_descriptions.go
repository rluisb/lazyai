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

func optionsWithDescriptions(options []huh.Option[string], descriptions map[string]string) []huh.Option[string] {
	if len(options) == 0 {
		return options
	}
	out := make([]huh.Option[string], 0, len(options))
	for _, opt := range options {
		description := optionDescription(opt.Value, descriptions, "")
		label := opt.Key
		if description != "" && !strings.Contains(label, " — ") {
			label += " — " + trimDescription(description)
		}
		out = append(out, huh.NewOption(label, opt.Value))
	}
	return out
}

func trimDescription(description string) string {
	description = strings.TrimSuffix(strings.TrimSpace(description), ".")
	const max = 72
	if len(description) <= max {
		return description
	}
	return strings.TrimSpace(description[:max-1]) + "…"
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

var projectIdentityActionDescriptions = map[string]string{
	"continue": "Save these optional identity values and continue.",
}

var conflictStrategyDescriptions = map[string]string{
	"align":              "Review each conflicting file before deciding.",
	"backup-and-replace": "Back up existing files, then replace them with LazyAI output.",
	"skip":               "Keep current files unchanged and skip conflicting outputs.",
}

var installConsentDescriptions = map[string]string{
	"yes": "Continue knowing these optional CLIs may be needed.",
	"no":  "Continue without approving optional CLI install prompts.",
}

var finalInstallDescriptions = map[string]string{
	"yes":  "Apply the dry-run plan now.",
	"edit": "Return to the wizard with current answers prefilled.",
	"no":   "Cancel without installing files.",
}

var opencodeCommandDescriptions = map[string]string{
	"review":               "Reviews branch changes against project guidance.",
	"test":                 "Runs the configured project test command.",
	"commit":               "Drafts a conventional commit message.",
	"speckit.analyze":      "Runs Spec Kit analysis guidance.",
	"speckit.checklist":    "Runs Spec Kit checklist guidance.",
	"speckit.clarify":      "Runs Spec Kit clarification guidance.",
	"speckit.constitution": "Runs Spec Kit constitution guidance.",
	"speckit.implement":    "Runs Spec Kit implementation guidance.",
	"speckit.plan":         "Runs Spec Kit planning guidance.",
	"speckit.specify":      "Runs Spec Kit specification guidance.",
	"speckit.tasks":        "Runs Spec Kit task breakdown guidance.",
}

var opencodeModeDescriptions = map[string]string{
	"plan":  "Constrains OpenCode to planning-first behavior.",
	"audit": "Focuses OpenCode on risks, defects, and review findings.",
}

var opencodePluginDescriptions = map[string]string{
	"https://github.com/Opencode-DCP/opencode-dynamic-context-pruning": "Reduces OpenCode context bloat dynamically.",
	"https://github.com/spoons-and-mirrors/subtask2":                   "Adds background subtask helpers.",
	"https://github.com/JRedeker/opencode-shell-strategy":              "Improves shell execution strategy.",
	"https://github.com/boxpositron/envsitter-guard":                   "Guards environment-variable use.",
	"https://github.com/kdcokenny/opencode-background-agents":          "Enables async background-agent workflows.",
}
