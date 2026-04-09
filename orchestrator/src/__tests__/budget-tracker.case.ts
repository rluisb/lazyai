import { describe, expect, it } from 'vitest'
import { aggregateBudgets, createBudgetState, updateBudget } from '../budget-tracker.js'

describe('budget-tracker', () => {
  it('tracks usage and warns at configured thresholds', () => {
    const policy = {
      id: 'budget-1',
      scope: 'chain' as const,
      defaultActionOnLimit: 'pause' as const,
      tokens: { limit: 100, warnAt: 0.5 },
    }

    const initial = createBudgetState(policy)
    const { budget, evaluation } = updateBudget({
      budget: initial,
      policy,
      stepId: 'step-1',
      usage: { inputTokens: 20, outputTokens: 40 },
      now: '2026-01-01T00:00:00.000Z',
    })

    expect(budget.tokens.consumed).toBe(60)
    expect(budget.tokens.remaining).toBe(40)
    expect(budget.byStep['step-1']).toEqual({
      inputTokens: 20,
      outputTokens: 40,
      totalTokens: 0,
      costUsd: 0,
      wallClockMs: 0,
    })
    expect(evaluation.overall).toBe('warning')
    expect(evaluation.recommendedAction).toBe('warn')
  })

  it('aggregates child budgets for workflow and team scopes', () => {
    const policy = {
      id: 'budget-2',
      scope: 'workflow' as const,
      defaultActionOnLimit: 'pause' as const,
      tokens: { limit: 200, warnAt: 0.5 },
    }

    const childA = createBudgetState({
      id: 'child-a',
      scope: 'chain',
      defaultActionOnLimit: 'pause',
      tokens: { limit: 100, warnAt: 0.5 },
    })
    const childB = createBudgetState({
      id: 'child-b',
      scope: 'team',
      defaultActionOnLimit: 'pause',
      tokens: { limit: 100, warnAt: 0.5 },
    })

    const withA = updateBudget({
      budget: childA,
      policy: { id: 'child-a', scope: 'chain', defaultActionOnLimit: 'pause', tokens: { limit: 100, warnAt: 0.5 } },
      stepId: 'step-1',
      usage: { totalTokens: 60 },
    }).budget
    const withB = updateBudget({
      budget: childB,
      policy: { id: 'child-b', scope: 'team', defaultActionOnLimit: 'pause', tokens: { limit: 100, warnAt: 0.5 } },
      stepId: 'task-1',
      usage: { totalTokens: 45 },
    }).budget

    const aggregated = aggregateBudgets({
      policy,
      scope: 'workflow',
      children: [
        { childId: 'chain-1', budget: withA },
        { childId: 'team-1', budget: withB },
      ],
      now: '2026-01-01T00:00:00.000Z',
    })

    expect(aggregated.budget.tokens.consumed).toBe(105)
    expect(aggregated.evaluation.overall).toBe('warning')
    expect(Object.keys(aggregated.budget.byStep)).toEqual(['chain-1:step-1', 'team-1:task-1'])
  })
})
