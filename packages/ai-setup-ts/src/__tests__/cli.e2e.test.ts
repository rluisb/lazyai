import { execSync } from 'node:child_process'
import fs from 'node:fs'
import os from 'node:os'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const binPath = join(__dirname, '../../bin/ai-setup.js')

function getFailureOutput(command: string, cwd?: string): string {
  try {
    execSync(command, { stdio: 'pipe', cwd })
    throw new Error('Expected command to fail')
  } catch (err) {
    const error = err as Error & { stdout?: Buffer; stderr?: Buffer }
    return `${error.stdout?.toString() ?? ''}${error.stderr?.toString() ?? ''}`
  }
}

describe('CLI End-to-End', () => {
  it('shows help output when run with --help', () => {
    const output = execSync(`node ${binPath} --help`).toString()
    expect(output).toContain('AI development environment scaffold')
    expect(output).toContain('init [options]')
    expect(output).toContain('update [options]')
    expect(output).toContain('doctor')
    expect(output).toContain('eject')
    expect(output).toContain('create [options]')
  })

  it('shows version when run with --version', () => {
    const output = execSync(`node ${binPath} --version`).toString()
    expect(output.trim()).toMatch(/^[0-9]+\.[0-9]+\.[0-9]+$/)
  })

  it('accepts --verbose flag', () => {
    const output = execSync(`node ${binPath} --verbose --help`).toString()
    expect(output).toContain('--verbose')
  })

  it('fails gracefully on unknown command', () => {
    try {
      execSync(`node ${binPath} potato`, { stdio: 'pipe' })
      expect.fail('Should have thrown an error')
    } catch (err) {
      const error = err as Error & { status?: number; stderr?: Buffer }
      expect(error.status).toBe(1)
      expect(error.stderr?.toString()).toContain("error: unknown command 'potato'")
    }
  })

  it('shows actionable help for an invalid migration strategy', () => {
    const output = getFailureOutput(`node ${binPath} import --strategy potato`)

    expect(output).toContain('Unknown merge strategy "potato"')
    expect(output).toContain('Supported strategies:')
    expect(output).toContain('smart')
    expect(output).toContain('preserve')
    expect(output).toContain('replace')
    expect(output).toContain('append')
  })

  it('shows clearer guidance when no supported setup is detected', () => {
    const tempDir = fs.mkdtempSync(join(os.tmpdir(), 'ai-setup-empty-'))

    const output = getFailureOutput(`node ${binPath} import --yes`, tempDir)

    expect(output).toContain('No supported AI setup detected')
    expect(output).toContain('Expected markers include:')
    expect(output).toContain('ai-setup import /path/to/project')
  })
})
