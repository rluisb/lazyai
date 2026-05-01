import path from 'node:path'
import type { ConflictStrategy, FileRecord, ToolId } from '../types.js'
import { applyStrategy } from '../utils/conflict-strategy.js'
import { ensureDir, fileExists, fileHash, readFile, writeFile } from '../utils/files.js'
import { ROOT_FILE_BY_TOOL } from './root-file-map.js'

const DEPRECATED_ROOT_TEMPLATE_BY_FILE: Partial<Record<string, string>> = {
  'AGENTS.md': 'root/AGENTS.template.md',
  '.github/copilot-instructions.md': 'root/copilot-instructions.template.md',
}

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
 * @deprecated Root files are now always generated through scaffoldCompiledRoot.
 * This fallback is kept temporarily for reference and test coverage only.
 *
 * Behavior:
 * - For each tool in `tools`, reads template from `library/root/` and writes substituted content
 * - `opencode` → reads AGENTS.template.md, writes AGENTS.md
 * - `claude-code` → reads AGENTS.template.md (if exists), writes AGENTS.md
 * - `copilot` → reads copilot-instructions.template.md (if exists), ensures .github/, writes .github/copilot-instructions.md
 * - All templates replace [YOUR_PROJECT_NAME] with config.projectName
 * - Uses deterministic applyStrategy for conflict handling
 * - Records each written file in fileRecords
 * - If multiple tools share a root filename, the last processed tool wins for that file
 */
export async function scaffoldRootFiles(opts: ScaffoldRootFilesOptions): Promise<void> {
  const { targetDir, libraryDir, tools, projectName, fileRecords, strategy, perFileOverrides } = opts
  const rootDir = path.join(libraryDir, 'root')

  for (const tool of tools) {
    const outputName = ROOT_FILE_BY_TOOL[tool]
    const templateSource = DEPRECATED_ROOT_TEMPLATE_BY_FILE[outputName]
    if (!templateSource) continue

    const templatePath = path.join(rootDir, path.basename(templateSource))
    if (!fileExists(templatePath)) continue

    const template = readFile(templatePath)
    const content = template.replace(/\[YOUR_PROJECT_NAME\]/g, projectName)
    const destPath = path.join(targetDir, outputName)
    if (outputName.startsWith('.github/')) {
      ensureDir(path.dirname(destPath))
    }

    writeRootFile(
      destPath,
      content,
      fileRecords,
      targetDir,
      templateSource,
      strategy,
      perFileOverrides,
    )
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
    owner: 'library',
  })
}
