import * as p from '@clack/prompts'
import { Errors } from '../errors/index.js'
import type { ConflictStrategy } from '../types.js'
import { computeLineDiff, renderDiffPreview } from '../utils/diff.js'
import { fileExists, readFile } from '../utils/files.js'

export interface PlannedFile {
  destPath: string
  srcContent: string
}

export interface Phase7Result {
  strategy: ConflictStrategy
  perFileOverrides: Map<string, ConflictStrategy>
}

function cancelAndExit(): never {
  p.cancel('Setup cancelled.')
  throw Errors.userCancelled()
}

export async function runPhase3(opts: {
  interactive: boolean
  force?: boolean
  targetDir: string
  plannedFiles: PlannedFile[]
}): Promise<Phase7Result> {
  if (!opts.interactive) {
    if (opts.force) {
      return { strategy: 'backup-and-replace', perFileOverrides: new Map() }
    }

    return { strategy: 'skip', perFileOverrides: new Map() }
  }

  const conflictingFiles = opts.plannedFiles.filter(file => fileExists(file.destPath))

  if (conflictingFiles.length === 0) {
    p.note('No conflicts found — all files are new.')
    return { strategy: 'skip', perFileOverrides: new Map() }
  }

  const globalStrategy = await p.select({
    message: `Found ${conflictingFiles.length} conflicting file(s). How should conflicts be handled?`,
    options: [
      {
        value: 'align',
        label: 'align',
        hint: 'Review each conflict and choose per file',
      },
      {
        value: 'backup-and-replace',
        label: 'backup-and-replace',
        hint: 'Create backups and overwrite all conflicts',
      },
      {
        value: 'skip',
        label: 'skip',
        hint: 'Keep existing files and skip conflicts',
      },
    ],
  })

  if (p.isCancel(globalStrategy)) {
    cancelAndExit()
  }

  const strategy = globalStrategy as ConflictStrategy
  const perFileOverrides = new Map<string, ConflictStrategy>()

  if (strategy !== 'align') {
    return { strategy, perFileOverrides }
  }

  for (const file of conflictingFiles) {
    const existingContent = readFile(file.destPath)
    const diff = computeLineDiff(existingContent, file.srcContent)
    const preview = renderDiffPreview(diff)

    p.note(preview || '  (no visible changes)', `Diff preview: ${file.destPath}`)

    const fileStrategy = await p.select({
      message: `Conflict strategy for ${file.destPath}?`,
      options: [
        {
          value: 'align',
          label: 'align',
          hint: 'Keep for manual merge/review workflow',
        },
        {
          value: 'backup-and-replace',
          label: 'backup-and-replace',
          hint: 'Backup existing file, then overwrite',
        },
        {
          value: 'skip',
          label: 'skip',
          hint: 'Keep existing file untouched',
        },
      ],
    })

    if (p.isCancel(fileStrategy)) {
      cancelAndExit()
    }

    perFileOverrides.set(file.destPath, fileStrategy as ConflictStrategy)
  }

  return {
    strategy,
    perFileOverrides,
  }
}
