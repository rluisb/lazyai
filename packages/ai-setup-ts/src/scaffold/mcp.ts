import { execFileSync } from 'node:child_process'
import path from 'node:path'
import type { ConflictStrategy, FileRecord } from '../types.js'
import { applyStrategy } from '../utils/conflict-strategy.js'
import { ensureDir, fileExists, fileHash, readFile, writeFile } from '../utils/files.js'

interface McpServer {
  command?: string
  args?: string[]
  enabled?: boolean
  requiresInstall?: boolean
  installHint?: string
}

interface McpCatalog {
  servers: Record<string, McpServer>
}

export interface ScaffoldMcpOptions {
  targetDir: string
  libraryDir: string
  fileRecords: FileRecord[]
  strategy: ConflictStrategy
  perFileOverrides: Map<string, ConflictStrategy>
  cliTools?: string[]
  /** MCP server names to enable (e.g., ['atlassian']) */
  enableServers?: string[]
}

function resolveOrchestratorPackageDir(libraryDir: string): string | null {
  const repoRoot = path.dirname(libraryDir)
  const packageDir = path.join(repoRoot, 'orchestrator')
  return fileExists(path.join(packageDir, 'package.json')) ? packageDir : null
}

function runOrchestratorBuildStep(packageDir: string, ...args: string[]): boolean {
  try {
    execFileSync('npm', ['--prefix', packageDir, ...args], { stdio: 'ignore', timeout: 120_000 })
    return true
  } catch (error) {
    console.warn(`⚠️  Orchestrator build step "npm ${args.join(' ')}" failed: ${(error as Error).message}`)
    return false
  }
}

function runOrchestratorSmokeTest(nodePath: string, entryPath: string): void {
  try {
    execFileSync(nodePath, [entryPath, 'catalog'], { stdio: 'ignore', timeout: 15_000 })
  } catch (error) {
    throw new Error(`prepare orchestrator MCP: smoke test failed: ${(error as Error).message}`)
  }
}

function prepareManagedOrchestratorServer(libraryDir: string, server: McpServer): McpServer {
  const packageDir = resolveOrchestratorPackageDir(libraryDir)
  if (!packageDir) return server

  const nodePath = process.execPath
  const entryPath = path.join(packageDir, 'dist', 'index.js')

  if (!fileExists(entryPath)) {
    console.log('Building orchestrator package...')
    const installOk = runOrchestratorBuildStep(packageDir, 'install')
    if (installOk) {
      runOrchestratorBuildStep(packageDir, 'run', 'build')
    }
  }

  if (!fileExists(entryPath)) {
    console.warn(`⚠️  Orchestrator build did not produce ${entryPath} — using catalog default command`)
    return server
  }

  if (/^(1|true)$/i.test(process.env.AI_SETUP_ORCHESTRATOR_SMOKE ?? '')) {
    runOrchestratorSmokeTest(nodePath, entryPath)
  }

  return {
    ...server,
    command: nodePath,
    args: [entryPath],
    requiresInstall: false,
  }
}

export async function scaffoldMcp(opts: ScaffoldMcpOptions): Promise<void> {
  const { targetDir, libraryDir, fileRecords, strategy, perFileOverrides } = opts
  const aiDir = path.join(targetDir, '.ai')
  ensureDir(aiDir)

  const catalogPath = path.join(libraryDir, 'mcp', 'catalog.json')
  if (!fileExists(catalogPath)) return

  const dest = path.join(aiDir, 'mcp.json')
  const relPath = path.relative(targetDir, dest)
  const action = applyStrategy(dest, strategy, perFileOverrides, targetDir)

  if (action === 'skip') {
    console.warn(`⚠️  Skipping existing file: ${relPath}`)
    return
  }

  const catalog = JSON.parse(readFile(catalogPath)) as McpCatalog
  const enabledServerNames = new Set<string>()

  if (opts.cliTools && opts.cliTools.length > 0) {
    for (const toolName of opts.cliTools) {
      enabledServerNames.add(toolName)
    }
  }

  if (opts.enableServers && opts.enableServers.length > 0) {
    for (const serverName of opts.enableServers) {
      enabledServerNames.add(serverName)
    }
  }

  if (enabledServerNames.has('orchestrator') && catalog.servers.orchestrator) {
    catalog.servers.orchestrator = prepareManagedOrchestratorServer(libraryDir, catalog.servers.orchestrator)
  }

  for (const serverName of enabledServerNames) {
    if (catalog.servers[serverName]) {
      catalog.servers[serverName].enabled = true
    }
  }

  writeFile(dest, `${JSON.stringify(catalog, null, 2)}\n`)
  fileRecords.push({
    path: relPath,
    hash: fileHash(dest),
    source: 'mcp/catalog.json',
    owner: 'library',
  })
}
