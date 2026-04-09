import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js'
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js'
import { z } from 'zod'
import {
  OrchestratorToolHandlers,
  type ComposeAgentInput,
  type EscalateStepInput,
  type GetBudgetInput,
  type GetStatusInput,
  type HandoffInput,
  type ListCatalogInput,
  type RetryStepInput,
  type ToolHandlerOptions,
} from './tool-handlers.js'
import type { AdvanceChainInput, HostCli, StartChainInput, StepOutputContract, StepUsage } from './types.js'

const CATALOG_KIND_SCHEMA = z.enum(['chain', 'team', 'workflow', 'domain', 'mode'])
const HOST_CLI_SCHEMA = z.enum(['claude-code', 'codex', 'opencode', 'gemini', 'copilot'])
const CHAIN_KIND_SCHEMA = z.literal('chain')

export interface OrchestratorServerContext {
  server: McpServer
  handlers: OrchestratorToolHandlers
}

type StructuredContent = Record<string, unknown>
type HandlerCliTool = NonNullable<ComposeAgentInput['cliTool']>
type ToolArgs = Record<string, unknown>

function definedEntries<T extends Record<string, unknown>>(value: T): Partial<T> {
  return Object.fromEntries(
    Object.entries(value).filter(([, entry]) => entry !== undefined),
  ) as Partial<T>
}

function toStructuredContent(data: unknown): StructuredContent | undefined {
  return isStructuredContent(data) ? data : undefined
}

