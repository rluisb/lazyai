import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { run } from '../cli.js'

describe('e2e project workflows', () => {
  let originalCwd: string

  const makeTempRepo = (): string => {
    const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-e2e-'))
    fs.mkdirSync(path.join(tempDir, '.git'), { recursive: true })
    return tempDir
  }

  const runCli = async (...args: string[]): Promise<void> => {
    await run(['node', 'ai-setup', ...args])
  }

  beforeEach(() => {
    originalCwd = process.cwd()
  })

  afterEach(() => {
    process.chdir(originalCwd)
  })

  it('runs full project init + compile cycle for opencode', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    await runCli(
      'init',
      '--scope',
      'project',
      '--tools',
      'opencode',
      '--name',
      'e2e-opencode',
      '--no-interactive',
    )

    expect(fs.existsSync(path.join(tempDir, '.ai'))).toBe(true)
    expect(fs.existsSync(path.join(tempDir, '.ai', 'constitution'))).toBe(true)
    expect(fs.existsSync(path.join(tempDir, '.opencode', 'agents', 'builder.md'))).toBe(true)
    expect(fs.existsSync(path.join(tempDir, '.opencode', 'skills', 'research', 'SKILL.md'))).toBe(true)

    fs.rmSync(path.join(tempDir, '.opencode', 'agents', 'builder.md'))
    expect(fs.existsSync(path.join(tempDir, '.opencode', 'agents', 'builder.md'))).toBe(false)

    await runCli('compile', '--tools', 'opencode', '--force')

    expect(fs.existsSync(path.join(tempDir, '.opencode', 'agents', 'builder.md'))).toBe(true)
  }, 20000)

  it('runs project init with opencode + claude-code outputs', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    await runCli(
      'init',
      '--scope',
      'project',
      '--tools',
      'opencode,claude-code',
      '--name',
      'e2e-multi-tool',
      '--no-interactive',
    )

    expect(fs.existsSync(path.join(tempDir, '.opencode'))).toBe(true)
    expect(fs.existsSync(path.join(tempDir, '.claude'))).toBe(true)
    expect(fs.existsSync(path.join(tempDir, 'AGENTS.md'))).toBe(true)
    expect(fs.existsSync(path.join(tempDir, 'CLAUDE.md'))).toBe(true)
  }, 20000)
})
