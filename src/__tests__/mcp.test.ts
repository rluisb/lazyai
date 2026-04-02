import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mkdtempSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { ensureDir, readFile, writeFile } from '../utils/files.js'
import type { FileRecord } from '../types.js'
import { scaffoldMcp } from '../scaffold/mcp.js'
import { compileMcp } from '../adapters/mcp-compiler.js'

function makeTempDir(prefix: string): string {
  return mkdtempSync(path.join(tmpdir(), prefix))
}

describe('MCP scaffold and compile', () => {
  let targetDir: string
  let libraryDir: string
  let fileRecords: FileRecord[]

  beforeEach(() => {
    targetDir = makeTempDir('ai-setup-mcp-target-')
    libraryDir = makeTempDir('ai-setup-mcp-library-')
    fileRecords = []

    ensureDir(path.join(libraryDir, 'mcp'))
    writeFile(
      path.join(libraryDir, 'mcp', 'catalog.json'),
      JSON.stringify(
        {
          servers: {
            stdioEnabled: {
              command: 'npx',
              args: ['-y', 'mcp-stdio-enabled'],
              env: { API_KEY: '${API_KEY}' },
              tools: ['alpha', 'beta'],
              enabled: true,
            },
            stdioDefaultEnabled: {
              command: 'qmd',
              args: ['mcp'],
            },
            stdioDisabled: {
              command: 'npx',
              args: ['-y', 'mcp-stdio-disabled'],
              enabled: false,
            },
            remoteDisabled: {
              url: 'https://example.com/mcp',
              headers: { REMOTE_API_KEY: '${REMOTE_API_KEY}' },
              enabled: false,
            },
            remoteEnabled: {
              url: 'https://example.com/remote-enabled',
              headers: { REMOTE_ENABLED_API_KEY: '${REMOTE_ENABLED_API_KEY}' },
              enabled: true,
            },
          },
        },
        null,
        2,
      ),
    )
  })

  afterEach(() => {
    rmSync(targetDir, { recursive: true, force: true })
    rmSync(libraryDir, { recursive: true, force: true })
  })

  it('scaffoldMcp creates .ai/mcp.json from catalog', async () => {
    await scaffoldMcp({
      targetDir,
      libraryDir,
      fileRecords,
      strategy: 'skip',
      perFileOverrides: new Map(),
    })

    const content = JSON.parse(readFile(path.join(targetDir, '.ai', 'mcp.json')))
    expect(content.servers.stdioEnabled.command).toBe('npx')
    expect(fileRecords.some((r) => r.path === '.ai/mcp.json')).toBe(true)
  })

  it('keeps requiresInstall servers disabled by default', async () => {
    writeFile(
      path.join(libraryDir, 'mcp', 'catalog.json'),
      JSON.stringify(
        {
          servers: {
            codegraph: {
              command: 'codegraph',
              args: ['serve', '--mcp'],
              requiresInstall: true,
              enabled: false,
            },
            qmd: {
              command: 'qmd',
              args: ['mcp'],
              requiresInstall: true,
              enabled: false,
            },
            memory: {
              command: 'npx',
              args: ['-y', '@modelcontextprotocol/server-memory'],
              requiresInstall: false,
              enabled: true,
            },
          },
        },
        null,
        2,
      ),
    )

    await scaffoldMcp({
      targetDir,
      libraryDir,
      fileRecords,
      strategy: 'skip',
      perFileOverrides: new Map(),
    })

    const catalog = JSON.parse(readFile(path.join(targetDir, '.ai', 'mcp.json')))
    expect(catalog.servers.codegraph.enabled).toBe(false)
    expect(catalog.servers.qmd.enabled).toBe(false)
    expect(catalog.servers.memory.enabled).toBe(true)
  })

  it('compileMcp generates .mcp.json and .opencode/mcp-servers.json for opencode', async () => {
    await scaffoldMcp({
      targetDir,
      libraryDir,
      fileRecords,
      strategy: 'skip',
      perFileOverrides: new Map(),
    })

    await compileMcp({
      canonicalDir: targetDir,
      toolTargetDir: targetDir,
      toolId: 'opencode',
      fileRecords,
    })

    const mcpJson = JSON.parse(readFile(path.join(targetDir, '.mcp.json')))
    expect(Object.keys(mcpJson.mcpServers)).toEqual(['stdioEnabled', 'stdioDefaultEnabled', 'remoteEnabled'])
    expect(mcpJson.mcpServers.remoteEnabled.url).toBe('https://example.com/remote-enabled')
    expect(mcpJson.mcpServers.remoteEnabled.headers.REMOTE_ENABLED_API_KEY).toBe('${REMOTE_ENABLED_API_KEY}')

    const opencodeMcp = JSON.parse(readFile(path.join(targetDir, '.opencode', 'mcp-servers.json')))
    expect(opencodeMcp.stdioEnabled.type).toBe('local')
    expect(opencodeMcp.stdioEnabled.environment.API_KEY).toBe('{env:API_KEY}')
    expect(opencodeMcp.stdioDisabled.enabled).toBe(false)
    expect(opencodeMcp.remoteDisabled.type).toBe('remote')
    expect(opencodeMcp.remoteDisabled.headers.REMOTE_API_KEY).toBe('{env:REMOTE_API_KEY}')
  })

  it('compileMcp generates .vscode/mcp.json for copilot with stdio and remote types', async () => {
    await scaffoldMcp({
      targetDir,
      libraryDir,
      fileRecords,
      strategy: 'skip',
      perFileOverrides: new Map(),
    })

    await compileMcp({
      canonicalDir: targetDir,
      toolTargetDir: targetDir,
      toolId: 'copilot',
      fileRecords,
    })

    const copilot = JSON.parse(readFile(path.join(targetDir, '.vscode', 'mcp.json')))
    expect(copilot.servers.stdioEnabled.type).toBe('stdio')
    expect(copilot.servers.remoteEnabled.type).toBe('sse')
    expect(copilot.servers.remoteEnabled.url).toBe('https://example.com/remote-enabled')
    expect(copilot.servers.remoteEnabled.headers.REMOTE_ENABLED_API_KEY).toBe('${REMOTE_ENABLED_API_KEY}')
    expect(copilot.servers.remoteDisabled).toBeUndefined()
    expect(copilot.servers.stdioDisabled).toBeUndefined()
  })

  it('compileMcp generates .gemini/settings.json with $VAR env syntax and warns on remote servers', async () => {
    const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => undefined)

    await scaffoldMcp({
      targetDir,
      libraryDir,
      fileRecords,
      strategy: 'skip',
      perFileOverrides: new Map(),
    })

    await compileMcp({
      canonicalDir: targetDir,
      toolTargetDir: targetDir,
      toolId: 'gemini',
      fileRecords,
    })

    const gemini = JSON.parse(readFile(path.join(targetDir, '.gemini', 'settings.json')))
    expect(gemini.mcpServers.stdioEnabled.env.API_KEY).toBe('$API_KEY')
    expect(gemini.mcpServers.stdioEnabled.includeTools).toBeUndefined()
    expect(warnSpy).toHaveBeenCalledWith('⚠️  Skipping remote server "remoteEnabled" for gemini (not supported)')
    warnSpy.mockRestore()
    expect(gemini.mcpServers.remoteDisabled).toBeUndefined()
    expect(gemini.mcpServers.remoteEnabled).toBeUndefined()
  })

  it('compileMcp generates .mcp.json for claude-code', async () => {
    await scaffoldMcp({
      targetDir,
      libraryDir,
      fileRecords,
      strategy: 'skip',
      perFileOverrides: new Map(),
    })

    await compileMcp({
      canonicalDir: targetDir,
      toolTargetDir: targetDir,
      toolId: 'claude-code',
      fileRecords,
    })

    const mcpJson = JSON.parse(readFile(path.join(targetDir, '.mcp.json')))
    expect(Object.keys(mcpJson.mcpServers)).toEqual(['stdioEnabled', 'stdioDefaultEnabled', 'remoteEnabled'])
    expect(mcpJson.mcpServers.remoteEnabled.url).toBe('https://example.com/remote-enabled')
  })
})
