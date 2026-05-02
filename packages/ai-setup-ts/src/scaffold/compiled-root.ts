import * as fs from 'node:fs'
import path from 'node:path'
import { type FragmentContext, TemplateCompiler } from '../compiler/index.js'
import type { FeatureFlags, GitConventions } from '../store/schema.js'
import type { ConflictStrategy, FileRecord, SetupScope, ToolId } from '../types.js'
import { applyStrategy } from '../utils/conflict-strategy.js'
import { ensureDir, fileHash, writeFile } from '../utils/files.js'
import { detectProjectStack } from '../utils/repo-detection.js'
import { ROOT_FILE_BY_TOOL } from './root-file-map.js'

const DEFAULT_ENABLED_FEATURES: FeatureFlags = {
  contextEngineering: true,
  rpiWorkflow: true,
  chainOfThought: true,
  treeOfThoughts: true,
  adrEnforcement: true,
  qualityGates: true,
  agentHarness: true,
  bugResolution: true,
  pivotHandling: true,
  adversarialDesign: false,
}

const CLAUDE_AGENTS_REFERENCE = '<!-- ai-setup: AGENTS.md reference -->\nThis project uses [AGENTS.md](./AGENTS.md) as the canonical AI agent instruction file.'

export interface TargetedUpdatePatch {
  file: string
  replacements: Array<{
    field: string
    oldText: string
    newText: string
    location: {
      section: string | null
      lineStart: number | null
      lineEnd: number | null
    }
  }>
  warnings: string[]
  preservedUnrecognizedContent: true
}

interface TargetedFieldSpec {
  field: string
  newText: string
  section: string
  placeholders: string[]
  linePrefixes: string[]
}

export interface ScaffoldCompiledRootOptions {
  targetDir: string
  libraryDir: string
  tools: ToolId[]
  projectName: string
  planningDir: string
  features?: FeatureFlags
  gitConventions?: GitConventions
  fileRecords: FileRecord[]
  strategy: ConflictStrategy
  perFileOverrides: Map<string, ConflictStrategy>
  setupScope?: SetupScope
  // Optional context overrides
  primaryLanguage?: string
  framework?: string
  workspaceType?: string
  projectInstructions?: string
  constitution?: FragmentContext['constitution']
  projectOverview?: string
  namingConventions?: string
  errorHandling?: string
  apiConventions?: string
  importOrder?: string
  protectedBranch?: string
  testCommand?: string
  lintCommand?: string
  buildCommand?: string
  coverageThreshold?: number
  codebaseMap?: Array<{ path: string; responsibility?: string }>
  /** Referenced repos for workspace scope — appended as context to compiled root */
  repos?: Array<{ name: string; path: string; type?: string; description?: string }>
}

/**
 * Compiles and writes root AI tool configuration files using the shared
 * fragment/template compilation system.
 *
 * Default behavior:
 * - When features are omitted, schema/wizard defaults are used (all feature flags enabled)
 * - Callers can pass explicit features to disable specific blocks
 * - Git-conventions blocks are included when gitConventions context is provided
 */
