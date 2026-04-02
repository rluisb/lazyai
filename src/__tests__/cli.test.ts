import { beforeEach, afterEach, describe, expect, it } from 'vitest'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { createProgram } from '../cli.js'
import { createImportCommand } from '../commands/import.js'
import { createMigrateCommand } from '../commands/migrate.js'
import type { AiSetupConfig } from '../types.js'

describe('cli init integration', () => {
  let originalCwd: string

  const makeTempRepo = (): string => {
    const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-init-'))
    fs.mkdirSync(path.join(tempDir, '.git'), { recursive: true })
    return tempDir
  }

  const runInit = async (args: string[]): Promise<void> => {
    const program = createProgram()
    await program.parseAsync(['node', 'ai-setup', 'init', ...args])
  }

  beforeEach(() => {
    originalCwd = process.cwd()
  })

  afterEach(() => {
    process.chdir(originalCwd)
  })

  it('runs full init and writes expected file tree', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    await runInit([
      '--scope',
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
      'docs/features/AGENTS.md',
      'docs/bugfixes/AGENTS.md',
      'docs/refactors/AGENTS.md',
      'docs/tech-debt/AGENTS.md',
      'docs/adrs/AGENTS.md',
      'docs/memory/AGENTS.md',
      'docs/standards/AGENTS.md',
      'docs/templates/AGENTS.md',
      'docs/rules/AGENTS.md',
      'AGENTS.md',
      'CLAUDE.md',
      '.git/hooks/pre-commit',
      '.pi/agents',
      '.pi/templates',
      '.pi/skills',
      '.opencode/agents',
      '.opencode/commands',
      '.opencode/templates',
      '.ai-setup.json',
    ]

    for (const rel of expectedPaths) {
      expect(fs.existsSync(path.join(tempDir, rel)), `${rel} should exist`).toBe(true)
    }

    const config = JSON.parse(fs.readFileSync(path.join(tempDir, '.ai-setup.json'), 'utf-8')) as any
    expect(config.config.projectName).toBe('integration-test')
    expect(config.config.setupScope).toBe('project')
    expect(config.config.tools).toEqual(['pi', 'opencode'])
    expect(config.files.length).toBeGreaterThan(20)
    expect(config.files.some((f: { path: string }) => f.path === '.pi/agents/builder.md')).toBe(true)
    expect(config.files.some((f: { path: string }) => f.path === '.opencode/agents/builder.md')).toBe(true)
    expect(config.files.some((f: { path: string }) => f.path === '.pi/skills/research.md')).toBe(true)
    expect(config.files.some((f: { path: string }) => f.path === '.opencode/commands/research.md')).toBe(true)
    expect(config.files.some((f: { path: string }) => f.path === '.opencode/templates/research.md')).toBe(true)
    expect(config.files.some((f: { path: string }) => f.path === 'docs/templates/task.md')).toBe(true)
    expect(config.files.some((f: { path: string }) => f.path === '.git/hooks/pre-commit')).toBe(true)
  })

  it('re-run with existing .ai-setup.json succeeds in non-interactive mode', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    const args = [
      '--scope',
      'project',
      '--tools',
      'pi,opencode',
      '--name',
      'rerun-test',
      '--no-interactive',
    ]

    await runInit(args)
    await runInit([...args, '--force'])

    const config = JSON.parse(fs.readFileSync(path.join(tempDir, '.ai-setup.json'), 'utf-8')) as any
    expect(config.config.projectName).toBe('rerun-test')
    expect(config.config.setupScope).toBe('project')
    expect(config.config.tools).toEqual(['pi', 'opencode'])
    expect(fs.existsSync(path.join(tempDir, 'AGENTS.md'))).toBe(true)
    expect(fs.existsSync(path.join(tempDir, '.opencode/agents'))).toBe(true)
  }, 15000)

  it('supports partial tool selection with --tools opencode', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    await runInit([
      '--scope',
      'project',
      '--tools',
      'opencode',
      '--name',
      'opencode-only-test',
      '--no-interactive',
    ])

    expect(fs.existsSync(path.join(tempDir, 'AGENTS.md'))).toBe(true)
    expect(fs.existsSync(path.join(tempDir, '.opencode/agents'))).toBe(true)
    expect(fs.existsSync(path.join(tempDir, '.opencode/commands'))).toBe(true)
    expect(fs.existsSync(path.join(tempDir, '.opencode/templates'))).toBe(true)

    expect(fs.existsSync(path.join(tempDir, 'CLAUDE.md'))).toBe(false)
    expect(fs.existsSync(path.join(tempDir, '.pi/agents'))).toBe(false)
    expect(fs.existsSync(path.join(tempDir, '.pi/skills'))).toBe(false)

    const config = JSON.parse(fs.readFileSync(path.join(tempDir, '.ai-setup.json'), 'utf-8')) as any
    expect(config.config.tools).toEqual(['opencode'])
    expect(config.files.some((f: { path: string }) => f.path.startsWith('.opencode/'))).toBe(true)
    expect(config.files.some((f: { path: string }) => f.path.startsWith('.pi/'))).toBe(false)
  })
})

describe('migration command options', () => {
  it('exposes --interactive on import command', () => {
    expect(createImportCommand().helpInformation()).toContain('--interactive')
  })

  it('exposes --interactive on migrate command', () => {
    expect(createMigrateCommand().helpInformation()).toContain('--interactive')
  })
})
