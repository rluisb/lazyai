import path from 'node:path'
import { fileExists, listDir, readFile } from '../utils/files.js'

export type OrchestrationArtifactType = 'workflow' | 'chain' | 'team' | 'domain' | 'mode'
export type OrchestrationListCategory = 'workflows' | 'chains' | 'teams' | 'domains' | 'modes'
export type OrchestrationSource = 'project' | 'library'

interface OrchestrationDirectoryConfig {
  category: OrchestrationListCategory
  type: OrchestrationArtifactType
  relativeDir: string
  extension: '.json' | '.md'
}

export interface OrchestrationCatalogItem {
  type: OrchestrationArtifactType
  category: OrchestrationListCategory
  name: string
  source: OrchestrationSource
  path: string
  description?: string
  content: string
  data?: Record<string, unknown>
  metadata?: Record<string, unknown>
  body?: string
}

function withOptionalDescription<T extends Record<string, unknown>>(description: string | undefined, value: T): T & { description?: string } {
  return description ? { ...value, description } : value
}

export interface OrchestrationCatalogSummary {
  name: string
  type: OrchestrationArtifactType
  category: OrchestrationListCategory
  source: OrchestrationSource
  path: string
  description?: string
}

const ORCHESTRATION_DIRECTORIES: OrchestrationDirectoryConfig[] = [
  { category: 'workflows', type: 'workflow', relativeDir: 'workflows', extension: '.json' },
  { category: 'chains', type: 'chain', relativeDir: 'chains', extension: '.json' },
  { category: 'teams', type: 'team', relativeDir: 'teams', extension: '.json' },
  { category: 'domains', type: 'domain', relativeDir: 'skills/domains', extension: '.md' },
  { category: 'modes', type: 'mode', relativeDir: 'skills/modes', extension: '.md' },
]

function normalizeName(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/\.[a-z0-9]+$/i, '')
    .replace(/\s+/g, '-')
}

function getProjectRoot(projectDir: string): string {
  return path.join(projectDir, '.ai', 'orchestration')
}

function getLibraryRoot(libraryDir: string): string {
  return path.join(libraryDir, 'orchestration')
}

function parseFrontmatter(content: string): { metadata: Record<string, unknown>; body: string } {
  const match = content.match(/^---\n([\s\S]*?)\n---\n?([\s\S]*)$/)
  if (!match) {
    return { metadata: {}, body: content }
  }

  const metadata: Record<string, unknown> = {}
  let currentArrayKey: string | null = null

  for (const rawLine of (match[1] ?? '').split('\n')) {
    const arrayMatch = rawLine.match(/^\s*-\s*(.+)$/)
    if (arrayMatch && currentArrayKey) {
      const existing = metadata[currentArrayKey]
      const nextValue = arrayMatch[1]?.trim() ?? ''
      if (Array.isArray(existing)) {
        existing.push(nextValue)
      } else {
        metadata[currentArrayKey] = [nextValue]
      }
      continue
    }

    const keyMatch = rawLine.match(/^([a-zA-Z0-9_]+):\s*(.*)$/)
    if (!keyMatch) {
      currentArrayKey = null
      continue
    }

    const [, rawKey = '', rawValue = ''] = keyMatch
    const key = rawKey.trim()
    const value = rawValue.trim()
    if (value.length === 0) {
      metadata[key] = []
      currentArrayKey = key
      continue
    }

    metadata[key] = value
    currentArrayKey = null
  }

  return {
    metadata,
    body: match[2] ?? '',
  }
}

function readCatalogDirectory(rootDir: string, config: OrchestrationDirectoryConfig, source: OrchestrationSource): OrchestrationCatalogItem[] {
  const dir = path.join(rootDir, config.relativeDir)
  if (!fileExists(dir)) {
    return []
  }

  return listDir(dir)
    .filter((entry) => entry.endsWith(config.extension))
    .sort((a, b) => a.localeCompare(b))
    .map((entry) => {
      const itemPath = path.join(dir, entry)
      const content = readFile(itemPath)
      const name = entry.slice(0, -config.extension.length)

      if (config.extension === '.json') {
        const data = JSON.parse(content) as Record<string, unknown>
        return withOptionalDescription(typeof data.description === 'string' ? data.description : undefined, {
          type: config.type,
          category: config.category,
          name,
          source,
          path: itemPath,
          content,
          data,
        }) satisfies OrchestrationCatalogItem
      }

      const { metadata, body } = parseFrontmatter(content)
      return withOptionalDescription(typeof metadata.description === 'string' ? metadata.description : undefined, {
        type: config.type,
        category: config.category,
        name,
        source,
        path: itemPath,
        content,
        metadata,
        body,
      }) satisfies OrchestrationCatalogItem
    })
}

