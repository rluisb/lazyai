import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import type {
  BaseAgentDefinition,
  ChainDefinition,
  DefinitionMetadata,
  DefinitionSource,
  OrchestrationCatalog,
  SkillDefinition,
  TeamDefinition,
  WorkflowDefinition,
} from './types.js'

const DEFAULT_ALLOWED_TOOLS = ['Read', 'Grep', 'Glob', 'Edit', 'Write', 'Bash']

export interface LoaderOptions {
  projectRoot: string
  libraryOrchestrationRoot?: string
  libraryAgentsRoot?: string
}

const SPEC_AGENT_FILE_ALIASES: Record<string, string> = {
  adr: 'adrs.md',
  adrs: 'adrs.md',
  feature: 'features.md',
  features: 'features.md',
  bugfix: 'bugfixes.md',
  bugfixes: 'bugfixes.md',
  refactor: 'refactors.md',
  refactors: 'refactors.md',
  'tech-debt': 'tech-debt.md',
  workflow: 'workflows.md',
  workflows: 'workflows.md',
  memory: 'memory.md',
  prompt: 'prompts.md',
  prompts: 'prompts.md',
  rule: 'rules.md',
  rules: 'rules.md',
  standard: 'standards.md',
  standards: 'standards.md',
  template: 'templates.md',
  templates: 'templates.md',
}

interface ParsedFrontmatter {
  attributes: Record<string, unknown>
  body: string
}

export function getDefaultLibraryRoots(): { orchestrationRoot: string; agentsRoot: string } {
  const currentDir = path.dirname(fileURLToPath(import.meta.url))
  const packageRoot = path.resolve(currentDir, '..')
  const repoRoot = path.resolve(packageRoot, '..')

  return {
    orchestrationRoot: path.join(repoRoot, 'library', 'orchestration'),
    agentsRoot: path.join(repoRoot, 'library', 'agents'),
  }
}

export function loadCatalog(options: LoaderOptions): OrchestrationCatalog {
  const defaults = getDefaultLibraryRoots()
  const libraryOrchestrationRoot = options.libraryOrchestrationRoot ?? defaults.orchestrationRoot
  const libraryAgentsRoot = options.libraryAgentsRoot ?? defaults.agentsRoot
  const projectOrchestrationRoot = path.join(options.projectRoot, '.ai', 'orchestration')

  return {
    agents: mergeMaps(
      loadAgents(libraryAgentsRoot, 'library'),
      loadAgents(path.join(options.projectRoot, '.ai', 'agents'), 'project'),
    ),
    domains: mergeMaps(
      loadSkills(path.join(libraryOrchestrationRoot, 'skills', 'domains'), 'domain', 'library'),
      loadSkills(path.join(projectOrchestrationRoot, 'skills', 'domains'), 'domain', 'project'),
    ),
    modes: mergeMaps(
      loadSkills(path.join(libraryOrchestrationRoot, 'skills', 'modes'), 'mode', 'library'),
      loadSkills(path.join(projectOrchestrationRoot, 'skills', 'modes'), 'mode', 'project'),
    ),
    chains: mergeMaps(
      loadJsonDefinitions<ChainDefinition>(path.join(libraryOrchestrationRoot, 'chains'), 'library'),
      loadJsonDefinitions<ChainDefinition>(path.join(projectOrchestrationRoot, 'chains'), 'project'),
    ),
    teams: mergeMaps(
      loadJsonDefinitions<TeamDefinition>(path.join(libraryOrchestrationRoot, 'teams'), 'library'),
      loadJsonDefinitions<TeamDefinition>(path.join(projectOrchestrationRoot, 'teams'), 'project'),
    ),
    workflows: mergeMaps(
      loadJsonDefinitions<WorkflowDefinition>(path.join(libraryOrchestrationRoot, 'workflows'), 'library'),
      loadJsonDefinitions<WorkflowDefinition>(path.join(projectOrchestrationRoot, 'workflows'), 'project'),
    ),
  }
}

