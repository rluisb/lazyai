export type DefinitionSource = 'library' | 'project' | 'db' | 'user_project' | 'user_global'
export type CatalogKind = 'agent' | 'domain' | 'mode' | 'chain' | 'team' | 'workflow'
export type HostCli = 'claude-code' | 'codex' | 'opencode' | 'gemini' | 'copilot'
export type DispatchMode = 'task-tool' | 'native-subagent' | 'sdk-session' | 'instruction-only'
export type ApprovalPolicy = 'minimal' | 'normal' | 'strict'
export type RunKind = 'chain' | 'team' | 'workflow'
export type ChainLifecycleState =
  | 'created'
  | 'running'
  | 'gated'
  | 'paused'
  | 'completed'
  | 'abandoned'
  | 'handoff'
export type StepLifecycleState =
  | 'pending'
  | 'running'
  | 'completed'
  | 'failed'
  | 'retrying'
  | 'escalated'
  | 'skipped'
  | 'abandoned'
export type TeamLifecycleState = 'created' | 'running' | 'synthesizing' | 'completed' | 'failed' | 'paused' | 'handoff'
export type TeamTaskLifecycleState = 'pending' | 'assigned' | 'claimed' | 'blocked' | 'completed' | 'failed'
export type WorkflowLifecycleState =
  | 'created'
  | 'waiting_on_child'
  | 'running'
  | 'gated'
  | 'awaiting_recovery'
  | 'completed'
  | 'failed'
  | 'handoff'
  | 'paused'
export type WorkflowPhaseLifecycleState =
  | 'pending'
  | 'waiting_on_child'
  | 'running'
  | 'gated'
  | 'completed'
  | 'failed'
  | 'skipped'
export type BudgetHealth = 'ok' | 'warning' | 'limit_reached'
export type ErrorCategory = 'transient' | 'logical' | 'budget' | 'permission' | 'validation' | 'fatal'
export type StepType = 'research' | 'plan' | 'implement' | 'review' | 'document' | 'custom'
export type DriftStatus = 'fresh' | 'stale' | 'stale_acked' | 'disabled' | 'unavailable' | 'unknown'
export type MaintenanceApprovalScope = 'per_action' | 'task_scoped' | 'session_scoped' | 'standing'

export interface DefinitionMetadata {
  name: string
  description: string
  version?: string
  source: DefinitionSource
  path: string
}

export interface BaseAgentDefinition extends DefinitionMetadata {
  kind: 'agent'
  prompt: string
  modelHint?: string
  allowedTools: string[]
  constraints: string[]
  displayName: string
}

export interface SkillDefinition extends DefinitionMetadata {
  kind: 'domain' | 'mode'
  prompt: string
  allowedTools?: string[]
  constraints: string[]
  modelHint?: string
  approvalPolicy?: ApprovalPolicy
  appliesTo?: string[]
}

export type StepTransition = string | { retry: number; then: string }

export interface ChainStepDefinition {
  id: string
  agent: string
  skills: string[]
  description: string
  taskType?: string
  gate?: 'user_approval' | 'severity_confirmation' | 'cost_confirmation'
  prompt?: string
  transitions: Record<string, StepTransition>
  allowedTools?: string[]
  model?: string
}

export interface ChainDefinition extends DefinitionMetadata {
  kind: 'chain'
  entry: string
  steps: ChainStepDefinition[]
  domain_skill_injection?: 'all_steps' | 'builder_steps_only' | 'none'
  mode_skill_injection?: 'all_steps' | 'builder_steps_only' | 'none'
}

export interface TeamMemberDefinition {
  role: string
  agent: string
  skills: string[]
  focus: string
}

export interface TeamDefinition extends DefinitionMetadata {
  kind: 'team'
  budget_multiplier?: number
  user_confirmation_required?: boolean
  parallel: TeamMemberDefinition[]
  synthesize: {
    agent: string
    description: string
  }
}

export interface WorkflowPhaseDefinition {
  id: string
  kind: 'chain' | 'team' | 'gate' | 'terminal'
  ref?: string
  gate?: 'user_approval' | 'severity_confirmation' | 'cost_confirmation'
  prompt?: string
  when?: string
  on?: Record<string, string>
}

export interface WorkflowDefinition extends DefinitionMetadata {
  kind: 'workflow'
  entry: string
  phases: WorkflowPhaseDefinition[]
}

