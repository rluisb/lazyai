import { describe, expect, it } from 'vitest'
import { createBudgetState } from '../budget-tracker.js'
import { advanceChainState, createChainState, retryChainStep } from '../chain-machine.js'
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

function retryThen(target: string): Extract<CompiledStepPlan['transitions'][string], { retry: number; then: string }> {
  const transition = { retry: 1 } as Extract<CompiledStepPlan['transitions'][string], { retry: number; then: string }>
  // biome-ignore lint/suspicious/noThenProperty: Chain retry transitions intentionally use the catalog's retry/then schema.
  transition.then = target
  return transition
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
        transitions: { success: 'step-2', failure: retryThen('done') },
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

function makePlanWithSteps(compiledSteps: CompiledStepPlan[], entrypoint?: string): ExecutionPlan {
  const resolvedEntrypoint = entrypoint ?? compiledSteps[0]?.id
  if (!resolvedEntrypoint) throw new Error('makePlanWithSteps requires at least one compiled step or an entrypoint')

  return {
    ...makePlan(),
    entrypoint: resolvedEntrypoint,
    compiledSteps,
  }
}

interface MergedGateReportFixture {
  schemaVersion: string
  summary: {
    planVerdict: string
    redTeamStatus: string
    blockingCount: number
    warningCount: number
  }
  planQuality: {
    findings: Array<{
      location: Record<string, unknown>
    }>
  }
  adversarialReview: Record<string, unknown> | null
}

const planQualityPassReport = {
  schemaVersion: 'plan-quality-report/v1',
  verdict: 'pass',
  findings: [],
  checkedAgainst: {
    spec: 'specs/features/001-ai-techniques-integration/spec.md',
    plan: 'specs/features/001-ai-techniques-integration/plan.md',
    research: 'specs/features/001-ai-techniques-integration/research.md',
    tasks: null,
  },
}

const planQualityFailReport = {
  schemaVersion: 'plan-quality-report/v1',
  verdict: 'fail',
  findings: [
    {
      rule: 'R1',
      severity: 'fail',
      message: 'AC-D6-001 lacks confident plan.md coverage; tasks.md is not inspected before task generation.',
      location: {
        file: 'specs/features/001-ai-techniques-integration/plan.md',
        section: 'D6 — Plan Validation',
        lineStart: 256,
        lineEnd: 267,
      },
    },
    {
      rule: 'R2',
      severity: 'warn',
      message: 'A phase exit criterion needs clearer observable evidence.',
      location: {
        file: 'specs/features/001-ai-techniques-integration/plan.md',
        section: 'Phases & Milestones',
        lineStart: 202,
        lineEnd: 205,
      },
    },
  ],
  checkedAgainst: {
    spec: 'specs/features/001-ai-techniques-integration/spec.md',
    plan: 'specs/features/001-ai-techniques-integration/plan.md',
    research: 'specs/features/001-ai-techniques-integration/research.md',
    tasks: null,
  },
}

const redTeamOkReport = {
  schemaVersion: 'red-team-plan-report/v1',
  status: 'ok',
  findings: [
    {
      category: 'security',
      severity: 'high',
      message: 'Provider-backed review needs a visible secret-redaction assumption.',
      recommendation: 'Ask the human approver to confirm no secrets are included before provider review.',
      location: {
        file: 'specs/features/001-ai-techniques-integration/spec.md',
        section: 'Out of Scope',
        lineStart: 243,
        lineEnd: 249,
      },
    },
    {
      category: 'operational',
      severity: 'medium',
      message: 'Active chain rollback policy must remain visible to the gate reviewer.',
      recommendation: 'Keep drain/no-op compatibility guidance in the approval packet.',
      location: {
        file: 'specs/features/001-ai-techniques-integration/plan.md',
        section: 'Rollback and Release Policy',
        lineStart: 357,
        lineEnd: 364,
      },
    },
  ],
}

const redTeamSoftFailReport = {
  schemaVersion: 'red-team-plan-report/v1',
  status: 'soft_fail',
  findings: [
    {
      category: 'operational',
      severity: 'medium',
      message: 'Red-team provider/API outage prevented adversarial design review.',
      recommendation: 'Surface this warning in plan-gate and proceed to the human approval gate instead of halting.',
      location: {
        file: 'specs/features/001-ai-techniques-integration/plan.md',
        section: 'D17 — Adversarial Self-Play During Design',
        lineStart: null,
        lineEnd: null,
      },
    },
  ],
}

function makeW1BFeaturePlan(adversarialDesign: boolean): ExecutionPlan {
  const planQualityTarget = adversarialDesign ? 'red-team-plan' : 'plan-gate'
  return makePlanWithSteps([
    makeCompiledStep({
      id: 'plan',
      agent: 'planner',
      skills: ['plan'],
      stepType: 'plan',
      transitions: { success: 'plan-quality' },
    }),
    makeCompiledStep({
      id: 'plan-quality',
      agent: 'planner',
      skills: ['plan'],
      stepType: 'plan',
      transitions: { success: planQualityTarget, pass: planQualityTarget, warn: planQualityTarget, fail: planQualityTarget },
    }),
    ...(adversarialDesign
      ? [
          makeCompiledStep({
            id: 'red-team-plan',
            agent: 'red-team',
            skills: ['red-team-plan'],
            stepType: 'plan',
            transitions: { success: 'plan-gate', soft_fail: 'plan-gate', failure: 'plan-gate' },
          }),
        ]
      : []),
    makeCompiledStep({
      id: 'plan-gate',
      agent: 'planner',
      skills: [],
      stepType: 'plan',
      gate: 'user_approval',
      transitions: { approved: 'implement', rejected: 'plan' },
    }),
    makeCompiledStep({
      id: 'implement',
      agent: 'builder',
      skills: ['implement'],
      stepType: 'implement',
      transitions: { success: 'done' },
    }),
  ], 'plan')
}

function createRunningW1BState(plan: ExecutionPlan) {
  const state = createChainState(plan)
  state.budget = createBudgetState(plan.budgetPolicy)
  return state
}

function runW1BToGate(input: {
  adversarialDesign: boolean
  planQualityOutcome?: string
  planQualityReport?: Record<string, unknown>
  redTeamOutcome?: string
  redTeamReport?: Record<string, unknown>
}) {
  const plan = makeW1BFeaturePlan(input.adversarialDesign)
  const state = createRunningW1BState(plan)

  const afterPlan = advanceChainState({
    state,
    plan,
    stepId: 'plan',
    outcome: 'success',
    output: { summary: 'plan.md emitted' },
  })

  const afterQuality = advanceChainState({
    state: afterPlan.stateSnapshot,
    plan,
    stepId: 'plan-quality',
    outcome: input.planQualityOutcome ?? 'pass',
    output: { summary: 'PlanQualityReport emitted', planQualityReport: input.planQualityReport ?? planQualityPassReport },
  })

  const beforeGate = input.adversarialDesign
    ? advanceChainState({
        state: afterQuality.stateSnapshot,
        plan,
        stepId: 'red-team-plan',
        outcome: input.redTeamOutcome ?? 'success',
        output: { summary: 'RedTeamPlanReport emitted', redTeamPlanReport: input.redTeamReport ?? redTeamOkReport },
      })
    : afterQuality

  return advanceChainState({
    state: beforeGate.stateSnapshot,
    plan,
    stepId: 'plan-gate',
    outcome: 'success',
    output: { summary: 'approval packet ready' },
  })
}

function extractMergedGateReport(prompt: string): MergedGateReportFixture {
  const match = prompt.match(/```json\n([\s\S]*?)\n```/)
  const json = match?.[1]
  if (!match) throw new Error(`expected merged gate report JSON block in prompt: ${prompt}`)
  if (!json) throw new Error(`expected merged gate report JSON payload in prompt: ${prompt}`)
  return JSON.parse(json) as MergedGateReportFixture
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
    const firstStep = state.steps[0]
    if (!firstStep) throw new Error('expected first runtime step')
    firstStep.state = 'failed'

    const retried = retryChainStep(state, plan, 'step-1')
    expect(retried.state.currentStepId).toBe('step-1')
    expect(retried.state.steps[0]?.state).toBe('running')
    expect(retried.attemptsRemaining).toBe(0)
  })

  it('stops at a sequential user approval gate and resumes through approved or rejected outcomes', () => {
    const plan = makePlanWithSteps([
      makeCompiledStep({
        id: 'plan',
        agent: 'planner',
        skills: ['plan'],
        stepType: 'plan',
        transitions: { success: 'plan-quality' },
      }),
      makeCompiledStep({
        id: 'plan-quality',
        agent: 'planner',
        skills: ['plan'],
        stepType: 'plan',
        transitions: { success: 'plan-gate', pass: 'plan-gate', warn: 'plan-gate', fail: 'plan-gate' },
      }),
      makeCompiledStep({
        id: 'plan-gate',
        agent: 'planner',
        skills: [],
        stepType: 'plan',
        gate: 'user_approval',
        transitions: { approved: 'implement', rejected: 'plan' },
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

    const afterPlan = advanceChainState({
      state,
      plan,
      stepId: 'plan',
      outcome: 'success',
      output: { summary: 'plan emitted' },
    })
    expect(afterPlan.state).toBe('running')
    expect(afterPlan.nextStep?.stepId).toBe('plan-quality')

    const afterQuality = advanceChainState({
      state: afterPlan.stateSnapshot,
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
    expect(rejected.nextStep?.stepId).toBe('plan')
  })

  it.each(['success', 'pass', 'warn', 'fail'])('routes plan-quality %s outcomes to the plan gate without auto-looping', (outcome) => {
    const plan = makePlanWithSteps([
      makeCompiledStep({
        id: 'plan',
        agent: 'planner',
        skills: ['plan'],
        stepType: 'plan',
        transitions: { success: 'plan-quality' },
      }),
      makeCompiledStep({
        id: 'plan-quality',
        agent: 'planner',
        skills: ['plan'],
        stepType: 'plan',
        transitions: { success: 'plan-gate', pass: 'plan-gate', warn: 'plan-gate', fail: 'plan-gate' },
      }),
      makeCompiledStep({
        id: 'plan-gate',
        agent: 'planner',
        skills: [],
        stepType: 'plan',
        gate: 'user_approval',
        transitions: { approved: 'implement', rejected: 'plan' },
      }),
      makeCompiledStep({
        id: 'implement',
        agent: 'builder',
        skills: ['implement'],
        stepType: 'implement',
        transitions: { success: 'done' },
      }),
    ], 'plan-quality')

    const state = createChainState(plan)
    state.budget = createBudgetState(plan.budgetPolicy)

    const result = advanceChainState({
      state,
      plan,
      stepId: 'plan-quality',
      outcome,
      output: { summary: `${outcome} quality report emitted` },
    })

    expect(result.state).toBe('running')
    expect(result.nextStep?.stepId).toBe('plan-gate')
    expect(result.nextStep?.stepId).not.toBe('plan')
    expect(result.nextStep?.stepId).not.toBe('implement')
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

  it('gates the base D6 path only after plan-quality and displays PlanQualityReport evidence', () => {
    const gated = runW1BToGate({ adversarialDesign: false, planQualityOutcome: 'pass', planQualityReport: planQualityPassReport })

    expect(gated.state).toBe('gated')
    expect(gated.gate?.type).toBe('user_approval')
    expect(gated.nextStep?.stepId).toBe('plan-gate')
    expect(gated.gate?.prompt).toContain('PlanQualityReport')
    expect(gated.gate?.prompt).toContain('plan-quality-report/v1')

    const report = extractMergedGateReport(gated.gate?.prompt ?? '')
    expect(report).toMatchObject({
      schemaVersion: 'plan-gate-report/v1',
      summary: { planVerdict: 'pass', redTeamStatus: 'skipped', blockingCount: 0, warningCount: 0 },
      planQuality: planQualityPassReport,
      adversarialReview: null,
    })
  })

  it('gates the adversarial D17 path after red-team-plan and displays both reports', () => {
    const gated = runW1BToGate({ adversarialDesign: true, planQualityOutcome: 'pass', planQualityReport: planQualityPassReport, redTeamOutcome: 'success', redTeamReport: redTeamOkReport })

    expect(gated.state).toBe('gated')
    expect(gated.gate?.prompt).toContain('PlanQualityReport')
    expect(gated.gate?.prompt).toContain('RedTeamPlanReport')

    const report = extractMergedGateReport(gated.gate?.prompt ?? '')
    expect(report).toMatchObject({
      schemaVersion: 'plan-gate-report/v1',
      summary: { planVerdict: 'pass', redTeamStatus: 'ok', blockingCount: 1, warningCount: 1 },
      planQuality: planQualityPassReport,
      adversarialReview: redTeamOkReport,
    })
  })

  it('surfaces a fail plan-quality verdict at the gate without automatic looping', () => {
    const gated = runW1BToGate({ adversarialDesign: false, planQualityOutcome: 'fail', planQualityReport: planQualityFailReport })

    expect(gated.state).toBe('gated')
    expect(gated.nextStep?.stepId).toBe('plan-gate')
    expect(gated.gate?.prompt).toContain('fail')
    expect(gated.gate?.prompt).toContain('AC-D6-001')

    const report = extractMergedGateReport(gated.gate?.prompt ?? '')
    expect(report.summary).toMatchObject({ planVerdict: 'fail', redTeamStatus: 'skipped', blockingCount: 1, warningCount: 1 })
    const firstFinding = report.planQuality.findings[0]
    expect(firstFinding).toBeDefined()
    expect(firstFinding?.location).toMatchObject({
      file: 'specs/features/001-ai-techniques-integration/plan.md',
      lineStart: 256,
      lineEnd: 267,
    })
  })

  it('treats red-team-plan soft_fail as gate-visible evidence instead of a chain halt', () => {
    const gated = runW1BToGate({
      adversarialDesign: true,
      planQualityOutcome: 'pass',
      planQualityReport: planQualityPassReport,
      redTeamOutcome: 'soft_fail',
      redTeamReport: redTeamSoftFailReport,
    })

    expect(gated.state).toBe('gated')
    expect(gated.nextStep?.stepId).toBe('plan-gate')
    expect(gated.gate?.prompt).toContain('soft_fail')
    expect(gated.gate?.prompt).toContain('provider/API outage')

    const report = extractMergedGateReport(gated.gate?.prompt ?? '')
    expect(report).toMatchObject({
      schemaVersion: 'plan-gate-report/v1',
      summary: { planVerdict: 'pass', redTeamStatus: 'soft_fail', blockingCount: 0, warningCount: 1 },
      adversarialReview: redTeamSoftFailReport,
    })
  })
})
