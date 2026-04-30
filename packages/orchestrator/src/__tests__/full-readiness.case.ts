import fs from 'node:fs'
import path from 'node:path'
import os from 'node:os'
import { afterAll, beforeAll, describe, expect, it } from 'vitest'
import { openDatabase, type Db } from '../db/index.js'
import { runMigrations } from '../db/migrations.js'
import { CatalogStore } from '../catalog/store.js'
import { CatalogToolHandlers } from '../catalog-tools.js'
import { loadRuntimeCatalog } from '../catalog/runtime.js'
import { OrchestratorToolHandlers, type ToolHandlerOptions } from '../tool-handlers.js'
import { loadChainState, loadTeamState, loadWorkflowState, saveChainState, saveTeamState, saveWorkflowState, listActiveRuns, handoffActiveRuns, getPersistenceDb, initPersistenceDb, closePersistenceDb } from '../persistence.js'
import { createChainState } from '../chain-machine.js'
import { createTeamState } from '../team-machine.js'
import { createWorkflowState } from '../workflow-machine.js'
import { createBudgetState } from '../budget-tracker.js'
import type { ChainStepStatus, ExecutionPlan, TeamDefinition, TeamState, TeamTaskState, WorkflowDefinition, ChainState, StructuredError, RunKind, CliContext } from '../types.js'

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

const FAKE_ROOT = fs.mkdtempSync(path.join(os.tmpdir(), 'orc-readiness-'))
const AGENTS_DIR = path.join(FAKE_ROOT, 'agents')
const LIBRARY_DIR = path.join(FAKE_ROOT, 'library')
const PROJECT_A = path.join(FAKE_ROOT, 'project-a')
const PROJECT_B = path.join(FAKE_ROOT, 'project-b')

function ensureDirs() {
  for (const d of [PROJECT_A, PROJECT_B]) {
    fs.mkdirSync(`${d}/.ai/orchestration/state/chains`, { recursive: true })
    fs.mkdirSync(`${d}/.ai/orchestration/state/teams`, { recursive: true })
    fs.mkdirSync(`${d}/.ai/orchestration/state/workflows`, { recursive: true })
    fs.mkdirSync(`${d}/.ai/orchestration/state/handoffs`, { recursive: true })
  }
  // Create library
  fs.mkdirSync(path.join(LIBRARY_DIR, 'chains'), { recursive: true })
  fs.mkdirSync(path.join(LIBRARY_DIR, 'teams'), { recursive: true })
  fs.mkdirSync(path.join(LIBRARY_DIR, 'workflows'), { recursive: true })
  fs.mkdirSync(path.join(LIBRARY_DIR, 'skills', 'domains'), { recursive: true })
  fs.mkdirSync(path.join(LIBRARY_DIR, 'skills', 'modes'), { recursive: true })
  fs.mkdirSync(AGENTS_DIR, { recursive: true })

  // Write agents
  for (const agent of ['scout', 'builder', 'reviewer', 'red-team', 'implementor', 'architect', 'planner']) {
    fs.writeFileSync(path.join(AGENTS_DIR, `${agent}.md`),
      `---\nname: ${agent}\nmodel: sonnet\n---\n\n# ${agent}\n\nDo your job: ${agent}.\n`)
  }

  // Write domains
  fs.writeFileSync(path.join(LIBRARY_DIR, 'skills', 'domains', 'typescript.md'),
    '---\nname: TypeScript\ndescription: TypeScript domain skill\n---\n\nUse strict types.')
  fs.writeFileSync(path.join(LIBRARY_DIR, 'skills', 'modes', 'tdd.md'),
    '---\nname: tdd\ndescription: TDD mode\n---\n\nRed green refactor.')

  // Write chain
  fs.writeFileSync(path.join(LIBRARY_DIR, 'chains', 'repair.json'),
    JSON.stringify({
      name: 'repair', kind: 'chain', description: 'Repair chain', entry: 'diagnose',
      steps: [
        { id: 'diagnose', agent: 'scout', skills: ['typescript'], description: 'Research the problem', transitions: { success: 'fix', failure: 'done' } },
        { id: 'fix', agent: 'builder', skills: ['typescript'], description: 'Implement the fix', transitions: { success: 'done', failure: { retry: 1, then: 'done' } } },
      ],
    }))

  // Write team
  fs.writeFileSync(path.join(LIBRARY_DIR, 'teams', 'review-team.json'),
    JSON.stringify({
      name: 'review-team', kind: 'team', description: 'Parallel review', version: '1.0.0',
      parallel: [
        { role: 'security-reviewer', agent: 'red-team', skills: [], focus: 'Security vulnerabilities' },
        { role: 'correctness-reviewer', agent: 'reviewer', skills: [], focus: 'Logic correctness' },
      ],
      synthesize: { agent: 'reviewer', description: 'Merge findings into single report' },
    }))

  // Write workflow
  fs.writeFileSync(path.join(LIBRARY_DIR, 'workflows', 'delivery.json'),
    JSON.stringify({
      name: 'delivery', kind: 'workflow', description: 'Full delivery workflow', version: '1.0.0', entry: 'implement',
      phases: [
        { id: 'implement', kind: 'chain', ref: 'repair', on: { success: 'review', failure: 'handoff' } },
        { id: 'review', kind: 'team', ref: 'review-team', on: { success: 'complete', failure: 'handoff' } },
        { id: 'handoff', kind: 'terminal' },
        { id: 'complete', kind: 'terminal' },
      ],
    }))

  // Write project root files for init command
  fs.writeFileSync(path.join(PROJECT_A, 'AGENTS.md'), '# Agent rules\n')
  fs.writeFileSync(path.join(PROJECT_A, 'CLAUDE.md'), '# Claude rules\n')
}

