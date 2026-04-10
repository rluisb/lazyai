import * as p from '@clack/prompts'
import { Errors } from '../errors/index.js'
import { defaultPresetForScope, PRESET_FEATURES, type PresetLevel, resolvePreset } from '../presets.js'
import type { FeatureFlags, GitConventions } from '../store/schema.js'
import type { SetupScope } from '../types.js'

function cancelWizard(): never {
  p.cancel('Setup cancelled.')
  throw Errors.userCancelled()
}

export interface Phase2Result {
  planningDir: string
  features: FeatureFlags
  gitConventions: GitConventions
  preset: PresetLevel
}

const FEATURE_KEYS: Array<keyof FeatureFlags> = [
  'contextEngineering',
  'rpiWorkflow',
  'chainOfThought',
  'treeOfThoughts',
  'adrEnforcement',
  'qualityGates',
  'agentHarness',
  'bugResolution',
  'pivotHandling',
]

// All features ON by default - users can disable what they don't need
const DEFAULT_FEATURES: FeatureFlags = {
  contextEngineering: true,
  rpiWorkflow: true,
  chainOfThought: true,
  treeOfThoughts: true,
  adrEnforcement: true,
  qualityGates: true,
  agentHarness: true,
  bugResolution: true,
  pivotHandling: true,
}

// Default git conventions - conventional commits style
const DEFAULT_GIT_CONVENTIONS: GitConventions = {
  branchPattern: '{type}/{ticket}-{description}',
  commitPattern: '{type}({scope}): {description}',
  types: ['feat', 'fix', 'docs', 'style', 'refactor', 'perf', 'test', 'build', 'ci', 'chore', 'revert', 'task', 'bug'],
  requireTicket: false,
  ticketPattern: '[A-Z]+-[0-9]+',
}

const BRANCH_PATTERN_OPTIONS = [
  { value: '{type}/{ticket}-{description}', label: 'Conventional', hint: 'feat/PROJ-123-add-login' },
  { value: '{type}/{ticket}/{description}', label: 'Jira Style', hint: 'task/PBG-35/creator-billing-endpoints' },
  { value: '{type}/{description}', label: 'Simple Type', hint: 'feat/add-login' },
  { value: '{ticket}/{description}', label: 'Ticket First', hint: 'PROJ-123/add-login' },
  { value: '{description}', label: 'Description Only', hint: 'add-login' },
  { value: 'custom', label: 'Custom Pattern', hint: 'Define your own pattern' },
]

const COMMIT_PATTERN_OPTIONS = [
  { value: '{type}({scope}): {description}', label: 'Conventional Commits', hint: 'feat(auth): add login' },
  { value: '{type}: {description}', label: 'Simple Type', hint: 'feat: add login' },
  { value: '[{ticket}] {description}', label: 'Ticket Prefix', hint: '[PROJ-123] add login' },
  { value: '{description}', label: 'Description Only', hint: 'add login' },
  { value: 'custom', label: 'Custom Pattern', hint: 'Define your own pattern' },
]

function buildFeaturesFromSelection(selectedFeatures: string[]): FeatureFlags {
  return {
    contextEngineering: selectedFeatures.includes('contextEngineering'),
    rpiWorkflow: selectedFeatures.includes('rpiWorkflow'),
    chainOfThought: selectedFeatures.includes('chainOfThought'),
    treeOfThoughts: selectedFeatures.includes('treeOfThoughts'),
    adrEnforcement: selectedFeatures.includes('adrEnforcement'),
    qualityGates: selectedFeatures.includes('qualityGates'),
    agentHarness: selectedFeatures.includes('agentHarness'),
    bugResolution: selectedFeatures.includes('bugResolution'),
    pivotHandling: selectedFeatures.includes('pivotHandling'),
  }
}

function inferPresetFromFeatures(features?: Partial<FeatureFlags>): PresetLevel | undefined {
  if (!features || !FEATURE_KEYS.every((key) => key in features)) {
    return undefined
  }

  const fullFeatures = features as FeatureFlags

  for (const [preset, presetFeatures] of Object.entries(PRESET_FEATURES) as Array<[
    Exclude<PresetLevel, 'custom'>,
    FeatureFlags,
  ]>) {
    if (FEATURE_KEYS.every((key) => fullFeatures[key] === presetFeatures[key])) {
      return preset
    }
  }

  return 'custom'
}

/**
 * Run Phase 2 of the interactive wizard: gather planningDir, feature flags, and git conventions.
 * These control which XML-tagged prompt engineering blocks are embedded and how git operations are formatted.
 */
