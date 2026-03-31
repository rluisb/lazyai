import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { dirname } from 'node:path'
import * as files from './utils/files.js'
import type { SetupConfig, AiSetupConfig, FileRecord } from './types.js'
import { AdapterRegistry } from './adapters/registry.js'
import { resolveConflict } from './utils/conflicts.js'

const __dirname = dirname(fileURLToPath(import.meta.url))
const libraryDir = path.join(__dirname, '../library')

export async function runScaffold(config: SetupConfig): Promise<void> {
  const targetDir = config.targetDir
  const fileRecords: FileRecord[] = []

  // 1. Create docs/ structure
  console.log('\n📂  Creating docs/ structure...')
  const docsDir = path.join(targetDir, 'docs')
  files.ensureDir(docsDir)
  files.ensureDir(path.join(docsDir, 'features'))
  files.ensureDir(path.join(docsDir, 'bugfixes'))
  files.ensureDir(path.join(docsDir, 'refactors'))
  files.ensureDir(path.join(docsDir, 'tech-debt'))
  files.ensureDir(path.join(docsDir, 'adrs'))
  files.ensureDir(path.join(docsDir, 'memory'))
  files.ensureDir(path.join(docsDir, 'memory/handoffs'))
  files.ensureDir(path.join(docsDir, 'prompts'))
  files.ensureDir(path.join(docsDir, 'standards'))
  files.ensureDir(path.join(docsDir, 'templates'))
  files.ensureDir(path.join(docsDir, 'rules'))

  // 2. Copy templates, rules, docs agents
  console.log('📄  Copying shared files...')
  await copyLibraryDir(path.join(libraryDir, 'templates'), path.join(docsDir, 'templates'), fileRecords, targetDir, config.force)
  await copyLibraryDir(path.join(libraryDir, 'rules'), path.join(docsDir, 'rules'), fileRecords, targetDir, config.force)
  await copyLibraryFile(path.join(libraryDir, 'docs-agents/docs.md'), path.join(docsDir, 'AGENTS.md'), fileRecords, targetDir, config.force)
  await copyLibraryFile(path.join(libraryDir, 'docs-agents/features.md'), path.join(docsDir, 'features/AGENTS.md'), fileRecords, targetDir, config.force)
  await copyLibraryFile(path.join(libraryDir, 'docs-agents/bugfixes.md'), path.join(docsDir, 'bugfixes/AGENTS.md'), fileRecords, targetDir, config.force)
  await copyLibraryFile(path.join(libraryDir, 'docs-agents/refactors.md'), path.join(docsDir, 'refactors/AGENTS.md'), fileRecords, targetDir, config.force)
  await copyLibraryFile(path.join(libraryDir, 'docs-agents/tech-debt.md'), path.join(docsDir, 'tech-debt/AGENTS.md'), fileRecords, targetDir, config.force)
  await copyLibraryFile(path.join(libraryDir, 'docs-agents/rules.md'), path.join(docsDir, 'rules/AGENTS.md'), fileRecords, targetDir, config.force)
  await copyLibraryFile(path.join(libraryDir, 'docs-agents/standards.md'), path.join(docsDir, 'standards/AGENTS.md'), fileRecords, targetDir, config.force)
  await copyLibraryFile(path.join(libraryDir, 'docs-agents/templates.md'), path.join(docsDir, 'templates/AGENTS.md'), fileRecords, targetDir, config.force)
  await copyLibraryFile(path.join(libraryDir, 'docs-agents/memory.md'), path.join(docsDir, 'memory/AGENTS.md'), fileRecords, targetDir, config.force)
  await copyLibraryFile(path.join(libraryDir, 'docs-agents/adrs.md'), path.join(docsDir, 'adrs/AGENTS.md'), fileRecords, targetDir, config.force)
  await copyLibraryDir(path.join(libraryDir, 'prompts/local-examples'), path.join(docsDir, 'prompts/local-examples'), fileRecords, targetDir, config.force)

  // 3. Copy infra
  console.log('🛠️  Copying infrastructure files...')
  await copyLibraryFile(path.join(libraryDir, 'infra/CODEOWNERS.template'), path.join(targetDir, 'CODEOWNERS'), fileRecords, targetDir, config.force)

  const hooksDir = path.join(targetDir, '.git/hooks')
  if (files.fileExists(path.join(targetDir, '.git'))) {
    files.ensureDir(hooksDir)
    await copyLibraryFile(path.join(libraryDir, 'infra/pre-commit.hook'), path.join(hooksDir, 'pre-commit'), fileRecords, targetDir, config.force)
  }

  await copyLibraryFile(path.join(libraryDir, 'infra/compliance.md'), path.join(docsDir, 'compliance.md'), fileRecords, targetDir, config.force)
  await copyLibraryFile(path.join(libraryDir, 'infra/KNOWLEDGE_MAP.template.md'), path.join(docsDir, 'KNOWLEDGE_MAP.md'), fileRecords, targetDir, config.force)

  // 4. Create root files (AGENTS.md, CLAUDE.md, GEMINI.md, etc.)
  console.log('📝  Creating root files...')
  const agentsTemplate = files.readFile(path.join(libraryDir, 'root/AGENTS.template.md'))
  const agentsContent = agentsTemplate.replace(/\[YOUR_PROJECT_NAME\]/g, config.projectName)

  if (config.tools.includes('opencode')) {
    await writeRootFile(path.join(targetDir, 'AGENTS.md'), agentsContent, fileRecords, targetDir, 'root/AGENTS.template.md', config.force)
  }
  if (config.tools.includes('pi')) {
    await writeRootFile(path.join(targetDir, 'CLAUDE.md'), agentsContent, fileRecords, targetDir, 'root/AGENTS.template.md', config.force)
  }
  if (config.tools.includes('claude-code')) {
    const claudeTemplatePath = path.join(libraryDir, 'root/CLAUDE.template.md')
    if (files.fileExists(claudeTemplatePath)) {
      const claudeTemplate = files.readFile(claudeTemplatePath)
      const claudeContent = claudeTemplate.replace(/\[YOUR_PROJECT_NAME\]/g, config.projectName)
      await writeRootFile(path.join(targetDir, 'CLAUDE.md'), claudeContent, fileRecords, targetDir, 'root/CLAUDE.template.md', config.force)
    }
  }
  if (config.tools.includes('gemini')) {
    const geminiTemplatePath = path.join(libraryDir, 'root/GEMINI.template.md')
    if (files.fileExists(geminiTemplatePath)) {
      const geminiTemplate = files.readFile(geminiTemplatePath)
      const geminiContent = geminiTemplate.replace(/\[YOUR_PROJECT_NAME\]/g, config.projectName)
      await writeRootFile(path.join(targetDir, 'GEMINI.md'), geminiContent, fileRecords, targetDir, 'root/GEMINI.template.md', config.force)
    }
  }
  if (config.tools.includes('copilot')) {
    const copilotTemplatePath = path.join(libraryDir, 'root/copilot-instructions.template.md')
    if (files.fileExists(copilotTemplatePath)) {
      const copilotTemplate = files.readFile(copilotTemplatePath)
      const copilotContent = copilotTemplate.replace(/\[YOUR_PROJECT_NAME\]/g, config.projectName)
      files.ensureDir(path.join(targetDir, '.github'))
      await writeRootFile(path.join(targetDir, '.github/copilot-instructions.md'), copilotContent, fileRecords, targetDir, 'root/copilot-instructions.template.md', config.force)
    }
  }

  // 5. Run Adapters
  const registry = new AdapterRegistry()
  const adapters = registry.getAll(config.tools)

  for (const adapter of adapters) {
      await adapter.install({
        targetDir,
        libraryDir,
        fileRecords,
        force: config.force,
      })
  }

  // 6. Write .ai-setup.json
  console.log('⚙️  Writing .ai-setup.json...')
  const aiSetupConfig: AiSetupConfig = {
    version: '0.1.0', // TODO: get from package.json
    setupType: config.setupType,
    tools: config.tools,
    projectName: config.projectName,
    installedAt: new Date().toISOString(),
    files: fileRecords,
  }
  files.writeFile(path.join(targetDir, '.ai-setup.json'), JSON.stringify(aiSetupConfig, null, 2))
}

