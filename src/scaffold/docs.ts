import path from 'node:path'
import { ensureDir, copyFile, fileExists, fileHash } from '../utils/files.js'
import { applyStrategy } from '../utils/conflict-strategy.js'
import type { DocsDirId, FileRecord, ConflictStrategy } from '../types.js'

export interface ScaffoldDocsOptions {
  targetDir: string
  libraryDir: string
  docsDirs: DocsDirId[]
  docsAgents: DocsDirId[]
  fileRecords: FileRecord[]
  strategy: ConflictStrategy
  perFileOverrides: Map<string, ConflictStrategy>
}

/**
 * Creates docs directory structure and copies AGENTS.md files.
 *
 * Behavior:
 * - For each dir in `docsDirs`: create `docs/<dir>/`
 * - Special case: `memory` also creates `docs/memory/handoffs/`
 * - For each dir in `docsAgents` (must be subset of docsDirs):
 *   copy `library/docs-agents/<dir>.md` → `docs/<dir>/AGENTS.md`
 * - Record each copied AGENTS.md in `fileRecords`
 * - If `docsDirs` is empty, create no directories
 */
export async function scaffoldDocs(opts: ScaffoldDocsOptions): Promise<void> {
  const { targetDir, libraryDir, docsDirs, docsAgents, fileRecords, strategy, perFileOverrides } = opts

  // Create docs root
  const docsDir = path.join(targetDir, 'docs')

  // 1. Create selected docs directories
  if (docsDirs.length > 0) {
    for (const dir of docsDirs) {
      ensureDir(path.join(docsDir, dir))

      // Special case: memory also needs handoffs subdirectory
      if (dir === 'memory') {
        ensureDir(path.join(docsDir, 'memory', 'handoffs'))
      }
    }
  }

  // 2. Copy AGENTS.md files for selected docs directories
  for (const dir of docsAgents) {
    const src = path.join(libraryDir, 'docs-agents', `${dir}.md`)
    const dest = path.join(docsDir, dir, 'AGENTS.md')

    await copyLibraryFile(src, dest, fileRecords, targetDir, strategy, perFileOverrides)
  }
}

/**
 * Helper: Copy a file from library with conflict resolution and recording.
 */
async function copyLibraryFile(
  src: string,
  dest: string,
  records: FileRecord[],
  targetDir: string,
  strategy: ConflictStrategy,
  perFileOverrides: Map<string, ConflictStrategy>
): Promise<void> {
  if (!fileExists(src)) return

  const relPath = path.relative(targetDir, dest)
  const action = applyStrategy(dest, strategy, perFileOverrides, targetDir)

  if (action === 'skip') {
    console.warn(`⚠️  Skipping existing file: ${relPath}`)
    return
  }

  copyFile(src, dest)
  records.push({
    path: relPath,
    hash: fileHash(dest),
    source: path.relative(targetDir, src),
  })
}
