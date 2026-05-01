import os from 'node:os'
import path from 'node:path'
import type { BaseAgentDefinition, ChainDefinition, OrchestrationCatalog, SkillDefinition, TeamDefinition, WorkflowDefinition } from '../types.js'
import type { Db } from '../db/index.js'
import { CatalogStore } from './store.js'
import { mergeScannedCatalogs, scanClaudeGlobal, scanClaudeProject, scanOpencodeGlobal } from './host-scanner.js'
import { chainBodySchema, teamBodySchema, workflowBodySchema } from './schemas.js'

const DEFAULT_ALLOWED_TOOLS = ['Read', 'Grep', 'Glob', 'Edit', 'Write', 'Bash']

function dbRowToAgent(row: { name: string; frontmatterJson: string; body: string }): BaseAgentDefinition {
  const fm = JSON.parse(row.frontmatterJson) as Record<string, unknown>
  const modelHint = typeof fm.model === 'string' ? fm.model : undefined
  return {
    kind: 'agent',
    name: row.name,
    displayName: (typeof fm.name === 'string' ? fm.name : row.name),
    description: typeof fm.description === 'string' ? fm.description : '',
    source: 'db',
    path: '',
    prompt: row.body,
    allowedTools: DEFAULT_ALLOWED_TOOLS,
    constraints: [],
    ...(modelHint ? { modelHint } : {}),
  }
}

function dbRowToSkill(
  row: { name: string; frontmatterJson: string; body: string },
  kind: 'domain' | 'mode' = 'domain',
): SkillDefinition {
  const fm = JSON.parse(row.frontmatterJson) as Record<string, unknown>
  const allowedTools = Array.isArray(fm.allowed_tools)
    ? (fm.allowed_tools as unknown[]).filter((v): v is string => typeof v === 'string')
    : undefined
  const modelHint = typeof fm.model_hint === 'string' ? fm.model_hint : undefined
  const rawPolicy = fm.approval_policy
  const approvalPolicy = rawPolicy === 'minimal' || rawPolicy === 'normal' || rawPolicy === 'strict' ? rawPolicy : undefined
  return {
    kind,
    name: row.name,
    description: typeof fm.description === 'string' ? fm.description : '',
    source: 'db',
    path: '',
    prompt: row.body,
    constraints: [],
    ...(allowedTools ? { allowedTools } : {}),
    ...(modelHint ? { modelHint } : {}),
    ...(approvalPolicy ? { approvalPolicy } : {}),
  }
}

type DbBodyRow = { name: string; version: number; body: string }

function parseJsonBody(body: string): unknown | null {
  try {
    return JSON.parse(body) as unknown
  } catch {
    return null
  }
}

function dbRowToChain(row: DbBodyRow): ChainDefinition | null {
  const parsed = chainBodySchema.safeParse(parseJsonBody(row.body))
  if (!parsed.success) return null
  const data = parsed.data
  return {
    kind: 'chain',
    name: row.name,
    description: data.description ?? '',
    version: String(row.version),
    source: 'db',
    path: `catalog://chain/${row.name}`,
    entry: data.entry,
    steps: data.steps.map((step) => ({
      id: step.id,
      agent: step.agent,
      skills: step.skills,
      description: step.description,
      transitions: step.transitions,
      ...(step.gate ? { gate: step.gate } : {}),
      ...(step.prompt ? { prompt: step.prompt } : {}),
      ...(step.allowedTools ? { allowedTools: step.allowedTools } : {}),
      ...(step.model ? { model: step.model } : {}),
    })),
    ...(data.domain_skill_injection ? { domain_skill_injection: data.domain_skill_injection } : {}),
    ...(data.mode_skill_injection ? { mode_skill_injection: data.mode_skill_injection } : {}),
  }
}

function dbRowToTeam(row: DbBodyRow): TeamDefinition | null {
  const parsed = teamBodySchema.safeParse(parseJsonBody(row.body))
  if (!parsed.success) return null
  const data = parsed.data
  return {
    kind: 'team',
    name: row.name,
    description: data.description ?? '',
    version: String(row.version),
    source: 'db',
    path: `catalog://team/${row.name}`,
    parallel: data.parallel,
    synthesize: data.synthesize,
    ...(data.budget_multiplier !== undefined ? { budget_multiplier: data.budget_multiplier } : {}),
    ...(data.user_confirmation_required !== undefined ? { user_confirmation_required: data.user_confirmation_required } : {}),
  }
}

