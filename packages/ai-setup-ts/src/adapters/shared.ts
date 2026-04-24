import path from 'node:path'
import { extractTools, normalizeToolsFrontmatter, stripYamlFrontmatter } from '../utils/frontmatter.js'
import { resolveConflict } from '../utils/conflicts.js'
import * as files from '../utils/files.js'
import type { AdapterContext } from './types.js'

const FALLBACK_ORCHESTRATOR_TOOLS = [
  'list_catalog',
  'compose_agent',
  'start_chain',
  'advance_chain',
  'get_status',
  'get_budget',
  'retry_step',
  'escalate_step',
  'handoff',
] as const

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
  includeFile?: (file: string) => boolean
}

interface WriteContentWithRecordOptions {
  dest: string
  content: string
  ctx: AdapterContext
  source: string
  dryRun?: boolean
  warnOnSkip?: boolean
}

interface ToolContextFilesOptions {
  ctx: AdapterContext
  toolDir: string
  contextFileName: string
  agentsDestDir: string
  skillsDestDir?: string
  templatesDestDir?: string
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
        owner: 'library',
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
      owner: 'library',
    })
    return
  }

  files.ensureDir(path.dirname(opts.dest))

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
    owner: 'library',
  })
}

export async function copyLibraryDirectory(opts: CopyLibraryDirectoryOptions): Promise<void> {
  const selected = getSelectionSet(opts.ctx, opts.selectionKey)
  const sourceDir = path.join(opts.ctx.libraryDir, opts.sourceSubdir)

  for (const file of files.listDir(sourceDir)) {
    if (opts.includeFile && !opts.includeFile(file)) continue

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

export async function writeContentWithRecord(opts: WriteContentWithRecordOptions): Promise<void> {
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

  if (resolution === 'backup-and-overwrite' && !dryRun) {
    files.backupFile(opts.dest, opts.ctx.targetDir)
  }

  if (dryRun) {
    console.log(`[dry-run] Would create: ${relPath}`)
    opts.ctx.fileRecords.push({
      path: relPath,
      hash: 'dry-run',
      source: opts.source,
      owner: 'library',
    })
    return
  }

  files.ensureDir(path.dirname(opts.dest))
  files.writeFile(opts.dest, opts.content)
  opts.ctx.fileRecords.push({
    path: relPath,
    hash: files.fileHash(opts.dest),
    source: opts.source,
    owner: 'library',
  })
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

  if (opts.skillsDestDir) {
    await copyWithRecord({
      src: path.join(toolAgentsDir, 'skills-dir.md'),
      dest: path.join(opts.toolDir, opts.skillsDestDir, opts.contextFileName),
      ctx: opts.ctx,
      ...withWarn,
    })
  }

  if (opts.templatesDestDir) {
    await copyWithRecord({
      src: path.join(toolAgentsDir, 'templates-dir.md'),
      dest: path.join(opts.toolDir, opts.templatesDestDir, opts.contextFileName),
      ctx: opts.ctx,
      ...withWarn,
    })
  }

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
    owner: 'library',
  })
}

export function isOrchestratorEnabled(ctx: AdapterContext): boolean {
  return ctx.enableServers?.includes('orchestrator') === true
}

function readOrchestratorAgentSource(ctx: AdapterContext): string {
  const sourcePath = path.join(ctx.libraryDir, 'agents', 'orchestrator.md')
  if (files.fileExists(sourcePath)) {
    return files.readFile(sourcePath)
  }

  return [
    '---',
    'name: Orchestrator',
    'model: opus',
    `tools: ${FALLBACK_ORCHESTRATOR_TOOLS.join(' ')}`,
    '---',
    '',
    '# Orchestrator Agent',
    '',
    'Use the orchestration MCP runtime to coordinate multi-agent execution.',
  ].join('\n')
}

export function readOrchestratorTools(ctx: AdapterContext): string[] {
  const source = readOrchestratorAgentSource(ctx)
  const tools = extractTools(source)
  return tools.length > 0 ? tools : [...FALLBACK_ORCHESTRATOR_TOOLS]
}

export function formatAllowedToolsSection(tools: string[]): string {
  if (tools.length === 0) return ''
  const lines = ['## Allowed MCP Tools', '']
  for (const tool of tools) {
    lines.push(`- ${tool}`)
  }
  return lines.join('\n')
}

export function getOrchestratorAgentContent(ctx: AdapterContext): string {
  return normalizeToolsFrontmatter(readOrchestratorAgentSource(ctx), 'comma')
}

export function getOrchestratorSkillContent(ctx: AdapterContext): string {
  const body = stripYamlFrontmatter(readOrchestratorAgentSource(ctx)).trim()
  const tools = readOrchestratorTools(ctx)

  return [
    '---',
    'name: orchestrator',
    'description: Orchestration MCP runtime guidance',
    '---',
    '',
    '# Orchestrator Skill',
    '',
    body,
    '',
    formatAllowedToolsSection(tools),
    '',
  ].join('\n')
}

export function getOrchestratorPromptContent(ctx: AdapterContext): string {
  const body = stripYamlFrontmatter(readOrchestratorAgentSource(ctx)).trim()
  const tools = readOrchestratorTools(ctx)

  return [
    '---',
    'mode: agent',
    '---',
    '',
    '# Orchestrator Prompt',
    '',
    body,
    '',
    formatAllowedToolsSection(tools),
    '',
  ].join('\n')
}
