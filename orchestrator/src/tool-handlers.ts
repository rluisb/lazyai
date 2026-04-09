import crypto from 'node:crypto'
import { aggregateBudgets, createBudgetState, evaluateBudgetHealth, updateBudget } from './budget-tracker.js'
import {
  advanceChainState,
  createChainState,
  escalateChainStep,
  retryChainStep,
} from './chain-machine.js'
import { buildBudgetPolicy, compileChainDefinition, validateStepOutput } from './compiler.js'
import { composeAgent as composePromptLayers } from './composer.js'
import {
  appendErrorJournalEntry,
  createErrorJournalEntry,
  createStructuredError,
  readErrorJournal,
} from './error-journal.js'
import { loadCatalog } from './loader.js'
import {
  loadChainState,
  loadExecutionPlan,
  loadTeamState,
  loadWorkflowState,
  saveChainState,
  saveExecutionPlan,
  saveHandoff,
  saveTeamState,
  saveWorkflowState,
} from './persistence.js'
import {
  assignTeamTask,
  completeTeamTask,
  createTeamState,
} from './team-machine.js'
import {
  advanceWorkflowState,
  applyWorkflowChildLaunch,
  createWorkflowState,
} from './workflow-machine.js'
import type {
  AdvanceChainInput,
  AdvanceChainResult,
  AdvanceWorkflowInput,
  AssignTaskInput,
  BaseAgentDefinition,
  BuildTeamInput,
  CatalogItem,
  ChainState,
  ChainStepStatus,
  ComposedAgentSpec,
  CompleteTaskInput,
  HandoffDocument,
  OrchestrationCatalog,
  RootContextLayer,
  StartChainInput,
  StartWorkflowInput,
  StructuredError,
  TeamState,
  WorkflowState,
} from './types.js'

export interface ToolHandlerOptions {
  projectRoot: string
  libraryOrchestrationRoot?: string
  libraryAgentsRoot?: string
}

export interface ComposeAgentInput {
  base: string
  domainSkill?: string
  modeSkill?: string
  stepInstructions?: string
  cliTool?: 'claude-code' | 'codex' | 'opencode' | 'gemini' | 'copilot'
  outputContract?: ComposedAgentSpec['outputContract']
  rootContext?: RootContextLayer
  allowedTools?: string[]
  model?: string
}

export interface ListCatalogInput {
  kinds?: Array<'chain' | 'team' | 'workflow' | 'domain' | 'mode'>
  includeProjectOverrides?: boolean
  query?: string
}

export interface GetStatusInput {
  runId: string
  kind: 'chain' | 'team' | 'workflow'
}

export interface GetBudgetInput {
  runId: string
  kind: 'chain' | 'team' | 'workflow'
}

export interface RetryStepInput {
  runId: string
  kind: 'chain'
  stepId: string
  reason?: string
}

export interface EscalateStepInput {
  runId: string
  kind: 'chain'
  stepId: string
  targetAgent: string
  domainSkill?: string
  modeSkill?: string
  reason?: string
}

export interface HandoffInput {
  runId: string
  kind: 'chain'
  summary?: string
  recipient?: string
  includeArtifacts?: boolean
}

export class OrchestratorToolHandlers {
  constructor(private readonly options: ToolHandlerOptions) {}

  listCatalog(input: ListCatalogInput = {}): { items: CatalogItem[] } {
    const catalog = this.getCatalog()
    const kinds = new Set(input.kinds ?? ['chain', 'team', 'workflow', 'domain', 'mode'])
    const items = [
      ...toCatalogItems('chain', catalog.chains),
      ...toCatalogItems('team', catalog.teams),
      ...toCatalogItems('workflow', catalog.workflows),
      ...toCatalogItems('domain', catalog.domains),
      ...toCatalogItems('mode', catalog.modes),
    ]
      .filter((item) => kinds.has(item.kind as never))
      .filter((item) => {
        if (input.includeProjectOverrides === false && item.source === 'project') return false
        if (!input.query) return true
        const haystack = `${item.kind} ${item.name} ${item.description}`.toLowerCase()
        return haystack.includes(input.query.toLowerCase())
      })
      .sort((left, right) => left.kind.localeCompare(right.kind) || left.name.localeCompare(right.name))

    return { items }
  }

