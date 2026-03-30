import { beforeEach, afterEach, describe, expect, it } from 'vitest'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { createProgram } from '../cli.js'
import type { AiSetupConfig } from '../types.js'

describe('cli init integration', () => {
  let originalCwd: string

  beforeEach(() => {
    originalCwd = process.cwd()
  })

  afterEach(() => {
    process.chdir(originalCwd)
  })

  it('runs full init and writes expected file tree', async () => {
    const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-init-'))
    fs.mkdirSync(path.join(tempDir, '.git'), { recursive: true })
    process.chdir(tempDir)

    const program = createProgram()
    await program.parseAsync([
      'node',
      'ai-setup',
      'init',
      '--type',
      'project',
      '--tools',
      'pi,opencode',
      '--name',
      'integration-test',
      '--no-interactive',
    ])

    const expectedPaths = [
      'docs/features',
      'docs/bugfixes',
      'docs/refactors',
      'docs/tech-debt',
      'docs/adrs',
      'docs/memory',
      'docs/standards',
      'docs/templates',
      'docs/rules',
      'docs/context',
      'CODEOWNERS',
      'AGENTS.md',
      'CLAUDE.md',
      '.git/hooks/pre-commit',
      '.pi/agents',
      '.pi/templates',
      '.pi/skills',
      '.opencode/agents',
      '.opencode/commands',
      '.ai-setup.json',
    ]

    for (const rel of expectedPaths) {
      expect(fs.existsSync(path.join(tempDir, rel)), `${rel} should exist`).toBe(true)
    }

    const config = JSON.parse(fs.readFileSync(path.join(tempDir, '.ai-setup.json'), 'utf-8')) as AiSetupConfig
    expect(config.projectName).toBe('integration-test')
    expect(config.setupType).toBe('project')
    expect(config.tools).toEqual(['pi', 'opencode'])
    expect(config.files.length).toBeGreaterThan(20)
    expect(config.files.some((f) => f.path === '.pi/agents/builder.md')).toBe(true)
    expect(config.files.some((f) => f.path === '.opencode/agents/builder.md')).toBe(true)
    expect(config.files.some((f) => f.path === 'docs/templates/prd.md')).toBe(true)
    expect(config.files.some((f) => f.path === '.git/hooks/pre-commit')).toBe(true)
  })
})