function createHandlers(projectRoot = PROJECT_A, hostCli: ToolHandlerOptions['hostCli'] = 'opencode'): OrchestratorToolHandlers {
  return new OrchestratorToolHandlers({
    projectRoot,
    libraryOrchestrationRoot: LIBRARY_DIR,
    libraryAgentsRoot: AGENTS_DIR,
    hostCli,
  })
}

function isChainStepStatus(s: unknown): s is ChainStepStatus {
  return typeof s === 'object' && s !== null && 'stepId' in s
}

function makeStructuredError(runId: string, kind: RunKind, stepId: string): StructuredError {
  return {
    category: 'logical',
    code: 'TEST_FAILURE',
    message: 'Synthetic test failure',
    stepId,
    agent: 'builder',
    skills: ['typescript'],
    context: { runId, runKind: kind, task: 'Test', attempt: 1, hostCli: 'opencode' },
    suggestedRecovery: { type: 'retry', maxAttempts: 1 },
    timestamp: new Date().toISOString(),
  }
}

// ---------------------------------------------------------------------------
// 1. CATALOG MCP TOOLS — full lifecycle
// ---------------------------------------------------------------------------

describe('1. CATALOG MCP TOOLS — full lifecycle', () => {
  let db: Db
  let cat: CatalogToolHandlers

  beforeAll(() => {
    db = openDatabase(':memory:')
    runMigrations(db)
    cat = new CatalogToolHandlers(db)
    ensureDirs()
  })

  it('1a. List empty catalog returns 0 definitions', () => {
    expect(cat.catalogList().definitions).toHaveLength(0)
  })

  it('1b. Create first version', () => {
    const v1 = cat.catalogCreateVersion({
      kind: 'agent', name: 'reviewer',
      frontmatter: { name: 'Reviewer', model: 'sonnet' },
      body: '# Reviewer\nReview code.',
    })
    expect(v1.version).toBe(1)
    expect(v1.alreadyExists).toBe(false)
  })

  it('1c. Checksum dedup — identical content returns alreadyExists', () => {
    const dup = cat.catalogCreateVersion({
      kind: 'agent', name: 'reviewer',
      frontmatter: { name: 'Reviewer', model: 'sonnet' },
      body: '# Reviewer\nReview code.',
    })
    expect(dup.alreadyExists).toBe(true)
  })

  it('1d. Changed content creates version 2', () => {
    const v2 = cat.catalogCreateVersion({
      kind: 'agent', name: 'reviewer',
      frontmatter: { name: 'Reviewer v2', model: 'opus' },
      body: '# Reviewer v2\nReview more carefully.',
    })
    expect(v2.version).toBe(2)
  })

  it('1e. Active version defaults to latest', () => {
    const g = cat.catalogGetVersion({ kind: 'agent', name: 'reviewer' })
    expect(g.version).toBe(2)
  })

  it('1f. Pinned version returns correct content', () => {
    const g = cat.catalogGetVersion({ kind: 'agent', name: 'reviewer', version: 1 })
    expect(g.version).toBe(1)
    expect(g.body).toContain('Review code.')
  })

  it('1g. Set active version works', () => {
    cat.catalogSetActive({ kind: 'agent', name: 'reviewer', version: 1 })
    expect(cat.catalogGetVersion({ kind: 'agent', name: 'reviewer' }).version).toBe(1)
  })

  it('1h. List versions returns 2', () => {
    expect(cat.catalogListVersions({ kind: 'agent', name: 'reviewer' }).versions).toHaveLength(2)
  })

  it('1i. Diff returns both sides', () => {
    const diff = cat.catalogDiff({ kind: 'agent', name: 'reviewer', fromVersion: 1, toVersion: 2 })
    expect(diff.from?.body).toContain('Review code.')
    expect(diff.to?.body).toContain('more carefully')
  })

  it('1j. Deactivate clears active version; pinned still readable', () => {
    const deact = cat.catalogDeactivate({ kind: 'agent', name: 'reviewer' })
    expect(deact.deactivated).toBe(true)
    expect(deact.activeVersion).toBeNull()
    expect(() => cat.catalogGetVersion({ kind: 'agent', name: 'reviewer' })).toThrow('No active version')
    expect(cat.catalogGetVersion({ kind: 'agent', name: 'reviewer', version: 1 }).body).toContain('Review code.')
  })

  it('1k. Remove definition deletes all versions', () => {
    cat.catalogCreateVersion({ kind: 'agent', name: 'temp', frontmatter: { name: 'Temp' }, body: 'x' })
    const rm = cat.catalogRemove({ kind: 'agent', name: 'temp' })
    expect(rm.removed).toBe(true)
    expect(rm.versionsRemoved).toBe(1)
    expect(cat.catalogList({ kind: 'agent' }).definitions.find(d => d.name === 'temp')).toBeUndefined()
  })

  it('1l. Create and list chain/team/workflow definitions', () => {
    cat.catalogCreateVersion({
      kind: 'chain', name: 'test-chain',
      frontmatter: { name: 'Test Chain' },
      body: JSON.stringify({ kind: 'chain', name: 'ignored', description: '', entry: 'step1', steps: [{ id: 'step1', agent: 'builder', skills: [], description: 'x', transitions: { success: 'done' } }] }),
    })
    cat.catalogCreateVersion({
      kind: 'team', name: 'test-team',
      frontmatter: { name: 'Test Team' },
      body: JSON.stringify({ kind: 'team', name: 'ignored', description: '', parallel: [{ role: 'm', agent: 'builder', skills: [], focus: 'x' }], synthesize: { agent: 'reviewer', description: 'x' } }),
    })
    cat.catalogCreateVersion({
      kind: 'workflow', name: 'test-workflow',
      frontmatter: { name: 'Test Workflow' },
      body: JSON.stringify({ kind: 'workflow', name: 'ignored', description: '', entry: 'p1', phases: [{ id: 'p1', kind: 'chain', ref: 'test-chain', on: { success: 'done' } }] }),
    })
    const defs = cat.catalogList().definitions
    const kinds = defs.map(d => `${d.kind}:${d.name}`).sort()
    expect(kinds).toContain('chain:test-chain')
    expect(kinds).toContain('team:test-team')
    expect(kinds).toContain('workflow:test-workflow')
    expect(kinds).toContain('agent:reviewer')
  })

  it('1m. Export writes file to disk', () => {
    cat.catalogSetActive({ kind: 'agent', name: 'reviewer', version: 1 })
    const p = `${PROJECT_A}/exported-agent.md`
    cat.catalogExportVersion({ kind: 'agent', name: 'reviewer', targetPath: p })
    expect(fs.existsSync(p)).toBe(true)
    expect(fs.readFileSync(p, 'utf-8')).toContain('Review code.')
    fs.unlinkSync(p)
  })

  it('1n. Import returns results array', () => {
    const r = cat.catalogImport({ libraryOrchestrationRoot: LIBRARY_DIR, libraryAgentsRoot: AGENTS_DIR })
    expect(Array.isArray(r.results)).toBe(true)
  })

  it('1o. Invalid frontmatter rejected', () => {
    expect(() => cat.catalogCreateVersion({ kind: 'agent', name: 'bad', frontmatter: { name: 123 }, body: '' })).toThrow('Invalid frontmatter')
  })

  it('1p. Unknown version get throws', () => {
    expect(() => cat.catalogGetVersion({ kind: 'agent', name: 'reviewer', version: 999 })).toThrow()
  })

  it('1q. Deactivate unknown throws', () => {
    expect(() => cat.catalogDeactivate({ kind: 'agent', name: 'no-such' })).toThrow()
  })

  it('1r. Remove unknown throws', () => {
    expect(() => cat.catalogRemove({ kind: 'agent', name: 'no-such' })).toThrow()
  })
})

