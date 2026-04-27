import crypto from 'node:crypto'
import { evaluateBudgetHealth, type AggregateBudgetInput, updateBudget } from './budget-tracker.js'
import type {
  BudgetEvaluation,
  BudgetPolicy,
  BudgetState,
  StepUsage,
  StructuredError,
  TeamDefinition,
  TeamState,
  TeamTaskState,
} from './types.js'

export interface CreateTeamStateInput {
  definition: TeamDefinition
  task: string
  policy: BudgetPolicy
  budget: BudgetState
  createdAt?: string
}

export interface AssignTeamTaskInput {
  state: TeamState
  taskId: string
  assignee: string
  claim?: boolean
  now?: string
}

export interface CompleteTeamTaskInput {
  state: TeamState
  taskId: string
  outcome: 'success' | 'failure'
  result?: Record<string, unknown>
  usage?: StepUsage
  error?: StructuredError
  policy: BudgetPolicy
  now?: string
}

export interface CompleteTeamTaskResult {
  state: TeamState
  evaluation: BudgetEvaluation
}

export function createTeamState(input: CreateTeamStateInput): TeamState {
  const now = input.createdAt ?? new Date().toISOString()
  const memberTasks = input.definition.parallel.map<TeamTaskState>((member, index) => ({
    taskId: `${input.definition.name}:${member.role}`,
    kind: 'member',
    role: member.role,
    agent: member.agent,
    skills: member.skills,
    focus: member.focus,
    state: 'pending',
    order: index,
    dependsOn: [],
    usage: {},
  }))
  const synthesisTaskId = `${input.definition.name}:synthesize`

  const synthesisTask: TeamTaskState = {
    taskId: synthesisTaskId,
    kind: 'synthesize',
    role: 'synthesizer',
    agent: input.definition.synthesize.agent,
    skills: [],
    focus: input.definition.synthesize.description,
    state: 'blocked',
    order: memberTasks.length,
    dependsOn: memberTasks.map((task) => task.taskId),
    usage: {},
  }

  return {
    teamId: crypto.randomUUID(),
    definitionName: input.definition.name,
    definitionVersion: input.definition.version ?? '1.0.0',
    state: 'running',
    task: input.task,
    tasks: [...memberTasks, synthesisTask],
    readyTaskIds: memberTasks.map((task) => task.taskId),
    synthesisTaskId,
    budgetPolicy: input.policy,
    budget: input.budget,
    createdAt: now,
    updatedAt: now,
  }
}

export function assignTeamTask(input: AssignTeamTaskInput): { state: TeamState } {
  const now = input.now ?? new Date().toISOString()
  const state = structuredClone(input.state)
  const task = requireTask(state, input.taskId)

  if (!isAssignable(task)) {
    throw new Error(`Task "${task.taskId}" cannot be assigned from state "${task.state}".`)
  }

  task.assignee = input.assignee
  task.assignedAt = now
  if (input.claim) {
    task.state = 'claimed'
    task.claimedBy = input.assignee
    task.claimedAt = now
  } else {
    task.state = 'assigned'
  }

  state.updatedAt = now
  state.readyTaskIds = computeReadyTaskIds(state.tasks)
  return { state }
}

export function completeTeamTask(input: CompleteTeamTaskInput): CompleteTeamTaskResult {
  const now = input.now ?? new Date().toISOString()
  const state = structuredClone(input.state)
  const task = requireTask(state, input.taskId)

  if (!isCompletable(task)) {
    throw new Error(`Task "${task.taskId}" cannot be completed from state "${task.state}".`)
  }

  task.state = input.outcome === 'success' ? 'completed' : 'failed'
  task.completedAt = now
  if (input.result) {
    task.result = input.result
  } else {
    delete task.result
  }
  task.usage = mergeUsage(task.usage, input.usage)
  if (input.error) {
    task.error = input.error
  } else {
    delete task.error
  }

  const budgetUpdate = updateBudget({
    budget: state.budget,
    policy: input.policy,
    stepId: task.taskId,
    ...(input.usage ? { usage: input.usage } : {}),
    now,
  })
  state.budget = budgetUpdate.budget

  if (task.kind === 'synthesize' && input.outcome === 'success') {
    state.state = 'completed'
    if (input.result) {
      state.summary = input.result
    } else {
      delete state.summary
    }
  } else if (input.outcome === 'failure') {
    state.state = 'failed'
  } else {
    updateSynthesisReadiness(state, now)
  }

  state.readyTaskIds = computeReadyTaskIds(state.tasks)
  state.updatedAt = now

  return {
    state,
    evaluation: evaluateBudgetHealth(state.budget, input.policy),
  }
}

function updateSynthesisReadiness(state: TeamState, now: string): void {
  if (state.state === 'failed') {
    const synthesisTask = requireTask(state, state.synthesisTaskId)
    synthesisTask.state = 'blocked'
    return
  }

  const memberTasks = state.tasks.filter((task) => task.kind === 'member')
  if (memberTasks.every((task) => task.state === 'completed')) {
    const synthesisTask = requireTask(state, state.synthesisTaskId)
    if (synthesisTask.state === 'blocked') {
      synthesisTask.state = 'pending'
    }
    state.state = 'synthesizing'
    state.updatedAt = now
  }
}

function computeReadyTaskIds(tasks: TeamTaskState[]): string[] {
  return tasks
    .filter((task) => task.state === 'pending')
    .filter((task) => task.dependsOn.every((dependency) => tasks.find((candidate) => candidate.taskId === dependency)?.state === 'completed'))
    .map((task) => task.taskId)
}

function requireTask(state: TeamState, taskId: string): TeamTaskState {
  const task = state.tasks.find((entry) => entry.taskId === taskId)
  if (!task) throw new Error(`Unknown team task: ${taskId}`)
  return task
}

function isAssignable(task: TeamTaskState): boolean {
  return task.state === 'pending' || task.state === 'assigned'
}

function isCompletable(task: TeamTaskState): boolean {
  return task.state === 'pending' || task.state === 'assigned' || task.state === 'claimed'
}

function mergeUsage(existing: StepUsage, incoming: StepUsage | undefined): StepUsage {
  return {
    inputTokens: (existing.inputTokens ?? 0) + (incoming?.inputTokens ?? 0),
    outputTokens: (existing.outputTokens ?? 0) + (incoming?.outputTokens ?? 0),
    totalTokens: (existing.totalTokens ?? 0) + (incoming?.totalTokens ?? 0),
    costUsd: (existing.costUsd ?? 0) + (incoming?.costUsd ?? 0),
    wallClockMs: (existing.wallClockMs ?? 0) + (incoming?.wallClockMs ?? 0),
  }
}
