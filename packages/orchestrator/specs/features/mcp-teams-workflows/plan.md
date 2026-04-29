# MCP Teams, Workflows, and Polymorphic Run Tools Plan

## Summary

Expose the already-implemented team and workflow orchestration engines through the MCP server by registering team/workflow tools in `src/server.ts`, adding input normalizers that exactly match the handler input types in `src/tool-handlers.ts`, and replacing chain-only MCP schemas for generic run tools with a polymorphic `RUN_KIND_SCHEMA`.

This plan intentionally keeps the orchestration engines (`team-machine.ts`, `workflow-machine.ts`) as the source of execution semantics and limits server changes to schema validation, normalization, and tool registration. Handler changes are required only where current handler interfaces are still chain-only (`retryStep`, `escalateStep`, `handoff`).

## Inputs Read

- `docs/agent-swarm-implementation.md`
- `src/server.ts`
- `src/tool-handlers.ts`
- Supporting context: `src/types.ts`, `src/team-machine.ts`, `src/workflow-machine.ts`, `src/persistence.ts`, `src/__tests__/tool-handlers.case.ts`, `package.json`, active constitution at `../../specs/001-store-and-errors/constitution.md`

## Technical Context

| Area | Current State | Planned Use |
|------|---------------|-------------|
| Language/runtime | TypeScript ESM, Node `>=20.12.0` | Preserve ESM imports and type-only imports. |
| MCP layer | `@modelcontextprotocol/sdk` `^1.18.1`; tools registered in `src/server.ts` | Add registrations through existing `server.registerTool(...)` pattern. |
| Validation | `zod` `^3.25.76` | Replace `CHAIN_KIND_SCHEMA` with `RUN_KIND_SCHEMA`; add shared schemas for budget, usage, context, recovery. |
| Persistence | SQLite-backed persistence helpers in `src/persistence.ts` | Reuse `saveTeamState`, `saveWorkflowState`, `saveHandoff`; no schema migration expected. |
| Existing engines | `buildTeam`, `assignTask`, `completeTask`, `startWorkflow`, `advanceWorkflow` already exist on `OrchestratorToolHandlers` | MCP server should call these directly after normalization. |
| Tests | Vitest via `npm run test`; typecheck via `npm run typecheck`; build via `npm run build` | Add handler coverage for polymorphic tools and MCP registration/normalization coverage where feasible. |

## Current Gap

`src/tool-handlers.ts` already exposes these methods:

- `buildTeam(input: BuildTeamInput)`
- `assignTask(input: AssignTaskInput)`
- `completeTask(input: CompleteTaskInput)`
- `startWorkflow(input: StartWorkflowInput)`
- `advanceWorkflow(input: AdvanceWorkflowInput)`
- `getStatus(input: GetStatusInput)` and `getBudget(input: GetBudgetInput)` already support `chain | team | workflow`

But `src/server.ts` currently:

- Registers no MCP tools for team or workflow lifecycle operations.
- Defines `CHAIN_KIND_SCHEMA = z.literal('chain')` and uses it for `get_status`, `get_budget`, `retry_step`, `escalate_step`, and `handoff`.
- Normalizes `kind` for retry/escalate/handoff as `'chain'` only.

## Architecture Decision

### Options Considered

#### Option A: Register only team/workflow start/advance tools; leave existing generic tools chain-only

- Pros: Smallest diff; lowest risk.
- Cons: Contradicts the requested polymorphic behavior and the handler capabilities already present for status/budget.
- Rejected because `get_status` and `get_budget` already have polymorphic handler support and should not remain artificially chain-only.

#### Option B: Make `server.ts` polymorphic while leaving `retryStep`, `escalateStep`, and `handoff` chain-only internally

- Pros: MCP schemas become broader quickly.
- Cons: Runtime would accept `kind: 'team' | 'workflow'` but fail or misroute; this is misleading and not truly polymorphic.
- Rejected because input schemas must align with handler behavior, not overpromise support.

