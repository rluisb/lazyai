import { beforeEach, afterEach, describe, expect, it } from 'vitest'
import { vi } from 'vitest'
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

  const runCompile = async (args: string[] = []): Promise<void> => {
    const program = createProgram()
    await program.parseAsync(['node', 'ai-setup', 'compile', ...args])
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

  it('compile restores tool files from store without modifying store file', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    await runInit([
      '--scope',
      'project',
      '--tools',
      'opencode,claude-code',
      '--name',
      'compile-project-test',
      '--no-interactive',
    ])

    const opencodeAgent = path.join(tempDir, '.opencode/agents/builder.md')
    const claudeAgent = path.join(tempDir, '.claude/agents/builder.md')
    fs.rmSync(opencodeAgent)
    fs.rmSync(claudeAgent)

    const storePath = path.join(tempDir, '.ai-setup.json')
    const beforeStore = fs.readFileSync(storePath, 'utf-8')

    await runCompile(['--tools', 'opencode'])

    const afterStore = fs.readFileSync(storePath, 'utf-8')
    expect(afterStore).toBe(beforeStore)
    expect(fs.existsSync(opencodeAgent)).toBe(true)
    expect(fs.existsSync(claudeAgent)).toBe(false)
  })

  it('compile global uses native global target paths and logs unsupported tools', async () => {
    const tempDir = makeTempRepo()
    const tempHome = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-home-'))
    const originalHome = process.env.HOME
    process.env.HOME = tempHome
    process.chdir(tempDir)

    try {
      await runInit([
        '--scope',
        'global',
        '--tools',
        'opencode,claude-code,copilot',
        '--no-interactive',
      ])

      const opencodeAgent = path.join(tempHome, '.config/opencode/agents/builder.md')
      const claudeAgent = path.join(tempHome, '.claude/builder.md')
      fs.rmSync(opencodeAgent)
      fs.rmSync(claudeAgent, { recursive: true, force: true })

      const infoSpy = vi.spyOn(console, 'info').mockImplementation(() => {})
      await runCompile(['--scope', 'global', '--tools', 'opencode,copilot'])

      expect(fs.existsSync(opencodeAgent)).toBe(true)
      expect(fs.existsSync(claudeAgent)).toBe(false)
      expect(infoSpy).toHaveBeenCalledWith("Copilot doesn't support file-based global config. Use project scope instead.")
      infoSpy.mockRestore()
    } finally {
      if (originalHome === undefined) {
        delete process.env.HOME
      } else {
        process.env.HOME = originalHome
      }
      fs.rmSync(tempHome, { recursive: true, force: true })
    }
  })

  it('supports workspace init + workspace compile from planning repo', async () => {
    const workspaceRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-workspace-cli-'))
    const planningRepoDir = path.join(workspaceRoot, 'planning-repo')
    const fedoraRepoDir = path.join(workspaceRoot, 'fedora')
    const checkoutRepoDir = path.join(workspaceRoot, 'creator-checkout')

    fs.mkdirSync(planningRepoDir, { recursive: true })
    fs.mkdirSync(fedoraRepoDir, { recursive: true })
    fs.mkdirSync(checkoutRepoDir, { recursive: true })

    process.chdir(workspaceRoot)

    await runInit([
      '--scope',
      'workspace',
      '--planning-repo',
      planningRepoDir,
      '--repos',
      '../fedora,../creator-checkout',
      '--tools',
      'opencode,claude-code',
      '--name',
      'teachable-workspace',
      '--no-interactive',
    ])

    expect(fs.existsSync(path.join(planningRepoDir, '.ai-setup.json'))).toBe(true)
    expect(fs.existsSync(path.join(fedoraRepoDir, '.ai-setup.json'))).toBe(false)
    expect(fs.existsSync(path.join(checkoutRepoDir, '.ai-setup.json'))).toBe(false)

    const opencodeAgent = path.join(planningRepoDir, '.opencode/agents/builder.md')
    fs.rmSync(opencodeAgent)

    process.chdir(planningRepoDir)
    await runCompile(['--scope', 'workspace', '--tools', 'opencode'])

    expect(fs.existsSync(opencodeAgent)).toBe(true)

    const config = JSON.parse(fs.readFileSync(path.join(planningRepoDir, '.ai-setup.json'), 'utf-8')) as any
    expect(config.config.setupScope).toBe('workspace')
    expect(config.config.workspaceName).toBe('teachable-workspace')
    expect(config.config.repos).toEqual([
      { name: 'fedora', path: '../fedora' },
      { name: 'creator-checkout', path: '../creator-checkout' },
    ])
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
