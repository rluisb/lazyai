import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import type { BaseAgentDefinition, SkillDefinition } from '../types.js'
import { parseMarkdownFrontmatter } from '../loader.js'

const DEFAULT_ALLOWED_TOOLS = ['Read', 'Grep', 'Glob', 'Edit', 'Write', 'Bash']

export interface ScannedCatalog {
  agents: Record<string, BaseAgentDefinition>
  skills: Record<string, SkillDefinition>
}

interface MtimeCacheEntry {
  mtime: number
  result: ScannedCatalog
}

const scanCache = new Map<string, MtimeCacheEntry>()

function dirMtime(dirPath: string): number {
  try {
    return fs.statSync(dirPath).mtimeMs
  } catch {
    return 0
  }
}

function normalizeName(value: unknown, fileName: string): string {
  const raw = typeof value === 'string' && value.trim() ? value : path.basename(fileName, path.extname(fileName))
  return raw.toLowerCase().replace(/\s+/g, '-')
}

function extractFirstParagraph(body: string): string {
  const cleaned = body.trim().split(/\n\s*\n/).find(Boolean)
  return cleaned?.replace(/^#+\s*/, '').trim() ?? ''
}

function extractConstraints(body: string, markers: RegExp[]): string[] {
  const lines = body.split(/\r?\n/)
  const items: string[] = []
  for (const marker of markers) {
    const start = lines.findIndex((l) => marker.test(l.trim()))
    if (start === -1) continue
    for (let i = start + 1; i < lines.length; i++) {
      const line = lines[i]?.trim() ?? ''
      if (!line) continue
      if (line.startsWith('#')) break
      const m = line.match(/^-\s+(.*)$/)
      if (m?.[1]) { items.push(m[1].trim()); continue }
      if (items.length > 0) break
    }
  }
  return items
}

function readAgentsMd(
  dir: string,
  source: 'user_global' | 'user_project',
): Record<string, BaseAgentDefinition> {
  const records: Record<string, BaseAgentDefinition> = {}
  if (!fs.existsSync(dir)) return records
  for (const entry of fs.readdirSync(dir)) {
    if (!entry.endsWith('.md')) continue
    const filePath = path.join(dir, entry)
    try {
      const raw = fs.readFileSync(filePath, 'utf-8')
      const parsed = parseMarkdownFrontmatter(raw)
      const key = normalizeName(parsed.attributes.name, entry)
      const modelHint = typeof parsed.attributes.model === 'string' ? parsed.attributes.model : undefined
      records[key] = {
        kind: 'agent',
        name: key,
        displayName: (typeof parsed.attributes.name === 'string' ? parsed.attributes.name : key),
        description: extractFirstParagraph(parsed.body),
        source,
        path: filePath,
        prompt: parsed.body.trim(),
        allowedTools: DEFAULT_ALLOWED_TOOLS,
        constraints: extractConstraints(parsed.body, [/^##\s+Constraints\s*$/i]),
        ...(modelHint ? { modelHint } : {}),
      }
    } catch {
      // Skip unparseable files
    }
  }
  return records
}

function readSkillsMd(
  dir: string,
  source: 'user_global' | 'user_project',
): Record<string, SkillDefinition> {
  const records: Record<string, SkillDefinition> = {}
  if (!fs.existsSync(dir)) return records

  // Two layouts: dir/SKILL.md (opencode style) or dir/*.md (flat)
  for (const entry of fs.readdirSync(dir)) {
    const entryPath = path.join(dir, entry)
    let filePath: string
    try {
      const stat = fs.statSync(entryPath)
      if (stat.isDirectory()) {
        filePath = path.join(entryPath, 'SKILL.md')
        if (!fs.existsSync(filePath)) continue
      } else if (entry.endsWith('.md')) {
        filePath = entryPath
      } else {
        continue
      }

      const raw = fs.readFileSync(filePath, 'utf-8')
      const parsed = parseMarkdownFrontmatter(raw)
      const key = normalizeName(parsed.attributes.name, entry)
      const allowedTools = Array.isArray(parsed.attributes.allowed_tools)
        ? (parsed.attributes.allowed_tools as unknown[]).filter((v): v is string => typeof v === 'string')
        : undefined
      const modelHint = typeof parsed.attributes.model_hint === 'string' ? parsed.attributes.model_hint : undefined
      const rawPolicy = parsed.attributes.approval_policy
      const approvalPolicy = rawPolicy === 'minimal' || rawPolicy === 'normal' || rawPolicy === 'strict' ? rawPolicy : undefined

      records[key] = {
        kind: 'domain',
        name: key,
        description: typeof parsed.attributes.description === 'string' ? parsed.attributes.description : '',
        source,
        path: filePath,
        prompt: parsed.body.trim(),
        constraints: extractConstraints(parsed.body, [
          /^When applying this skill:\s*$/i,
          /^You should:\s*$/i,
        ]),
        ...(allowedTools ? { allowedTools } : {}),
        ...(modelHint ? { modelHint } : {}),
        ...(approvalPolicy ? { approvalPolicy } : {}),
      }
    } catch {
      // Skip unparseable entries
    }
  }
  return records
}

function scanDir(dir: string, source: 'user_global' | 'user_project'): ScannedCatalog {
  return {
    agents: readAgentsMd(path.join(dir, 'agents'), source),
    skills: readSkillsMd(path.join(dir, 'skills'), source),
  }
}

function cachedScan(dir: string, source: 'user_global' | 'user_project'): ScannedCatalog {
  const agentsDir = path.join(dir, 'agents')
  const skillsDir = path.join(dir, 'skills')
  const mtime = dirMtime(agentsDir) + dirMtime(skillsDir)
  const hit = scanCache.get(dir)
  if (hit && hit.mtime === mtime) return hit.result
  const result = scanDir(dir, source)
  scanCache.set(dir, { mtime, result })
  return result
}

export function scanOpencodeGlobal(): ScannedCatalog {
  const dir = path.join(os.homedir(), '.config', 'opencode')
  return cachedScan(dir, 'user_global')
}

export function scanClaudeGlobal(): ScannedCatalog {
  const dir = path.join(os.homedir(), '.claude')
  return cachedScan(dir, 'user_global')
}

export function scanClaudeProject(projectRoot: string): ScannedCatalog {
  const dir = path.join(projectRoot, '.claude')
  return cachedScan(dir, 'user_project')
}

export function mergeScannedCatalogs(...catalogs: ScannedCatalog[]): ScannedCatalog {
  const result: ScannedCatalog = { agents: {}, skills: {} }
  for (const c of catalogs) {
    Object.assign(result.agents, c.agents)
    Object.assign(result.skills, c.skills)
  }
  return result
}

export function clearScanCache(): void {
  scanCache.clear()
}
