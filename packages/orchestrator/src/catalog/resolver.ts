import os from 'node:os'
import path from 'node:path'
import type { BaseAgentDefinition, OrchestrationCatalog, SkillDefinition } from '../types.js'
import type { Db } from '../db/index.js'
import { CatalogStore } from './store.js'
import { mergeScannedCatalogs, scanClaudeGlobal, scanClaudeProject, scanOpencodeGlobal } from './host-scanner.js'

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

export type ResolverHostCli = 'opencode' | 'claude-code' | 'codex'

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

  // Layers 3 + 4: Host scans (only when caller declares which CLI is in use)
  if (opts.hostCli) {
    const scans = resolveHostScans(opts.projectRoot, opts.hostCli)
    Object.assign(agents, scans.agents)
    Object.assign(domains, scans.skills)
  }

  return { ...base, agents, domains, modes }
}

function resolveHostScans(
  projectRoot: string,
  hostCli?: string,
): ReturnType<typeof mergeScannedCatalogs> {
  // Scan the union of all relevant host dirs; project-level wins last (highest priority)
  const catalogs = []

  // opencode global (unless running in codex mode)
  if (hostCli !== 'codex') {
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
