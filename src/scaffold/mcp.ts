import path from 'node:path'
import { ensureDir, fileExists, fileHash, readFile, writeFile } from '../utils/files.js'
import { applyStrategy } from '../utils/conflict-strategy.js'
import type { FileRecord, ConflictStrategy } from '../types.js'

interface McpServer {
  enabled?: boolean
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

  let content = readFile(catalogPath)

  if (opts.cliTools && opts.cliTools.length > 0) {
    const catalog = JSON.parse(content) as McpCatalog
    for (const toolName of opts.cliTools) {
      if (catalog.servers[toolName]) {
        catalog.servers[toolName].enabled = true
      }
    }
    content = JSON.stringify(catalog, null, 2)
  }

  writeFile(dest, content)
  fileRecords.push({
    path: relPath,
    hash: fileHash(dest),
    source: 'mcp/catalog.json',
  })
}
