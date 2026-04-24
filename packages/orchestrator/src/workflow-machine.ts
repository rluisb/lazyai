import crypto from 'node:crypto'
import type {
  BudgetState,
  RecoveryAction,
  StructuredError,
  WorkflowAction,
  WorkflowChildRun,
  WorkflowDefinition,
  WorkflowPhaseDefinition,
  WorkflowPhaseState,
  WorkflowRecoveryDecision,
  WorkflowState,
} from './types.js'

const TERMINAL_PHASE_IDS = new Set(['complete', 'completed'])
const HANDOFF_PHASE_IDS = new Set(['handoff'])

export interface CreateWorkflowStateInput {
  definition: WorkflowDefinition
  task: string
  policy: WorkflowState['budgetPolicy']
  budget: BudgetState
  createdAt?: string
  runtime?: WorkflowState['runtime']
}

export interface CreateWorkflowStateResult {
  state: WorkflowState
  nextAction: WorkflowAction | null
}

export interface ApplyWorkflowChildLaunchInput {
  state: WorkflowState
  phaseId: string
  childRun: WorkflowChildRun
  now?: string
}

export interface AdvanceWorkflowStateInput {
  state: WorkflowState
  definition: WorkflowDefinition
  outcome?: string
  childError?: StructuredError
  recovery?: WorkflowRecoveryDecision
  now?: string
}

export interface AdvanceWorkflowStateResult {
  state: WorkflowState
  nextAction: WorkflowAction | null
  recoveryOptions: RecoveryAction[]
}

export function createWorkflowState(input: CreateWorkflowStateInput): CreateWorkflowStateResult {
  const now = input.createdAt ?? new Date().toISOString()
  const phases = input.definition.phases.map<WorkflowPhaseState>((phase) => ({
    phaseId: phase.id,
    kind: phase.kind,
    state: 'pending',
    ...(phase.ref ? { ref: phase.ref } : {}),
    ...(phase.gate ? { gate: phase.gate } : {}),
    ...(phase.prompt ? { prompt: phase.prompt } : {}),
  }))

  const state: WorkflowState = {
    workflowId: crypto.randomUUID(),
    definitionName: input.definition.name,
    definitionVersion: input.definition.version ?? '1.0.0',
    state: 'created',
    task: input.task,
    entryPhaseId: input.definition.entry,
    currentPhaseId: input.definition.entry,
    phases,
    childRuns: [],
    budgetPolicy: input.policy,
    budget: input.budget,
    createdAt: now,
    updatedAt: now,
    ...(input.runtime ? { runtime: input.runtime } : {}),
  }

  const nextAction = enterPhase(state, input.definition, input.definition.entry, now)
  return { state, nextAction }
}

export function applyWorkflowChildLaunch(input: ApplyWorkflowChildLaunchInput): { state: WorkflowState } {
  const now = input.now ?? new Date().toISOString()
  const state = structuredClone(input.state)
  const phase = requirePhaseState(state, input.phaseId)

  phase.childRun = input.childRun
  phase.startedAt = phase.startedAt ?? now
  state.childRuns = [...state.childRuns, input.childRun]
  state.state = 'waiting_on_child'
  state.updatedAt = now

  return { state }
}

