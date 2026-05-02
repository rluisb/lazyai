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
  coverageThreshold?: number
}

export async function scaffoldTemplatesRules(opts: ScaffoldTemplatesRulesOptions): Promise<void> {
  const { targetDir, libraryDir, templates, rules, fileRecords, strategy, perFileOverrides, coverageThreshold } = opts
  const specsDir = path.join(targetDir, 'specs')

  // Copy selected templates
  if (templates.length > 0) {
    const templatesDir = path.join(specsDir, 'templates')
    files.ensureDir(templatesDir)

    for (const templateId of templates) {
      const src = path.join(libraryDir, 'templates', `${templateId}.md`)
      const dest = path.join(templatesDir, `${templateId}.md`)
      await copyLibraryFile(src, dest, fileRecords, targetDir, libraryDir, strategy, perFileOverrides, coverageThreshold)
    }
  }

  // Copy selected rules
  if (rules.length > 0) {
    const rulesDir = path.join(specsDir, 'rules')
    files.ensureDir(rulesDir)

    for (const ruleId of rules) {
      const src = path.join(libraryDir, 'rules', `${ruleId}.md`)
      const dest = path.join(rulesDir, `${ruleId}.md`)
      await copyLibraryFile(src, dest, fileRecords, targetDir, libraryDir, strategy, perFileOverrides, coverageThreshold)
    }
  }

  // Always copy prompts/local-examples directory
  await copyLibraryDir(
    path.join(libraryDir, 'prompts/local-examples'),
    path.join(specsDir, 'prompts/local-examples'),
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
  perFileOverrides: Map<string, ConflictStrategy>,
  coverageThreshold?: number,
): Promise<void> {
  if (!files.fileExists(src)) return

  const relPath = path.relative(targetDir, dest)
  const action = applyStrategy(dest, strategy, perFileOverrides, targetDir)

  if (action === 'skip') {
    console.warn(`⚠️  Skipping existing file: ${relPath}`)
    return
  }

  files.writeFile(dest, applyTemplateRuleSubstitutions(files.readFile(src), coverageThreshold))
  records.push({
    path: relPath,
    hash: files.fileHash(dest),
    source: path.relative(libraryDir, src),
    owner: 'library',
  })
}

function applyTemplateRuleSubstitutions(content: string, coverageThreshold?: number): string {
  const value = coverageThreshold ?? 80
  const threshold = Number.isInteger(value) && value >= 1 && value <= 100 ? value : 80
  return content.replaceAll('{{COVERAGE_THRESHOLD}}', String(threshold))
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
