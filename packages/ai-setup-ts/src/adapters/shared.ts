import path from 'node:path'
import { discoverExtensions } from '../extensions/discovery.js'
import { resolveConflict } from '../utils/conflicts.js'
import * as files from '../utils/files.js'
import { extractTools, normalizeToolsFrontmatter, stripYamlFrontmatter } from '../utils/frontmatter.js'
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
  skipRootIfExists?: boolean
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

  // Symlink mode: create symlink to library source instead of copying
  if (opts.ctx.installMode === 'symlink' && !opts.transform) {
    files.symlinkFile(opts.src, opts.dest)
    opts.ctx.fileRecords.push({
      path: relPath,
      hash: files.fileHash(opts.dest),
      source: path.relative(opts.ctx.libraryDir, opts.src),
      owner: 'library',
      kind: 'symlink',
      linkTarget: path.resolve(opts.src),
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
    owner: 'library',
  })
}

export async function copyLibraryDirectory(opts: CopyLibraryDirectoryOptions): Promise<void> {
  const selected = getSelectionSet(opts.ctx, opts.selectionKey)
  const sourceDir = path.join(opts.ctx.libraryDir, opts.sourceSubdir)

  const entries = files.listDir(sourceDir)
  const processedNames = new Set<string>()

  // First pass: directory-based agents (library/agents/<name>/AGENT.md)
  for (const entry of entries) {
    const entryPath = path.join(sourceDir, entry)
    if (!files.isDirectory(entryPath)) continue

    const agentMd = path.join(entryPath, 'AGENT.md')
    if (!files.fileExists(agentMd)) continue

    const fileId = entry
    if (selected && !selected.has(fileId)) continue

    processedNames.add(fileId)
    const copyOpts: CopyWithRecordOptions = {
      src: agentMd,
      dest: opts.toDestPath(`${entry}.md`),
      ctx: opts.ctx,
      ...(opts.dryRun !== undefined ? { dryRun: opts.dryRun } : {}),
    }
    if (opts.warnOnSkip !== undefined) copyOpts.warnOnSkip = opts.warnOnSkip
    if (opts.transform !== undefined) copyOpts.transform = opts.transform
    await copyWithRecord(copyOpts)

    // Copy agent-local mcp.json if it exists
    const agentMcpJson = path.join(entryPath, 'mcp.json')
    if (files.fileExists(agentMcpJson)) {
      // Agent-local MCP is informational — install alongside agent config
      const mcpDest = path.join(path.dirname(opts.toDestPath(`${entry}.md`)), `${entry}.mcp.json`)
      const mcpCopyOpts: CopyWithRecordOptions = {
        src: agentMcpJson,
        dest: mcpDest,
        ctx: opts.ctx,
        warnOnSkip: false,
      }
      await copyWithRecord(mcpCopyOpts)
    }
  }

  // Second pass: flat files (legacy format: library/agents/<name>.md)
  for (const file of entries) {
    if (opts.includeFile && !opts.includeFile(file)) continue

    const srcPath = path.join(sourceDir, file)
    if (files.isDirectory(srcPath)) continue

    const fileId = path.parse(file).name
    if (processedNames.has(fileId)) continue
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

  // Third pass: extension content (same directory structure as library)
  const extensions = discoverExtensions(opts.ctx.targetDir)
  for (const ext of extensions) {
    const extSourceDir = path.join(ext.path, opts.sourceSubdir)
    if (!files.isDirectory(extSourceDir)) continue

    for (const entry of files.listDir(extSourceDir)) {
      const entryPath = path.join(extSourceDir, entry)

      if (files.isDirectory(entryPath)) {
        const agentMd = path.join(entryPath, 'AGENT.md')
        if (!files.fileExists(agentMd)) continue
        const fileId = entry
        if (processedNames.has(fileId)) continue
        if (selected && !selected.has(fileId)) continue

        processedNames.add(fileId)
        const copyOpts: CopyWithRecordOptions = {
          src: agentMd,
          dest: opts.toDestPath(`${entry}.md`),
          ctx: opts.ctx,
          ...(opts.dryRun !== undefined ? { dryRun: opts.dryRun } : {}),
        }
        if (opts.warnOnSkip !== undefined) copyOpts.warnOnSkip = opts.warnOnSkip
        await copyWithRecord(copyOpts)
      } else if (entry.endsWith('.md')) {
        const fileId = path.parse(entry).name
        if (processedNames.has(fileId)) continue
        if (selected && !selected.has(fileId)) continue

        processedNames.add(fileId)
        const copyOpts: CopyWithRecordOptions = {
          src: entryPath,
          dest: opts.toDestPath(entry),
          ctx: opts.ctx,
          ...(opts.dryRun !== undefined ? { dryRun: opts.dryRun } : {}),
        }
        if (opts.warnOnSkip !== undefined) copyOpts.warnOnSkip = opts.warnOnSkip
        await copyWithRecord(copyOpts)
      }
    }
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

  const rootDest = path.join(opts.toolDir, opts.contextFileName)
  if (!(opts.skipRootIfExists === true && files.fileExists(rootDest))) {
    await copyWithRecord({
      src: path.join(toolAgentsDir, 'root-dir.md'),
      dest: rootDest,
      ctx: opts.ctx,
      ...withWarn,
    })
  }
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
