import fs from 'node:fs'
import path from 'node:path'
import { readFile } from '../utils/files.js'

export type ContractSeverity = 'warn' | 'error'

export interface ContractIssue {
  source: string
  severity: ContractSeverity
  code: string
  message: string
}

interface SkillContract {
  source: string
  name: string
  output: string | undefined
  consumes: string[]
  producesFor: string[]
  mcpTools: string[]
}

interface FrontmatterAttrs {
  name?: string | undefined
  output?: string | undefined
  consumes?: string[] | string | undefined
  produces_for?: string[] | string | undefined
  mcp_tools?: string[] | string | undefined
}

/**
 * Walk the library skills and agents directories, parse frontmatter,
 * and return normalized SkillContract records.
 */
export function loadSkillContracts(libraryDir: string): SkillContract[] {
  const contracts: SkillContract[] = []
  const dirs = [`${libraryDir}/skills`, `${libraryDir}/agents`]

  for (const dir of dirs) {
    try {
      const entries = listMarkdownFiles(dir)
      for (const entry of entries) {
        const source = entry.replace(`${libraryDir}/`, '')
        const raw = readFile(entry)
        const attrs = parseFrontmatter(raw)

        // Skip files without a name (not a valid skill/agent).
        if (!attrs.name) continue
        // Skip AGENTS.md, AGENT.md, and underscore-prefixed.
        const base = entry.split('/').pop() ?? ''
        if (base === 'AGENTS.md' || base === 'AGENT.md' || base.startsWith('_')) continue

        contracts.push({
          source,
          name: attrs.name,
          output: attrs.output,
          consumes: asStringArray(attrs.consumes),
          producesFor: asStringArray(attrs.produces_for),
          mcpTools: asStringArray(attrs.mcp_tools),
        })
      }
    } catch {
      // Directory may not exist — skip.
    }
  }

  return contracts
}

/**
 * Validate the producer/consumer chain across all skill contracts.
 * Returns issues found. Empty array = chain is clean.
 */
export function validateChain(contracts: SkillContract[]): ContractIssue[] {
  const issues: ContractIssue[] = []
  const byName = new Map<string, SkillContract>()

  // Index by name for fast lookup.
  for (const c of contracts) {
    if (byName.has(c.name)) {
      issues.push({
        source: c.source,
        severity: 'error',
        code: 'duplicate-name',
        message: `skill/agent name "${c.name}" is declared in multiple files (last: ${c.source})`,
      })
    }
    byName.set(c.name, c)
  }

  // Index producers: which contracts declare `output`?
  const producers = new Map<string, string>() // artifact name → skill name
  for (const c of contracts) {
    if (c.output) {
      producers.set(c.output, c.name)
    }
  }

  // Validate each contract.
  for (const c of contracts) {
    // Check produces_for: every downstream name must exist.
    for (const target of c.producesFor) {
      if (!byName.has(target)) {
        issues.push({
          source: c.source,
          severity: 'warn',
          code: 'missing-downstream',
          message: `"${c.name}" declares produces_for "${target}", but no skill/agent with that name exists`,
        })
      }
    }

    // Check consumes: every artifact must have a producer.
    for (const artifact of c.consumes) {
      if (!producers.has(artifact)) {
        // Check if it's a physical file (e.g., "constitution.md").
        // For now, flag as warn — physical files can be produced by the
        // scaffold step which isn't tracked as a skill.
        if (!artifact.startsWith('.') && !artifact.includes('/')) {
          issues.push({
            source: c.source,
            severity: 'warn',
            code: 'missing-producer',
            message: `"${c.name}" declares consumes "${artifact}", but no skill produces it (may be a scaffold artifact)`,
          })
        }
      }
    }

    // Check output: if declared, the contract name should make sense.
    if (c.output && c.output === c.name) {
      issues.push({
        source: c.source,
        severity: 'warn',
        code: 'self-output',
        message: `"${c.name}" declares output "${c.output}" which matches its own name — may indicate a misconfiguration`,
      })
    }
  }

  // Check for orphans: skills not consumed by anyone and not root skills.
  const consumedBy = new Set<string>()
  for (const c of contracts) {
    for (const target of c.producesFor) {
      consumedBy.add(target)
    }
  }

  const rootSkills = new Set(['speckit-constitution', 'rpi', 'bugfix', 'spike', 'proof-of-concept', 'housekeeping'])

  for (const c of contracts) {
    if (!consumedBy.has(c.name) && !rootSkills.has(c.name) && c.source.startsWith('skills/')) {
      issues.push({
        source: c.source,
        severity: 'warn',
        code: 'orphan-skill',
        message: `"${c.name}" is not consumed by any other skill and is not a root skill — may be unreachable`,
      })
    }
  }

  return issues
}

/**
 * Parse the YAML-like frontmatter from a markdown file.
 * Returns attribute keys in snake_case (normalized).
 */
function parseFrontmatter(content: string): FrontmatterAttrs {
  if (!content.startsWith('---\n')) return {}

  const closingIndex = content.indexOf('\n---', 4)
  if (closingIndex === -1) return {}

  const fm = content.slice(4, closingIndex)
  const attrs: Record<string, unknown> = {}

  for (const line of fm.split('\n')) {
    const trimmed = line.trim()
    if (!trimmed || trimmed.startsWith('#')) continue

    const match = trimmed.match(/^([a-zA-Z0-9_-]+):\s*(.*)$/)
    if (!match) continue

    const [, rawKey, rawValue] = match
    if (!rawKey) continue
    const key: string = rawKey
    const val = rawValue?.trim() ?? ''

    if (!val) {
      // List format — read next lines.
      attrs[key] = []
      continue
    }

    // Handle bracket arrays: [item1, item2]
    if (val.startsWith('[') && val.endsWith(']')) {
      attrs[key] = val
        .slice(1, -1)
        .split(',')
        .map((s) => s.trim().replace(/^['"]|['"]$/g, ''))
        .filter(Boolean)
      continue
    }

    // Scalar values.
    if (val === 'true') attrs[key] = true
    else if (val === 'false') attrs[key] = false
    else attrs[key] = val.replace(/^['"]|['"]$/g, '')
  }

  return {
    name: attrs.name as string | undefined,
    output: attrs.output as string | undefined,
    consumes: (attrs.consumes as string[] | string) ?? (attrs.consumes as string[] | string | undefined),
    produces_for: (attrs.produces_for as string[] | string) ?? (attrs.produces_for as string[] | string | undefined),
    mcp_tools: (attrs.mcp_tools as string[] | string) ?? (attrs.mcp_tools as string[] | string | undefined),
  }
}

function asStringArray(value?: string[] | string): string[] {
  if (!value) return []
  if (Array.isArray(value)) return value
  return [value]
}

function listMarkdownFiles(dir: string): string[] {
  try {
    return fs
      .readdirSync(dir)
      .filter((f: string) => f.endsWith('.md'))
      .map((f: string) => path.join(dir, f))
  } catch {
    return []
  }
}
