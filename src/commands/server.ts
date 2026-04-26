import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import * as p from '@clack/prompts'
import { Client } from '@modelcontextprotocol/sdk/client/index.js'
import { StdioClientTransport } from '@modelcontextprotocol/sdk/client/stdio.js'
import type { Command } from 'commander'
import pc from 'picocolors'
import { compileMcp } from '../adapters/mcp-compiler.js'
import { AdapterRegistry } from '../adapters/registry.js'
import { Errors } from '../errors/index.js'
import { scaffoldMcp } from '../scaffold/mcp.js'
import { scaffoldOrchestration } from '../scaffold/orchestration.js'
import { createStore, writeStore } from '../store/index.js'
import type { FileRecord, ToolId } from '../types.js'
import { fileExists, readFile, resolveLibraryDir } from '../utils/files.js'
import { stripJsonComments } from '../utils/jsonc.js'
import { showSummaryBox } from '../utils/ui.js'

const libraryDir = resolveLibraryDir(dirname(fileURLToPath(import.meta.url)))

interface CatalogServer {
  description?: string
  command?: string
  args?: string[]
  env?: Record<string, string>
  url?: string
  headers?: Record<string, string>
  tools?: string[]
  enabled?: boolean
  requiresInstall?: boolean
  installHint?: string
  type?: string
}

interface Catalog {
  servers: Record<string, CatalogServer>
}

type CheckStatus = 'pass' | 'fail' | 'skip'

interface CheckResult {
  name: string
  status: CheckStatus
  message: string
  remediation?: string
}

interface ServerHealthReport {
  server: string
  overall: 'healthy' | 'unhealthy' | 'partial'
  checks: CheckResult[]
}

const PER_TOOL_MCP_CONFIG: Record<ToolId, string | null> = {
  opencode: 'opencode.jsonc',
  'claude-code': '.mcp.json',
  copilot: '.vscode/mcp.json',
  gemini: '.gemini/settings.json',
  codex: null,
  pi: '.pi/settings.json',
}

export function registerServer(program: Command): void {
  const server = program
    .command('server')
    .description('Manage MCP servers enabled in your setup')

  server
    .command('add <name>')
    .description('Enable an MCP server in this project')
    .action(async (name: string) => {
      await runServerAdd(name)
    })

  server
    .command('remove <name>')
    .description('Disable an MCP server in this project')
    .action(async (name: string) => {
      await runServerRemove(name)
    })

  server
    .command('list')
    .description('List MCP servers available in the catalog and which are enabled')
    .option('--json', 'Output as JSON')
    .action((opts: { json?: boolean }) => {
      runServerList(opts)
    })

  server
    .command('doctor [name]')
    .description('Validate that enabled MCP servers are healthy (L1 config checks + L3 stdio handshake)')
    .option('--json', 'Output as JSON')
    .option('--timeout <ms>', 'Stdio handshake timeout in ms', '5000')
    .action(async (name: string | undefined, opts: { json?: boolean; timeout?: string }) => {
      await runServerDoctor(name, opts)
    })
}

async function runServerAdd(name: string): Promise<void> {
  const targetDir = process.cwd()
  const configPath = join(targetDir, '.ai-setup.json')
  if (!fileExists(configPath)) throw Errors.manifestNotFound(targetDir)

  const catalog = readCatalog()
  const entry = catalog.servers[name]
  if (!entry) {
    throw Errors.invalidInput(`unknown MCP server: ${name}`, {
      available: Object.keys(catalog.servers),
    })
  }

  const db = await createStore(targetDir)
  const data = db.data
  const current = new Set(data.config.enableServers ?? [])

  if (current.has(name)) {
    p.log.info(`${name} is already enabled.`)
    return
  }

  p.intro(pc.bold(`Enabling MCP server: ${name}`))

  if (entry.requiresInstall && entry.installHint) {
    p.log.warn(`${name} requires installation: ${entry.installHint}`)
  }

  current.add(name)
  const enableServers = [...current].sort()
  data.config.enableServers = enableServers

  const spinner = p.spinner()
  spinner.start('Re-running scaffold and per-tool compile')
  try {
    await rerunPipeline(targetDir, data.config.tools as ToolId[], enableServers)
    await writeStore(targetDir, data)
    spinner.stop(`Enabled ${name}`)
  } catch (error) {
    spinner.stop(`Failed to enable ${name}`)
    throw error
  }

  // Run a quick post-add doctor for this server
  const report = await runHealthChecks(targetDir, name, catalog, data.config.tools as ToolId[], 5000)
  renderReport([report], false)

  p.outro(pc.green(`✓ ${name} is enabled`))
}

