import path from 'node:path'
import { discoverExtensions } from '../extensions/discovery.js'
import type { ConflictStrategy, FileRecord } from '../types.js'
import { applyStrategy } from '../utils/conflict-strategy.js'
import { copyFile, ensureDir, fileExists, fileHash, isDirectory, listDir } from '../utils/files.js'

export interface ScaffoldOrchestrationOptions {
  targetDir: string
  libraryDir: string
  fileRecords: FileRecord[]
  strategy: ConflictStrategy
  perFileOverrides: Map<string, ConflictStrategy>
}

export async function scaffoldOrchestration(opts: ScaffoldOrchestrationOptions): Promise<void> {
  const sourceRoot = path.join(opts.libraryDir, 'orchestration')
  if (!fileExists(sourceRoot)) return

  const targetRoot = path.join(opts.targetDir, '.ai', 'orchestration')
  ensureDir(targetRoot)

  await copyTree({
    sourceRoot,
    currentSourceDir: sourceRoot,
    currentTargetDir: targetRoot,
    targetDir: opts.targetDir,
    recordSourceRoot: opts.libraryDir,
    fileRecords: opts.fileRecords,
    strategy: opts.strategy,
    perFileOverrides: opts.perFileOverrides,
  })

  for (const extension of discoverExtensions(opts.targetDir)) {
    const extensionOrchestrationDir = path.join(extension.path, 'orchestration')
    if (!isDirectory(extensionOrchestrationDir)) continue

    for (const category of ['chains', 'teams', 'workflows'] as const) {
      const sourceDir = path.join(extensionOrchestrationDir, category)
      if (!isDirectory(sourceDir)) continue

      const targetDir = path.join(targetRoot, category)
      ensureDir(targetDir)

      await copyTree({
        sourceRoot: sourceDir,
        currentSourceDir: sourceDir,
        currentTargetDir: targetDir,
        targetDir: opts.targetDir,
        recordSourceRoot: extension.path,
        fileRecords: opts.fileRecords,
        strategy: opts.strategy,
        perFileOverrides: opts.perFileOverrides,
      })
    }
  }
}

interface CopyTreeOptions {
  sourceRoot: string
  currentSourceDir: string
  currentTargetDir: string
  targetDir: string
  recordSourceRoot: string
  fileRecords: FileRecord[]
  strategy: ConflictStrategy
  perFileOverrides: Map<string, ConflictStrategy>
}

async function copyTree(opts: CopyTreeOptions): Promise<void> {
  const entries = listDir(opts.currentSourceDir)

  for (const entry of entries) {
    const sourcePath = path.join(opts.currentSourceDir, entry)
    const targetPath = path.join(opts.currentTargetDir, entry)

    if (isDirectory(sourcePath)) {
      ensureDir(targetPath)
      await copyTree({
        ...opts,
        currentSourceDir: sourcePath,
        currentTargetDir: targetPath,
      })
      continue
    }

    const relPath = path.relative(opts.targetDir, targetPath)
    const action = applyStrategy(targetPath, opts.strategy, opts.perFileOverrides, opts.targetDir)

    if (action === 'skip') {
      console.warn(`⚠️  Skipping existing file: ${relPath}`)
      continue
    }

    copyFile(sourcePath, targetPath)
    opts.fileRecords.push({
      path: relPath,
      hash: fileHash(targetPath),
      source: path.relative(opts.recordSourceRoot, sourcePath),
      owner: 'library',
    })
  }
}
