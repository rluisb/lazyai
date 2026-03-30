import * as p from '@clack/prompts'
import { fileExists } from './files.js'

/**
 * Checks if a file exists at the given path and prompts the user to confirm replacement.
 * Returns true if the file doesn't exist (safe to write) or if the user confirms replacement.
 * Returns false if the file exists and the user declines replacement.
 */
export async function confirmReplace(filePath: string, displayName?: string): Promise<boolean> {
  if (!fileExists(filePath)) return true

  const label = displayName || filePath
  const replace = await p.confirm({
    message: `${label} already exists. Replace?`,
  })

  if (p.isCancel(replace)) {
    p.cancel('Operation cancelled.')
    process.exit(0)
  }

  return replace as boolean
}