export async function scaffoldCompiledRoot(opts: ScaffoldCompiledRootOptions): Promise<void> {
  const {
    targetDir,
    libraryDir,
    tools,
    projectName,
    planningDir,
    features,
    gitConventions,
    fileRecords,
    strategy,
    perFileOverrides,
    setupScope,
    primaryLanguage,
    framework,
    workspaceType,
    projectInstructions,
    constitution,
    projectOverview,
    namingConventions,
    errorHandling,
    apiConventions,
    importOrder,
    protectedBranch,
    testCommand,
    lintCommand,
    buildCommand,
    coverageThreshold,
    codebaseMap,
    repos,
  } = opts

  const effectiveFeatures: FeatureFlags = {
    ...DEFAULT_ENABLED_FEATURES,
    ...(features ?? {}),
  }
  const stack = setupScope === 'project' ? detectProjectStack(targetDir) : undefined
  const effectivePrimaryLanguage = primaryLanguage ?? normalizeDetectedStackValue(stack?.language)
  const effectiveFramework = framework ?? normalizeDetectedStackValue(stack?.framework)

  // Build fragment context from options
  const context: FragmentContext = {
    projectName,
    planningDir,
    ...(effectivePrimaryLanguage !== undefined ? { primaryLanguage: effectivePrimaryLanguage } : {}),
    ...(effectiveFramework !== undefined ? { framework: effectiveFramework } : {}),
    ...(workspaceType != null ? { workspaceType } : {}),
    ...(projectInstructions != null ? { projectInstructions } : {}),
    ...(stack?.testFramework ? { testFramework: stack.testFramework } : {}),
    ...(stack?.packageManager ? { packageManager: stack.packageManager } : {}),
    ...(stack?.commands.test ? { testCommand: stack.commands.test } : {}),
    ...(stack?.commands.lint ? { lintCommand: stack.commands.lint } : {}),
    ...(stack?.commands.build ? { buildCommand: stack.commands.build } : {}),
    ...(stack?.commands.dev ? { devCommand: stack.commands.dev } : {}),
    ...(stack?.commands.install ? { installCommand: stack.commands.install } : {}),
    ...(stack?.description ? { projectDescription: stack.description } : {}),
    features: {
      contextEngineering: effectiveFeatures.contextEngineering,
      rpiWorkflow: effectiveFeatures.rpiWorkflow,
      chainOfThought: effectiveFeatures.chainOfThought,
      treeOfThoughts: effectiveFeatures.treeOfThoughts,
      adrEnforcement: effectiveFeatures.adrEnforcement,
      qualityGates: effectiveFeatures.qualityGates,
      agentHarness: effectiveFeatures.agentHarness,
      bugResolution: effectiveFeatures.bugResolution,
      pivotHandling: effectiveFeatures.pivotHandling,
      adversarialDesign: effectiveFeatures.adversarialDesign,
      gitConventions: Boolean(gitConventions),
      // Legacy aliases for existing snake_case template conditionals
      context_engineering: effectiveFeatures.contextEngineering,
      rpi_workflow: effectiveFeatures.rpiWorkflow,
      chain_of_thought: effectiveFeatures.chainOfThought,
      tree_of_thoughts: effectiveFeatures.treeOfThoughts,
      adr_enforcement: effectiveFeatures.adrEnforcement,
      quality_gates: effectiveFeatures.qualityGates,
      agent_harness: effectiveFeatures.agentHarness,
      bug_resolution: effectiveFeatures.bugResolution,
      pivot_handling: effectiveFeatures.pivotHandling,
      adversarial_design: effectiveFeatures.adversarialDesign,
      git_conventions: Boolean(gitConventions),
    },
    constitution: {
      ...(constitution ?? {}),
      ...(projectOverview != null ? { projectOverview } : {}),
      ...(namingConventions != null ? { namingConventions } : {}),
      ...(errorHandling != null ? { errorHandling } : {}),
      ...(apiConventions != null ? { apiConventions } : {}),
      ...(importOrder != null ? { importOrder } : {}),
      ...(protectedBranch != null ? { protectedBranch } : {}),
      ...(testCommand != null ? { testCommand } : {}),
      ...(lintCommand != null ? { lintCommand } : {}),
      ...(buildCommand != null ? { buildCommand } : {}),
      ...(coverageThreshold != null ? { coverageThreshold } : {}),
      ...optionalCodebaseMap(codebaseMap ?? detectTopLevelCodebaseMap(targetDir)),
      stack: {
        ...(constitution?.stack ?? {}),
        ...(effectivePrimaryLanguage !== undefined ? { language: effectivePrimaryLanguage } : {}),
        ...(effectiveFramework !== undefined ? { framework: effectiveFramework } : {}),
        ...(stack?.testFramework ? { testing: stack.testFramework } : {}),
        ...(stack?.packageManager ? { packageManager: stack.packageManager } : {}),
      },
      commands: {
        ...(constitution?.commands ?? {}),
        ...optionalCommandValue('test', testCommand ?? stack?.commands.test),
        ...optionalCommandValue('lint', lintCommand ?? stack?.commands.lint),
        ...optionalCommandValue('build', buildCommand ?? stack?.commands.build),
      },
    },
  }

  // Compile for each tool
  for (const tool of tools) {
    if (tool === 'claude-code') {
      appendClaudeAgentsReference(targetDir, setupScope)
    }

    const compiler = new TemplateCompiler({
      libraryDir,
      outputDir: targetDir,
      tool,
      context,
    })

    const result = compiler.compile()

    // For workspace scope, append repo context to compiled output
    let workspaceReposSection = ''
    if (repos && repos.length > 0) {
      const lines = [
        '',
        '## Workspace Repos',
        '',
        'This workspace contains the following repositories:',
        '',
      ]
      for (const repo of repos) {
        lines.push(`### ${repo.name}`)
        lines.push('')
        lines.push(`- **Path**: \`${repo.path}\``)
        if (repo.type && repo.type !== 'unknown') lines.push(`- **Type**: ${repo.type}`)
        if (repo.description) lines.push(`- **Description**: ${repo.description}`)
        lines.push('')
      }
      lines.push('When working in a repo, refer to its README or package.json for repo-specific details.')
      lines.push('')
      workspaceReposSection = lines.join('\n')
    }

    // Write each compiled file
    for (const file of result.files) {
      // Map 'root.md' to the current root instruction filename (e.g., AGENTS.md).
      let outputPath = file.relativePath
      if (outputPath === 'root.md') {
        outputPath = ROOT_FILE_BY_TOOL[tool]
      }

      const destPath = path.join(targetDir, outputPath)
      const destDir = path.dirname(destPath)

      // Ensure parent directory exists
      ensureDir(destDir)

      // Check conflict strategy
      const action = applyStrategy(destPath, strategy, perFileOverrides, targetDir)
      if (action === 'skip') continue

      // Write the compiled content
      let content = workspaceReposSection ? file.content + workspaceReposSection : file.content
      if (outputPath === 'AGENTS.md' && fs.existsSync(destPath)) {
        const existing = fs.readFileSync(destPath, 'utf-8')
        const result = buildTargetedAgentsUpdatePatch(outputPath, existing, context)
        for (const warning of result.patch.warnings) {
          console.warn(`Warning: targeted ${outputPath} update: ${warning}`)
        }
        content = result.content
      }
      writeFile(destPath, content)

      // Record the file
      fileRecords.push({
        path: outputPath,
        hash: fileHash(destPath),
        source: `compiled:${tool}`,
        owner: 'library',
      })
    }
  }
}

