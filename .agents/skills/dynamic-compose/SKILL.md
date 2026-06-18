---
name: dynamic-compose
description: Compose ephemeral chains, teams, and workflows at runtime from available agents when no static catalog entry matches the task.
argument-hint: "[task-description]"
trigger: /dynamic-compose
phase: meta
---

# Dynamic Compose Skill

Compose ad-hoc orchestration definitions (chains, teams, workflows) at runtime using only the agents, domains, and modes the user actually has. When no static catalog entry matches a task, this skill assembles the right shape on demand.

## When to Use

- No static chain/team/workflow in the catalog matches the task
- The task is multi-step and genuinely needs orchestration (not a single-agent job)
- The Orchestrator has already checked `list_catalog` and confirmed no match

## When NOT to Use

- A static catalog entry already covers the task → use it instead
- Single-agent task → dispatch directly, no orchestration needed
- Task is trivial (under 20 lines) → implement directly

## The Three Shapes

### Chain (default — sequential)

Use when: steps have dependencies, output of step N feeds step N+1.

```json
{
  "kind": "chain",
  "name": "<slug>",
  "description": "<what this chain does>",
  "version": "0.1.0",
  "entry": "<first-step-id>",
  "domain_skill_injection": "all_steps",
  "mode_skill_injection": "builder_steps_only",
  "steps": [
    {
      "id": "<step-id>",
      "agent": "<agent-name>",
      "skills": ["<skill-1>"],
      "description": "<what this step does>",
      "gate": "<optional: user_approval>",
      "domain": "<optional: domain-skill-name>",
      "mode": "<optional: mode-skill-name>",
      "transitions": {
        "success": "<next-step-id>",
        "failure": { "retry": 1, "then": "handoff" }
      }
    }
  ]
}
```

### Team (parallel — requires explicit user confirmation)

Use when: multiple independent perspectives on the same artifact, or non-overlapping files to modify in parallel.

```json
{
  "kind": "team",
  "name": "<slug>",
  "description": "<what this team does>",
  "version": "0.1.0",
  "budget_multiplier": <N>,
  "user_confirmation_required": true,
  "parallel": [
    {
      "role": "<role-name>",
      "agent": "<agent-name>",
      "skills": ["<skill-1>"],
      "domain": "<optional: domain-skill-name>",
      "mode": "<optional: mode-skill-name>",
      "focus": "<what this role investigates>"
    }
  ],
  "synthesize": {
    "agent": "orchestrator",
    "description": "<how results are merged>"
  }
}
```

### Workflow (multi-phase — chains + teams + gates)

Use when: the task needs conditional routing (e.g., assess severity → route to bugfix OR feature chain).

```json
{
  "kind": "workflow",
  "name": "<slug>",
  "description": "<what this workflow does>",
  "version": "0.1.0",
  "entry": "<first-phase-id>",
  "phases": [
    {
      "id": "<phase-id>",
      "kind": "chain|team|gate",
      "ref": "<catalog-ref>",
      "gate": "<optional: gate-type>",
      "prompt": "<optional: gate-prompt>",
      "on": {
        "success": "<next-phase-id>",
        "failure": "<fallback-phase-id>"
      }
    }
  ]
}
```

## Composition Procedure

### Step 1: Classify the Task

| Task Pattern | Typical Shape | Example |
|-------------|---------------|---------|
| Build a new feature | Chain (research → plan → implement → review) | "Add WebSocket support to notification service" |
| Investigate a problem | Chain (research → analyze → report) | "Why is the payment webhook failing intermittently?" |
| Multi-perspective assessment | Team (parallel reviews) | "Assess readiness for production launch" |
| Conditional routing | Workflow (assess → route) | "Handle this incident — severity determines path" |
| Refactor with phased plan | Chain (research → plan [gate] → implement → review) | "Migrate auth from session to JWT" |
| Parallel non-overlapping work | Team (parallel builders) | "Set up 3 independent microservices" |

### Step 2: Inventory Available Agents

Call `list_catalog({ kinds: ["domain", "mode"] })` to discover:
- **Agent types available** in the current environment (from the task subagent_types + skill list)
- **Domain skills** (backend, frontend, security, data, devops)
- **Mode skills** (autonomous, junior, senior)

Only compose from what exists. If an agent type is missing that the task needs:
- **STOP.** Do not invent agents.
- Report to the user: "This task needs a [role] agent, but none is available. Options: [a] proceed without it, [b] install/register it, [c] assign a different agent."
- Wait for user decision.

### Step 3: Select the Minimal Shape

1. **Default to chain** — the cheapest shape (1× token cost).
2. **Switch to team ONLY when**:
   - Multiple independent perspectives on one artifact (review lenses)
   - Multiple non-overlapping file groups to modify in parallel
   - User explicitly asks for parallel execution
   - **Always show cost multiplier** before confirming a team
