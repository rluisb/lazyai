import crypto from 'node:crypto'
import type {
  AdvanceChainResult,
  BudgetDimensionState,
  BudgetState,
  ChainState,
  ChainStepStatus,
  CompiledStepPlan,
  ExecutionPlan,
  GateState,
  RecoveryAction,
  StepLifecycleState,
  StepState,
  StructuredError,
} from './types.js'

export interface MachineAdvanceInput {
  state: ChainState
  plan: ExecutionPlan
  stepId: string
  outcome: string
  output?: Record<string, unknown>
  validationError?: StructuredError
}

export interface MachineAdvanceResult extends AdvanceChainResult {
  stateSnapshot: ChainState
}

export function createChainState(plan: ExecutionPlan): ChainState {
  const steps = (plan.compiledSteps ?? []).map<StepState>((step, index) => ({
    stepId: step.id,
    order: index,
    agent: step.agent,
    skills: step.skills,
    stepType: step.stepType,
    ...(step.domainSkill ? { domainSkill: step.domainSkill } : {}),
    ...(step.modeSkill ? { modeSkill: step.modeSkill } : {}),
    state: step.id === plan.entrypoint ? 'running' : 'pending',
    attempts: step.id === plan.entrypoint ? 1 : 0,
    maxRetries: getMaxRetries(step),
    ...(step.id === plan.entrypoint ? { startedAt: plan.createdAt } : {}),
    usage: {},
  }))

  return {
    chainId: crypto.randomUUID(),
    definitionName: plan.definition.name,
    definitionVersion: plan.definition.version,
    executionPlanId: plan.id,
    state: 'running',
    task: plan.task,
    currentStepId: plan.entrypoint,
    entryStepId: plan.entrypoint,
    steps,
    completedStepIds: [],
    budget: createEmptyBudgetPlaceholder(plan),
    createdAt: plan.createdAt,
    updatedAt: plan.createdAt,
  }
}

export function advanceChainState(input: MachineAdvanceInput): MachineAdvanceResult {
  const now = new Date().toISOString()
  const state = cloneChainState(input.state)
  const step = requireStep(state, input.stepId)
  const compiledStep = requireCompiledStep(input.plan, input.stepId)

  if (state.state === 'gated' && step.gate?.status === 'pending') {
    return applyGateDecision(state, input.plan, compiledStep, step, input.outcome, now)
  }

  if (step.state !== 'running' && step.state !== 'escalated' && step.state !== 'retrying') {
    throw new Error(`Step "${step.stepId}" is not active.`)
  }

  if (input.validationError) {
    step.state = 'failed'
    step.error = input.validationError
    step.lastOutcome = input.outcome
    state.updatedAt = now

    return {
      state: state.state,
      nextStep: toChainStepStatus(input.plan, step.stepId, step.state),
      gate: null,
      recovery: input.validationError.suggestedRecovery,
      budget: state.budget,
      error: input.validationError,
      stateSnapshot: state,
    }
  }

  if (input.output) {
    step.output = input.output
    step.outputValid = true
  }

  if (compiledStep.gate && input.outcome === 'success') {
    step.state = 'completed'
    step.completedAt = now
    step.lastOutcome = input.outcome
    step.gate = buildPendingGate(compiledStep.gate, compiledStep.id)
    markCompleted(state, step.stepId)
    state.state = 'gated'
    state.updatedAt = now

    return {
      state: state.state,
      nextStep: toChainStepStatus(input.plan, step.stepId, step.state),
      gate: step.gate,
      recovery: null,
      budget: state.budget,
      stateSnapshot: state,
    }
  }

  if (input.outcome === 'failure') {
    return handleFailureTransition(state, compiledStep, step, now, input.plan)
  }

  const transition = compiledStep.transitions[input.outcome]
  if (!transition || typeof transition !== 'string') {
    throw new Error(`Outcome "${input.outcome}" is not valid for step "${input.stepId}".`)
  }

  step.state = 'completed'
  step.completedAt = now
  step.lastOutcome = input.outcome
  markCompleted(state, step.stepId)

  return moveToTarget(state, input.plan, transition, now)
}

