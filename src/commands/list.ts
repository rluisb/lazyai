/**
 * List Command
 *
 * Lists available templates, skills, agents, rules, and MCP servers.
 */

import path from 'node:path'
import * as p from '@clack/prompts'
import type { Command } from 'commander'
import { glob } from 'glob'
import { resolveLibraryDir } from '../utils/files.js'
import { showSummaryBox } from '../utils/ui.js'

interface McpServer {
  description: string
  command?: string
  url?: string
  tools: string[]
  enabled: boolean
  requiresInstall: boolean
  installHint?: string
  env?: Record<string, string>
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

type Category = 'all' | 'agents' | 'skills' | 'templates' | 'rules' | 'servers' | 'mcp' | 'tools' | 'cli'

function extractNameFromPath(filePath: string): string {
  const basename = path.basename(filePath, path.extname(filePath))
  // Convert kebab-case to Title Case
  return basename
    .split('-')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

async function listLibraryItems(libraryDir: string, category: string): Promise<string[]> {
  const pattern = path.join(libraryDir, category, '*.md')
  const files = await glob(pattern)
  return files.map(extractNameFromPath).sort()
}

async function loadMcpCatalog(libraryDir: string): Promise<McpCatalog> {
  const catalogPath = path.join(libraryDir, 'mcp', 'catalog.json')
  const { readFile } = await import('../utils/files.js')
  const content = readFile(catalogPath)
  return JSON.parse(content) as McpCatalog
}

function formatServerStatus(server: McpServer): string {
  const parts: string[] = []

  if (server.enabled) {
    parts.push('✓ enabled')
  } else {
    parts.push('○ disabled')
  }

  if (server.requiresInstall) {
    parts.push('requires install')
  }

  if (server.env && Object.keys(server.env).length > 0) {
    parts.push('needs env vars')
  }

  return parts.join(', ')
}

function formatToolsList(tools: string[], maxDisplay = 4): string {
  if (tools.length <= maxDisplay) {
    return tools.join(', ')
  }
  const displayed = tools.slice(0, maxDisplay)
  const remaining = tools.length - maxDisplay
  return `${displayed.join(', ')} +${remaining} more`
}

export function registerList(program: Command): void {
  program
    .command('list [category]')
    .description('List available agents, skills, templates, rules, and MCP servers')
    .option('--json', 'Output as JSON')
    .option('--enabled', 'Show only enabled items (for servers/tools)')
    .action(async (category: Category | undefined, opts: { json?: boolean; enabled?: boolean }) => {
      const libraryDir = resolveLibraryDir(path.dirname(import.meta.url))
      // Normalize category aliases
      let selectedCategory = category ?? 'all'
      if (selectedCategory === 'mcp') selectedCategory = 'servers'
      if (selectedCategory === 'cli') selectedCategory = 'tools'

      // biome-ignore lint/suspicious/noExplicitAny: flexible result structure for JSON output
      const results: Record<string, any> = {}

      // Load data based on category
      const showAgents = selectedCategory === 'all' || selectedCategory === 'agents'
      const showSkills = selectedCategory === 'all' || selectedCategory === 'skills'
      const showTemplates = selectedCategory === 'all' || selectedCategory === 'templates'
      const showRules = selectedCategory === 'all' || selectedCategory === 'rules'
      const showServers = selectedCategory === 'all' || selectedCategory === 'servers'
      const showTools = selectedCategory === 'all' || selectedCategory === 'tools'

      if (showAgents) {
        results.agents = await listLibraryItems(libraryDir, 'agents')
      }

      if (showSkills) {
        results.skills = await listLibraryItems(libraryDir, 'skills')
      }

      if (showTemplates) {
        results.templates = await listLibraryItems(libraryDir, 'templates')
      }

      if (showRules) {
        results.rules = await listLibraryItems(libraryDir, 'rules')
      }

      let mcpCatalog: McpCatalog | undefined
      if (showServers || showTools) {
        mcpCatalog = await loadMcpCatalog(libraryDir)
      }

      if (showServers && mcpCatalog) {
        const servers = Object.entries(mcpCatalog.servers)
          .filter(([, server]) => !opts.enabled || server.enabled)
          .map(([name, server]) => ({
            name,
            ...server,
          }))
        results.servers = servers
      }

      if (showTools && mcpCatalog) {
        const tools = Object.entries(mcpCatalog.cliTools)
          .filter(([, tool]) => !opts.enabled || tool.enabled)
          .map(([name, tool]) => ({
            name,
            ...tool,
          }))
        results.cliTools = tools
      }

      // Output
      if (opts.json) {
        console.log(JSON.stringify(results, null, 2))
        return
      }

      // Pretty print
      p.intro('ai-setup library')

      if (showAgents && results.agents) {
        const agents = results.agents as string[]
        if (agents.length > 0) {
          showSummaryBox('🤖 Agents', [{ label: 'Available', value: `${agents.length} agents` }])
          for (const agent of agents) {
            p.log.message(`  • ${agent}`)
          }
        }
      }

      if (showSkills && results.skills) {
        const skills = results.skills as string[]
        if (skills.length > 0) {
          console.log('')
          showSummaryBox('⚡ Skills', [{ label: 'Available', value: `${skills.length} skills` }])
          for (const skill of skills) {
            p.log.message(`  • ${skill}`)
          }
        }
      }

      if (showTemplates && results.templates) {
        const templates = results.templates as string[]
        if (templates.length > 0) {
          console.log('')
          showSummaryBox('📄 Templates', [{ label: 'Available', value: `${templates.length} templates` }])
          for (const template of templates) {
            p.log.message(`  • ${template}`)
          }
        }
      }

      if (showRules && results.rules) {
        const rules = results.rules as string[]
        if (rules.length > 0) {
          console.log('')
          showSummaryBox('📏 Rules', [{ label: 'Available', value: `${rules.length} rules` }])
          for (const rule of rules) {
            p.log.message(`  • ${rule}`)
          }
        }
      }

      if (showServers && results.servers) {
        const servers = results.servers as Array<McpServer & { name: string }>
        if (servers.length > 0) {
          console.log('')
          const enabledCount = servers.filter((s) => s.enabled).length
          showSummaryBox('🔌 MCP Servers', [
            { label: 'Available', value: `${servers.length} servers` },
            { label: 'Enabled', value: `${enabledCount} by default` },
          ])
          for (const server of servers) {
            const status = formatServerStatus(server)
            const tools = formatToolsList(server.tools)
            p.log.message(`  ${server.enabled ? '✓' : '○'} ${server.name}`)
            p.log.message(`    ${server.description}`)
            p.log.message(`    Tools: ${tools}`)
            if (server.requiresInstall && server.installHint) {
              p.log.message(`    Install: ${server.installHint}`)
            }
          }
        }
      }

      if (showTools && results.cliTools) {
        const tools = results.cliTools as Array<CliTool & { name: string }>
        if (tools.length > 0) {
          console.log('')
          const enabledCount = tools.filter((t) => t.enabled).length
          showSummaryBox('🛠️  CLI Tools', [
            { label: 'Available', value: `${tools.length} tools` },
            { label: 'Enabled', value: `${enabledCount} by default` },
          ])
          for (const tool of tools) {
            p.log.message(`  ${tool.enabled ? '✓' : '○'} ${tool.name}`)
            p.log.message(`    ${tool.description}`)
            if (tool.installHint) {
              p.log.message(`    Install: ${tool.installHint}`)
            }
          }
        }
      }

      // Summary
      console.log('')
      const totalItems =
        ((results.agents as string[])?.length ?? 0) +
        ((results.skills as string[])?.length ?? 0) +
        ((results.templates as string[])?.length ?? 0) +
        ((results.rules as string[])?.length ?? 0) +
        ((results.servers as unknown[])?.length ?? 0) +
        ((results.cliTools as unknown[])?.length ?? 0)

      p.outro(`${totalItems} items available`)
    })
}