async function copyLibraryDir(src: string, dest: string, records: FileRecord[], targetDir: string, force?: boolean): Promise<void> {
  if (!files.fileExists(src)) return

  files.ensureDir(dest)
  const entries = files.listDir(src)
  for (const entry of entries) {
    const srcPath = path.join(src, entry)
    const destPath = path.join(dest, entry)
    if (files.isDirectory(srcPath)) {
      await copyLibraryDir(srcPath, destPath, records, targetDir, force)
      continue
    }
    await copyLibraryFile(srcPath, destPath, records, targetDir, force)
  }
}

async function copyLibraryFile(src: string, dest: string, records: FileRecord[], targetDir: string, force?: boolean): Promise<void> {
  if (!files.fileExists(src)) return
  const relPath = path.relative(targetDir, dest)
  const resolution = await resolveConflict(dest, relPath, { force })

  if (resolution === 'skip') {
    console.warn(`⚠️  Skipping existing file: ${relPath}`)
    return
  }

  if (resolution === 'backup-and-overwrite') {
    files.backupFile(dest, targetDir)
  }

  files.copyFile(src, dest)
  records.push({
    path: relPath,
    hash: files.fileHash(dest),
    source: path.relative(libraryDir, src),
  })
}

async function writeRootFile(dest: string, content: string, records: FileRecord[], targetDir: string, source: string, force?: boolean): Promise<void> {
  const relPath = path.relative(targetDir, dest)
  const resolution = await resolveConflict(dest, relPath, { force })
  if (resolution === 'skip') return
  if (resolution === 'backup-and-overwrite') {
    files.backupFile(dest, targetDir)
  }
  files.writeFile(dest, content)
  records.push({
    path: relPath,
    hash: files.fileHash(dest),
    source,
  })
}
