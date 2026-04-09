import { existsSync, mkdirSync, mkdtempSync, readdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

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
        tools: ['opencode', 'claude-code', 'gemini', 'copilot', 'codex'],
        name: 'test-project',
      },
      targetDir: tempDir,
    })

    expect(existsSync(path.join(tempDir, 'specs', 'features'))).toBe(true)
    expect(existsSync(path.join(tempDir, 'specs', 'bugfixes'))).toBe(true)

    expect(existsSync(path.join(tempDir, 'AGENTS.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, 'CLAUDE.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, 'GEMINI.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.ai', 'constitution'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.ai', 'constitution', 'constitution.md'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.ai', 'mcp.json'))).toBe(true)
    expect(existsSync(path.join(tempDir, 'opencode.jsonc'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.vscode', 'mcp.json'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.gemini', 'settings.json'))).toBe(true)

    expect(existsSync(path.join(tempDir, '.ai-setup.json'))).toBe(true)

    const manifest = JSON.parse(readFileSync(path.join(tempDir, '.ai-setup.json'), 'utf-8'))
    expect(manifest.selections).toBeDefined()
    expect(manifest.config.setupScope).toBe('project')
  })

  it('creates .ai/orchestration when orchestrator is enabled', async () => {
    await runWizard({
      interactive: false,
      cliOverrides: {
        scope: 'project',
        tools: ['opencode', 'claude-code'],
        name: 'test-project',
        enableServers: ['orchestrator'],
      },
      targetDir: tempDir,
    })

    expect(existsSync(path.join(tempDir, '.ai', 'orchestration'))).toBe(true)
    expect(existsSync(path.join(tempDir, '.ai', 'orchestration', 'chains', 'feature.json'))).toBe(true)

    const manifest = JSON.parse(readFileSync(path.join(tempDir, '.ai-setup.json'), 'utf-8'))
    expect(manifest.config.enableServers).toEqual(['orchestrator'])
  })

  it('does not create .ai/orchestration when orchestrator is not enabled', async () => {
    await runWizard({
      interactive: false,
      cliOverrides: {
        scope: 'project',
        tools: ['opencode', 'claude-code'],
        name: 'test-project',
      },
      targetDir: tempDir,
    })

    expect(existsSync(path.join(tempDir, '.ai', 'orchestration'))).toBe(false)
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

  it('absorbs existing tool files into .ai when absorb is enabled', async () => {
    mkdirSync(path.join(tempDir, '.opencode', 'agents'), { recursive: true })
    const customAgentPath = path.join(tempDir, '.opencode', 'agents', 'custom-agent.md')
    const customAgentContent = '# Custom Agent\n\nKeep my custom instructions.'
    writeFileSync(customAgentPath, customAgentContent, 'utf-8')

    await runWizard({
      interactive: false,
      absorb: true,
      cliOverrides: {
        scope: 'project',
        tools: ['opencode'],
        name: 'test-project',
      },
      targetDir: tempDir,
    })

    const canonicalAgentPath = path.join(tempDir, '.ai', 'agents', 'custom-agent.md')
    expect(existsSync(canonicalAgentPath)).toBe(true)
    expect(readFileSync(canonicalAgentPath, 'utf-8')).toContain('Keep my custom instructions.')
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
        tools: ['opencode', 'claude-code', 'copilot', 'gemini', 'codex'],
      },
      targetDir: tempDir,
    })

    const canonicalDir = path.join(homeDir, '.ai')
    expect(existsSync(path.join(canonicalDir, '.ai-setup.json'))).toBe(true)
    expect(existsSync(path.join(canonicalDir, 'specs', 'templates', 'task.md'))).toBe(true)

    expect(existsSync(path.join(homeDir, '.config', 'opencode', 'agents', 'builder.md'))).toBe(true)
    expect(existsSync(path.join(homeDir, '.config', 'opencode', 'skills', 'implement', 'SKILL.md'))).toBe(true)
    expect(existsSync(path.join(homeDir, '.config', 'opencode', 'commands'))).toBe(true)
    expect(existsSync(path.join(homeDir, '.config', 'opencode', 'templates'))).toBe(false)
    expect(existsSync(path.join(homeDir, '.claude', 'builder.md'))).toBe(true)

    const manifest = JSON.parse(readFileSync(path.join(canonicalDir, '.ai-setup.json'), 'utf-8'))
    expect(manifest.config.setupScope).toBe('global')
    expect(manifest.config.targetDir).toBe(canonicalDir)
    expect(manifest.config.tools).toEqual(['opencode', 'claude-code'])
    expect(manifest.config.projectName).toBe('global')

    expect(infoSpy).toHaveBeenCalledWith("Copilot doesn't support file-based global config. Use project scope instead.")
    expect(infoSpy).toHaveBeenCalledWith("Gemini doesn't support file-based global config. Use project scope instead.")

    rmSync(homeDir, { recursive: true, force: true })
    infoSpy.mockRestore()
  })

  it('workspace scope scaffolds planning repo and ledgers without writing into referenced repos', async () => {
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
    expect(existsSync(path.join(planningRepoDir, 'specs', 'memory', 'decisions'))).toBe(true)
    expect(existsSync(path.join(planningRepoDir, 'specs', 'memory', 'handoffs'))).toBe(true)
    expect(existsSync(path.join(planningRepoDir, 'specs', 'memory', 'patterns'))).toBe(true)
    expect(existsSync(path.join(planningRepoDir, 'specs', 'memory', 'projects'))).toBe(true)
    expect(existsSync(path.join(planningRepoDir, 'specs', 'memory', 'repos', 'fedora', 'ledger.md'))).toBe(true)
    expect(existsSync(path.join(planningRepoDir, 'specs', 'memory', 'repos', 'fedora', 'last-known-state.md'))).toBe(true)
    expect(existsSync(path.join(planningRepoDir, 'specs', 'memory', 'repos', 'creator-checkout', 'ledger.md'))).toBe(true)
    expect(readFileSync(path.join(planningRepoDir, 'AGENTS.md'), 'utf-8')).toContain('## Workspace Repos')
    expect(readFileSync(path.join(planningRepoDir, 'AGENTS.md'), 'utf-8')).toContain('### fedora')
    expect(readFileSync(path.join(planningRepoDir, 'AGENTS.md'), 'utf-8')).toContain('### creator-checkout')

    expect(existsSync(path.join(fedoraRepoDir, '.ai-setup.json'))).toBe(false)
    expect(existsSync(path.join(checkoutRepoDir, '.ai-setup.json'))).toBe(false)
    expect(existsSync(path.join(fedoraRepoDir, 'AGENTS.md'))).toBe(false)
    expect(existsSync(path.join(fedoraRepoDir, 'CLAUDE.md'))).toBe(false)
    expect(existsSync(path.join(fedoraRepoDir, '.claude', 'settings.json'))).toBe(false)
    expect(existsSync(path.join(checkoutRepoDir, 'AGENTS.md'))).toBe(false)
    expect(existsSync(path.join(checkoutRepoDir, 'CLAUDE.md'))).toBe(false)
    expect(existsSync(path.join(checkoutRepoDir, '.claude', 'settings.json'))).toBe(false)

    const manifest = JSON.parse(readFileSync(path.join(planningRepoDir, '.ai-setup.json'), 'utf-8'))
    expect(manifest.config.setupScope).toBe('workspace')
    expect(manifest.config.targetDir).toBe(path.resolve(planningRepoDir))
    expect(manifest.config.planningRepoPath).toBe(path.resolve(planningRepoDir))
    expect(manifest.config.workspaceName).toBe('teachable-workspace')
    expect(manifest.config.projectName).toBe('planning-repo')
    expect(manifest.config.repos).toEqual([
      { name: 'fedora', path: '../fedora', type: 'unknown' },
      { name: 'creator-checkout', path: '../creator-checkout', type: 'unknown' },
    ])

    const fedoraEntries = readdirSync(fedoraRepoDir).sort()
    const checkoutEntries = readdirSync(checkoutRepoDir).sort()
    expect(fedoraEntries).toEqual([])
    expect(checkoutEntries).toEqual([])

    rmSync(workspaceRoot, { recursive: true, force: true })
  })
})