async function runServerRemove(name: string): Promise<void> {
  const targetDir = process.cwd()
  const configPath = join(targetDir, '.ai-setup.json')
  if (!fileExists(configPath)) throw Errors.manifestNotFound(targetDir)

  const catalog = readCatalog()
  if (!catalog.servers[name]) {
    throw Errors.invalidInput(`unknown MCP server: ${name}`, {
      available: Object.keys(catalog.servers),
    })
  }

  const db = await createStore(targetDir)
  const data = db.data
  const current = new Set(data.config.enableServers ?? [])

  if (!current.has(name)) {
    p.log.info(`${name} is not currently enabled.`)
    return
  }

  p.intro(pc.bold(`Disabling MCP server: ${name}`))

  current.delete(name)
  const enableServers = [...current].sort()
  data.config.enableServers = enableServers

  const spinner = p.spinner()
  spinner.start('Re-running scaffold and per-tool compile')
  try {
    await rerunPipeline(targetDir, data.config.tools as ToolId[], enableServers)
    await writeStore(targetDir, data)
    spinner.stop(`Disabled ${name}`)
  } catch (error) {
    spinner.stop(`Failed to disable ${name}`)
    throw error
  }

  if (name === 'orchestrator') {
    const orchDir = join(targetDir, '.ai', 'orchestration')
    if (fileExists(orchDir)) {
      p.log.warn(`${orchDir} was left on disk. Delete it manually if you want a clean state.`)
    }
  }

  p.outro(pc.green(`✓ ${name} is disabled`))
}

function runServerList(opts: { json?: boolean }): void {
  const catalog = readCatalog()
  const targetDir = process.cwd()

  let enabledFromStore: string[] = []
  const configPath = join(targetDir, '.ai-setup.json')
  if (fileExists(configPath)) {
    try {
      const raw = JSON.parse(readFile(configPath)) as { config?: { enableServers?: string[] } }
      enabledFromStore = raw.config?.enableServers ?? []
    } catch {
      enabledFromStore = []
    }
  }

  const enabledSet = new Set(enabledFromStore)

  const rows = Object.entries(catalog.servers).map(([name, entry]) => ({
    name,
    enabled: enabledSet.has(name) || entry.enabled === true,
    description: entry.description ?? '',
    requiresInstall: entry.requiresInstall === true,
    installHint: entry.installHint ?? null,
    command: entry.command ?? (entry.url ? `remote: ${entry.url}` : null),
  }))

  if (opts.json) {
    console.log(JSON.stringify(rows, null, 2))
    return
  }

  p.intro(pc.bold('MCP Servers'))
  for (const row of rows) {
    const marker = row.enabled ? pc.green('●') : pc.dim('○')
    const nameCol = row.enabled ? pc.bold(row.name) : pc.dim(row.name)
    const extras: string[] = []
    if (row.requiresInstall) extras.push(pc.yellow('requires install'))
    const extraStr = extras.length > 0 ? ` [${extras.join(', ')}]` : ''
    p.log.message(`  ${marker} ${nameCol} — ${pc.dim(row.description)}${extraStr}`)
  }
  p.outro(pc.dim(`${rows.filter((r) => r.enabled).length} of ${rows.length} enabled`))
}

