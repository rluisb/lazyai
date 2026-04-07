import { mkdtempSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { compileMcp } from '../adapters/mcp-compiler.js'
import { scaffoldEnvExample } from '../scaffold/env-example.js'
import { scaffoldMcp } from '../scaffold/mcp.js'
import type { FileRecord } from '../types.js'
import { ensureDir, fileExists, readFile, writeFile } from '../utils/files.js'

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
              // biome-ignore lint/suspicious/noTemplateCurlyInString: intentional catalog template syntax
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
              // biome-ignore lint/suspicious/noTemplateCurlyInString: intentional catalog template syntax
              headers: { REMOTE_API_KEY: '${REMOTE_API_KEY}' },
              enabled: false,
            },
            remoteEnabled: {
              url: 'https://example.com/remote-enabled',
              // biome-ignore lint/suspicious/noTemplateCurlyInString: intentional catalog template syntax
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

  it('compileMcp generates opencode.jsonc with mcp config for opencode', async () => {
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

    const opencodeConfig = JSON.parse(readFile(path.join(targetDir, 'opencode.jsonc')))
    expect(opencodeConfig.$schema).toBe('https://opencode.ai/config.json')
    expect(opencodeConfig.mcp.stdioEnabled.type).toBe('local')
    expect(opencodeConfig.mcp.stdioEnabled.environment.API_KEY).toBe('{env:API_KEY}')
    expect(opencodeConfig.mcp.stdioDisabled.enabled).toBe(false)
    expect(opencodeConfig.mcp.remoteDisabled.type).toBe('remote')
    expect(opencodeConfig.mcp.remoteDisabled.headers.REMOTE_API_KEY).toBe('{env:REMOTE_API_KEY}')
  })

  it('compileMcp merges existing opencode.jsonc and preserves non-mcp keys', async () => {
    await scaffoldMcp({
      targetDir,
      libraryDir,
      fileRecords,
      strategy: 'skip',
      perFileOverrides: new Map(),
    })

    writeFile(
      path.join(targetDir, 'opencode.jsonc'),
      `${JSON.stringify(
        {
          plugin: ['foo-plugin'],
          permission: { default: 'allow' },
          mcp: {
            legacy: {
              type: 'local',
              command: ['legacy-cmd'],
            },
          },
        },
        null,
        2,
      )}\n`
    )

    await compileMcp({
      canonicalDir: targetDir,
      toolTargetDir: targetDir,
      toolId: 'opencode',
      fileRecords,
    })

    const opencodeConfig = JSON.parse(readFile(path.join(targetDir, 'opencode.jsonc')))
    expect(opencodeConfig.plugin).toEqual(['foo-plugin'])
    expect(opencodeConfig.permission).toEqual({ default: 'allow' })
    expect(opencodeConfig.$schema).toBe('https://opencode.ai/config.json')
    expect(opencodeConfig.mcp.legacy).toBeUndefined()
    expect(opencodeConfig.mcp.stdioEnabled.type).toBe('local')
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
    // biome-ignore lint/suspicious/noTemplateCurlyInString: intentional placeholder assertion
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

  it('enableServers option enables disabled MCP servers by name', async () => {
    await scaffoldMcp({
      targetDir,
      libraryDir,
      fileRecords,
      strategy: 'skip',
      perFileOverrides: new Map(),
      enableServers: ['stdioDisabled'],
    })

    const catalog = JSON.parse(readFile(path.join(targetDir, '.ai', 'mcp.json')))
    expect(catalog.servers.stdioDisabled.enabled).toBe(true)
    expect(catalog.servers.stdioEnabled.enabled).toBe(true)
  })

  it('enableServers ignores unknown server names', async () => {
    await scaffoldMcp({
      targetDir,
      libraryDir,
      fileRecords,
      strategy: 'skip',
      perFileOverrides: new Map(),
      enableServers: ['nonexistent'],
    })

    const catalog = JSON.parse(readFile(path.join(targetDir, '.ai', 'mcp.json')))
    expect(catalog.servers.nonexistent).toBeUndefined()
  })

  describe('env-example generation', () => {
    it('generates .env.example from enabled servers with env vars', async () => {
      await scaffoldMcp({
        targetDir,
        libraryDir,
        fileRecords,
        strategy: 'skip',
        perFileOverrides: new Map(),
      })

      await scaffoldEnvExample({
        targetDir,
        fileRecords,
        strategy: 'skip',
        perFileOverrides: new Map(),
      })

      const envExample = readFile(path.join(targetDir, '.env.example'))
      expect(envExample).toContain('API_KEY=')
      expect(envExample).toContain('Required by: stdioEnabled')
      expect(envExample).toContain('NEVER commit .env')
      expect(fileRecords.some((r) => r.path === '.env.example')).toBe(true)
    })

    it('does not generate .env.example when no enabled servers have env vars', async () => {
      // Create a catalog with no env vars on enabled servers
      writeFile(
        path.join(libraryDir, 'mcp', 'catalog.json'),
        JSON.stringify(
          {
            servers: {
              noenv: {
                command: 'npx',
                args: ['-y', 'no-env-server'],
                enabled: true,
              },
            },
          },
          null,
          2,
        ),
      )

      const records: FileRecord[] = []
      await scaffoldMcp({
        targetDir,
        libraryDir,
        fileRecords: records,
        strategy: 'skip',
        perFileOverrides: new Map(),
      })

      await scaffoldEnvExample({
        targetDir,
        fileRecords: records,
        strategy: 'skip',
        perFileOverrides: new Map(),
      })

      expect(fileExists(path.join(targetDir, '.env.example'))).toBe(false)
    })
  })
})
