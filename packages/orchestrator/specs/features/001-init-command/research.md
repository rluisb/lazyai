# Research: Orchestrator `init` Command

**Date:** 2026-04-29
**Package:** `packages/orchestrator`
**Goal:** Add a one-shot CLI command that inspects available orchestration meta (agents, skills, commands/chains/teams/workflows, host CLI capabilities, root context files) and recommends the best orchestration shape for a user's task.

---

## 1. Confirmed Product Direction

User choices:

1. **Output target:** one-shot command, not interactive TUI.
2. **Meta source:** compare options; recommended foundation is discovered meta from the existing ai-setup/orchestrator loader, with optional fallback later.
3. **Recommendation engine:** rule-based MVP; LLM-powered/smart recommendations can be added as a later `--smart`/`--reason` mode.
4. **Side effects:** read-only report by default; optional `--scaffold` / `--start` can be layered later.
5. **CLI awareness:** recommendations should consider host CLI capability (parallel teams, subagents, structured output).

Recommended MVP:

```text
ai-setup-orchestrator init --task "build auth from scratch" --host claude-code
```

Outputs a structured capability report plus deterministic suggestions.

---

## 2. Existing Catalog Discovery Flow

### File-based catalog loading

`src/loader.ts` exposes:

- `getDefaultLibraryRoots()`
- `loadCatalog(options: LoaderOptions): OrchestrationCatalog`

`loadCatalog()` discovers definitions from:

| Definition kind | Library path | Project override path |
|---|---|---|
| agents | `library/agents` | `.ai/agents` |
| domain skills | `library/orchestration/skills/domains` | `.ai/orchestration/skills/domains` |
| mode skills | `library/orchestration/skills/modes` | `.ai/orchestration/skills/modes` |
| chains | `library/orchestration/chains` | `.ai/orchestration/chains` |
| teams | `library/orchestration/teams` | `.ai/orchestration/teams` |
| workflows | `library/orchestration/workflows` | `.ai/orchestration/workflows` |

Default library discovery walks upward from the compiled module location until it finds `library/mcp/catalog.json`.

### DB + host-file resolution

`src/catalog/resolver.ts` exposes:

- `resolveCatalog(base, { db, projectRoot, hostCli })`

Resolver priority comment:

1. base catalog (library + project files)
2. DB internal active versions
3. user-global host files
4. user-project host files

However, current implementation only merges **DB active agents and skills** into the base catalog:

```ts
const dbAgentDefs = store.listDefinitions('agent')
const dbSkillDefs = store.listDefinitions('skill')
```

It does **not yet merge active DB chains, teams, or workflows** into the runtime `OrchestrationCatalog`.

### Important gap discovered

The new internal catalog tools can create teams/chains/workflows, but `buildTeam`, `startChain`, and `startWorkflow` use `OrchestratorToolHandlers.getCatalog()`, which currently calls `resolveCatalog()`.

Because `resolveCatalog()` only merges DB agents/skills, a team created via `catalog_create_version(kind='team', ...)` may appear in `catalog_list`, but not be available to `build_team` unless it also exists as a file definition.

This impacts both:

- the previously created `pedro-ernesto-e-banda` team
- the future `init --scaffold` path if it stores teams/chains/workflows in the DB catalog

**Recommendation:** before or alongside `init`, extend `resolveCatalog()` to merge active DB chains, teams, and workflows.

---

## 3. Existing CLI Command Pattern

Main entrypoint: `src/index.ts`.

Current subcommands:

- `catalog`
- `invoke`
- `connect`
- `serve`
- `start`
- `tail`

Pattern:

```ts
if (subcommand === 'catalog') {
  await runCatalog(getPersistenceDb(), rest)
  return
}
```

CLI modules live in `src/cli/` and expose:

- `HELP` constant
- parse function for args when needed
- `runX(...)` function

`src/cli/catalog.ts` is the best model for testable one-shot output because `runCatalog(db, args, out = process.stdout)` allows injecting an output stream.

Recommended new module:

```text
src/cli/init.ts
```

Exports:

- `INIT_HELP`
- `parseInitArgs(args)`
- `runInit(options, out = process.stdout)`

Add to `src/index.ts`:

```ts
import { parseInitArgs, runInit, INIT_HELP } from './cli/init.js'

if (subcommand === 'init') { ... }
```

---

## 4. CLI Capability Context

`src/compiler.ts` exposes:

```ts
getCliContext(host: HostCli = 'opencode'): CliContext
```

`CliContext` fields:

```ts
interface CliContext {
  host: HostCli
  dispatchMode: DispatchMode
  supportsSubagents: boolean
  supportsParallelTeams: boolean
  supportsStructuredOutput: boolean
  mcpServerName: string
}
```

Current host capabilities:

| Host | dispatchMode | supportsSubagents | supportsParallelTeams | supportsStructuredOutput |
|---|---|---:|---:|---:|
| claude-code | task-tool | true | true | true |
| codex | native-subagent | true | false | true |
| opencode | task-tool | true | false | true |
| gemini | instruction-only | false | false | false |
| copilot | instruction-only | false | false | false |

Recommendation engine can start with a single important branch:

- If task appears multi-agent and `supportsParallelTeams === true`, prefer `team`.
- Otherwise prefer `chain` for multi-step work.

---

## 5. Existing Tool Handler Flow

`src/tool-handlers.ts` uses a private `getCatalog()` method that:

1. calls `loadCatalog()` using `projectRoot`, optional library roots
2. calls `resolveCatalog()` with DB + host CLI when supported

