import { existsSync, readdirSync, statSync } from 'node:fs'
import { join } from 'node:path'
import { loadConfig } from '../utils/toml.js'

export interface ExtensionDescriptor {
  name: string
  path: string
  kind: 'toml' | 'local'
  content: ExtensionContent
}

export interface ExtensionContent {
  agents: string[]
  skills: string[]
  prompts: string[]
  rules: string[]
  templates: string[]
  mcpServers: Array<{ name: string; path: string }>
}

const CONTENT_DIRS = ['agents', 'skills', 'prompts', 'rules', 'templates'] as const

/**
 * Discover all extensions from TOML config and local .ai/extensions/.
 */
export function discoverExtensions(targetDir: string): ExtensionDescriptor[] {
  const results: ExtensionDescriptor[] = []

  // 1. TOML-configured extensions
  const config = loadConfig(targetDir)
  const extSection = config.extensions as Record<string, { path?: string }> | undefined
  if (extSection) {
    for (const [name, entry] of Object.entries(extSection)) {
      if (!entry.path) continue
      const absPath = entry.path.startsWith('~')
        ? join(process.env.HOME ?? '', entry.path.slice(1))
        : entry.path.startsWith('/')
          ? entry.path
          : join(targetDir, entry.path)

      if (!existsSync(absPath)) continue
      const content = scanExtensionContent(absPath)
      if (hasAnyContent(content)) {
        results.push({ name, path: absPath, kind: 'toml', content })
      }
    }
  }

  // 2. Local extensions in .ai/extensions/
  const localExtDir = join(targetDir, '.ai', 'extensions')
  if (existsSync(localExtDir)) {
    for (const entry of readdirSync(localExtDir)) {
      const entryPath = join(localExtDir, entry)
      if (!statSync(entryPath).isDirectory()) continue
      if (results.some((e) => e.name === entry)) continue

      const content = scanExtensionContent(entryPath)
      if (hasAnyContent(content)) {
        results.push({ name: entry, path: entryPath, kind: 'local', content })
      }
    }
  }

  return results
}

function scanExtensionContent(extPath: string): ExtensionContent {
  const content: ExtensionContent = { agents: [], skills: [], prompts: [], rules: [], templates: [], mcpServers: [] }

  for (const dir of CONTENT_DIRS) {
    const dirPath = join(extPath, dir)
    if (!existsSync(dirPath)) continue

    for (const entry of readdirSync(dirPath)) {
      const entryPath = join(dirPath, entry)

      if (dir === 'agents') {
        if (statSync(entryPath).isDirectory() && existsSync(join(entryPath, 'AGENT.md'))) {
          content.agents.push(entry)
        } else if (entry.endsWith('.md')) {
          content.agents.push(entry.replace(/\.md$/, ''))
        }
      } else {
        if (entry.endsWith('.md')) {
          content[dir].push(entry.replace(/\.md$/, ''))
        }
      }
    }
  }

  // Scan for MCP server configs
  const mcpJsonPath = join(extPath, 'mcp.json')
  if (existsSync(mcpJsonPath)) {
    content.mcpServers.push({ name: 'extension-root', path: mcpJsonPath })
  }

  // Scan agent-local MCP configs
  const agentsDir = join(extPath, 'agents')
  if (existsSync(agentsDir)) {
    for (const entry of readdirSync(agentsDir)) {
      const agentMcpPath = join(agentsDir, entry, 'mcp.json')
      if (statSync(join(agentsDir, entry)).isDirectory() && existsSync(agentMcpPath)) {
        content.mcpServers.push({ name: entry, path: agentMcpPath })
      }
    }
  }

  return content
}

function hasAnyContent(content: ExtensionContent): boolean {
  return Object.values(content).some((arr) => Array.isArray(arr) && arr.length > 0)
}

/**
 * Get merged available names from built-in library + all extensions.
 */
export function getExtendedAvailable(
  targetDir: string,
  category: Exclude<keyof ExtensionContent, 'mcpServers'>,
): string[] {
  const builtIn = getBuiltInContent(category, targetDir)
  const extensions = discoverExtensions(targetDir)
  const names = new Set(builtIn)

  for (const ext of extensions) {
    for (const name of ext.content[category]) {
      names.add(name)
    }
  }

  return [...names].sort()
}

function getBuiltInContent(category: Exclude<keyof ExtensionContent, 'mcpServers'>, targetDir: string): string[] {
  const catDir = join(targetDir, 'library', category)
  if (!existsSync(catDir)) return []

  return readdirSync(catDir)
    .filter((e) => {
      const ep = join(catDir, e)
      if (category === 'agents') {
        return statSync(ep).isDirectory() ? existsSync(join(ep, 'AGENT.md')) : e.endsWith('.md')
      }
      return e.endsWith('.md')
    })
    .map((e) => e.replace(/\.md$/, ''))
    .sort()
}