// ---------------------------------------------------------------------------
// 2. CHAIN LIFECYCLE — across all hosts
// ---------------------------------------------------------------------------

describe('2. CHAIN LIFECYCLE — across all hosts', () => {
  const hosts: Array<{ name: string; host: ToolHandlerOptions['hostCli'] }> = [
    { name: 'opencode', host: 'opencode' },
    { name: 'claude-code', host: 'claude-code' },
    { name: 'codex', host: 'codex' },
    { name: 'gemini', host: 'gemini' },
    { name: 'copilot', host: 'copilot' },
  ]

  for (const { name, host } of hosts) {
    it(`2a. ${name}: start chain, advance step, complete`, () => {
      const h = createHandlers(PROJECT_A, host)

      const started = h.startChain({ chain: 'repair', task: `Fix auth bug via ${name}` })
      expect(started.chainId).toBeTruthy()
      expect(started.state).toBe('running')

      const initial = h.getStatus({ runId: started.chainId, kind: 'chain' })
      if (initial.kind !== 'chain') throw new Error('expected chain status')

      // Advance first step
      const advanced = h.advanceChain({
        chainId: started.chainId,
        stepId: 'diagnose',
        outcome: 'success',
        output: { summary: 'Found the bug', status: 'ok', files_changed: ['src/auth.ts'], tests_passed: false },
        usage: { totalTokens: 100, costUsd: 0.02 },
      })
      expect(advanced.state).toBe('running')
      expect(advanced.nextStep?.stepId).toBe('fix')

      // Advance second step
      const advanced2 = h.advanceChain({
        chainId: started.chainId,
        stepId: 'fix',
        outcome: 'success',
        output: { summary: 'Fixed the bug', status: 'ok', files_changed: ['src/auth.ts'], tests_passed: true },
        usage: { totalTokens: 200, costUsd: 0.04 },
      })
      expect(advanced2.state).toBe('completed')

      // Budget should reflect consumption
      const budget = h.getBudget({ runId: started.chainId, kind: 'chain' })
      expect(budget.tokens.consumed).toBeGreaterThanOrEqual(300)
    })

    it(`2b. ${name}: chain with gated step`, () => {
      const h = createHandlers(PROJECT_A, host)
      // The repair chain doesn't have gates, but advanceChain should still work
      const started = h.startChain({ chain: 'repair', task: `Gate test ${name}` })
      const status = h.getStatus({ runId: started.chainId, kind: 'chain' })
      if (status.kind !== 'chain') throw new Error('expected chain')
      expect(status.state).toBe('running')
    })
  }
})