export function advanceWorkflowState(input: AdvanceWorkflowStateInput): AdvanceWorkflowStateResult {
  const now = input.now ?? new Date().toISOString()
  const state = structuredClone(input.state)
  const currentPhase = requireCurrentPhase(state)
  const definitionPhase = requireDefinitionPhase(input.definition, currentPhase.phaseId)

  if (state.state === 'awaiting_recovery') {
    if (!input.recovery) {
      throw new Error('Workflow is awaiting an explicit recovery decision.')
    }
    return applyRecoveryDecision(state, input.definition, definitionPhase, input.recovery, now)
  }

  if (definitionPhase.kind === 'gate') {
    if (!input.outcome) throw new Error('Gate phases require an outcome to advance.')
    currentPhase.state = 'completed'
    currentPhase.completedAt = now
    currentPhase.lastOutcome = input.outcome
    const target = definitionPhase.on?.[input.outcome]
    if (!target) throw new Error(`Workflow gate phase "${currentPhase.phaseId}" does not support outcome "${input.outcome}".`)
    return moveToTarget(state, input.definition, target, now)
  }

  if (definitionPhase.kind === 'chain' || definitionPhase.kind === 'team') {
    if (!input.outcome) throw new Error('Child phases require an outcome to advance.')

    if (input.outcome === 'failure') {
      currentPhase.state = 'failed'
      currentPhase.lastOutcome = 'failure'
      currentPhase.completedAt = now
      state.state = 'awaiting_recovery'
      state.updatedAt = now
      state.lastError = input.childError ?? buildGenericChildError(state, definitionPhase)

      return {
        state,
        nextAction: null,
        recoveryOptions: buildRecoveryOptions(definitionPhase, state.lastError),
      }
    }

    currentPhase.state = 'completed'
    currentPhase.completedAt = now
    currentPhase.lastOutcome = input.outcome
    if (currentPhase.childRun) {
      currentPhase.childRun.completedAt = now
      currentPhase.childRun.outcome = input.outcome
      syncChildRun(state, currentPhase.childRun)
    }

    const target = definitionPhase.on?.[input.outcome]
    if (!target) throw new Error(`Workflow phase "${currentPhase.phaseId}" does not support outcome "${input.outcome}".`)
    return moveToTarget(state, input.definition, target, now)
  }

  return {
    state,
    nextAction: null,
    recoveryOptions: [],
  }
}

function applyRecoveryDecision(
  state: WorkflowState,
  definition: WorkflowDefinition,
  currentDefinitionPhase: WorkflowPhaseDefinition,
  recovery: WorkflowRecoveryDecision,
  now: string,
): AdvanceWorkflowStateResult {
  if (recovery.type === 'retry') {
    const currentPhase = requireCurrentPhase(state)
    currentPhase.state = 'waiting_on_child'
    delete currentPhase.completedAt
    currentPhase.lastOutcome = 'retry'
    delete currentPhase.childRun
    state.state = 'waiting_on_child'
    state.updatedAt = now
    delete state.lastError

    return {
      state,
      nextAction: {
        type: 'start_child',
        phaseId: currentPhase.phaseId,
        childKind: currentDefinitionPhase.kind === 'chain' || currentDefinitionPhase.kind === 'team' ? currentDefinitionPhase.kind : 'chain',
        ...(currentDefinitionPhase.ref ? { ref: currentDefinitionPhase.ref } : {}),
      },
      recoveryOptions: [],
    }
  }

  if (recovery.type === 'handoff') {
    state.state = 'handoff'
    state.handoffSummary = recovery.summary ?? state.lastError?.message ?? 'Workflow requested handoff.'
    state.updatedAt = now
    delete state.currentPhaseId

    return {
      state,
      nextAction: null,
      recoveryOptions: [],
    }
  }

  const target = recovery.targetPhaseId ?? currentDefinitionPhase.on?.failure
  if (!target) {
    throw new Error(`Workflow phase "${currentDefinitionPhase.id}" has no failure transition to escalate to.`)
  }

  return moveToTarget(state, definition, target, now)
}

function moveToTarget(
  state: WorkflowState,
  definition: WorkflowDefinition,
  targetPhaseId: string,
  now: string,
): AdvanceWorkflowStateResult {
  if (HANDOFF_PHASE_IDS.has(targetPhaseId)) {
    state.state = 'handoff'
    state.currentPhaseId = targetPhaseId
    state.updatedAt = now
    markPhaseCompleted(state, targetPhaseId, now)

    return {
      state,
      nextAction: null,
      recoveryOptions: [],
    }
  }

  if (TERMINAL_PHASE_IDS.has(targetPhaseId)) {
    state.state = 'completed'
    state.currentPhaseId = targetPhaseId
    state.updatedAt = now
    markPhaseCompleted(state, targetPhaseId, now)

    return {
      state,
      nextAction: null,
      recoveryOptions: [],
    }
  }

  const nextAction = enterPhase(state, definition, targetPhaseId, now)
  return {
    state,
    nextAction,
    recoveryOptions: [],
  }
}