export function retryChainStep(
  state: ChainState,
  plan: ExecutionPlan,
  stepId: string,
): { state: ChainState; attemptsRemaining: number } {
  const next = cloneChainState(state)
  const step = requireStep(next, stepId)
  requireCompiledStep(plan, stepId)

  if (step.state !== 'failed') {
    throw new Error(`Step "${stepId}" is not eligible for retry.`)
  }

  const attemptsRemaining = getAttemptsRemaining(step)
  if (attemptsRemaining <= 0) {
    throw new Error(`Step "${stepId}" has no retries remaining.`)
  }

  step.state = 'running'
  step.attempts += 1
  step.startedAt = new Date().toISOString()
  delete step.completedAt
  delete step.error
  delete step.gate
  delete step.output
  delete step.outputValid
  step.lastOutcome = 'retry'
  next.state = 'running'
  next.currentStepId = stepId
  next.updatedAt = step.startedAt

  return {
    state: next,
    attemptsRemaining: getAttemptsRemaining(step),
  }
}

export function escalateChainStep(
  state: ChainState,
  plan: ExecutionPlan,
  stepId: string,
  targetAgent: string,
  domainSkill?: string,
  modeSkill?: string,
): { state: ChainState; plan: ExecutionPlan } {
  const nextState = cloneChainState(state)
  const nextPlan = structuredClone(plan)
  const step = requireStep(nextState, stepId)
  const compiled = requireCompiledStep(nextPlan, stepId)
  const now = new Date().toISOString()

  step.state = 'running'
  step.agent = targetAgent
  step.attempts += 1
  step.startedAt = now
  delete step.completedAt
  delete step.error
  delete step.gate
  step.lastOutcome = 'escalated'
  if (domainSkill) step.domainSkill = domainSkill
  if (modeSkill) step.modeSkill = modeSkill

  compiled.agent = targetAgent
  if (domainSkill) compiled.domainSkill = domainSkill
  if (modeSkill) compiled.modeSkill = modeSkill
  nextState.state = 'running'
  nextState.currentStepId = stepId
  nextState.updatedAt = now

  return {
    state: nextState,
    plan: nextPlan,
  }
}

function handleFailureTransition(
  state: ChainState,
  compiledStep: CompiledStepPlan,
  step: StepState,
  now: string,
  plan: ExecutionPlan,
): MachineAdvanceResult {
  step.state = 'failed'
  step.lastOutcome = 'failure'

  const transition = compiledStep.transitions.failure
  if (transition && typeof transition !== 'string') {
    if (getAttemptsRemaining(step) > 0) {
      const recovery: RecoveryAction = {
        type: 'retry',
        maxAttempts: step.maxRetries,
        guidance: `Retry step "${step.stepId}" before routing to "${transition.then}".`,
      }

      state.updatedAt = now
      return {
        state: state.state,
        nextStep: toChainStepStatus(plan, step.stepId, step.state),
        gate: null,
        recovery,
        budget: state.budget,
        stateSnapshot: state,
      }
    }

    return moveToTarget(state, plan, transition.then, now)
  }

  if (typeof transition === 'string') {
    return moveToTarget(state, plan, transition, now)
  }

  state.updatedAt = now
  return {
    state: state.state,
    nextStep: toChainStepStatus(plan, step.stepId, step.state),
    gate: null,
    recovery: null,
    budget: state.budget,
    stateSnapshot: state,
  }
}

function applyGateDecision(
  state: ChainState,
  plan: ExecutionPlan,
  compiledStep: CompiledStepPlan,
  step: StepState,
  outcome: string,
  now: string,
): MachineAdvanceResult {
  if (!step.gate) {
    throw new Error(`Step "${step.stepId}" is not waiting on a gate.`)
  }

  if (outcome !== 'approved' && outcome !== 'rejected') {
    throw new Error('Gate outcome must be "approved" or "rejected".')
  }

  step.gate.status = outcome
  step.gate.decidedAt = now
  const transition = compiledStep.transitions[outcome]
  if (!transition || typeof transition !== 'string') {
    throw new Error(`Gate outcome "${outcome}" is not valid for step "${step.stepId}".`)
  }

  return moveToTarget(state, plan, transition, now)
}