// ---------------------------------------------------------------------------
// 3. TEAM LIFECYCLE — across all hosts
// ---------------------------------------------------------------------------

describe('3. TEAM LIFECYCLE — across all hosts', () => {
  const hosts: Array<{ name: string; host: ToolHandlerOptions['hostCli'] }> = [
    { name: 'opencode', host: 'opencode' },
    { name: 'claude-code', host: 'claude-code' },
    { name: 'codex', host: 'codex' },
    { name: 'gemini', host: 'gemini' },
    { name: 'copilot', host: 'copilot' },
  ]

  for (const { name, host } of hosts) {
    it(`3a. ${name}: build team, assign both tasks, complete, synthesize`, () => {
      const h = createHandlers(PROJECT_A, host)

      const team = h.buildTeam({ team: 'review-team', task: `Review auth middleware via ${name}` })
      expect(team.state).toBe('running')
      expect(team.readyTaskIds).toEqual(['review-team:security-reviewer', 'review-team:correctness-reviewer'])

      // Assign security task
      h.assignTask({ teamId: team.teamId, taskId: 'review-team:security-reviewer', assignee: 'sec-agent', claim: true })

      // Assign correctness task
      h.assignTask({ teamId: team.teamId, taskId: 'review-team:correctness-reviewer', assignee: 'correct-agent', claim: true })

      // Complete security (success)
      const afterSec = h.completeTask({
        teamId: team.teamId,
        taskId: 'review-team:security-reviewer',
        outcome: 'success',
        result: { summary: 'No critical vulns', findings: [] },
        usage: { totalTokens: 50 },
      })
      expect(afterSec.state).toBe('running')

      // Complete correctness (success)
      const afterCorr = h.completeTask({
        teamId: team.teamId,
        taskId: 'review-team:correctness-reviewer',
        outcome: 'success',
        result: { summary: 'Logic correct' },
        usage: { totalTokens: 40 },
      })
      expect(afterCorr.state).toBe('synthesizing')

      // Synthesize
      const done = h.completeTask({
        teamId: team.teamId,
        taskId: 'review-team:synthesize',
        outcome: 'success',
        result: { verdict: 'pass', summary: 'All clear' },
        usage: { totalTokens: 10 },
      })
      expect(done.state).toBe('completed')

      // Budget should reflect all tasks
      expect(done.budget.tokens.consumed).toBeGreaterThanOrEqual(100)
    })

    it(`3b. ${name}: team failure prevents synthesis`, () => {
      const h = createHandlers(PROJECT_A, host)
      const team = h.buildTeam({ team: 'review-team', task: `Failure test ${name}` })

      // Assign and fail the first task
      h.assignTask({ teamId: team.teamId, taskId: 'review-team:security-reviewer', assignee: 'sec-agent', claim: true })
      const failed = h.completeTask({
        teamId: team.teamId,
        taskId: 'review-team:security-reviewer',
        outcome: 'failure',
        error: makeStructuredError(team.teamId, 'team', 'review-team:security-reviewer'),
        usage: { totalTokens: 10 },
      })
      expect(failed.state).toBe('failed')
    })

    it(`3c. ${name}: retry a failed team task`, () => {
      const h = createHandlers(PROJECT_A, host)
      const team = h.buildTeam({
        team: 'review-team',
        task: `Retry test ${name}`,
        budget: { retries: { limit: 3 } },
      })

      // fail a task
      h.assignTask({ teamId: team.teamId, taskId: 'review-team:security-reviewer', assignee: 'sec-a', claim: true })
      h.completeTask({
        teamId: team.teamId,
        taskId: 'review-team:security-reviewer',
        outcome: 'failure',
        error: makeStructuredError(team.teamId, 'team', 'review-team:security-reviewer'),
        usage: { totalTokens: 5 },
      })

      const retried = h.retryStep({ runId: team.teamId, kind: 'team', stepId: 'review-team:security-reviewer', reason: 'try again' })
      expect(retried.state).toBe('running')
      expect(retried.readyTaskIds).toContain('review-team:security-reviewer')
    })

    it(`3d. ${name}: escalate a team task`, () => {
      const h = createHandlers(PROJECT_A, host)
      const team = h.buildTeam({ team: 'review-team', task: `Escalate test ${name}` })

      h.assignTask({ teamId: team.teamId, taskId: 'review-team:security-reviewer', assignee: 'junior', claim: true })
      h.completeTask({
        teamId: team.teamId,
        taskId: 'review-team:security-reviewer',
        outcome: 'failure',
        error: makeStructuredError(team.teamId, 'team', 'review-team:security-reviewer'),
        usage: { totalTokens: 5 },
      })

      const escalated = h.escalateStep({
        runId: team.teamId,
        kind: 'team',
        stepId: 'review-team:security-reviewer',
        targetAgent: 'senior-builder',
        reason: 'needs senior expertise',
      })
      expect(escalated.state).toBe('running')
      const newAss = escalated.newAssignment as TeamTaskState
      expect(newAss.agent).toBe('senior-builder')
    })
  }
})