function detectTopLevelCodebaseMap(targetDir: string): Array<{ path: string; responsibility?: string }> | undefined {
  if (!fs.existsSync(targetDir)) return undefined
  const ignored = new Set(['node_modules', 'dist', '.git', 'vendor'])
  const entries = fs.readdirSync(targetDir, { withFileTypes: true })
    .filter((entry) => entry.isDirectory() && !ignored.has(entry.name))
    .map((entry) => ({ path: entry.name }))

  return entries.length > 0 ? entries : undefined
}

function normalizeDetectedStackValue(value: string | undefined): string | undefined {
  if (value === undefined || value === 'Unknown') return undefined
  return value
}

function optionalCodebaseMap(
  value: Array<{ path: string; responsibility?: string }> | undefined,
): Pick<NonNullable<FragmentContext['constitution']>, 'codebaseMap'> | Record<string, never> {
  return value !== undefined ? { codebaseMap: value } : {}
}

function optionalCommandValue(key: 'test' | 'lint' | 'build', value: string | undefined): Partial<NonNullable<FragmentContext['constitution']>['commands']> {
  return value != null ? { [key]: value } : {}
}

export function buildTargetedAgentsUpdatePatch(
  file: string,
  existing: string,
  context: FragmentContext,
): { content: string; patch: TargetedUpdatePatch } {
  const patch: TargetedUpdatePatch = {
    file,
    replacements: [],
    warnings: [],
    preservedUnrecognizedContent: true,
  }
  let content = existing

  for (const spec of targetedAgentsFieldSpecs(context)) {
    if (spec.newText === '') continue
    for (const placeholder of spec.placeholders) {
      content = replaceTargetedExact(content, placeholder, spec, patch)
    }
  }

  for (const spec of targetedAgentsFieldSpecs(context)) {
    if (spec.newText === '' || spec.linePrefixes.length === 0) continue
    content = replaceTargetedLineSlots(content, spec, patch)
  }
  warnUnsafeProjectOverview(content, context, patch)

  return { content, patch }
}

