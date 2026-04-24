import { describe, expect, it } from 'vitest'
import { createBudgetState } from '../budget-tracker.js'
import { assignTeamTask, completeTeamTask, createTeamState } from '../team-machine.js'
import type { BudgetPolicy, TeamDefinition } from '../types.js'

const budgetPolicy: BudgetPolicy = {
  id: 'team-budget-1',
  scope: 'team',
  defaultActionOnLimit: 'pause',
  tokens: { limit: 200, warnAt: 0.5 },
}

const teamDefinition: TeamDefinition = {
  kind: 'team',
  name: 'review-team',
  description: 'Review a change in parallel.',
  version: '1.0.0',
  source: 'library',
  path: '/teams/review-team.json',
  parallel: [
    {
      role: 'correctness-reviewer',
      agent: 'reviewer',
      skills: [],
      focus: 'Logic and behavior',
    },
    {
      role: 'security-reviewer',
      agent: 'red-team',
      skills: [],
      focus: 'Security',
    },
  ],
  synthesize: {
    agent: 'orchestrator',
    description: 'Merge the review findings.',
  },
}

describe('team-machine', () => {
  it('assigns, claims, and synthesizes a completed team run', () => {
    const state = createTeamState({
      definition: teamDefinition,
      task: 'Review auth middleware',
      policy: budgetPolicy,
      budget: createBudgetState(budgetPolicy),
      createdAt: '2026-01-01T00:00:00.000Z',
    })

    const correctnessTaskId = `${teamDefinition.name}:correctness-reviewer`
    const securityTaskId = `${teamDefinition.name}:security-reviewer`
    const synthesisTaskId = `${teamDefinition.name}:synthesize`

    const assigned = assignTeamTask({
      state,
      taskId: correctnessTaskId,
      assignee: 'reviewer-a',
    })
    const claimed = assignTeamTask({
      state: assigned.state,
      taskId: correctnessTaskId,
      assignee: 'reviewer-a',
      claim: true,
    })

    expect(claimed.state.tasks.find((task) => task.taskId === correctnessTaskId)?.state).toBe('claimed')

    const afterCorrectness = completeTeamTask({
      state: claimed.state,
      taskId: correctnessTaskId,
      outcome: 'success',
      result: { summary: 'Looks correct' },
      usage: { inputTokens: 10, outputTokens: 20 },
      policy: budgetPolicy,
      now: '2026-01-01T00:01:00.000Z',
    })

    const afterSecurity = completeTeamTask({
      state: afterCorrectness.state,
      taskId: securityTaskId,
      outcome: 'success',
      result: { summary: 'No security issues' },
      usage: { inputTokens: 15, outputTokens: 25 },
      policy: budgetPolicy,
      now: '2026-01-01T00:02:00.000Z',
    })

    expect(afterSecurity.state.state).toBe('synthesizing')
    expect(afterSecurity.state.tasks.find((task) => task.taskId === synthesisTaskId)?.state).toBe('pending')

    const afterSynthesis = completeTeamTask({
      state: afterSecurity.state,
      taskId: synthesisTaskId,
      outcome: 'success',
      result: { verdict: 'pass', summary: 'Ready to merge' },
      usage: { inputTokens: 5, outputTokens: 10 },
      policy: budgetPolicy,
      now: '2026-01-01T00:03:00.000Z',
    })

    expect(afterSynthesis.state.state).toBe('completed')
    expect(afterSynthesis.state.summary?.verdict).toBe('pass')
    expect(afterSynthesis.evaluation.overall).toBe('ok')
    expect(afterSynthesis.state.budget.tokens.consumed).toBe(85)
  })

  it('fails the team when a member task fails', () => {
    const state = createTeamState({
      definition: teamDefinition,
      task: 'Review auth middleware',
      policy: budgetPolicy,
      budget: createBudgetState(budgetPolicy),
      createdAt: '2026-01-01T00:00:00.000Z',
    })

    const failure = completeTeamTask({
      state,
      taskId: `${teamDefinition.name}:security-reviewer`,
      outcome: 'failure',
      error: {
        category: 'logical',
        code: 'TEAM_MEMBER_FAILED',
        message: 'Security review found a blocker.',
        stepId: 'security-reviewer',
        agent: 'red-team',
        skills: [],
        context: {
          runId: 'team-1',
          runKind: 'team',
          task: 'Review auth middleware',
          attempt: 1,
          hostCli: 'opencode',
        },
        suggestedRecovery: { type: 'escalate', targetAgent: 'planner', reason: 'Re-plan the change' },
        timestamp: '2026-01-01T00:02:00.000Z',
      },
      policy: budgetPolicy,
      now: '2026-01-01T00:02:00.000Z',
    })

    expect(failure.state.state).toBe('failed')
    expect(failure.state.tasks.find((task) => task.taskId === `${teamDefinition.name}:synthesize`)?.state).toBe('blocked')
  })
})
