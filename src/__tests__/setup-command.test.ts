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

  it('setup --scan emits current and desired state inventory', async () => {
    const repoDir = makeTempRepo('ai-setup-setup-scan-')
    const resolvedRepoDir = realPath(repoDir)
    const homeDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-home-'))
    process.chdir(repoDir)
    process.env.HOME = homeDir

    fs.mkdirSync(path.join(repoDir, '.claude', 'agents'), { recursive: true })
    fs.writeFileSync(path.join(repoDir, '.claude', 'settings.json'), JSON.stringify({ version: '1.2.3' }), 'utf-8')

    const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
    await runSetup(['--scan'])

    const result = JSON.parse(String(logSpy.mock.calls[0]?.[0])) as {
      currentState: {
        sharedPaths: Array<{ id: string; exists: boolean }>
        targets: Array<{
          id: string
          detections: Array<{
            scope: string
            origin: string
            status: string
            state: string
            version: string
            observedFiles: string[]
          }>
        }>
      }
      desiredState: {
        sharedPaths: Array<{ id: string }>
        targets: Array<{
          id: string
          candidateRoots: Array<{ scope: string; expectedFiles: string[] }>
        }>
      }
    }

    expect(result.currentState.sharedPaths).toEqual([
      { id: 'global-ai-setup', exists: false, path: path.join(homeDir, '.ai-setup') },
      { id: 'project-ai', exists: false, path: path.join(resolvedRepoDir, '.ai') },
    ])
    expect(result.desiredState.sharedPaths.map(({ id }) => id)).toEqual(['global-ai-setup', 'project-ai'])

    expect(result.currentState.targets.find(({ id }) => id === 'claude-code')?.detections).toContainEqual({
      scope: 'project',
      origin: 'project',
      rootPath: path.join(resolvedRepoDir, '.claude'),
      status: 'detected',
      state: 'adoptable',
      version: '1.2.3',
      observedFiles: ['agents', 'settings.json'],
    })

    expect(result.desiredState.targets.find(({ id }) => id === 'claude-code')?.candidateRoots).toContainEqual({
      scope: 'project',
      origin: 'project',
      rootPath: path.join(resolvedRepoDir, '.claude'),
      expectedFiles: ['agents', 'commands', 'output-styles', 'settings.json', 'settings.local.json', 'skills'],
    })
  })

  it('setup --scan marks missing detections as missing state', async () => {
    const repoDir = makeTempRepo('ai-setup-setup-scan-missing-')
    const resolvedRepoDir = realPath(repoDir)
    const homeDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-home-'))
    process.chdir(repoDir)
    process.env.HOME = homeDir

    const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
    await runSetup(['--scan'])

    const result = JSON.parse(String(logSpy.mock.calls[0]?.[0])) as {
      currentState: { targets: Array<{ id: string; detections: Array<{ scope: string; status: string; state: string; observedFiles: string[] }> }> }
    }

    expect(result.currentState.targets.find(({ id }) => id === 'opencode')?.detections.find(({ scope }) => scope === 'project')).toEqual({
      scope: 'project',
      origin: 'project',
      rootPath: path.join(resolvedRepoDir, '.opencode'),
      status: 'missing',
      state: 'missing',
      version: 'unknown',
      observedFiles: [],
    })
  })

  it('setup --scan preserves count-root-only copilot global detection', async () => {
    const repoDir = makeTempRepo('ai-setup-setup-scan-copilot-')
    const homeDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-home-'))
    process.chdir(repoDir)
    process.env.HOME = homeDir
    fs.mkdirSync(path.join(homeDir, '.copilot'), { recursive: true })

    const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
    await runSetup(['--scan'])

    const result = JSON.parse(String(logSpy.mock.calls[0]?.[0])) as {
      currentState: { targets: Array<{ id: string; detections: Array<{ scope: string; status: string; state: string; observedFiles: string[] }> }> }
    }

    expect(result.currentState.targets.find(({ id }) => id === 'copilot')?.detections.find(({ scope }) => scope === 'global')).toEqual({
      scope: 'global',
      origin: 'global',
      rootPath: path.join(homeDir, '.copilot'),
      status: 'detected',
      state: 'adoptable',
      version: 'unknown',
      observedFiles: [],
    })
  })

  it('setup --scan extracts valid reusable agent metadata', async () => {
    const repoDir = makeTempRepo('ai-setup-setup-scan-agent-')
    const resolvedRepoDir = realPath(repoDir)
    process.chdir(repoDir)

    const agentDir = path.join(resolvedRepoDir, '.ai', 'agents', 'reviewer')
    fs.mkdirSync(agentDir, { recursive: true })
    fs.writeFileSync(path.join(agentDir, 'AGENT.md'), `---
title: Review Agent
description: Finds issues
tools:
  - write
  - bash
  -
---

# Reviewer

Checks pull requests.
`, 'utf-8')
    fs.writeFileSync(path.join(agentDir, 'mcp.json'), JSON.stringify({ mcpServers: { zed: {}, alpha: {} } }), 'utf-8')

    const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
    await runSetup(['--scan'])

    const result = JSON.parse(String(logSpy.mock.calls[0]?.[0])) as {
      currentState: { agents?: Array<{ id: string; status: string; title?: string; description?: string; tools?: string[]; mcp?: { serverNames?: string[]; serverCount: number } }> }
    }

    expect(result.currentState.agents).toEqual([
      {
        id: 'reviewer',
        directory: agentDir,
        promptPath: path.join(agentDir, 'AGENT.md'),
        status: 'detected',
        title: 'Review Agent',
        description: 'Finds issues',
        tools: ['bash', 'write'],
        mcp: {
          configPath: path.join(agentDir, 'mcp.json'),
          scoped: true,
          serverNames: ['alpha', 'zed'],
          serverCount: 2,
        },
      },
    ])
  })

  it('setup --scan reports invalid reusable agents', async () => {
    const repoDir = makeTempRepo('ai-setup-setup-scan-agent-invalid-')
    const resolvedRepoDir = realPath(repoDir)
    process.chdir(repoDir)

    const badIdDir = path.join(resolvedRepoDir, '.ai', 'agents', 'Bad_ID')
    fs.mkdirSync(badIdDir, { recursive: true })
    fs.writeFileSync(path.join(badIdDir, 'AGENT.md'), '# Heading\n\nBody', 'utf-8')

    const missingAgentMdDir = path.join(resolvedRepoDir, '.ai', 'agents', 'missing-agent')
    fs.mkdirSync(missingAgentMdDir, { recursive: true })

    const badMcpDir = path.join(resolvedRepoDir, '.ai', 'agents', 'bad-mcp')
    fs.mkdirSync(badMcpDir, { recursive: true })
    fs.writeFileSync(path.join(badMcpDir, 'AGENT.md'), '# Bad MCP\n\nBody', 'utf-8')
    fs.writeFileSync(path.join(badMcpDir, 'mcp.json'), JSON.stringify({ mcpServers: {}, extra: true }), 'utf-8')

    const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
    await runSetup(['--scan'])

    const result = JSON.parse(String(logSpy.mock.calls[0]?.[0])) as {
      currentState: { agents?: Array<{ id: string; status: string; reasons?: string[] }> }
    }

    expect(result.currentState.agents).toEqual([
      {
        id: 'Bad_ID',
        directory: badIdDir,
        promptPath: path.join(badIdDir, 'AGENT.md'),
        status: 'invalid',
        title: 'Heading',
        description: 'Body',
        reasons: ['invalid-agent-id'],
      },
      {
        id: 'bad-mcp',
        directory: badMcpDir,
        promptPath: path.join(badMcpDir, 'AGENT.md'),
        status: 'invalid',
        title: 'Bad MCP',
        description: 'Body',
        reasons: ['invalid-agent-mcp-schema'],
      },
      {
        id: 'missing-agent',
        directory: missingAgentMdDir,
        promptPath: path.join(missingAgentMdDir, 'AGENT.md'),
        status: 'invalid',
        reasons: ['missing-agent-md'],
      },
    ])
  })

  it('setup --list includes agents when .ai/agents directory exists', async () => {
    const repoDir = makeTempRepo('ai-setup-setup-list-agents-')
    const resolvedRepoDir = realPath(repoDir)
    const homeDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-home-'))
    process.chdir(repoDir)
    process.env.HOME = homeDir

    const agentDir = path.join(resolvedRepoDir, '.ai', 'agents', 'reviewer')
    fs.mkdirSync(agentDir, { recursive: true })
    fs.writeFileSync(path.join(agentDir, 'AGENT.md'), `---
title: Review Agent
description: Finds issues
tools:
  - write
  - bash
---
# Reviewer
`, 'utf-8')

    const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
    await runSetup(['--list'])

    const result = JSON.parse(String(logSpy.mock.calls[0]?.[0])) as {
      mode: string
      agents?: Array<{ id: string; status: string; title?: string }>
    }

    expect(result.mode).toBe('list')
    expect(result.agents).toBeDefined()
    expect(result.agents?.some(({ id }) => id === 'reviewer')).toBe(true)
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

  it('rejects setup --scan with --tool', async () => {
    const repoDir = makeTempRepo('ai-setup-setup-scan-tool-')
    process.chdir(repoDir)

    await expect(runSetup(['--scan', '--tool', 'opencode'])).rejects.toThrow(
      '--tool, --all, and --global are only supported with --list or --dry-run',
    )
  })

  it('rejects --adopt without --scan', async () => {
    const repoDir = makeTempRepo('ai-setup-setup-adopt-no-scan-')
    process.chdir(repoDir)

    await expect(runSetup(['--adopt'])).rejects.toThrow('--adopt and --import require --scan')
  })

  it('adopts adoptable targets and rescans them as managed', async () => {
    const repoDir = makeTempRepo('ai-setup-setup-adopt-')
    const resolvedRepoDir = realPath(repoDir)
    const homeDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-home-'))
    process.chdir(repoDir)
    process.env.HOME = homeDir
    fs.mkdirSync(path.join(repoDir, '.claude'), { recursive: true })
    fs.writeFileSync(path.join(repoDir, '.claude', 'settings.json'), JSON.stringify({ version: '1.0.0' }), 'utf-8')

    const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
    await runSetup(['--scan', '--adopt'])

    const result = JSON.parse(String(logSpy.mock.calls[0]?.[0])) as {
      currentState: { targets: Array<{ id: string; detections: Array<{ scope: string; state: string }> }> }
      operation: { mode: string; registryPath: string; importRoot: string; adopted?: Array<{ targetId: string; rootPath: string }> }
    }

    expect(result.operation.mode).toBe('adopt')
    expect(result.operation.registryPath).toBe(path.join(homeDir, '.ai-setup', 'config', 'setup-scan-registry.json'))
    expect(result.operation.importRoot).toBe(path.join(homeDir, '.ai-setup', 'imports'))
    expect(result.operation.adopted).toContainEqual({
      targetId: 'claude-code',
      scope: 'project',
      origin: 'project',
      rootPath: path.join(resolvedRepoDir, '.claude'),
    })
    expect(result.currentState.targets.find(({ id }) => id === 'claude-code')?.detections.find(({ scope }) => scope === 'project')?.state).toBe('managed')

    const registry = JSON.parse(
      fs.readFileSync(path.join(homeDir, '.ai-setup', 'config', 'setup-scan-registry.json'), 'utf-8'),
    ) as { resources: Array<{ targetId: string; state: string; rootPath: string }> }
    expect(registry.resources).toContainEqual(expect.objectContaining({
      targetId: 'claude-code',
      state: 'managed',
      rootPath: path.join(resolvedRepoDir, '.claude'),
    }))
  })

  it('adopt skips missing targets with not-adoptable', async () => {
    const repoDir = makeTempRepo('ai-setup-setup-adopt-skip-')
    const resolvedRepoDir = realPath(repoDir)
    const homeDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-home-'))
    process.chdir(repoDir)
    process.env.HOME = homeDir

    const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
    await runSetup(['--scan', '--adopt'])

    const result = JSON.parse(String(logSpy.mock.calls[0]?.[0])) as {
      operation: { skipped?: Array<{ targetId: string; scope: string; rootPath: string; state: string; reason: string }> }
    }

    expect(result.operation.skipped).toContainEqual({
      targetId: 'claude-code',
      scope: 'project',
      origin: 'project',
      rootPath: path.join(resolvedRepoDir, '.claude'),
      state: 'missing',
      reason: 'not-adoptable',
    })
  })

  it('imports observed files into the ai-setup import store and records the import', async () => {
    const repoDir = makeTempRepo('ai-setup-setup-import-')
    const homeDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-home-'))
    process.chdir(repoDir)
    process.env.HOME = homeDir
    fs.mkdirSync(path.join(repoDir, '.claude'), { recursive: true })
    fs.writeFileSync(path.join(repoDir, '.claude', 'settings.json'), JSON.stringify({ version: '2.0.0' }), 'utf-8')

    const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
    await runSetup(['--scan', '--import'])

    const result = JSON.parse(String(logSpy.mock.calls[0]?.[0])) as {
      operation: {
        mode: string
        registryPath: string
        importRoot: string
        imported?: Array<{ targetId: string; sourcePath: string; destinationPath: string }>
      }
    }

    expect(result.operation.mode).toBe('import')
    const imported = result.operation.imported?.find(({ targetId, sourcePath }) => targetId === 'claude-code' && sourcePath.endsWith('settings.json'))
    expect(imported).toBeDefined()
    expect(fs.readFileSync(imported?.destinationPath ?? '', 'utf-8')).toBe(fs.readFileSync(path.join(repoDir, '.claude', 'settings.json'), 'utf-8'))

    const registry = JSON.parse(
      fs.readFileSync(path.join(homeDir, '.ai-setup', 'config', 'setup-scan-registry.json'), 'utf-8'),
    ) as { imports: Array<{ targetId: string; importedPaths: string[]; destinationRoot: string }> }
    expect(registry.imports).toContainEqual(expect.objectContaining({
      targetId: 'claude-code',
      importedPaths: ['settings.json'],
      destinationRoot: path.dirname(imported?.destinationPath ?? ''),
    }))
  })

  it('import skips missing and conflict detections as not-importable', async () => {
    const repoDir = makeTempRepo('ai-setup-setup-import-skip-')
    const homeDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-home-'))
    process.chdir(repoDir)
    process.env.HOME = homeDir
    fs.mkdirSync(path.join(repoDir, '.claude'), { recursive: true })
    fs.writeFileSync(path.join(repoDir, '.claude', 'settings.json'), JSON.stringify({ version: '1.0.0' }), 'utf-8')

    await runSetup(['--scan', '--adopt'])
    fs.writeFileSync(path.join(repoDir, '.claude', 'settings.json'), JSON.stringify({ version: '9.9.9' }), 'utf-8')

    const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
    await runSetup(['--scan', '--import'])

    const result = JSON.parse(String(logSpy.mock.calls[0]?.[0])) as {
      operation: { skipped?: Array<{ targetId: string; scope: string; state: string; reason: string }> }
    }

    expect(result.operation.skipped).toEqual(expect.arrayContaining([
      expect.objectContaining({
        targetId: 'claude-code',
        scope: 'project',
        state: 'conflict',
        reason: 'not-importable',
      }),
      expect.objectContaining({
        targetId: 'opencode',
        scope: 'project',
        state: 'missing',
        reason: 'not-importable',
      }),
    ]))
  })

  it('supports combined adopt and import mode output', async () => {
    const repoDir = makeTempRepo('ai-setup-setup-adopt-import-')
    const homeDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-home-'))
    process.chdir(repoDir)
    process.env.HOME = homeDir
    fs.mkdirSync(path.join(repoDir, '.claude'), { recursive: true })
    fs.writeFileSync(path.join(repoDir, '.claude', 'settings.json'), JSON.stringify({ version: '3.0.0' }), 'utf-8')

    const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
    await runSetup(['--scan', '--adopt', '--import'])

    const result = JSON.parse(String(logSpy.mock.calls[0]?.[0])) as {
      operation: {
        mode: string
        adopted?: Array<{ targetId: string }>
        imported?: Array<{ targetId: string }>
      }
    }

    expect(result.operation.mode).toBe('adopt-import')
    expect(result.operation.adopted?.some(({ targetId }) => targetId === 'claude-code')).toBe(true)
    expect(result.operation.imported?.some(({ targetId }) => targetId === 'claude-code')).toBe(true)
  })

  it('reports managed registry conflicts when an adopted file changes', async () => {
    const repoDir = makeTempRepo('ai-setup-setup-conflict-')
    process.chdir(repoDir)
    const homeDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-home-'))
    process.env.HOME = homeDir
    fs.mkdirSync(path.join(repoDir, '.claude'), { recursive: true })
    fs.writeFileSync(path.join(repoDir, '.claude', 'settings.json'), JSON.stringify({ version: '1.0.0' }), 'utf-8')

    await runSetup(['--scan', '--adopt'])
    fs.writeFileSync(path.join(repoDir, '.claude', 'settings.json'), JSON.stringify({ version: '1.0.1' }), 'utf-8')

    const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
    await runSetup(['--scan'])

    const result = JSON.parse(String(logSpy.mock.calls[0]?.[0])) as {
      currentState: {
        targets: Array<{ id: string; detections: Array<{ scope: string; state: string; reasons?: string[] }> }>
      }
    }

    expect(result.currentState.targets.find(({ id }) => id === 'claude-code')?.detections.find(({ scope }) => scope === 'project')).toEqual(
      expect.objectContaining({
        state: 'conflict',
        reasons: ['changed-path:settings.json'],
      }),
    )
  })
})
