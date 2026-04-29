import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { createBudgetState } from '../budget-tracker.js'
import { createChainState } from '../chain-machine.js'
import {
  closePersistenceDb,
  getPersistenceDb,
  handoffActiveRuns,
  initPersistenceDb,
  listActiveRuns,
  loadChainState,
  loadTeamState,
  loadWorkflowState,
  saveChainState,
  saveTeamState,
  saveWorkflowState,
} from '../persistence.js'
import { createTeamState } from '../team-machine.js'
import type { BudgetPolicy, ExecutionPlan, HandoffDocument, TeamDefinition, WorkflowDefinition } from '../types.js'
import { createWorkflowState } from '../workflow-machine.js'

const PROJECT_ROOT = '/tmp/orchestrator-auto-handoff-test'
const CREATED_AT = '2026-04-29T00:00:00.000Z'

beforeEach(() => {
  initPersistenceDb(':memory:')
})

afterEach(() => {
  closePersistenceDb()
})

describe('auto-shutdown handoff persistence', () => {
  it('listActiveRuns returns only active runs', () => {
    insertChainRun('active-chain', 'running')
    insertChainRun('completed-chain', 'completed')
    insertChainRun('failed-chain', 'failed')

    expect(listActiveRuns()).toEqual([{ runId: 'active-chain', kind: 'chain', state: 'running' }])
  })

  it('listActiveRuns returns runs from all three tables', () => {
    insertChainRun('chain-1', 'gated')
    insertTeamRun('team-1', 'synthesizing')
    insertWorkflowRun('workflow-1', 'awaiting_recovery')

    expect(listActiveRuns()).toEqual([
      { runId: 'chain-1', kind: 'chain', state: 'gated' },
      { runId: 'team-1', kind: 'team', state: 'synthesizing' },
      { runId: 'workflow-1', kind: 'workflow', state: 'awaiting_recovery' },
    ])
  })

  it('handoffActiveRuns creates handoffs for active runs and transitions state', () => {
    const chainState = makeChainState('chain-active')
    const teamState = makeTeamState('team-active')
    const workflowState = makeWorkflowState('workflow-active')
    saveChainState(PROJECT_ROOT, chainState)
    saveTeamState(PROJECT_ROOT, teamState)
    saveWorkflowState(PROJECT_ROOT, workflowState)

    const result = handoffActiveRuns(PROJECT_ROOT)

    expect(result).toEqual({ handoffsCreated: 3, errors: [] })
    expect(loadChainState(PROJECT_ROOT, 'chain-active').state).toBe('handoff')
    expect(loadTeamState(PROJECT_ROOT, 'team-active').state).toBe('handoff')
    expect(loadWorkflowState(PROJECT_ROOT, 'workflow-active').state).toBe('handoff')

    const handoffs = loadHandoffDocuments()
    expect(handoffs).toHaveLength(3)
    expect(handoffs.map((handoff) => `${handoff.kind}:${handoff.runId}`).sort()).toEqual([
      'chain:chain-active',
      'team:team-active',
      'workflow:workflow-active',
    ])
    expect(handoffs.every((handoff) => handoff.resumable)).toBe(true)
  })

  it('handoffActiveRuns handles empty state gracefully', () => {
    expect(handoffActiveRuns(PROJECT_ROOT)).toEqual({ handoffsCreated: 0, errors: [] })
    expect(loadHandoffDocuments()).toEqual([])
  })

  it('handoffActiveRuns is resilient to per-run failures', () => {
    insertChainRun('corrupt-chain', 'running', '{not-json')
    saveTeamState(PROJECT_ROOT, makeTeamState('healthy-team'))

    const result = handoffActiveRuns(PROJECT_ROOT)

    expect(result.handoffsCreated).toBe(1)
    expect(result.errors).toHaveLength(1)
    expect(result.errors[0]).toContain('corrupt-chain')
    expect(loadTeamState(PROJECT_ROOT, 'healthy-team').state).toBe('handoff')

    const handoffs = loadHandoffDocuments()
    expect(handoffs).toHaveLength(1)
    expect(handoffs[0]?.runId).toBe('healthy-team')
  })
})

function insertChainRun(id: string, state: string, stateJson = '{}'): void {
  getPersistenceDb()
    .prepare<[string, string, string, string, string, string, string, string, string]>(`
      INSERT INTO chain_runs (id, definition_name, definition_version, state, current_step_id, project_root, state_json, created_at, updated_at)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `)
    .run(id, 'test-chain', '1.0.0', state, 'step-1', PROJECT_ROOT, stateJson, CREATED_AT, CREATED_AT)
}

