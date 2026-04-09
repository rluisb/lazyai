import { describe, expect, it } from 'vitest'
import { createBudgetState } from '../budget-tracker.js'
import {
  applyWorkflowChildLaunch,
  advanceWorkflowState,
  createWorkflowState,
} from '../workflow-machine.js'
import type {
  BudgetPolicy,
  StructuredError,
  WorkflowDefinition,
  WorkflowRecoveryDecision,
} from '../types.js'

const budgetPolicy: BudgetPolicy = {
  id: 'workflow-budget-1',
  scope: 'workflow',
  defaultActionOnLimit: 'pause',
}

const workflowDefinition: WorkflowDefinition = {
  kind: 'workflow',
  name: 'deliver-feature',
  description: 'Run implementation, request approval, then parallel review.',
  version: '1.0.0',
  source: 'library',
  path: '/workflows/deliver-feature.json',
  entry: 'implement',
  phases: [
    {
      id: 'implement',
      kind: 'chain',
      ref: 'feature',
      on: {
        success: 'approve',
        failure: 'review',
      },
    },
    {
      id: 'approve',
      kind: 'gate',
      gate: 'user_approval',
      prompt: 'Approve implementation?',
      on: {
        approved: 'review',
        rejected: 'handoff',
      },
    },
    {
      id: 'review',
      kind: 'team',
      ref: 'review-team',
      on: {
        success: 'complete',
        failure: 'handoff',
      },
    },
    {
      id: 'handoff',
      kind: 'terminal',
    },
    {
      id: 'complete',
      kind: 'terminal',
    },
  ],
}

function makeChildError(): StructuredError {
  return {
    category: 'logical',
    code: 'CHAIN_FAILED',
    message: 'Implementation chain failed review.',
    stepId: 'review',
    agent: 'reviewer',
    skills: ['extract-standards'],
    context: {
      runId: 'chain-1',
      runKind: 'chain',
      task: 'Deliver auth middleware',
      attempt: 1,
      hostCli: 'opencode',
      child: {
        runId: 'chain-1',
        runKind: 'chain',
        definitionName: 'feature',
        stepId: 'review',
        phaseId: 'implement',
      },
    },
    suggestedRecovery: { type: 'retry', maxAttempts: 1 },
    timestamp: '2026-01-01T00:01:00.000Z',
  }
}

describe('workflow-machine', () => {
  it('advances through child, gate, and terminal phases', () => {
    const created = createWorkflowState({
      definition: workflowDefinition,
      task: 'Deliver auth middleware',
      policy: budgetPolicy,
      budget: createBudgetState(budgetPolicy),
      createdAt: '2026-01-01T00:00:00.000Z',
    })

    expect(created.nextAction?.type).toBe('start_child')
    expect(created.nextAction?.phaseId).toBe('implement')

    const withChain = applyWorkflowChildLaunch({
      state: created.state,
      phaseId: 'implement',
      childRun: {
        phaseId: 'implement',
        runId: 'chain-1',
        runKind: 'chain',
        definitionName: 'feature',
        launchedAt: '2026-01-01T00:00:00.000Z',
      },
      now: '2026-01-01T00:00:00.000Z',
    })

    const afterChain = advanceWorkflowState({
      state: withChain.state,
      definition: workflowDefinition,
      outcome: 'success',
      now: '2026-01-01T00:01:00.000Z',
    })

    expect(afterChain.state.state).toBe('gated')
    expect(afterChain.nextAction?.type).toBe('gate')

    const afterApproval = advanceWorkflowState({
      state: afterChain.state,
      definition: workflowDefinition,
      outcome: 'approved',
      now: '2026-01-01T00:02:00.000Z',
    })

    expect(afterApproval.nextAction?.type).toBe('start_child')
    expect(afterApproval.nextAction?.childKind).toBe('team')

    const withTeam = applyWorkflowChildLaunch({
      state: afterApproval.state,
      phaseId: 'review',
      childRun: {
        phaseId: 'review',
        runId: 'team-1',
        runKind: 'team',
        definitionName: 'review-team',
        launchedAt: '2026-01-01T00:02:00.000Z',
      },
      now: '2026-01-01T00:02:00.000Z',
    })

    const afterTeam = advanceWorkflowState({
      state: withTeam.state,
      definition: workflowDefinition,
      outcome: 'success',
      now: '2026-01-01T00:03:00.000Z',
    })

    expect(afterTeam.state.state).toBe('completed')
    expect(afterTeam.nextAction).toBeNull()
  })

  it('captures child failure context and supports retry recovery', () => {
    const created = createWorkflowState({
      definition: workflowDefinition,
      task: 'Deliver auth middleware',
      policy: budgetPolicy,
      budget: createBudgetState(budgetPolicy),
      createdAt: '2026-01-01T00:00:00.000Z',
    })

    const withChain = applyWorkflowChildLaunch({
      state: created.state,
      phaseId: 'implement',
      childRun: {
        phaseId: 'implement',
        runId: 'chain-1',
        runKind: 'chain',
        definitionName: 'feature',
        launchedAt: '2026-01-01T00:00:00.000Z',
      },
      now: '2026-01-01T00:00:00.000Z',
    })

    const failed = advanceWorkflowState({
      state: withChain.state,
      definition: workflowDefinition,
      outcome: 'failure',
      childError: makeChildError(),
      now: '2026-01-01T00:01:00.000Z',
    })

    expect(failed.state.state).toBe('awaiting_recovery')
    expect(failed.state.lastError?.context.child?.runId).toBe('chain-1')
    expect(failed.recoveryOptions.map((option) => option.type)).toContain('retry')

    const retryDecision: WorkflowRecoveryDecision = {
      type: 'retry',
    }

    const retried = advanceWorkflowState({
      state: failed.state,
      definition: workflowDefinition,
      recovery: retryDecision,
      now: '2026-01-01T00:02:00.000Z',
    })

    expect(retried.state.state).toBe('waiting_on_child')
    expect(retried.nextAction?.type).toBe('start_child')
    expect(retried.nextAction?.phaseId).toBe('implement')
  })
})