async function runServerDoctor(
  name: string | undefined,
  opts: { json?: boolean; timeout?: string },
): Promise<void> {
  const targetDir = process.cwd()
  const configPath = join(targetDir, '.ai-setup.json')
  if (!fileExists(configPath)) throw Errors.manifestNotFound(targetDir)

  const catalog = readCatalog()
  const db = await createStore(targetDir)
  const data = db.data
  const tools = data.config.tools as ToolId[]
  const enabled = data.config.enableServers ?? []

  const targets = name ? [name] : enabled
  if (targets.length === 0) {
    if (opts.json) {
      console.log(JSON.stringify({ reports: [] }, null, 2))
      return
    }
    p.log.info('No enabled MCP servers to check.')
    return
  }

  const timeoutMs = Number.parseInt(opts.timeout ?? '5000', 10)

  if (!opts.json) p.intro(pc.bold('ai-setup server doctor'))

  const reports: ServerHealthReport[] = []
  for (const target of targets) {
    if (!catalog.servers[target]) {
      reports.push({
        server: target,
        overall: 'unhealthy',
        checks: [
          {
            name: 'catalog',
            status: 'fail',
            message: `server '${target}' not found in library/mcp/catalog.json`,
            remediation: `Run 'ai-setup server list' to see available servers`,
          },
        ],
      })
      continue
    }
    const report = await runHealthChecks(targetDir, target, catalog, tools, timeoutMs)
    reports.push(report)
  }

  if (opts.json) {
    console.log(JSON.stringify({ reports }, null, 2))
    const hasUnhealthy = reports.some((r) => r.overall === 'unhealthy')
    if (hasUnhealthy) throw Errors.unknown(`doctor found issues in ${reports.filter((r) => r.overall === 'unhealthy').length} server(s)`)
    return
  }

  renderReport(reports, true)

  const hasUnhealthy = reports.some((r) => r.overall === 'unhealthy')
  if (hasUnhealthy) {
    p.outro(pc.yellow('⚠ Some servers are unhealthy'))
    throw Errors.unknown(`doctor found issues in ${reports.filter((r) => r.overall === 'unhealthy').length} server(s)`)
  }
  p.outro(pc.green('✓ All enabled servers healthy'))
}

async function rerunPipeline(
  targetDir: string,
  tools: ToolId[],
  enableServers: string[],
): Promise<void> {
  const fileRecords: FileRecord[] = []
  const strategy = 'backup-and-replace'
  const perFileOverrides = new Map()

  await scaffoldMcp({
    targetDir,
    libraryDir,
    fileRecords,
    strategy,
    perFileOverrides,
    enableServers,
  })

  if (enableServers.includes('orchestrator')) {
    await scaffoldOrchestration({
      targetDir,
      libraryDir,
      fileRecords,
      strategy,
      perFileOverrides,
    })
  }

  const registry = new AdapterRegistry()
  for (const toolId of tools) {
    const adapter = registry.get(toolId)
    if (!adapter) continue
    await adapter.install({
      targetDir,
      libraryDir,
      fileRecords,
      force: true,
      strategy,
      enableServers,
    })
    await compileMcp({
      canonicalDir: targetDir,
      toolTargetDir: targetDir,
      toolId,
      fileRecords,
      setupScope: 'project',
    })
  }
}