  composeAgent(input: ComposeAgentInput): ComposedAgentSpec {
    const catalog = this.getCatalog()
    const agent = this.requireAgent(catalog, input.base)
    const domain = input.domainSkill ? catalog.domains[input.domainSkill] : undefined
    const mode = input.modeSkill ? catalog.modes[input.modeSkill] : undefined

    return composePromptLayers({
      ...(input.rootContext
        ? {
            root: {
              source: 'root',
              name: 'root-context',
              prompt: input.rootContext.prompt ?? '',
              ...(input.rootContext.allowedTools ? { allowedTools: input.rootContext.allowedTools } : {}),
              ...(input.rootContext.modelHint ? { modelHint: input.rootContext.modelHint } : {}),
              ...(input.rootContext.constraints ? { constraints: input.rootContext.constraints } : {}),
              ...(input.rootContext.approvalPolicy ? { approvalPolicy: input.rootContext.approvalPolicy } : {}),
            },
          }
        : {}),
      base: {
        source: 'base',
        name: agent.name,
        prompt: agent.prompt,
        allowedTools: agent.allowedTools,
        ...(agent.modelHint ? { modelHint: agent.modelHint } : {}),
        constraints: agent.constraints,
        approvalPolicy: 'minimal',
      },
      ...(domain
        ? {
            domain: {
              source: 'domain',
              name: domain.name,
              prompt: domain.prompt,
              ...(domain.allowedTools ? { allowedTools: domain.allowedTools } : {}),
              ...(domain.modelHint ? { modelHint: domain.modelHint } : {}),
              constraints: domain.constraints,
              ...(domain.approvalPolicy ? { approvalPolicy: domain.approvalPolicy } : {}),
            },
          }
        : {}),
      ...(mode
        ? {
            mode: {
              source: 'mode',
              name: mode.name,
              prompt: mode.prompt,
              ...(mode.allowedTools ? { allowedTools: mode.allowedTools } : {}),
              ...(mode.modelHint ? { modelHint: mode.modelHint } : {}),
              constraints: mode.constraints,
              ...(mode.approvalPolicy ? { approvalPolicy: mode.approvalPolicy } : {}),
            },
          }
        : {}),
      step: {
        source: 'step',
        name: 'ad-hoc-step',
        prompt: input.stepInstructions ?? '',
        ...(input.allowedTools ? { allowedTools: input.allowedTools } : {}),
        ...(input.model ? { modelHint: input.model } : {}),
        constraints: [],
        approvalPolicy: 'minimal',
        ...(input.outputContract ? { outputContract: input.outputContract } : {}),
      },
    })
  }

  startChain(input: StartChainInput) {
    const catalog = this.getCatalog()
    const plan = compileChainDefinition({
      catalog,
      projectRoot: this.options.projectRoot,
      chainName: input.chain,
      task: input.task,
      ...(input.context?.cliTool ? { cliTool: input.context.cliTool } : {}),
      ...(input.domainSkill ? { domainSkill: input.domainSkill } : {}),
      ...(input.modeSkill ? { modeSkill: input.modeSkill } : {}),
      ...(input.budget ? { budget: input.budget } : {}),
      ...(input.context?.rootContext ? { rootContext: input.context.rootContext } : {}),
      ...(input.context?.project ? { project: input.context.project } : {}),
    })

    saveExecutionPlan(this.options.projectRoot, plan)

    const state = createChainState(plan)
    state.budget = createBudgetState(plan.budgetPolicy)
    saveChainState(this.options.projectRoot, state)

    return {
      chainId: state.chainId,
      state: state.state,
      currentStep: this.toCurrentStepStatus(plan, state),
      budget: state.budget,
      executionPlanId: plan.id,
    }
  }

