import { describe, expect, it } from 'vitest'
import { existsSync, mkdirSync, writeFileSync, rmSync } from 'node:fs'
import { join } from 'node:path'
import { tmpdir } from 'node:os'
import { parseToml, loadConfig } from '../utils/toml.js'

function makeTempDir(prefix: string): string {
  const dir = join(tmpdir(), `${prefix}-${Date.now()}`)
  mkdirSync(dir, { recursive: true })
  return dir
}

describe('parseToml', () => {
  it('parses simple key-value pairs', () => {
    const result = parseToml('default_scope = "project"\ndefault_tools = ["opencode", "claude-code"]')
    expect(result.default_scope).toBe('project')
    expect(result.default_tools).toEqual(['opencode', 'claude-code'])
  })

  it('parses boolean and numeric values', () => {
    const result = parseToml('install_mode = "symlink"\nenabled = true\nport = 8080')
    expect(result.install_mode).toBe('symlink')
    expect(result.enabled).toBe(true)
    expect(result.port).toBe(8080)
  })

  it('parses section headers', () => {
    const result = parseToml('[wizard]\npreset = "minimal"\nshow_preview = true')
    expect(result.wizard).toEqual({ preset: 'minimal', show_preview: true })
  })

  it('handles comments and blank lines', () => {
    const result = parseToml('# This is a comment\ndefault_scope = "global"\n\n# Another comment\ndefault_tools = ["opencode"]')
    expect(result.default_scope).toBe('global')
    expect(result.default_tools).toEqual(['opencode'])
  })

  it('handles empty arrays', () => {
    const result = parseToml('default_tools = []')
    expect(result.default_tools).toEqual([])
  })

  it('handles single-quoted strings', () => {
    const result = parseToml("project_name = 'my-app'")
    expect(result.project_name).toBe('my-app')
  })

  it('handles empty input gracefully', () => {
    const result = parseToml('')
    expect(result).toEqual({})
    const result2 = parseToml('# just a comment')
    expect(result2).toEqual({})
  })
})

describe('loadConfig', () => {
  it('returns empty config when no files exist', () => {
    const dir = makeTempDir('ai-setup-toml-empty')
    const config = loadConfig(dir)
    expect(config).toEqual({})
    rmSync(dir, { recursive: true, force: true })
  })

  it('loads project config', () => {
    const dir = makeTempDir('ai-setup-toml-project')
    writeFileSync(join(dir, '.ai-setup.toml'), 'default_scope = "workspace"\ndefault_tools = ["opencode"]\ninstall_mode = "symlink"')
    const config = loadConfig(dir)
    expect(config.default_scope).toBe('workspace')
    expect(config.default_tools).toEqual(['opencode'])
    expect(config.install_mode).toBe('symlink')
    rmSync(dir, { recursive: true, force: true })
  })

  it('handles malformed TOML gracefully', () => {
    const dir = makeTempDir('ai-setup-toml-bad')
    writeFileSync(join(dir, '.ai-setup.toml'), 'this is not valid toml {{{')
    const config = loadConfig(dir)
    expect(config).toEqual({}) // Fallback to empty
    rmSync(dir, { recursive: true, force: true })
  })
})
