import { describe, expect, it } from 'vitest'
import { advanceChainState, createChainState, retryChainStep } from '../chain-machine.js'
import { createBudgetState } from '../budget-tracker.js'
import type { ExecutionPlan } from '../types.js'

function makePlan(): ExecutionPlan {
  return {
    id: 'plan-1',
    kind: 'chain',
    definition: {
      kind: 'chain',
      name: 'repair',
      version: '1.0.0',
      source: 'library',
      path: '/chains/repair.json',
    },
    cli: {
      host: 'opencode',
      dispatchMode: 'task-tool',
      supportsSubagents: true,
      supportsParallelTeams: false,
      supportsStructuredOutput: true,
      mcpServerName: 'ai-setup-orchestrator',
    },
    project: { rootPath: '/repo' },
    budgetPolicy: { id: 'budget-1', scope: 'chain', defaultActionOnLimit: 'pause' },
    entrypoint: 'step-1',
    createdAt: '2026-01-01T00:00:00.000Z',
    task: 'repair issue',
    compiledSteps: [
      {
        id: 'step-1',
        kind: 'step',
        agent: 'builder',
        skills: [],
        stepType: 'implement',
        instructions: 'do work',
        allowedTools: ['Read'],
        model: 'sonnet',
        outputContract: {
          stepType: 'implement',
          requiredFields: ['summary'],
          allowAdditionalProperties: true,
          schema: { type: 'object' },
          onValidationFailure: { category: 'validation', defaultRecovery: { type: 'retry' } },
        },
        transitions: { success: 'step-2', failure: { retry: 1, then: 'done' } },
        composedAgent: {
          id: 'builder:step-1',
          base: 'builder',
          model: 'sonnet',
          tools: ['Read'],
          approvalPolicy: 'minimal',
          constraints: [],
          prompt: 'do work',
          mergedFrom: [],
        },
      },
      {
        id: 'step-2',
        kind: 'step',
        agent: 'reviewer',
        skills: [],
        stepType: 'review',
        instructions: 'review work',
        allowedTools: ['Read'],
        model: 'sonnet',
        outputContract: {
          stepType: 'review',
          requiredFields: ['summary'],
          allowAdditionalProperties: true,
          schema: { type: 'object' },
          onValidationFailure: { category: 'validation', defaultRecovery: { type: 'retry' } },
        },
        transitions: { success: 'done' },
        composedAgent: {
          id: 'reviewer:step-2',
          base: 'reviewer',
          model: 'sonnet',
          tools: ['Read'],
          approvalPolicy: 'minimal',
          constraints: [],
          prompt: 'review work',
          mergedFrom: [],
        },
      },
    ],
  }
}

describe('chain-machine', () => {
  it('advances to the next step on success', () => {
    const plan = makePlan()
    const state = createChainState(plan)
    state.budget = createBudgetState(plan.budgetPolicy)

    const result = advanceChainState({
      state,
      plan,
      stepId: 'step-1',
      outcome: 'success',
      output: { summary: 'done' },
    })

    expect(result.state).toBe('running')
    expect(result.nextStep?.stepId).toBe('step-2')
    expect(result.stateSnapshot.currentStepId).toBe('step-2')
    expect(result.stateSnapshot.completedStepIds).toContain('step-1')
  })

  it('retries a failed step when retries remain', () => {
    const plan = makePlan()
    const state = createChainState(plan)
    state.budget = createBudgetState(plan.budgetPolicy)
    state.steps[0]!.state = 'failed'

    const retried = retryChainStep(state, plan, 'step-1')
    expect(retried.state.currentStepId).toBe('step-1')
    expect(retried.state.steps[0]?.state).toBe('running')
    expect(retried.attemptsRemaining).toBe(0)
  })
})
