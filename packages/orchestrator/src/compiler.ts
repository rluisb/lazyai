import crypto from 'node:crypto'
import { z } from 'zod'
import { composeAgent } from './composer.js'
import { resolveSpecAgentContent } from './loader.js'
import type {
  ApprovalPolicy,
  BudgetPolicy,
  ChainDefinition,
  ChainStepDefinition,
  CliContext,
  CompiledStepPlan,
  ExecutionPlan,
  OrchestrationCatalog,
  ProjectStackContext,
  RootContextLayer,
  StepOutputContract,
  StepType,
} from './types.js'

const TERMINAL_TRANSITIONS = new Set(['done', 'handoff', 'abandon'])

const outputValidators = {
  research: z.object({
    summary: z.string(),
    status: z.string(),
    findings: z.array(z.unknown()),
  }).passthrough(),
  plan: z.object({
    summary: z.string(),
    status: z.string(),
    plan: z.unknown(),
    tasks: z.array(z.unknown()),
  }).passthrough(),
  implement: z.object({
    summary: z.string(),
    status: z.string(),
    files_changed: z.array(z.string()),
    tests_passed: z.boolean(),
  }).passthrough(),
  review: z.object({
    summary: z.string(),
    status: z.string(),
    verdict: z.string(),
    findings: z.array(z.unknown()),
  }).passthrough(),
  document: z.object({
    summary: z.string(),
    status: z.string(),
    files_created: z.array(z.string()),
  }).passthrough(),
  custom: z.record(z.unknown()),
} as const

export interface CompileChainOptions {
  catalog: OrchestrationCatalog
  projectRoot: string
  chainName: string
  task: string
  cliTool?: CliContext['host']
  domainSkill?: string
  modeSkill?: string
  budget?: Partial<BudgetPolicy>
  rootContext?: RootContextLayer
  project?: Omit<ProjectStackContext, 'rootPath'>
}

export interface ValidationResult {
  valid: boolean
  issues: string[]
}

export function compileChainDefinition(options: CompileChainOptions): ExecutionPlan {
  const chain = options.catalog.chains[options.chainName]
  if (!chain) {
    throw new Error(`Unknown chain definition: ${options.chainName}`)
  }

  validateChainDefinition(chain)

  const cli = getCliContext(options.cliTool)
  const budgetPolicy = buildBudgetPolicy('chain', options.budget)
  const rootContext = options.rootContext
  const compiledSteps = chain.steps.map((step) => compileStep(options.catalog, chain, step, options, rootContext))

  return {
    id: crypto.randomUUID(),
    kind: 'chain',
    definition: {
      kind: 'chain',
      name: chain.name,
      version: chain.version ?? '1.0.0',
      source: chain.source,
      path: chain.path,
    },
    cli,
    project: {
      rootPath: options.projectRoot,
      ...options.project,
    },
    budgetPolicy,
    entrypoint: chain.entry,
    compiledSteps,
    createdAt: new Date().toISOString(),
    task: options.task,
    ...(rootContext ? { rootContext } : {}),
  }
}

export function buildBudgetPolicy(
  scope: BudgetPolicy['scope'],
  overrides?: Partial<BudgetPolicy>,
): BudgetPolicy {
  return {
    id: overrides?.id ?? crypto.randomUUID(),
    scope,
    defaultActionOnLimit: overrides?.defaultActionOnLimit ?? 'pause',
    ...(overrides?.tokens ? { tokens: overrides.tokens } : {}),
    ...(overrides?.costUsd ? { costUsd: overrides.costUsd } : {}),
    ...(overrides?.wallClockMs ? { wallClockMs: overrides.wallClockMs } : {}),
    ...(overrides?.retries ? { retries: overrides.retries } : {}),
    ...(typeof overrides?.requireUserApprovalForTeamMultiplier === 'number'
      ? { requireUserApprovalForTeamMultiplier: overrides.requireUserApprovalForTeamMultiplier }
      : {}),
  }
}