#### Option C: Register team/workflow tools and extend handler-level generic run operations by `kind`

- Pros: End-to-end MCP behavior matches schemas; existing engines are reused; server remains a thin normalization layer.
- Cons: Requires modest changes beyond `server.ts` in `tool-handlers.ts` and `types.ts` for polymorphic handoff/retry/escalate.
- Decision: **Use Option C**. It is the simplest complete implementation that satisfies the request without creating a parallel orchestration path.

## Server Schema Changes (`src/server.ts`)

### Replace Chain-only Kind Schema

Remove:

```ts
const CHAIN_KIND_SCHEMA = z.literal('chain')
```

Add:

```ts
const RUN_KIND_SCHEMA = z.enum(['chain', 'team', 'workflow'])
```

Use `RUN_KIND_SCHEMA` for:

- `get_status`
- `get_budget`
- `retry_step`
- `escalate_step`
- `handoff`

### Add Shared Server Schemas

Add these helper schemas near existing constants to avoid duplicating shapes:

```ts
const BUDGET_SCHEMA = z.record(z.unknown())

const STEP_USAGE_SCHEMA = z.object({
  inputTokens: z.number().nonnegative().optional(),
  outputTokens: z.number().nonnegative().optional(),
  totalTokens: z.number().nonnegative().optional(),
  costUsd: z.number().nonnegative().optional(),
  wallClockMs: z.number().nonnegative().optional(),
})

const ROOT_CONTEXT_SCHEMA = z.object({
  prompt: z.string().optional(),
  constraints: z.array(z.string()).optional(),
  allowedTools: z.array(z.string()).optional(),
  modelHint: z.string().optional(),
  approvalPolicy: z.enum(['minimal', 'normal', 'strict']).optional(),
})

const START_CONTEXT_SCHEMA = z.object({
  cliTool: HOST_CLI_SCHEMA.optional(),
  rootContext: ROOT_CONTEXT_SCHEMA.optional(),
  project: z.record(z.unknown()).optional(),
})

const TEAM_TASK_OUTCOME_SCHEMA = z.enum(['success', 'failure'])

const WORKFLOW_RECOVERY_SCHEMA = z.object({
  type: z.enum(['retry', 'escalate', 'handoff']),
  targetPhaseId: z.string().min(1).optional(),
  reason: z.string().optional(),
  recipient: z.string().optional(),
  summary: z.string().optional(),
})
```

Notes:

- `START_CONTEXT_SCHEMA` must match `StartChainInput['context']`, which is also used by `StartWorkflowInput['context']`.
- `BUDGET_SCHEMA` remains `z.record(z.unknown())` in the server because handler types accept `Partial<BudgetPolicy>` and runtime budget validation already happens downstream via `buildBudgetPolicy(...)`.
- Do not introduce new runtime dependencies.

## Exact MCP Tools to Register in `server.ts`

Register these new tools after `advance_chain` and before generic run tools (`get_status`, `get_budget`, `retry_step`, `escalate_step`, `handoff`) so lifecycle tools are grouped before cross-run tools.

### `build_team`

Handler call:

```ts
handlers.buildTeam(normalizeBuildTeamInput(args))
```

Input schema:

```ts
inputSchema: {
  team: z.string().min(1),
  task: z.string().min(1),
  budget: BUDGET_SCHEMA.optional(),
}
```

Normalization target: `BuildTeamInput`

```ts
{
  team: args.team as string,
  task: args.task as string,
  ...(isStructuredContent(args.budget) ? { budget: args.budget as Partial<BudgetPolicy> } : {}),
}
```

Expected result from handler: `{ teamId, state, readyTaskIds, tasks, budget }`.

### `assign_team_task`

Handler call:

```ts
handlers.assignTask(normalizeAssignTaskInput(args))
```

Input schema:

```ts
inputSchema: {
  teamId: z.string().min(1),
  taskId: z.string().min(1),
  assignee: z.string().min(1),
  claim: z.boolean().optional(),
}
```

