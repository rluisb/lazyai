import path from 'node:path'
import type { ConflictStrategy, FileRecord } from '../types.js'
import { applyStrategy } from '../utils/conflict-strategy.js'
import { ensureDir, fileExists, fileHash, readFile, writeFile } from '../utils/files.js'

export interface ScaffoldConstitutionOptions {
  targetDir: string
  libraryDir: string
  projectName: string
  fileRecords: FileRecord[]
  strategy: ConflictStrategy
  perFileOverrides: Map<string, ConflictStrategy>
}

export async function scaffoldConstitution(opts: ScaffoldConstitutionOptions): Promise<void> {
  const { targetDir, libraryDir, projectName, fileRecords, strategy, perFileOverrides } = opts
  const templatePath = path.join(libraryDir, 'constitution', 'constitution.template.md')
  if (!fileExists(templatePath)) return

  const dest = path.join(targetDir, '.specify', 'memory', 'constitution.md')
  const relPath = path.relative(targetDir, dest)
  ensureDir(path.dirname(dest))

  const action = applyStrategy(dest, strategy, perFileOverrides, targetDir)
  if (action === 'skip') {
    console.warn(`⚠️  Skipping existing file: ${relPath}`)
    return
  }

  const template = readFile(templatePath)
  const content = template.replace(/\[YOUR_PROJECT_NAME\]/g, projectName)
  writeFile(dest, content)
  fileRecords.push({
    path: relPath,
    hash: fileHash(dest),
    source: 'constitution/constitution.template.md',
    owner: 'library',
  })
}
