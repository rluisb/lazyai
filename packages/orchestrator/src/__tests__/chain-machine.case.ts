import { describe, expect, it } from 'vitest'
import { advanceChainState, createChainState, retryChainStep } from '../chain-machine.js'
import { createBudgetState } from '../budget-tracker.js'
import type { CompiledStepPlan, ExecutionPlan, StepType } from '../types.js'

function makeCompiledStep(input: {
  id: string
  agent?: string
  skills?: string[]
  stepType?: StepType
  transitions: CompiledStepPlan['transitions']
  gate?: CompiledStepPlan['gate']
}): CompiledStepPlan {
  const agent = input.agent ?? 'builder'
  const stepType = input.stepType ?? 'custom'

  return {
    id: input.id,
    kind: 'step',
    agent,
    skills: input.skills ?? [],
    stepType,
    instructions: `${input.id} instructions`,
    allowedTools: ['Read'],
    model: 'sonnet',
    outputContract: {
      stepType,
      requiredFields: ['summary'],
      allowAdditionalProperties: true,
      schema: { type: 'object' },
      onValidationFailure: { category: 'validation', defaultRecovery: { type: 'retry' } },
    },
    transitions: input.transitions,
    ...(input.gate ? { gate: input.gate } : {}),
    composedAgent: {
      id: `${agent}:${input.id}`,
      base: agent,
      model: 'sonnet',
      tools: ['Read'],
      approvalPolicy: input.gate ? 'strict' : 'minimal',
      constraints: [],
      prompt: `${input.id} instructions`,
      mergedFrom: [],
    },
  }
}

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

function makePlanWithSteps(compiledSteps: CompiledStepPlan[], entrypoint = compiledSteps[0]!.id): ExecutionPlan {
  return {
    ...makePlan(),
    entrypoint,
    compiledSteps,
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

  it('stops at a sequential user approval gate and resumes through approved or rejected outcomes', () => {
    const plan = makePlanWithSteps([
      makeCompiledStep({
        id: 'plan-quality',
        agent: 'planner',
        skills: ['plan'],
        stepType: 'plan',
        transitions: { success: 'plan-gate' },
      }),
      makeCompiledStep({
        id: 'plan-gate',
        agent: 'planner',
        skills: [],
        stepType: 'plan',
        gate: 'user_approval',
        transitions: { approved: 'implement', rejected: 'plan-quality' },
      }),
      makeCompiledStep({
        id: 'implement',
        agent: 'builder',
        skills: ['implement'],
        stepType: 'implement',
        transitions: { success: 'done' },
      }),
    ])

    const state = createChainState(plan)
    state.budget = createBudgetState(plan.budgetPolicy)

    const afterQuality = advanceChainState({
      state,
      plan,
      stepId: 'plan-quality',
      outcome: 'success',
      output: { summary: 'quality report emitted' },
    })
    expect(afterQuality.state).toBe('running')
    expect(afterQuality.nextStep?.stepId).toBe('plan-gate')

    const gated = advanceChainState({
      state: afterQuality.stateSnapshot,
      plan,
      stepId: 'plan-gate',
      outcome: 'success',
      output: { summary: 'approval packet ready' },
    })
    expect(gated.state).toBe('gated')
    expect(gated.gate).toMatchObject({ type: 'user_approval', status: 'pending' })
    expect(gated.nextStep?.stepId).toBe('plan-gate')

    const approved = advanceChainState({
      state: gated.stateSnapshot,
      plan,
      stepId: 'plan-gate',
      outcome: 'approved',
    })
    expect(approved.state).toBe('running')
    expect(approved.nextStep?.stepId).toBe('implement')

    const rejected = advanceChainState({
      state: gated.stateSnapshot,
      plan,
      stepId: 'plan-gate',
      outcome: 'rejected',
    })
    expect(rejected.state).toBe('running')
    expect(rejected.nextStep?.stepId).toBe('plan-quality')
  })

  it('does not skip steps using unsupported runtime condition metadata', () => {
    const redTeamStep = {
      ...makeCompiledStep({
        id: 'red-team-plan',
        agent: 'reviewer',
        skills: ['red-team-plan'],
        stepType: 'plan',
        transitions: { success: 'plan-gate' },
      }),
      optionalByFeature: 'adversarialDesign',
    } as CompiledStepPlan & { optionalByFeature: string }

    const plan = makePlanWithSteps([
      makeCompiledStep({
        id: 'plan-quality',
        agent: 'planner',
        skills: ['plan'],
        stepType: 'plan',
        transitions: { success: 'red-team-plan' },
      }),
      redTeamStep,
      makeCompiledStep({
        id: 'plan-gate',
        agent: 'planner',
        skills: [],
        stepType: 'plan',
        gate: 'user_approval',
        transitions: { approved: 'implement', rejected: 'plan-quality' },
      }),
    ])
    const state = createChainState(plan)
    state.budget = createBudgetState(plan.budgetPolicy)

    expect(state.steps.map((step) => step.stepId)).toEqual(['plan-quality', 'red-team-plan', 'plan-gate'])

    const result = advanceChainState({
      state,
      plan,
      stepId: 'plan-quality',
      outcome: 'success',
      output: { summary: 'quality report emitted' },
    })

    expect(result.nextStep?.stepId).toBe('red-team-plan')
    expect(result.nextStep?.stepId).not.toBe('plan-gate')
  })
})
