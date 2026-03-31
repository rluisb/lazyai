import { beforeEach, describe, expect, it, vi } from 'vitest'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import {
  copyDir,
  copyFile,
  ensureDir,
  fileExists,
  fileHash,
  listDir,
  readFile,
  writeFile,
} from '../utils/files.js'
import { PiAdapter } from '../adapters/pi.js'
import { OpenCodeAdapter } from '../adapters/opencode.js'
import type { FileRecord } from '../types.js'

function makeTempDir(prefix: string): string {
  return fs.mkdtempSync(path.join(os.tmpdir(), prefix))
}

describe('file utils', () => {
  it('creates directories and writes/reads files', () => {
    const tempDir = makeTempDir('ai-setup-files-')
    const nested = path.join(tempDir, 'a/b/c')
    const filePath = path.join(nested, 'note.txt')

    ensureDir(nested)
    expect(fileExists(nested)).toBe(true)

    writeFile(filePath, 'hello world')
    expect(readFile(filePath)).toBe('hello world')
    expect(fileExists(filePath)).toBe(true)
  })

  it('copies files and directories recursively', () => {
    const tempDir = makeTempDir('ai-setup-copy-')
    const src = path.join(tempDir, 'src')
    const dest = path.join(tempDir, 'dest')
    ensureDir(path.join(src, 'nested'))
    writeFile(path.join(src, 'root.txt'), 'root')
    writeFile(path.join(src, 'nested/child.txt'), 'child')

    copyDir(src, dest)
    expect(readFile(path.join(dest, 'root.txt'))).toBe('root')
    expect(readFile(path.join(dest, 'nested/child.txt'))).toBe('child')

    const copiedSingle = path.join(tempDir, 'single/one.txt')
    copyFile(path.join(src, 'root.txt'), copiedSingle)
    expect(readFile(copiedSingle)).toBe('root')
  })

  it('hashes deterministically and listDir handles missing directories', () => {
    const tempDir = makeTempDir('ai-setup-hash-')
    const filePath = path.join(tempDir, 'hash.txt')
    writeFile(filePath, 'same content')

    const hash1 = fileHash(filePath)
    const hash2 = fileHash(filePath)
    expect(hash1).toBe(hash2)
    expect(hash1).toHaveLength(16)

    writeFile(filePath, 'different content')
    expect(fileHash(filePath)).not.toBe(hash1)

    expect(listDir(path.join(tempDir, 'does-not-exist'))).toEqual([])
  })

  it('throws when copying from a missing source directory', () => {
    const tempDir = makeTempDir('ai-setup-copy-error-')
    expect(() => copyDir(path.join(tempDir, 'missing'), path.join(tempDir, 'dest'))).toThrow(
      'Source directory does not exist',
    )
  })
})

describe('tool adapters', () => {
  let libraryDir: string
  let targetDir: string
  let fileRecords: FileRecord[]

  beforeEach(() => {
    libraryDir = makeTempDir('ai-setup-library-')
    targetDir = makeTempDir('ai-setup-target-')
    fileRecords = []

    ensureDir(path.join(libraryDir, 'agents'))
    ensureDir(path.join(libraryDir, 'prompts'))
    ensureDir(path.join(libraryDir, 'tool-agents'))
    writeFile(path.join(libraryDir, 'agents/builder.md'), '# builder')
    writeFile(path.join(libraryDir, 'agents/reviewer.md'), '# reviewer')
    writeFile(path.join(libraryDir, 'prompts/plan.md'), '# plan')
    writeFile(path.join(libraryDir, 'tool-agents/agents-dir.md'), '# agents context')
    writeFile(path.join(libraryDir, 'tool-agents/skills-dir.md'), '# skills context')
    writeFile(path.join(libraryDir, 'tool-agents/templates-dir.md'), '# templates context')
    writeFile(path.join(libraryDir, 'tool-agents/root-dir.md'), '# root context')
  })

  it('Pi adapter installs agents/templates and records metadata', async () => {
    const adapter = new PiAdapter()

    await adapter.install({ targetDir, libraryDir, fileRecords })

    expect(fileExists(path.join(targetDir, '.pi/agents/builder.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.pi/agents/reviewer.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.pi/templates/plan.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.pi/skills'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.pi/agents/AGENTS.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.pi/skills/AGENTS.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.pi/templates/AGENTS.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.pi/AGENTS.md'))).toBe(true)

    expect(fileRecords.map((f) => f.path).sort()).toEqual([
      '.pi/AGENTS.md',
      '.pi/agents/AGENTS.md',
      '.pi/agents/builder.md',
      '.pi/agents/reviewer.md',
      '.pi/skills/AGENTS.md',
      '.pi/templates/AGENTS.md',
      '.pi/templates/plan.md',
    ])
    expect(fileRecords.every((f) => f.hash.length === 16)).toBe(true)
  })

  it('OpenCode adapter installs agents and force-overwrites existing files with backup', async () => {
    const existingPath = path.join(targetDir, '.opencode/agents/builder.md')
    ensureDir(path.dirname(existingPath))
    writeFile(existingPath, 'customized')

    const adapter = new OpenCodeAdapter()
    await adapter.install({ targetDir, libraryDir, fileRecords, force: true })

    expect(readFile(existingPath)).toBe('# builder')
    expect(fileExists(path.join(targetDir, '.opencode/agents/reviewer.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.opencode/commands'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.opencode/agents/AGENTS.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.opencode/commands/AGENTS.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.opencode/templates/AGENTS.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.opencode/AGENTS.md'))).toBe(true)
    expect(fileExists(path.join(targetDir, '.ai-setup-backup/.opencode/agents/builder.md'))).toBe(true)

    expect(fileRecords.map((f) => f.path).sort()).toEqual([
      '.opencode/AGENTS.md',
      '.opencode/agents/AGENTS.md',
      '.opencode/agents/builder.md',
      '.opencode/agents/reviewer.md',
      '.opencode/commands/AGENTS.md',
      '.opencode/templates/AGENTS.md',
      '.opencode/templates/plan.md',
    ])
  })
})