Normalization target: `AssignTaskInput`

```ts
{
  teamId: args.teamId as string,
  taskId: args.taskId as string,
  assignee: args.assignee as string,
  ...(typeof args.claim === 'boolean' ? { claim: args.claim } : {}),
}
```

Expected result from handler: `{ teamId, state, readyTaskIds, task }`.

### `complete_team_task`

Handler call:

```ts
handlers.completeTask(normalizeCompleteTaskInput(args))
```

Input schema:

```ts
inputSchema: {
  teamId: z.string().min(1),
  taskId: z.string().min(1),
  outcome: TEAM_TASK_OUTCOME_SCHEMA,
  result: z.record(z.unknown()).optional(),
  usage: STEP_USAGE_SCHEMA.optional(),
  error: z.record(z.unknown()).optional(),
}
```

Normalization target: `CompleteTaskInput`

```ts
{
  teamId: args.teamId as string,
  taskId: args.taskId as string,
  outcome: args.outcome as 'success' | 'failure',
  ...(isStructuredContent(args.result) ? { result: args.result } : {}),
  ...(isStructuredContent(args.usage) ? { usage: definedEntries(args.usage) as StepUsage } : {}),
  ...(isStructuredContent(args.error) ? { error: args.error as StructuredError } : {}),
}
```

Expected result from handler: `{ teamId, state, readyTaskIds, budget, summary, evaluation }`. When member tasks complete and the synthesis task becomes ready, `readyTaskIds` must include the synthesis task id.

### `start_workflow`

Handler call:

```ts
handlers.startWorkflow(normalizeStartWorkflowInput(args))
```

Input schema:

```ts
inputSchema: {
  workflow: z.string().min(1),
  task: z.string().min(1),
  domainSkill: z.string().min(1).optional(),
  modeSkill: z.string().min(1).optional(),
  budget: BUDGET_SCHEMA.optional(),
  context: START_CONTEXT_SCHEMA.optional(),
}
```

Normalization target: `StartWorkflowInput`

```ts
{
  workflow: args.workflow as string,
  task: args.task as string,
  ...(typeof args.domainSkill === 'string' ? { domainSkill: args.domainSkill } : {}),
  ...(typeof args.modeSkill === 'string' ? { modeSkill: args.modeSkill } : {}),
  ...(isStructuredContent(args.budget) ? { budget: args.budget as Partial<BudgetPolicy> } : {}),
  ...(isStructuredContent(args.context) ? { context: normalizeStartContext(args.context) } : {}),
}
```

Expected result from handler: `{ workflowId, state, currentPhase, budget }`. If the entry phase launches a child chain/team, `currentPhase.childRun` should describe that child.

### `advance_workflow`

Handler call:

```ts
handlers.advanceWorkflow(normalizeAdvanceWorkflowInput(args))
```

Input schema:

```ts
inputSchema: {
  workflowId: z.string().min(1),
  outcome: z.string().min(1).optional(),
  recovery: WORKFLOW_RECOVERY_SCHEMA.optional(),
}
```

Normalization target: `AdvanceWorkflowInput`

```ts
{
  workflowId: args.workflowId as string,
  ...(typeof args.outcome === 'string' ? { outcome: args.outcome } : {}),
  ...(isStructuredContent(args.recovery) ? { recovery: normalizeWorkflowRecovery(args.recovery) } : {}),
}
```

Expected result from handler: `{ workflowId, state, currentPhase, budget, recoveryOptions, error? }`.

## Existing Tools to Make Polymorphic

### `get_status`

Current server schema uses `CHAIN_KIND_SCHEMA`; change it to:

```ts
inputSchema: {
  runId: z.string().min(1),
  kind: RUN_KIND_SCHEMA,
}
```

Normalize to `GetStatusInput`:

```ts
function normalizeGetStatusInput(args: Record<string, unknown>): GetStatusInput {
  return { runId: args.runId as string, kind: args.kind as RunKind }
}
```

