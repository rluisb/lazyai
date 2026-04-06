import type { FeatureFlags } from './store/schema.js'
import type { SetupScope } from './types.js'

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
