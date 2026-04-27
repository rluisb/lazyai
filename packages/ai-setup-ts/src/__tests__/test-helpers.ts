import fs from 'node:fs'
import path from 'node:path'

/**
 * Walk up from CWD until we find a directory named "library"
 * that contains "mcp/catalog.json" (the ai-setup library sentinel).
 *
 * Works in both:
 * - Monorepo layout (library symlink at repo root)
 * - Flat layout (library/ real directory)
 *
 * Use this in tests instead of hardcoding `path.resolve(process.cwd(), 'library')`.
 */
export function findMonorepoLibraryDir(): string {
  let dir = process.cwd()
  for (let i = 0; i < 20; i++) {
    const candidate = path.join(dir, 'library')
    if (fs.existsSync(path.join(candidate, 'mcp', 'catalog.json'))) {
      return candidate
    }
    const parent = path.dirname(dir)
    if (parent === dir) break
    dir = parent
  }
  throw new Error(
    `Could not find library directory (looked for library/mcp/catalog.json sentinel) from: ${process.cwd()}`,
  )
}
