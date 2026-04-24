import { existsSync } from 'node:fs'
import path from 'node:path'

/**
 * Walk up from `process.cwd()` (which is `packages/ai-setup-ts/` when vitest
 * runs) until we find the sentinel `library/mcp/catalog.json`. Returns the
 * absolute path to the repo's canonical `library/` directory.
 *
 * Replaces pre-monorepo assumptions that `library/` lived at `process.cwd()`.
 */
export function findMonorepoLibraryDir(): string {
  let dir = process.cwd()
  for (let i = 0; i < 20; i++) {
    const candidate = path.join(dir, 'library')
    if (existsSync(path.join(candidate, 'mcp', 'catalog.json'))) {
      return candidate
    }
    const parent = path.dirname(dir)
    if (parent === dir) break
    dir = parent
  }
  throw new Error(`Could not find library directory from: ${process.cwd()}`)
}