3. **Switch to workflow ONLY when**:
   - The task has conditional routing (different paths based on assessment)
   - The task composes existing chains/teams into a larger flow

### Step 4: Assign Agents and Skills

Map each step/role to the best available agent:

| Step Purpose | Agent | Common Skills |
|-------------|-------|---------------|
| Research / investigate | scout | research |
| Plan / design | planner | plan |
| Implement (complex) | builder | implement, anti-speculation |
| Implement (single task) | implementor | implement |
| Implement (senior, high-risk) | implementor-senior | implement |
| Implement (junior, low-risk) | implementor-junior | implement |
| Review code | reviewer | review, extract-standards |
| Adversarial testing | red-team | review |
| Document | documenter | — |
| Security audit | security | review |
| SRE / reliability | sre | — |
| Architecture | architect | — |
| Ops / deployment | ops | — |

Apply domain skills to steps that need domain knowledge:
- Backend-heavy task → backend domain on research + plan + implement steps
- Security-sensitive task → security domain on review step
- Frontend task → frontend domain on implement step

Apply mode skills based on risk:
- High-risk (breaking changes, auth, data loss) → senior mode on implementor
- Low-risk (docs, minor fixes) → junior or autonomous mode
- Uncertain scope → senior mode (asks more questions)

### Step 5: Define Transitions

Standard transition patterns:

```
# Linear (most common)
success → next_step
failure → { retry: 1, then: "handoff" }

# Gate (user approval needed)
approved → next_step
rejected → prev_step

# Review cycle
pass → next_step
minor_issues → fix_step
blocking → fix_step
design_issues → plan_step

# Fix loop (max 2 cycles to prevent infinite loops)
fix_step → success → review_step
fix_step → failure → { retry: 1, then: "handoff" }
```

### Step 6: Budget Gate

Before executing, estimate and present:
1. **Token estimate**: steps × ~2-4K input + ~1-2K output per step
2. **Team cost multiplier**: 3× for 3-agent team, etc.
3. **Wall-clock estimate**: steps × ~30-60s per step (sequential), or parallel savings

Wait for user confirmation. Never skip.

### Step 7: Execute via MCP

Use the orchestrator MCP server:
1. `catalog_create_version` — register the composed definition in the runtime catalog
2. `catalog_set_active` — activate it
3. `start_chain` / `build_team` / `start_workflow` — begin execution
4. Follow the standard `orchestrate` skill loop (advance, observe, recover)

### Step 8: Post-Run Decision

After the composed definition completes:

| Outcome | Action |
|---------|--------|
| Success, first time | Ask user: "Save this pattern for future use?" |
| Success, same pattern used 3+ times | Flag for auto-promotion: "This pattern has succeeded 3 times. Recommend promoting to a static catalog entry." |
| Failure | Discard. Do not persist failed compositions. |
| Partial success | Keep ephemeral. Note what went wrong for future composition improvement. |

If the user says "save it" or auto-promotion triggers:
1. Write the JSON to `.ai/orchestration/<kind>s/<name>.json`
2. Add `source: "project"` metadata
3. The next `list_catalog` call will surface it alongside library entries

## Hard Rules

1. **Only use agents that exist.** Never invent or assume an agent is available.
2. **Only one agent per file at a time** in teams. Verify no file overlap before parallelizing.
3. **Budget gate is mandatory.** Show cost estimate before any execution.
4. **Teams require explicit user confirmation** — always show the budget multiplier.
5. **Max 2 fix loops.** If review → fix → review cycles 3 times, escalate to handoff.
6. **No auto-persist on failure.** Only successful compositions can be promoted.
7. **Ephemeral by default.** Composed definitions are runtime-only unless explicitly saved.
8. **Domain/mode layering is optional but recommended** for domain-specific tasks.

## Anti-Patterns

- Composing a chain for a task that a single agent can handle → use the agent directly
- Creating a 5-agent team for a 2-file change → overkill, use a chain
- Persisting every ad-hoc composition → catalog bloat; ephemeral by default
- Using "architect" agent for implementation steps → wrong agent for the job
- Skipping the budget gate because "it's just a small team" → always show the cost
- Composing a workflow when a simple chain suffices → workflow is for conditional routing

## Compose vs. Existing Catalog Priority

```
1. Check existing catalog → list_catalog
2. If match found → use static entry (tested, stable)
3. If no match → dynamic-compose (ephemeral)
4. After 3+ successful uses of same composition → promote to static
```

Static entries always take priority over dynamic compositions because they've been proven across multiple runs.

## Integration

- **Primary agent:** Orchestrator
- **Fallback from:** `orchestrate` skill (when "No catalog entry matches")
- **MCP tools:** `list_catalog`, `catalog_create_version`, `catalog_set_active`, `start_chain`, `build_team`, `start_workflow`
- **Output:** composed JSON definition (ephemeral or persisted to `.ai/orchestration/`)
- **See also:** `orchestrate` skill, `parallel-execution` skill