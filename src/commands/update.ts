import type { Command } from 'commander'
import { readFileSync, writeFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import path, { dirname, join } from 'node:path'
import * as p from '@clack/prompts'
import type { AiSetupConfig, FileRecord, ToolId } from '../types.js'
import { backupFile, fileExists, fileHash, listDir, readFile, writeFile } from '../utils/files.js'
import { resolveConflict } from '../utils/conflicts.js'

interface ExpectedFile {
  path: string
  source: string
  content: string
}

interface UpdateOptions {
  force?: boolean
}

const __dirname = dirname(fileURLToPath(import.meta.url))
const libraryDir = join(__dirname, '../library')

const docsAgentMappings: Array<{ source: string; destination: string }> = [
  { source: 'docs.md', destination: 'docs/AGENTS.md' },
  { source: 'features.md', destination: 'docs/features/AGENTS.md' },
  { source: 'bugfixes.md', destination: 'docs/bugfixes/AGENTS.md' },
  { source: 'refactors.md', destination: 'docs/refactors/AGENTS.md' },
  { source: 'tech-debt.md', destination: 'docs/tech-debt/AGENTS.md' },
  { source: 'rules.md', destination: 'docs/rules/AGENTS.md' },
  { source: 'standards.md', destination: 'docs/standards/AGENTS.md' },
  { source: 'templates.md', destination: 'docs/templates/AGENTS.md' },
  { source: 'memory.md', destination: 'docs/memory/AGENTS.md' },
  { source: 'adrs.md', destination: 'docs/adrs/AGENTS.md' },
]

function buildExpectedFiles(config: AiSetupConfig, targetDir: string): ExpectedFile[] {
  const expected: ExpectedFile[] = []

  const upsertExpected = (entry: ExpectedFile): void => {
    const index = expected.findIndex((item) => item.path === entry.path)
    if (index >= 0) {
      expected[index] = entry
      return
    }
    expected.push(entry)
  }

  const addDir = (librarySubDir: string, targetSubDir: string): void => {
    const srcDir = join(libraryDir, librarySubDir)
    if (!fileExists(srcDir)) return
    for (const file of listDir(srcDir)) {
      const srcPath = join(srcDir, file)
      const targetPath = join(targetDir, targetSubDir, file)
      upsertExpected({
        path: path.relative(targetDir, targetPath),
        source: path.join(librarySubDir, file).replaceAll('\\', '/'),
        content: readFile(srcPath),
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

  const addDocsAgents = (): void => {
    for (const mapping of docsAgentMappings) {
      addFile(`docs-agents/${mapping.source}`, join(targetDir, mapping.destination))
    }
  }

  const addSkillsAndPromptsForTool = (tool: ToolId): void => {
    const addSkill = (name: string, targetPath: string): void => {
      addFile(`skills/${name}.md`, join(targetDir, targetPath))
    }
    const addPrompt = (name: string, targetPath: string): void => {
      addFile(`prompts/${name}.md`, join(targetDir, targetPath))
    }

    if (tool === 'pi') {
      for (const name of ['research', 'plan', 'implement', 'iterate']) {
        addSkill(name, `.pi/skills/${name}.md`)
      }
      for (const name of ['research', 'plan', 'implement', 'compact', 'local-example']) {
        addPrompt(name, `.pi/templates/${name}.md`)
      }
    }

    if (tool === 'opencode') {
      for (const name of ['research', 'plan', 'implement', 'iterate']) {
        addSkill(name, `.opencode/commands/${name}.md`)
      }
      for (const name of ['research', 'plan', 'implement', 'compact', 'local-example']) {
        addPrompt(name, `.opencode/templates/${name}.md`)
      }
    }

    if (tool === 'claude-code') {
      for (const name of ['research', 'plan', 'implement', 'iterate']) {
        addSkill(name, `.claude/commands/${name}.md`)
      }
      for (const name of ['research', 'plan', 'implement', 'compact', 'local-example']) {
        addPrompt(name, `.claude/templates/${name}.md`)
      }
    }

    if (tool === 'gemini') {
      for (const name of ['research', 'plan', 'implement', 'iterate']) {
        addSkill(name, `.gemini/skills/${name}.md`)
      }
      for (const name of ['research', 'plan', 'implement', 'compact', 'local-example']) {
        addPrompt(name, `.gemini/templates/${name}.md`)
      }
    }

    if (tool === 'copilot') {
      for (const name of ['research', 'plan', 'implement', 'iterate']) {
        const source = `skills/${name}.md`
        const srcPath = join(libraryDir, source)
        if (fileExists(srcPath)) {
          const transformed = `---\nmode: agent\n---\n\n${readFile(srcPath)}`
          addContent(`.github/prompts/${name}.prompt.md`, source, transformed)
        }
      }
      for (const name of ['research', 'plan', 'implement', 'compact', 'local-example']) {
        addPrompt(name, `.github/templates/${name}.md`)
      }
    }
  }

  addDir('templates', 'docs/templates')
  addDir('rules', 'docs/rules')
  addDocsAgents()

  addFile('infra/CODEOWNERS.template', join(targetDir, 'CODEOWNERS'))
  addFile('infra/compliance.md', join(targetDir, 'docs/compliance.md'))
  addFile('infra/KNOWLEDGE_MAP.template.md', join(targetDir, 'docs/KNOWLEDGE_MAP.md'))

  if (fileExists(join(targetDir, '.git'))) {
    addFile('infra/pre-commit.hook', join(targetDir, '.git/hooks/pre-commit'))
  }

  const rootAgentsTemplatePath = join(libraryDir, 'root/AGENTS.template.md')
  if (fileExists(rootAgentsTemplatePath)) {
    const rootContent = readFile(rootAgentsTemplatePath).replace(/\[YOUR_PROJECT_NAME\]/g, config.projectName)
    if (config.tools.includes('opencode')) {
      addContent('AGENTS.md', 'root/AGENTS.template.md', rootContent)
    }
    if (config.tools.includes('pi') && !config.tools.includes('claude-code')) {
      addContent('CLAUDE.md', 'root/AGENTS.template.md', rootContent)
    }
  }

  if (config.tools.includes('claude-code')) {
    const rootClaudeTemplatePath = join(libraryDir, 'root/CLAUDE.template.md')
    if (fileExists(rootClaudeTemplatePath)) {
      addContent(
        'CLAUDE.md',
        'root/CLAUDE.template.md',
        readFile(rootClaudeTemplatePath).replace(/\[YOUR_PROJECT_NAME\]/g, config.projectName),
      )
    }
  }

  if (config.tools.includes('gemini')) {
    const rootGeminiTemplatePath = join(libraryDir, 'root/GEMINI.template.md')
    if (fileExists(rootGeminiTemplatePath)) {
      addContent(
        'GEMINI.md',
        'root/GEMINI.template.md',
        readFile(rootGeminiTemplatePath).replace(/\[YOUR_PROJECT_NAME\]/g, config.projectName),
      )
    }
  }

  if (config.tools.includes('copilot')) {
    const rootCopilotTemplatePath = join(libraryDir, 'root/copilot-instructions.template.md')
    if (fileExists(rootCopilotTemplatePath)) {
      addContent(
        '.github/copilot-instructions.md',
        'root/copilot-instructions.template.md',
        readFile(rootCopilotTemplatePath).replace(/\[YOUR_PROJECT_NAME\]/g, config.projectName),
      )
    }
  }

  for (const tool of config.tools) {
    addDir('agents', tool === 'pi' ? '.pi/agents' : tool === 'opencode' ? '.opencode/agents' : tool === 'claude-code' ? '.claude' : tool === 'gemini' ? '.gemini' : '.github')
    addSkillsAndPromptsForTool(tool)
  }

  return expected
}

export function registerUpdate(program: Command): void {
  program
    .command('update')
    .description('Update ai-setup library files (preserves or prompts on conflicts)')
    .option('--force', 'Overwrite all existing managed files (creates backups)')
    .action(async (opts: UpdateOptions) => {
      const targetDir = process.cwd()
      const configPath = join(targetDir, '.ai-setup.json')

      if (!fileExists(configPath)) {
        p.log.error('No .ai-setup.json found. Please run init first.')
        process.exit(1)
      }

      const config = JSON.parse(readFileSync(configPath, 'utf-8')) as AiSetupConfig
      const expectedFiles = buildExpectedFiles(config, targetDir)
      const trackedByPath = new Map(config.files.map((record) => [record.path, record]))

      const updatedRecords = new Map<string, FileRecord>()
      const newRecords: FileRecord[] = []
      const missing: string[] = []
      const skipped: string[] = []

      let updatedCount = 0
      let addedCount = 0

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
          backupFile(absPath, targetDir)
        }

        writeFile(absPath, entry.content)

        const nextRecord: FileRecord = {
          path: entry.path,
          hash: fileHash(absPath),
          source: entry.source,
        }

        if (tracked) {
          updatedRecords.set(entry.path, nextRecord)
          updatedCount += 1
        } else {
          newRecords.push(nextRecord)
          addedCount += 1
        }
      }

      const nextFiles: FileRecord[] = config.files.map((record) => updatedRecords.get(record.path) ?? record)
      for (const record of newRecords) {
        if (!nextFiles.some((existing) => existing.path === record.path)) {
          nextFiles.push(record)
        }
      }

      config.files = nextFiles
      writeFileSync(configPath, JSON.stringify(config, null, 2), 'utf-8')

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