No handler change is needed because `getStatus` already switches on `chain | team | workflow`.

### `get_budget`

Current server schema uses `CHAIN_KIND_SCHEMA`; change it to:

```ts
inputSchema: {
  runId: z.string().min(1),
  kind: RUN_KIND_SCHEMA,
}
```

Normalize to `GetBudgetInput`:

```ts
function normalizeGetBudgetInput(args: Record<string, unknown>): GetBudgetInput {
  return { runId: args.runId as string, kind: args.kind as RunKind }
}
```

No handler change is needed because `getBudget` already switches on `chain | team | workflow`.

### `retry_step`

Schema change:

```ts
inputSchema: {
  runId: z.string().min(1),
  kind: RUN_KIND_SCHEMA,
  stepId: z.string().min(1),
  reason: z.string().optional(),
}
```

Normalize to updated `RetryStepInput`:

```ts
{
  runId: args.runId as string,
  kind: args.kind as RunKind,
  stepId: args.stepId as string,
  ...(typeof args.reason === 'string' ? { reason: args.reason } : {}),
}
```

Handler changes required:

- Update `RetryStepInput.kind` from `'chain'` to `RunKind`.
- Keep existing chain behavior unchanged.
- Add team behavior: treat `stepId` as `TeamTaskState.taskId`; retry a failed task by moving it back to `pending`, clearing failure/completion fields, recomputing `readyTaskIds`, incrementing budget retry count with `updateBudget({ retryIncrement: 1, stepId })`, saving via `saveTeamState`, and returning `{ runId, stepId, state, readyTaskIds, attemptsRemaining: null }`.
- Add workflow behavior: treat `stepId` as a workflow phase id and validate it matches `state.currentPhaseId` when present; call `advanceWorkflow({ workflowId: runId, recovery: { type: 'retry', reason } })`; return `{ runId, stepId, state, currentPhase, budget }`.

### `escalate_step`

Schema change:

```ts
inputSchema: {
  runId: z.string().min(1),
  kind: RUN_KIND_SCHEMA,
  stepId: z.string().min(1),
  targetAgent: z.string().min(1),
  targetPhaseId: z.string().min(1).optional(),
  domainSkill: z.string().optional(),
  modeSkill: z.string().optional(),
  reason: z.string().optional(),
}
```

Normalize to updated `EscalateStepInput`:

```ts
{
  runId: args.runId as string,
  kind: args.kind as RunKind,
  stepId: args.stepId as string,
  targetAgent: args.targetAgent as string,
  ...(typeof args.targetPhaseId === 'string' ? { targetPhaseId: args.targetPhaseId } : {}),
  ...(typeof args.domainSkill === 'string' ? { domainSkill: args.domainSkill } : {}),
  ...(typeof args.modeSkill === 'string' ? { modeSkill: args.modeSkill } : {}),
  ...(typeof args.reason === 'string' ? { reason: args.reason } : {}),
}
```

Handler changes required:

- Update `EscalateStepInput.kind` from `'chain'` to `RunKind`.
- Add optional `targetPhaseId?: string` to `EscalateStepInput`.
- Keep existing chain behavior unchanged; `targetAgent`, `domainSkill`, and `modeSkill` keep their current meaning.
- Add team behavior: treat `stepId` as `TeamTaskState.taskId`; update that task's `agent` to `targetAgent`, clear assignment/claim/failure fields, set task state to `pending`, recompute `readyTaskIds`, save via `saveTeamState`, and return `{ runId, stepId, state, readyTaskIds, newAssignment: task }`.
- Add workflow behavior: treat `stepId` as the current workflow phase id; set recovery target to `targetPhaseId ?? targetAgent` because workflow escalation routes to a phase, not an agent; call `advanceWorkflow({ workflowId: runId, recovery: { type: 'escalate', targetPhaseId: targetPhaseId ?? targetAgent, reason } })`; return `{ runId, stepId, state, currentPhase, budget }`.

