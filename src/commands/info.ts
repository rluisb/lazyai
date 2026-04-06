/**
 * Info Command
 *
 * Shows detailed information about a specific library item.
 */

import path from 'node:path'
import * as p from '@clack/prompts'
import type { Command } from 'commander'
import { glob } from 'glob'
import pc from 'picocolors'
import { readFile, resolveLibraryDir } from '../utils/files.js'
import { showSummaryBox } from '../utils/ui.js'

interface McpServer {
  description: string
  command?: string
  url?: string
  args?: string[]
  env?: Record<string, string>
  headers?: Record<string, string>
  tools: string[]
  enabled: boolean
  requiresInstall: boolean
  installHint?: string
}

interface CliTool {
  description: string
  installHint?: string
  enabled: boolean
}

interface McpCatalog {
  servers: Record<string, McpServer>
  cliTools: Record<string, CliTool>
}

interface ItemInfo {
  type: 'agent' | 'skill' | 'template' | 'rule' | 'server' | 'cli-tool'
  name: string
  path?: string
  content?: string
  metadata?: Record<string, unknown>
}

function extractFrontmatter(content: string): { frontmatter: Record<string, unknown>; body: string } {
  const match = content.match(/^---\n([\s\S]*?)\n---\n([\s\S]*)$/)
  if (!match) {
    return { frontmatter: {}, body: content }
  }

  const fmRaw = match[1] ?? ''
  const body = match[2] ?? ''
  const frontmatter: Record<string, unknown> = {}

  for (const line of fmRaw.split('\n')) {
    const colonIdx = line.indexOf(':')
    if (colonIdx > 0) {
      const key = line.slice(0, colonIdx).trim()
      const value = line.slice(colonIdx + 1).trim()
      frontmatter[key] = value
    }
  }

  return { frontmatter, body }
}

function extractFirstParagraph(content: string): string {
  // Skip heading and get first non-empty paragraph
  const lines = content.split('\n')
  let inParagraph = false
  const paragraphLines: string[] = []

  for (const line of lines) {
    const trimmed = line.trim()

    // Skip headings
    if (trimmed.startsWith('#')) continue

    // Empty line ends paragraph if we're in one
    if (trimmed === '') {
      if (inParagraph && paragraphLines.length > 0) break
      continue
    }

    // Start/continue paragraph
    inParagraph = true
    paragraphLines.push(trimmed)
  }

  return paragraphLines.join(' ').slice(0, 200) + (paragraphLines.join(' ').length > 200 ? '...' : '')
}

function extractSections(content: string): string[] {
  const headings: string[] = []
  for (const line of content.split('\n')) {
    if (line.startsWith('## ')) {
      headings.push(line.slice(3).trim())
    }
  }
  return headings
}

function toKebabCase(str: string): string {
  return str
    .toLowerCase()
    .replace(/\s+/g, '-')
    .replace(/[^a-z0-9-]/g, '')
}

async function findItem(libraryDir: string, query: string): Promise<ItemInfo | null> {
  const queryLower = query.toLowerCase()
  const queryKebab = toKebabCase(query)

  // Check MCP catalog first
  const catalogPath = path.join(libraryDir, 'mcp', 'catalog.json')
  try {
    const catalogContent = readFile(catalogPath)
    const catalog: McpCatalog = JSON.parse(catalogContent)

    // Check servers
    for (const [name, server] of Object.entries(catalog.servers)) {
      if (name.toLowerCase() === queryLower || name === queryKebab) {
        return {
          type: 'server',
          name,
          metadata: server as unknown as Record<string, unknown>,
        }
      }
    }

    // Check CLI tools
    for (const [name, tool] of Object.entries(catalog.cliTools)) {
      if (name.toLowerCase() === queryLower || name === queryKebab) {
        return {
          type: 'cli-tool',
          name,
          metadata: tool as unknown as Record<string, unknown>,
        }
      }
    }
  } catch {
    // Catalog not found, continue
  }

  // Check markdown files in library directories
  const categories: Array<{ dir: string; type: ItemInfo['type'] }> = [
    { dir: 'agents', type: 'agent' },
    { dir: 'skills', type: 'skill' },
    { dir: 'templates', type: 'template' },
    { dir: 'rules', type: 'rule' },
  ]

  for (const { dir, type } of categories) {
    const pattern = path.join(libraryDir, dir, '*.md')
    const files = await glob(pattern)

    for (const filePath of files) {
      const basename = path.basename(filePath, '.md')
      const baseLower = basename.toLowerCase()

      if (baseLower === queryLower || basename === queryKebab || baseLower === queryKebab) {
        const content = readFile(filePath)
        const { frontmatter, body } = extractFrontmatter(content)

        return {
          type,
          name: basename,
          path: filePath,
          content: body,
          metadata: frontmatter,
        }
      }
    }
  }

  return null
}

