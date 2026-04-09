import type {
  BudgetDimensionState,
  BudgetEvaluation,
  BudgetHealth,
  BudgetPolicy,
  BudgetState,
  StepUsage,
} from './types.js'

export interface BudgetUpdateInput {
  budget: BudgetState
  policy: BudgetPolicy
  stepId?: string
  usage?: StepUsage
  retryIncrement?: number
  now?: string
}

export interface AggregateBudgetInput {
  policy: BudgetPolicy
  scope: BudgetState['scope']
  children: Array<{
    childId: string
    budget: BudgetState
  }>
  now?: string
}

export function createBudgetState(policy: BudgetPolicy): BudgetState {
  const now = new Date().toISOString()
  return {
    policyId: policy.id,
    scope: policy.scope,
    tokens: createDimensionState(policy.tokens),
    costUsd: createDimensionState(policy.costUsd),
    wallClockMs: createDimensionState(policy.wallClockMs),
    retries: createDimensionState(policy.retries),
    byStep: {},
    lastUpdatedAt: now,
  }
}

export function updateBudget(input: BudgetUpdateInput): { budget: BudgetState; evaluation: BudgetEvaluation } {
  const usage = input.usage ?? {}
  const now = input.now ?? new Date().toISOString()
  const totalTokens = usage.totalTokens ?? (usage.inputTokens ?? 0) + (usage.outputTokens ?? 0)

  const byStep = {
    ...input.budget.byStep,
  }

  if (input.stepId) {
    byStep[input.stepId] = mergeUsage(byStep[input.stepId], usage)
  }

  const budget: BudgetState = {
    ...input.budget,
    tokens: applyDimensionUpdate(input.budget.tokens, input.policy.tokens, totalTokens),
    costUsd: applyDimensionUpdate(input.budget.costUsd, input.policy.costUsd, usage.costUsd ?? 0),
    wallClockMs: applyDimensionUpdate(input.budget.wallClockMs, input.policy.wallClockMs, usage.wallClockMs ?? 0),
    retries: applyDimensionUpdate(input.budget.retries, input.policy.retries, input.retryIncrement ?? 0),
    byStep,
    lastUpdatedAt: now,
  }

  return {
    budget,
    evaluation: evaluateBudgetHealth(budget, input.policy),
  }
}

export function evaluateBudgetHealth(budget: BudgetState, policy: BudgetPolicy): BudgetEvaluation {
  const dimensions = {
    tokens: evaluateDimension(budget.tokens, policy.tokens),
    costUsd: evaluateDimension(budget.costUsd, policy.costUsd),
    wallClockMs: evaluateDimension(budget.wallClockMs, policy.wallClockMs),
    retries: evaluateDimension(budget.retries, policy.retries),
  }

  const values = Object.values(dimensions)
  const overall = values.includes('limit_reached')
    ? 'limit_reached'
    : values.includes('warning')
      ? 'warning'
      : 'ok'

  const recommendedAction = values.includes('limit_reached')
    ? policy.defaultActionOnLimit === 'abort'
      ? 'abort'
      : 'pause'
    : values.includes('warning')
      ? 'warn'
      : 'continue'

  return {
    overall,
    dimensions,
    recommendedAction,
    shouldPause: recommendedAction === 'pause' || recommendedAction === 'abort',
  }
}

export function aggregateBudgets(input: AggregateBudgetInput): { budget: BudgetState; evaluation: BudgetEvaluation } {
  const now = input.now ?? new Date().toISOString()
  const aggregated = createBudgetState({
    ...input.policy,
    scope: input.scope,
  })

  aggregated.tokens = mergeDimensionStates(input.children.map((child) => child.budget.tokens), input.policy.tokens)
  aggregated.costUsd = mergeDimensionStates(input.children.map((child) => child.budget.costUsd), input.policy.costUsd)
  aggregated.wallClockMs = mergeDimensionStates(input.children.map((child) => child.budget.wallClockMs), input.policy.wallClockMs)
  aggregated.retries = mergeDimensionStates(input.children.map((child) => child.budget.retries), input.policy.retries)
  aggregated.byStep = Object.fromEntries(
    input.children.flatMap((child) =>
      Object.entries(child.budget.byStep).map(([stepId, usage]) => [`${child.childId}:${stepId}`, usage] as const),
    ),
  )
  aggregated.lastUpdatedAt = now

  return {
    budget: aggregated,
    evaluation: evaluateBudgetHealth(aggregated, input.policy),
  }
}

function createDimensionState(threshold?: BudgetPolicy['tokens']): BudgetDimensionState {
  return {
    consumed: 0,
    warningTriggered: false,
    pausedAtLimit: false,
    ...(typeof threshold?.limit === 'number'
      ? {
          limit: threshold.limit,
          remaining: threshold.limit,
        }
      : {}),
  }
}

function applyDimensionUpdate(
  state: BudgetDimensionState,
  threshold: BudgetPolicy['tokens'],
  increment: number,
): BudgetDimensionState {
  const consumed = state.consumed + increment
  const warnAt = resolveThresholdValue(threshold?.limit, threshold?.warnAt)
  const pauseAt = resolveThresholdValue(threshold?.limit, threshold?.pauseAt)
  const limit = threshold?.limit

  return {
    consumed,
    warningTriggered: state.warningTriggered || (typeof warnAt === 'number' && consumed >= warnAt),
    pausedAtLimit:
      state.pausedAtLimit ||
      (typeof pauseAt === 'number' && consumed >= pauseAt) ||
      Boolean(threshold?.hardStop && typeof limit === 'number' && consumed >= limit),
    ...(typeof limit === 'number'
      ? {
          limit,
          remaining: Math.max(limit - consumed, 0),
        }
      : {}),
  }
}

function resolveThresholdValue(limit: number | undefined, value: number | undefined): number | undefined {
  if (value === undefined) return undefined
  if (limit !== undefined && value > 0 && value <= 1) return limit * value
  return value
}

function evaluateDimension(state: BudgetDimensionState, threshold?: BudgetPolicy['tokens']): BudgetHealth {
  if (state.pausedAtLimit) return 'limit_reached'
  if (typeof threshold?.limit === 'number' && state.consumed >= threshold.limit) return 'limit_reached'
  if (state.warningTriggered) return 'warning'
  return 'ok'
}

function mergeUsage(existing: StepUsage | undefined, incoming: StepUsage): StepUsage {
  return {
    inputTokens: (existing?.inputTokens ?? 0) + (incoming.inputTokens ?? 0),
    outputTokens: (existing?.outputTokens ?? 0) + (incoming.outputTokens ?? 0),
    totalTokens: (existing?.totalTokens ?? 0) + (incoming.totalTokens ?? 0),
    costUsd: (existing?.costUsd ?? 0) + (incoming.costUsd ?? 0),
    wallClockMs: (existing?.wallClockMs ?? 0) + (incoming.wallClockMs ?? 0),
  }
}

function mergeDimensionStates(
  states: BudgetDimensionState[],
  threshold: BudgetPolicy['tokens'],
): BudgetDimensionState {
  const consumed = states.reduce((total, state) => total + state.consumed, 0)
  const limit = threshold?.limit

  return {
    consumed,
    warningTriggered: states.some((state) => state.warningTriggered),
    pausedAtLimit: states.some((state) => state.pausedAtLimit),
    ...(typeof limit === 'number'
      ? {
          limit,
          remaining: Math.max(limit - consumed, 0),
        }
      : {}),
  }
}
