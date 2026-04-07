import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createProgram } from '../cli.js'
import { ejectCommand } from '../commands/eject.js'
import { createImportCommand } from '../commands/import.js'
import { createMigrateCommand } from '../commands/migrate.js'
import type { StoreData } from '../store/schema.js'

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

  const runStatus = async (): Promise<void> => {
    const program = createProgram()
    await program.parseAsync(['node', 'ai-setup', 'status'])
  }

  const _runEject = async (): Promise<void> => {
    const program = createProgram()
    await program.parseAsync(['node', 'ai-setup', 'eject'])
  }

  const runAdd = async (tool: string): Promise<void> => {
    const program = createProgram()
    await program.parseAsync(['node', 'ai-setup', 'add', tool])
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
      'specs/features',
      'specs/bugfixes',
      'specs/refactors',
      'specs/tech-debt',
      'specs/adrs',
      'specs/memory',
      'specs/standards',
      'specs/templates',
      'specs/rules',
      'specs/features/AGENTS.md',
      'specs/bugfixes/AGENTS.md',
      'specs/refactors/AGENTS.md',
      'specs/tech-debt/AGENTS.md',
      'specs/adrs/AGENTS.md',
      'specs/memory/AGENTS.md',
      'specs/standards/AGENTS.md',
      'specs/templates/AGENTS.md',
      'specs/rules/AGENTS.md',
      'AGENTS.md',
      '.git/hooks/pre-commit',
      '.pi/prompts',
      '.pi/skills',
      '.pi/settings.json',
      '.opencode/agents',
      '.opencode/skills',
      '.ai-setup.json',
    ]

    for (const rel of expectedPaths) {
      expect(fs.existsSync(path.join(tempDir, rel)), `${rel} should exist`).toBe(true)
    }

    const config = JSON.parse(fs.readFileSync(path.join(tempDir, '.ai-setup.json'), 'utf-8')) as StoreData
    expect(config.config.projectName).toBe('integration-test')
    expect(config.config.setupScope).toBe('project')
    expect(config.config.tools).toEqual(['pi', 'opencode'])
    expect(config.files.length).toBeGreaterThan(20)
    expect(config.files.some((f: { path: string }) => f.path === '.pi/settings.json')).toBe(true)
    expect(config.files.some((f: { path: string }) => f.path === '.opencode/agents/builder.md')).toBe(true)
    expect(config.files.some((f: { path: string }) => f.path === '.pi/skills/research/SKILL.md')).toBe(true)
    expect(config.files.some((f: { path: string }) => f.path === '.pi/prompts/plan.md')).toBe(true)
    expect(config.files.some((f: { path: string }) => f.path === '.opencode/skills/research/SKILL.md')).toBe(true)
    expect(config.files.some((f: { path: string }) => f.path === 'specs/templates/task.md')).toBe(true)
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

    const config = JSON.parse(fs.readFileSync(path.join(tempDir, '.ai-setup.json'), 'utf-8')) as StoreData
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
    expect(fs.existsSync(path.join(tempDir, 'opencode.json'))).toBe(true)
    expect(fs.existsSync(path.join(tempDir, '.opencode/agents'))).toBe(true)
    expect(fs.existsSync(path.join(tempDir, '.opencode/skills'))).toBe(true)
    expect(fs.existsSync(path.join(tempDir, '.opencode/commands'))).toBe(true)
    expect(fs.existsSync(path.join(tempDir, '.opencode/templates'))).toBe(false)

    expect(fs.existsSync(path.join(tempDir, 'CLAUDE.md'))).toBe(false)
    expect(fs.existsSync(path.join(tempDir, '.pi/agents'))).toBe(false)
    expect(fs.existsSync(path.join(tempDir, '.pi/skills'))).toBe(false)

    const config = JSON.parse(fs.readFileSync(path.join(tempDir, '.ai-setup.json'), 'utf-8')) as StoreData
    expect(config.config.tools).toEqual(['opencode'])
    expect(config.files.some((f: { path: string }) => f.path.startsWith('.opencode/'))).toBe(true)
    expect(config.files.some((f: { path: string }) => f.path.startsWith('.pi/'))).toBe(false)
  })

  it('init --dry-run shows plan output and writes no files', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})

    await runInit([
      '--scope',
      'project',
      '--tools',
      'opencode',
      '--name',
      'dry-run-test',
      '--no-interactive',
      '--dry-run',
    ])

    const logOutput = logSpy.mock.calls.map((call) => call.map((value) => String(value)).join(' ')).join('\n')
    expect(logOutput).toContain('[dry-run] Would create:')
    expect(logOutput).toContain('Dry run complete. Would create')

    expect(fs.existsSync(path.join(tempDir, '.ai-setup.json'))).toBe(false)
    expect(fs.existsSync(path.join(tempDir, 'AGENTS.md'))).toBe(false)
    expect(fs.existsSync(path.join(tempDir, '.opencode'))).toBe(false)

    logSpy.mockRestore()
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

  it('compile uses stored phase-2 settings for compiled root regeneration', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    await runInit([
      '--scope',
      'project',
      '--tools',
      'opencode',
      '--name',
      'compile-phase2-test',
      '--planning-dir',
      '.specs',
      '--disable-features',
      'treeOfThoughts',
      '--no-interactive',
    ])

    const rootFile = path.join(tempDir, 'AGENTS.md')
    fs.rmSync(rootFile)
    expect(fs.existsSync(rootFile)).toBe(false)

    await runCompile(['--tools', 'opencode'])

    expect(fs.existsSync(rootFile)).toBe(true)
    const content = fs.readFileSync(rootFile, 'utf-8')
    expect(content).toContain('<planning-dir>.specs</planning-dir>')
    expect(content).not.toContain('<decision-protocol>')
    expect(content).toContain('<git-conventions>')
  })

  it('compile with --scope global reads manifest from ~/.ai and restores global tool paths', async () => {
    const tempDir = makeTempRepo()
    const tempHome = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-home-global-compile-'))
    const originalHome = process.env.HOME
    process.env.HOME = tempHome
    process.chdir(tempDir)

    try {
      await runInit([
        '--scope',
        'global',
        '--tools',
        'opencode,claude-code',
        '--no-interactive',
      ])

      const opencodeAgent = path.join(tempHome, '.config/opencode/agents/builder.md')
      const claudeAgent = path.join(tempHome, '.claude/builder.md')

      fs.rmSync(opencodeAgent)
      fs.rmSync(claudeAgent, { recursive: true, force: true })

      await runCompile(['--scope', 'global'])

      expect(fs.existsSync(opencodeAgent)).toBe(true)
      expect(fs.existsSync(claudeAgent)).toBe(true)
      expect(fs.existsSync(path.join(tempDir, '.ai-setup.json'))).toBe(false)
      expect(fs.existsSync(path.join(tempHome, '.ai/.ai-setup.json'))).toBe(true)
    } finally {
      if (originalHome === undefined) {
        delete process.env.HOME
      } else {
        process.env.HOME = originalHome
      }
      fs.rmSync(tempHome, { recursive: true, force: true })
    }
  })

  it('compile fails when setup manifest does not exist', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    await expect(runCompile()).rejects.toThrow(/Setup manifest not found/)
  })

  it('compile --dry-run shows preview and performs no writes', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    await runInit([
      '--scope',
      'project',
      '--tools',
      'opencode',
      '--name',
      'compile-dry-run-test',
      '--no-interactive',
    ])

    const opencodeAgent = path.join(tempDir, '.opencode/agents/builder.md')
    fs.rmSync(opencodeAgent)
    expect(fs.existsSync(opencodeAgent)).toBe(false)

    const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
    await runCompile(['--tools', 'opencode', '--dry-run'])

    const logOutput = logSpy.mock.calls.map((call) => call.map((value) => String(value)).join(' ')).join('\n')
    expect(logOutput).toContain('[dry-run] Compile preview:')
    expect(logOutput).toContain('[dry-run] Would compile tool: opencode')
    expect(logOutput).toContain('Dry run complete. Would compile 1 tool(s): opencode')

    expect(fs.existsSync(opencodeAgent)).toBe(false)
    logSpy.mockRestore()
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

  it('status shows installed tools and setup scope after init', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    await runInit([
      '--scope',
      'project',
      '--tools',
      'opencode,claude-code',
      '--name',
      'status-test',
      '--no-interactive',
    ])

    const writeSpy = vi.spyOn(process.stdout, 'write').mockImplementation(() => true)
    await runStatus()

    const output = writeSpy.mock.calls.map((call) => String(call[0])).join('')
    // New summary box format
    expect(output).toContain('Scope')
    expect(output).toContain('project')
    expect(output).toContain('opencode, claude-code')
    expect(output).toContain('Planning dir')
    expect(output).toContain('.planning')
    expect(output).toContain('Features')
    expect(output).toContain('Git Conventions')
    expect(output).toContain('File Health')
    expect(output).not.toContain('coming soon')

    writeSpy.mockRestore()
  })

  it('eject removes .ai-setup.json after confirmation', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    await runInit([
      '--scope',
      'project',
      '--tools',
      'opencode',
      '--name',
      'eject-test',
      '--no-interactive',
    ])

    expect(fs.existsSync(path.join(tempDir, '.ai-setup.json'))).toBe(true)

    await ejectCommand(tempDir, {
      confirmEject: async () => true,
    })

    expect(fs.existsSync(path.join(tempDir, '.ai-setup.json'))).toBe(false)
  })

  it('add installs claude-code files and updates manifest tools list', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    await runInit([
      '--scope',
      'project',
      '--tools',
      'opencode',
      '--name',
      'add-test',
      '--no-interactive',
    ])

    expect(fs.existsSync(path.join(tempDir, '.claude'))).toBe(false)

    await runAdd('claude-code')

    expect(fs.existsSync(path.join(tempDir, '.claude', 'skills'))).toBe(true)
    expect(fs.existsSync(path.join(tempDir, '.claude', 'rules'))).toBe(true)
    expect(fs.existsSync(path.join(tempDir, 'CLAUDE.md'))).toBe(true)
    const compiledRootContent = fs.readFileSync(path.join(tempDir, 'CLAUDE.md'), 'utf-8')
    expect(compiledRootContent).toContain('<system-context>')

    const config = JSON.parse(fs.readFileSync(path.join(tempDir, '.ai-setup.json'), 'utf-8')) as StoreData
    expect(config.config.tools).toContain('claude-code')
    expect(config.files.some((f: { path: string }) => f.path.startsWith('.claude/'))).toBe(true)
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

    const config = JSON.parse(fs.readFileSync(path.join(planningRepoDir, '.ai-setup.json'), 'utf-8')) as StoreData
    expect(config.config.setupScope).toBe('workspace')
    expect(config.config.workspaceName).toBe('teachable-workspace')
    expect(config.config.repos).toEqual([
      { name: 'fedora', path: '../fedora', type: 'unknown' },
      { name: 'creator-checkout', path: '../creator-checkout', type: 'unknown' },
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

describe('basic command smoke tests', () => {
  let originalCwd: string

  const makeTempRepo = (): string => {
    const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-cli-smoke-'))
    fs.mkdirSync(path.join(tempDir, '.git'), { recursive: true })
    return tempDir
  }

  const runCreate = async (args: string[]): Promise<void> => {
    const program = createProgram()
    await program.parseAsync(['node', 'ai-setup', 'create', ...args])
  }

  beforeEach(() => {
    originalCwd = process.cwd()
  })

  afterEach(() => {
    process.chdir(originalCwd)
  })

  it('add without prior init throws manifest-not-found error', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    const program = createProgram()
    await expect(program.parseAsync(['node', 'ai-setup', 'add', 'opencode'])).rejects.toThrow(/Setup manifest not found/)
  })

  it('create --help lists available subcommands', () => {
    const program = createProgram()
    const createCommand = program.commands.find((command) => command.name() === 'create')

    expect(createCommand?.helpInformation()).toContain('agent [options] [name]')
    expect(createCommand?.helpInformation()).toContain('skill [options] [name]')
    expect(createCommand?.helpInformation()).toContain('command [options] [name]')
    expect(createCommand?.helpInformation()).toContain('prompt [options] [name]')
    expect(createCommand?.helpInformation()).toContain('template [options] [name]')
    expect(createCommand?.helpInformation()).toContain('workflow [options] [name]')
  })

  it('create agent writes agent file to library/agents', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    await runCreate(['agent', 'my-agent', '--model', 'gpt-4o', '--mode', 'interactive', '--tools', 'fs', '--no-interactive'])

    expect(fs.existsSync(path.join(tempDir, 'library/agents/my-agent.md'))).toBe(true)
  })

  it('create skill writes skill file to library/skills', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    await runCreate(['skill', 'my-skill', '--command', 'my-skill', '--steps', 'Clarify scope', '--no-interactive'])

    expect(fs.existsSync(path.join(tempDir, 'library/skills/my-skill.md'))).toBe(true)
  })

  it('create prompt writes prompt file to library/prompts', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    await runCreate(['prompt', 'my-prompt', '--task-context', 'test task', '--output-format', 'markdown', '--no-interactive'])

    expect(fs.existsSync(path.join(tempDir, 'library/prompts/my-prompt.md'))).toBe(true)
  })

  it('eject without prior init throws manifest-not-found error', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    const program = createProgram()
    await expect(program.parseAsync(['node', 'ai-setup', 'eject'])).rejects.toThrow(/Setup manifest not found/)
  })
})
