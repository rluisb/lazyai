import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { beforeEach, describe, expect, it } from 'vitest'
import { runHealthChecks } from '../commands/server.js'
import { ensureDir, writeFile } from '../utils/files.js'

function makeTempDir(): string {
  return fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-server-doctor-'))
}

interface Catalog {
  servers: Record<
    string,
    {
      description?: string
      command?: string
      args?: string[]
      tools?: string[]
      enabled?: boolean
      requiresInstall?: boolean
      installHint?: string
      env?: Record<string, string>
    }
  >
}

function buildCatalog(): Catalog {
  return {
    servers: {
      orchestrator: {
        description: 'Test orchestrator',
        command: 'node',
        args: ['-e', 'process.exit(0)'],
        tools: ['list_catalog', 'compose_agent'],
        enabled: true,
      },
      memory: {
        description: 'Memory',
        command: 'npx',
        args: ['-y', '@modelcontextprotocol/server-memory'],
        tools: ['create_entities'],
        enabled: false,
      },
    },
  }
}

function seedCanonicalMcp(targetDir: string, catalog: Catalog): void {
  ensureDir(path.join(targetDir, '.ai'))
  writeFile(path.join(targetDir, '.ai', 'mcp.json'), JSON.stringify(catalog, null, 2))
}

function seedOpencodeConfig(targetDir: string, serverName: string): void {
  const content = {
    $schema: 'https://opencode.ai/config.json',
    mcp: {
      [serverName]: { type: 'local', command: ['node', '-e', 'process.exit(0)'] },
    },
  }
  writeFile(path.join(targetDir, 'opencode.jsonc'), JSON.stringify(content, null, 2))
}