export function resolveSpecAgentContent(taskType: string | undefined, specsAgentsRoot?: string): string {
  if (!taskType) return ''

  const defaults = getDefaultLibraryRoots()
  const root = specsAgentsRoot ?? path.join(path.dirname(defaults.orchestrationRoot), 'specs-agents')
  const normalized = taskType.trim().toLowerCase()
  const fileName = SPEC_AGENT_FILE_ALIASES[normalized] ?? `${normalized}.md`
  const filePath = path.join(root, fileName)

  if (!fs.existsSync(filePath)) return ''
  return fs.readFileSync(filePath, 'utf-8').trim()
}

function mergeMaps<T>(base: Record<string, T>, overrides: Record<string, T>): Record<string, T> {
  return {
    ...base,
    ...overrides,
  }
}

function loadJsonDefinitions<T extends DefinitionMetadata>(dirPath: string, source: DefinitionSource): Record<string, T> {
  const records: Record<string, T> = {}
  if (!fs.existsSync(dirPath)) return records

  for (const fileName of fs.readdirSync(dirPath)) {
    if (!fileName.endsWith('.json')) continue

    const absolutePath = path.join(dirPath, fileName)
    const raw = fs.readFileSync(absolutePath, 'utf-8')
    const parsed = JSON.parse(raw) as Record<string, unknown>
    const name = typeof parsed.name === 'string'
      ? parsed.name
      : path.basename(fileName, '.json')

    const description = typeof parsed.description === 'string'
      ? parsed.description
      : ''

    records[name] = {
      ...parsed,
      name,
      description,
      source,
      path: absolutePath,
    } as T
  }

  return records
}

function loadSkills(
  dirPath: string,
  kind: 'domain' | 'mode',
  source: DefinitionSource,
): Record<string, SkillDefinition> {
  const records: Record<string, SkillDefinition> = {}
  if (!fs.existsSync(dirPath)) return records

  for (const fileName of fs.readdirSync(dirPath)) {
    if (!fileName.endsWith('.md')) continue
    if (fileName === 'AGENTS.md' || fileName === 'AGENT.md' || fileName.startsWith('_')) continue

    const absolutePath = path.join(dirPath, fileName)
    const raw = fs.readFileSync(absolutePath, 'utf-8')
    const parsed = parseMarkdownFrontmatter(raw)
    const key = normalizeName(parsed.attributes.name, fileName)

    const allowedTools = asStringArray(parsed.attributes.allowed_tools)
    const modelHint = asOptionalString(parsed.attributes.model_hint)
    const approvalPolicy = asApprovalPolicy(parsed.attributes.approval_policy)
    const appliesTo = asStringArray(parsed.attributes.applies_to)

    records[key] = {
      kind,
      name: key,
      description: asString(parsed.attributes.description),
      source,
      path: absolutePath,
      prompt: parsed.body.trim(),
      constraints: extractSkillConstraints(parsed.body),
      ...(allowedTools ? { allowedTools } : {}),
      ...(modelHint ? { modelHint } : {}),
      ...(approvalPolicy ? { approvalPolicy } : {}),
      ...(appliesTo ? { appliesTo } : {}),
    }
  }

  return records
}

function loadAgents(dirPath: string, source: DefinitionSource): Record<string, BaseAgentDefinition> {
  const records: Record<string, BaseAgentDefinition> = {}
  if (!fs.existsSync(dirPath)) return records

  for (const fileName of fs.readdirSync(dirPath)) {
    if (!fileName.endsWith('.md')) continue
    if (fileName === 'AGENTS.md' || fileName === 'AGENT.md' || fileName.startsWith('_')) continue

    const absolutePath = path.join(dirPath, fileName)
    const raw = fs.readFileSync(absolutePath, 'utf-8')
    const parsed = parseMarkdownFrontmatter(raw)
    const key = normalizeName(parsed.attributes.name, fileName)

    const modelHint = asOptionalString(parsed.attributes.model)

    records[key] = {
      kind: 'agent',
      name: key,
      displayName: asString(parsed.attributes.name) || key,
      description: extractFirstParagraph(parsed.body),
      source,
      path: absolutePath,
      prompt: parsed.body.trim(),
      allowedTools: DEFAULT_ALLOWED_TOOLS,
      constraints: extractAgentConstraints(parsed.body),
      ...(modelHint ? { modelHint } : {}),
    }
  }

  return records
}

