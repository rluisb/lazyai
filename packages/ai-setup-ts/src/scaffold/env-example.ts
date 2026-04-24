import path from 'node:path'
import type { ConflictStrategy, FileRecord } from '../types.js'
import { applyStrategy } from '../utils/conflict-strategy.js'
import { fileExists, fileHash, readFile, writeFile } from '../utils/files.js'

interface McpServer {
  env?: Record<string, string>
  enabled?: boolean
}

interface McpCatalog {
  servers: Record<string, McpServer>
}

export interface ScaffoldEnvExampleOptions {
  targetDir: string
  fileRecords: FileRecord[]
  strategy: ConflictStrategy
  perFileOverrides: Map<string, ConflictStrategy>
}

/**
 * Generates a .env.example file listing required environment variables
 * from enabled MCP servers in the canonical .ai/mcp.json.
 *
 * Only creates the file if at least one enabled server requires env vars.
 * Never includes actual values — only variable names with empty placeholders.
 */
export async function scaffoldEnvExample(opts: ScaffoldEnvExampleOptions): Promise<void> {
  const { targetDir, fileRecords, strategy, perFileOverrides } = opts

  const mcpPath = path.join(targetDir, '.ai', 'mcp.json')
  if (!fileExists(mcpPath)) return

  let catalog: McpCatalog
  try {
    catalog = JSON.parse(readFile(mcpPath)) as McpCatalog
  } catch {
    return
  }

  // Collect env vars from enabled servers
  const envVars: Array<{ name: string; server: string }> = []
  for (const [serverName, server] of Object.entries(catalog.servers)) {
    if (server.enabled === false) continue
    if (!server.env) continue

    for (const key of Object.keys(server.env)) {
      // Extract the actual env var name from patterns like "${VAR_NAME}" or "{env:VAR_NAME}"
      const varName = key
      envVars.push({ name: varName, server: serverName })
    }
  }

  if (envVars.length === 0) return

  const dest = path.join(targetDir, '.env.example')
  const relPath = path.relative(targetDir, dest)
  const action = applyStrategy(dest, strategy, perFileOverrides, targetDir)

  if (action === 'skip') {
    console.warn(`⚠️  Skipping existing file: ${relPath}`)
    return
  }

  // Group by server for readability
  const lines: string[] = [
    '# Environment variables required by enabled MCP servers',
    '# Copy this file to .env and fill in the values',
    '# NEVER commit .env to version control',
    '',
  ]

  const seen = new Set<string>()
  for (const { name, server } of envVars) {
    if (seen.has(name)) continue
    seen.add(name)
    lines.push(`# Required by: ${server}`)
    lines.push(`${name}=`)
    lines.push('')
  }

  writeFile(dest, lines.join('\n'))
  fileRecords.push({
    path: relPath,
    hash: fileHash(dest),
    source: 'generated:env-example',
    owner: 'library',
  })
}
