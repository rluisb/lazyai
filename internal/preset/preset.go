// Package preset defines preset configurations that map PresetLevel values to
// feature flags, specs directories, templates, and rules. It is ported from
// the TypeScript module src/presets.ts.
package preset

import (
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// PresetFeatures maps each non-custom PresetLevel to its FeatureFlags,
// matching the TypeScript PRESET_FEATURES record.
var PresetFeatures = map[types.PresetLevel]types.FeatureFlags{
	types.PresetLevelMinimal: {
		ContextEngineering: false,
		RPIWorkflow:        false,
		ChainOfThought:     false,
		TreeOfThoughts:     false,
		ADREnforcement:     false,
		QualityGates:       true,
		AgentHarness:       false,
		BugResolution:      false,
		PivotHandling:      false,
	},
	types.PresetLevelStandard: {
		ContextEngineering: false,
		RPIWorkflow:        true,
		ChainOfThought:     true,
		TreeOfThoughts:     false,
		ADREnforcement:     false,
		QualityGates:       true,
		AgentHarness:       false,
		BugResolution:      true,
		PivotHandling:      false,
	},
	types.PresetLevelFull: {
		ContextEngineering: true,
		RPIWorkflow:        true,
		ChainOfThought:     true,
		TreeOfThoughts:     true,
		ADREnforcement:     true,
		QualityGates:       true,
		AgentHarness:       true,
		BugResolution:      true,
		PivotHandling:      true,
	},
}

// DefaultPresetForScope returns the default PresetLevel for a given SetupScope.
// Global scope defaults to minimal; project and workspace default to standard.
func DefaultPresetForScope(scope types.SetupScope) types.PresetLevel {
	switch scope {
	case types.SetupScopeGlobal:
		return types.PresetLevelMinimal
	case types.SetupScopeProject, types.SetupScopeWorkspace:
		return types.PresetLevelStandard
	default:
		return types.PresetLevelStandard
	}
}

// ResolvePreset returns the FeatureFlags for a given PresetLevel.
// Returns nil for PresetLevelCustom, signalling that the caller must supply
// their own feature flags.
func ResolvePreset(preset types.PresetLevel) *types.FeatureFlags {
	if preset == types.PresetLevelCustom {
		return nil
	}
	flags, ok := PresetFeatures[preset]
	if !ok {
		return nil
	}
	// Return a copy so callers cannot mutate the map value.
	cp := flags
	return &cp
}

// SpecsDirsForPreset returns the specs directories to create for a given
// preset level.
func SpecsDirsForPreset(preset types.PresetLevel) []string {
	switch preset {
	case types.PresetLevelMinimal:
		return []string{"standards", "memory"}
	case types.PresetLevelStandard:
		return []string{"features", "bugfixes", "rules", "adrs", "standards", "templates", "memory"}
	case types.PresetLevelFull, types.PresetLevelCustom:
		return []string{
			"features", "bugfixes", "refactors", "tech-debt",
			"adrs", "memory", "prompts", "standards", "templates", "rules",
		}
	default:
		return []string{}
	}
}

// TemplatesForPreset returns the templates to install for a given preset.
func TemplatesForPreset(preset types.PresetLevel) []types.TemplateId {
	switch preset {
	case types.PresetLevelMinimal:
		return []types.TemplateId{}
	case types.PresetLevelStandard:
		return []types.TemplateId{
			types.TemplateIdPlanTemplate,
			types.TemplateIdSpecTemplate,
			types.TemplateIdTask,
			types.TemplateIdAdr,
			types.TemplateIdBugfixRcaTemplate,
			types.TemplateIdStandard,
			types.TemplateIdChecklistTemplate,
		}
	case types.PresetLevelFull, types.PresetLevelCustom:
		return []types.TemplateId{
			types.TemplateIdPlanTemplate,
			types.TemplateIdSpecTemplate,
			types.TemplateIdTask,
			types.TemplateIdAdr,
			types.TemplateIdBugfixRcaTemplate,
			types.TemplateIdStandard,
			types.TemplateIdChecklistTemplate,
			types.TemplateIdCodeReviewTemplate,
			types.TemplateIdPostmortemTemplate,
			types.TemplateIdTechDebtTemplate,
		}
	default:
		return []types.TemplateId{}
	}
}

// RulesForPreset returns the rules to install for a given preset.
func RulesForPreset(preset types.PresetLevel) []types.RuleId {
	switch preset {
	case types.PresetLevelMinimal:
		return []types.RuleId{}
	case types.PresetLevelStandard:
		return []types.RuleId{
			types.RuleIdCodeStyle,
			types.RuleIdTesting,
			types.RuleIdSecurity,
			types.RuleIdWorkflow,
			types.RuleIdAccess,
		}
	case types.PresetLevelFull, types.PresetLevelCustom:
		return []types.RuleId{
			types.RuleIdAccess,
			types.RuleIdAgentSecurity,
			types.RuleIdCodeStyle,
			types.RuleIdCost,
			types.RuleIdReview,
			types.RuleIdSecurity,
			types.RuleIdTesting,
			types.RuleIdToolUse,
			types.RuleIdWorkflow,
		}
	default:
		return []types.RuleId{}
	}
}