  advanceChain(input: AdvanceChainInput): AdvanceChainResult {
    const state = loadChainState(this.options.projectRoot, input.chainId)
    const plan = loadExecutionPlan(this.options.projectRoot, state.executionPlanId)
    const currentStep = this.requireCurrentStep(state)
    const compiled = this.requireCompiledStep(plan, input.stepId)

    let validationError: StructuredError | undefined
    const needsValidation = state.state !== 'gated' && input.outcome !== 'failure'
    if (needsValidation) {
      const validation = validateStepOutput(compiled.stepType, input.output)
      if (!validation.valid) {
        validationError = createStructuredError({
          category: 'validation',
          code: 'STEP_OUTPUT_SCHEMA_MISMATCH',
          message: `Step output for "${compiled.id}" did not match its contract.`,
          stepId: compiled.id,
          agent: compiled.agent,
          skills: compiled.skills,
          context: {
            runId: state.chainId,
            runKind: 'chain',
            task: state.task,
            attempt: currentStep.attempts,
            hostCli: plan.cli.host,
            budgetSnapshot: state.budget,
            ...(input.output ? { rawOutput: input.output } : {}),
            notes: validation.issues,
          },
          suggestedRecovery: {
            type: 'retry',
            maxAttempts: currentStep.maxRetries,
            guidance: validation.issues.join(' '),
          },
        })
      }
    }

    let nextBudget = state.budget
    if (input.usage) {
      const updated = updateBudget({
        budget: state.budget,
        policy: plan.budgetPolicy,
        stepId: input.stepId,
        usage: input.usage,
      })
      nextBudget = updated.budget
      state.budget = nextBudget
      if (updated.evaluation.shouldPause) {
        state.state = 'paused'
      }
    }

    const result = advanceChainState({
      state,
      plan,
      stepId: input.stepId,
      outcome: input.outcome,
      ...(input.output ? { output: input.output } : {}),
      ...(validationError ? { validationError } : {}),
    })

    result.stateSnapshot.budget = nextBudget
    saveChainState(this.options.projectRoot, result.stateSnapshot)

    if (result.error) {
      appendErrorJournalEntry(
        this.options.projectRoot,
        createErrorJournalEntry({
          runId: state.chainId,
          runKind: 'chain',
          definitionName: state.definitionName,
          stepId: input.stepId,
          error: result.error,
        }),
      )
    }

    return {
      state: result.state,
      nextStep: result.nextStep,
      gate: result.gate,
      recovery: result.recovery,
      budget: result.stateSnapshot.budget,
      ...(result.error ? { error: result.error } : {}),
    }
  }

  buildTeam(input: BuildTeamInput) {
    const catalog = this.getCatalog()
    const definition = catalog.teams[input.team]
    if (!definition) throw new Error(`Unknown team definition: ${input.team}`)

    const policy = buildBudgetPolicy('team', input.budget)
    const state = createTeamState({
      definition,
      task: input.task,
      policy,
      budget: createBudgetState(policy),
    })
    saveTeamState(this.options.projectRoot, state)

    return {
      teamId: state.teamId,
      state: state.state,
      readyTaskIds: state.readyTaskIds,
      tasks: state.tasks,
      budget: state.budget,
    }
  }

  assignTask(input: AssignTaskInput) {
    const state = loadTeamState(this.options.projectRoot, input.teamId)
    const assigned = assignTeamTask({
      state,
      taskId: input.taskId,
      assignee: input.assignee,
      ...(typeof input.claim === 'boolean' ? { claim: input.claim } : {}),
    })
    saveTeamState(this.options.projectRoot, assigned.state)

    return {
      teamId: assigned.state.teamId,
      state: assigned.state.state,
      readyTaskIds: assigned.state.readyTaskIds,
      task: assigned.state.tasks.find((task) => task.taskId === input.taskId) ?? null,
    }
  }

  completeTask(input: CompleteTaskInput) {
    const state = loadTeamState(this.options.projectRoot, input.teamId)
    const completion = completeTeamTask({
      state,
      taskId: input.taskId,
      outcome: input.outcome,
      ...(input.result ? { result: input.result } : {}),
      ...(input.usage ? { usage: input.usage } : {}),
      ...(input.error ? { error: input.error } : {}),
      policy: state.budgetPolicy,
    })
    saveTeamState(this.options.projectRoot, completion.state)

    if (input.outcome === 'failure' && input.error) {
      appendErrorJournalEntry(
        this.options.projectRoot,
        createErrorJournalEntry({
          runId: completion.state.teamId,
          runKind: 'team',
          definitionName: completion.state.definitionName,
          stepId: input.taskId,
          error: input.error,
        }),
      )
    }

    return {
      teamId: completion.state.teamId,
      state: completion.state.state,
      readyTaskIds: completion.state.readyTaskIds,
      budget: completion.state.budget,
      summary: completion.state.summary ?? null,
      evaluation: completion.evaluation,
    }
  }

