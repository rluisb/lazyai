import path from 'node:path'
import type { ConflictStrategy, FileRecord, SetupScope } from '../types.js'
import { applyStrategy } from '../utils/conflict-strategy.js'
import { copyFile, ensureDir, fileExists, fileHash } from '../utils/files.js'

export interface ScaffoldSpecsOptions {
  targetDir: string
  setupScope?: SetupScope
  libraryDir: string
  specsDirs: string[]
  specsAgents: string[]
  fileRecords: FileRecord[]
  strategy: ConflictStrategy
  perFileOverrides: Map<string, ConflictStrategy>
}

/**
 * Creates specs directory structure and copies AGENTS.md files.
 *
 * Behavior:
 * - For each dir in `specsDirs`: create `specs/<dir>/`
 * - Special case: `memory` also creates `specs/memory/handoffs/`
 * - For each dir in `specsAgents` (must be subset of specsDirs):
 *   copy `library/specs-agents/<dir>.md` → `specs/<dir>/AGENTS.md`
 * - Record each copied AGENTS.md in `fileRecords`
 * - If `specsDirs` is empty, create no directories
 */
export async function scaffoldSpecs(opts: ScaffoldSpecsOptions): Promise<void> {
  const { targetDir, setupScope, libraryDir, specsDirs, specsAgents, fileRecords, strategy, perFileOverrides } = opts

  // Create specs root
  const specsDir = path.join(targetDir, 'specs')

  // 1. Create selected specs directories
  if (specsDirs.length > 0) {
    for (const dir of specsDirs) {
      ensureDir(path.join(specsDir, dir))

      // Special case: memory also needs handoffs subdirectory
      if (dir === 'memory') {
        if (setupScope === 'workspace') {
          ensureDir(path.join(specsDir, 'memory', 'decisions'))
          ensureDir(path.join(specsDir, 'memory', 'patterns'))
          ensureDir(path.join(specsDir, 'memory', 'projects'))
        }

        ensureDir(path.join(specsDir, 'memory', 'handoffs'))
      }
    }
  }

  // 2. Copy AGENTS.md files for selected specs directories
  for (const dir of specsAgents) {
    const src = path.join(libraryDir, 'specs-agents', `${dir}.md`)
    const dest = path.join(specsDir, dir, 'AGENTS.md')

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
