import * as p from '@clack/prompts'
import { fileExists, fileHash } from './files.js'
import type { ConflictStrategy } from '../types.js'
import { Errors } from '../errors/index.js'

export type ConflictResolution = 'overwrite' | 'skip' | 'backup-and-overwrite'

export interface ConflictOptions {
  force?: boolean | undefined
  trackedHash?: string | undefined
  strategy?: ConflictStrategy | undefined
}

export async function resolveConflict(
  destPath: string,
  displayName: string,
  options: ConflictOptions = {}
): Promise<ConflictResolution> {
  if (!fileExists(destPath)) return 'overwrite'

  // Deterministic strategy from Phase 7 wizard (bypasses interactive prompts)
  if (options.strategy) {
    switch (options.strategy) {
      case 'skip':
        return 'skip'
      case 'backup-and-replace':
      case 'align':
        return 'backup-and-overwrite'
    }
  }

  if (options.force) {
    return 'backup-and-overwrite'
  }

  if (options.trackedHash) {
    const currentHash = fileHash(destPath)
    if (currentHash === options.trackedHash) {
      return 'overwrite'
    }

    const replaceCustomized = await p.confirm({
      message: `File ${displayName} has been customized. Overwrite? (backup will be created)`,
    })

    if (p.isCancel(replaceCustomized)) {
      p.cancel('Operation cancelled.')
      throw Errors.userCancelled()
    }

    return replaceCustomized ? 'backup-and-overwrite' : 'skip'
  }

  const replaceExisting = await p.confirm({
    message: `File ${displayName} already exists. Replace? (backup will be created)`,
  })

  if (p.isCancel(replaceExisting)) {
    p.cancel('Operation cancelled.')
    throw Errors.userCancelled()
  }

  return replaceExisting ? 'backup-and-overwrite' : 'skip'
}

/**
 * Backward-compatible wrapper around resolveConflict.
 */
export async function confirmReplace(filePath: string, displayName?: string): Promise<boolean> {
  const label = displayName || filePath
  const resolution = await resolveConflict(filePath, label)
  return resolution !== 'skip'
}