### `handoff`

Schema change:

```ts
inputSchema: {
  runId: z.string().min(1),
  kind: RUN_KIND_SCHEMA,
  summary: z.string().optional(),
  recipient: z.string().optional(),
  includeArtifacts: z.boolean().optional(),
}
```

Normalize to updated `HandoffInput`:

```ts
{
  runId: args.runId as string,
  kind: args.kind as RunKind,
  ...(typeof args.summary === 'string' ? { summary: args.summary } : {}),
  ...(typeof args.recipient === 'string' ? { recipient: args.recipient } : {}),
  ...(typeof args.includeArtifacts === 'boolean' ? { includeArtifacts: args.includeArtifacts } : {}),
}
```

Handler changes required:

- Update `HandoffInput.kind` from `'chain'` to `RunKind`.
- Update `HandoffDocument.status` from `ChainState` to `ChainState | TeamState | WorkflowState`.
- Make `HandoffDocument.plan` optional because teams/workflows do not have a single `ExecutionPlan`.
- Keep current chain behavior, including saving the chain state as `handoff` and including the execution plan to preserve existing handoff detail.
- Add team behavior: load `TeamState`, set `state.state = 'handoff'`, update `updatedAt`, create a handoff with `kind: 'team'`, save via `saveHandoff`, then save the team state.
- Add workflow behavior: load `WorkflowState`, set `state.state = 'handoff'`, set `handoffSummary` from input summary when present, update `updatedAt`, create a handoff with `kind: 'workflow'`, save via `saveHandoff`, then save the workflow state.
- Do not add new artifact collection. `includeArtifacts` should remain accepted for API compatibility; the implementation should not invent artifact discovery beyond the existing chain plan snapshot.

## Normalization Refactor in `server.ts`

To avoid drift, extract the context normalization currently embedded in `normalizeStartChainInput`:

```ts
function normalizeStartContext(args: StructuredContent): NonNullable<StartChainInput['context']> {
  const nextContext: NonNullable<StartChainInput['context']> = {}
  if (typeof args.cliTool === 'string') nextContext.cliTool = args.cliTool as HostCli
  if (isStructuredContent(args.rootContext)) {
    nextContext.rootContext = definedEntries(args.rootContext) as NonNullable<NonNullable<StartChainInput['context']>['rootContext']>
  }
  if (isStructuredContent(args.project)) {
    nextContext.project = definedEntries(args.project) as NonNullable<NonNullable<StartChainInput['context']>['project']>
  }
  return nextContext
}
```

Then use it from both `normalizeStartChainInput` and `normalizeStartWorkflowInput`.

## Type Imports Needed in `server.ts`

Extend the existing type import from `./types.js` to include:

```ts
type AdvanceWorkflowInput,
type AssignTaskInput,
type BuildTeamInput,
type CompleteTaskInput,
type RunKind,
type StartWorkflowInput,
type StructuredError,
```

`BudgetPolicy` may also be imported as a type if normalizers cast `budget` explicitly to `Partial<BudgetPolicy>`.

## Files to Change During Implementation

| File | Required Changes |
|------|------------------|
| `src/server.ts` | Add schemas, normalizers, and MCP registrations; replace `CHAIN_KIND_SCHEMA` usages with `RUN_KIND_SCHEMA`. |
| `src/tool-handlers.ts` | Extend `RetryStepInput`, `EscalateStepInput`, `HandoffInput` behavior by `kind`; keep chain behavior unchanged. |
| `src/types.ts` | Update `RetryStepInput`, `EscalateStepInput`, `HandoffInput`, and `HandoffDocument` types if these interfaces stay centralized there; currently some input interfaces live in `tool-handlers.ts`, while `HandoffDocument` lives in `types.ts`. |
| `src/team-machine.ts` | Prefer adding small helpers for retry/escalate team task state transitions rather than mutating team state ad hoc in `tool-handlers.ts`. |
| `src/__tests__/tool-handlers.case.ts` | Add polymorphic handler tests for team/workflow retry, escalation, handoff, status, and budget. |
| `src/__tests__/server.case.ts` or equivalent | Add MCP registration/normalization tests if the SDK exposes a stable way to call registered tools; otherwise rely on typecheck plus handler tests and a manual MCP smoke checklist. |

