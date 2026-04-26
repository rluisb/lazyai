import { execFileSync } from 'node:child_process'
import { homedir } from 'node:os'
import path from 'node:path'
import type { FileRecord, SetupScope, ToolId } from '../types.js'
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

interface CopilotPromptInput {
  type: 'promptString'
  id: string
  description: string
  password: true
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
  return { mcpServers: toClaudeCodeMcpInner(servers) }
}

function toClaudeCodeMcpInner(servers: Record<string, McpServer>): Record<string, unknown> {
  const mcpServers: Record<string, unknown> = {}
  for (const [name, server] of Object.entries(servers)) {
    if (server.url) {
      const entry: Record<string, unknown> = { url: server.url }
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
  return mcpServers
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
      continue
    }

    const entry: Record<string, unknown> = {
      type: 'local',
      enabled: isEnabled,
      command: [server.command, ...(server.args ?? [])],
    }
    if (server.env) entry.environment = transformEnvSyntax(server.env, '{env:$1}')
    mcp[name] = entry
  }
  return mcp
}

function mergeOpenCodeMcp(existingRaw: unknown, managed: Record<string, unknown>): Record<string, unknown> {
  const merged: Record<string, unknown> = {}

  if (existingRaw && typeof existingRaw === 'object' && !Array.isArray(existingRaw)) {
    for (const [name, entry] of Object.entries(existingRaw)) {
      if (!(name in managed)) {
        merged[name] = entry
      }
    }
  }

  for (const [name, entry] of Object.entries(managed)) {
    merged[name] = entry
  }

  return merged
}

function toCopilotServerEntries(servers: Record<string, McpServer>): {
  entries: Record<string, unknown>
  placeholderIds: Set<string>
} {
  const entries: Record<string, unknown> = {}
  const placeholderIds = new Set<string>()

  for (const [name, server] of Object.entries(servers)) {
    if (server.url) {
      const entry: Record<string, unknown> = {
        type: 'sse',
        url: server.url,
      }
      if (server.headers) entry.headers = server.headers
      entries[name] = entry
      continue
    }

    const entry: Record<string, unknown> = {
      type: 'stdio',
      command: server.command,
      args: server.args,
    }
    if (server.env) {
      entry.env = server.env
      for (const value of Object.values(server.env)) {
        for (const match of value.matchAll(/\$\{(\w+)\}/g)) {
          if (match[1]) placeholderIds.add(match[1])
        }
      }
    }
    entries[name] = entry
  }

  return { entries, placeholderIds }
}

function toCopilotPromptInputs(placeholderIds: Set<string>): CopilotPromptInput[] {
  return [...placeholderIds]
    .sort((a, b) => a.localeCompare(b))
    .map((id) => ({
      type: 'promptString',
      id,
      description: id,
      password: true,
    }))
}

function toCopilotVSCodeMcp(servers: Record<string, McpServer>): Record<string, unknown> {
  const { entries, placeholderIds } = toCopilotServerEntries(servers)
  const output: Record<string, unknown> = { servers: entries }
  if (placeholderIds.size > 0) {
    output.inputs = toCopilotPromptInputs(placeholderIds)
  }
  return output
}

