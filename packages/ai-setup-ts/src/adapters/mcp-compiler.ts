import * as childProcess from 'node:child_process'
import { homedir } from 'node:os'
import path from 'node:path'
import { parse as parseToml, stringify as stringifyToml } from 'smol-toml'
import type { FileRecord, SetupScope, ToolId } from '../types.js'
import { copyFile, ensureDir, fileExists, fileHash, readFile, writeFile } from '../utils/files.js'
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

export const mcpCompilerInternals = {
  execFileSync(file: string, args: readonly string[], options?: childProcess.ExecFileSyncOptions) {
    return childProcess.execFileSync(file, args, options)
  },
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

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
}

function deepMerge(target: Record<string, unknown>, source: Record<string, unknown>): Record<string, unknown> {
  const result = { ...target }

  for (const [key, value] of Object.entries(source)) {
    if (isPlainObject(value) && isPlainObject(result[key])) {
      result[key] = deepMerge(result[key], value)
    } else {
      result[key] = value
    }
  }

  return result
}

function sortDeep(value: unknown): unknown {
  if (Array.isArray(value)) {
    return value.map((item) => sortDeep(item))
  }

  if (!isPlainObject(value)) {
    return value
  }

  const sorted: Record<string, unknown> = {}
  for (const key of Object.keys(value).sort((a, b) => a.localeCompare(b))) {
    sorted[key] = sortDeep(value[key])
  }
  return sorted
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error)
}

function readJsonConfigObject(pathname: string): Record<string, unknown> {
  if (!fileExists(pathname)) return {}

  let parsed: unknown
  try {
    parsed = JSON.parse(stripJsonComments(readFile(pathname))) as unknown
  } catch (error) {
    throw new Error(`Failed to parse JSON config ${pathname}: ${errorMessage(error)}`)
  }

  if (!isPlainObject(parsed)) {
    throw new Error(`Failed to parse JSON config ${pathname}: expected a JSON object`)
  }

  return parsed
}

function backupJsonFile(pathname: string): void {
  if (!fileExists(pathname)) return

  const backupPath = `${pathname}.bak`
  if (fileExists(backupPath)) return

  copyFile(pathname, backupPath)
}

function mergeJsonFile(pathname: string, patch: Record<string, unknown>): Record<string, unknown> {
  if (fileExists(pathname)) {
    const existing = readJsonConfigObject(pathname)
    const merged = deepMerge(existing, patch)
    backupJsonFile(pathname)
    writeFile(pathname, `${JSON.stringify(sortDeep(merged), null, 2)}\n`)
    return merged
  }

  const merged = deepMerge({}, patch)
  writeFile(pathname, `${JSON.stringify(sortDeep(merged), null, 2)}\n`)
  return merged
}

function backupTomlFile(pathname: string): void {
  if (!fileExists(pathname)) return

  const backupPath = `${pathname}.bak`
  if (fileExists(backupPath)) return

  writeFile(backupPath, readFile(pathname))
}

function readTomlObject(pathname: string): Record<string, unknown> {
  if (!fileExists(pathname)) return {}

  try {
    const parsed = parseToml(readFile(pathname)) as unknown
    return isPlainObject(parsed) ? parsed : {}
  } catch (error) {
    throw new Error(`Failed to parse TOML config ${pathname}: ${errorMessage(error)}`)
  }
}

function mergeCodexTomlFile(pathname: string, servers: Record<string, McpServer>): Record<string, unknown> {
  const existing = readTomlObject(pathname)
  if (fileExists(pathname)) {
    backupTomlFile(pathname)
  }

  const existingMcpServers = isPlainObject(existing.mcp_servers) ? existing.mcp_servers : {}
  const managedNames = new Set(Object.keys(servers))
  const preservedServers: Record<string, unknown> = {}

  for (const [name, config] of Object.entries(existingMcpServers)) {
    if (!managedNames.has(name)) {
      preservedServers[name] = config
    }
  }

  const patch = toCodexTomlPatch(servers)
  const managedServers = isPlainObject(patch.mcp_servers) ? patch.mcp_servers : {}
  const merged: Record<string, unknown> = {
    ...existing,
    mcp_servers: {
      ...preservedServers,
      ...managedServers,
    },
  }

  const content = stringifyToml(sortDeep(merged) as Record<string, unknown>)
  writeFile(pathname, content.endsWith('\n') ? content : `${content}\n`)
  return merged
}

function claudeCliAvailable(): boolean {
  try {
    mcpCompilerInternals.execFileSync('claude', ['--version'], { stdio: 'ignore' })
    return true
  } catch {
    return false
  }
}

function cliWorkingDirectory(setupScope: SetupScope, toolTargetDir: string): string | undefined {
  return setupScope === 'global' ? undefined : toolTargetDir
}