## Implementation Phases

### Phase 1: Server schema and normalizer foundation

1. Replace `CHAIN_KIND_SCHEMA` with `RUN_KIND_SCHEMA`.
2. Add shared schemas: `BUDGET_SCHEMA`, `STEP_USAGE_SCHEMA`, `ROOT_CONTEXT_SCHEMA`, `START_CONTEXT_SCHEMA`, `TEAM_TASK_OUTCOME_SCHEMA`, `WORKFLOW_RECOVERY_SCHEMA`.
3. Extract `normalizeStartContext` and update `normalizeStartChainInput` to use it.
4. Add normalizers for build/assign/complete team and start/advance workflow.

Done when: `src/server.ts` typechecks with the new normalizers and no registered tool behavior has changed yet.

### Phase 2: Register team and workflow MCP tools

1. Register `build_team`.
2. Register `assign_team_task`.
3. Register `complete_team_task`.
4. Register `start_workflow`.
5. Register `advance_workflow`.

Done when: each new MCP tool calls the matching `OrchestratorToolHandlers` method with the normalizer listed above.

### Phase 3: Make status and budget polymorphic in `server.ts`

1. Change `get_status.kind` to `RUN_KIND_SCHEMA`.
2. Change `get_budget.kind` to `RUN_KIND_SCHEMA`.
3. Update their normalizers to return `kind: RunKind`.
4. Update descriptions from “Phase 2 chain” to “chain, team, or workflow run”.

Done when: `get_status` and `get_budget` can pass `team` and `workflow` kinds to the already-polymorphic handlers.

### Phase 4: Make retry, escalation, and handoff polymorphic end to end

1. Update server schemas and normalizers for `retry_step`, `escalate_step`, and `handoff`.
2. Update handler input interfaces in `src/tool-handlers.ts`.
3. Add team retry/escalation state-transition helpers or equivalent minimal logic.
4. Add workflow retry/escalation routing through `advanceWorkflow(...)` recovery decisions.
5. Add team/workflow handoff branches and update `HandoffDocument` typing.

Done when: all five generic run tools accept `chain | team | workflow` and either perform kind-specific behavior or route to existing workflow recovery semantics.

### Phase 5: Verification

1. Add/extend Vitest tests for handler behavior.
2. If practical, add server registration/normalization tests.
3. Run:
   - `npm run typecheck`
   - `npm run test`
   - `npm run build`

Done when: all gates pass and tests prove server inputs align with handler inputs.

## Test Plan

### Handler tests

Extend `src/__tests__/tool-handlers.case.ts` with cases for:

- `getStatus({ kind: 'team' })` after `buildTeam(...)` returns team current tasks.
- `getBudget({ kind: 'team' })` after completing a team task reflects token usage.
- `getStatus({ kind: 'workflow' })` and `getBudget({ kind: 'workflow' })` remain covered and should be asserted more explicitly.
- `handoff({ kind: 'team' })` sets team state to `handoff` and saves a handoff document.
- `handoff({ kind: 'workflow' })` sets workflow state to `handoff` and saves a handoff document.
- `retryStep({ kind: 'team' })` moves a failed team task back to ready/pending and updates budget retry count.
- `escalateStep({ kind: 'team' })` changes the task agent to `targetAgent` and makes it assignable again.
- `retryStep({ kind: 'workflow' })` routes through workflow recovery and launches a replacement child run.
- `escalateStep({ kind: 'workflow', targetPhaseId })` routes through workflow recovery to the target phase.
- Existing chain retry/escalate/handoff tests continue to pass unchanged.

