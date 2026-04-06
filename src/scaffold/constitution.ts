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

const CONSTITUTION_FILES = [
  'constitution',
  'constraints',
  'quality-gates',
  'uncertainty',
]

export async function scaffoldConstitution(opts: ScaffoldConstitutionOptions): Promise<void> {
  const { targetDir, libraryDir, projectName, fileRecords, strategy, perFileOverrides } = opts
  const constitutionDir = path.join(targetDir, '.ai', 'constitution')
  ensureDir(constitutionDir)

  for (const file of CONSTITUTION_FILES) {
    const templatePath = path.join(libraryDir, 'constitution', `${file}.template.md`)
    if (!fileExists(templatePath)) continue

    const template = readFile(templatePath)
    const content = template.replace(/\[YOUR_PROJECT_NAME\]/g, projectName)
    const dest = path.join(constitutionDir, `${file}.md`)
    const relPath = path.relative(targetDir, dest)
    const action = applyStrategy(dest, strategy, perFileOverrides, targetDir)

    if (action === 'skip') {
      console.warn(`⚠️  Skipping existing file: ${relPath}`)
      continue
    }

    writeFile(dest, content)
    fileRecords.push({
      path: relPath,
      hash: fileHash(dest),
      source: `constitution/${file}.template.md`,
      owner: 'library',
    })
  }
}