export async function runHealthChecks(
  targetDir: string,
  name: string,
  catalog: Catalog,
  tools: ToolId[],
  timeoutMs: number,
): Promise<ServerHealthReport> {
  const entry = catalog.servers[name]
  const checks: CheckResult[] = []

  // L1.1: canonical .ai/mcp.json has the server enabled
  const canonicalPath = join(targetDir, '.ai', 'mcp.json')
  if (!fileExists(canonicalPath)) {
    checks.push({
      name: 'canonical mcp.json',
      status: 'fail',
      message: '.ai/mcp.json is missing',
      remediation: `Run 'ai-setup compile --force' to regenerate it`,
    })
    return finalizeReport(name, checks)
  }

  let canonical: Catalog
  try {
    canonical = JSON.parse(readFile(canonicalPath)) as Catalog
  } catch (error) {
    checks.push({
      name: 'canonical mcp.json',
      status: 'fail',
      message: `.ai/mcp.json is not valid JSON: ${error instanceof Error ? error.message : String(error)}`,
    })
    return finalizeReport(name, checks)
  }

  const canonicalEntry = canonical.servers?.[name]
  if (!canonicalEntry) {
    checks.push({
      name: 'canonical mcp.json entry',
      status: 'fail',
      message: `'${name}' missing from .ai/mcp.json`,
      remediation: `Run 'ai-setup server add ${name}'`,
    })
    return finalizeReport(name, checks)
  }
  if (canonicalEntry.enabled !== true) {
    checks.push({
      name: 'canonical mcp.json entry',
      status: 'fail',
      message: `'${name}' is present but not enabled in .ai/mcp.json`,
      remediation: `Run 'ai-setup server add ${name}'`,
    })
    return finalizeReport(name, checks)
  }
  checks.push({
    name: 'canonical mcp.json entry',
    status: 'pass',
    message: `'${name}' enabled in .ai/mcp.json`,
  })

  // L1.2: per-tool compiled config has the entry
  for (const tool of tools) {
    const relPath = PER_TOOL_MCP_CONFIG[tool]
    if (!relPath) {
      checks.push({
        name: `${tool} mcp config`,
        status: 'skip',
        message: `${tool} has no project-local MCP config file (managed globally)`,
      })
      continue
    }
    const abs = join(targetDir, relPath)
    if (!fileExists(abs)) {
      checks.push({
        name: `${tool} mcp config`,
        status: 'fail',
        message: `${relPath} is missing`,
        remediation: `Run 'ai-setup compile --force'`,
      })
      continue
    }
    const raw = readFile(abs)
    try {
      const parsed = JSON.parse(stripJsonComments(raw)) as {
        mcp?: Record<string, unknown>
        mcpServers?: Record<string, unknown>
        servers?: Record<string, unknown>
      }
      const serverMap = parsed.mcp ?? parsed.mcpServers ?? parsed.servers ?? {}
      if (serverMap[name]) {
        checks.push({
          name: `${tool} mcp config`,
          status: 'pass',
          message: `${relPath} contains '${name}'`,
        })
      } else {
        checks.push({
          name: `${tool} mcp config`,
          status: 'fail',
          message: `${relPath} exists but does not contain '${name}'`,
          remediation: `Run 'ai-setup compile --force'`,
        })
      }
    } catch (error) {
      checks.push({
        name: `${tool} mcp config`,
        status: 'fail',
        message: `${relPath} is not valid JSON: ${error instanceof Error ? error.message : String(error)}`,
      })
    }
  }

  // L1.3: orchestrator-specific checks
  if (name === 'orchestrator') {
    const chainsDir = join(targetDir, '.ai', 'orchestration', 'chains')
    if (!fileExists(chainsDir)) {
      checks.push({
        name: 'orchestration chains',
        status: 'fail',
        message: '.ai/orchestration/chains/ is missing',
        remediation: `Run 'ai-setup server add orchestrator'`,
      })
    } else {
      checks.push({
        name: 'orchestration chains',
        status: 'pass',
        message: '.ai/orchestration/chains/ present',
      })
    }

    const registry = new AdapterRegistry()
    for (const tool of tools) {
      const adapter = registry.get(tool)
      if (!adapter) continue
      const agentPath = getOrchestratorAgentPathForTool(tool, targetDir)
      if (!agentPath) {
        checks.push({
          name: `${tool} orchestrator agent`,
          status: 'skip',
          message: `${tool} has no agent-file path convention`,
        })
        continue
      }
      if (fileExists(agentPath)) {
        checks.push({
          name: `${tool} orchestrator agent`,
          status: 'pass',
          message: agentPath,
        })
      } else {
        checks.push({
          name: `${tool} orchestrator agent`,
          status: 'fail',
          message: `${agentPath} missing`,
          remediation: `Run 'ai-setup compile --force'`,
        })
      }
    }
  }

  // L3: stdio handshake (use canonical entry — reflects user overrides)
  const effectiveCommand = canonicalEntry.command
  const effectiveArgs = canonicalEntry.args
  const effectiveTools = canonicalEntry.tools ?? entry?.tools
  if (effectiveCommand && effectiveArgs) {
    try {
      const handshake = await runStdioHandshake(
        {
          command: effectiveCommand,
          args: effectiveArgs,
          ...(canonicalEntry.env ? { env: canonicalEntry.env } : {}),
        },
        targetDir,
        timeoutMs,
      )
      if (effectiveTools && effectiveTools.length > 0) {
        const expected = new Set(effectiveTools)
        const actual = new Set(handshake.tools)
        const missing = [...expected].filter((t) => !actual.has(t))
        if (missing.length === 0) {
          checks.push({
            name: 'stdio handshake',
            status: 'pass',
            message: `server registered ${handshake.tools.length} tools (${[...expected].length} expected, all present)`,
          })
        } else {
          checks.push({
            name: 'stdio handshake',
            status: 'fail',
            message: `server is missing tools: ${missing.join(', ')}`,
            remediation: `Check the server's registerTool calls and rebuild`,
          })
        }
      } else {
        checks.push({
          name: 'stdio handshake',
          status: 'pass',
          message: `server registered ${handshake.tools.length} tools`,
        })
      }
    } catch (error) {
      checks.push({
        name: 'stdio handshake',
        status: 'fail',
        message: `spawn failed: ${error instanceof Error ? error.message : String(error)}`,
        remediation: entry?.requiresInstall && entry?.installHint
          ? entry.installHint
          : `Verify '${effectiveCommand} ${effectiveArgs.join(' ')}' runs from ${targetDir}`,
      })
    }
  } else {
    checks.push({
      name: 'stdio handshake',
      status: 'skip',
      message: 'server has no stdio command (remote or url-based)',
    })
  }

  return finalizeReport(name, checks)
}

