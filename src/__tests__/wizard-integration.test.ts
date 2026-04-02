import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { existsSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'

vi.mock('@clack/prompts', () => ({
  select: vi.fn(),
  multiselect: vi.fn(),
  text: vi.fn(),
  confirm: vi.fn(),
  note: vi.fn(),
  cancel: vi.fn(),
  intro: vi.fn(),
  outro: vi.fn(),
  spinner: vi.fn(() => ({ start: vi.fn(), stop: vi.fn() })),
  isCancel: vi.fn(() => false),
  log: {
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
    success: vi.fn(),
    step: vi.fn(),
    message: vi.fn(),
  },
}))

import { runWizard } from '../wizard/index.js'

describe('wizard integration (non-interactive)', () => {
  let tempDir: string

  beforeEach(() => {
    tempDir = mkdtempSync(path.join(tmpdir(), 'ai-setup-wizard-integration-'))
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
    vi.clearAllMocks()
  })

  it('creates complete file tree in non-interactive mode', async () => {
    await runWizard({
      interactive: false,
      cliOverrides: {
        scope: 'project',
        tools: ['pi', 'opencode', 'claude-code', 'gemini', 'copilot'],
        name: 'test-project',
      },
      targetDir: tempDir,
    })

    expect(existsSync(path.join(tempDir, 'docs', 'features'))).toBe(true)
    expect(existsSync(path.join(tempDir, 'docs', 'bugfixes'))).toBe(true)

    expect(existsSync(path.join(tempDir, 'AGENTS.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, 'CLAUDE.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, 'GEMINI.md'))).toBe(true)

    expect(existsSync(path.join(tempDir, '.ai-setup.json'))).toBe(true)

    const manifest = JSON.parse(readFileSync(path.join(tempDir, '.ai-setup.json'), 'utf-8'))
    expect(manifest.selections).toBeDefined()
    expect(manifest.config.setupScope).toBe('project')
  })

  it('creates only opencode files when only opencode selected', async () => {
    await runWizard({
      interactive: false,
      cliOverrides: {
        scope: 'project',
        tools: ['opencode'],
        name: 'test-project',
      },
      targetDir: tempDir,
    })

    expect(existsSync(path.join(tempDir, 'AGENTS.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, 'GEMINI.md'))).toBe(false)
    expect(existsSync(path.join(tempDir, '.github', 'copilot-instructions.md'))).toBe(false)
  })

  it('re-run preserves manifest selections field', async () => {
    await runWizard({
      interactive: false,
      cliOverrides: {
        scope: 'project',
        tools: ['opencode'],
        name: 'test-project',
      },
      targetDir: tempDir,
    })

    await runWizard({
      interactive: false,
      cliOverrides: {
        scope: 'project',
        tools: ['opencode'],
        name: 'test-project',
      },
      targetDir: tempDir,
    })

    const manifest = JSON.parse(readFileSync(path.join(tempDir, '.ai-setup.json'), 'utf-8'))
    expect(manifest.selections).toBeDefined()
    expect(manifest.config.tools).toEqual(['opencode'])
  })

  it('global scope scaffolds into ~/.ai and logs unsupported tools', async () => {
    const homeDir = mkdtempSync(path.join(tmpdir(), 'ai-setup-home-'))
    const infoSpy = vi.spyOn(console, 'info').mockImplementation(() => {})

    await runWizard({
      interactive: false,
      homeDir,
      cliOverrides: {
        scope: 'global',
        tools: ['opencode', 'claude-code', 'copilot', 'gemini', 'pi'],
      },
      targetDir: tempDir,
    })

    const canonicalDir = path.join(homeDir, '.ai')
    expect(existsSync(path.join(canonicalDir, '.ai-setup.json'))).toBe(true)
    expect(existsSync(path.join(canonicalDir, 'docs', 'templates', 'task.md'))).toBe(true)

    expect(existsSync(path.join(homeDir, '.config', 'opencode', 'agents', 'builder.md'))).toBe(true)
    expect(existsSync(path.join(homeDir, '.config', 'opencode', 'command', 'implement.md'))).toBe(true)
    expect(existsSync(path.join(homeDir, '.config', 'opencode', 'commands'))).toBe(false)
    expect(existsSync(path.join(homeDir, '.claude', 'builder.md'))).toBe(true)

    const manifest = JSON.parse(readFileSync(path.join(canonicalDir, '.ai-setup.json'), 'utf-8'))
    expect(manifest.config.setupScope).toBe('global')
    expect(manifest.config.targetDir).toBe(canonicalDir)
    expect(manifest.config.tools).toEqual(['opencode', 'claude-code'])
    expect(manifest.config.projectName).toBe('global')

    expect(infoSpy).toHaveBeenCalledWith("Copilot doesn't support file-based global config. Use project scope instead.")
    expect(infoSpy).toHaveBeenCalledWith("Gemini doesn't support file-based global config. Use project scope instead.")
    expect(infoSpy).toHaveBeenCalledWith("Pi doesn't support file-based global config. Use project scope instead.")

    rmSync(homeDir, { recursive: true, force: true })
    infoSpy.mockRestore()
  })

  it('workspace scope scaffolds only planning repo and stores repo references', async () => {
    const workspaceRoot = mkdtempSync(path.join(tmpdir(), 'ai-setup-workspace-'))
    const planningRepoDir = path.join(workspaceRoot, 'planning-repo')
    const fedoraRepoDir = path.join(workspaceRoot, 'fedora')
    const checkoutRepoDir = path.join(workspaceRoot, 'creator-checkout')

    mkdirSync(planningRepoDir, { recursive: true })
    mkdirSync(fedoraRepoDir, { recursive: true })
    mkdirSync(checkoutRepoDir, { recursive: true })

    await runWizard({
      interactive: false,
      cliOverrides: {
        scope: 'workspace',
        tools: ['opencode', 'claude-code'],
        name: 'teachable-workspace',
        planningRepo: planningRepoDir,
        repos: ['../fedora', '../creator-checkout'],
      },
      targetDir: workspaceRoot,
    })

    expect(existsSync(path.join(planningRepoDir, '.ai-setup.json'))).toBe(true)
    expect(existsSync(path.join(planningRepoDir, '.opencode', 'agents', 'builder.md'))).toBe(true)
    expect(existsSync(path.join(planningRepoDir, '.claude', 'agents', 'builder.md'))).toBe(true)
    expect(existsSync(path.join(planningRepoDir, 'docs', 'memory', 'decisions'))).toBe(true)
    expect(existsSync(path.join(planningRepoDir, 'docs', 'memory', 'handoffs'))).toBe(true)
    expect(existsSync(path.join(planningRepoDir, 'docs', 'memory', 'patterns'))).toBe(true)
    expect(existsSync(path.join(planningRepoDir, 'docs', 'memory', 'projects'))).toBe(true)

    expect(existsSync(path.join(fedoraRepoDir, '.ai-setup.json'))).toBe(false)
    expect(existsSync(path.join(checkoutRepoDir, '.ai-setup.json'))).toBe(false)
    expect(existsSync(path.join(fedoraRepoDir, '.opencode'))).toBe(false)
    expect(existsSync(path.join(checkoutRepoDir, '.claude'))).toBe(false)

    const manifest = JSON.parse(readFileSync(path.join(planningRepoDir, '.ai-setup.json'), 'utf-8'))
    expect(manifest.config.setupScope).toBe('workspace')
    expect(manifest.config.targetDir).toBe(path.resolve(planningRepoDir))
    expect(manifest.config.planningRepoPath).toBe(path.resolve(planningRepoDir))
    expect(manifest.config.workspaceName).toBe('teachable-workspace')
    expect(manifest.config.projectName).toBe('planning-repo')
    expect(manifest.config.repos).toEqual([
      { name: 'fedora', path: '../fedora' },
      { name: 'creator-checkout', path: '../creator-checkout' },
    ])

    const fedoraEntries = readdirSync(fedoraRepoDir)
    const checkoutEntries = readdirSync(checkoutRepoDir)
    expect(fedoraEntries).toEqual([])
    expect(checkoutEntries).toEqual([])

    rmSync(workspaceRoot, { recursive: true, force: true })
  })
})
