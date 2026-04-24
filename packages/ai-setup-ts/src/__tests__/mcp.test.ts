import { mkdtempSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
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

  it('compileMcp generates .opencode/opencode.jsonc with mcp config for opencode', async () => {
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

    const opencodeConfig = JSON.parse(readFile(path.join(targetDir, '.opencode', 'opencode.jsonc')))
    expect(opencodeConfig.$schema).toBe('https://opencode.ai/config.json')
    expect(opencodeConfig.mcp.stdioEnabled.type).toBe('local')
    expect(opencodeConfig.mcp.stdioEnabled.environment.API_KEY).toBe('{env:API_KEY}')
    expect(opencodeConfig.mcp.stdioDisabled.enabled).toBe(false)
    expect(opencodeConfig.mcp.remoteDisabled.type).toBe('remote')
    expect(opencodeConfig.mcp.remoteDisabled.headers.REMOTE_API_KEY).toBe('{env:REMOTE_API_KEY}')
  })

  it('compileMcp preserves user-authored MCP servers and non-mcp keys on re-run', async () => {
    await scaffoldMcp({
      targetDir,
      libraryDir,
      fileRecords,
      strategy: 'skip',
      perFileOverrides: new Map(),
    })

    ensureDir(path.join(targetDir, '.opencode'))
    writeFile(
      path.join(targetDir, '.opencode', 'opencode.jsonc'),
      `${JSON.stringify(
        {
          plugin: ['foo-plugin'],
          permission: { default: 'allow' },
          mcp: {
            userOwned: {
              type: 'local',
              command: ['user-cmd'],
            },
          },
        },
        null,
        2,
      )}\n`,
    )

    await compileMcp({
      canonicalDir: targetDir,
      toolTargetDir: targetDir,
      toolId: 'opencode',
      fileRecords,
    })

    const opencodeConfig = JSON.parse(readFile(path.join(targetDir, '.opencode', 'opencode.jsonc')))
    expect(opencodeConfig.plugin).toEqual(['foo-plugin'])
    expect(opencodeConfig.permission).toEqual({ default: 'allow' })
    expect(opencodeConfig.$schema).toBe('https://opencode.ai/config.json')
    expect(opencodeConfig.mcp.userOwned).toEqual({ type: 'local', command: ['user-cmd'] })
    expect(opencodeConfig.mcp.stdioEnabled.type).toBe('local')
  })

  it('compileMcp managed servers win on key collision with user-authored entries', async () => {
    await scaffoldMcp({
      targetDir,
      libraryDir,
      fileRecords,
      strategy: 'skip',
      perFileOverrides: new Map(),
    })

    ensureDir(path.join(targetDir, '.opencode'))
    writeFile(
      path.join(targetDir, '.opencode', 'opencode.jsonc'),
      `${JSON.stringify(
        {
          mcp: {
            stdioEnabled: { type: 'local', command: ['user-override'] },
          },
        },
        null,
        2,
      )}\n`,
    )

    await compileMcp({
      canonicalDir: targetDir,
      toolTargetDir: targetDir,
      toolId: 'opencode',
      fileRecords,
    })

    const opencodeConfig = JSON.parse(readFile(path.join(targetDir, '.opencode', 'opencode.jsonc')))
    expect(opencodeConfig.mcp.stdioEnabled.command).toEqual(['npx', '-y', 'mcp-stdio-enabled'])
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
