import * as p from '@clack/prompts'
import type { FeatureFlags, GitConventions } from '../store/schema.js'

export interface Phase2Result {
  planningDir: string
  features: FeatureFlags
  gitConventions: GitConventions
}

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
  types: ['feat', 'fix', 'docs', 'style', 'refactor', 'perf', 'test', 'build', 'ci', 'chore', 'revert'],
  requireTicket: false,
  ticketPattern: '[A-Z]+-[0-9]+',
}

const BRANCH_PATTERN_OPTIONS = [
  { value: '{type}/{ticket}-{description}', label: 'Conventional', hint: 'feat/PROJ-123-add-login' },
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

/**
 * Run Phase 2 of the interactive wizard: gather planningDir, feature flags, and git conventions.
 * These control which XML-tagged prompt engineering blocks are embedded and how git operations are formatted.
 */
export async function runPhase2Features(opts: {
  interactive: boolean
  prior?: {
    planningDir?: string
    features?: Partial<FeatureFlags>
    gitConventions?: Partial<GitConventions>
  }
  cliOverrides?: {
    planningDir?: string
    features?: string[]
    disableFeatures?: string[]
    branchPattern?: string
    commitPattern?: string
  }
}): Promise<Phase2Result> {
  // Non-interactive mode: use cliOverrides or defaults
  if (!opts.interactive) {
    const planningDir = opts.cliOverrides?.planningDir ?? opts.prior?.planningDir ?? '.planning'
    const features = { ...DEFAULT_FEATURES }

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
      for (const flag of opts.cliOverrides.disableFeatures) {
        if (flag in features) {
          features[flag as keyof FeatureFlags] = false
        }
      }
    }

    // Apply prior feature overrides
    if (opts.prior?.features) {
      Object.assign(features, opts.prior.features)
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

    return { planningDir, features, gitConventions }
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
    p.cancel('Setup cancelled.')
    process.exit(0)
  }

  const planningDir = planningDirResult

  // Prompt 2: Feature flags (multiselect with all defaults ON)
  const featureOptions = [
    { value: 'contextEngineering', label: 'Context Engineering', hint: 'Optimal context selection principles' },
    { value: 'rpiWorkflow', label: 'RPI Workflow', hint: 'Research → Plan → Implement phases' },
    { value: 'chainOfThought', label: 'Chain of Thought', hint: 'Structured reasoning: understand → analyze → synthesize → verify' },
    { value: 'treeOfThoughts', label: 'Tree of Thoughts', hint: 'Parallel exploration of multiple approaches' },
    { value: 'adrEnforcement', label: 'ADR Enforcement', hint: 'Architecture Decision Records for significant changes' },
    { value: 'qualityGates', label: 'Quality Gates', hint: 'Lint, typecheck, test, build checks' },
    { value: 'agentHarness', label: 'Agent Harness', hint: 'Multi-agent coordination (planner, builder, reviewer, scout, documenter)' },
    { value: 'bugResolution', label: 'Bug Resolution', hint: 'Structured debugging: reproduce → diagnose → fix → verify' },
    { value: 'pivotHandling', label: 'Pivot Handling', hint: 'Detection and ADR process when plans change' },
  ]

  // Pre-select: all enabled by default
  const priorFeatures = opts.prior?.features ?? {}
  const initialValues = featureOptions
    .filter(opt => {
      const key = opt.value as keyof FeatureFlags
      return priorFeatures[key] ?? DEFAULT_FEATURES[key]
    })
    .map(opt => opt.value)

  const featuresResult = await p.multiselect({
    message: 'Which prompt engineering features to enable? (all recommended)',
    options: featureOptions,
    initialValues,
    required: false,
  })

  if (p.isCancel(featuresResult)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }

  const selectedFeatures = featuresResult as string[]

  // Build FeatureFlags from selection
  const features: FeatureFlags = {
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

  // Prompt 3: Branch naming pattern
  const priorBranch = opts.prior?.gitConventions?.branchPattern ?? DEFAULT_GIT_CONVENTIONS.branchPattern
  const branchPatternResult = await p.select({
    message: 'Branch naming pattern?',
    options: BRANCH_PATTERN_OPTIONS,
    initialValue: BRANCH_PATTERN_OPTIONS.find(o => o.value === priorBranch)?.value ?? BRANCH_PATTERN_OPTIONS[0]?.value ?? DEFAULT_GIT_CONVENTIONS.branchPattern,
  })

  if (p.isCancel(branchPatternResult)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }

  let branchPattern = branchPatternResult as string
  if (branchPattern === 'custom') {
    const customBranch = await p.text({
      message: 'Custom branch pattern (use {type}, {ticket}, {description}):',
      placeholder: '{type}/{ticket}-{description}',
      defaultValue: '{type}/{ticket}-{description}',
    })
    if (p.isCancel(customBranch)) {
      p.cancel('Setup cancelled.')
      process.exit(0)
    }
    branchPattern = customBranch
  }

  // Prompt 4: Commit message pattern
  const priorCommit = opts.prior?.gitConventions?.commitPattern ?? DEFAULT_GIT_CONVENTIONS.commitPattern
  const commitPatternResult = await p.select({
    message: 'Commit message pattern?',
    options: COMMIT_PATTERN_OPTIONS,
    initialValue: COMMIT_PATTERN_OPTIONS.find(o => o.value === priorCommit)?.value ?? COMMIT_PATTERN_OPTIONS[0]?.value ?? DEFAULT_GIT_CONVENTIONS.commitPattern,
  })

  if (p.isCancel(commitPatternResult)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }

  let commitPattern = commitPatternResult as string
  if (commitPattern === 'custom') {
    const customCommit = await p.text({
      message: 'Custom commit pattern (use {type}, {scope}, {ticket}, {description}):',
      placeholder: '{type}({scope}): {description}',
      defaultValue: '{type}({scope}): {description}',
    })
    if (p.isCancel(customCommit)) {
      p.cancel('Setup cancelled.')
      process.exit(0)
    }
    commitPattern = customCommit
  }

  // Prompt 5: Require ticket in branch/commit?
  const requireTicketResult = await p.confirm({
    message: 'Require ticket ID in branches/commits?',
    initialValue: opts.prior?.gitConventions?.requireTicket ?? false,
  })

  if (p.isCancel(requireTicketResult)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }

  const gitConventions: GitConventions = {
    branchPattern,
    commitPattern,
    types: DEFAULT_GIT_CONVENTIONS.types,
    requireTicket: requireTicketResult,
    ticketPattern: DEFAULT_GIT_CONVENTIONS.ticketPattern,
  }

  return { planningDir, features, gitConventions }
}
