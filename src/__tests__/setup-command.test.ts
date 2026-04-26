import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createProgram } from '../cli.js'

function makeTempRepo(prefix: string): string {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), prefix))
  fs.mkdirSync(path.join(tempDir, '.git'), { recursive: true })
  return tempDir
}

function realPath(value: string): string {
  return fs.realpathSync.native(value)
}

async function runSetup(args: string[]): Promise<void> {
  const program = createProgram()
  await program.parseAsync(['node', 'ai-setup', 'setup', ...args])
}

describe('setup command parity', () => {
  let originalCwd: string
  let originalHome: string | undefined

  beforeEach(() => {
    originalCwd = process.cwd()
    originalHome = process.env.HOME
  })

  afterEach(() => {
    process.chdir(originalCwd)
    if (originalHome === undefined) {
      delete process.env.HOME
    } else {
      process.env.HOME = originalHome
    }
    vi.restoreAllMocks()
  })

  it('setup --list emits all candidate roots by default', async () => {
    const repoDir = makeTempRepo('ai-setup-setup-list-')
    const homeDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-home-'))
    process.chdir(repoDir)
    process.env.HOME = homeDir

    const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
    await runSetup(['--list'])

    const result = JSON.parse(String(logSpy.mock.calls[0]?.[0])) as {
      mode: string
      sharedPaths: Array<{ id: string }>
      targets: Array<{ id: string; supportedScopes: string[]; candidateRoots: Array<{ scope: string; rootPath: string }> }>
      agents?: unknown[]
      scopeFilter?: string
    }

    expect(result.mode).toBe('list')
    expect(result.scopeFilter).toBeUndefined()
    expect(result.sharedPaths.map(({ id }) => id)).toEqual(['global-ai-setup', 'project-ai'])
    expect(result.targets.map(({ id }) => id)).toEqual(['claude-code', 'codex', 'copilot', 'gemini', 'opencode', 'pi'])
    expect(result.targets.find(({ id }) => id === 'pi')?.supportedScopes).toEqual(['project', 'workspace'])
    expect(result.targets.find(({ id }) => id === 'claude-code')?.candidateRoots.map(({ scope }) => scope)).toEqual([
      'global',
      'project',
      'workspace',
    ])
    expect(result.targets.find(({ id }) => id === 'opencode')?.candidateRoots[0]?.rootPath).toBe(path.join(homeDir, '.config', 'opencode'))
    expect(result).not.toHaveProperty('agents')
  })

  it('setup --list --global filters shared paths and roots to global scope', async () => {
    const repoDir = makeTempRepo('ai-setup-setup-list-global-')
    const homeDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-home-'))
    process.chdir(repoDir)
    process.env.HOME = homeDir

    const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
    await runSetup(['--list', '--global', '--tool', 'claude-code'])

    const result = JSON.parse(String(logSpy.mock.calls[0]?.[0])) as {
      scopeFilter: string
      sharedPaths: Array<{ id: string }>
      targets: Array<{ id: string; supportedScopes: string[]; candidateRoots: Array<{ scope: string; rootPath: string }> }>
    }

    expect(result.scopeFilter).toBe('global')
    expect(result.sharedPaths.map(({ id }) => id)).toEqual(['global-ai-setup'])
    expect(result.targets).toHaveLength(1)
    expect(result.targets[0]?.id).toBe('claude-code')
    expect(result.targets[0]?.supportedScopes).toEqual(['global'])
    expect(result.targets[0]?.candidateRoots).toEqual([
      {
        scope: 'global',
        origin: 'global',
        rootPath: path.join(homeDir, '.claude'),
        expectedFiles: ['agents', 'commands', 'output-styles', 'settings.json', 'settings.local.json', 'skills'],
      },
    ])
  })

  it('setup --dry-run defaults to project scope and preserves detected project files', async () => {
    const repoDir = makeTempRepo('ai-setup-setup-dry-run-')
    const homeDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-home-'))
    process.chdir(repoDir)
    process.env.HOME = homeDir
    fs.mkdirSync(path.join(repoDir, '.claude'), { recursive: true })
    fs.writeFileSync(path.join(repoDir, '.claude', 'settings.json'), '{}', 'utf-8')

    const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
    await runSetup(['--dry-run', '--tool', 'claude-code'])

    const result = JSON.parse(String(logSpy.mock.calls[0]?.[0])) as {
      mode: string
      scope: string
      sharedPaths: Array<{ id: string }>
      targets: Array<{
        id: string
        scope: string
        rootPath: string
        observedFiles?: string[]
        existingStatus: string
        existingState?: string
        action: string
      }>
    }

    expect(result.mode).toBe('dry-run')
    expect(result.scope).toBe('project')
    expect(result.sharedPaths.map(({ id }) => id)).toEqual(['project-ai'])
    expect(result.targets).toEqual([
      {
        id: 'claude-code',
        name: 'Claude Code',
        scope: 'project',
        origin: 'project',
        rootPath: realPath(path.join(repoDir, '.claude')),
        expectedFiles: ['agents', 'commands', 'output-styles', 'settings.json', 'settings.local.json', 'skills'],
        observedFiles: ['settings.json'],
        existingStatus: 'detected',
        existingState: 'adoptable',
        action: 'preserve-existing',
      },
    ])
  })

  it('setup --dry-run --global preserves count-root-only global detections', async () => {
    const repoDir = makeTempRepo('ai-setup-setup-dry-run-global-')
    const homeDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-home-'))
    process.chdir(repoDir)
    process.env.HOME = homeDir
    fs.mkdirSync(path.join(homeDir, '.copilot'), { recursive: true })

    const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
    await runSetup(['--dry-run', '--global', '--tool', 'copilot'])

    const result = JSON.parse(String(logSpy.mock.calls[0]?.[0])) as {
      scope: string
      sharedPaths: Array<{ id: string }>
      targets: Array<{
        id: string
        existingStatus: string
        existingState?: string
        action: string
        rootPath: string
      }>
    }

    expect(result.scope).toBe('global')
    expect(result.sharedPaths.map(({ id }) => id)).toEqual(['global-ai-setup'])
    expect(result.targets).toEqual([
      {
        id: 'copilot',
        name: 'GitHub Copilot CLI',
        scope: 'global',
        origin: 'global',
        rootPath: path.join(homeDir, '.copilot'),
        expectedFiles: ['mcp-config.json'],
        existingStatus: 'detected',
        existingState: 'adoptable',
        action: 'preserve-existing',
      },
    ])
  })

  it('rejects unknown setup tools', async () => {
    const repoDir = makeTempRepo('ai-setup-setup-unknown-')
    process.chdir(repoDir)

    await expect(runSetup(['--list', '--tool', 'unknown-tool'])).rejects.toThrow('unknown tool "unknown-tool"')
  })

  it('rejects unsupported Pi global setup filtering', async () => {
    const repoDir = makeTempRepo('ai-setup-setup-pi-global-')
    process.chdir(repoDir)

    await expect(runSetup(['--dry-run', '--global', '--tool', 'pi'])).rejects.toThrow('tool "pi" does not support scope "global"')
  })

  it('rejects combining --all with --tool', async () => {
    const repoDir = makeTempRepo('ai-setup-setup-all-tool-')
    process.chdir(repoDir)

    await expect(runSetup(['--list', '--all', '--tool', 'opencode'])).rejects.toThrow('--all cannot be combined with --tool')
  })
})
