import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('node:child_process', () => ({
  execSync: vi.fn(),
}))

// Mock @clack/prompts before importing phases
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
}))

import { execSync } from 'node:child_process'
import * as p from '@clack/prompts'
import { ALL_AGENTS, ALL_SKILLS } from '../types.js'
import { GO_BACK } from '../utils/ui.js'
import {
  defaultMcpServersForPreset,
  detectInstalledCliToolsFromCatalog,
  filterToolsByScope,
  runPhase1,
  toolOptionsForScope,
} from '../wizard/phase1-context.js'
import { runPhase3 } from '../wizard/phase3-conflicts.js'

/**
 * Type-narrowing helper for Phase 1 results.
 * Tests never trigger GO_BACK since they mock @clack/prompts.
 */
function unwrapPhase1(result: Awaited<ReturnType<typeof runPhase1>>) {
  if (result === GO_BACK) throw new Error('Unexpected GO_BACK')
  return result
}

function makeTempDir(prefix: string): string {
  return mkdtempSync(path.join(tmpdir(), prefix))
}

describe('wizard phases 1 and 3', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(p.isCancel).mockReturnValue(false)
    vi.mocked(execSync).mockImplementation(() => {
      throw new Error('not found')
    })
  })

  it('Phase 1: non-interactive with all CLI overrides returns correct values', async () => {
    const result = await runPhase1({
      interactive: false,
      prior: {},
      cliOverrides: { scope: 'project', tools: ['opencode', 'claude-code'], name: 'my-project' },
      targetDir: '/tmp',
    })

    expect(result).toEqual({
      setupScope: 'project',
      tools: ['opencode', 'claude-code'],
      skills: ALL_SKILLS,
      agents: ALL_AGENTS,
      mcpPreset: 'recommended',
      enableServers: ['codegraph', 'filesystem', 'graphify', 'memoria', 'memory', 'obsidian', 'qmd', 'ripgrep'],
      projectName: 'my-project',
    })
  })

  it('Phase 1: interactive follows parity step ordering and defaults', async () => {
    vi.mocked(p.select).mockResolvedValueOnce('project')
      .mockResolvedValueOnce('recommended')
    vi.mocked(p.multiselect)
      .mockResolvedValueOnce(['opencode'])
      .mockResolvedValueOnce(ALL_SKILLS)
      .mockResolvedValueOnce(ALL_AGENTS)
      .mockResolvedValueOnce(['filesystem', 'memoria', 'memory', 'ripgrep'])
      .mockResolvedValueOnce(['gh'])
    vi.mocked(p.text)
      .mockResolvedValueOnce('my-project')
      .mockResolvedValueOnce('Acme')
      .mockResolvedValueOnce('Platform')

    const result = unwrapPhase1(await runPhase1({
      interactive: true,
      prior: {},
      cliOverrides: {},
      targetDir: process.cwd(),
    }))

    expect(result).toMatchObject({
      setupScope: 'project',
      tools: ['opencode'],
      skills: ALL_SKILLS,
      agents: ALL_AGENTS,
      mcpPreset: 'recommended',
      enableServers: ['filesystem', 'memoria', 'memory', 'ripgrep'],
      projectName: 'my-project',
      cliTools: ['gh'],
      organization: 'Acme',
      team: 'Platform',
    })

    const selectMessages = vi.mocked(p.select).mock.calls.map(([arg]) => arg.message)
    expect(selectMessages).toEqual([
      'Setup scope: (previous: Project)',
      'Which MCP preset should be enabled? (previous: recommended)',
    ])

    const multiselectMessages = vi.mocked(p.multiselect).mock.calls.map(([arg]) => arg.message)
    expect(multiselectMessages).toEqual([
      'Which AI tools are you using? (previous: opencode, claude-code, gemini, copilot, codex, pi)',
      `Which skills should be installed? (previous: ${ALL_SKILLS.join(', ')})`,
      `Which agents should be installed? (previous: ${ALL_AGENTS.join(', ')})`,
      'Which MCP servers would you like to enable?',
      'Which CLI tools do you have installed? (press space to select, enter to confirm or skip)',
    ])

    const textMessages = vi.mocked(p.text).mock.calls.map(([arg]) => arg.message)
    expect(textMessages).toEqual([
      'Project name?',
      'Organization? (optional)',
      'Team? (optional)',
    ])
  })

  it('Phase 1: non-interactive with scope=global returns setupScope=global', async () => {
    const result = await runPhase1({
      interactive: false,
      prior: {},
      cliOverrides: { scope: 'global', tools: ['opencode'], name: 'my-project' },
      targetDir: '/tmp',
    })

    expect(unwrapPhase1(result).setupScope).toBe('global')
  })

  it('Phase 1: non-interactive with scope=global defaults projectName=global', async () => {
    const result = await runPhase1({
      interactive: false,
      prior: {},
      cliOverrides: { scope: 'global', tools: ['opencode'], name: 'ignored-name' },
      targetDir: '/tmp',
    })

    expect(unwrapPhase1(result).projectName).toBe('global')
  })

  it('Phase 1: mcp preset expansion matches parity contract', () => {
    expect(defaultMcpServersForPreset('minimal', process.cwd())).toEqual(['filesystem', 'ripgrep'])
    expect(defaultMcpServersForPreset('recommended', process.cwd())).toEqual([
      'codegraph',
      'filesystem',
      'graphify',
      'memoria',
      'memory',
      'obsidian',
      'qmd',
      'ripgrep',
    ])
    expect(defaultMcpServersForPreset('full', process.cwd())).toEqual([
      'atlassian',
      'brave-search',
      'codegraph',
      'context7',
      'fetch',
      'filesystem',
      'graphify',
      'memoria',
      'memory',
      'obsidian',
      'orchestrator',
      'playwright',
      'qmd',
      'ripgrep',
    ])
  })

  it('Phase 1: scope tool filtering excludes global pi and keeps global copilot', () => {
    expect(toolOptionsForScope('global').map(({ value }) => value)).toEqual([
      'opencode',
      'claude-code',
      'gemini',
      'copilot',
      'codex',
    ])
    expect(toolOptionsForScope('project').map(({ value }) => value)).toContain('pi')
    expect(toolOptionsForScope('workspace').map(({ value }) => value)).toContain('pi')
    expect(filterToolsByScope(['opencode', 'copilot', 'pi'], 'global')).toEqual(['opencode', 'copilot'])
  })

  it('Phase 1: installed CLI defaults come from catalog tools', () => {
    vi.mocked(execSync).mockImplementation((command) => {
      if (command === 'which gh') return Buffer.from('/usr/bin/gh') as never
      throw new Error('not found')
    })

    expect(detectInstalledCliToolsFromCatalog(process.cwd())).toEqual(['gh'])
  })

  it('Phase 1: non-interactive with scope=workspace returns setupScope=workspace', async () => {
    const result = await runPhase1({
      interactive: false,
      prior: {},
      cliOverrides: {
        scope: 'workspace',
        tools: ['opencode'],
        name: 'my-project',
        planningRepo: '/tmp/planning-repo',
      },
      targetDir: '/tmp',
    })

    const unwrapped = unwrapPhase1(result)
    expect(unwrapped.setupScope).toBe('workspace')
    expect(unwrapped.planningRepoPath).toBe(path.resolve('/tmp/planning-repo'))
    expect(unwrapped.projectName).toBe('planning-repo')
  })

  it('Phase 1: non-interactive workspace missing --planning-repo throws', async () => {
    await expect(
      runPhase1({
        interactive: false,
        prior: {},
        cliOverrides: {
          scope: 'workspace',
          tools: ['opencode'],
          name: 'workspace-name',
        },
        targetDir: '/tmp',
      }),
    ).rejects.toThrow('--planning-repo is required in non-interactive mode when --scope=workspace')
  })

  it('Phase 1: workspace parent-directory scan detects repo types', async () => {
    const workspaceRoot = makeTempDir('ai-setup-phase1-workspace-')
    const planningRepoDir = path.join(workspaceRoot, 'planning-repo')
    const railsRepoDir = path.join(workspaceRoot, 'fedora')
    const nextRepoDir = path.join(workspaceRoot, 'creator-checkout')

    mkdirSync(path.join(planningRepoDir, '.git'), { recursive: true })
    mkdirSync(path.join(railsRepoDir, '.git'), { recursive: true })
    mkdirSync(path.join(nextRepoDir, '.git'), { recursive: true })

    writeFileSync(path.join(railsRepoDir, 'Gemfile'), 'source "https://rubygems.org"')
    mkdirSync(path.join(railsRepoDir, 'config'), { recursive: true })
    writeFileSync(path.join(railsRepoDir, 'config', 'routes.rb'), 'Rails.application.routes.draw do end')

    writeFileSync(path.join(nextRepoDir, 'package.json'), JSON.stringify({ dependencies: { next: '14.0.0' } }))

    const result = await runPhase1({
      interactive: false,
      prior: {},
      cliOverrides: {
        scope: 'workspace',
        tools: ['opencode'],
        name: 'workspace-name',
        planningRepo: planningRepoDir,
        repos: ['..'],
      },
      targetDir: workspaceRoot,
    })

    expect(unwrapPhase1(result).repos).toEqual([
      { name: 'creator-checkout', path: '../creator-checkout', type: 'nextjs-typescript' },
      { name: 'fedora', path: '../fedora', type: 'ruby-rails' },
    ])

    rmSync(workspaceRoot, { recursive: true, force: true })
  })

  it('Phase 1: workspace scan filters non-git directories and warns', async () => {
    const workspaceRoot = makeTempDir('ai-setup-phase1-filter-')
    const planningRepoDir = path.join(workspaceRoot, 'planning-repo')
    const gitRepoDir = path.join(workspaceRoot, 'git-repo')
    const nonGitDir = path.join(workspaceRoot, 'not-a-repo')

    mkdirSync(path.join(planningRepoDir, '.git'), { recursive: true })
    mkdirSync(path.join(gitRepoDir, '.git'), { recursive: true })
    mkdirSync(nonGitDir, { recursive: true })

    const result = await runPhase1({
      interactive: false,
      prior: {},
      cliOverrides: {
        scope: 'workspace',
        tools: ['opencode'],
        name: 'workspace-name',
        planningRepo: planningRepoDir,
        repos: ['..'],
      },
      targetDir: workspaceRoot,
    })

    expect(unwrapPhase1(result).repos).toEqual([{ name: 'git-repo', path: '../git-repo', type: 'unknown' }])
    expect(p.note).toHaveBeenCalledWith('Skipped non-git directories: not-a-repo')

    rmSync(workspaceRoot, { recursive: true, force: true })
  })

  it('Phase 1: non-interactive missing --scope throws required in non-interactive mode', async () => {
    await expect(
      runPhase1({
        interactive: false,
        prior: {},
        cliOverrides: { tools: ['opencode'], name: 'my-project' },
        targetDir: '/tmp',
      }),
    ).rejects.toThrow('required in non-interactive mode')
  })

  it('Phase 1: non-interactive missing --tools throws required in non-interactive mode', async () => {
    await expect(
      runPhase1({
        interactive: false,
        prior: {},
        cliOverrides: { scope: 'project', name: 'my-project' },
        targetDir: '/tmp',
      }),
    ).rejects.toThrow('required in non-interactive mode')
  })

  it("Phase 3: non-interactive + force returns strategy 'backup-and-replace'", async () => {
    const result = await runPhase3({
      interactive: false,
      force: true,
      targetDir: '/tmp',
      plannedFiles: [],
    })

    expect(result.strategy).toBe('backup-and-replace')
    expect(result.perFileOverrides.size).toBe(0)
  })

  it("Phase 3: non-interactive without force returns strategy 'skip'", async () => {
    const result = await runPhase3({
      interactive: false,
      targetDir: '/tmp',
      plannedFiles: [],
    })

    expect(result.strategy).toBe('skip')
    expect(result.perFileOverrides.size).toBe(0)
  })
})