  startWorkflow(input: StartWorkflowInput) {
    const catalog = this.getCatalog()
    const definition = catalog.workflows[input.workflow]
    if (!definition) throw new Error(`Unknown workflow definition: ${input.workflow}`)

    const policy = buildBudgetPolicy('workflow', input.budget)
    const created = createWorkflowState({
      definition,
      task: input.task,
      policy,
      budget: createBudgetState(policy),
      runtime: {
        ...(input.domainSkill ? { domainSkill: input.domainSkill } : {}),
        ...(input.modeSkill ? { modeSkill: input.modeSkill } : {}),
        ...(input.context ? { context: input.context } : {}),
      },
    })

    const state = this.materializeWorkflowAction(created.state, created.nextAction)
    this.refreshWorkflowBudget(state)
    saveWorkflowState(this.options.projectRoot, state)

    return {
      workflowId: state.workflowId,
      state: state.state,
      currentPhase: this.toCurrentWorkflowPhase(state),
      budget: state.budget,
    }
  }

  advanceWorkflow(input: AdvanceWorkflowInput) {
    const state = loadWorkflowState(this.options.projectRoot, input.workflowId)
    const definition = this.requireWorkflowDefinition(state.definitionName)
    const advanced = advanceWorkflowState({
      state,
      definition,
      ...(input.outcome ? { outcome: input.outcome } : {}),
      ...(input.recovery ? { recovery: input.recovery } : {}),
    })

    const nextState = this.materializeWorkflowAction(advanced.state, advanced.nextAction)
    this.refreshWorkflowBudget(nextState)
    saveWorkflowState(this.options.projectRoot, nextState)

    if (nextState.lastError) {
      appendErrorJournalEntry(
        this.options.projectRoot,
        createErrorJournalEntry({
          runId: nextState.workflowId,
          runKind: 'workflow',
          definitionName: nextState.definitionName,
          ...(nextState.currentPhaseId ? { stepId: nextState.currentPhaseId } : {}),
          error: nextState.lastError,
        }),
      )
    }

    return {
      workflowId: nextState.workflowId,
      state: nextState.state,
      currentPhase: this.toCurrentWorkflowPhase(nextState),
      budget: nextState.budget,
      recoveryOptions: advanced.recoveryOptions,
      ...(nextState.lastError ? { error: nextState.lastError } : {}),
    }
  }

  getStatus(input: GetStatusInput) {
    if (input.kind === 'chain') {
      const state = loadChainState(this.options.projectRoot, input.runId)
      const plan = loadExecutionPlan(this.options.projectRoot, state.executionPlanId)

      return {
        kind: 'chain',
        state: state.state,
        summary: {
          definitionName: state.definitionName,
          totalSteps: state.steps.length,
          completedSteps: state.completedStepIds.length,
          currentStepId: state.currentStepId ?? null,
        },
        current: this.toCurrentStepStatus(plan, state),
        budget: state.budget,
      }
    }

    if (input.kind === 'team') {
      const state = loadTeamState(this.options.projectRoot, input.runId)
      return {
        kind: 'team',
        state: state.state,
        summary: {
          definitionName: state.definitionName,
          totalTasks: state.tasks.length,
          completedTasks: state.tasks.filter((task) => task.state === 'completed').length,
          readyTaskIds: state.readyTaskIds,
        },
        current: this.toCurrentTeamTasks(state),
        budget: state.budget,
      }
    }

    const state = loadWorkflowState(this.options.projectRoot, input.runId)
    return {
      kind: 'workflow',
      state: state.state,
      summary: {
        definitionName: state.definitionName,
        totalPhases: state.phases.length,
        completedPhases: state.phases.filter((phase) => phase.state === 'completed').length,
        currentPhaseId: state.currentPhaseId ?? null,
      },
      current: this.toCurrentWorkflowPhase(state),
      budget: state.budget,
      ...(state.lastError ? { error: state.lastError } : {}),
    }
  }