function toCopilotCliMcp(servers: Record<string, McpServer>): Record<string, unknown> {
  const { entries } = toCopilotServerEntries(servers)
  return { mcpServers: entries }
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

function resolveToolRoot(toolId: ToolId, targetDir: string, setupScope: SetupScope): string {
  if (setupScope === 'global') return targetDir

  switch (toolId) {
    case 'opencode':
      return path.join(targetDir, '.opencode')
    case 'claude-code':
      return path.join(targetDir, '.claude')
    case 'gemini':
      return path.join(targetDir, '.gemini')
    case 'codex':
      return path.join(targetDir, '.codex')
    default:
      return targetDir
  }
}

function resolveHomeDir(homeDir?: string): string {
  return homeDir ?? homedir()
}

function mergeJsonFile(pathname: string, patch: Record<string, unknown>): Record<string, unknown> {
  let existing: Record<string, unknown> = {}
  if (fileExists(pathname)) {
    try {
      existing = JSON.parse(readFile(pathname)) as Record<string, unknown>
    } catch {
      existing = {}
    }
  }

  const merged = { ...existing, ...patch }
  writeFile(pathname, `${JSON.stringify(merged, null, 2)}\n`)
  return merged
}

function copilotProbePasses(homeDir: string): boolean {
  if (fileExists(path.join(homeDir, '.copilot'))) {
    return true
  }

  try {
    execFileSync('copilot', ['--version'], { stdio: 'ignore' })
    return true
  } catch {
    return false
  }
}

function codexSectionPrefix(sectionName: string): string {
  return `[mcp_servers.${sectionName}]`
}

function stripManagedCodexSections(existing: string, managedNames: Set<string>): string {
  const lines = existing.split('\n')
  const kept: string[] = []
  let skippingManagedSection = false

  for (const line of lines) {
    const match = line.match(/^\[mcp_servers\.([^\].]+)(?:\.env)?\]$/)
    if (match?.[1]) {
      skippingManagedSection = managedNames.has(match[1])
      if (!skippingManagedSection) kept.push(line)
      continue
    }

    if (line.match(/^\[[^\]]+\]$/)) {
      skippingManagedSection = false
      kept.push(line)
      continue
    }

    if (!skippingManagedSection) {
      kept.push(line)
    }
  }

  return kept.join('\n').trim()
}

function renderCodexServer(name: string, server: McpServer): string {
  const lines = [codexSectionPrefix(name)]

  if (server.command) lines.push(`command = ${JSON.stringify(server.command)}`)
  if (server.args && server.args.length > 0) {
    lines.push(`args = [${server.args.map((arg) => JSON.stringify(arg)).join(', ')}]`)
  }
  if (server.env && Object.keys(server.env).length > 0) {
    lines.push('', `[mcp_servers.${name}.env]`)
    for (const [key, value] of Object.entries(server.env).sort(([a], [b]) => a.localeCompare(b))) {
      lines.push(`${key} = ${JSON.stringify(value)}`)
    }
  }

  return lines.join('\n')
}

function mergeCodexToml(existing: string, servers: Record<string, McpServer>): string {
  const managedNames = new Set(Object.keys(servers))
  const preserved = stripManagedCodexSections(existing, managedNames)
  const renderedManaged = Object.entries(servers)
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([name, server]) => renderCodexServer(name, server))
    .join('\n\n')

  return `${[preserved, renderedManaged].filter(Boolean).join('\n\n').trimEnd()}\n`
}

function upsertFileRecord(fileRecords: FileRecord[], record: FileRecord): void {
  const index = fileRecords.findIndex((existing) => existing.path === record.path)
  if (index >= 0) {
    fileRecords[index] = record
    return
  }
  fileRecords.push(record)
}

export interface CompileMcpOptions {
  canonicalDir: string
  toolTargetDir: string
  toolId: ToolId
  fileRecords: FileRecord[]
  setupScope?: SetupScope
  homeDir?: string
  localSecrets?: boolean
}

