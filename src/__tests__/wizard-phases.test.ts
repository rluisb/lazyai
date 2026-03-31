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
import { runPhase2 } from '../wizard/phase2-docs.js'
import { runPhase3 } from '../wizard/phase3-templates.js'
import { runPhase4 } from '../wizard/phase4-agents.js'
import { runPhase5 } from '../wizard/phase5-infra.js'
import { runPhase6 } from '../wizard/phase6-root.js'
import { runPhase7 } from '../wizard/phase7-conflicts.js'

function makeTempDir(prefix: string): string {
  return mkdtempSync(path.join(tmpdir(), prefix))
}

describe('wizard phases 1-7', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(p.isCancel).mockReturnValue(false)
  })

  it('Phase 1: non-interactive with all CLI overrides returns correct values', async () => {
    const result = await runPhase1({
      interactive: false,
      prior: {},
      cliOverrides: { type: 'project', tools: ['pi', 'opencode'], name: 'my-project' },
      targetDir: '/tmp',
    })

    expect(result).toEqual({
      setupType: 'project',
      tools: ['pi', 'opencode'],
      projectName: 'my-project',
    })
  })

  it('Phase 1: non-interactive missing --type throws required in non-interactive mode', async () => {
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
        cliOverrides: { type: 'project', name: 'my-project' },
        targetDir: '/tmp',
      }),
    ).rejects.toThrow('required in non-interactive mode')
  })

  it('Phase 2: non-interactive returns all 10 docs dirs', async () => {
    const result = await runPhase2({ interactive: false, prior: {} })

    expect(result.docsDirs).toEqual([
      'features',
      'bugfixes',
      'refactors',
      'tech-debt',
      'adrs',
      'memory',
      'prompts',
      'standards',
      'templates',
      'rules',
    ])
    expect(result.docsDirs).toHaveLength(10)
  })

  it('Phase 2: non-interactive returns all docs agents = all docs dirs', async () => {
    const result = await runPhase2({ interactive: false, prior: {} })

    expect(result.docsAgents).toEqual(result.docsDirs)
  })

  it('Phase 2: docsAgents is always subset of docsDirs', async () => {
    vi.mocked(p.multiselect)
      .mockResolvedValueOnce(['features', 'bugfixes'])
      .mockResolvedValueOnce(['features'])

    const result = await runPhase2({ interactive: true, prior: {} })

    expect(result.docsDirs).toEqual(['features', 'bugfixes'])
    expect(result.docsAgents).toEqual(['features'])
    expect(result.docsAgents.every((agent) => result.docsDirs.includes(agent))).toBe(true)
  })

  it('Phase 3: non-interactive returns all 8 templates and all 4 rules', async () => {
    const result = await runPhase3({ interactive: false, prior: {} })

    expect(result.templates).toEqual([
      'adr',
      'prd',
      'progress',
      'standard',
      'task',
      'tasks',
      'tech-debt',
      'techspec',
    ])
    expect(result.rules).toEqual(['cost', 'review', 'security', 'workflow'])
  })

  it('Phase 4: non-interactive returns all 6 agents, 4 skills, 5 prompts', async () => {
    const result = await runPhase4({ interactive: false, prior: {} })

    expect(result.agents).toEqual(['builder', 'documenter', 'planner', 'red-team', 'reviewer', 'scout'])
    expect(result.skills).toEqual(['implement', 'iterate', 'plan', 'research'])
    expect(result.prompts).toEqual(['compact', 'implement', 'local-example', 'plan', 'research'])
  })

  it('Phase 5: non-interactive returns all 4 infra items when .git exists', async () => {
    const tempDir = makeTempDir('ai-setup-phase5-with-git-')

    try {
      mkdirSync(path.join(tempDir, '.git'), { recursive: true })

      const result = await runPhase5({
        interactive: false,
        prior: {},
        targetDir: tempDir,
      })

      expect(result.infra).toEqual(['CODEOWNERS', 'pre-commit', 'compliance', 'KNOWLEDGE_MAP'])
    } finally {
      rmSync(tempDir, { recursive: true, force: true })
    }
  })

  it('Phase 5: non-interactive without .git excludes pre-commit from result', async () => {
    const tempDir = makeTempDir('ai-setup-phase5-no-git-')

    try {
      const result = await runPhase5({
        interactive: false,
        prior: {},
        targetDir: tempDir,
      })

      expect(result.infra).toEqual(['CODEOWNERS', 'compliance', 'KNOWLEDGE_MAP'])
      expect(result.infra).not.toContain('pre-commit')
    } finally {
      rmSync(tempDir, { recursive: true, force: true })
    }
  })

  it("Phase 6: non-interactive returns correct root files for tools ['opencode', 'claude-code']", async () => {
    const result = await runPhase6({
      interactive: false,
      tools: ['opencode', 'claude-code'],
      projectName: 'my-project',
    })

    expect(result.rootFiles).toEqual(['AGENTS.md', 'CLAUDE.md'])
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
