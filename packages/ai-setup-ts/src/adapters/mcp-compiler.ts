import path from 'node:path'
import type { FileRecord, ToolId } from '../types.js'
import { ensureDir, fileExists, fileHash, readFile, writeFile } from '../utils/files.js'
import { stripJsonComments } from '../utils/jsonc.js'

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

  const ocDir = path.join(opts.toolTargetDir, '.opencode')
  const configPath = path.join(ocDir, 'opencode.jsonc')
  const ocMcpContent = toOpenCodeJsonc(catalog.servers)

  let existingConfig: Record<string, unknown> = {}
  if (fileExists(configPath)) {
    try {
      const raw = readFile(configPath)
      existingConfig = JSON.parse(stripJsonComments(raw)) as Record<string, unknown>
    } catch {
      // If parse fails, start fresh
    }
  }

  existingConfig.$schema = 'https://opencode.ai/config.json'
  existingConfig.mcp = mergeOpenCodeMcpServers(existingConfig.mcp, ocMcpContent)

  ensureDir(ocDir)
  writeFile(configPath, `${JSON.stringify(existingConfig, null, 2)}\n`)
  opts.fileRecords.push({
    path: '.opencode/opencode.jsonc',
    hash: fileHash(configPath),
    source: 'compiled:mcp:opencode',
    owner: 'library',
  })
}

/**
 * Merge ai-setup-managed MCP entries into whatever the user currently has
 * under `mcp` in their opencode.jsonc. User-authored servers NOT in the
 * managed set are preserved; managed servers win on key collision.
 *
 * Mirrors Go's `mergeOpenCodeMcpServers` at internal/adapter/mcp_compiler.go.
 */
function mergeOpenCodeMcpServers(
  existingRaw: unknown,
  managed: Record<string, unknown>,
): Record<string, unknown> {
  const merged: Record<string, unknown> = {}
  if (existingRaw !== null && typeof existingRaw === 'object' && !Array.isArray(existingRaw)) {
    for (const [name, entry] of Object.entries(existingRaw)) {
      if (name in managed) continue
      merged[name] = entry
    }
  }
  for (const [name, entry] of Object.entries(managed)) {
    merged[name] = entry
  }
  return merged
}
