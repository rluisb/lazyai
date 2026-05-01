import { existsSync, mkdirSync, mkdtempSync, readdirSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { scaffoldOrchestration } from '../scaffold/orchestration.js'
import type { FileRecord } from '../types.js'
import { findMonorepoLibraryDir } from './test-helpers.js'

const libraryDir = findMonorepoLibraryDir()

describe('scaffoldOrchestration', () => {
  let tempDir: string
  let fileRecords: FileRecord[]

  beforeEach(() => {
    tempDir = mkdtempSync(path.join(tmpdir(), 'ai-setup-orchestration-'))
    fileRecords = []
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
  })

  it('copies the orchestration library tree into .ai/orchestration', async () => {
    await scaffoldOrchestration({
      targetDir: tempDir,
      libraryDir,
      fileRecords,
      strategy: 'skip',
      perFileOverrides: new Map(),
    })

    expect(existsSync(path.join(tempDir, '.ai', 'orchestration', 'chains', 'feature.json'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.ai', 'orchestration', 'teams', 'review-team.json'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.ai', 'orchestration', 'workflows', 'rpi.json'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.ai', 'orchestration', 'skills', 'domains', 'backend.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.ai', 'orchestration', 'skills', 'modes', 'senior.md'))).toBe(true)
    expect(fileRecords.some((record) => record.path === '.ai/orchestration/chains/feature.json')).toBe(true)
  })

  it('scaffolds the expected top-level orchestration directories', async () => {
    await scaffoldOrchestration({
      targetDir: tempDir,
      libraryDir,
      fileRecords,
      strategy: 'skip',
      perFileOverrides: new Map(),
    })

    const entries = readdirSync(path.join(tempDir, '.ai', 'orchestration')).sort()
    expect(entries).toEqual(['chains', 'skills', 'teams', 'workflows'])
  })

  it('adds orchestration content from discovered extensions', async () => {
    const extensionDir = path.join(tempDir, '.ai', 'extensions', 'team-pack')
    mkdirSync(path.join(extensionDir, 'skills'), { recursive: true })
    mkdirSync(path.join(extensionDir, 'orchestration', 'chains'), { recursive: true })
    mkdirSync(path.join(extensionDir, 'orchestration', 'workflows'), { recursive: true })
    writeFileSync(path.join(extensionDir, 'skills', 'custom-skill.md'), '# Custom Skill')
    writeFileSync(path.join(extensionDir, 'orchestration', 'chains', 'release.json'), '{"name":"release"}')
    writeFileSync(path.join(extensionDir, 'orchestration', 'workflows', 'deploy.json'), '{"name":"deploy"}')

    await scaffoldOrchestration({
      targetDir: tempDir,
      libraryDir,
      fileRecords,
      strategy: 'skip',
      perFileOverrides: new Map(),
    })

    expect(existsSync(path.join(tempDir, '.ai', 'orchestration', 'chains', 'feature.json'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.ai', 'orchestration', 'chains', 'release.json'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.ai', 'orchestration', 'workflows', 'deploy.json'))).toBe(true)
    expect(fileRecords.some((record) => record.path === '.ai/orchestration/chains/release.json')).toBe(true)
  })
})
