import type { FeatureFlags } from './store/schema.js'
import type { RuleId, SetupScope, TemplateId } from './types.js'

export type PresetLevel = 'minimal' | 'standard' | 'full' | 'custom'

export const PRESET_FEATURES: Record<Exclude<PresetLevel, 'custom'>, FeatureFlags> = {
  minimal: {
    contextEngineering: false,
    rpiWorkflow: false,
    chainOfThought: false,
    treeOfThoughts: false,
    adrEnforcement: false,
    qualityGates: true,
    agentHarness: false,
    bugResolution: false,
    pivotHandling: false,
  },
  standard: {
    contextEngineering: false,
    rpiWorkflow: true,
    chainOfThought: true,
    treeOfThoughts: false,
    adrEnforcement: false,
    qualityGates: true,
    agentHarness: false,
    bugResolution: true,
    pivotHandling: false,
  },
  full: {
    contextEngineering: true,
    rpiWorkflow: true,
    chainOfThought: true,
    treeOfThoughts: true,
    adrEnforcement: true,
    qualityGates: true,
    agentHarness: true,
    bugResolution: true,
    pivotHandling: true,
  },
}

/** Returns the default preset level for a given setup scope */
export function defaultPresetForScope(scope: SetupScope): PresetLevel {
  switch (scope) {
    case 'global':
      return 'minimal'
    case 'project':
    case 'workspace':
      return 'standard'
    default:
      return 'standard'
  }
}

/** Resolve a preset name to feature flags. Returns undefined for 'custom'. */
export function resolvePreset(preset: PresetLevel): FeatureFlags | undefined {
  if (preset === 'custom') return undefined
  return { ...PRESET_FEATURES[preset] }
}

/** Returns the specs directories to create for a given preset level */
export function specsDirsForPreset(preset: PresetLevel): string[] {
  switch (preset) {
    case 'minimal':
      return ['standards', 'memory']
    case 'standard':
      return ['features', 'bugfixes', 'rules', 'adrs', 'standards', 'templates', 'memory']
    case 'full':
    case 'custom':
      return ['features', 'bugfixes', 'refactors', 'tech-debt', 'adrs', 'memory', 'prompts', 'standards', 'templates', 'rules']
  }
}

/** Returns the templates to install for a given preset */
export function templatesForPreset(preset: PresetLevel): TemplateId[] {
  switch (preset) {
    case 'minimal':
      return []
    case 'standard':
      return ['plan-template', 'spec-template', 'task', 'adr', 'bugfix-rca-template', 'standard', 'checklist-template']
    case 'full':
    case 'custom':
      return [
        'plan-template',
        'spec-template',
        'task',
        'adr',
        'bugfix-rca-template',
        'standard',
        'checklist-template',
        'code-review-template',
        'postmortem-template',
        'tech-debt-template',
      ]
  }
}

/** Returns the rules to install for a given preset */
export function rulesForPreset(preset: PresetLevel): RuleId[] {
  switch (preset) {
    case 'minimal':
      return []
    case 'standard':
      return ['code-style', 'testing', 'security', 'workflow', 'access']
    case 'full':
    case 'custom':
      return ['access', 'agent-security', 'code-style', 'cost', 'review', 'security', 'testing', 'tool-use', 'workflow']
  }
}