function targetedAgentsFieldSpecs(context: FragmentContext): TargetedFieldSpec[] {
  const constitution = context.constitution ?? {}
  const stack = constitution.stack ?? {}
  const conventions = constitution.conventions ?? {}
  const commands = constitution.commands ?? {}
  const coverage = constitution.coverageThreshold != null ? String(constitution.coverageThreshold) : ''

  return [
    { field: 'PROJECT_OVERVIEW', newText: cleanTargetedString(constitution.projectOverview), section: 'Project Overview', placeholders: ['[YOUR_PROJECT_OVERVIEW]'], linePrefixes: [] },
    { field: 'LANGUAGE', newText: cleanTargetedString(stack.language), section: 'Project Overview', placeholders: ['[YOUR_LANGUAGE]'], linePrefixes: ['- Language: '] },
    { field: 'FRAMEWORK', newText: cleanTargetedString(stack.framework), section: 'Project Overview', placeholders: ['[YOUR_FRAMEWORK]'], linePrefixes: ['- Framework: '] },
    { field: 'DATABASE', newText: cleanTargetedString(stack.database), section: 'Project Overview', placeholders: ['[YOUR_DATABASE]'], linePrefixes: ['- Database: '] },
    { field: 'ORM', newText: cleanTargetedString(stack.orm), section: 'Project Overview', placeholders: ['[YOUR_ORM]'], linePrefixes: ['- ORM/Query: '] },
    { field: 'TEST_FRAMEWORK', newText: cleanTargetedString(stack.testing), section: 'Project Overview', placeholders: ['[YOUR_TEST_FRAMEWORK]'], linePrefixes: ['- Testing: '] },
    { field: 'PACKAGE_MANAGER', newText: cleanTargetedString(stack.packageManager), section: 'Project Overview', placeholders: ['[YOUR_PACKAGE_MANAGER]'], linePrefixes: ['- Package manager: '] },
    { field: 'NAMING_CONVENTIONS', newText: cleanTargetedString(conventions.naming ?? constitution.namingConventions), section: 'Conventions', placeholders: ['[YOUR_NAMING_CONVENTION]'], linePrefixes: [] },
    { field: 'ERROR_HANDLING', newText: cleanTargetedString(conventions.errorHandling ?? constitution.errorHandling), section: 'Conventions', placeholders: ['[YOUR_ERROR_PATTERN]'], linePrefixes: [] },
    { field: 'API_CONVENTIONS', newText: cleanTargetedString(conventions.apiResponses ?? constitution.apiConventions), section: 'Conventions', placeholders: ['[YOUR_API_CONVENTION]'], linePrefixes: [] },
    { field: 'IMPORT_ORDER', newText: cleanTargetedString(conventions.importOrder ?? constitution.importOrder), section: 'Conventions', placeholders: ['[YOUR_IMPORT_ORDER]'], linePrefixes: [] },
    { field: 'PROTECTED_BRANCH', newText: cleanTargetedString(constitution.protectedBranch), section: 'Do NOT', placeholders: ['[YOUR_PROTECTED_BRANCH]'], linePrefixes: [] },
    { field: 'TEST_COMMAND', newText: cleanTargetedString(commands.test ?? constitution.testCommand), section: 'Key Commands', placeholders: ['<!-- fill-in: test command -->'], linePrefixes: [] },
    { field: 'LINT_COMMAND', newText: cleanTargetedString(commands.lint ?? constitution.lintCommand), section: 'Key Commands', placeholders: ['[YOUR_LINT_COMMAND]'], linePrefixes: [] },
    { field: 'BUILD_COMMAND', newText: cleanTargetedString(commands.build ?? constitution.buildCommand), section: 'Key Commands', placeholders: ['<!-- fill-in: build command -->'], linePrefixes: [] },
    { field: 'COVERAGE_THRESHOLD', newText: coverage, section: 'Testing', placeholders: ['[YOUR_COVERAGE_THRESHOLD]'], linePrefixes: ['- Minimum coverage: ', '- Minimum coverage threshold: '] },
  ]
}

function replaceTargetedExact(content: string, oldText: string, spec: TargetedFieldSpec, patch: TargetedUpdatePatch): string {
  if (oldText === '' || spec.newText === '' || oldText === spec.newText) return content
  let updated = content
  let searchStart = 0
  while (searchStart <= updated.length) {
    const index = updated.indexOf(oldText, searchStart)
    if (index < 0) return updated
    const line = lineNumberAt(updated, index)
    patch.replacements.push({
      field: spec.field,
      oldText,
      newText: spec.newText,
      location: { section: spec.section, lineStart: line, lineEnd: line },
    })
    updated = `${updated.slice(0, index)}${spec.newText}${updated.slice(index + oldText.length)}`
    searchStart = index + spec.newText.length
  }
  return updated
}

