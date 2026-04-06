import { existsSync } from 'node:fs'
import path from 'node:path'
import type { ConflictStrategy, FileRecord, InfraId } from '../types.js'
import { applyStrategy } from '../utils/conflict-strategy.js'
import { copyFile, ensureDir, fileExists, fileHash, readFile, writeFile } from '../utils/files.js'

export interface ScaffoldInfraOptions {
  targetDir: string
  libraryDir: string
  infra: InfraId[]
  projectName: string
  fileRecords: FileRecord[]
  strategy: ConflictStrategy
  perFileOverrides: Map<string, ConflictStrategy>
}

/**
 * Scaffolds infrastructure files into the target directory.
 *
 * Behavior:
 * - `pre-commit`: if `.git` exists at targetDir, create `.git/hooks/` and copy `library/infra/pre-commit.hook` → `.git/hooks/pre-commit`
 * - `compliance`: copy `library/infra/compliance.md` → `specs/compliance.md`
 * - `KNOWLEDGE_MAP`: read `library/infra/KNOWLEDGE_MAP.template.md`, replace `[YOUR_PROJECT_NAME]` with projectName, write to `targetDir/KNOWLEDGE_MAP.md`
 * - `codeowners`: copy `library/infra/CODEOWNERS.template` → `targetDir/CODEOWNERS`
 * - Records each written file in `fileRecords`
 * - If `infra` is empty, does nothing
 */
export async function scaffoldInfra(opts: ScaffoldInfraOptions): Promise<void> {
  const { targetDir, libraryDir, infra, projectName, fileRecords, strategy, perFileOverrides } = opts

  if (infra.length === 0) {
    return
  }

  // Process pre-commit hook
  if (infra.includes('pre-commit')) {
    if (existsSync(path.join(targetDir, '.git'))) {
      const hookDir = path.join(targetDir, '.git', 'hooks')
      ensureDir(hookDir)

      const src = path.join(libraryDir, 'infra', 'pre-commit.hook')
      const dest = path.join(hookDir, 'pre-commit')

      if (fileExists(src)) {
        await copyLibraryFile(src, dest, fileRecords, targetDir, strategy, perFileOverrides)
      }
    }
  }

  // Process compliance.md
  if (infra.includes('compliance')) {
    const specsDir = path.join(targetDir, 'specs')
    ensureDir(specsDir)

    const src = path.join(libraryDir, 'infra', 'compliance.md')
    const dest = path.join(specsDir, 'compliance.md')

    if (fileExists(src)) {
      await copyLibraryFile(src, dest, fileRecords, targetDir, strategy, perFileOverrides)
    }
  }

  // Process KNOWLEDGE_MAP
  if (infra.includes('KNOWLEDGE_MAP')) {
    const src = path.join(libraryDir, 'infra', 'KNOWLEDGE_MAP.template.md')
    const dest = path.join(targetDir, 'KNOWLEDGE_MAP.md')

    if (fileExists(src)) {
      let content = readFile(src)
      content = content.replace(/\[YOUR_PROJECT_NAME\]/g, projectName)
      await writeRootFile(dest, content, fileRecords, targetDir, 'infra/KNOWLEDGE_MAP.template.md', strategy, perFileOverrides)
    }
  }

  // Process CODEOWNERS
  if (infra.includes('codeowners')) {
    const src = path.join(libraryDir, 'infra', 'CODEOWNERS.template')
    const dest = path.join(targetDir, 'CODEOWNERS')

    if (fileExists(src)) {
      await copyLibraryFile(src, dest, fileRecords, targetDir, strategy, perFileOverrides)
    }
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

/**
 * Helper: Write content to file with conflict resolution and recording.
 */
async function writeRootFile(
  dest: string,
  content: string,
  records: FileRecord[],
  targetDir: string,
  source: string,
  strategy: ConflictStrategy,
  perFileOverrides: Map<string, ConflictStrategy>
): Promise<void> {
  const relPath = path.relative(targetDir, dest)
  const action = applyStrategy(dest, strategy, perFileOverrides, targetDir)

  if (action === 'skip') {
    console.warn(`⚠️  Skipping existing file: ${relPath}`)
    return
  }

  writeFile(dest, content)
  records.push({
    path: relPath,
    hash: fileHash(dest),
    source,
  })
}