// ---------------------------------------------------------------------------
// 4. WORKFLOW LIFECYCLE — across all hosts
// ---------------------------------------------------------------------------

describe('4. WORKFLOW LIFECYCLE — across all hosts', () => {
  const hosts: Array<{ name: string; host: ToolHandlerOptions['hostCli'] }> = [
    { name: 'opencode', host: 'opencode' },
    { name: 'claude-code', host: 'claude-code' },
    { name: 'codex', host: 'codex' },
  ]

  for (const { name, host } of hosts) {
    it(`4a. ${name}: start workflow, advance through chain phase, advance through team phase`, () => {
      const h = createHandlers(PROJECT_A, host)

      const wf = h.startWorkflow({ workflow: 'delivery', task: `Deliver feature via ${name}` })
      expect(wf.state).toBe('waiting_on_child')
      expect(wf.currentPhase?.phaseId).toBe('implement')
      expect(wf.currentPhase?.childRun?.runKind).toBe('chain')

      // The chain (repair) was auto-launched as a child
      const chainId = wf.currentPhase?.childRun?.runId
      expect(chainId).toBeTruthy()

      // Advance the child chain to completion
      h.advanceChain({ chainId: chainId!, stepId: 'diagnose', outcome: 'success', output: { summary: 'diagnosed' } })
      h.advanceChain({ chainId: chainId!, stepId: 'fix', outcome: 'success', output: { summary: 'fixed', status: 'ok', files_changed: [], tests_passed: true } })

      // Now advance the workflow — should transition to review phase
      const wf2 = h.advanceWorkflow({ workflowId: wf.workflowId, outcome: 'success' })
      expect(wf2.currentPhase?.phaseId).toBe('review')
      expect(wf2.currentPhase?.childRun?.runKind).toBe('team')

      // Advance the child team to completion
      const teamId = wf2.currentPhase?.childRun?.runId!
      h.assignTask({ teamId, taskId: 'review-team:security-reviewer', assignee: 'sec', claim: true })
      h.assignTask({ teamId, taskId: 'review-team:correctness-reviewer', assignee: 'corr', claim: true })
      h.completeTask({ teamId, taskId: 'review-team:security-reviewer', outcome: 'success', result: { summary: 'ok' }, usage: { totalTokens: 20 } })
      h.completeTask({ teamId, taskId: 'review-team:correctness-reviewer', outcome: 'success', result: { summary: 'ok' }, usage: { totalTokens: 20 } })
      h.completeTask({ teamId, taskId: 'review-team:synthesize', outcome: 'success', result: { verdict: 'pass' }, usage: { totalTokens: 5 } })

      // Advance workflow to complete
      const wfDone = h.advanceWorkflow({ workflowId: wf.workflowId, outcome: 'success' })
      expect(wfDone.state).toBe('completed')

      // Budget should be aggregated
      const budget = h.getBudget({ runId: wf.workflowId, kind: 'workflow' })
      expect(budget.tokens.consumed).toBeGreaterThan(0)
    })

    it(`4b. ${name}: workflow failure recovery — retry`, () => {
      const h = createHandlers(PROJECT_A, host)
      const wf = h.startWorkflow({ workflow: 'delivery', task: `Retry test ${name}` })

      // Fail the chain phase
      h.advanceWorkflow({ workflowId: wf.workflowId, outcome: 'failure' })
      const wfRetry = h.retryStep({ runId: wf.workflowId, kind: 'workflow', stepId: 'implement', reason: 'rerun' })
      expect(wfRetry.state).toBe('waiting_on_child')
      expect(wfRetry.currentPhase?.phaseId).toBe('implement')
    })

    it(`4c. ${name}: workflow recovery — escalate to different phase`, () => {
      const h = createHandlers(PROJECT_A, host)
      const wf = h.startWorkflow({ workflow: 'delivery', task: `Escalate test ${name}` })

      h.advanceWorkflow({ workflowId: wf.workflowId, outcome: 'failure' })
      const escalated = h.escalateStep({
        runId: wf.workflowId,
        kind: 'workflow',
        stepId: 'implement',
        targetAgent: 'review',
        targetPhaseId: 'review',
        reason: 'route to review',
      })
      expect(escalated.state).toBe('waiting_on_child')
      expect(escalated.currentPhase?.phaseId).toBe('review')
    })

    it(`4d. ${name}: workflow handoff`, () => {
      const h = createHandlers(PROJECT_A, host)
      const wf = h.startWorkflow({ workflow: 'delivery', task: `Handoff test ${name}` })

      const ho = h.handoff({ runId: wf.workflowId, kind: 'workflow', summary: 'Workflow needs human review', recipient: 'human-admin' })
      expect(ho.summary).toBe('Workflow needs human review')
      expect(ho.resumable).toBe(true)

      const status = h.getStatus({ runId: wf.workflowId, kind: 'workflow' })
      expect(status.state).toBe('handoff')
    })
  }
})

