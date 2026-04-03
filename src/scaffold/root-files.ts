import path from 'node:path'
import { ensureDir, writeFile, fileHash, readFile, fileExists } from '../utils/files.js'
import { applyStrategy } from '../utils/conflict-strategy.js'
import type { ToolId, FileRecord, ConflictStrategy } from '../types.js'

export interface ScaffoldRootFilesOptions {
  targetDir: string
  libraryDir: string
  tools: ToolId[]
  projectName: string
  fileRecords: FileRecord[]
  strategy: ConflictStrategy
  perFileOverrides: Map<string, ConflictStrategy>
}

/**
 * Creates root-level AI tool configuration files.
 *
 * Behavior:
 * - For each tool in `tools`, reads template from `library/root/` and writes substituted content
 * - `opencode` → reads AGENTS.template.md, writes AGENTS.md
 * - `codex` → reads AGENTS.template.md, writes AGENTS.md
 * - `pi` → reads AGENTS.template.md, writes INSTRUCTIONS.md
 * - `claude-code` → reads CLAUDE.template.md (if exists), writes CLAUDE.md
 * - `gemini` → reads GEMINI.template.md (if exists), writes GEMINI.md
 * - `copilot` → reads copilot-instructions.template.md (if exists), ensures .github/, writes .github/copilot-instructions.md
 * - All templates replace [YOUR_PROJECT_NAME] with config.projectName
 * - Uses deterministic applyStrategy for conflict handling
 * - Records each written file in fileRecords
 * - If both pi and claude-code are selected, CLAUDE.md is written by the last one (claude-code wins)
 */
export async function scaffoldRootFiles(opts: ScaffoldRootFilesOptions): Promise<void> {
  const { targetDir, libraryDir, tools, projectName, fileRecords, strategy, perFileOverrides } = opts
  const rootDir = path.join(libraryDir, 'root')

  // Track if CLAUDE.md has been written (pi writes AGENTS.template, claude-code writes CLAUDE.template)
  let claudeWritten = false

  for (const tool of tools) {
    if (tool === 'opencode') {
      const templatePath = path.join(rootDir, 'AGENTS.template.md')
      if (fileExists(templatePath)) {
        const template = readFile(templatePath)
        const content = template.replace(/\[YOUR_PROJECT_NAME\]/g, projectName)
        writeRootFile(
          path.join(targetDir, 'AGENTS.md'),
          content,
          fileRecords,
          targetDir,
          'root/AGENTS.template.md',
          strategy,
          perFileOverrides
        )
      }
    }

    if (tool === 'codex') {
      // Codex uses AGENTS.md like opencode
      const templatePath = path.join(rootDir, 'AGENTS.template.md')
      if (fileExists(templatePath)) {
        const template = readFile(templatePath)
        const content = template.replace(/\[YOUR_PROJECT_NAME\]/g, projectName)
        writeRootFile(
          path.join(targetDir, 'AGENTS.md'),
          content,
          fileRecords,
          targetDir,
          'root/AGENTS.template.md',
          strategy,
          perFileOverrides
        )
      }
    }

    if (tool === 'pi') {
      const templatePath = path.join(rootDir, 'AGENTS.template.md')
      if (fileExists(templatePath)) {
        const template = readFile(templatePath)
        const content = template.replace(/\[YOUR_PROJECT_NAME\]/g, projectName)
        writeRootFile(
          path.join(targetDir, 'INSTRUCTIONS.md'),
          content,
          fileRecords,
          targetDir,
          'root/AGENTS.template.md',
          strategy,
          perFileOverrides
        )
        claudeWritten = true
      }
    }

    if (tool === 'claude-code') {
      const templatePath = path.join(rootDir, 'CLAUDE.template.md')
      if (fileExists(templatePath)) {
        const template = readFile(templatePath)
        const content = template.replace(/\[YOUR_PROJECT_NAME\]/g, projectName)
        writeRootFile(
          path.join(targetDir, 'CLAUDE.md'),
          content,
          fileRecords,
          targetDir,
          'root/CLAUDE.template.md',
          strategy,
          perFileOverrides
        )
        claudeWritten = true
      }
    }

    if (tool === 'gemini') {
      const templatePath = path.join(rootDir, 'GEMINI.template.md')
      if (fileExists(templatePath)) {
        const template = readFile(templatePath)
        const content = template.replace(/\[YOUR_PROJECT_NAME\]/g, projectName)
        writeRootFile(
          path.join(targetDir, 'GEMINI.md'),
          content,
          fileRecords,
          targetDir,
          'root/GEMINI.template.md',
          strategy,
          perFileOverrides
        )
      }
    }

    if (tool === 'copilot') {
      const templatePath = path.join(rootDir, 'copilot-instructions.template.md')
      if (fileExists(templatePath)) {
        const template = readFile(templatePath)
        const content = template.replace(/\[YOUR_PROJECT_NAME\]/g, projectName)
        ensureDir(path.join(targetDir, '.github'))
        writeRootFile(
          path.join(targetDir, '.github', 'copilot-instructions.md'),
          content,
          fileRecords,
          targetDir,
          'root/copilot-instructions.template.md',
          strategy,
          perFileOverrides
        )
      }
    }
  }
}

/**
 * Internal helper to write a root file with deterministic conflict strategy.
 */
function writeRootFile(
  dest: string,
  content: string,
  records: FileRecord[],
  targetDir: string,
  source: string,
  strategy: ConflictStrategy,
  perFileOverrides: Map<string, ConflictStrategy>
): void {
  const relPath = path.relative(targetDir, dest)
  const action = applyStrategy(dest, strategy, perFileOverrides, targetDir)
  if (action === 'skip') return
  writeFile(dest, content)
  records.push({
    path: relPath,
    hash: fileHash(dest),
    source,
  })
}