function isStructuredContent(value: unknown): value is StructuredContent {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function normalizeListCatalogInput(args: Record<string, unknown>): ListCatalogInput {
  return definedEntries(args)
}

function normalizeComposeAgentInput(args: Record<string, unknown>): ComposeAgentInput {
  const normalized: ComposeAgentInput = {
    base: args.base as string,
  }

  if (typeof args.domainSkill === 'string') normalized.domainSkill = args.domainSkill
  if (typeof args.modeSkill === 'string') normalized.modeSkill = args.modeSkill
  if (typeof args.stepInstructions === 'string') normalized.stepInstructions = args.stepInstructions
  if (typeof args.cliTool === 'string') normalized.cliTool = args.cliTool as HandlerCliTool
  if (Array.isArray(args.allowedTools)) normalized.allowedTools = args.allowedTools as string[]
  if (typeof args.model === 'string') normalized.model = args.model
  if (isStructuredContent(args.outputContract)) normalized.outputContract = args.outputContract as unknown as StepOutputContract
  if (isStructuredContent(args.rootContext)) normalized.rootContext = definedEntries(args.rootContext) as NonNullable<ComposeAgentInput['rootContext']>

  return normalized
}

function normalizeStartChainInput(args: Record<string, unknown>): StartChainInput {
  const normalized: StartChainInput = {
    chain: args.chain as string,
    task: args.task as string,
  }

  if (typeof args.domainSkill === 'string') normalized.domainSkill = args.domainSkill
  if (typeof args.modeSkill === 'string') normalized.modeSkill = args.modeSkill
  if (isStructuredContent(args.budget)) normalized.budget = args.budget as Partial<NonNullable<StartChainInput['budget']>>
  if (isStructuredContent(args.context)) {
    const context = args.context
    const nextContext: NonNullable<StartChainInput['context']> = {}
    if (typeof context.cliTool === 'string') nextContext.cliTool = context.cliTool as HostCli
    if (isStructuredContent(context.rootContext)) {
      nextContext.rootContext = definedEntries(context.rootContext) as NonNullable<NonNullable<StartChainInput['context']>['rootContext']>
    }
    if (isStructuredContent(context.project)) {
      nextContext.project = definedEntries(context.project) as NonNullable<NonNullable<StartChainInput['context']>['project']>
    }
    normalized.context = nextContext
  }

  return normalized
}

function normalizeAdvanceChainInput(args: Record<string, unknown>): AdvanceChainInput {
  const normalized: AdvanceChainInput = {
    chainId: args.chainId as string,
    stepId: args.stepId as string,
    outcome: args.outcome as string,
  }

  if (isStructuredContent(args.output)) normalized.output = args.output
  if (isStructuredContent(args.usage)) normalized.usage = definedEntries(args.usage) as StepUsage

  return normalized
}

function normalizeGetStatusInput(args: { runId: string; kind: 'chain' }): GetStatusInput {
  return args
}

function normalizeGetBudgetInput(args: { runId: string; kind: 'chain' }): GetBudgetInput {
  return args
}

function normalizeRetryStepInput(args: Record<string, unknown>): RetryStepInput {
  const normalized: RetryStepInput = {
    runId: args.runId as string,
    kind: args.kind as 'chain',
    stepId: args.stepId as string,
  }

  if (typeof args.reason === 'string') normalized.reason = args.reason
  return normalized
}

function normalizeEscalateStepInput(args: Record<string, unknown>): EscalateStepInput {
  const normalized: EscalateStepInput = {
    runId: args.runId as string,
    kind: args.kind as 'chain',
    stepId: args.stepId as string,
    targetAgent: args.targetAgent as string,
  }

  if (typeof args.domainSkill === 'string') normalized.domainSkill = args.domainSkill
  if (typeof args.modeSkill === 'string') normalized.modeSkill = args.modeSkill
  if (typeof args.reason === 'string') normalized.reason = args.reason
  return normalized
}

function normalizeHandoffInput(args: Record<string, unknown>): HandoffInput {
  const normalized: HandoffInput = {
    runId: args.runId as string,
    kind: args.kind as 'chain',
  }

  if (typeof args.summary === 'string') normalized.summary = args.summary
  if (typeof args.recipient === 'string') normalized.recipient = args.recipient
  if (typeof args.includeArtifacts === 'boolean') normalized.includeArtifacts = args.includeArtifacts
  return normalized
}

export function createOrchestratorServer(options: ToolHandlerOptions): OrchestratorServerContext {
  const handlers = new OrchestratorToolHandlers(options)
  const server = new McpServer({
    name: 'ai-setup-orchestrator',
    version: '0.1.0',
  })

  server.registerTool(
    'list_catalog',
    {
      description: 'List orchestration catalog definitions available to the runtime.',
      inputSchema: {
        kinds: z.array(CATALOG_KIND_SCHEMA).optional(),
        includeProjectOverrides: z.boolean().optional(),
        query: z.string().min(1).optional(),
      },
    },
    async (args: ToolArgs) => formatToolResult(handlers.listCatalog(normalizeListCatalogInput(args))),
  )

  server.registerTool(
    'compose_agent',
    {
      description: 'Compose a runtime agent prompt from base, domain, mode, and step layers.',
      inputSchema: {
        base: z.string().min(1),
        domainSkill: z.string().min(1).optional(),
        modeSkill: z.string().min(1).optional(),
        stepInstructions: z.string().optional(),
        cliTool: HOST_CLI_SCHEMA.optional(),
        outputContract: z.record(z.unknown()).optional(),
        rootContext: z
          .object({
            prompt: z.string().optional(),
            constraints: z.array(z.string()).optional(),
            allowedTools: z.array(z.string()).optional(),
            modelHint: z.string().optional(),
            approvalPolicy: z.enum(['minimal', 'normal', 'strict']).optional(),
          })
          .optional(),
        allowedTools: z.array(z.string()).optional(),
        model: z.string().optional(),
      },
    },
    async (args: ToolArgs) => formatToolResult(handlers.composeAgent(normalizeComposeAgentInput(args))),
  )

  server.registerTool(
    'start_chain',
    {
      description: 'Compile and start a chain execution plan.',
      inputSchema: {
        chain: z.string().min(1),
        task: z.string().min(1),
        domainSkill: z.string().min(1).optional(),
        modeSkill: z.string().min(1).optional(),
        budget: z.record(z.unknown()).optional(),
        context: z
          .object({
            cliTool: HOST_CLI_SCHEMA.optional(),
            rootContext: z
              .object({
                prompt: z.string().optional(),
                constraints: z.array(z.string()).optional(),
                allowedTools: z.array(z.string()).optional(),
                modelHint: z.string().optional(),
                approvalPolicy: z.enum(['minimal', 'normal', 'strict']).optional(),
              })
              .optional(),
            project: z.record(z.unknown()).optional(),
          })
          .optional(),
      },
    },
    async (args: ToolArgs) => formatToolResult(handlers.startChain(normalizeStartChainInput(args))),
  )

  server.registerTool(
    'advance_chain',
    {
      description: 'Advance a running chain after a step completes or fails.',
      inputSchema: {
        chainId: z.string().min(1),
        stepId: z.string().min(1),
        outcome: z.string().min(1),
        output: z.record(z.unknown()).optional(),
        usage: z
          .object({
            inputTokens: z.number().nonnegative().optional(),
            outputTokens: z.number().nonnegative().optional(),
            totalTokens: z.number().nonnegative().optional(),
            costUsd: z.number().nonnegative().optional(),
            wallClockMs: z.number().nonnegative().optional(),
          })
          .optional(),
      },
    },
    async (args: ToolArgs) => formatToolResult(handlers.advanceChain(normalizeAdvanceChainInput(args))),
  )

  server.registerTool(
    'get_status',
    {
      description: 'Get the current runtime status for a Phase 2 chain.',
      inputSchema: {
        runId: z.string().min(1),
        kind: CHAIN_KIND_SCHEMA,
      },
    },
    async (args: { runId: string; kind: 'chain' }) => formatToolResult(handlers.getStatus(normalizeGetStatusInput(args))),
  )

  server.registerTool(
    'get_budget',
    {
      description: 'Get the tracked budget state for a Phase 2 chain.',
      inputSchema: {
        runId: z.string().min(1),
        kind: CHAIN_KIND_SCHEMA,
      },
    },
    async (args: { runId: string; kind: 'chain' }) => formatToolResult(handlers.getBudget(normalizeGetBudgetInput(args))),
  )

  server.registerTool(
    'retry_step',
    {
      description: 'Retry a failed chain step if retries remain.',
      inputSchema: {
        runId: z.string().min(1),
        kind: CHAIN_KIND_SCHEMA,
        stepId: z.string().min(1),
        reason: z.string().optional(),
      },
    },
    async (args: ToolArgs) => formatToolResult(handlers.retryStep(normalizeRetryStepInput(args))),
  )

  server.registerTool(
    'escalate_step',
    {
      description: 'Reassign a chain step to a different agent.',
      inputSchema: {
        runId: z.string().min(1),
        kind: CHAIN_KIND_SCHEMA,
        stepId: z.string().min(1),
        targetAgent: z.string().min(1),
        domainSkill: z.string().optional(),
        modeSkill: z.string().optional(),
        reason: z.string().optional(),
      },
    },
    async (args: ToolArgs) => formatToolResult(handlers.escalateStep(normalizeEscalateStepInput(args))),
  )

  server.registerTool(
    'handoff',
    {
      description: 'Persist a resumable handoff document for a running chain.',
      inputSchema: {
        runId: z.string().min(1),
        kind: CHAIN_KIND_SCHEMA,
        summary: z.string().optional(),
        recipient: z.string().optional(),
        includeArtifacts: z.boolean().optional(),
      },
    },
    async (args: ToolArgs) => formatToolResult(handlers.handoff(normalizeHandoffInput(args))),
  )

  return { server, handlers }
}

export async function startStdioServer(options: ToolHandlerOptions): Promise<OrchestratorServerContext> {
  const context = createOrchestratorServer(options)
  const transport = new StdioServerTransport()
  await context.server.connect(transport)
  return context
}

function formatToolResult(data: unknown): { content: Array<{ type: 'text'; text: string }>; structuredContent: Record<string, unknown> | undefined } {
  return {
    content: [
      {
        type: 'text' as const,
        text: JSON.stringify(data, null, 2),
      },
    ],
    structuredContent: toStructuredContent(data),
  }
}
