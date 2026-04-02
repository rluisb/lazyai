import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mkdirSync, mkdtempSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'

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
import { runPhase1 } from '../wizard/phase1-context.js'
import { runPhase7 } from '../wizard/phase7-conflicts.js'

function makeTempDir(prefix: string): string {
  return mkdtempSync(path.join(tmpdir(), prefix))
}

describe('wizard phases 1 and 7', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(p.isCancel).mockReturnValue(false)
  })

  it('Phase 1: non-interactive with all CLI overrides returns correct values', async () => {
    const result = await runPhase1({
      interactive: false,
      prior: {},
      cliOverrides: { scope: 'project', tools: ['pi', 'opencode'], name: 'my-project' },
      targetDir: '/tmp',
    })

    expect(result).toEqual({
      setupScope: 'project',
      tools: ['pi', 'opencode'],
      projectName: 'my-project',
    })
  })

  it('Phase 1: non-interactive with scope=global returns setupScope=global', async () => {
    const result = await runPhase1({
      interactive: false,
      prior: {},
      cliOverrides: { scope: 'global', tools: ['pi'], name: 'my-project' },
      targetDir: '/tmp',
    })

    expect(result.setupScope).toBe('global')
  })

  it('Phase 1: non-interactive with scope=global defaults projectName=global', async () => {
    const result = await runPhase1({
      interactive: false,
      prior: {},
      cliOverrides: { scope: 'global', tools: ['opencode'] },
      targetDir: '/tmp',
    })

    expect(result.projectName).toBe('global')
  })

  it('Phase 1: non-interactive with scope=workspace returns setupScope=workspace', async () => {
    const result = await runPhase1({
      interactive: false,
      prior: {},
      cliOverrides: {
        scope: 'workspace',
        tools: ['pi'],
        name: 'my-project',
        planningRepo: '/tmp/planning-repo',
      },
      targetDir: '/tmp',
    })

    expect(result.setupScope).toBe('workspace')
    expect(result.planningRepoPath).toBe(path.resolve('/tmp/planning-repo'))
    expect(result.projectName).toBe('planning-repo')
  })

  it('Phase 1: non-interactive workspace missing --planning-repo throws', async () => {
    await expect(
      runPhase1({
        interactive: false,
        prior: {},
        cliOverrides: {
          scope: 'workspace',
          tools: ['pi'],
          name: 'workspace-name',
        },
        targetDir: '/tmp',
      }),
    ).rejects.toThrow('--planning-repo is required in non-interactive mode when --scope=workspace')
  })

  it('Phase 1: non-interactive missing --scope throws required in non-interactive mode', async () => {
    await expect(
      runPhase1({
        interactive: false,
        prior: {},
        cliOverrides: { tools: ['pi'], name: 'my-project' },
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

  it("Phase 7: non-interactive + force returns strategy 'backup-and-replace'", async () => {
    const result = await runPhase7({
      interactive: false,
      force: true,
      targetDir: '/tmp',
      plannedFiles: [],
    })

    expect(result.strategy).toBe('backup-and-replace')
    expect(result.perFileOverrides.size).toBe(0)
  })

  it("Phase 7: non-interactive without force returns strategy 'skip'", async () => {
    const result = await runPhase7({
      interactive: false,
      targetDir: '/tmp',
      plannedFiles: [],
    })

    expect(result.strategy).toBe('skip')
    expect(result.perFileOverrides.size).toBe(0)
  })
})
