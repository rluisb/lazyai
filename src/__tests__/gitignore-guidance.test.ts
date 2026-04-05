import { mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { checkGitignoreGuidance } from '../scaffold/gitignore.js'

describe('gitignore guidance', () => {
  let tempDir: string

  beforeEach(() => {
    tempDir = mkdtempSync(path.join(tmpdir(), 'ai-setup-gitignore-guidance-'))
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
    vi.restoreAllMocks()
  })

  it('logs guidance when .gitignore is missing', () => {
    const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})

    checkGitignoreGuidance(tempDir)

    const output = logSpy.mock.calls.map((call) => call.map((value) => String(value)).join(' ')).join('\n')
    expect(output).toContain('Consider creating a .gitignore')
    expect(output).toContain('.ai/memory/')
  })

  it('logs guidance when .gitignore does not include .ai/memory', () => {
    writeFileSync(path.join(tempDir, '.gitignore'), 'node_modules/\n', 'utf-8')
    const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})

    checkGitignoreGuidance(tempDir)

    const output = logSpy.mock.calls.map((call) => call.map((value) => String(value)).join(' ')).join('\n')
    expect(output).toContain('Consider adding to .gitignore')
    expect(output).toContain('.ai/memory/')
  })

  it('does not log guidance when .gitignore already includes .ai/memory', () => {
    writeFileSync(path.join(tempDir, '.gitignore'), '.ai/memory/\n', 'utf-8')
    const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})

    checkGitignoreGuidance(tempDir)

    expect(logSpy).not.toHaveBeenCalled()
  })
})