  getBudget(input: GetBudgetInput) {
    if (input.kind === 'chain') {
      const state = loadChainState(this.options.projectRoot, input.runId)
      const plan = loadExecutionPlan(this.options.projectRoot, state.executionPlanId)
      return {
        ...state.budget,
        health: evaluateBudgetHealth(state.budget, plan.budgetPolicy),
      }
    }

    if (input.kind === 'team') {
      const state = loadTeamState(this.options.projectRoot, input.runId)
      return {
        ...state.budget,
        health: evaluateBudgetHealth(state.budget, state.budgetPolicy),
      }
    }

    const state = loadWorkflowState(this.options.projectRoot, input.runId)
    this.refreshWorkflowBudget(state)
    saveWorkflowState(this.options.projectRoot, state)
    return {
      ...state.budget,
      health: evaluateBudgetHealth(state.budget, state.budgetPolicy),
    }
  }

  retryStep(input: RetryStepInput) {
    const state = loadChainState(this.options.projectRoot, input.runId)
    const plan = loadExecutionPlan(this.options.projectRoot, state.executionPlanId)
    const retried = retryChainStep(state, plan, input.stepId)
    const budgetUpdate = updateBudget({
      budget: retried.state.budget,
      policy: plan.budgetPolicy,
      retryIncrement: 1,
      stepId: input.stepId,
    })
    retried.state.budget = budgetUpdate.budget
    saveChainState(this.options.projectRoot, retried.state)

    return {
      runId: retried.state.chainId,
      stepId: input.stepId,
      state: retried.state.state,
      attemptsRemaining: retried.attemptsRemaining,
    }
  }

  escalateStep(input: EscalateStepInput) {
    const state = loadChainState(this.options.projectRoot, input.runId)
    const plan = loadExecutionPlan(this.options.projectRoot, state.executionPlanId)
    const escalated = escalateChainStep(state, plan, input.stepId, input.targetAgent, input.domainSkill, input.modeSkill)

    saveExecutionPlan(this.options.projectRoot, escalated.plan)
    saveChainState(this.options.projectRoot, escalated.state)

    return {
      runId: escalated.state.chainId,
      stepId: input.stepId,
      state: escalated.state.state,
      newAssignment: this.toCurrentStepStatus(escalated.plan, escalated.state),
    }
  }

  handoff(input: HandoffInput) {
    const state = loadChainState(this.options.projectRoot, input.runId)
    const plan = loadExecutionPlan(this.options.projectRoot, state.executionPlanId)

    state.state = 'handoff'
    state.updatedAt = new Date().toISOString()

    const handoff: HandoffDocument = {
      id: crypto.randomUUID(),
      runId: state.chainId,
      kind: 'chain',
      summary: input.summary ?? `Handoff for chain ${state.chainId}`,
      ...(input.recipient ? { recipient: input.recipient } : {}),
      createdAt: new Date().toISOString(),
      resumable: true,
      status: state,
      plan,
    }

    const filePath = saveHandoff(this.options.projectRoot, handoff)
    state.handoffPath = filePath
    saveChainState(this.options.projectRoot, state)

    return {
      handoffId: handoff.id,
      path: filePath,
      summary: handoff.summary,
      resumable: true,
    }
  }

  getErrorJournal() {
    return readErrorJournal(this.options.projectRoot)
  }

  private materializeWorkflowAction(state: WorkflowState, action: ReturnType<typeof createWorkflowState>['nextAction']): WorkflowState {
    if (!action || action.type !== 'start_child' || !action.ref || !action.childKind) {
      return state
    }

    if (action.childKind === 'chain') {
      const child = this.startChain({
        chain: action.ref,
        task: state.task,
        ...(state.runtime?.domainSkill ? { domainSkill: state.runtime.domainSkill } : {}),
        ...(state.runtime?.modeSkill ? { modeSkill: state.runtime.modeSkill } : {}),
        ...(state.runtime?.context ? { context: state.runtime.context } : {}),
      })

      return applyWorkflowChildLaunch({
        state,
        phaseId: action.phaseId,
        childRun: {
          phaseId: action.phaseId,
          runId: child.chainId,
          runKind: 'chain',
          definitionName: action.ref,
          launchedAt: new Date().toISOString(),
        },
      }).state
    }

    const child = this.buildTeam({
      team: action.ref,
      task: state.task,
    })

    return applyWorkflowChildLaunch({
      state,
      phaseId: action.phaseId,
      childRun: {
        phaseId: action.phaseId,
        runId: child.teamId,
        runKind: 'team',
        definitionName: action.ref,
        launchedAt: new Date().toISOString(),
      },
    }).state
  }