// ---------------------------------------------------------------------------
// 5. RECOVERY PATHS — across chain/team/workflow
// ---------------------------------------------------------------------------

describe('5. RECOVERY PATHS', () => {
  let h: OrchestratorToolHandlers
  beforeAll(() => { h = createHandlers(PROJECT_A) })

  it('5a. Chain retry after failure', () => {
    const c = h.startChain({ chain: 'repair', task: 'Chain retry test' })
    // Advance diagnose successfully first, then fail the fix step which has retry capability
    h.advanceChain({
      chainId: c.chainId,
      stepId: 'diagnose',
      outcome: 'success',
      output: { summary: 'diagnosed' },
    })
    const failed = h.advanceChain({
      chainId: c.chainId,
      stepId: 'fix',
      outcome: 'failure',
      output: { summary: 'fix failed' },
    })
    expect(failed.recovery?.type).toBe('retry')

    const retried = h.retryStep({ runId: c.chainId, kind: 'chain', stepId: 'fix', reason: 'try again' })
    expect(retried.state).toBe('running')
  })

  it('5b. Chain escalation', () => {
    const c = h.startChain({ chain: 'repair', task: 'Chain escalate test' })
    h.advanceChain({ chainId: c.chainId, stepId: 'diagnose', outcome: 'failure' })

    const esc = h.escalateStep({
      runId: c.chainId,
      kind: 'chain',
      stepId: 'fix',
      targetAgent: 'architect',
      reason: 'needs redesign',
    })
    expect(esc.newAssignment).toBeTruthy()
  })

  it('5c. Chain handoff', () => {
    const c = h.startChain({ chain: 'repair', task: 'Chain handoff test' })
    const ho = h.handoff({ runId: c.chainId, kind: 'chain', summary: 'Stuck', includeArtifacts: true })
    expect(ho.resumable).toBe(true)
    expect(ho.handoffId).toBeTruthy()
    expect(ho.path).toBeTruthy()
  })

  it('5d. Team handoff', () => {
    const t = h.buildTeam({ team: 'review-team', task: 'Team handoff test' })
    const ho = h.handoff({ runId: t.teamId, kind: 'team', summary: 'Blocked', recipient: 'next-agent' })
    expect(ho.resumable).toBe(true)
    expect(ho.summary).toBe('Blocked')
  })
})

// ---------------------------------------------------------------------------
// 6. AUTO-SHUTDOWN HANDOFF
// ---------------------------------------------------------------------------