export interface OrchestrationCatalog {
  agents: Record<string, BaseAgentDefinition>
  domains: Record<string, SkillDefinition>
  modes: Record<string, SkillDefinition>
  chains: Record<string, ChainDefinition>
  teams: Record<string, TeamDefinition>
  workflows: Record<string, WorkflowDefinition>
}

export interface PromptLayer {
  source: 'root' | 'base' | 'domain' | 'mode' | 'step'
  name: string
  prompt: string
  allowedTools?: string[]
  modelHint?: string
  constraints?: string[]
  approvalPolicy?: ApprovalPolicy
}

export interface MaintenanceContractRecord {
  id: string
  status?: string
  approvalScope: MaintenanceApprovalScope
  approvalExpiresAt?: string
  permittedActions: string[]
  workflowRunId?: string
  chainId?: string
}

export interface SyncToolSnapshot {
  enabled: boolean
  indexPath?: string
  dataPath?: string
  lastIndexTime?: string
  sourceFingerprint?: string
  driftStatus?: DriftStatus
}

export interface SyncStateSnapshot {
  schemaVersion: number
  updatedAt: string
  qmd: SyncToolSnapshot
  codegraph: SyncToolSnapshot
  staleAcked: {
    qmd: Array<{ fingerprint: string; ackedAt: string; reason?: string }>
    codegraph: Array<{ fingerprint: string; ackedAt: string; reason?: string }>
  }
  repairProposals: Array<{ tool: string; proposalId: string; status: string; createdAt: string; reason: string }>
}

export interface BootstrapReport {
  memoryPath: string
  specsAvailable: boolean
  syncStatePresent: boolean
  qmdStatus: DriftStatus
  codegraphStatus: DriftStatus
  contractStatus: 'active' | 'expired' | 'missing'
  contractScope?: MaintenanceApprovalScope
  approvalsRequested: string[]
  contextLoaded: string[]
  deferredMaintenance: string[]
  warnings: string[]
}

export interface HousekeepingReport {
  phase: 'pre' | 'inline' | 'post' | 'combined'
  contractStatus: 'active' | 'expired' | 'missing'
  approvalsRequested: string[]
  contextLoaded: string[]
  stagedMemoryEntries: string[]
  deferredMaintenance: string[]
  qmdStatus: DriftStatus
  codegraphStatus: DriftStatus
  warnings: string[]
}

export interface RootContextLayer {
  prompt?: string
  constraints?: string[]
  allowedTools?: string[]
  modelHint?: string
  approvalPolicy?: ApprovalPolicy
}

export interface StepOutputContract {
  stepType: StepType
  requiredFields: string[]
  allowAdditionalProperties: boolean
  schema: Record<string, unknown>
  onValidationFailure: {
    category: 'validation'
    defaultRecovery: RecoveryAction
  }
}

export interface ComposedAgentSpec {
  id: string
  base: string
  domainSkill?: string
  modeSkill?: string
  model: string
  tools: string[]
  approvalPolicy: ApprovalPolicy
  constraints: string[]
  prompt: string
  outputContract?: StepOutputContract
  mergedFrom: PromptLayer[]
}

export interface CliContext {
  host: HostCli
  dispatchMode: DispatchMode
  supportsSubagents: boolean
  supportsParallelTeams: boolean
  supportsStructuredOutput: boolean
  mcpServerName: string
}

export interface ProjectStackContext {
  rootPath: string
  rootInstructionFile?: string
  language?: string
  framework?: string
  database?: string
  packageManager?: string
  testCommand?: string
  buildCommand?: string
}

export interface BudgetThreshold {
  limit?: number
  warnAt?: number
  pauseAt?: number
  hardStop?: boolean
}

export interface BudgetPolicy {
  id: string
  scope: 'chain' | 'team' | 'workflow'
  tokens?: BudgetThreshold
  costUsd?: BudgetThreshold
  wallClockMs?: BudgetThreshold
  retries?: BudgetThreshold
  requireUserApprovalForTeamMultiplier?: number
  defaultActionOnLimit: 'warn' | 'pause' | 'abort'
}

export interface BudgetDimensionState {
  limit?: number
  consumed: number
  remaining?: number
  warningTriggered: boolean
  pausedAtLimit: boolean
}

export interface StepUsage {
  inputTokens?: number
  outputTokens?: number
  totalTokens?: number
  costUsd?: number
  wallClockMs?: number
}

