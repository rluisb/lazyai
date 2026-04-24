import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { OrchestratorToolHandlers } from '../tool-handlers.js'
import { resetEventBus, getEventBus } from '../events/bus.js'
import type { RunEvent } from '../events/bus.js'
import { initPersistenceDb, closePersistenceDb, loadChainState, getPersistenceDb } from '../persistence.js'
import { CatalogStore } from '../catalog/store.js'

const tempDirs: string[] = []

beforeEach(() => {
  initPersistenceDb(':memory:')
})

afterEach(() => {
  closePersistenceDb()
  resetEventBus()
  for (const dir of tempDirs.splice(0)) {
    fs.rmSync(dir, { recursive: true, force: true })
  }
})

function setupFixture() {
  const projectRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'orchestrator-invoke-'))
  const agentsRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'orchestrator-invoke-agents-'))
  const libraryRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'orchestrator-invoke-library-'))
  tempDirs.push(projectRoot, agentsRoot, libraryRoot)

  fs.mkdirSync(path.join(libraryRoot, 'skills', 'domains'), { recursive: true })

  fs.writeFileSync(
    path.join(agentsRoot, 'analyst.md'),
    ['---', 'name: Analyst', 'model: sonnet', '---', '', '# Analyst', '', 'Analyse and report.'].join('\n'),
  )
  fs.writeFileSync(
    path.join(libraryRoot, 'skills', 'domains', 'security.md'),
    ['---', 'name: Security', 'description: security domain', '---', '', 'Apply security checks.'].join('\n'),
  )

  return {
    projectRoot,
    handlers: new OrchestratorToolHandlers({
      projectRoot,
      libraryOrchestrationRoot: libraryRoot,
      libraryAgentsRoot: agentsRoot,
    }),
  }
}

describe('invokeAgent', () => {
  it('resolves agent and returns composed spec', () => {
    const { handlers } = setupFixture()
    const result = handlers.invokeAgent({ agent: 'analyst', task: 'Review auth module' })

    expect(result.agentName).toBe('analyst')
    expect(result.invocationId).toBeTruthy()
    expect(result.composed.prompt).toContain('Review auth module')
  })

  it('creates a real chain run that can be loaded', () => {
    const { handlers, projectRoot } = setupFixture()
    const result = handlers.invokeAgent({ agent: 'analyst', task: 'Scan codebase' })

    const chainState = loadChainState(projectRoot, result.invocationId)
    expect(chainState.chainId).toBe(result.invocationId)
    expect(chainState.state).toBe('running')
    expect(chainState.definitionName).toBe('agent:analyst')
  })

  it('applies domain skill when provided', () => {
    const { handlers } = setupFixture()
    const result = handlers.invokeAgent({
      agent: 'analyst',
      task: 'Check for vulnerabilities',
      domainSkill: 'security',
    })

    expect(result.composed.domainSkill).toBe('security')
    expect(result.composed.prompt).toContain('Apply security checks')
  })

  it('emits agent.invoked event', () => {
    const { handlers } = setupFixture()
    const events: RunEvent[] = []
    getEventBus().onAny((e) => events.push(e))

    const result = handlers.invokeAgent({ agent: 'analyst', task: 'Scan dependencies' })

    expect(events).toHaveLength(1)
    expect(events[0]?.eventType).toBe('agent.invoked')
    expect(events[0]?.runId).toBe(result.invocationId)
    expect(events[0]?.payload).toMatchObject({ agentName: 'analyst', task: 'Scan dependencies' })
  })

  it('throws for unknown agent', () => {
    const { handlers } = setupFixture()
    expect(() => handlers.invokeAgent({ agent: 'nonexistent', task: 'anything' }))
      .toThrow(/Unknown base agent/)
  })

  it('uses pinned DB version body when version is specified', () => {
    const store = new CatalogStore(getPersistenceDb())
    store.createVersion({
      kind: 'agent',
      name: 'analyst',
      frontmatter: { name: 'Analyst' },
      body: '# Analyst v2\n\nVersion 2 prompt.',
    })

    const { handlers } = setupFixture()
    const result = handlers.invokeAgent({ agent: 'analyst', task: 'Run checks', version: 1 })

    expect(result.resolvedVersion).toBe(1)
    expect(result.composed.prompt).toContain('Version 2 prompt.')
  })
})

// ---------------------------------------------------------------------------
// Phase 7: chain_retry queue handler
// ---------------------------------------------------------------------------
import { openDatabase } from '../db/index.js'
import { runMigrations } from '../db/migrations.js'
import { JobQueue } from '../queue/queue.js'
import { QueueWorker } from '../queue/worker.js'
import { registerBuiltinHandlers } from '../queue/handlers.js'

describe('chain_retry queue handler', () => {
  it('resets a failed step to running and emits chain.retry_ready', async () => {
    const db = openDatabase(':memory:')
    runMigrations(db)
    initPersistenceDb(':memory:')

    const projectRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'orchestrator-retry-'))
    const agentsRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'orchestrator-retry-agents-'))
    const libraryRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'orchestrator-retry-library-'))
    tempDirs.push(projectRoot, agentsRoot, libraryRoot)

    fs.mkdirSync(path.join(libraryRoot, 'chains'), { recursive: true })
    fs.mkdirSync(agentsRoot, { recursive: true })

    fs.writeFileSync(
      path.join(agentsRoot, 'builder.md'),
      ['---', 'name: Builder', 'model: sonnet', '---', '', '# Builder', '', 'Build things.'].join('\n'),
    )
    fs.writeFileSync(
      path.join(libraryRoot, 'chains', 'fix.json'),
      JSON.stringify({
        name: 'fix',
        kind: 'chain',
        entry: 'fix-step',
        description: 'Fix chain',
        steps: [{
          id: 'fix-step',
          agent: 'builder',
          skills: [],
          description: 'Apply fix',
          prompt: 'Fix it',
          transitions: { success: 'done', failure: { retry: 2, then: 'done' } },
        }],
      }),
    )

    const handlers = new OrchestratorToolHandlers({ projectRoot, libraryOrchestrationRoot: libraryRoot, libraryAgentsRoot: agentsRoot })
    const chain = handlers.startChain({ chain: 'fix', task: 'Patch the auth bug' })

    // Simulate a step failure with retries remaining
    const events: RunEvent[] = []
    getEventBus().onAny((e) => events.push(e))

    const advance = handlers.advanceChain({
      chainId: chain.chainId,
      stepId: 'fix-step',
      outcome: 'failure',
    })
    expect(advance.recovery?.type).toBe('retry')

    // The retry should have been enqueued in the persistence DB
    const q = new JobQueue(getPersistenceDb())
    expect(q.pendingCount('chain_retry')).toBe(1)

    // Run the worker to process it
    const worker = new QueueWorker({ db: getPersistenceDb(), pollIntervalMs: 50 })
    registerBuiltinHandlers(worker)
    worker.start()

    await vi.waitFor(() => {
      const retryEvents = events.filter((e) => e.eventType === 'chain.retry_ready')
      expect(retryEvents).toHaveLength(1)
    }, { timeout: 2000 })
    worker.stop()

    const retryReady = events.find((e) => e.eventType === 'chain.retry_ready')!
    expect(retryReady.runId).toBe(chain.chainId)
    expect(retryReady.payload.stepId).toBe('fix-step')

    // Chain state should now be 'running' again
    const state = loadChainState(projectRoot, chain.chainId)
    expect(state.state).toBe('running')

    closePersistenceDb()
  })
})
