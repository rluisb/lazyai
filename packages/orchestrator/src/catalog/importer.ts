import fs from 'node:fs'
import path from 'node:path'
import type { Db } from '../db/index.js'
import type { BaseAgentDefinition, SkillDefinition, ChainDefinition, TeamDefinition, WorkflowDefinition } from '../types.js'
import { parseMarkdownFrontmatter } from '../loader.js'
import { CatalogStore } from './store.js'
import { scanClaudeGlobal, scanClaudeProject, scanOpencodeGlobal } from './host-scanner.js'

export interface CatalogImportResult {
  imported: number
  skipped: number
  errors: Array<{ source: string; error: string }>
}

function importAgent(store: CatalogStore, agent: BaseAgentDefinition, createdBy: string): void {
  const frontmatter: Record<string, unknown> = { name: agent.displayName, description: agent.description }
  if (agent.modelHint) frontmatter.model = agent.modelHint
  store.createVersion({
    kind: 'agent',
    name: agent.name,
    frontmatter,
    body: agent.prompt,
    createdBy,
  })
}

function importSkill(store: CatalogStore, skill: SkillDefinition, createdBy: string): void {
  const frontmatter: Record<string, unknown> = {
    name: skill.name,
    description: skill.description,
    kind: skill.kind,
  }
  if (skill.allowedTools) frontmatter.allowed_tools = skill.allowedTools
  if (skill.modelHint) frontmatter.model_hint = skill.modelHint
  if (skill.approvalPolicy) frontmatter.approval_policy = skill.approvalPolicy
  store.createVersion({
    kind: 'skill',
    name: skill.name,
    frontmatter,
    body: skill.prompt,
    createdBy,
  })
}

function importJsonDefinitions<T extends ChainDefinition | TeamDefinition | WorkflowDefinition>(
  store: CatalogStore,
  dir: string,
  kind: 'chain' | 'team' | 'workflow',
  result: CatalogImportResult,
  createdBy: string,
): void {
  if (!fs.existsSync(dir)) return
  for (const entry of fs.readdirSync(dir)) {
    if (!entry.endsWith('.json')) continue
    const filePath = path.join(dir, entry)
    try {
      const raw = fs.readFileSync(filePath, 'utf-8')
      const def = JSON.parse(raw) as T
      const name = (typeof def.name === 'string' ? def.name : path.basename(entry, '.json')).toLowerCase()
      const versionResult = store.createVersion({
        kind,
        name,
        frontmatter: {},
        body: raw,
        createdBy,
      })
      if (versionResult.alreadyExists) result.skipped++
      else result.imported++
    } catch (err) {
      result.errors.push({ source: filePath, error: err instanceof Error ? err.message : String(err) })
    }
  }
}

export function importFromLibrary(db: Db, libraryOrchestrationRoot: string, libraryAgentsRoot?: string): CatalogImportResult {
  const store = new CatalogStore(db)
  const result: CatalogImportResult = { imported: 0, skipped: 0, errors: [] }
  const createdBy = 'library-import'

  // Agents from library
  if (libraryAgentsRoot && fs.existsSync(libraryAgentsRoot)) {
    for (const entry of fs.readdirSync(libraryAgentsRoot)) {
      if (!entry.endsWith('.md')) continue
      const filePath = path.join(libraryAgentsRoot, entry)
      try {
        const raw = fs.readFileSync(filePath, 'utf-8')
        const parsed = parseMarkdownFrontmatter(raw)
        const name = (
          typeof parsed.attributes.name === 'string'
            ? parsed.attributes.name
            : path.basename(entry, '.md')
        ).toLowerCase().replace(/\s+/g, '-')
        const frontmatter: Record<string, unknown> = { ...parsed.attributes }
        const vResult = store.createVersion({ kind: 'agent', name, frontmatter, body: parsed.body.trim(), createdBy })
        if (vResult.alreadyExists) result.skipped++
        else result.imported++
      } catch (err) {
        result.errors.push({ source: filePath, error: err instanceof Error ? err.message : String(err) })
      }
    }
  }

  // Chains / teams / workflows
  importJsonDefinitions(store, path.join(libraryOrchestrationRoot, 'chains'), 'chain', result, createdBy)
  importJsonDefinitions(store, path.join(libraryOrchestrationRoot, 'teams'), 'team', result, createdBy)
  importJsonDefinitions(store, path.join(libraryOrchestrationRoot, 'workflows'), 'workflow', result, createdBy)

  return result
}

export function importFromHostFiles(
  db: Db,
  hosts: Array<'opencode' | 'claude-code'>,
  projectRoot?: string,
): CatalogImportResult {
  const store = new CatalogStore(db)
  const result: CatalogImportResult = { imported: 0, skipped: 0, errors: [] }

  const process = (agents: Record<string, BaseAgentDefinition>, skills: Record<string, SkillDefinition>, createdBy: string): void => {
    for (const agent of Object.values(agents)) {
      try {
        const r = (() => { importAgent(store, agent, createdBy); return { alreadyExists: false } })()
        if (r.alreadyExists) result.skipped++
        else result.imported++
      } catch (err) {
        result.errors.push({ source: agent.path, error: err instanceof Error ? err.message : String(err) })
      }
    }
    for (const skill of Object.values(skills)) {
      try {
        importSkill(store, skill, createdBy)
        result.imported++
      } catch (err) {
        result.errors.push({ source: skill.path, error: err instanceof Error ? err.message : String(err) })
      }
    }
  }

  for (const host of hosts) {
    if (host === 'opencode') {
      const scan = scanOpencodeGlobal()
      process(scan.agents, scan.skills, 'opencode-global-import')
    }
    if (host === 'claude-code') {
      const globalScan = scanClaudeGlobal()
      process(globalScan.agents, globalScan.skills, 'claude-code-global-import')
      if (projectRoot) {
        const projectScan = scanClaudeProject(projectRoot)
        process(projectScan.agents, projectScan.skills, 'claude-code-project-import')
      }
    }
  }

  return result
}