describe('runHealthChecks', () => {
  let targetDir: string

  beforeEach(() => {
    targetDir = makeTempDir()
  })

  it('fails if .ai/mcp.json is missing', async () => {
    const report = await runHealthChecks(targetDir, 'orchestrator', buildCatalog(), [], 1000)
    expect(report.overall).toBe('unhealthy')
    expect(report.checks[0]?.status).toBe('fail')
    expect(report.checks[0]?.message).toContain('.ai/mcp.json is missing')
  })

  it('fails if the server is not in canonical mcp.json', async () => {
    seedCanonicalMcp(targetDir, { servers: {} })
    const report = await runHealthChecks(targetDir, 'orchestrator', buildCatalog(), [], 1000)
    expect(report.overall).toBe('unhealthy')
    const canonicalCheck = report.checks.find((c) => c.name === 'canonical mcp.json entry')
    expect(canonicalCheck?.status).toBe('fail')
    expect(canonicalCheck?.message).toContain("'orchestrator' missing from .ai/mcp.json")
  })

  it('fails if canonical entry is present but not enabled', async () => {
    const catalog = buildCatalog()
    catalog.servers.orchestrator!.enabled = false
    seedCanonicalMcp(targetDir, catalog)
    const report = await runHealthChecks(targetDir, 'orchestrator', buildCatalog(), [], 1000)
    expect(report.overall).toBe('unhealthy')
    const check = report.checks.find((c) => c.name === 'canonical mcp.json entry')
    expect(check?.status).toBe('fail')
    expect(check?.message).toContain('not enabled')
  })

  it('passes L1 config checks when all files are in place but fails stdio for a server that exits immediately', async () => {
    const catalog = buildCatalog()
    seedCanonicalMcp(targetDir, catalog)
    seedOpencodeConfig(targetDir, 'orchestrator')
    // orchestration dir required for L1 orchestrator-specific check
    ensureDir(path.join(targetDir, '.ai', 'orchestration', 'chains'))
    // write a fake compiled agent file so the L1 tool-agent check passes
    ensureDir(path.join(targetDir, '.opencode', 'agents'))
    writeFile(path.join(targetDir, '.opencode', 'agents', 'orchestrator.md'), '# test')

    const report = await runHealthChecks(targetDir, 'orchestrator', catalog, ['opencode'], 2000)
    const passing = report.checks.filter((c) => c.status === 'pass').map((c) => c.name)
    expect(passing).toContain('canonical mcp.json entry')
    expect(passing).toContain('opencode mcp config')
    expect(passing).toContain('orchestration chains')
    expect(passing).toContain('opencode orchestrator agent')

    // stdio handshake runs `node -e 'process.exit(0)'` which exits before speaking MCP
    // so the handshake fails — but that's expected for this fixture.
    const handshake = report.checks.find((c) => c.name === 'stdio handshake')
    expect(handshake?.status).toBe('fail')
  })

  it('skips stdio handshake for url-based (remote) servers', async () => {
    const catalog: Catalog = {
      servers: {
        context7: {
          description: 'Remote doc lookup',
          tools: ['query'],
          enabled: true,
        },
      },
    }
    seedCanonicalMcp(targetDir, catalog)
    const report = await runHealthChecks(targetDir, 'context7', catalog, [], 1000)
    const handshake = report.checks.find((c) => c.name === 'stdio handshake')
    expect(handshake?.status).toBe('skip')
    expect(handshake?.message).toContain('no stdio command')
  })

  it('reports corrupt .ai/mcp.json as fail', async () => {
    ensureDir(path.join(targetDir, '.ai'))
    writeFile(path.join(targetDir, '.ai', 'mcp.json'), '{ not valid json')
    const report = await runHealthChecks(targetDir, 'orchestrator', buildCatalog(), [], 1000)
    expect(report.overall).toBe('unhealthy')
    expect(report.checks.some((c) => c.message.includes('not valid JSON'))).toBe(true)
  })

  it('reports missing per-tool mcp config file', async () => {
    const catalog = buildCatalog()
    seedCanonicalMcp(targetDir, catalog)
    ensureDir(path.join(targetDir, '.ai', 'orchestration', 'chains'))
    // intentionally do NOT create opencode.jsonc
    const report = await runHealthChecks(targetDir, 'orchestrator', catalog, ['opencode'], 1000)
    const check = report.checks.find((c) => c.name === 'opencode mcp config')
    expect(check?.status).toBe('fail')
    expect(check?.message).toContain('is missing')
  })

  it('reports entry missing from per-tool mcp config', async () => {
    const catalog = buildCatalog()
    seedCanonicalMcp(targetDir, catalog)
    ensureDir(path.join(targetDir, '.ai', 'orchestration', 'chains'))
    // opencode config exists but has a different server
    writeFile(
      path.join(targetDir, 'opencode.jsonc'),
      JSON.stringify({ $schema: 'https://opencode.ai/config.json', mcp: { memory: {} } }),
    )
    const report = await runHealthChecks(targetDir, 'orchestrator', catalog, ['opencode'], 1000)
    const check = report.checks.find((c) => c.name === 'opencode mcp config')
    expect(check?.status).toBe('fail')
    expect(check?.message).toContain("does not contain 'orchestrator'")
  })

  it('skips per-tool mcp check for codex (no project-local config)', async () => {
    const catalog = buildCatalog()
    seedCanonicalMcp(targetDir, catalog)
    ensureDir(path.join(targetDir, '.ai', 'orchestration', 'chains'))
    const report = await runHealthChecks(targetDir, 'orchestrator', catalog, ['codex'], 1000)
    const check = report.checks.find((c) => c.name === 'codex mcp config')
    expect(check?.status).toBe('skip')
  })
})

describe('runHealthChecks stdio handshake (integration)', () => {
  const dogfoodDist = path.resolve(__dirname, '../../orchestrator/dist/index.js')
  const distExists = fs.existsSync(dogfoodDist)

  it.skipIf(!distExists)(
    'handshakes the real orchestrator dist and finds all 9 tools',
    async () => {
      const targetDir = makeTempDir()
      const catalog: Catalog = {
        servers: {
          orchestrator: {
            description: 'Real orchestrator via local dist',
            command: 'node',
            args: [dogfoodDist],
            tools: [
              'list_catalog',
              'compose_agent',
              'start_chain',
              'advance_chain',
              'get_status',
              'get_budget',
              'retry_step',
              'escalate_step',
              'handoff',
            ],
            enabled: true,
          },
        },
      }
      seedCanonicalMcp(targetDir, catalog)
      ensureDir(path.join(targetDir, '.ai', 'orchestration', 'chains'))

      const report = await runHealthChecks(targetDir, 'orchestrator', catalog, [], 10000)
      const handshake = report.checks.find((c) => c.name === 'stdio handshake')
      expect(handshake?.status).toBe('pass')
      expect(handshake?.message).toContain('9 tools')
    },
    15000,
  )
})
