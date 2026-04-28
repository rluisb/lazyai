import { createRequire } from 'node:module'
import { z } from 'zod'

const require = createRequire(import.meta.url)
const pkg = (() => {
  try {
    return require('../package.json') as { version: string }
  } catch {
    return require('../../package.json') as { version: string }
  }
})()

export const CURRENT_SCHEMA_VERSION = 1

export const setupScopeSchema = z.enum(['global', 'workspace', 'project'])
export const setupTypeSchema = z.enum(['project', 'workspace'])
export const presetLevelSchema = z.enum(['minimal', 'standard', 'full', 'custom'])
export const toolIdSchema = z.enum(['opencode', 'claude-code', 'gemini', 'copilot', 'codex', 'pi'])
export const agentIdSchema = z.enum(['builder', 'documenter', 'implementor', 'orchestrator', 'planner', 'red-team', 'reviewer', 'scout'])
export const skillIdSchema = z.enum([
  'anti-speculation',
  'bugfix',
  'extract-standards',
  'housekeeping',
  'impact-check',
  'implement',
  'iterate',
  'memory-write',
  'orchestrate',
  'parallel-execution',
  'plan',
  'process-audit',
  'proof-of-concept',
  'research',
  'review',
  'rpi',
  'self-improve',
  'speckit-analyze',
  'speckit-checklist',
  'speckit-clarify',
  'speckit-constitution',
  'speckit-implement',
  'speckit-plan',
  'speckit-specify',
  'speckit-tasks',
  'spike',
  'tdd-loop',
  'update-memory',
])
export const promptIdSchema = z.enum(['compact', 'implement', 'local-example', 'plan', 'research'])
export const templateIdSchema = z.enum([
  'adr',
  'bugfix-rca-template',
  'checklist-template',
  'code-review-template',
  'plan-template',
  'postmortem-template',
  'spec-template',
  'standard',
  'task',
  'tech-debt-template',
])
export const ruleIdSchema = z.enum(['access', 'agent-security', 'code-style', 'cost', 'review', 'security', 'testing', 'tool-use', 'workflow'])
export const infraIdSchema = z.enum(['pre-commit', 'compliance', 'KNOWLEDGE_MAP', 'codeowners'])

export const metaSchema = z.object({
  schemaVersion: z.number(),
  cliVersion: z.string(),
  installedAt: z.string(),
  lastUpdatedAt: z.string(),
})

// Feature flags for compilation - ALL ON by default
export const featureFlagsSchema = z.object({
  contextEngineering: z.boolean().default(true),
  rpiWorkflow: z.boolean().default(true),
  chainOfThought: z.boolean().default(true),
  treeOfThoughts: z.boolean().default(true),
  adrEnforcement: z.boolean().default(true),
  qualityGates: z.boolean().default(true),
  agentHarness: z.boolean().default(true),
  bugResolution: z.boolean().default(true),
  pivotHandling: z.boolean().default(true),
})

// Git conventions for branch/commit patterns
export const gitConventionsSchema = z.object({
  branchPattern: z.string().default('{type}/{ticket}-{description}'),
  commitPattern: z.string().default('{type}({scope}): {description}'),
  types: z.array(z.string()).default(['feat', 'fix', 'docs', 'style', 'refactor', 'perf', 'test', 'build', 'ci', 'chore', 'revert']),
  requireTicket: z.boolean().default(false),
  ticketPattern: z.string().default('[A-Z]+-[0-9]+'),
})

export const configSchema = z.object({
  setupScope: setupScopeSchema,
  setupType: setupTypeSchema.optional(),
  tools: toolIdSchema.array(),
  cliTools: z.array(z.string()).optional(),
  enableServers: z.array(z.string()).optional(),
  projectName: z.string(),
  workspaceName: z.string().optional(),
  targetDir: z.string(),
  planningDir: z.string().optional(),
  planningRepoPath: z.string().optional(),
  repos: z
    .array(
      z.object({
        name: z.string(),
        path: z.string(),
        type: z.string().optional(),
        description: z.string().optional(),
      }),
    )
    .optional(),
  globalRef: z.string().optional(),
})

export const wizardSelectionsSchema = z.object({
  templates: templateIdSchema.array(),
  rules: ruleIdSchema.array(),
  agents: agentIdSchema.array(),
  skills: skillIdSchema.array(),
  prompts: promptIdSchema.array(),
  infra: infraIdSchema.array(),
  constitution: z.array(z.string()),
  features: featureFlagsSchema.optional(),
  gitConventions: gitConventionsSchema.optional(),
})

export const trackedFileSchema = z.object({
  path: z.string(),
  hash: z.string(),
  source: z.string(),
  owner: z.enum(['library', 'user', 'migrated']).default('library'),
  status: z.enum(['installed', 'modified', 'missing', 'conflict']).optional(),
  installedAt: z.string().optional(),
  lastCheckedAt: z.string().optional(),
  kind: z.enum(['file', 'symlink']).optional(),
  linkTarget: z.string().optional(),
})

export const syncSchema = z.object({
  lastSyncAt: z.string(),
  dirty: z.boolean(),
})

export const operationSchema = z.object({
  id: z.string(),
  type: z.string(),
  timestamp: z.string(),
  filesAffected: z.array(z.string()),
  result: z.enum(['success', 'failure', 'partial']),
  backupPaths: z.array(z.string()).optional(),
  error: z.string().optional(),
})

export const storeDataSchema = z.object({
  meta: metaSchema,
  config: configSchema,
  selections: wizardSelectionsSchema,
  files: trackedFileSchema.array(),
  sync: syncSchema,
  operations: operationSchema.array(),
})

export type StoreData = z.infer<typeof storeDataSchema>
export type TrackedFile = z.infer<typeof trackedFileSchema>
export type Operation = z.infer<typeof operationSchema>
export type WizardSelections = z.infer<typeof wizardSelectionsSchema>
export type Meta = z.infer<typeof metaSchema>
export type Config = z.infer<typeof configSchema>
export type Sync = z.infer<typeof syncSchema>
export type FeatureFlags = z.infer<typeof featureFlagsSchema>
export type GitConventions = z.infer<typeof gitConventionsSchema>

export function defaultStore(): StoreData {
  const now = new Date().toISOString()

  return {
    meta: {
      schemaVersion: CURRENT_SCHEMA_VERSION,
      cliVersion: pkg.version,
      installedAt: now,
      lastUpdatedAt: now,
    },
    config: {
      setupScope: 'project',
      tools: [],
      projectName: '',
      targetDir: '',
    },
    selections: {
      templates: [],
      rules: [],
      agents: [],
      skills: [],
      prompts: [],
      infra: [],
      constitution: [],
    },
    files: [],
    sync: {
      lastSyncAt: now,
      dirty: true,
    },
    operations: [],
  }
}