function mergeCatalogItems(projectItems: OrchestrationCatalogItem[], libraryItems: OrchestrationCatalogItem[]): OrchestrationCatalogItem[] {
  const merged = new Map<string, OrchestrationCatalogItem>()

  for (const item of projectItems) {
    merged.set(normalizeName(item.name), item)
  }

  for (const item of libraryItems) {
    const key = normalizeName(item.name)
    if (!merged.has(key)) {
      merged.set(key, item)
    }
  }

  return Array.from(merged.values()).sort((a, b) => a.name.localeCompare(b.name))
}

function getDirectoryConfig(category: OrchestrationListCategory): OrchestrationDirectoryConfig {
  const config = ORCHESTRATION_DIRECTORIES.find((item) => item.category === category)
  if (!config) {
    throw new Error(`Unsupported orchestration category: ${category}`)
  }
  return config
}

export function listOrchestrationItems(projectDir: string, libraryDir: string, category: OrchestrationListCategory): OrchestrationCatalogSummary[] {
  const config = getDirectoryConfig(category)
  const projectItems = readCatalogDirectory(getProjectRoot(projectDir), config, 'project')
  const libraryItems = readCatalogDirectory(getLibraryRoot(libraryDir), config, 'library')

  return mergeCatalogItems(projectItems, libraryItems).map((item) => ({
    name: item.name,
    type: item.type,
    category: item.category,
    source: item.source,
    path: item.path,
    ...(item.description ? { description: item.description } : {}),
  }))
}

export function findOrchestrationItem(projectDir: string, libraryDir: string, query: string): OrchestrationCatalogItem | null {
  const normalizedQuery = normalizeName(query)

  for (const config of ORCHESTRATION_DIRECTORIES) {
    const items = listOrchestrationItems(projectDir, libraryDir, config.category)
    const match = items.find((item) => normalizeName(item.name) === normalizedQuery)
    if (!match) {
      continue
    }

    const rootDir = match.source === 'project' ? getProjectRoot(projectDir) : getLibraryRoot(libraryDir)
    const fullItems = readCatalogDirectory(rootDir, config, match.source)
    return fullItems.find((item) => normalizeName(item.name) === normalizedQuery) ?? null
  }

  return null
}

function countCategory(rootDir: string, category: OrchestrationListCategory): number {
  const config = getDirectoryConfig(category)
  const dir = path.join(rootDir, config.relativeDir)
  if (!fileExists(dir)) {
    return 0
  }

  return listDir(dir).filter((entry) => entry.endsWith(config.extension)).length
}

export function getOrchestrationCounts(projectDir: string, libraryDir: string): {
  scaffolded: boolean
  project: Record<OrchestrationListCategory, number>
  library: Record<OrchestrationListCategory, number>
} {
  const projectRoot = getProjectRoot(projectDir)
  const libraryRoot = getLibraryRoot(libraryDir)

  const project = {
    workflows: countCategory(projectRoot, 'workflows'),
    chains: countCategory(projectRoot, 'chains'),
    teams: countCategory(projectRoot, 'teams'),
    domains: countCategory(projectRoot, 'domains'),
    modes: countCategory(projectRoot, 'modes'),
  }

  const library = {
    workflows: countCategory(libraryRoot, 'workflows'),
    chains: countCategory(libraryRoot, 'chains'),
    teams: countCategory(libraryRoot, 'teams'),
    domains: countCategory(libraryRoot, 'domains'),
    modes: countCategory(libraryRoot, 'modes'),
  }

  return {
    scaffolded: fileExists(projectRoot),
    project,
    library,
  }
}
