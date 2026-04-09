import { describe, expect, it } from 'vitest'
import { createBudgetState, updateBudget } from '../budget-tracker.js'

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
})