function parseMarkdownFrontmatter(content: string): ParsedFrontmatter {
  if (!content.startsWith('---\n')) {
    return { attributes: {}, body: content }
  }

  const closingIndex = content.indexOf('\n---', 4)
  if (closingIndex === -1) {
    return { attributes: {}, body: content }
  }

  const frontmatter = content.slice(4, closingIndex)
  const body = content.slice(closingIndex + 4).replace(/^\s*\n/, '')

  return {
    attributes: parseYamlLikeFrontmatter(frontmatter),
    body,
  }
}

function parseYamlLikeFrontmatter(frontmatter: string): Record<string, unknown> {
  const result: Record<string, unknown> = {}
  const lines = frontmatter.split(/\r?\n/)

  let index = 0
  while (index < lines.length) {
    const line = lines[index]?.trim() ?? ''
    if (!line || line.startsWith('#')) {
      index += 1
      continue
    }

    const match = line.match(/^([A-Za-z0-9_-]+):\s*(.*)$/)
    if (!match) {
      index += 1
      continue
    }

    const [, rawKey = '', rawValue = ''] = match
    const key = rawKey.trim()
    if (!key) {
      index += 1
      continue
    }
    if (!rawValue) {
      const items: unknown[] = []
      index += 1
      while (index < lines.length) {
        const listLine = lines[index] ?? ''
        const listMatch = listLine.match(/^\s*-\s+(.*)$/)
        if (!listMatch) break
        items.push(parseScalar(listMatch[1] ?? ''))
        index += 1
      }
      result[key] = items
      continue
    }

    result[key] = parseScalar(rawValue)
    index += 1
  }

  return result
}

function parseScalar(value: string): unknown {
  const normalized = value.trim()
  if (normalized === 'true') return true
  if (normalized === 'false') return false
  if (/^-?\d+(?:\.\d+)?$/.test(normalized)) return Number(normalized)
  return normalized.replace(/^['"]|['"]$/g, '')
}

function normalizeName(value: unknown, fileName: string): string {
  const raw = typeof value === 'string' && value.trim() ? value : path.basename(fileName, path.extname(fileName))
  return raw.toLowerCase().replace(/\s+/g, '-')
}

function asString(value: unknown): string {
  return typeof value === 'string' ? value : ''
}

function asOptionalString(value: unknown): string | undefined {
  return typeof value === 'string' && value.trim() ? value : undefined
}

function asStringArray(value: unknown): string[] | undefined {
  if (!Array.isArray(value)) return undefined
  return value.filter((entry): entry is string => typeof entry === 'string')
}

function asApprovalPolicy(value: unknown): 'minimal' | 'normal' | 'strict' | undefined {
  return value === 'minimal' || value === 'normal' || value === 'strict' ? value : undefined
}

function extractFirstParagraph(body: string): string {
  const cleaned = body.trim().split(/\n\s*\n/).find(Boolean)
  return cleaned?.replace(/^#+\s*/, '').trim() ?? ''
}

function extractAgentConstraints(body: string): string[] {
  return extractBulletsAfterMarker(body, /^##\s+Constraints\s*$/i)
}

function extractSkillConstraints(body: string): string[] {
  return [
    ...extractBulletsAfterMarker(body, /^When applying this skill:\s*$/i),
    ...extractBulletsAfterMarker(body, /^You should:\s*$/i),
  ]
}

function extractBulletsAfterMarker(body: string, marker: RegExp): string[] {
  const lines = body.split(/\r?\n/)
  const startIndex = lines.findIndex((line) => marker.test(line.trim()))
  if (startIndex === -1) return []

  const items: string[] = []
  for (let index = startIndex + 1; index < lines.length; index += 1) {
    const line = lines[index]?.trim() ?? ''
    if (!line) continue
    if (line.startsWith('#')) break

    const bulletMatch = line.match(/^-\s+(.*)$/)
    if (bulletMatch?.[1]) {
      items.push(bulletMatch[1].trim())
      continue
    }

    if (items.length > 0) break
  }

  return items
}