export async function runPhase2Features(opts: {
  interactive: boolean
  setupScope?: SetupScope
  prior?: {
    planningDir?: string
    features?: Partial<FeatureFlags>
    gitConventions?: Partial<GitConventions>
  }
  cliOverrides?: {
    planningDir?: string
    preset?: PresetLevel
    features?: string[]
    disableFeatures?: string[]
    branchPattern?: string
    commitPattern?: string
  }
}): Promise<Phase2Result> {
  // Non-interactive mode: use cliOverrides or defaults
  if (!opts.interactive) {
    const planningDir = opts.cliOverrides?.planningDir ?? opts.prior?.planningDir ?? '.planning'

    let features: FeatureFlags
    let preset: PresetLevel
    if (opts.cliOverrides?.preset) {
      preset = opts.cliOverrides.preset
      const resolved = resolvePreset(opts.cliOverrides.preset)
      features = resolved ?? { ...DEFAULT_FEATURES }
    } else {
      features = { ...DEFAULT_FEATURES }
      // Preset will be inferred after feature overrides are applied
      preset = 'full' // placeholder — recalculated below
    }

    // Apply prior feature overrides (baseline from last run)
    if (opts.prior?.features) {
      Object.assign(features, opts.prior.features)
    }

    // Apply CLI feature enables
    if (opts.cliOverrides?.features) {
      for (const flag of opts.cliOverrides.features) {
        if (flag in features) {
          features[flag as keyof FeatureFlags] = true
        }
      }
    }

    // Apply CLI feature disables
    if (opts.cliOverrides?.disableFeatures) {
      const disableFeatures = opts.cliOverrides.disableFeatures
      if (disableFeatures.includes('all')) {
        for (const key of Object.keys(features) as Array<keyof FeatureFlags>) {
          features[key] = false
        }
      } else {
        for (const flag of disableFeatures) {
          if (flag in features) {
            features[flag as keyof FeatureFlags] = false
          }
        }
      }
    }

    // Infer preset from final resolved features (when no explicit --preset)
    if (!opts.cliOverrides?.preset) {
      preset = inferPresetFromFeatures(features) ?? defaultPresetForScope(opts.setupScope ?? 'project')
    }

    // Git conventions
    const gitConventions: GitConventions = {
      ...DEFAULT_GIT_CONVENTIONS,
      ...opts.prior?.gitConventions,
    }
    if (opts.cliOverrides?.branchPattern) {
      gitConventions.branchPattern = opts.cliOverrides.branchPattern
    }
    if (opts.cliOverrides?.commitPattern) {
      gitConventions.commitPattern = opts.cliOverrides.commitPattern
    }

    return { planningDir, features, gitConventions, preset }
  }

  // Interactive mode
  p.note('Configure prompt engineering features and git conventions for your AI tools.')

  // Prompt 1: Planning directory
  const defaultDir = opts.prior?.planningDir ?? '.planning'
  const planningDirResult = await p.text({
    message: 'Planning directory for specs, ADRs, and research?',
    placeholder: defaultDir,
    defaultValue: defaultDir,
    validate: (value) => {
      if (!value) return 'Planning directory is required'
      if (value.includes('..')) return 'Path cannot contain ..'
      return undefined
    },
  })

  if (p.isCancel(planningDirResult)) {
    cancelWizard()
  }

  const planningDir = planningDirResult

  // Prompt 2: Preset selection
  const presetOptions: Array<{ value: PresetLevel; label: string; hint: string }> = [
    {
      value: 'minimal',
      label: 'Minimal',
      hint: 'Quality gates + git conventions only. Best for cheap models or small context.',
    },
    {
      value: 'standard',
      label: 'Standard (recommended)',
      hint: '+ Reasoning protocol, RPI workflow, bug resolution. Best for most teams.',
    },
    {
      value: 'full',
      label: 'Full',
      hint: '+ Context discipline, decision protocol, agent coordination, ADR enforcement.',
    },
    {
      value: 'custom',
      label: 'Custom',
      hint: 'Pick features individually.',
    },
  ]

  const presetResult =
    opts.cliOverrides?.preset ??
    (await p.select({
      message: 'Feature preset?',
      options: presetOptions,
      initialValue:
        inferPresetFromFeatures(opts.prior?.features) ?? defaultPresetForScope(opts.setupScope ?? 'project'),
    }))

  if (p.isCancel(presetResult)) {
    cancelWizard()
  }

  const preset = presetResult as PresetLevel

  // Prompt 3: Feature flags (multiselect only for custom preset)
  const featureOptions = [
    { value: 'qualityGates', label: 'Quality Gates', hint: 'Lint, typecheck, test, build checks' },
    { value: 'rpiWorkflow', label: 'RPI Workflow', hint: 'Research → Plan → Implement phases' },
    { value: 'chainOfThought', label: 'Reasoning Protocol', hint: 'Structured <cot> reasoning before acting' },
    { value: 'bugResolution', label: 'Bug Resolution', hint: 'Structured debugging: reproduce → diagnose → fix → verify' },
    { value: 'contextEngineering', label: 'Context Discipline', hint: 'File budget, session hygiene, read priority' },
    { value: 'treeOfThoughts', label: 'Decision Protocol', hint: 'Evaluate multiple approaches before choosing' },
    { value: 'adrEnforcement', label: 'ADR Enforcement', hint: 'Architecture Decision Records for significant changes' },
    { value: 'agentHarness', label: 'Agent Coordination', hint: 'Multi-agent handoff and escalation rules' },
    { value: 'pivotHandling', label: 'Pivot Handling', hint: 'What to do when plans change mid-implementation' },
  ]

  // Pre-select: all enabled by default
  const priorFeatures = opts.prior?.features ?? {}
  const initialValues = featureOptions
    .filter((opt) => {
      const key = opt.value as keyof FeatureFlags
      return priorFeatures[key] ?? DEFAULT_FEATURES[key]
    })
    .map((opt) => opt.value)

  let features: FeatureFlags
  if (preset === 'custom') {
    const featuresResult = await p.multiselect({
      message: 'Which features should be enabled?',
      options: featureOptions,
      initialValues,
      required: false,
    })

    if (p.isCancel(featuresResult)) {
      cancelWizard()
    }

    features = buildFeaturesFromSelection(featuresResult as string[])
  } else {
    features = resolvePreset(preset) ?? { ...DEFAULT_FEATURES }
  }

  // Prompt 4: Branch naming pattern
  const priorBranch = opts.prior?.gitConventions?.branchPattern ?? DEFAULT_GIT_CONVENTIONS.branchPattern
  const branchPatternResult = await p.select({
    message: 'Branch naming pattern?',
    options: BRANCH_PATTERN_OPTIONS,
    initialValue:
      BRANCH_PATTERN_OPTIONS.find((o) => o.value === priorBranch)?.value ??
      BRANCH_PATTERN_OPTIONS[0]?.value ??
      DEFAULT_GIT_CONVENTIONS.branchPattern,
  })

  if (p.isCancel(branchPatternResult)) {
    cancelWizard()
  }

  let branchPattern = branchPatternResult as string
  if (branchPattern === 'custom') {
    const customBranch = await p.text({
      message: 'Custom branch pattern (use {type}, {ticket}, {description}):',
      placeholder: '{type}/{ticket}-{description}',
      defaultValue: '{type}/{ticket}-{description}',
    })
    if (p.isCancel(customBranch)) {
      cancelWizard()
    }
    branchPattern = customBranch
  }

  // Prompt 5: Commit message pattern
  const priorCommit = opts.prior?.gitConventions?.commitPattern ?? DEFAULT_GIT_CONVENTIONS.commitPattern
  const commitPatternResult = await p.select({
    message: 'Commit message pattern?',
    options: COMMIT_PATTERN_OPTIONS,
    initialValue:
      COMMIT_PATTERN_OPTIONS.find((o) => o.value === priorCommit)?.value ??
      COMMIT_PATTERN_OPTIONS[0]?.value ??
      DEFAULT_GIT_CONVENTIONS.commitPattern,
  })

  if (p.isCancel(commitPatternResult)) {
    cancelWizard()
  }

  let commitPattern = commitPatternResult as string
  if (commitPattern === 'custom') {
    const customCommit = await p.text({
      message: 'Custom commit pattern (use {type}, {scope}, {ticket}, {description}):',
      placeholder: '{type}({scope}): {description}',
      defaultValue: '{type}({scope}): {description}',
    })
    if (p.isCancel(customCommit)) {
      cancelWizard()
    }
    commitPattern = customCommit
  }

  // Prompt 6: Require ticket in branch/commit?
  const requireTicketResult = await p.confirm({
    message: 'Require ticket ID in branches/commits?',
    initialValue: opts.prior?.gitConventions?.requireTicket ?? false,
  })

  if (p.isCancel(requireTicketResult)) {
    cancelWizard()
  }

  const gitConventions: GitConventions = {
    branchPattern,
    commitPattern,
    types: DEFAULT_GIT_CONVENTIONS.types,
    requireTicket: requireTicketResult,
    ticketPattern: DEFAULT_GIT_CONVENTIONS.ticketPattern,
  }

  return { planningDir, features, gitConventions, preset }
}
