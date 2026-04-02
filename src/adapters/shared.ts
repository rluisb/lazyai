import path from 'node:path'
import * as files from '../utils/files.js'
import { resolveConflict } from '../utils/conflicts.js'
import type { AdapterContext } from './types.js'

type SelectionKey = 'agents' | 'skills' | 'prompts'

interface CopyWithRecordOptions {
  src: string
  dest: string
  ctx: AdapterContext
  dryRun?: boolean
  warnOnSkip?: boolean
  transform?: (content: string) => string
}

interface CopyLibraryDirectoryOptions {
  ctx: AdapterContext
  sourceSubdir: SelectionKey
  selectionKey: SelectionKey
  toDestPath: (file: string) => string
  dryRun?: boolean
  warnOnSkip?: boolean
  transform?: (content: string) => string
}

interface ToolContextFilesOptions {
  ctx: AdapterContext
  toolDir: string
  contextFileName: string
  agentsDestDir: string
  skillsDestDir: string
  templatesDestDir: string
  warnOnSkip?: boolean
}

interface RootTemplateOptions {
  ctx: AdapterContext
  recordPath: string
  destPath: string
  templateSource: string
}

function getSelectionSet(ctx: AdapterContext, key: SelectionKey): Set<string> | undefined {
  const selected = ctx.selections?.[key]
  return selected ? new Set(selected) : undefined
}

export async function copyWithRecord(opts: CopyWithRecordOptions): Promise<void> {
  const relPath = path.relative(opts.ctx.targetDir, opts.dest)
  const dryRun = opts.dryRun ?? opts.ctx.dryRun === true
  const effectiveStrategy = opts.ctx.perFileOverrides?.get(opts.dest) ?? opts.ctx.strategy
  const resolution = await resolveConflict(opts.dest, relPath, {
    force: opts.ctx.force,
    ...(effectiveStrategy ? { strategy: effectiveStrategy } : {}),
  })

  if (resolution === 'skip') {
    if (opts.warnOnSkip) {
      console.warn(`⚠️  Skipping existing file: ${relPath}`)
    }
    return
  }

  if (resolution === 'backup-and-overwrite') {
    if (dryRun) {
      console.log(`[dry-run] Would create: ${relPath}`)
      opts.ctx.fileRecords.push({
        path: relPath,
        hash: 'dry-run',
        source: path.relative(opts.ctx.libraryDir, opts.src),
      })
      return
    }

    files.backupFile(opts.dest, opts.ctx.targetDir)
  }

  if (dryRun) {
    console.log(`[dry-run] Would create: ${relPath}`)
    opts.ctx.fileRecords.push({
      path: relPath,
      hash: 'dry-run',
      source: path.relative(opts.ctx.libraryDir, opts.src),
    })
    return
  }

  if (opts.transform) {
    const transformed = opts.transform(files.readFile(opts.src))
    files.writeFile(opts.dest, transformed)
  } else {
    files.copyFile(opts.src, opts.dest)
  }

  opts.ctx.fileRecords.push({
    path: relPath,
    hash: files.fileHash(opts.dest),
    source: path.relative(opts.ctx.libraryDir, opts.src),
  })
}

export async function copyLibraryDirectory(opts: CopyLibraryDirectoryOptions): Promise<void> {
  const selected = getSelectionSet(opts.ctx, opts.selectionKey)
  const sourceDir = path.join(opts.ctx.libraryDir, opts.sourceSubdir)

  for (const file of files.listDir(sourceDir)) {
    const srcPath = path.join(sourceDir, file)
    if (files.isDirectory(srcPath)) continue

    const fileId = path.parse(file).name
    if (selected && !selected.has(fileId)) continue

    const copyOpts: CopyWithRecordOptions = {
      src: srcPath,
      dest: opts.toDestPath(file),
      ctx: opts.ctx,
      ...(opts.dryRun !== undefined ? { dryRun: opts.dryRun } : {}),
    }
    if (opts.warnOnSkip !== undefined) {
      copyOpts.warnOnSkip = opts.warnOnSkip
    }
    if (opts.transform !== undefined) {
      copyOpts.transform = opts.transform
    }

    await copyWithRecord(copyOpts)
  }
}

export async function installToolContextFiles(opts: ToolContextFilesOptions): Promise<void> {
  const toolAgentsDir = path.join(opts.ctx.libraryDir, 'tool-agents')
  const withWarn = opts.warnOnSkip !== undefined ? { warnOnSkip: opts.warnOnSkip } : {}

  await copyWithRecord({
    src: path.join(toolAgentsDir, 'agents-dir.md'),
    dest: path.join(opts.toolDir, opts.agentsDestDir, opts.contextFileName),
    ctx: opts.ctx,
    ...withWarn,
  })

  await copyWithRecord({
    src: path.join(toolAgentsDir, 'skills-dir.md'),
    dest: path.join(opts.toolDir, opts.skillsDestDir, opts.contextFileName),
    ctx: opts.ctx,
    ...withWarn,
  })

  await copyWithRecord({
    src: path.join(toolAgentsDir, 'templates-dir.md'),
    dest: path.join(opts.toolDir, opts.templatesDestDir, opts.contextFileName),
    ctx: opts.ctx,
    ...withWarn,
  })

  await copyWithRecord({
    src: path.join(toolAgentsDir, 'root-dir.md'),
    dest: path.join(opts.toolDir, opts.contextFileName),
    ctx: opts.ctx,
    ...withWarn,
  })
}

export async function installRootTemplateIfMissing(opts: RootTemplateOptions): Promise<void> {
  const alreadyCreated = opts.ctx.fileRecords.some((r) => r.path === opts.recordPath)
  if (alreadyCreated) return

  const templatePath = path.join(opts.ctx.libraryDir, opts.templateSource)
  if (!files.fileExists(templatePath)) return

  const effectiveStrategy = opts.ctx.perFileOverrides?.get(opts.destPath) ?? opts.ctx.strategy
  const resolution = await resolveConflict(opts.destPath, opts.recordPath, {
    force: opts.ctx.force,
    ...(effectiveStrategy ? { strategy: effectiveStrategy } : {}),
  })
  if (resolution === 'skip') return

  if (resolution === 'backup-and-overwrite') {
    files.backupFile(opts.destPath, opts.ctx.targetDir)
  }

  files.writeFile(opts.destPath, files.readFile(templatePath))
  opts.ctx.fileRecords.push({
    path: opts.recordPath,
    hash: files.fileHash(opts.destPath),
    source: opts.templateSource,
  })
}
