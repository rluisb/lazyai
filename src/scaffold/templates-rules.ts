import path from 'node:path'
import type { ConflictStrategy, FileRecord, RuleId, TemplateId } from '../types.js'
import { applyStrategy } from '../utils/conflict-strategy.js'
import * as files from '../utils/files.js'

export interface ScaffoldTemplatesRulesOptions {
  targetDir: string
  libraryDir: string
  templates: TemplateId[]
  rules: RuleId[]
  fileRecords: FileRecord[]
  strategy: ConflictStrategy
  perFileOverrides: Map<string, ConflictStrategy>
}

export async function scaffoldTemplatesRules(opts: ScaffoldTemplatesRulesOptions): Promise<void> {
  const { targetDir, libraryDir, templates, rules, fileRecords, strategy, perFileOverrides } = opts
  const docsDir = path.join(targetDir, 'docs')

  // Copy selected templates
  if (templates.length > 0) {
    const templatesDir = path.join(docsDir, 'templates')
    files.ensureDir(templatesDir)

    for (const templateId of templates) {
      const src = path.join(libraryDir, 'templates', `${templateId}.md`)
      const dest = path.join(templatesDir, `${templateId}.md`)
      await copyLibraryFile(src, dest, fileRecords, targetDir, libraryDir, strategy, perFileOverrides)
    }
  }

  // Copy selected rules
  if (rules.length > 0) {
    const rulesDir = path.join(docsDir, 'rules')
    files.ensureDir(rulesDir)

    for (const ruleId of rules) {
      const src = path.join(libraryDir, 'rules', `${ruleId}.md`)
      const dest = path.join(rulesDir, `${ruleId}.md`)
      await copyLibraryFile(src, dest, fileRecords, targetDir, libraryDir, strategy, perFileOverrides)
    }
  }

  // Always copy prompts/local-examples directory
  await copyLibraryDir(
    path.join(libraryDir, 'prompts/local-examples'),
    path.join(docsDir, 'prompts/local-examples'),
    fileRecords,
    targetDir,
    libraryDir,
    strategy,
    perFileOverrides
  )
}

async function copyLibraryFile(
  src: string,
  dest: string,
  records: FileRecord[],
  targetDir: string,
  libraryDir: string,
  strategy: ConflictStrategy,
  perFileOverrides: Map<string, ConflictStrategy>
): Promise<void> {
  if (!files.fileExists(src)) return

  const relPath = path.relative(targetDir, dest)
  const action = applyStrategy(dest, strategy, perFileOverrides, targetDir)

  if (action === 'skip') {
    console.warn(`⚠️  Skipping existing file: ${relPath}`)
    return
  }

  files.copyFile(src, dest)
  records.push({
    path: relPath,
    hash: files.fileHash(dest),
    source: path.relative(libraryDir, src),
  })
}

async function copyLibraryDir(
  src: string,
  dest: string,
  records: FileRecord[],
  targetDir: string,
  libraryDir: string,
  strategy: ConflictStrategy,
  perFileOverrides: Map<string, ConflictStrategy>
): Promise<void> {
  if (!files.fileExists(src)) return

  files.ensureDir(dest)
  const entries = files.listDir(src)

  for (const entry of entries) {
    const srcPath = path.join(src, entry)
    const destPath = path.join(dest, entry)

    if (files.isDirectory(srcPath)) {
      await copyLibraryDir(srcPath, destPath, records, targetDir, libraryDir, strategy, perFileOverrides)
      continue
    }

    await copyLibraryFile(srcPath, destPath, records, targetDir, libraryDir, strategy, perFileOverrides)
  }
}
