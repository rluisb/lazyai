import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { beforeEach, describe, expect, it, vi } from 'vitest'

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

import * as p from '@clack/prompts'
import { GO_BACK } from '../utils/ui.js'
import { runPhase1 } from '../wizard/phase1-context.js'
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
      projectName: 'my-project',
    })
  })

  it('Phase 1: interactive can opt into orchestrator as an integration', async () => {
    vi.mocked(p.select).mockResolvedValueOnce('project')
    vi.mocked(p.multiselect)
      .mockResolvedValueOnce(['opencode'])
      .mockResolvedValueOnce(['orchestrator'])
    vi.mocked(p.text).mockResolvedValueOnce('my-project')

    const result = unwrapPhase1(await runPhase1({
      interactive: true,
      prior: {},
      cliOverrides: { cliTools: [] },
      targetDir: process.cwd(),
    }))

    expect(result.enableServers).toEqual(['orchestrator'])
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
      cliOverrides: { scope: 'global', tools: ['opencode'] },
      targetDir: '/tmp',
    })

    expect(unwrapPhase1(result).projectName).toBe('global')
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
