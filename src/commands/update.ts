import type { Command } from 'commander'
import { readFileSync, writeFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import path, { dirname, join } from 'node:path'
import * as p from '@clack/prompts'
import type { AiSetupConfig, FileRecord } from '../types.js'
import { fileExists, fileHash, listDir, readFile, writeFile } from '../utils/files.js'

interface ExpectedFile {
  path: string
  source: string
  content: string
}

const __dirname = dirname(fileURLToPath(import.meta.url))
const libraryDir = join(__dirname, '../../library')

function buildExpectedFiles(config: AiSetupConfig, targetDir: string): ExpectedFile[] {
  const expected: ExpectedFile[] = []

  const addDir = (librarySubDir: string, targetSubDir: string): void => {
    const srcDir = join(libraryDir, librarySubDir)
    if (!fileExists(srcDir)) return
    for (const file of listDir(srcDir)) {
      const srcPath = join(srcDir, file)
      const targetPath = join(targetDir, targetSubDir, file)
      expected.push({
        path: path.relative(targetDir, targetPath),
        source: path.join(librarySubDir, file).replaceAll('\\', '/'),
        content: readFile(srcPath),
      })
    }
  }

  const addFile = (librarySource: string, targetPath: string): void => {
    const srcPath = join(libraryDir, librarySource)
    if (!fileExists(srcPath)) return
    expected.push({
      path: path.relative(targetDir, targetPath),
      source: librarySource,
      content: readFile(srcPath),
    })
  }

  addDir('templates', 'docs/templates')
  addDir('rules', 'docs/rules')
  addDir('context', 'docs/context')

  addFile('infra/CODEOWNERS.template', join(targetDir, 'CODEOWNERS'))
  addFile('infra/compliance.md', join(targetDir, 'docs/compliance.md'))
  addFile('infra/KNOWLEDGE_MAP.template.md', join(targetDir, 'docs/KNOWLEDGE_MAP.md'))

  if (fileExists(join(targetDir, '.git'))) {
    addFile('infra/pre-commit.hook', join(targetDir, '.git/hooks/pre-commit'))
  }

  const rootTemplatePath = join(libraryDir, 'root/AGENTS.template.md')
  if (fileExists(rootTemplatePath)) {
    const rootContent = readFile(rootTemplatePath).replace(/\[YOUR_PROJECT_NAME\]/g, config.projectName)
    if (config.tools.includes('opencode')) {
      expected.push({
        path: 'AGENTS.md',
        source: 'root/AGENTS.template.md',
        content: rootContent,
      })
    }
    if (config.tools.includes('pi')) {
      expected.push({
        path: 'CLAUDE.md',
        source: 'root/AGENTS.template.md',
        content: rootContent,
      })
    }
  }

  if (config.tools.includes('pi')) {
    addDir('agents', '.pi/agents')
    addDir('prompts', '.pi/templates')
  }

  if (config.tools.includes('opencode')) {
    addDir('agents', '.opencode/agents')
  }

  return expected
}

export function registerUpdate(program: Command): void {
  program
    .command('update')
    .description('Update ai-setup library files (skips customized files)')
    .action(() => {
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
      const customized: string[] = []
      const missing: string[] = []
      const conflicts: string[] = []

      let updatedCount = 0
      let addedCount = 0

      p.intro('Updating ai-setup managed files...')

      for (const entry of expectedFiles) {
        const absPath = join(targetDir, entry.path)
        const tracked = trackedByPath.get(entry.path)

        if (tracked) {
          if (!fileExists(absPath)) {
            missing.push(entry.path)
            continue
          }

          const currentHash = fileHash(absPath)
          if (currentHash !== tracked.hash) {
            customized.push(entry.path)
            continue
          }

          writeFile(absPath, entry.content)
          updatedRecords.set(entry.path, {
            path: entry.path,
            hash: fileHash(absPath),
            source: entry.source,
          })
          updatedCount += 1
          continue
        }

        if (fileExists(absPath)) {
          conflicts.push(entry.path)
          continue
        }

        writeFile(absPath, entry.content)
        newRecords.push({
          path: entry.path,
          hash: fileHash(absPath),
          source: entry.source,
        })
        addedCount += 1
      }

      const nextFiles: FileRecord[] = config.files.map((record) => updatedRecords.get(record.path) ?? record)
      for (const record of newRecords) {
        nextFiles.push(record)
      }

      config.files = nextFiles
      writeFileSync(configPath, JSON.stringify(config, null, 2), 'utf-8')

      p.log.success(`Updated ${updatedCount} tracked file(s)`)
      p.log.success(`Added ${addedCount} new file(s)`)

      if (customized.length > 0) {
        p.log.warn(`Skipped ${customized.length} customized file(s):`)
        for (const file of customized) {
          p.log.message(`  - ${file}`)
        }
      }

      if (missing.length > 0) {
        p.log.warn(`Skipped ${missing.length} missing tracked file(s):`)
        for (const file of missing) {
          p.log.message(`  - ${file}`)
        }
      }

      if (conflicts.length > 0) {
        p.log.warn(`Skipped ${conflicts.length} untracked existing file(s):`)
        for (const file of conflicts) {
          p.log.message(`  - ${file}`)
        }
      }

      p.outro('✅ Update completed')
    })
}