export async function compileMcp(opts: CompileMcpOptions): Promise<void> {
  const catalog = readCanonicalMcp(opts.canonicalDir)
  if (!catalog) return

  const enabledServers = getEnabledServers(catalog)
  if (Object.keys(enabledServers).length === 0) return

  const setupScope = opts.setupScope ?? 'project'
  const homeDir = resolveHomeDir(opts.homeDir)

  switch (opts.toolId) {
    case 'opencode': {
      const toolRoot = resolveToolRoot('opencode', opts.toolTargetDir, setupScope)
      ensureDir(toolRoot)
      const configPath = path.join(toolRoot, 'opencode.jsonc')
      const ocMcpContent = toOpenCodeJsonc(catalog.servers)

      let existingConfig: Record<string, unknown> = {}
      if (fileExists(configPath)) {
        try {
          existingConfig = JSON.parse(stripJsonComments(readFile(configPath))) as Record<string, unknown>
        } catch {
          existingConfig = {}
        }
      }

      const merged = {
        ...existingConfig,
        $schema: 'https://opencode.ai/config.json',
        mcp: mergeOpenCodeMcp(existingConfig.mcp, ocMcpContent),
      }

      writeFile(configPath, `${JSON.stringify(merged, null, 2)}\n`)
      upsertFileRecord(opts.fileRecords, {
        path: path.relative(opts.toolTargetDir, configPath),
        hash: fileHash(configPath),
        source: 'compiled:mcp:opencode',
        owner: 'library',
      })
      break
    }

    case 'claude-code': {
      if (opts.localSecrets === true) {
        const settingsPath =
          setupScope === 'global'
            ? path.join(homeDir, '.claude', 'settings.local.json')
            : path.join(resolveToolRoot('claude-code', opts.toolTargetDir, setupScope), 'settings.local.json')
        ensureDir(path.dirname(settingsPath))
        mergeJsonFile(settingsPath, { mcpServers: toClaudeCodeMcpInner(enabledServers) })
        upsertFileRecord(opts.fileRecords, {
          path: path.relative(opts.toolTargetDir, settingsPath),
          hash: fileHash(settingsPath),
          source: 'compiled:mcp:claude-local',
          owner: 'library',
        })
        break
      }

      if (setupScope === 'global') {
        break
      }

      const mcpPath = path.join(opts.toolTargetDir, '.mcp.json')
      const content = toMcpJson(enabledServers)
      writeFile(mcpPath, `${JSON.stringify(content, null, 2)}\n`)
      upsertFileRecord(opts.fileRecords, {
        path: '.mcp.json',
        hash: fileHash(mcpPath),
        source: 'compiled:mcp',
        owner: 'library',
      })
      break
    }

    case 'copilot': {
      if (setupScope !== 'global') {
        const vscodeMcpPath = path.join(opts.toolTargetDir, '.vscode', 'mcp.json')
        ensureDir(path.join(opts.toolTargetDir, '.vscode'))
        const content = toCopilotVSCodeMcp(enabledServers)
        writeFile(vscodeMcpPath, `${JSON.stringify(content, null, 2)}\n`)
        upsertFileRecord(opts.fileRecords, {
          path: '.vscode/mcp.json',
          hash: fileHash(vscodeMcpPath),
          source: 'compiled:mcp:copilot',
          owner: 'library',
        })
      }

      if (!copilotProbePasses(homeDir)) {
        break
      }

      const cliRoot = path.join(homeDir, '.copilot')
      ensureDir(cliRoot)
      const cliPath = path.join(cliRoot, 'mcp-config.json')
      mergeJsonFile(cliPath, toCopilotCliMcp(enabledServers))
      upsertFileRecord(opts.fileRecords, {
        path: path.relative(opts.toolTargetDir, cliPath),
        hash: fileHash(cliPath),
        source: 'compiled:mcp:copilot-cli',
        owner: 'library',
      })
      break
    }

    case 'gemini': {
      const toolRoot = resolveToolRoot('gemini', opts.toolTargetDir, setupScope)
      ensureDir(toolRoot)
      const settingsPath = path.join(toolRoot, 'settings.json')
      const content = toGeminiSettings(enabledServers)
      writeFile(settingsPath, `${JSON.stringify(content, null, 2)}\n`)
      upsertFileRecord(opts.fileRecords, {
        path: path.relative(opts.toolTargetDir, settingsPath),
        hash: fileHash(settingsPath),
        source: 'compiled:mcp:gemini',
        owner: 'library',
      })
      break
    }

    case 'codex': {
      const toolRoot = resolveToolRoot('codex', opts.toolTargetDir, setupScope)
      ensureDir(toolRoot)
      const configPath = path.join(toolRoot, 'config.toml')
      const existing = fileExists(configPath) ? readFile(configPath) : ''
      const content = mergeCodexToml(existing, enabledServers)
      writeFile(configPath, content)
      upsertFileRecord(opts.fileRecords, {
        path: path.relative(opts.toolTargetDir, configPath),
        hash: fileHash(configPath),
        source: 'compiled:mcp:codex',
        owner: 'library',
      })
      break
    }
  }
}