export interface BudgetState {
  policyId: string
  scope: 'chain' | 'team' | 'workflow'
  tokens: BudgetDimensionState
  costUsd: BudgetDimensionState
  wallClockMs: BudgetDimensionState
  retries: BudgetDimensionState
  byStep: Record<string, StepUsage>
  lastUpdatedAt: string
}

export interface BudgetEvaluation {
  overall: BudgetHealth
  dimensions: {
    tokens: BudgetHealth
    costUsd: BudgetHealth
    wallClockMs: BudgetHealth
    retries: BudgetHealth
  }
  recommendedAction: 'continue' | 'warn' | 'pause' | 'abort'
  shouldPause: boolean
}

export type RecoveryAction =
  | { type: 'retry'; maxAttempts?: number; guidance?: string }
  | { type: 'fix_and_resume'; instructions: string }
  | { type: 'escalate'; targetAgent: string; reason: string }
  | { type: 'pause'; reason: string }
  | { type: 'handoff'; recipient?: string; summary?: string }
  | { type: 'abort'; reason: string }

export interface StructuredError {
  category: ErrorCategory
  code: string
  message: string
  stepId: string
  agent: string
  skills: string[]
  context: {
    runId: string
    runKind: RunKind
    task: string
    attempt: number
    hostCli: HostCli
    budgetSnapshot?: BudgetState
    rawOutput?: Record<string, unknown>
    notes?: string[]
    child?: {
      runId: string
      runKind: Exclude<RunKind, 'workflow'>
      definitionName: string
      phaseId: string
      stepId?: string
      taskId?: string
    }
  }
  suggestedRecovery: RecoveryAction
  timestamp: string
}

export interface ErrorJournalEntry {
  id: string
  runId: string
  runKind: RunKind
  definitionName: string
  stepId?: string
  error: StructuredError
  resolution?: {
    action: RecoveryAction
    resolvedAt?: string
    notes?: string
  }
  lesson?: {
    summary: string
    prevention: string[]
    tags: string[]
  }
}

export interface DefinitionRef {
  kind: 'chain' | 'team' | 'workflow'
  name: string
  version: string
  source: DefinitionSource
  path: string
}

export interface CompiledStepPlan {
  id: string
  kind: 'step'
  agent: string
  skills: string[]
  taskType?: string
  stepType: StepType
  domainSkill?: string
  modeSkill?: string
  instructions: string
  allowedTools: string[]
  model: string
  outputContract: StepOutputContract
  transitions: Record<string, StepTransition>
  gate?: 'user_approval' | 'severity_confirmation' | 'cost_confirmation'
  composedAgent: ComposedAgentSpec
}

export interface ExecutionPlan {
  id: string
  kind: 'chain' | 'team' | 'workflow'
  definition: DefinitionRef
  cli: CliContext
  project: ProjectStackContext
  budgetPolicy: BudgetPolicy
  entrypoint: string
  compiledSteps?: CompiledStepPlan[]
  createdAt: string
  task: string
  rootContext?: RootContextLayer
}

export interface GateState {
  type: 'user_approval' | 'severity_confirmation' | 'cost_confirmation'
  prompt: string
  status: 'pending' | 'approved' | 'rejected'
  decidedAt?: string
}

export interface StepState {
  stepId: string
  order: number
  agent: string
  skills: string[]
  stepType: StepType
  domainSkill?: string
  modeSkill?: string
  state: StepLifecycleState
  attempts: number
  maxRetries: number
  startedAt?: string
  completedAt?: string
  output?: Record<string, unknown>
  outputValid?: boolean
  usage: StepUsage
  gate?: GateState
  lastOutcome?: string
  error?: StructuredError
  nextStepId?: string
  housekeepingReport?: HousekeepingReport
}

export interface ChainState {
  chainId: string
  definitionName: string
  definitionVersion: string
  executionPlanId: string
  state: ChainLifecycleState
  task: string
  currentStepId?: string
  entryStepId: string
  steps: StepState[]
  completedStepIds: string[]
  budget: BudgetState
  createdAt: string
  updatedAt: string
  handoffPath?: string
  bootstrapReport?: BootstrapReport
}

export interface StartChainInput {
  chain: string
  task: string
  domainSkill?: string
  modeSkill?: string
  budget?: Partial<BudgetPolicy>
  context?: {
    cliTool?: HostCli
    rootContext?: RootContextLayer
    project?: Omit<ProjectStackContext, 'rootPath'>
  }
}

