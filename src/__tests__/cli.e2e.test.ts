import { describe, it, expect } from 'vitest'
import { execSync } from 'node:child_process'
import { join } from 'node:path'

const binPath = join(__dirname, '../../bin/ai-setup.js')

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
    } catch (err: any) {
      expect(err.status).toBe(1)
      expect(err.stderr.toString()).toContain("error: unknown command 'potato'")
    }
  })
})
