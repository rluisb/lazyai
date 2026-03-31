import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { applyStrategy } from '../utils/conflict-strategy.js'
import { mkdtempSync, writeFileSync, rmSync, existsSync } from 'node:fs'
import path from 'node:path'
import { tmpdir } from 'node:os'
import type { ConflictStrategy } from '../types.js'

describe('applyStrategy', () => {
  let tempDir: string

  beforeEach(() => {
    tempDir = mkdtempSync(path.join(tmpdir(), 'conflict-test-'))
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
  })

  it('returns write for new file regardless of strategy', () => {
    const dest = path.join(tempDir, 'new-file.md')
    const result = applyStrategy(dest, 'skip', new Map(), tempDir)
    expect(result).toBe('write')
  })

  it('returns skip for existing file with skip strategy', () => {
    const dest = path.join(tempDir, 'existing.md')
    writeFileSync(dest, 'existing content')
    const result = applyStrategy(dest, 'skip', new Map(), tempDir)
    expect(result).toBe('skip')
  })

  it('returns write and creates backup for backup-and-replace strategy', () => {
    const dest = path.join(tempDir, 'existing.md')
    writeFileSync(dest, 'existing content')

    const result = applyStrategy(dest, 'backup-and-replace', new Map(), tempDir)

    expect(result).toBe('write')
    const backupDir = path.join(tempDir, '.ai-setup-backup')
    expect(existsSync(backupDir)).toBe(true)
  })

  it('returns write and creates backup for align strategy with existing file', () => {
    const dest = path.join(tempDir, 'existing.md')
    writeFileSync(dest, 'existing content')

    const result = applyStrategy(dest, 'align', new Map(), tempDir)

    expect(result).toBe('write')
    const backupDir = path.join(tempDir, '.ai-setup-backup')
    expect(existsSync(backupDir)).toBe(true)
  })

  it('uses per-file override over global strategy', () => {
    const dest = path.join(tempDir, 'override.md')
    writeFileSync(dest, 'existing content')

    const overrides = new Map<string, ConflictStrategy>()
    overrides.set(dest, 'skip')

    const result = applyStrategy(dest, 'backup-and-replace', overrides, tempDir)
    expect(result).toBe('skip')
  })
})
