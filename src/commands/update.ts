import path, { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import * as p from '@clack/prompts'
import type { Command } from 'commander'
import pc from 'picocolors'
import { Errors } from '../errors/index.js'
import { OperationTracker } from '../errors/operation.js'
import { appendOperation, createStore, writeStore } from '../store/index.js'
import type { StoreData, TrackedFile } from '../store/schema.js'
import type { ToolId } from '../types.js'
import { ALL_SKILLS } from '../types.js'
import { resolveConflict } from '../utils/conflicts.js'
import { backupFile, fileExists, fileHash, listDir, readFile, resolveLibraryDir, writeFile } from '../utils/files.js'
import { stripFrontmatterAndInjectModel } from '../utils/frontmatter.js'
import { compareLibrarySkills } from '../utils/library-compare.js'
import { showSummaryBox } from '../utils/ui.js'

interface ExpectedFile {
  path: string
  source: string
  content: string
}

interface UpdateOptions {
  force?: boolean
}

const libraryDir = resolveLibraryDir(dirname(fileURLToPath(import.meta.url)))

const specsAgentMappings: Array<{ source: string; destination: string }> = [
  { source: 'docs.md', destination: 'specs/AGENTS.md' },
  { source: 'features.md', destination: 'specs/features/AGENTS.md' },
  { source: 'bugfixes.md', destination: 'specs/bugfixes/AGENTS.md' },
  { source: 'refactors.md', destination: 'specs/refactors/AGENTS.md' },
  { source: 'tech-debt.md', destination: 'specs/tech-debt/AGENTS.md' },
  { source: 'rules.md', destination: 'specs/rules/AGENTS.md' },
  { source: 'standards.md', destination: 'specs/standards/AGENTS.md' },
  { source: 'templates.md', destination: 'specs/templates/AGENTS.md' },
  { source: 'memory.md', destination: 'specs/memory/AGENTS.md' },
  { source: 'adrs.md', destination: 'specs/adrs/AGENTS.md' },
]

function buildExpectedFiles(data: StoreData, targetDir: string): ExpectedFile[] {
  const expected: ExpectedFile[] = []

  const upsertExpected = (entry: ExpectedFile): void => {
    const index = expected.findIndex((item) => item.path === entry.path)
    if (index >= 0) {
      expected[index] = entry
      return
    }
    expected.push(entry)
  }

  const addDir = (librarySubDir: string, targetSubDir: string, transform?: (content: string) => string): void => {
    const srcDir = join(libraryDir, librarySubDir)
    if (!fileExists(srcDir)) return
    for (const file of listDir(srcDir)) {
      const srcPath = join(srcDir, file)
      const targetPath = join(targetDir, targetSubDir, file)
      const raw = readFile(srcPath)
      upsertExpected({
        path: path.relative(targetDir, targetPath),
        source: path.join(librarySubDir, file).replaceAll('\\', '/'),
        content: transform ? transform(raw) : raw,
      })
    }
  }

  const addFile = (librarySource: string, targetPath: string): void => {
    const srcPath = join(libraryDir, librarySource)
    if (!fileExists(srcPath)) return
    upsertExpected({
      path: path.relative(targetDir, targetPath),
      source: librarySource,
      content: readFile(srcPath),
    })
  }

  const addContent = (targetPath: string, source: string, content: string): void => {
    upsertExpected({
      path: path.relative(targetDir, join(targetDir, targetPath)),
      source,
      content,
    })
  }

  const addSpecsAgents = (): void => {
    for (const mapping of specsAgentMappings) {
      addFile(`specs-agents/${mapping.source}`, join(targetDir, mapping.destination))
    }
  }

  const addSkillsAndPromptsForTool = (tool: ToolId): void => {
    const addSkill = (name: string, targetPath: string): void => {
      addFile(`skills/${name}.md`, join(targetDir, targetPath))
    }
    const addPrompt = (name: string, targetPath: string): void => {
      addFile(`prompts/${name}.md`, join(targetDir, targetPath))
    }

    if (tool === 'opencode') {
      for (const name of ALL_SKILLS) {
        addSkill(name, `.opencode/skills/${name}/SKILL.md`)
      }
    }

    if (tool === 'claude-code') {
      for (const name of ALL_SKILLS) {
        addSkill(name, `.claude/skills/${name}/SKILL.md`)
      }
    }

    if (tool === 'gemini') {
      for (const name of ALL_SKILLS) {
        addSkill(name, `.gemini/skills/${name}/SKILL.md`)
      }
      // Gemini has no templates/ concept — skip prompts
    }

    if (tool === 'copilot') {
      for (const name of ALL_SKILLS) {
        const source = `skills/${name}.md`
        const srcPath = join(libraryDir, source)
        if (fileExists(srcPath)) {
          const transformed = `---\nmode: agent\n---\n\n${readFile(srcPath)}`
          addContent(`.github/prompts/${name}.prompt.md`, source, transformed)
        }
      }
      for (const name of ['research', 'plan', 'implement', 'compact', 'local-example']) {
        addPrompt(name, `.github/prompts/${name}.prompt.md`)
      }
    }
  }

  addDir('templates', 'specs/templates')
  addDir('rules', 'specs/rules')
  addSpecsAgents()

  addFile('infra/CODEOWNERS.template', join(targetDir, 'CODEOWNERS'))
  addFile('infra/compliance.md', join(targetDir, 'specs/compliance.md'))
  addFile('infra/KNOWLEDGE_MAP.template.md', join(targetDir, 'specs/KNOWLEDGE_MAP.md'))

  if (fileExists(join(targetDir, '.git'))) {
    addFile('infra/pre-commit.hook', join(targetDir, '.git/hooks/pre-commit'))
  }

  const rootAgentsTemplatePath = join(libraryDir, 'root/AGENTS.template.md')
  if (fileExists(rootAgentsTemplatePath)) {
    const rootContent = readFile(rootAgentsTemplatePath).replace(/\[YOUR_PROJECT_NAME\]/g, data.config.projectName)
    if (
      data.config.tools.includes('opencode')
      || data.config.tools.includes('copilot')
      || data.config.tools.includes('codex')
    ) {
      addContent('AGENTS.md', 'root/AGENTS.template.md', rootContent)
    }
  }

  if (data.config.tools.includes('claude-code')) {
    const rootClaudeTemplatePath = join(libraryDir, 'root/CLAUDE.template.md')
    if (fileExists(rootClaudeTemplatePath)) {
      addContent(
        'CLAUDE.md',
        'root/CLAUDE.template.md',
        readFile(rootClaudeTemplatePath).replace(/\[YOUR_PROJECT_NAME\]/g, data.config.projectName),
      )
    }
  }

  if (data.config.tools.includes('gemini')) {
    const rootGeminiTemplatePath = join(libraryDir, 'root/GEMINI.template.md')
    if (fileExists(rootGeminiTemplatePath)) {
      addContent(
        'GEMINI.md',
        'root/GEMINI.template.md',
        readFile(rootGeminiTemplatePath).replace(/\[YOUR_PROJECT_NAME\]/g, data.config.projectName),
      )
    }
  }

  if (data.config.tools.includes('copilot')) {
    const rootCopilotTemplatePath = join(libraryDir, 'root/copilot-instructions.template.md')
    if (fileExists(rootCopilotTemplatePath)) {
      addContent(
        '.github/copilot-instructions.md',
        'root/copilot-instructions.template.md',
        readFile(rootCopilotTemplatePath).replace(/\[YOUR_PROJECT_NAME\]/g, data.config.projectName),
      )
    }
  }

  for (const tool of data.config.tools) {
    if (tool === 'opencode' || tool === 'claude-code') {
      const targetSubdir = tool === 'opencode' ? '.opencode/agents' : '.claude/agents'
      const transform = tool === 'opencode' ? stripFrontmatterAndInjectModel : undefined
      addDir('agents', targetSubdir, transform)
    }
    addSkillsAndPromptsForTool(tool)
  }

  return expected
}

export function registerUpdate(program: Command): void {
  program
    .command('update')
    .description('Update ai-setup library files (preserves or prompts on conflicts)')
    .option('--force', 'Overwrite all existing managed files (creates backups)')
    .option('--check', 'Preview which skills are outdated without applying changes')
    .action(async (opts: UpdateOptions & { check?: boolean }) => {
      if (opts.check) {
        await runUpdateCheck()
        return
      }
      const targetDir = process.cwd()
      const tracker = new OperationTracker('update')
      const configPath = join(targetDir, '.ai-setup.json')

      if (!fileExists(configPath)) {
        throw Errors.manifestNotFound(targetDir)
      }

      const db = await createStore(targetDir)
      const data = db.data
      const expectedFiles = buildExpectedFiles(data, targetDir)
      const trackedByPath = new Map(data.files.map((record) => [record.path, record]))

      const updatedRecords = new Map<string, TrackedFile>()
      const newRecords: TrackedFile[] = []
      const missing: string[] = []
      const skipped: string[] = []

      let updatedCount = 0
      let addedCount = 0
      const now = new Date().toISOString()

      p.intro('Updating ai-setup managed files...')

      for (const entry of expectedFiles) {
        const absPath = join(targetDir, entry.path)
        const tracked = trackedByPath.get(entry.path)

        if (tracked && !fileExists(absPath)) {
          missing.push(entry.path)
          continue
        }

        const resolution = await resolveConflict(absPath, entry.path, {
          force: opts.force,
          trackedHash: tracked?.hash,
        })

        if (resolution === 'skip') {
          skipped.push(entry.path)
          continue
        }

        if (resolution === 'backup-and-overwrite') {
          const backupPath = backupFile(absPath, targetDir)
          tracker.registerBackup(absPath, backupPath)
        }

        try {
          writeFile(absPath, entry.content)
          tracker.trackSuccess(entry.path)
        } catch (error) {
          tracker.trackFailure(entry.path, error instanceof Error ? error.message : String(error))
          throw error
        }

        const nextRecord: TrackedFile = {
          path: entry.path,
          hash: fileHash(absPath),
          source: entry.source,
          owner: tracked?.owner ?? 'library',
          status: 'installed',
          installedAt: tracked?.installedAt ?? now,
          lastCheckedAt: now,
        }

        if (tracked) {
          updatedRecords.set(entry.path, nextRecord)
          updatedCount += 1
        } else {
          newRecords.push(nextRecord)
          addedCount += 1
        }
      }

      const nextFiles: TrackedFile[] = data.files.map((record) => updatedRecords.get(record.path) ?? record)
      for (const record of newRecords) {
        if (!nextFiles.some((existing) => existing.path === record.path)) {
          nextFiles.push(record)
        }
      }

      data.files = nextFiles
      await writeStore(targetDir, data)
      await appendOperation(targetDir, tracker.toOperation())

      p.log.success(`Updated ${updatedCount} tracked file(s)`)
      p.log.success(`Added ${addedCount} new file(s)`)

      if (missing.length > 0) {
        p.log.warn(`Skipped ${missing.length} missing tracked file(s):`)
        for (const file of missing) {
          p.log.message(`  - ${file}`)
        }
      }

      if (skipped.length > 0) {
        p.log.warn(`Skipped ${skipped.length} file(s):`)
        for (const file of skipped) {
          p.log.message(`  - ${file}`)
        }
      }

      p.outro('✅ Update completed')
    })
}

async function runUpdateCheck(): Promise<void> {
  const targetDir = process.cwd()
  p.intro(pc.bold('ai-setup update --check'))

  const results = await compareLibrarySkills(targetDir)

  if (results.length === 0) {
    p.log.info('No library skills found. Nothing to check.')
    p.outro(pc.dim('Run ai-setup init first to install skills.'))
    return
  }

  const drifted = results.filter((r) => r.status === 'drifted')
  const modified = results.filter((r) => r.status === 'modified')
  const missing = results.filter((r) => r.status === 'missing')
  const upgradable = [...drifted, ...modified, ...missing]

  if (upgradable.length === 0) {
    p.log.success('All skills are up-to-date. No updates needed.')
    p.outro(pc.green('✓ Nothing to update'))
    return
  }

  showSummaryBox('📋 Skills to update', [
    { label: 'Drifted (library newer)', value: pc.yellow(`${drifted.length}`) },
    { label: 'Modified (user changed)', value: pc.yellow(`${modified.length}`) },
    { label: 'Missing (not installed)', value: pc.red(`${missing.length}`) },
    { label: 'Total to update', value: pc.bold(`${upgradable.length}`) },
  ])

  console.log('')
  p.log.info('Skills that would be updated:')
  for (const skill of upgradable) {
    const icon = skill.status === 'drifted' ? '↻' : skill.status === 'modified' ? '~' : '✗'
    p.log.message(`  ${pc.yellow(icon)} ${skill.name} (${skill.status})`)
  }

  console.log('')
  showSummaryBox('💡 To apply', [
    { label: '1', value: `Run ${pc.cyan('ai-setup update --force')} to refresh all` },
    { label: '2', value: `Run ${pc.cyan('ai-setup compile')} to regenerate from library` },
  ])

  p.outro(pc.yellow(`⚠ ${upgradable.length} skill(s) would be updated`))
}