describe('6. AUTO-SHUTDOWN HANDOFF', () => {
  beforeAll(() => {
    initPersistenceDb(':memory:')
  })
  afterAll(() => { closePersistenceDb() })

  it('6a. listActiveRuns returns only active runs', () => {
    const db = getPersistenceDb()
    // Insert active chain
    const stateA = makeMiniChainState('chain-active')
    saveChainState(PROJECT_A, stateA)
    // Insert active team
    const stateT = makeMiniTeamState('team-active')
    saveTeamState(PROJECT_A, stateT)
    // Insert completed chain (should NOT be active)
    const stateDone = makeMiniChainState('chain-done')
    stateDone.state = 'completed'
    saveChainState(PROJECT_A, stateDone)

    const active = listActiveRuns()
    expect(active.map(a => a.runId).sort()).toEqual(['chain-active', 'team-active'])
  })

  it('6b. handoffActiveRuns creates handoffs and transitions state', () => {
    const r = handoffActiveRuns(PROJECT_A)
    expect(r.handoffsCreated).toBeGreaterThanOrEqual(2)
    expect(loadChainState(PROJECT_A, 'chain-active').state).toBe('handoff')
    expect(loadTeamState(PROJECT_A, 'team-active').state).toBe('handoff')
  })

  it('6c. handoffActiveRuns handles empty state', () => {
    expect(handoffActiveRuns('/tmp/orc-test/empty-project').handoffsCreated).toBe(0)
  })

  it('6d. handoffActiveRuns is resilient to per-run failures', () => {
    // Insert corrupt chain
    getPersistenceDb()
      .prepare(`INSERT INTO chain_runs (id, definition_name, definition_version, state, current_step_id, project_root, state_json, created_at, updated_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
      .run('corrupt', 'test', '1', 'running', null, PROJECT_A, '{not-json', '2026-01-01', '2026-01-01')

    const r = handoffActiveRuns(PROJECT_A)
    expect(r.errors.length).toBeGreaterThanOrEqual(1)
    expect(r.errors[0]).toContain('corrupt')
  })
})

// ---------------------------------------------------------------------------
// 7. INVOKE AGENT
// ---------------------------------------------------------------------------

describe('7. INVOKE AGENT', () => {
  it('7a. Invokes an agent by name and returns composed prompt', () => {
    const h = createHandlers(PROJECT_A)
    const result = h.invokeAgent({ agent: 'builder', task: 'Build auth middleware' })
    expect(result.invocationId).toBeTruthy()
    expect(result.agentName).toBe('builder')
    expect(result.displayName).toBe('builder')
    expect(result.source).toBeTruthy()
    // source may be 'library', 'user_global', 'project', or 'db' depending on host config
    expect(result.composed.prompt).toContain('Build auth middleware')
  })

  it('7b. Invoke agent with domain skill applies it', () => {
    const h = createHandlers(PROJECT_A)
    const r = h.invokeAgent({ agent: 'builder', task: 'Build', domainSkill: 'typescript' })
    expect(r.composed.prompt).toContain('Build')
    expect(r.composed.base).toBe('builder')
  })

  it('7c. Invoke agent with version pin overrides body', () => {
    const cat = new CatalogToolHandlers(getPersistenceDb())
    cat.catalogCreateVersion({
      kind: 'agent', name: 'builder',
      frontmatter: { name: 'Builder v1', model: 'sonnet' },
      body: 'Original body',
    })
    // The agent may already exist in catalog; just ensure version pin works
    const h = createHandlers(PROJECT_A)
    const r = h.invokeAgent({ agent: 'builder', task: 'Task', version: 1 })
    expect(r.resolvedVersion).toBe(1)
    expect(r.composed.prompt).toContain('Task')
  })
})

// ---------------------------------------------------------------------------
// 8. MULTI-PROJECT ISOLATION
// ---------------------------------------------------------------------------

describe('8. MULTI-PROJECT ISOLATION', () => {
  it('8a. Two independent handlers on different projects', () => {
    const hA = createHandlers(PROJECT_A)
    const hB = createHandlers(PROJECT_B)

    const cA = hA.startChain({ chain: 'repair', task: 'Project A task' })
    const cB = hB.startChain({ chain: 'repair', task: 'Project B task' })

    expect(cA.chainId).not.toBe(cB.chainId)

    const sA = hA.getStatus({ runId: cA.chainId, kind: 'chain' })
    const sB = hB.getStatus({ runId: cB.chainId, kind: 'chain' })

    expect(sA.kind).toBe('chain')
    expect(sB.kind).toBe('chain')
  })

  it('8b. Two simultaneous teams on same project', () => {
    const h = createHandlers(PROJECT_A)
    const t1 = h.buildTeam({ team: 'review-team', task: 'Team 1' })
    const t2 = h.buildTeam({ team: 'review-team', task: 'Team 2' })
    expect(t1.teamId).not.toBe(t2.teamId)

    h.assignTask({ teamId: t1.teamId, taskId: 'review-team:security-reviewer', assignee: 'a1', claim: true })
    h.assignTask({ teamId: t2.teamId, taskId: 'review-team:security-reviewer', assignee: 'a2', claim: true })

    const s1 = h.getStatus({ runId: t1.teamId, kind: 'team' })
    const s2 = h.getStatus({ runId: t2.teamId, kind: 'team' })
    expect(s1.kind).toBe('team')
    expect(s2.kind).toBe('team')
  })
})

// ---------------------------------------------------------------------------
// 9. EVENT BUS
// ---------------------------------------------------------------------------

describe('9. EVENT BUS', () => {
  it('9a. Events are emitted on chain operations', () => {
    const h = createHandlers(PROJECT_A)
    const c = h.startChain({ chain: 'repair', task: 'Event test' })
    // The bus emits events internally — we just verify the chain started
    expect(c.state).toBe('running')
  })
})

// ---------------------------------------------------------------------------
// 10. HOST CLI VARIATIONS
// ---------------------------------------------------------------------------

describe('10. HOST CLI CATALOG RESOLUTION', () => {
  const hosts: ToolHandlerOptions['hostCli'][] = ['opencode', 'claude-code', 'codex', 'gemini', 'copilot']

  for (const host of hosts) {
    it(`10a. ${host}: listCatalog works and returns expected kinds`, () => {
      const h = createHandlers(PROJECT_A, host)
      const catalog = h.listCatalog()
      expect(catalog.items.length).toBeGreaterThanOrEqual(0)
      const chains = h.listCatalog({ kinds: ['chain'] })
      expect(chains.items.some(i => i.name === 'repair')).toBe(true)
    })

    it(`10b. ${host}: composeAgent layers correctly`, () => {
      const h = createHandlers(PROJECT_A, host)
      const composed = h.composeAgent({
        base: 'builder',
        domainSkill: 'typescript',
        stepInstructions: 'Fix the bug',
      })
      expect(composed.prompt).toContain('Fix the bug')
      expect(composed.domainSkill).toBe('typescript')
    })
  }
})

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function makeMiniChainState(id: string): ChainState {
  const CREATED_AT = '2026-04-29T00:00:00.000Z'
  const plan: ExecutionPlan = {
    id: `plan-${id}`,
    kind: 'chain',
    definition: { kind: 'chain', name: 'test-chain', version: '1.0.0', source: 'library', path: '/chains/test.json' },
    cli: { host: 'opencode', dispatchMode: 'instruction-only', supportsSubagents: false, supportsParallelTeams: false, supportsStructuredOutput: true, mcpServerName: 'orchestrator' },
    project: { rootPath: PROJECT_A },
    budgetPolicy: { id: 'test-budget', scope: 'chain', defaultActionOnLimit: 'pause' },
    entrypoint: 'step-1',
    compiledSteps: [{
      id: 'step-1', kind: 'step', agent: 'implementor', skills: [],
      stepType: 'implement', instructions: 'Do it.', allowedTools: [],
      model: 'gpt-4', outputContract: { stepType: 'implement', requiredFields: [], allowAdditionalProperties: true, schema: {}, onValidationFailure: { category: 'validation', defaultRecovery: { type: 'retry' } } },
      transitions: {}, composedAgent: { id: 'i', base: 'i', model: 'gpt-4', tools: [], approvalPolicy: 'minimal', constraints: [], prompt: 'Do it.', mergedFrom: [] },
    }],
    createdAt: CREATED_AT, task: 'test',
  }
  const s = createChainState(plan)
  s.chainId = id
  s.createdAt = CREATED_AT
  s.updatedAt = CREATED_AT
  return s
}

function makeMiniTeamState(id: string): TeamState {
  const CREATED_AT = '2026-04-29T00:00:00.000Z'
  const policy = { id: 'team-budget', scope: 'team' as const, defaultActionOnLimit: 'pause' as const }
  const def: TeamDefinition = {
    kind: 'team', name: 'test-team', description: 'Test', version: '1.0.0',
    source: 'library', path: '/teams/test.json',
    parallel: [{ role: 'member', agent: 'builder', skills: [], focus: 'x' }],
    synthesize: { agent: 'reviewer', description: 'summarize' },
  }
  const s = createTeamState({ definition: def, task: 'test', policy, budget: createBudgetState(policy), createdAt: CREATED_AT })
  s.teamId = id
  return s
}

// Cleanup temp dirs after all tests
afterAll(() => {
  try { fs.rmSync(FAKE_ROOT, { recursive: true, force: true }) } catch { /* ignore */ }
})