function moveToTarget(state: ChainState, plan: ExecutionPlan, target: string, now: string): MachineAdvanceResult {
  if (target === 'done') {
    state.state = 'completed'
    delete state.currentStepId
    state.updatedAt = now

    return {
      state: state.state,
      nextStep: null,
      gate: null,
      recovery: null,
      budget: state.budget,
      stateSnapshot: state,
    }
  }

  if (target === 'handoff') {
    state.state = 'handoff'
    delete state.currentStepId
    state.updatedAt = now

    return {
      state: state.state,
      nextStep: null,
      gate: null,
      recovery: { type: 'handoff', summary: 'Definition requested a handoff transition.' },
      budget: state.budget,
      stateSnapshot: state,
    }
  }

  if (target === 'abandon') {
    state.state = 'abandoned'
    delete state.currentStepId
    state.updatedAt = now

    return {
      state: state.state,
      nextStep: null,
      gate: null,
      recovery: { type: 'abort', reason: 'Definition requested an abandon transition.' },
      budget: state.budget,
      stateSnapshot: state,
    }
  }

  const nextRuntimeStep = requireStep(state, target)
  nextRuntimeStep.state = 'running'
  nextRuntimeStep.startedAt = now
  nextRuntimeStep.attempts = nextRuntimeStep.attempts === 0 ? 1 : nextRuntimeStep.attempts + 1
  delete nextRuntimeStep.error
  delete nextRuntimeStep.gate
  state.currentStepId = target
  state.state = 'running'
  state.updatedAt = now

  return {
    state: state.state,
    nextStep: toChainStepStatus(plan, target, nextRuntimeStep.state),
    gate: null,
    recovery: null,
    budget: state.budget,
    stateSnapshot: state,
  }
}

function toChainStepStatus(
  plan: ExecutionPlan,
  stepId: string | undefined,
  state: StepLifecycleState,
): ChainStepStatus | null {
  if (!stepId) return null

  const compiled = requireCompiledStep(plan, stepId)
  return {
    stepId: compiled.id,
    agent: compiled.agent,
    skills: compiled.skills,
    stepType: compiled.stepType,
    state,
    model: compiled.model,
    tools: compiled.allowedTools,
    instructions: compiled.instructions,
    outputContract: compiled.outputContract,
    ...(compiled.gate ? { gate: buildPendingGate(compiled.gate, compiled.id) } : {}),
    composedAgent: compiled.composedAgent,
  }
}

function buildPendingGate(type: NonNullable<CompiledStepPlan['gate']>, stepId: string): GateState {
  return {
    type,
    prompt: `Awaiting ${type} for step "${stepId}".`,
    status: 'pending',
  }
}

function getMaxRetries(step: CompiledStepPlan): number {
  const failure = step.transitions.failure
  return typeof failure === 'object' ? failure.retry : 0
}

function getAttemptsRemaining(step: StepState): number {
  return step.maxRetries - Math.max(step.attempts - 1, 0)
}

function requireStep(state: ChainState, stepId: string): StepState {
  const step = state.steps.find((entry) => entry.stepId === stepId)
  if (!step) throw new Error(`Unknown step state: ${stepId}`)
  return step
}

function requireCompiledStep(plan: ExecutionPlan, stepId: string): CompiledStepPlan {
  const step = plan.compiledSteps?.find((entry) => entry.id === stepId)
  if (!step) throw new Error(`Unknown compiled step: ${stepId}`)
  return step
}

function createEmptyBudgetPlaceholder(plan: ExecutionPlan): BudgetState {
  return {
    policyId: plan.budgetPolicy.id,
    scope: 'chain',
    tokens: buildDimension(plan.budgetPolicy.tokens?.limit),
    costUsd: buildDimension(plan.budgetPolicy.costUsd?.limit),
    wallClockMs: buildDimension(plan.budgetPolicy.wallClockMs?.limit),
    retries: buildDimension(plan.budgetPolicy.retries?.limit),
    byStep: {},
    lastUpdatedAt: plan.createdAt,
  }
}

function buildDimension(limit: number | undefined): BudgetDimensionState {
  return {
    consumed: 0,
    warningTriggered: false,
    pausedAtLimit: false,
    ...(typeof limit === 'number' ? { limit, remaining: limit } : {}),
  }
}

function markCompleted(state: ChainState, stepId: string): void {
  if (!state.completedStepIds.includes(stepId)) {
    state.completedStepIds.push(stepId)
  }
}

function cloneChainState(state: ChainState): ChainState {
  return structuredClone(state)
}
