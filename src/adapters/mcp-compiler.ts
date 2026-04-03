import path from 'node:path'
import { ensureDir, fileExists, fileHash, readFile, writeFile } from '../utils/files.js'
import type { FileRecord, ToolId } from '../types.js'

interface McpServer {
  description?: string
  command?: string
  args?: string[]
  env?: Record<string, string>
  url?: string
  headers?: Record<string, string>
  tools?: string[]
  enabled?: boolean
}

interface McpCatalog {
  servers: Record<string, McpServer>
}

function readCanonicalMcp(targetDir: string): McpCatalog | null {
  const mcpPath = path.join(targetDir, '.ai', 'mcp.json')
  if (!fileExists(mcpPath)) return null
  try {
    return JSON.parse(readFile(mcpPath)) as McpCatalog
  } catch {
    return null
  }
}

function getEnabledServers(catalog: McpCatalog): Record<string, McpServer> {
  const result: Record<string, McpServer> = {}
  for (const [name, server] of Object.entries(catalog.servers)) {
    if (server.enabled !== false) {
      result[name] = server
    }
  }
  return result
}

function toMcpJson(servers: Record<string, McpServer>): Record<string, unknown> {
  const mcpServers: Record<string, unknown> = {}
  for (const [name, server] of Object.entries(servers)) {
    if (server.url) {
      const entry: Record<string, unknown> = {
        url: server.url,
      }
      if (server.headers) entry.headers = server.headers
      mcpServers[name] = entry
      continue
    }
    const entry: Record<string, unknown> = {
      command: server.command,
      args: server.args,
    }
    if (server.env) entry.env = server.env
    mcpServers[name] = entry
  }
  return { mcpServers }
}

function toOpenCodeJsonc(allServers: Record<string, McpServer>): Record<string, unknown> {
  const mcp: Record<string, unknown> = {}
  for (const [name, server] of Object.entries(allServers)) {
    const isEnabled = server.enabled !== false
    if (server.url) {
      mcp[name] = {
        type: 'remote',
        enabled: isEnabled,
        url: server.url,
        ...(server.headers ? { headers: transformEnvSyntax(server.headers, '{env:$1}') } : {}),
      }
    } else {
      const entry: Record<string, unknown> = {
        type: 'local',
        enabled: isEnabled,
        command: [server.command, ...(server.args ?? [])],
      }
      if (server.env) entry.environment = transformEnvSyntax(server.env, '{env:$1}')
      mcp[name] = entry
    }
  }
  return mcp
}

function toCopilotMcp(servers: Record<string, McpServer>): Record<string, unknown> {
  const result: Record<string, unknown> = {}
  for (const [name, server] of Object.entries(servers)) {
    if (server.url) {
      const entry: Record<string, unknown> = {
        type: 'sse',
        url: server.url,
      }
      if (server.headers) entry.headers = server.headers
      result[name] = entry
      continue
    }
    const entry: Record<string, unknown> = {
      type: 'stdio',
      command: server.command,
      args: server.args,
    }
    if (server.env) entry.env = server.env
    result[name] = entry
  }
  return { servers: result }
}

function toGeminiSettings(servers: Record<string, McpServer>): Record<string, unknown> {
  const mcpServers: Record<string, unknown> = {}
  for (const [name, server] of Object.entries(servers)) {
    if (server.url) {
      console.warn(`⚠️  Skipping remote server "${name}" for gemini (not supported)`)
      continue
    }
    const entry: Record<string, unknown> = {
      command: server.command,
      args: server.args,
    }
    if (server.env) entry.env = transformEnvSyntax(server.env, '$$$1')
    mcpServers[name] = entry
  }
  return { mcpServers }
}

function transformEnvSyntax(envObj: Record<string, string>, targetPattern: string): Record<string, string> {
  const result: Record<string, string> = {}
  for (const [key, value] of Object.entries(envObj)) {
    result[key] = value.replace(/\$\{(\w+)\}/g, targetPattern)
  }
  return result
}

export interface CompileMcpOptions {
  canonicalDir: string
  toolTargetDir: string
  toolId: ToolId
  fileRecords: FileRecord[]
}

export async function compileMcp(opts: CompileMcpOptions): Promise<void> {
  const catalog = readCanonicalMcp(opts.canonicalDir)
  if (!catalog) return

  const enabledServers = getEnabledServers(catalog)
  if (Object.keys(enabledServers).length === 0) return

  switch (opts.toolId) {
    case 'opencode': {
      const configPath = path.join(opts.toolTargetDir, 'opencode.jsonc')
      const ocMcpContent = toOpenCodeJsonc(catalog.servers)

      let existingConfig: Record<string, unknown> = {}
      if (fileExists(configPath)) {
        try {
          const raw = readFile(configPath)
          const stripped = raw.replace(/\/\/.*$/gm, '').replace(/\/\*[\s\S]*?\*\//g, '')
          existingConfig = JSON.parse(stripped) as Record<string, unknown>
        } catch {
          // If parse fails, start fresh
        }
      }

      const merged = {
        ...existingConfig,
        $schema: 'https://opencode.ai/config.json',
        mcp: ocMcpContent,
      }

      writeFile(configPath, JSON.stringify(merged, null, 2) + '\n')
      opts.fileRecords.push({
        path: 'opencode.jsonc',
        hash: fileHash(configPath),
        source: 'compiled:mcp:opencode',
      })
      break
    }

    case 'claude-code': {
      const mcpPath = path.join(opts.toolTargetDir, '.mcp.json')
      const content = toMcpJson(enabledServers)
      writeFile(mcpPath, JSON.stringify(content, null, 2) + '\n')
      opts.fileRecords.push({
        path: '.mcp.json',
        hash: fileHash(mcpPath),
        source: 'compiled:mcp',
      })
      break
    }

    case 'copilot': {
      const vscodeMcpPath = path.join(opts.toolTargetDir, '.vscode', 'mcp.json')
      ensureDir(path.join(opts.toolTargetDir, '.vscode'))
      const content = toCopilotMcp(enabledServers)
      writeFile(vscodeMcpPath, JSON.stringify(content, null, 2) + '\n')
      opts.fileRecords.push({
        path: '.vscode/mcp.json',
        hash: fileHash(vscodeMcpPath),
        source: 'compiled:mcp:copilot',
      })
      break
    }

    case 'gemini': {
      const settingsPath = path.join(opts.toolTargetDir, '.gemini', 'settings.json')
      ensureDir(path.join(opts.toolTargetDir, '.gemini'))
      const content = toGeminiSettings(enabledServers)
      writeFile(settingsPath, JSON.stringify(content, null, 2) + '\n')
      opts.fileRecords.push({
        path: '.gemini/settings.json',
        hash: fileHash(settingsPath),
        source: 'compiled:mcp:gemini',
      })
      break
    }

    case 'pi': {
      const mcpPath = path.join(opts.toolTargetDir, '.mcp.json')
      const content = toMcpJson(enabledServers)
      writeFile(mcpPath, JSON.stringify(content, null, 2) + '\n')
      opts.fileRecords.push({
        path: '.mcp.json',
        hash: fileHash(mcpPath),
        source: 'compiled:mcp',
      })
      break
    }
  }
}
