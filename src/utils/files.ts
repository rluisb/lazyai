import fs from 'node:fs'
import path from 'node:path'
import crypto from 'node:crypto'

export function ensureDir(dirPath: string): void {
  try {
    fs.mkdirSync(dirPath, { recursive: true })
  } catch (err) {
    throw new Error(`Cannot create directory ${dirPath}: ${(err as Error).message}`)
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
    throw new Error(`Cannot read ${filePath}: ${(err as Error).message}`)
  }
}

export function writeFile(filePath: string, content: string): void {
  try {
    ensureDir(path.dirname(filePath))
    fs.writeFileSync(filePath, content, 'utf-8')
  } catch (err) {
    throw new Error(`Cannot write to ${filePath}: ${(err as Error).message}`)
  }
}

export function copyFile(src: string, dest: string): void {
  try {
    ensureDir(path.dirname(dest))
    fs.copyFileSync(src, dest)
  } catch (err) {
    throw new Error(`Cannot copy ${src} → ${dest}: ${(err as Error).message}`)
  }
}

export function copyDir(src: string, dest: string): void {
  if (!fileExists(src)) {
    throw new Error(`Source directory does not exist: ${src}`)
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
    throw new Error(`Cannot hash ${filePath}: ${(err as Error).message}`)
  }
}

export function listDir(dirPath: string): string[] {
  try {
    return fs.readdirSync(dirPath)
  } catch {
    return []
  }
}