`listCatalog()` returns a flattened list of catalog items.

`buildTeam`, `startChain`, and `startWorkflow` call `getCatalog()` and require definitions by name.

Implication:

- If `init` wants to recommend definitions users can actually run, it should use the same discovery flow as handlers, not just `loadCatalog()`.
- Since `getCatalog()` is private, either:
  - add a public helper such as `loadRuntimeCatalog(options)`; or
  - implement `runInit()` using `loadCatalog()` + `resolveCatalog()` directly.

Recommended: extract shared helper:

```ts
export function loadRuntimeCatalog(options: ToolHandlerOptions): OrchestrationCatalog
```

Then use it from both `OrchestratorToolHandlers.getCatalog()` and `runInit()`.

This reduces drift.

---

## 6. Recommendation Engine MVP

Suggested rule-based output:

```ts
interface InitRecommendation {
  kind: 'direct-agent' | 'chain' | 'team' | 'workflow'
  name?: string
  confidence: 'low' | 'medium' | 'high'
  reason: string
  nextCommand?: string
  alternatives: Array<{ kind: string; name?: string; reason: string }>
}
```

Rules, in order:

1. If task contains review/audit/security and reviewer/red-team/team exists:
   - Prefer team when host supports parallel teams.
   - Otherwise suggest review chain or direct reviewer.
2. If task contains design/architecture/spec/from scratch/greenfield:
   - Prefer workflow/chain with architect/planner/implementor when available.
3. If task contains build/implement/refactor and multiple phases exist:
   - Prefer RPI workflow or feature chain.
4. If only one relevant agent exists:
   - Suggest direct `invoke_agent`.
5. If no task is provided:
   - Print capability inventory and generic examples only.

Example CLI-aware recommendation:

```text
Host: claude-code
Capabilities: parallel teams yes, structured output yes
Suggestion: team
Reason: Task looks multi-agent and host supports parallel team dispatch.
```

```text
Host: opencode
Capabilities: parallel teams no, structured output yes
Suggestion: chain
Reason: Task is multi-step but host does not support parallel teams; sequential RPI is safer.
```

---

## 7. Read-only vs scaffold vs start

MVP should be read-only:

```bash
ai-setup-orchestrator init --task "review auth code"
```

Future flags:

```bash
ai-setup-orchestrator init --task "review auth code" --scaffold
ai-setup-orchestrator init --task "review auth code" --scaffold --start
```

If `--scaffold` uses the DB internal catalog, runtime catalog must first support DB chains/teams/workflows.

If `--scaffold` writes JSON files to `.ai/orchestration`, that avoids DB runtime merge issues but creates project files. This is likely less desirable than fixing DB catalog merge.

Recommendation:

1. MVP read-only report.
2. Fix DB runtime merge for chains/teams/workflows.
3. Later add `--scaffold` using internal catalog tools.
4. Later add `--start` after budget/approval gates are explicit.

---

## 8. Testing Patterns

Relevant tests:

- `src/__tests__/catalog-store.case.ts`
- `src/__tests__/catalog-tools.case.ts`
- `src/__tests__/invoke-agent.case.ts`
- `src/__tests__/loader.case.ts`
- `src/__tests__/catalog-resolver.case.ts`

Recommended new tests:

1. `src/__tests__/init-cli.case.ts`
   - Inject temp project root + temp library roots.
   - Call `runInit()` with a fake writable stream.
   - Assert output includes inventory and recommendation.
2. `src/__tests__/catalog-resolver.case.ts`
   - Extend to verify active DB `chain`, `team`, and `workflow` definitions are merged into runtime catalog.
3. Optional server/MCP test later if exposing init through MCP (not part of MVP).

---

## 9. Risks and Watchouts

| Risk | Detail | Mitigation |
|---|---|---|
| Catalog drift | `runInit()` could discover different definitions than runtime handlers | Extract shared `loadRuntimeCatalog()` helper |
| DB catalog mismatch | DB teams/chains/workflows currently not merged into runtime catalog | Fix resolver before scaffold/start |
| Over-smart recommendations | LLM-style suggestions may be nondeterministic | Rule-based MVP; optional `--smart` later |
| Side effects from init | Users asked for one-shot command, not wizard/scaffold by default | Default read-only; explicit flags for scaffold/start later |
| Host detection ambiguity | CLI may not know whether caller is OpenCode/Claude/Codex | Add explicit `--host`; default `opencode`; optionally inspect env later |

---

## 10. Research Conclusion

Implement in two increments:

### Increment 1 — Runtime catalog correctness

- Extend `resolveCatalog()` to merge active DB chains, teams, workflows.
- Add tests.

This makes internal catalog definitions actually runnable by `start_chain`, `build_team`, and `start_workflow`.

### Increment 2 — `init` one-shot report

- Add `src/cli/init.ts`.
- Add `init` subcommand to `src/index.ts`.
- Use shared runtime catalog loading.
- Print:
  - project root
  - library roots
  - host CLI context
  - inventory counts and names
  - root context files present (`AGENTS.md`, `CLAUDE.md`)
  - recommendations when `--task` is provided

Suggested flags:

```text
--task <text>       Task/request to recommend orchestration for
--host <host>       claude-code | codex | opencode | gemini | copilot (default: opencode)
--project <path>    Project root (default: cwd)
--json              Machine-readable JSON output
--verbose           Include definition source/path details
```

Defer:

- `--scaffold`
- `--start`
- `--smart` / LLM-powered recommendations