export function getCliContext(host: CliContext['host'] = 'opencode'): CliContext {
  const contexts: Record<CliContext['host'], CliContext> = {
    'claude-code': {
      host: 'claude-code',
      dispatchMode: 'task-tool',
      supportsSubagents: true,
      supportsParallelTeams: true,
      supportsStructuredOutput: true,
      mcpServerName: 'ai-setup-orchestrator',
    },
    codex: {
      host: 'codex',
      dispatchMode: 'native-subagent',
      supportsSubagents: true,
      supportsParallelTeams: false,
      supportsStructuredOutput: true,
      mcpServerName: 'ai-setup-orchestrator',
    },
    opencode: {
      host: 'opencode',
      dispatchMode: 'task-tool',
      supportsSubagents: true,
      supportsParallelTeams: false,
      supportsStructuredOutput: true,
      mcpServerName: 'ai-setup-orchestrator',
    },
    gemini: {
      host: 'gemini',
      dispatchMode: 'instruction-only',
      supportsSubagents: false,
      supportsParallelTeams: false,
      supportsStructuredOutput: false,
      mcpServerName: 'ai-setup-orchestrator',
    },
    copilot: {
      host: 'copilot',
      dispatchMode: 'instruction-only',
      supportsSubagents: false,
      supportsParallelTeams: false,
      supportsStructuredOutput: false,
      mcpServerName: 'ai-setup-orchestrator',
    },
  }

  return contexts[host]
}

export function inferStepType(step: ChainStepDefinition): StepType {
  const identifiers = [step.id, ...step.skills].map((value) => value.toLowerCase())
  if (identifiers.some((value) => value.includes('research'))) return 'research'
  if (identifiers.some((value) => value.includes('plan'))) return 'plan'
  if (identifiers.some((value) => value.includes('review'))) return 'review'
  if (identifiers.some((value) => value.includes('document'))) return 'document'
  if (identifiers.some((value) => ['implement', 'fix', 'iterate'].some((token) => value.includes(token)))) return 'implement'
  return 'custom'
}

export function getOutputContract(stepType: StepType): StepOutputContract {
  switch (stepType) {
    case 'research':
      return contractFor(stepType, ['summary', 'status', 'findings'])
    case 'plan':
      return contractFor(stepType, ['summary', 'status', 'plan', 'tasks'])
    case 'implement':
      return contractFor(stepType, ['summary', 'status', 'files_changed', 'tests_passed'])
    case 'review':
      return contractFor(stepType, ['summary', 'status', 'verdict', 'findings'])
    case 'document':
      return contractFor(stepType, ['summary', 'status', 'files_created'])
    case 'custom':
      return {
        stepType,
        requiredFields: [],
        allowAdditionalProperties: true,
        schema: { type: 'object' },
        onValidationFailure: {
          category: 'validation',
          defaultRecovery: { type: 'retry', guidance: 'Return a structured JSON object for the custom step.' },
        },
      }
  }
}

export function validateStepOutput(stepType: StepType, payload: Record<string, unknown> | undefined): ValidationResult {
  if (!payload) {
    return {
      valid: false,
      issues: ['No output payload was provided.'],
    }
  }

  const result = outputValidators[stepType].safeParse(payload)
  return {
    valid: result.success,
    issues: result.success ? [] : result.error.issues.map((issue) => issue.message),
  }
}

function contractFor(stepType: StepType, requiredFields: string[]): StepOutputContract {
  return {
    stepType,
    requiredFields,
    allowAdditionalProperties: true,
    schema: {
      type: 'object',
      required: requiredFields,
    },
    onValidationFailure: {
      category: 'validation',
      defaultRecovery: {
        type: 'retry',
        guidance: `Return the required structured fields: ${requiredFields.join(', ')}.`,
      },
    },
  }
}

