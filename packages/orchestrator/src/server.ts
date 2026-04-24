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
  type InvokeAgentInput,
  type ListCatalogInput,
  type RetryStepInput,
  type ToolHandlerOptions,
} from './tool-handlers.js'
import { CatalogToolHandlers } from './catalog-tools.js'
import { getPersistenceDb } from './persistence.js'
import type { AdvanceChainInput, HostCli, StartChainInput, StepOutputContract, StepUsage } from './types.js'
import { getEventBus } from './events/bus.js'
import { JobQueue } from './queue/queue.js'
import { startQueueWorker } from './queue/worker.js'
import { registerBuiltinHandlers } from './queue/handlers.js'

const CATALOG_KIND_EXTENDED_SCHEMA = z.enum(['agent', 'skill', 'chain', 'team', 'workflow', 'mode', 'command'])

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
  const catalogHandlers = new CatalogToolHandlers(getPersistenceDb())
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

  // -----------------------------------------------------------------------
  // Catalog management tools
  // -----------------------------------------------------------------------

  server.registerTool(
    'catalog_list',
    {
      description: 'List internal versioned catalog definitions.',
      inputSchema: { kind: CATALOG_KIND_EXTENDED_SCHEMA.optional() },
    },
    async (args: ToolArgs) => formatToolResult(catalogHandlers.catalogList(args.kind ? { kind: args.kind as never } : {})),
  )

  server.registerTool(
    'catalog_list_versions',
    {
      description: 'List all immutable versions of a definition.',
      inputSchema: { kind: CATALOG_KIND_EXTENDED_SCHEMA, name: z.string().min(1) },
    },
    async (args: ToolArgs) => formatToolResult(catalogHandlers.catalogListVersions({ kind: args.kind as never, name: args.name as string })),
  )

  server.registerTool(
    'catalog_get_version',
    {
      description: 'Get a catalog definition by kind/name, optionally pinning a version. Defaults to active version.',
      inputSchema: {
        kind: CATALOG_KIND_EXTENDED_SCHEMA,
        name: z.string().min(1),
        version: z.number().int().positive().optional(),
      },
    },
    async (args: ToolArgs) => {
      const version = args.version as number | undefined
      const input = version !== undefined
        ? { kind: args.kind as never, name: args.name as string, version }
        : { kind: args.kind as never, name: args.name as string }
      return formatToolResult(catalogHandlers.catalogGetVersion(input))
    },
  )

  server.registerTool(
    'catalog_create_version',
    {
      description: 'Create a new immutable version of a definition. Checksum-deduplication — identical content is a no-op.',
      inputSchema: {
        kind: CATALOG_KIND_EXTENDED_SCHEMA,
        name: z.string().min(1),
        frontmatter: z.record(z.unknown()),
        body: z.string(),
        createdBy: z.string().optional(),
      },
    },
    async (args: ToolArgs) => {
      const createdBy = args.createdBy as string | undefined
      return formatToolResult(catalogHandlers.catalogCreateVersion({
        kind: args.kind as never,
        name: args.name as string,
        frontmatter: args.frontmatter as Record<string, unknown>,
        body: args.body as string,
        ...(createdBy !== undefined ? { createdBy } : {}),
      }))
    },
  )

  server.registerTool(
    'catalog_set_active',
    {
      description: 'Move the active version pointer for a definition.',
      inputSchema: {
        kind: CATALOG_KIND_EXTENDED_SCHEMA,
        name: z.string().min(1),
        version: z.number().int().positive(),
      },
    },
    async (args: ToolArgs) => formatToolResult(catalogHandlers.catalogSetActive({
      kind: args.kind as never,
      name: args.name as string,
      version: args.version as number,
    })),
  )

  server.registerTool(
    'catalog_diff',
    {
      description: 'Compare two versions of a definition (returns both side-by-side).',
      inputSchema: {
        kind: CATALOG_KIND_EXTENDED_SCHEMA,
        name: z.string().min(1),
        fromVersion: z.number().int().positive(),
        toVersion: z.number().int().positive(),
      },
    },
    async (args: ToolArgs) => formatToolResult(catalogHandlers.catalogDiff({
      kind: args.kind as never,
      name: args.name as string,
      fromVersion: args.fromVersion as number,
      toVersion: args.toVersion as number,
    })),
  )

  server.registerTool(
    'catalog_export_version',
    {
      description: 'Write a catalog definition version body to a file path. The only orchestrator-initiated write to a host catalog path; always explicit and user-driven.',
      inputSchema: {
        kind: CATALOG_KIND_EXTENDED_SCHEMA,
        name: z.string().min(1),
        targetPath: z.string().min(1),
        version: z.number().int().positive().optional(),
      },
    },
    async (args: ToolArgs) => {
      const version = args.version as number | undefined
      return formatToolResult(catalogHandlers.catalogExportVersion({
        kind: args.kind as never,
        name: args.name as string,
        targetPath: args.targetPath as string,
        ...(version !== undefined ? { version } : {}),
      }))
    },
  )

  server.registerTool(
    'catalog_import',
    {
      description: 'Bulk-import definitions from host config files or the library path.',
      inputSchema: {
        hosts: z.array(z.enum(['opencode', 'claude-code'])).optional(),
        libraryOrchestrationRoot: z.string().optional(),
        libraryAgentsRoot: z.string().optional(),
        projectRoot: z.string().optional(),
      },
    },
    async (args: ToolArgs) => formatToolResult(catalogHandlers.catalogImport(args as never)),
  )

  // -----------------------------------------------------------------------
  // Agent invocation
  // -----------------------------------------------------------------------

  server.registerTool(
    'invoke_agent',
    {
      description:
        'Resolve a named agent from the catalog and return a fully composed prompt spec ready for execution. ' +
        'Emits an agent.invoked event so watchers are notified. The caller is responsible for running the agent.',
      inputSchema: {
        agent: z.string().min(1),
        task: z.string().min(1),
        version: z.number().int().positive().optional(),
        domainSkill: z.string().min(1).optional(),
        modeSkill: z.string().min(1).optional(),
      },
    },
    async (args: ToolArgs) => {
      const input: InvokeAgentInput = {
        agent: args.agent as string,
        task: args.task as string,
        ...(typeof args.version === 'number' ? { version: args.version } : {}),
        ...(typeof args.domainSkill === 'string' ? { domainSkill: args.domainSkill } : {}),
        ...(typeof args.modeSkill === 'string' ? { modeSkill: args.modeSkill } : {}),
      }
      return formatToolResult(handlers.invokeAgent(input))
    },
  )

  // -----------------------------------------------------------------------
  // Watch tools
  // -----------------------------------------------------------------------

  server.registerTool(
    'subscribe_run',
    {
      description:
        'Subscribe to state-change events for a run (chain / team / workflow). ' +
        'Returns immediately with past events and then sends MCP log notifications ' +
        'for each future event until unsubscribe_run is called or the run completes.',
      inputSchema: {
        runId: z.string().min(1),
        sinceEventId: z.number().int().nonnegative().optional(),
      },
    },
    async (args: ToolArgs) => {
      const runId = args.runId as string
      const sinceEventId = args.sinceEventId as number | undefined
      const db = getPersistenceDb()
      const bus = getEventBus()

      // sinceEventId=0 means "from start"; undefined also means "from start"
      const past = bus.replayFromDb(db, runId, sinceEventId)

      bus.onRun(runId, (event) => {
        server.server.sendLoggingMessage({
          level: 'info',
          data: { _type: 'run_event', ...event },
        }).catch(() => { /* client may have disconnected */ })
      })

      return formatToolResult({
        subscribed: true,
        runId,
        pastEvents: past,
      })
    },
  )

  server.registerTool(
    'unsubscribe_run',
    {
      description: 'Remove all subscriptions for a run ID on this connection.',
      inputSchema: { runId: z.string().min(1) },
    },
    async (args: ToolArgs) => {
      const runId = args.runId as string
      getEventBus().removeRunListeners(runId)
      return formatToolResult({ unsubscribed: true, runId })
    },
  )

  // -----------------------------------------------------------------------
  // Message queue tools
  // -----------------------------------------------------------------------

  server.registerTool(
    'enqueue_job',
    {
      description:
        'Enqueue a background job. The worker processes it asynchronously and emits a run_event ' +
        'notification on completion. Built-in job types: agent_invoke.',
      inputSchema: {
        jobType: z.string().min(1),
        payload: z.record(z.unknown()).optional(),
        priority: z.number().int().optional(),
        maxAttempts: z.number().int().positive().optional(),
        id: z.string().optional(),
      },
    },
    async (args: ToolArgs) => {
      const q = new JobQueue(getPersistenceDb())
      const job = q.enqueue({
        jobType: args.jobType as string,
        ...(args.payload !== undefined ? { payload: args.payload as Record<string, unknown> } : {}),
        ...(args.priority !== undefined ? { priority: args.priority as number } : {}),
        ...(args.maxAttempts !== undefined ? { maxAttempts: args.maxAttempts as number } : {}),
        ...(args.id !== undefined ? { id: args.id as string } : {}),
      })
      return formatToolResult({ jobId: job.id, status: job.status, createdAt: job.createdAt })
    },
  )

  server.registerTool(
    'get_job',
    {
      description: 'Get the current status and result of an enqueued job.',
      inputSchema: { jobId: z.string().min(1) },
    },
    async (args: ToolArgs) => {
      const q = new JobQueue(getPersistenceDb())
      const job = q.getJob(args.jobId as string)
      if (!job) return formatToolResult({ found: false })
      return formatToolResult({ found: true, job })
    },
  )

  server.registerTool(
    'list_jobs',
    {
      description: 'List queued jobs, optionally filtered by status.',
      inputSchema: {
        status: z.enum(['pending', 'claimed', 'completed', 'failed']).optional(),
        limit: z.number().int().positive().max(200).optional(),
      },
    },
    async (args: ToolArgs) => {
      const q = new JobQueue(getPersistenceDb())
      const status = args.status as 'pending' | 'claimed' | 'completed' | 'failed' | undefined
      const jobs = q.listJobs(status, (args.limit as number | undefined) ?? 50)
      return formatToolResult({ jobs, total: jobs.length })
    },
  )

  return { server, handlers }
}

export async function startStdioServer(options: ToolHandlerOptions): Promise<OrchestratorServerContext> {
  registerBuiltinHandlers(startQueueWorker({ db: getPersistenceDb() }))
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