async function runStdioHandshake(
  entry: { command: string; args: string[]; env?: Record<string, string> },
  cwd: string,
  timeoutMs: number,
): Promise<{ tools: string[] }> {
  const transport = new StdioClientTransport({
    command: entry.command,
    args: entry.args,
    env: { ...(process.env as Record<string, string>), ...(entry.env ?? {}) },
    cwd,
  })
  const client = new Client({ name: 'ai-setup-doctor', version: '0.2.0' })

  const connectPromise = (async () => {
    await client.connect(transport)
    const result = await client.listTools()
    return { tools: result.tools.map((t) => t.name) }
  })()

  const timeoutPromise = new Promise<never>((_, reject) => {
    setTimeout(() => reject(new Error(`timed out after ${timeoutMs}ms`)), timeoutMs)
  })

  try {
    return await Promise.race([connectPromise, timeoutPromise])
  } finally {
    try {
      await transport.close()
    } catch {
      // ignore
    }
  }
}

function finalizeReport(server: string, checks: CheckResult[]): ServerHealthReport {
  const hasFail = checks.some((c) => c.status === 'fail')
  const hasPass = checks.some((c) => c.status === 'pass')
  const overall: ServerHealthReport['overall'] = hasFail ? 'unhealthy' : hasPass ? 'healthy' : 'partial'
  return { server, overall, checks }
}

function renderReport(reports: ServerHealthReport[], includeSummary: boolean): void {
  for (const report of reports) {
    const emoji = report.overall === 'healthy' ? '✅' : report.overall === 'unhealthy' ? '❌' : '⚠️'
    if (includeSummary) {
      const pass = report.checks.filter((c) => c.status === 'pass').length
      const fail = report.checks.filter((c) => c.status === 'fail').length
      const skip = report.checks.filter((c) => c.status === 'skip').length
      showSummaryBox(`${emoji} ${report.server}`, [
        { label: 'Overall', value: report.overall },
        { label: 'Pass', value: pc.green(`${pass}`) },
        { label: 'Fail', value: fail > 0 ? pc.red(`${fail}`) : pc.dim('0') },
        { label: 'Skip', value: pc.dim(`${skip}`) },
      ])
    }
    for (const check of report.checks) {
      const glyph =
        check.status === 'pass' ? pc.green('✓') : check.status === 'fail' ? pc.red('✗') : pc.dim('—')
      p.log.message(`  ${glyph} ${check.name} — ${check.message}`)
      if (check.status === 'fail' && check.remediation) {
        p.log.message(`    ${pc.dim('→')} ${pc.dim(check.remediation)}`)
      }
    }
  }
}

function readCatalog(): Catalog {
  const catalogPath = join(libraryDir, 'mcp', 'catalog.json')
  if (!fileExists(catalogPath)) {
    throw Errors.missingDependency(`catalog:library/mcp/catalog.json`)
  }
  return JSON.parse(readFile(catalogPath)) as Catalog
}

function getOrchestratorAgentPathForTool(tool: ToolId, targetDir: string): string | null {
  switch (tool) {
    case 'opencode':
      return join(targetDir, '.opencode', 'agents', 'orchestrator.md')
    case 'claude-code':
      return join(targetDir, '.claude', 'agents', 'orchestrator.md')
    case 'gemini':
      return join(targetDir, '.gemini', 'skills', 'orchestrator', 'SKILL.md')
    case 'codex':
      return join(targetDir, '.agents', 'skills', 'orchestrator', 'SKILL.md')
    case 'copilot':
      return join(targetDir, '.github', 'prompts', 'orchestrator.prompt.md')
    case 'pi':
      return null
  }
}