function dbRowToWorkflow(row: DbBodyRow): WorkflowDefinition | null {
  const parsed = workflowBodySchema.safeParse(parseJsonBody(row.body))
  if (!parsed.success) return null
  const data = parsed.data
  return {
    kind: 'workflow',
    name: row.name,
    description: data.description ?? '',
    version: String(row.version),
    source: 'db',
    path: `catalog://workflow/${row.name}`,
    entry: data.entry,
    phases: data.phases.map((phase) => ({
      id: phase.id,
      kind: phase.kind,
      ...(phase.ref ? { ref: phase.ref } : {}),
      ...(phase.gate ? { gate: phase.gate } : {}),
      ...(phase.prompt ? { prompt: phase.prompt } : {}),
      ...(phase.when ? { when: phase.when } : {}),
      ...(phase.on ? { on: phase.on } : {}),
    })),
  }
}

export type ResolverHostCli = 'opencode' | 'claude-code' | 'copilot'

export interface ResolverOptions {
  db: Db
  projectRoot: string
  hostCli?: ResolverHostCli
}

export function resolveCatalog(base: OrchestrationCatalog, opts: ResolverOptions): OrchestrationCatalog {
  const store = new CatalogStore(opts.db)

  // Priority (lowest → highest):
  // 1. base (file library + file project — already merged by loadCatalog)
  // 2. DB internal active versions
  // 3. User-global host files
  // 4. User-project host files

  const agents = { ...base.agents }
  const domains = { ...base.domains }
  const modes = { ...base.modes }
  const chains = { ...base.chains }
  const teams = { ...base.teams }
  const workflows = { ...base.workflows }

  // Layer 2: DB internal
  const dbAgentDefs = store.listDefinitions('agent')
  for (const def of dbAgentDefs) {
    const row = store.getActiveVersion('agent', def.name)
    if (!row) continue
    agents[def.name] = dbRowToAgent({ name: def.name, frontmatterJson: row.frontmatterJson, body: row.body })
  }
  const dbSkillDefs = store.listDefinitions('skill')
  for (const def of dbSkillDefs) {
    const row = store.getActiveVersion('skill', def.name)
    if (!row) continue
    const fm = JSON.parse(row.frontmatterJson) as Record<string, unknown>
    const skillKind = fm.kind === 'mode' ? 'mode' : 'domain'
    const skill = dbRowToSkill({ name: def.name, frontmatterJson: row.frontmatterJson, body: row.body }, skillKind)
    if (skillKind === 'mode') {
      modes[def.name] = skill
    } else {
      domains[def.name] = skill
    }
  }
  for (const def of store.listDefinitions('chain')) {
    const row = store.getActiveVersion('chain', def.name)
    if (!row) continue
    const chain = dbRowToChain({ name: def.name, version: row.version, body: row.body })
    if (chain) chains[def.name] = chain
  }
  for (const def of store.listDefinitions('team')) {
    const row = store.getActiveVersion('team', def.name)
    if (!row) continue
    const team = dbRowToTeam({ name: def.name, version: row.version, body: row.body })
    if (team) teams[def.name] = team
  }
  for (const def of store.listDefinitions('workflow')) {
    const row = store.getActiveVersion('workflow', def.name)
    if (!row) continue
    const workflow = dbRowToWorkflow({ name: def.name, version: row.version, body: row.body })
    if (workflow) workflows[def.name] = workflow
  }

  // Layers 3 + 4: Host scans (only when caller declares which CLI is in use)
  if (opts.hostCli) {
    const scans = resolveHostScans(opts.projectRoot, opts.hostCli)
    Object.assign(agents, scans.agents)
    Object.assign(domains, scans.skills)
  }

  return { ...base, agents, domains, modes, chains, teams, workflows }
}

function resolveHostScans(
  projectRoot: string,
  hostCli?: ResolverHostCli,
): ReturnType<typeof mergeScannedCatalogs> {
  // Scan the union of all relevant host dirs; project-level wins last (highest priority)
  const catalogs = []

  // opencode global
  if (hostCli === 'opencode' || hostCli === 'claude-code' || !hostCli) {
    try { catalogs.push(scanOpencodeGlobal()) } catch { /* skip if path doesn't exist */ }
  }

  // claude-code global
  if (hostCli === 'claude-code' || !hostCli) {
    try { catalogs.push(scanClaudeGlobal()) } catch { /* skip */ }
  }

  // claude-code project (highest priority)
  if (hostCli === 'claude-code' || !hostCli) {
    const isUnderHome = projectRoot.startsWith(os.homedir())
    const isValidProject = isUnderHome || path.isAbsolute(projectRoot)
    if (isValidProject) {
      try { catalogs.push(scanClaudeProject(projectRoot)) } catch { /* skip */ }
    }
  }

  return mergeScannedCatalogs(...catalogs)
}
