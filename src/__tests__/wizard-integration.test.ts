import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { existsSync, mkdtempSync, readFileSync, rmSync } from 'node:fs'
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
        type: 'project',
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
    expect(manifest.selections.docsDirs).toHaveLength(10)
  })

  it('creates only opencode files when only opencode selected', async () => {
    await runWizard({
      interactive: false,
      cliOverrides: {
        type: 'project',
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
        type: 'project',
        tools: ['opencode'],
        name: 'test-project',
      },
      targetDir: tempDir,
    })

    await runWizard({
      interactive: false,
      cliOverrides: {
        type: 'project',
        tools: ['opencode'],
        name: 'test-project',
      },
      targetDir: tempDir,
    })

    const manifest = JSON.parse(readFileSync(path.join(tempDir, '.ai-setup.json'), 'utf-8'))
    expect(manifest.selections).toBeDefined()
    expect(manifest.config.tools).toEqual(['opencode'])
  })
})