function reconcileClaudeMcp(
  servers: Record<string, unknown>,
  setupScope: SetupScope,
  toolTargetDir: string,
): boolean {
  if (!claudeCliAvailable()) return false

  try {
    const cwd = cliWorkingDirectory(setupScope, toolTargetDir)
    const scopeFlag = setupScope === 'global' ? 'user' : 'project'
    const options: childProcess.ExecFileSyncOptions = cwd ? { stdio: 'pipe', cwd } : { stdio: 'pipe' }

    for (const [name, config] of Object.entries(servers).sort(([a], [b]) => a.localeCompare(b))) {
      try {
        mcpCompilerInternals.execFileSync('claude', ['mcp', 'get', name], options)
        continue
      } catch {
        // Not registered yet; add below.
      }

      mcpCompilerInternals.execFileSync('claude', ['mcp', 'add-json', name, JSON.stringify(config), '-s', scopeFlag], {
        stdio: 'pipe',
        ...(cwd ? { cwd } : {}),
      })
    }
    return true
  } catch {
    return false
  }
}

function codexCliAvailable(): boolean {
  try {
    mcpCompilerInternals.execFileSync('codex', ['--version'], { stdio: 'ignore' })
    return true
  } catch {
    return false
  }
}

function copilotProbePasses(homeDir: string): boolean {
  if (fileExists(path.join(homeDir, '.copilot'))) {
    return true
  }

  try {
    mcpCompilerInternals.execFileSync('copilot', ['--version'], { stdio: 'ignore' })
    return true
  } catch {
    return false
  }
}

function toCodexTomlPatch(servers: Record<string, McpServer>): Record<string, unknown> {
  const mcpServers: Record<string, unknown> = {}

  for (const [name, server] of Object.entries(servers).sort(([a], [b]) => a.localeCompare(b))) {
    if (server.url || !server.command) continue

    const entry: Record<string, unknown> = {}
    entry.command = server.command
    if (server.args && server.args.length > 0) entry.args = server.args
    if (server.env && Object.keys(server.env).length > 0) entry.env = sortDeep(server.env)
    mcpServers[name] = entry
  }

  return { mcp_servers: mcpServers }
}

function reconcileCodexMcp(servers: Record<string, McpServer>, toolTargetDir: string): boolean {
  if (!codexCliAvailable()) return false

  let success = false
  for (const [name, server] of Object.entries(servers).sort(([a], [b]) => a.localeCompare(b))) {
    if (server.url || !server.command) continue

    const args = ['mcp', 'add', name]
    for (const [key, value] of Object.entries(server.env ?? {}).sort(([a], [b]) => a.localeCompare(b))) {
      args.push('--env', `${key}=${value}`)
    }
    args.push('--', server.command, ...(server.args ?? []))

    try {
      mcpCompilerInternals.execFileSync('codex', args, { stdio: 'pipe', cwd: toolTargetDir })
      success = true
    } catch {
      return false
    }
  }
  return success
}

export function driveClaudeMcpViaCli(
  canonicalDir: string,
  toolTargetDir: string,
  setupScope: SetupScope = 'project',
): boolean {
  const catalog = readCanonicalMcp(canonicalDir)
  if (!catalog) return false
  const enabledServers = getEnabledServers(catalog)
  if (Object.keys(enabledServers).length === 0) return false
  return reconcileClaudeMcp(toClaudeCodeMcpInner(enabledServers), setupScope, toolTargetDir)
}

export function driveCodexMcpViaCli(canonicalDir: string, toolTargetDir: string): boolean {
  const catalog = readCanonicalMcp(canonicalDir)
  if (!catalog) return false
  const enabledServers = getEnabledServers(catalog)
  if (Object.keys(enabledServers).length === 0) return false
  return reconcileCodexMcp(enabledServers, toolTargetDir)
}

function isDriveCliEnabled(opts: { driveCLI?: boolean; driveCli?: boolean }): boolean {
  return opts.driveCLI === true || opts.driveCli === true
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
  driveCLI?: boolean
  driveCli?: boolean
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
        existingConfig = readJsonConfigObject(configPath)
        backupJsonFile(configPath)
      }

      const merged = {
        ...existingConfig,
        $schema: 'https://opencode.ai/config.json',
        mcp: mergeOpenCodeMcp(existingConfig.mcp, ocMcpContent),
      }

      writeFile(configPath, `${JSON.stringify(sortDeep(merged), null, 2)}\n`)
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

      if (isDriveCliEnabled(opts)) {
        driveClaudeMcpViaCli(opts.canonicalDir, opts.toolTargetDir, setupScope)
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
      if (isDriveCliEnabled(opts)) {
        driveCodexMcpViaCli(opts.canonicalDir, opts.toolTargetDir)
      }
      mergeCodexTomlFile(configPath, enabledServers)
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
