import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { Writable } from 'node:stream'
import { afterEach, describe, expect, it } from 'vitest'
import { openDatabase, type Db } from '../db/index.js'
import { runMigrations } from '../db/migrations.js'
import { INIT_HELP, parseInitArgs, runInit } from '../cli/init.js'

const tempDirs: string[] = []

afterEach(() => {
  for (const dir of tempDirs.splice(0)) {
    fs.rmSync(dir, { recursive: true, force: true })
  }
})

function createBuffer(): { out: NodeJS.WritableStream; output: () => string } {
  const chunks: string[] = []
  return {
    out: new Writable({
      write(chunk, _encoding, callback) {
        chunks.push(Buffer.isBuffer(chunk) ? chunk.toString('utf8') : String(chunk))
        callback()
      },
    }),
    output: () => chunks.join(''),
  }
}

function setupFixture(): { projectRoot: string; libraryRoot: string; agentsRoot: string; db: Db } {
  const projectRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'orchestrator-init-project-'))
  const libraryRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'orchestrator-init-library-'))
  const agentsRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'orchestrator-init-agents-'))
  tempDirs.push(projectRoot, libraryRoot, agentsRoot)

  fs.mkdirSync(path.join(libraryRoot, 'chains'), { recursive: true })
  fs.mkdirSync(path.join(libraryRoot, 'teams'), { recursive: true })
  fs.mkdirSync(path.join(libraryRoot, 'workflows'), { recursive: true })
  fs.mkdirSync(path.join(libraryRoot, 'skills', 'domains'), { recursive: true })
  fs.mkdirSync(path.join(libraryRoot, 'skills', 'modes'), { recursive: true })
  fs.mkdirSync(agentsRoot, { recursive: true })

  fs.writeFileSync(path.join(projectRoot, 'AGENTS.md'), '# Agent rules\n')
  fs.writeFileSync(path.join(projectRoot, 'CLAUDE.md'), '# Claude rules\n')

  for (const agent of ['architect', 'implementor-senior', 'reviewer']) {
    fs.writeFileSync(
      path.join(agentsRoot, `${agent}.md`),
      ['---', `name: ${agent}`, '---', '', `# ${agent}`, '', 'Use existing project conventions.'].join('\n'),
    )
  }

  fs.writeFileSync(
    path.join(libraryRoot, 'skills', 'domains', 'typescript.md'),
    ['---', 'name: TypeScript', 'description: TypeScript domain', '---', '', 'Prefer exact types.'].join('\n'),
  )
  fs.writeFileSync(
    path.join(libraryRoot, 'skills', 'modes', 'tdd.md'),
    ['---', 'name: TDD', 'description: Test first', '---', '', 'Use red green refactor.'].join('\n'),
  )

  fs.writeFileSync(
    path.join(libraryRoot, 'chains', 'feature-rpi.json'),
    JSON.stringify({
      name: 'feature-rpi',
      kind: 'chain',
      description: 'Research, plan, implement',
      entry: 'plan',
      steps: [
        { id: 'plan', agent: 'architect', skills: [], description: 'Plan', transitions: { success: 'implement' } },
        { id: 'implement', agent: 'implementor-senior', skills: [], description: 'Implement', transitions: { success: 'done' } },
      ],
    }),
  )
  fs.writeFileSync(
    path.join(libraryRoot, 'teams', 'review-team.json'),
    JSON.stringify({
      name: 'review-team',
      kind: 'team',
      description: 'Parallel review team',
      parallel: [{ role: 'reviewer', agent: 'reviewer', skills: [], focus: 'security and correctness' }],
      synthesize: { agent: 'reviewer', description: 'Summarize review findings' },
    }),
  )
  fs.writeFileSync(
    path.join(libraryRoot, 'workflows', 'rpi.json'),
    JSON.stringify({
      name: 'rpi',
      kind: 'workflow',
      description: 'RPI delivery workflow',
      entry: 'delivery',
      phases: [{ id: 'delivery', kind: 'chain', ref: 'feature-rpi', on: { success: 'done' } }],
    }),
  )

  const db = openDatabase(':memory:')
  runMigrations(db)
  return { projectRoot, libraryRoot, agentsRoot, db }
}

describe('init cli', () => {
  it('parseInitArgs handles flags', () => {
    expect(parseInitArgs([
      '--task', 'review auth',
      '--host', 'claude-code',
      '--project', '/tmp/project',
      '--json',
      '--verbose',
    ])).toEqual({
      task: 'review auth',
      host: 'claude-code',
      projectRoot: '/tmp/project',
      json: true,
      verbose: true,
    })
  })

  it('parseInitArgs rejects removed host cli tools', () => {
    expect(() => parseInitArgs(['--host', 'codex'])).toThrow('Unknown host: codex')
    expect(() => parseInitArgs(['--host', 'gemini'])).toThrow('Unknown host: gemini')
  })

  it('runInit --json emits valid JSON with inventory and root files', async () => {
    const fixture = setupFixture()
    const { out, output } = createBuffer()

    await runInit({
      projectRoot: fixture.projectRoot,
      host: 'opencode',
      json: true,
      libraryOrchestrationRoot: fixture.libraryRoot,
      libraryAgentsRoot: fixture.agentsRoot,
      db: fixture.db,
    }, out)

    const parsed = JSON.parse(output())
    expect(parsed.projectRoot).toBe(fixture.projectRoot)
    expect(parsed.host.host).toBe('opencode')
    expect(parsed.rootFiles).toEqual({ agentsMd: true, claudeMd: true })
    expect(parsed.inventory.agents).toEqual(expect.arrayContaining(['architect', 'implementor-senior', 'reviewer']))
    expect(parsed.inventory.chains).toEqual(['feature-rpi'])
    expect(parsed.inventory.teams).toEqual(['review-team'])
    expect(parsed.inventory.workflows).toEqual(['rpi'])
    expect(parsed.recommendation).toBeUndefined()
  })

  it('Claude Code review task can recommend a team when a team exists', async () => {
    const fixture = setupFixture()
    const { out, output } = createBuffer()

    await runInit({
      projectRoot: fixture.projectRoot,
      host: 'claude-code',
      task: 'security review auth middleware',
      json: true,
      libraryOrchestrationRoot: fixture.libraryRoot,
      libraryAgentsRoot: fixture.agentsRoot,
      db: fixture.db,
    }, out)

    const parsed = JSON.parse(output())
    expect(parsed.recommendation).toMatchObject({ kind: 'team', name: 'review-team', confidence: 'high' })
  })

  it('OpenCode build/from-scratch task recommends workflow or chain instead of team', async () => {
    const fixture = setupFixture()
    const { out, output } = createBuffer()

    await runInit({
      projectRoot: fixture.projectRoot,
      host: 'opencode',
      task: 'build auth from scratch',
      json: true,
      libraryOrchestrationRoot: fixture.libraryRoot,
      libraryAgentsRoot: fixture.agentsRoot,
      db: fixture.db,
    }, out)

    const parsed = JSON.parse(output())
    expect(['workflow', 'chain']).toContain(parsed.recommendation.kind)
    expect(parsed.recommendation.kind).not.toBe('team')
  })

  it('help output includes usage', async () => {
    const { out, output } = createBuffer()

    await runInit({ help: true }, out)

    expect(output()).toContain('Usage: ai-setup-orchestrator init [options]')
    expect(INIT_HELP).toContain('--task <text>')
  })
})
