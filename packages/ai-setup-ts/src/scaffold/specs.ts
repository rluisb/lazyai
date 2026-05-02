import * as fs from 'node:fs'
import path from 'node:path'
import type { ConflictStrategy, FileRecord, SetupScope } from '../types.js'
import { applyStrategy } from '../utils/conflict-strategy.js'
import { copyFile, ensureDir, fileExists, fileHash, isDirectory, writeFile } from '../utils/files.js'

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

const SPECIFY_TEMPLATES = [
  'spec-template.md',
  'plan-template.md',
  'tasks-template.md',
  'checklist-template.md',
  'task-harness-template.md',
]

const SPECIFY_SCRIPTS = ['create-new-feature.sh']

const STARTER_STANDARD_FILES = [
  'agent-security.md',
  'context-loading.md',
  'error-handling.md',
  'orchestration-patterns.md',
  'test-patterns.md',
]

/**
 * Detects whether an existing spec-kit structure is present.
 * Returns true if .specify/ directory exists.
 */
export function hasSpecKitStructure(targetDir: string): boolean {
  return isDirectory(path.join(targetDir, '.specify'))
}

/**
 * Detects numbered spec directories (###-slug pattern).
 * Returns the highest existing number, or 0 if none.
 */
export function detectExistingSpecs(targetDir: string): { hasSpecs: boolean; highestNumber: number } {
  const specsDir = path.join(targetDir, 'specs')
  if (!isDirectory(specsDir)) return { hasSpecs: false, highestNumber: 0 }

  let highest = 0
  const entries = fileExists(specsDir) ? fs.readdirSync(specsDir, { withFileTypes: true }) : []
  for (const entry of entries) {
    if (!entry.isDirectory()) continue
    const match = entry.name.match(/^(\d{3})-/)
    if (match?.[1]) {
      const num = Number.parseInt(match[1], 10)
      if (num > highest) highest = num
    }
  }

  return { hasSpecs: highest > 0, highestNumber: highest }
}

/**
 * Creates the speckit-compatible .specify/ and specs/ directory structure.
 *
 * Behavior:
 * - If `.specify/` already exists: skip .specify/ scaffolding entirely (respect existing)
 * - If `specs/###-slug/` directories exist: skip spec directory creation
 * - Greenfield: create `specs/.gitkeep` + full `.specify/` with templates, memory, scripts
 * - Workspace mode also creates `.specify/memory/repos/` for ledger files
 * - Always copies template files from library to `.specify/templates/`
 */
export async function scaffoldSpecs(opts: ScaffoldSpecsOptions): Promise<void> {
  const { targetDir, setupScope, libraryDir, specsDirs, specsAgents, fileRecords, strategy, perFileOverrides } = opts

  const hasSpecify = hasSpecKitStructure(targetDir)
  const existing = detectExistingSpecs(targetDir)

  // 1. Create .specify/ directory (speckit core)
  if (!hasSpecify) {
    const specifyDir = path.join(targetDir, '.specify')

    // 1a. .specify/templates/ — copy template files from library
    const templatesDir = path.join(specifyDir, 'templates')
    ensureDir(templatesDir)
    for (const tplFile of SPECIFY_TEMPLATES) {
      const src = path.join(libraryDir, 'templates', tplFile)
      const dest = path.join(templatesDir, tplFile)
      await copyLibraryFile(src, dest, fileRecords, targetDir, strategy, perFileOverrides)
    }

    // 1b. .specify/memory/ — constitution is written by scaffoldConstitution
    //     (with placeholder substitution); just ensure the dir + workspace repos.
    const memoryDir = path.join(specifyDir, 'memory')
    ensureDir(memoryDir)

    // Workspace: create repos/ ledger directory
    if (setupScope === 'workspace') {
      ensureDir(path.join(memoryDir, 'repos'))
    }

    // 1c. .specify/scripts/
    if (SPECIFY_SCRIPTS.length > 0) {
      const scriptsDir = path.join(specifyDir, 'scripts', 'bash')
      ensureDir(scriptsDir)
    }
  }

  // 2. Create specs/ directory
  const specsDir = path.join(targetDir, 'specs')
  ensureDir(specsDir)

  // Greenfield: create .gitkeep to signal empty spec directory
  if (!existing.hasSpecs && specsDirs.length === 0) {
    const gitkeepPath = path.join(specsDir, '.gitkeep')
    const action = applyStrategy(gitkeepPath, strategy, perFileOverrides, targetDir)
    if (action !== 'skip') {
      writeFile(gitkeepPath, '# Specs\n\nFeature specifications are created by the `/speckit.specify` command.\n')
      fileRecords.push({
        path: path.relative(targetDir, gitkeepPath).replaceAll('\\', '/'),
        hash: fileHash(gitkeepPath),
        source: 'speckit:specs-root',
        owner: 'library',
      })
    }
  }

  // Seed starter standards file-by-file. Metadata files such as .gitkeep or
  // README.md must not suppress missing standards, and existing user-authored
  // files at the same paths are always preserved regardless of conflict strategy.
  copyStarterStandards(libraryDir, targetDir, fileRecords)

  // Legacy: create selected specs directories (only if existing spec-kit structure AND user selected them)
  if (specsDirs.length > 0) {
    for (const dir of specsDirs) {
      ensureDir(path.join(specsDir, dir))

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

  // 3. Copy AGENTS.md files for selected specs directories (unchanged behavior)
  for (const dir of specsAgents) {
    const src = path.join(libraryDir, 'specs-agents', `${dir}.md`)
    const dest = path.join(specsDir, dir, 'AGENTS.md')

    await copyLibraryFile(src, dest, fileRecords, targetDir, strategy, perFileOverrides)
  }
}

function copyStarterStandards(libraryDir: string, targetDir: string, records: FileRecord[]): void {
  const starterDir = path.join(targetDir, 'specs', 'standards', 'starter')

  for (const standard of STARTER_STANDARD_FILES) {
    const sourcePath = path.join(libraryDir, 'standards', 'starter', standard)
    if (!fileExists(sourcePath)) continue

    const destPath = path.join(starterDir, standard)
    if (fileExists(destPath)) continue

    copyFile(sourcePath, destPath)
    records.push({
      path: path.relative(targetDir, destPath).replaceAll('\\', '/'),
      hash: fileHash(destPath),
      source: `standards:standards/starter/${standard}`,
      owner: 'library',
    })
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
  perFileOverrides: Map<string, ConflictStrategy>,
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
    owner: 'library',
  })
}