function displayServerInfo(name: string, server: McpServer): void {
  p.intro(`🔌 MCP Server: ${pc.bold(name)}`)

  const summaryItems = [
    { label: 'Status', value: server.enabled ? pc.green('✓ Enabled by default') : pc.dim('○ Disabled by default') },
    { label: 'Description', value: server.description },
  ]

  if (server.command) {
    const fullCommand = server.args ? `${server.command} ${server.args.join(' ')}` : server.command
    summaryItems.push({ label: 'Command', value: fullCommand })
  }

  if (server.url) {
    summaryItems.push({ label: 'URL', value: server.url })
  }

  if (server.requiresInstall && server.installHint) {
    summaryItems.push({ label: 'Install', value: server.installHint })
  }

  showSummaryBox('Configuration', summaryItems)

  // Tools
  console.log('')
  p.log.step('Available Tools')
  for (const tool of server.tools) {
    p.log.message(`  • ${tool}`)
  }

  // Environment variables
  if (server.env && Object.keys(server.env).length > 0) {
    console.log('')
    p.log.step('Required Environment Variables')
    for (const [key, value] of Object.entries(server.env)) {
      p.log.message(`  ${key}=${pc.dim(value)}`)
    }
  }

  // Headers
  if (server.headers && Object.keys(server.headers).length > 0) {
    console.log('')
    p.log.step('HTTP Headers')
    for (const [key, value] of Object.entries(server.headers)) {
      p.log.message(`  ${key}: ${pc.dim(value)}`)
    }
  }

  p.outro(`Enable with: ${pc.cyan(`ai-setup init --enable-servers ${name}`)}`)
}

function displayCliToolInfo(name: string, tool: CliTool): void {
  p.intro(`🛠️  CLI Tool: ${pc.bold(name)}`)

  const summaryItems = [
    { label: 'Status', value: tool.enabled ? pc.green('✓ Enabled by default') : pc.dim('○ Disabled by default') },
    { label: 'Description', value: tool.description },
  ]

  if (tool.installHint) {
    summaryItems.push({ label: 'Install', value: tool.installHint })
  }

  showSummaryBox('Configuration', summaryItems)

  p.outro('CLI tools are integrated into the generated AGENTS.md instructions')
}

function displayMarkdownInfo(item: ItemInfo): void {
  const typeEmoji: Record<string, string> = {
    agent: '🤖',
    skill: '⚡',
    template: '📄',
    rule: '📏',
  }

  const typeLabel: Record<string, string> = {
    agent: 'Agent',
    skill: 'Skill',
    template: 'Template',
    rule: 'Rule',
  }

  p.intro(`${typeEmoji[item.type]} ${typeLabel[item.type]}: ${pc.bold(item.name)}`)

  // Show frontmatter metadata if present
  if (item.metadata && Object.keys(item.metadata).length > 0) {
    const summaryItems = Object.entries(item.metadata).map(([key, value]) => ({
      label: key.charAt(0).toUpperCase() + key.slice(1),
      value: String(value),
    }))
    showSummaryBox('Metadata', summaryItems)
    console.log('')
  }

  // Show first paragraph as description
  if (item.content) {
    const description = extractFirstParagraph(item.content)
    if (description) {
      p.log.step('Description')
      p.log.message(`  ${description}`)
      console.log('')
    }

    // Show sections
    const sections = extractSections(item.content)
    if (sections.length > 0) {
      p.log.step('Sections')
      for (const section of sections) {
        p.log.message(`  • ${section}`)
      }
      console.log('')
    }
  }

  if (item.path) {
    p.log.info(`Source: ${pc.dim(item.path)}`)
  }

  p.outro(`View full content with: ${pc.cyan(`cat "${item.path}"`)}`)
}

export function registerInfo(program: Command): void {
  program
    .command('info <item>')
    .description('Show detailed information about a library item (agent, skill, template, rule, or MCP server)')
    .option('--json', 'Output as JSON')
    .action(async (item: string, opts: { json?: boolean }) => {
      const libraryDir = resolveLibraryDir(path.dirname(import.meta.url))

      const found = await findItem(libraryDir, item)

      if (!found) {
        p.log.error(`Item "${item}" not found in library`)
        p.log.info('Use `ai-setup list` to see all available items')
        process.exitCode = 1
        return
      }

      if (opts.json) {
        console.log(JSON.stringify(found, null, 2))
        return
      }

      // Display based on type
      switch (found.type) {
        case 'server':
          displayServerInfo(found.name, found.metadata as unknown as McpServer)
          break
        case 'cli-tool':
          displayCliToolInfo(found.name, found.metadata as unknown as CliTool)
          break
        default:
          displayMarkdownInfo(found)
      }
    })
}
