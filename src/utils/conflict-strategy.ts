import type { ConflictStrategy } from '../types.js'
import { fileExists, backupFile } from './files.js'

/**
 * Deterministic conflict resolution — no interactive prompts.
 * Called during scaffold execution AFTER the wizard collected strategy in Phase 7.
 *
 * @returns 'write' if file should be written, 'skip' if it should be left alone
 */
export function applyStrategy(
  destPath: string,
  strategy: ConflictStrategy,
  perFileOverrides: Map<string, ConflictStrategy>,
  targetDir: string
): 'write' | 'skip' {
  const effectiveStrategy = perFileOverrides.get(destPath) ?? strategy

  const exists = fileExists(destPath)

  // New file — always write regardless of strategy
  if (!exists) return 'write'

  switch (effectiveStrategy) {
    case 'skip':
      return 'skip'
    case 'backup-and-replace':
      backupFile(destPath, targetDir)
      return 'write'
    case 'align':
      // For align strategy, also backup and write
      // (the diff preview already happened in Phase 7)
      backupFile(destPath, targetDir)
      return 'write'
    default: {
      const _exhaustive: never = effectiveStrategy
      return 'write'
    }
  }
}