export interface AdvanceChainInput {
  chainId: string
  stepId: string
  outcome: string
  output?: Record<string, unknown>
  usage?: StepUsage
}

export interface ChainStepStatus {
  stepId: string
  agent: string
  skills: string[]
  stepType: StepType
  state: StepLifecycleState
  model: string
  tools: string[]
  instructions: string
  outputContract: StepOutputContract
  gate?: GateState
  composedAgent: ComposedAgentSpec
}

export interface AdvanceChainResult {
  state: ChainLifecycleState
  nextStep: ChainStepStatus | null
  gate: GateState | null
  recovery: RecoveryAction | null
  budget: BudgetState
  error?: StructuredError
}

export interface CatalogItem {
  kind: CatalogKind
  name: string
  source: DefinitionSource
  description: string
  version?: string
  path: string
}

export interface HandoffDocument {
  id: string
  runId: string
  kind: RunKind
  summary: string
  recipient?: string
  createdAt: string
  resumable: boolean
  status: ChainState
  plan: ExecutionPlan
}

export interface TeamTaskState {
  taskId: string
  kind: 'member' | 'synthesize'
  role: string
  agent: string
  skills: string[]
  focus: string
  state: TeamTaskLifecycleState
  order: number
  dependsOn: string[]
  assignee?: string
  claimedBy?: string
  assignedAt?: string
  claimedAt?: string
  completedAt?: string
  result?: Record<string, unknown>
  usage: StepUsage
  error?: StructuredError
}

export interface TeamState {
  teamId: string
  definitionName: string
  definitionVersion: string
  state: TeamLifecycleState
  task: string
  tasks: TeamTaskState[]
  readyTaskIds: string[]
  synthesisTaskId: string
  budgetPolicy: BudgetPolicy
  budget: BudgetState
  createdAt: string
  updatedAt: string
  summary?: Record<string, unknown>
}

export interface BuildTeamInput {
  team: string
  task: string
  budget?: Partial<BudgetPolicy>
}

export interface AssignTaskInput {
  teamId: string
  taskId: string
  assignee: string
  claim?: boolean
}

export interface CompleteTaskInput {
  teamId: string
  taskId: string
  outcome: 'success' | 'failure'
  result?: Record<string, unknown>
  usage?: StepUsage
  error?: StructuredError
}

export interface WorkflowChildRun {
  phaseId: string
  runId: string
  runKind: 'chain' | 'team'
  definitionName: string
  launchedAt: string
  completedAt?: string
  outcome?: string
}

export interface WorkflowPhaseState {
  phaseId: string
  kind: WorkflowPhaseDefinition['kind']
  state: WorkflowPhaseLifecycleState
  ref?: string
  gate?: WorkflowPhaseDefinition['gate']
  prompt?: string
  startedAt?: string
  completedAt?: string
  lastOutcome?: string
  childRun?: WorkflowChildRun
}

export interface WorkflowRecoveryDecision {
  type: 'retry' | 'escalate' | 'handoff'
  targetPhaseId?: string
  reason?: string
  recipient?: string
  summary?: string
}

export interface WorkflowState {
  workflowId: string
  definitionName: string
  definitionVersion: string
  state: WorkflowLifecycleState
  task: string
  entryPhaseId: string
  currentPhaseId?: string
  phases: WorkflowPhaseState[]
  childRuns: WorkflowChildRun[]
  budgetPolicy: BudgetPolicy
  budget: BudgetState
  createdAt: string
  updatedAt: string
  lastError?: StructuredError
  handoffSummary?: string
  runtime?: {
    domainSkill?: string
    modeSkill?: string
    context?: StartChainInput['context']
  }
}

export interface StartWorkflowInput {
  workflow: string
  task: string
  domainSkill?: string
  modeSkill?: string
  budget?: Partial<BudgetPolicy>
  context?: StartChainInput['context']
}

export interface AdvanceWorkflowInput {
  workflowId: string
  outcome?: string
  recovery?: WorkflowRecoveryDecision
}

export interface WorkflowAction {
  type: 'start_child' | 'gate' | 'terminal'
  phaseId: string
  childKind?: 'chain' | 'team'
  ref?: string
  gate?: WorkflowPhaseDefinition['gate']
  prompt?: string
}