function compileStep(
  catalog: OrchestrationCatalog,
  chain: ChainDefinition,
  step: ChainStepDefinition,
  options: CompileChainOptions,
  rootContext?: RootContextLayer,
): CompiledStepPlan {
  const agent = catalog.agents[step.agent]
  if (!agent) {
    throw new Error(`Unknown base agent: ${step.agent}`)
  }

  const domainName = shouldInjectSkill(chain.domain_skill_injection, step) ? options.domainSkill : undefined
  const modeName = shouldInjectSkill(chain.mode_skill_injection, step) ? options.modeSkill : undefined
  const domain = domainName ? catalog.domains[domainName] : undefined
  const mode = modeName ? catalog.modes[modeName] : undefined
  const stepType = inferStepType(step)
  const outputContract = getOutputContract(stepType)
  const stepAllowedTools = step.allowedTools
  const stepApprovalPolicy: ApprovalPolicy = step.gate ? 'strict' : 'minimal'
  const specAgentContent = resolveSpecAgentContent(step.taskType)
  const instructions = buildStepInstructions(step, outputContract, specAgentContent)

  const composed = composeAgent({
    ...(rootContext
      ? {
          root: {
            source: 'root',
            name: 'root-context',
            prompt: rootContext.prompt ?? '',
            ...(rootContext.allowedTools ? { allowedTools: rootContext.allowedTools } : {}),
            ...(rootContext.modelHint ? { modelHint: rootContext.modelHint } : {}),
            ...(rootContext.constraints ? { constraints: rootContext.constraints } : {}),
            ...(rootContext.approvalPolicy ? { approvalPolicy: rootContext.approvalPolicy } : {}),
          },
        }
      : {}),
    base: {
      source: 'base',
      name: agent.name,
      prompt: agent.prompt,
      allowedTools: agent.allowedTools,
      constraints: agent.constraints,
      approvalPolicy: 'minimal',
      ...(agent.modelHint ? { modelHint: agent.modelHint } : {}),
    },
    ...(domain
      ? {
          domain: {
            source: 'domain',
            name: domain.name,
            prompt: domain.prompt,
            constraints: domain.constraints,
            ...(domain.allowedTools ? { allowedTools: domain.allowedTools } : {}),
            ...(domain.modelHint ? { modelHint: domain.modelHint } : {}),
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
            constraints: mode.constraints,
            ...(mode.allowedTools ? { allowedTools: mode.allowedTools } : {}),
            ...(mode.modelHint ? { modelHint: mode.modelHint } : {}),
            ...(mode.approvalPolicy ? { approvalPolicy: mode.approvalPolicy } : {}),
          },
        }
      : {}),
      step: {
        source: 'step',
        name: step.id,
        prompt: instructions,
        constraints: [],
        approvalPolicy: stepApprovalPolicy,
        outputContract,
        ...(stepAllowedTools ? { allowedTools: stepAllowedTools } : {}),
        ...(step.model ? { modelHint: step.model } : {}),
      },
    })

  return {
    id: step.id,
    kind: 'step',
    agent: agent.name,
    skills: step.skills,
    ...(step.taskType ? { taskType: step.taskType } : {}),
    stepType,
    ...(domain ? { domainSkill: domain.name } : {}),
    ...(mode ? { modeSkill: mode.name } : {}),
    instructions,
    allowedTools: composed.tools,
    model: composed.model,
    outputContract,
    transitions: step.transitions,
    ...(step.gate ? { gate: step.gate } : {}),
    composedAgent: composed,
  }
}

function shouldInjectSkill(
  injection: ChainDefinition['domain_skill_injection'] | ChainDefinition['mode_skill_injection'],
  step: ChainStepDefinition,
): boolean {
  if (injection === 'none') return false
  if (injection === 'builder_steps_only') return step.agent === 'builder'
  return true
}

function buildStepInstructions(
  step: ChainStepDefinition,
  outputContract: StepOutputContract,
  specAgentContent?: string,
): string {
  const sections = [
    `Step: ${step.id}`,
    step.description,
    step.prompt,
    step.taskType ? `Task Type: ${step.taskType}` : undefined,
    step.skills.length > 0 ? `Apply supporting skills: ${step.skills.join(', ')}.` : undefined,
    `Return structured output with required fields: ${outputContract.requiredFields.join(', ')}.`,
    `Valid outcomes: ${Object.keys(step.transitions).join(', ')}.`,
    step.gate ? `A ${step.gate} gate must be satisfied before the chain can continue.` : undefined,
    specAgentContent ? `Relevant spec-agent guidance:\n\n${specAgentContent}` : undefined,
  ]

  return sections.filter(Boolean).join('\n\n')
}

function validateChainDefinition(chain: ChainDefinition): void {
  const stepIds = new Set(chain.steps.map((step) => step.id))
  if (!stepIds.has(chain.entry)) {
    throw new Error(`Chain entry step \"${chain.entry}\" does not exist.`)
  }

  for (const step of chain.steps) {
    if (Object.keys(step.transitions).length === 0) {
      throw new Error(`Chain step \"${step.id}\" must define at least one transition.`)
    }

    for (const transition of Object.values(step.transitions)) {
      const target = typeof transition === 'string' ? transition : transition.then
      if (!stepIds.has(target) && !TERMINAL_TRANSITIONS.has(target)) {
        throw new Error(`Chain step \"${step.id}\" references unknown transition target \"${target}\".`)
      }
    }
  }
}
