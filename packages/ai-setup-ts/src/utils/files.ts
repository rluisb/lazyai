import crypto from 'node:crypto'
import fs from 'node:fs'
import path from 'node:path'
import { Errors } from '../errors/index.js'


/**
 * Walk up from startDir until we find a directory containing package.json.
 * Works both when running from compiled dist/ and from TypeScript source src/.
 */
export function findPackageRoot(startDir: string): string {
  let dir = startDir
  while (true) {
    if (fs.existsSync(path.join(dir, 'package.json'))) return dir
    const parent = path.dirname(dir)
    if (parent === dir) throw Errors.dirNotFound(`Could not find package root from: ${startDir}`)
    dir = parent
  }
}

/**
 * Resolve the bundled library directory regardless of whether we are running
 * from compiled output, TypeScript source, or a monorepo layout.
 *
 * Walks up from `fromDir` looking for a directory named `library` that
 * contains the sentinel file `mcp/catalog.json`. The sentinel distinguishes
 * our library from unrelated `library/` directories (e.g. macOS `/Library`).
 *
 * Mirrors Go's `walkUpFromLibrary` in `internal/library/embed.go`.
 */
export function resolveLibraryDir(fromDir: string): string {
  let dir = fromDir
  for (let i = 0; i < 20; i++) {
    const candidate = path.join(dir, 'library')
    if (fs.existsSync(path.join(candidate, 'mcp', 'catalog.json'))) {
      return candidate
    }
    const parent = path.dirname(dir)
    if (parent === dir) break
    dir = parent
  }
  throw Errors.dirNotFound(`Could not find library directory from: ${fromDir}`)
}

export function ensureDir(dirPath: string): void {
  try {
    fs.mkdirSync(dirPath, { recursive: true })
  } catch (err) {
    throw Errors.filePermission(dirPath, `create directory (${(err as Error).message})`)
  }
}

export function fileExists(filePath: string): boolean {
  try {
    fs.accessSync(filePath, fs.constants.F_OK)
    return true
  } catch {
    return false
  }
}

export function readFile(filePath: string): string {
  try {
    return fs.readFileSync(filePath, 'utf-8')
  } catch (err) {
    throw Errors.filePermission(filePath, `read (${(err as Error).message})`)
  }
}

export function writeFile(filePath: string, content: string): void {
  try {
    ensureDir(path.dirname(filePath))
    fs.writeFileSync(filePath, content, 'utf-8')
  } catch (err) {
    throw Errors.filePermission(filePath, `write (${(err as Error).message})`)
  }
}

export function copyFile(src: string, dest: string): void {
  try {
    ensureDir(path.dirname(dest))
    fs.copyFileSync(src, dest)
  } catch (err) {
    throw Errors.unknown(`copy failed: ${src} → ${dest} (${(err as Error).message})`)
  }
}

export function copyDir(src: string, dest: string): void {
  if (!fileExists(src)) {
    throw Errors.dirNotFound(src)
  }
  ensureDir(dest)
  const entries = fs.readdirSync(src, { withFileTypes: true })
  for (const entry of entries) {
    const srcPath = path.join(src, entry.name)
    const destPath = path.join(dest, entry.name)
    if (entry.isDirectory()) {
      copyDir(srcPath, destPath)
    } else {
      copyFile(srcPath, destPath)
    }
  }
}

export function fileHash(filePath: string): string {
  try {
    const content = fs.readFileSync(filePath)
    return crypto.createHash('sha256').update(content).digest('hex').slice(0, 16)
  } catch (err) {
    throw Errors.filePermission(filePath, `hash (${(err as Error).message})`)
  }
}

export function listDir(dirPath: string): string[] {
  try {
    return fs.readdirSync(dirPath)
  } catch {
    return []
  }
}

export function isDirectory(filePath: string): boolean {
  try {
    return fs.statSync(filePath).isDirectory()
  } catch {
    return false
  }
}

export function backupFile(filePath: string, targetDir: string): string {
  const backupRoot = path.join(targetDir, '.ai-setup-backup')
  ensureDir(backupRoot)

  let relativePath = path.relative(targetDir, filePath)
  if (relativePath.startsWith('..') || path.isAbsolute(relativePath)) {
    relativePath = path.basename(filePath)
  }

  const normalizedRelativePath = relativePath.replaceAll('\\', '/')
  let backupPath = path.join(backupRoot, normalizedRelativePath)
  if (fileExists(backupPath)) {
    backupPath = `${backupPath}.${Date.now()}`
  }

  ensureDir(path.dirname(backupPath))
  fs.copyFileSync(filePath, backupPath)

  const backupRelative = path.relative(backupRoot, backupPath).replaceAll('\\', '/')
  console.log(`📦 Backed up: ${normalizedRelativePath} → .ai-setup-backup/${backupRelative}`)

  return backupPath
}