function replaceTargetedLineSlots(content: string, spec: TargetedFieldSpec, patch: TargetedUpdatePatch): string {
  const lines = content.split(/(?<=\n)/)
  let changed = false
  for (const [index, line] of lines.entries()) {
    const { body, ending } = splitLineEnding(line)
    for (const prefix of spec.linePrefixes) {
      if (!body.startsWith(prefix)) continue
      const oldValue = body.slice(prefix.length).trim()
      if (normalizeSlotValue(oldValue) === spec.newText) continue
      if (!isSafeTargetedSlot(oldValue, spec)) {
        patch.warnings.push(`left ${spec.field} unchanged at line ${index + 1} because existing value is not a recognized placeholder/value slot`)
        continue
      }
      lines[index] = `${prefix}${preserveSlotDelimiters(oldValue, spec.newText)}${ending}`
      patch.replacements.push({
        field: spec.field,
        oldText: oldValue,
        newText: spec.newText,
        location: { section: spec.section, lineStart: index + 1, lineEnd: index + 1 },
      })
      changed = true
    }
  }
  return changed ? lines.join('') : content
}

function warnUnsafeProjectOverview(content: string, context: FragmentContext, patch: TargetedUpdatePatch): void {
  const overview = cleanTargetedString(context.constitution?.projectOverview)
  if (overview === '' || content.includes(overview) || content.includes('[YOUR_PROJECT_OVERVIEW]')) return
  const lines = content.split('\n')
  const headerIndex = lines.findIndex((line) => line.trim() === '## Project Overview')
  if (headerIndex < 0) return
  for (let index = headerIndex + 1; index < lines.length; index += 1) {
    const line = lines[index]
    if (line == null) continue
    const trimmed = line.trim()
    if (trimmed === '' || trimmed.startsWith('<!--')) continue
    if (trimmed.startsWith('## ') || trimmed.startsWith('**Stack:**')) return
    patch.warnings.push(`left PROJECT_OVERVIEW unchanged at line ${index + 1} because existing value is not a recognized placeholder/value slot`)
    return
  }
}

function splitLineEnding(line: string): { body: string; ending: string } {
  if (line.endsWith('\r\n')) return { body: line.slice(0, -2), ending: '\r\n' }
  if (line.endsWith('\n')) return { body: line.slice(0, -1), ending: '\n' }
  return { body: line, ending: '' }
}

function isSafeTargetedSlot(oldValue: string, spec: TargetedFieldSpec): boolean {
  const normalized = normalizeSlotValue(oldValue)
  if (normalized === '' || normalized.includes('[YOUR_') || normalized.includes('{{') || normalized.includes('fill-in:')) return true
  if (spec.placeholders.some((placeholder) => normalized === placeholder)) return true
  return spec.field === 'COVERAGE_THRESHOLD' && normalized === '80'
}

function normalizeSlotValue(value: string): string {
  return value.trim().replace(/%$/, '').replace(/^`/, '').replace(/`$/, '').trim()
}

function preserveSlotDelimiters(oldValue: string, newText: string): string {
  const trimmed = oldValue.trim()
  if (trimmed.startsWith('`') && trimmed.endsWith('`')) return `\`${newText}\``
  return newText
}

function lineNumberAt(content: string, index: number): number {
  if (index <= 0) return 1
  return content.slice(0, index).split('\n').length
}

function cleanTargetedString(value: string | null | undefined): string {
  return value?.trim() ?? ''
}

function appendClaudeAgentsReference(targetDir: string, setupScope?: SetupScope): void {
  if (setupScope != null && setupScope !== 'project' && setupScope !== 'workspace') return

  const claudePath = path.join(targetDir, 'CLAUDE.md')
  if (!fs.existsSync(claudePath)) return

  const content = fs.readFileSync(claudePath, 'utf-8')
  if (content.includes(CLAUDE_AGENTS_REFERENCE)) return

  const separator = content.endsWith('\n') ? '\n' : '\n\n'
  writeFile(claudePath, `${content}${separator}${CLAUDE_AGENTS_REFERENCE}\n`)
}
