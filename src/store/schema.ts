import { z } from 'zod'
import { createRequire } from 'node:module'

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
export const toolIdSchema = z.enum(['pi', 'opencode', 'claude-code', 'gemini', 'copilot'])
export const agentIdSchema = z.enum(['builder', 'documenter', 'planner', 'red-team', 'reviewer', 'scout'])
export const skillIdSchema = z.enum([
  'anti-speculation',
  'implement',
  'iterate',
  'lessons-learned',
  'memory-write',
  'parallel-execution',
  'plan',
  'research',
  'tdd-loop',
])
export const promptIdSchema = z.enum(['compact', 'implement', 'local-example', 'plan', 'research'])
export const templateIdSchema = z.enum([
  'adr',
  'bugfix-rca-template',
  'code-review-template',
  'postmortem-template',
  'prd-template',
  'progress',
  'standard',
  'task',
  'tasks-template',
  'tech-debt-template',
  'techspec-template',
])
export const ruleIdSchema = z.enum(['access', 'agent-security', 'code-style', 'cost', 'review', 'security', 'testing', 'tool-use', 'workflow'])
export const infraIdSchema = z.enum(['pre-commit', 'compliance', 'KNOWLEDGE_MAP', 'codeowners'])

export const metaSchema = z.object({
  schemaVersion: z.number(),
  cliVersion: z.string(),
  installedAt: z.string(),
  lastUpdatedAt: z.string(),
})

export const configSchema = z.object({
  setupScope: setupScopeSchema,
  setupType: setupTypeSchema.optional(),
  tools: toolIdSchema.array(),
  cliTools: z.array(z.string()).optional(),
  projectName: z.string(),
  workspaceName: z.string().optional(),
  targetDir: z.string(),
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
})

export const trackedFileSchema = z.object({
  path: z.string(),
  hash: z.string(),
  source: z.string(),
  status: z.enum(['installed', 'modified', 'missing', 'conflict']).optional(),
  installedAt: z.string().optional(),
  lastCheckedAt: z.string().optional(),
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
    },
    files: [],
    sync: {
      lastSyncAt: now,
      dirty: true,
    },
    operations: [],
  }
}