function insertTeamRun(id: string, state: string): void {
  getPersistenceDb()
    .prepare<[string, string, string, string, string, string, string, string]>(`
      INSERT INTO team_runs (id, definition_name, definition_version, state, project_root, state_json, created_at, updated_at)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `)
    .run(id, 'test-team', '1.0.0', state, PROJECT_ROOT, '{}', CREATED_AT, CREATED_AT)
}

function insertWorkflowRun(id: string, state: string): void {
  getPersistenceDb()
    .prepare<[string, string, string, string, string, string, string, string, string]>(`
      INSERT INTO workflow_runs (id, definition_name, definition_version, state, current_phase_id, project_root, state_json, created_at, updated_at)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `)
    .run(id, 'test-workflow', '1.0.0', state, 'phase-1', PROJECT_ROOT, '{}', CREATED_AT, CREATED_AT)
}

function makeChainState(chainId: string) {
  const plan: ExecutionPlan = {
    id: `plan-${chainId}`,
    kind: 'chain',
    definition: { kind: 'chain', name: 'test-chain', version: '1.0.0', source: 'library', path: '/chains/test-chain.json' },
    cli: { host: 'opencode', dispatchMode: 'instruction-only', supportsSubagents: false, supportsParallelTeams: false, supportsStructuredOutput: true, mcpServerName: 'orchestrator' },
    project: { rootPath: PROJECT_ROOT },
    budgetPolicy: makeBudgetPolicy('chain'),
    entrypoint: 'step-1',
    compiledSteps: [
      {
        id: 'step-1',
        kind: 'step',
        agent: 'implementor',
        skills: [],
        stepType: 'implement',
        instructions: 'Implement the change.',
        allowedTools: [],
        model: 'gpt-5.5',
        outputContract: {
          stepType: 'implement',
          requiredFields: [],
          allowAdditionalProperties: true,
          schema: {},
          onValidationFailure: { category: 'validation', defaultRecovery: { type: 'retry' } },
        },
        transitions: {},
        composedAgent: {
          id: 'implementor',
          base: 'implementor',
          model: 'gpt-5.5',
          tools: [],
          approvalPolicy: 'minimal',
          constraints: [],
          prompt: 'Implement the change.',
          mergedFrom: [],
        },
      },
    ],
    createdAt: CREATED_AT,
    task: 'Test chain handoff',
  }
  const state = createChainState(plan)
  state.chainId = chainId
  return state
}

function makeTeamState(teamId: string) {
  const policy = makeBudgetPolicy('team')
  const definition: TeamDefinition = {
    kind: 'team',
    name: 'test-team',
    description: 'Test team.',
    version: '1.0.0',
    source: 'library',
    path: '/teams/test-team.json',
    parallel: [{ role: 'implementor', agent: 'implementor', skills: [], focus: 'Implement' }],
    synthesize: { agent: 'synthesizer', description: 'Synthesize results.' },
  }
  const state = createTeamState({ definition, task: 'Test team handoff', policy, budget: createBudgetState(policy), createdAt: CREATED_AT })
  state.teamId = teamId
  return state
}

function makeWorkflowState(workflowId: string) {
  const policy = makeBudgetPolicy('workflow')
  const definition: WorkflowDefinition = {
    kind: 'workflow',
    name: 'test-workflow',
    description: 'Test workflow.',
    version: '1.0.0',
    source: 'library',
    path: '/workflows/test-workflow.json',
    entry: 'implement',
    phases: [{ id: 'implement', kind: 'chain', ref: 'test-chain', on: { success: 'complete', failure: 'handoff' } }],
  }
  const { state } = createWorkflowState({ definition, task: 'Test workflow handoff', policy, budget: createBudgetState(policy), createdAt: CREATED_AT })
  state.workflowId = workflowId
  return state
}

function makeBudgetPolicy(scope: BudgetPolicy['scope']): BudgetPolicy {
  return { id: `${scope}-budget`, scope, defaultActionOnLimit: 'pause' }
}

function loadHandoffDocuments(): HandoffDocument[] {
  const rows = getPersistenceDb()
    .prepare<[], { doc_json: string }>('SELECT doc_json FROM handoffs ORDER BY run_kind, run_id')
    .all()
  return rows.map((row) => JSON.parse(row.doc_json) as HandoffDocument)
}