  private refreshWorkflowBudget(state: WorkflowState): void {
    const children = state.childRuns.map((child) => {
      if (child.runKind === 'chain') {
        const childState = loadChainState(this.options.projectRoot, child.runId)
        return { childId: child.runId, budget: childState.budget }
      }

      const childState = loadTeamState(this.options.projectRoot, child.runId)
      return { childId: child.runId, budget: childState.budget }
    })

    const aggregated = aggregateBudgets({
      policy: state.budgetPolicy,
      scope: 'workflow',
      children,
    })
    state.budget = aggregated.budget
  }

  private getCatalog(): OrchestrationCatalog {
    return loadCatalog({
      projectRoot: this.options.projectRoot,
      ...(this.options.libraryOrchestrationRoot ? { libraryOrchestrationRoot: this.options.libraryOrchestrationRoot } : {}),
      ...(this.options.libraryAgentsRoot ? { libraryAgentsRoot: this.options.libraryAgentsRoot } : {}),
    })
  }

  private requireAgent(catalog: OrchestrationCatalog, name: string): BaseAgentDefinition {
    const agent = catalog.agents[name]
    if (!agent) throw new Error(`Unknown base agent: ${name}`)
    return agent
  }

  private requireCurrentStep(state: ChainState) {
    const step = state.steps.find((entry) => entry.stepId === state.currentStepId)
    if (!step) throw new Error('Chain has no current step.')
    return step
  }

  private requireCompiledStep(plan: ReturnType<typeof loadExecutionPlan>, stepId: string) {
    const step = plan.compiledSteps?.find((entry) => entry.id === stepId)
    if (!step) throw new Error(`Unknown compiled step: ${stepId}`)
    return step
  }

  private requireWorkflowDefinition(name: string) {
    const workflow = this.getCatalog().workflows[name]
    if (!workflow) throw new Error(`Unknown workflow definition: ${name}`)
    return workflow
  }

  private toCurrentStepStatus(plan: ReturnType<typeof loadExecutionPlan>, state: ChainState): ChainStepStatus | null {
    if (!state.currentStepId) return null
    const runtimeStep = state.steps.find((entry) => entry.stepId === state.currentStepId)
    const compiledStep = plan.compiledSteps?.find((entry) => entry.id === state.currentStepId)
    if (!runtimeStep || !compiledStep) return null

    return {
      stepId: compiledStep.id,
      agent: runtimeStep.agent,
      skills: runtimeStep.skills,
      stepType: runtimeStep.stepType,
      state: runtimeStep.state,
      model: compiledStep.model,
      tools: compiledStep.allowedTools,
      instructions: compiledStep.instructions,
      outputContract: compiledStep.outputContract,
      ...(runtimeStep.gate ? { gate: runtimeStep.gate } : {}),
      composedAgent: compiledStep.composedAgent,
    }
  }

  private toCurrentTeamTasks(state: TeamState) {
    return state.tasks.filter((task) => state.readyTaskIds.includes(task.taskId) || task.state === 'assigned' || task.state === 'claimed')
  }

  private toCurrentWorkflowPhase(state: WorkflowState) {
    if (!state.currentPhaseId) return null
    return state.phases.find((phase) => phase.phaseId === state.currentPhaseId) ?? null
  }
}

function toCatalogItems<T extends { name: string; source: CatalogItem['source']; description: string; version?: string; path: string }>(
  kind: CatalogItem['kind'],
  collection: Record<string, T>,
): CatalogItem[] {
  return Object.values(collection).map((item) => ({
    kind,
    name: item.name,
    source: item.source,
    description: item.description,
    ...(item.version ? { version: item.version } : {}),
    path: item.path,
  }))
}