function enterPhase(state: WorkflowState, definition: WorkflowDefinition, phaseId: string, now: string): WorkflowAction | null {
  const definitionPhase = requireDefinitionPhase(definition, phaseId)
  const runtimePhase = requirePhaseState(state, phaseId)

  state.currentPhaseId = phaseId
  runtimePhase.startedAt = now

  if (definitionPhase.kind === 'chain' || definitionPhase.kind === 'team') {
    runtimePhase.state = 'waiting_on_child'
    state.state = 'waiting_on_child'
    state.updatedAt = now
    return {
      type: 'start_child',
      phaseId,
      childKind: definitionPhase.kind,
      ...(definitionPhase.ref ? { ref: definitionPhase.ref } : {}),
    }
  }

  if (definitionPhase.kind === 'gate') {
    runtimePhase.state = 'gated'
    state.state = 'gated'
    state.updatedAt = now
    return {
      type: 'gate',
      phaseId,
      ...(definitionPhase.gate ? { gate: definitionPhase.gate } : {}),
      ...(definitionPhase.prompt ? { prompt: definitionPhase.prompt } : {}),
    }
  }

  runtimePhase.state = 'completed'
  runtimePhase.completedAt = now
  state.updatedAt = now

  if (HANDOFF_PHASE_IDS.has(phaseId)) {
    state.state = 'handoff'
  } else {
    state.state = 'completed'
  }

  return null
}

function buildRecoveryOptions(phase: WorkflowPhaseDefinition, error: StructuredError): RecoveryAction[] {
  const actions: RecoveryAction[] = [
    { type: 'retry', maxAttempts: 1, guidance: `Retry workflow phase "${phase.id}".` },
    { type: 'handoff', summary: error.message },
  ]

  if (phase.on?.failure) {
    actions.push({
      type: 'escalate',
      targetAgent: phase.on.failure,
      reason: `Route workflow phase "${phase.id}" to "${phase.on.failure}".`,
    })
  }

  return actions
}

function buildGenericChildError(state: WorkflowState, phase: WorkflowPhaseDefinition): StructuredError {
  return {
    category: 'logical',
    code: 'WORKFLOW_CHILD_FAILED',
    message: `Workflow child phase "${phase.id}" failed.`,
    stepId: phase.id,
    agent: phase.kind,
    skills: [],
    context: {
      runId: state.workflowId,
      runKind: 'workflow',
      task: state.task,
      attempt: 1,
      hostCli: 'opencode',
      ...(requireCurrentPhase(state).childRun
        ? {
            child: {
              runId: requireCurrentPhase(state).childRun!.runId,
              runKind: requireCurrentPhase(state).childRun!.runKind,
              definitionName: requireCurrentPhase(state).childRun!.definitionName,
              phaseId: phase.id,
            },
          }
        : {}),
    },
    suggestedRecovery: { type: 'retry', maxAttempts: 1 },
    timestamp: new Date().toISOString(),
  }
}

function markPhaseCompleted(state: WorkflowState, phaseId: string, now: string): void {
  const phase = state.phases.find((entry) => entry.phaseId === phaseId)
  if (!phase) return
  phase.state = 'completed'
  phase.completedAt = now
}

function syncChildRun(state: WorkflowState, childRun: WorkflowChildRun): void {
  state.childRuns = state.childRuns.map((entry) => (entry.runId === childRun.runId ? childRun : entry))
}

function requireCurrentPhase(state: WorkflowState): WorkflowPhaseState {
  const phaseId = state.currentPhaseId
  if (!phaseId) throw new Error('Workflow has no current phase.')
  return requirePhaseState(state, phaseId)
}

function requirePhaseState(state: WorkflowState, phaseId: string): WorkflowPhaseState {
  const phase = state.phases.find((entry) => entry.phaseId === phaseId)
  if (!phase) throw new Error(`Unknown workflow phase state: ${phaseId}`)
  return phase
}

function requireDefinitionPhase(definition: WorkflowDefinition, phaseId: string): WorkflowPhaseDefinition {
  const phase = definition.phases.find((entry) => entry.id === phaseId)
  if (!phase) throw new Error(`Unknown workflow definition phase: ${phaseId}`)
  return phase
}