### Server tests

Add `src/__tests__/server.case.ts` if MCP SDK internals allow stable invocation of registered handlers. The test should verify at least:

- `build_team`, `assign_team_task`, `complete_team_task`, `start_workflow`, and `advance_workflow` are registered.
- `get_status`, `get_budget`, `handoff`, `retry_step`, and `escalate_step` schemas accept `team` and `workflow` kinds.
- Normalizers omit optional fields when not provided and preserve structured `budget`, `context`, `usage`, `result`, `error`, and `recovery` objects.

If stable server-level invocation is not available, keep coverage at handler level and use `npm run typecheck` as the contract check for `server.ts` normalizers.

## Acceptance Criteria

- MCP exposes these new tools: `build_team`, `assign_team_task`, `complete_team_task`, `start_workflow`, `advance_workflow`.
- MCP generic tools `get_status`, `get_budget`, `handoff`, `retry_step`, and `escalate_step` accept `kind: 'chain' | 'team' | 'workflow'`.
- New `server.ts` normalizers produce inputs matching the handler interfaces exactly.
- Existing chain tool behavior remains backward compatible.
- Team runs can be built, assigned, completed, queried, budget-checked, retried, escalated, and handed off through MCP.
- Workflow runs can be started, advanced, queried, budget-checked, retried/escalated through recovery semantics, and handed off through MCP.
- `npm run typecheck`, `npm run test`, and `npm run build` pass in `packages/orchestrator`.

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Server schema accepts a kind the handler does not truly support | MCP callers get runtime failures | Extend handler branches before widening schemas for retry/escalate/handoff. |
| Workflow escalation naming mismatch (`targetAgent` vs phase routing) | Confusing API | Add optional `targetPhaseId`; for backward compatibility, use `targetAgent` as fallback phase id for workflow only. |
| Team retry/escalate invents too much workflow policy | Scope creep | Implement only minimal task reset/reassignment; no max retry policy unless already present. |
| Handoff document assumes chain-only `plan` | Type/runtime errors for teams/workflows | Make `plan` optional and `status` a union of run state types. |
| Duplicate context schema/normalization drifts | Different behavior between chain and workflow starts | Extract `START_CONTEXT_SCHEMA` and `normalizeStartContext`. |

## Constitution Check

### Active Constitution (`../../specs/001-store-and-errors/constitution.md`)

- P1 Zod as single source for runtime validation: satisfied in `server.ts` by using zod schemas for MCP inputs; no parallel runtime validators.
- P4/P5 error discipline: no `process.exit`; handler errors continue to throw to the MCP boundary.
- P8 backward compatibility: preserve existing chain tool names, required fields, and handler behavior.
- C1 ESM-only: use existing `.js` ESM import style and type-only imports.
- C4 zero new runtime dependencies: no new dependencies.
- C5 quality gates: plan requires `npm run typecheck`, `npm run test`, `npm run build`.

### Planner Article Checklist

- Article I Library-First: reuse MCP SDK, zod, existing handlers, team/workflow machines, persistence helpers.
- Article II TDD: add handler tests before/alongside behavior changes; typecheck server normalizers.
- Article III Docs: this plan aligns the MCP layer with `docs/agent-swarm-implementation.md` and current handler methods.
- Article IV YAGNI: no new orchestration engine, no new artifact collection, no new approval protocol beyond existing budget return data.
- Article V Simplicity: server remains a thin schema/normalization layer; orchestration semantics stay in handlers/machines.
- Article VI Anti-Overengineering: shared schemas/helpers only where duplication already exists; no generic framework for all future run kinds.

## Out of Scope

- Adding a new MCP “user approval” protocol for team budget gates. `build_team` should surface `budget` as returned by the handler; a future change can define a concrete approval contract if needed.
- Changing catalog definition formats.
- Adding new persistence tables or migrations.
- Changing queue/worker behavior for background execution.
- Renaming existing chain tools.
