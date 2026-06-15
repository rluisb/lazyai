# Retired Orchestrator Research — Consolidated Findings

> Historical only. Spec 025 removed this runtime from the active LazyAI product surface. This research is preserved as decision context, not as current product or setup documentation.

---

## Table of Contents

- [1. Problem Statement](#1-problem-statement)
- [2. Architecture Decision: Three Layers](#2-architecture-decision-three-layers)
- [3. CLI Tool Capabilities (Programmatic)](#3-cli-tool-capabilities-programmatic)
- [4. Communication Protocols](#4-communication-protocols)
- [5. Seven Questions — Answered](#5-seven-questions--answered)
- [6. Composable Agent Model](#6-composable-agent-model)
- [7. Building Blocks: Chains, Teams, Workflows](#7-building-blocks-chains-teams-workflows)
- [8. Workflow Definitions (8)](#8-workflow-definitions-8)
- [9. Chain Definitions](#9-chain-definitions)
- [10. Team Definitions](#10-team-definitions)
- [11. Context Engineering Inheritance](#11-context-engineering-inheritance)
- [12. MCP Server Design](#12-mcp-server-design)
- [13. Token Budgets](#13-token-budgets)
- [14. Error Handling & Recovery](#14-error-handling--recovery)
- [15. CLI Commands](#15-cli-commands)
- [16. Product Fit](#16-product-fit)
- [17. Design Decisions Log](#17-design-decisions-log)
- [18. What Does NOT Change](#18-what-does-not-change)
- [19. Open Questions for Planning Phase](#19-open-questions-for-planning-phase)

---

## 1. Problem Statement

ai-setup today is a **scaffolding CLI** — it creates canonical `.ai/` directories, compiles tool-native configs, and gets out of the way. It runs once (or occasionally) at setup time. What it does NOT do is coordinate agents at runtime.

The orchestration layer answers a different question: **how do multiple agents work together on complex tasks?** Real software engineering involves research, planning, implementation, review, documentation — these are distinct roles that benefit from distinct agents with distinct instructions and tool restrictions.

### What Orchestration Solves

1. **Sequential workflows** — a feature needs research before planning, planning before implementation, implementation before review. Today this is manual (the human dispatches each agent).
2. **Parallel review** — code review benefits from multiple perspectives (correctness, security, quality) running simultaneously, then synthesizing findings. No CLI tool coordinates this natively across tools.
3. **Agent specialization** — a "builder" agent that also knows backend patterns and operates in senior mode is more effective than a generic builder. Composition of base agent + domain knowledge + behavioral mode is not supported today.
4. **State tracking** — when a chain of agents runs, someone needs to track what happened, what's next, what failed, and what the token budget looks like. Today that's entirely the human.
5. **Error recovery** — when step 3 of 5 fails, the human has to figure out what to do. The orchestrator provides structured recovery: retry, fix-resume, escalate to different agent, or handoff to new session.
6. **Cross-tool compatibility** — Claude Code, Codex, OpenCode, Gemini, Copilot all have different capabilities. The orchestration layer provides a universal interface (MCP) that works with any tool that supports the protocol.

### Why Now

MCP (Model Context Protocol) is adopted by Anthropic, OpenAI, Google, and Microsoft. It provides a standard way for CLI tools to call external servers. AgentSkills provides a standard for skill injection. Together they make tool-agnostic orchestration viable for the first time.

---

## 2. Architecture Decision: Three Layers

The architecture has three layers (plus the existing Layer 0):

```
Layer 3: Orchestrator Agent (.claude/agents/orchestrator.md)
         ↓ calls MCP tools
Layer 2: Orchestration MCP Server (@ai-setup/orchestrator)
         ↓ reads definitions from
Layer 1: Definition Files (.ai/orchestration/, library/)
         ↓ scaffolded by
Layer 0: ai-setup CLI (setup-time only)
```

### Layer 0: ai-setup CLI (what exists today)

- Scaffolds files, compiles tool directories
- Runs once at setup time, then gets out of the way
- NO fundamental changes needed to current implementation
- Adds one new scaffold function (`scaffoldOrchestration()`) and one new feature flag (`orchestration`)

### Layer 1: Definition Files (scaffolded by ai-setup)

- Chain configs, team configs, workflow configs (JSON)
- Domain skills and mode skills (Markdown)
- Agent base definitions (the 7 we already have)
- Constitution, fragments (already exist)
- Static files on disk — no runtime dependency

### Layer 2: Orchestration MCP Server (NEW runtime)

- Long-running process during coding sessions
- Reads definitions from Layer 1
- Exposes tools: `compose_agent`, `start_chain`, `build_team`, etc.
- Maintains state: chain progress, team tasks, token budgets
- Returns structured results to any CLI tool via MCP protocol
- **Pure logic — no LLM access needed** (see Section 12 for rationale)

### Layer 3: Orchestrator Agent (lives in each CLI tool)

- Loaded as `.claude/agents/orchestrator.md` or `.opencode/agents/orchestrator.md` or `.codex/agents/orchestrator.toml`
- System prompt tells the LLM how to use the MCP tools
- Knows how to interpret MCP results and display them
- Dispatches subagents for each chain/team step
- Watches subagent work, reads output, manages flow

### How It Actually Works (Step by Step)

User says: "Implement the auth feature"

1. User opens Claude Code (or OpenCode, or Codex)
2. CLI tool loads orchestrator agent definition
3. Orchestrator's system prompt includes available MCP tools, chain interpretation, confirmation rules, error recovery patterns
4. Orchestrator calls MCP: `start_chain("feature", {task: "auth feature"})`
5. MCP server returns: `{chainId: "abc", currentStep: {agent: "scout", skill: "research"}}`
6. Orchestrator dispatches Scout subagent with research skill
7. Scout returns findings
8. Orchestrator calls MCP: `advance_chain("abc", {result: findings})`
9. MCP server returns: `{currentStep: {agent: "planner", skill: "plan"}, gate: "user_approval"}`
10. Orchestrator shows plan to user, waits for approval
11. User approves
12. Orchestrator calls MCP: `advance_chain("abc", {approved: true})`
13. ...continues through chain until done

### What the Orchestrator Agent Sees

The orchestrator runs INSIDE the CLI tool session. It has access to:

- Subagent output (via Task tool in Claude Code, task tool in OpenCode)
- MCP tool results (structured chain/team state)
- The conversation history (what the user said, what agents returned)
- Token usage (via `/cost` or equivalent)

For Claude Code specifically:

```
Orchestrator (main session)
├── calls MCP: start_chain() → gets chain state
├── dispatches Scout via Task tool → gets research output
├── calls MCP: advance_chain() → gets next step
├── dispatches Planner via Task tool → gets plan
├── shows plan to user → waits
├── dispatches Builder via Task tool → gets code changes
│   └── Builder's stdout/stderr visible to orchestrator
├── dispatches Reviewer via Task tool → gets review
└── calls MCP: complete_chain() → final state
```

---

## 3. CLI Tool Capabilities (Programmatic)

### Capability Matrix

| Tool | Non-Interactive Mode | SDK | Structured Output | Subagents |
|------|---------------------|-----|-------------------|-----------|
| **Claude Code** | `claude -p 'prompt'` | Claude Agent SDK (Python/TS) | `--output-format json`, `--json-schema` | Native subagents + experimental teams |
| **OpenCode** | SDK: `@opencode-ai/sdk` | Full TypeScript SDK | Session/message streaming | Task tool with subagent types |
| **Codex** | `codex exec 'prompt'` | None (CLI only) | `--json` JSONL stream, `--output-schema` | Native subagents (`.codex/agents/`) |
| **Gemini** | Limited CLI | None | None | None |
| **Copilot** | VS Code extension only | None | None | None |

### Claude Code — Most Capable

**Subagents (stable):** Each gets own context, system prompt, tools, model. Report back to caller only.

**Agent Teams (experimental):** Full independent sessions with:
- Cross-agent messaging (mailbox)
- Shared task list with dependencies
- Self-claiming tasks
- Split tmux panes per teammate
- Plan approval gates
- Hooks: TeammateIdle, TaskCreated, TaskCompleted

**Claude Agent SDK:** Python/TS SDK gives same tools as Claude Code programmatically. Provides `query()` function for agentic loop, streaming or single-turn, custom system prompts, tool restrictions, MCP server connections, permission modes, file-based config support (CLAUDE.md, skills, etc.).

### Codex — Strong Second

**Subagents:** Native support with custom agent definitions in TOML format.
- Built-in agents: default, worker, explorer
- Custom agents: `.codex/agents/<name>.toml`
- CSV batch processing for parallel work
- Max threads (default 6), max depth (default 1)

**codex exec:** Non-interactive mode with JSON streaming, structured output schemas.

### OpenCode — SDK Available

**SDK:** Full TypeScript SDK (`@opencode-ai/sdk`) with:
- Session management
- Message streaming
- Programmatic control

### Gemini / Copilot — Limited

No programmatic orchestration capability. Gemini has basic CLI; Copilot is VS Code only. For these tools, the orchestrator runs as a skill (instructions only, no subagent dispatch). Sequential execution only — LLM handles each step directly. MCP tools still available for state management.

### Per-Tool Orchestrator Implementation

| Tool | Agent Location | Dispatch Mechanism | MCP Config |
|------|---------------|-------------------|------------|
| Claude Code | `.claude/agents/orchestrator.md` | Task tool for subagents | `.mcp.json` |
| OpenCode | `.opencode/agents/orchestrator.md` | Task tool with subagent_type | `.opencode/opencode.jsonc` |
| Codex | `.codex/agents/orchestrator.toml` | Native subagent system | MCP config |
| Gemini | `.gemini/skills/orchestrator/SKILL.md` | Instructions only (no dispatch) | MCP config |
| Copilot | Orchestrator as prompt (limited) | None | `.vscode/mcp.json` |

---

## 4. Communication Protocols

### Available Protocols

| Protocol | Used By | Direction | Best For |
|----------|---------|-----------|----------|
| **MCP tools** | All tools | Tool invocation | Custom tools callable by LLM |
| **RPC over stdin/stdout** | Pi (`--mode rpc`) | Bidirectional JSONL | Full agent control from external process |
| **Print mode + JSON** | Claude Code (`-p --output-format json`) | One-shot, structured output | Script orchestration, CI |
| **codex exec + JSON** | Codex (`codex exec --json`) | JSONL streaming | Automation, scripting |
| **TypeScript SDK** | OpenCode (`@opencode-ai/sdk`) | Full programmatic | Building on top of OpenCode |
| **Claude Agent SDK** | Claude Code (Python/TS) | Full programmatic | Building custom agents with Claude tools |
| **AgentSkills** | Claude, Codex, Pi, OpenCode | Skill loading | Passive instruction injection |

### What We Chose: MCP as the Universal Protocol

MCP was chosen as the primary communication protocol because:

1. **Universal adoption** — MCP is supported by Anthropic, OpenAI, Google, Microsoft
2. **Tool invocation model** — LLMs already know how to call tools; MCP tools appear as native tools
3. **Stateful servers** — MCP servers are long-running processes with in-memory state and session management (`Mcp-Session-Id`)
4. **Structured results** — tool results are displayed natively in every CLI tool's conversation
5. **Resource exposure** — MCP `resources` spec supports subscribable state (chain status, team tasks)
6. **No tool-specific code** — any future tool supporting MCP works automatically

### Pi's RPC Model — Noted for Future Reference

Pi exposes a complete RPC protocol over stdin/stdout (`pi --mode rpc`) with commands like `prompt`, `steer`, `follow_up`, `abort`, `get_state`, `get_messages`, `new_session`, and events like `agent_start`, `turn_start`, `tool_call`, `message_update`, `agent_end`. Pi's extension model allows registering tools, commands, intercepting events, sending messages programmatically, managing sessions, and controlling active tools at runtime. This is noted as the gold standard for external control but not our primary protocol (MCP has broader adoption).

### AgentSkills as Complementary Protocol

AgentSkills enables composability as a passive mechanism:
- Base agent definition (system prompt, tools, model)
- Skills loaded on activation (progressive disclosure)
- `allowed-tools` field for tool restriction per skill
- Skills can include scripts, references, assets
- Works alongside MCP — skills define WHAT the agent knows, MCP tools define WHAT the agent can do

---

## 5. Seven Questions — Answered

### Q1: How Do Chains Pass Context Between Steps?

Three proven patterns identified:

**Pattern A: File-Based Handoff** — Each step writes output to a file, next step reads it. Used by Codex (`codex exec -o output.md`), Claude Code (`claude -p > output.md`). Pros: simple, tool-agnostic, inspectable, resumable. Cons: no streaming, context may be too large.

**Pattern B: Session Continuity** — Same session persists, context stays in conversation window. Used by Claude Code (`-c` flag), Codex (`resume --last`), Pi (session persistence). Pros: context preserved naturally. Cons: can't mix tools between steps.

**Pattern C: MCP Resource Handoff** — MCP server acts as shared memory between steps. MCP `resources` spec natively supports this. Memory MCP server provides knowledge graph. Pros: tool-agnostic, structured, subscribable. Cons: requires running MCP server.

**Decision:** Mix of A + C. Files for simple chains, MCP resources for cross-tool chains. In practice, the orchestrator agent runs inside a single CLI tool session, so session continuity (B) is the primary mechanism for passing subagent results. MCP state (C) provides the structured chain metadata.

### Q2: How Do Teams Coordinate in Parallel?

**Native team coordination:** Claude Code Teams have shared task list + mailbox + dependency tracking + self-claiming. Codex subagents support parallel spawning (6 threads), results return to parent. Mastra uses Council pattern: parallel agents → synthesis step.

**For tools without native teams:** An MCP server manages a task list. LLM calls `get_next_task()`, works, calls `complete_task(result)`. Server handles assignment and dependencies.

**Decision:** Two tiers — native teams when available, MCP-based task list as universal fallback.

### Q3: How Do CLI Tools Display External Agent Output?

MCP tool results are the universal display mechanism. Every MCP-capable tool displays tool call results in its conversation.

| Tool | Display Mechanism |
|------|------------------|
| Claude Code | MCP tool result + stream-json |
| Codex | MCP tool result + exec --json |
| OpenCode | MCP tool result + SDK streaming |
| Pi | Extension TUI widgets + custom renderers |
| Gemini | MCP tool result (text only) |

**Decision:** MCP tool results for universal display. Pi extension for rich visualization (future).

### Q4: Expert Agent Discovery Mechanism

**Build-time:** Scan `.agents/skills/`, `.claude/agents/`, `.opencode/agents/` — each has structured frontmatter.

**Runtime:** Orchestration MCP server exposes `list_agents`, `list_skills`, `list_chains`, `list_teams` tools.

**Decision:** MCP `tools/list` + `resources/list` as the discovery mechanism. The orchestrator IS the expert (not a separate agent) — it already knows the catalog because it needs to for dispatching. Discovery is exposed as MCP tools for tools without native orchestrator support: `list_agents`, `list_skills`, `list_chains`, `list_teams`, `suggest_composition(task)`.

### Q5: Runtime vs Build-Time Agent Composition

**Build-time (static files):** Pre-generated by ai-setup. No runtime dependency. Problem: combinatorial explosion (every base × skill combo = separate file).

**Runtime (dynamic composition via MCP):** `compose_agent(base: 'engineer', skills: ['backend', 'typescript', 'senior'])` — LLM applies composed prompt to its behavior or spawns subagent.

**Hybrid (templates + runtime parameters):** Pre-define templates with slots, fill at runtime.

**Decision:** Hybrid — ai-setup generates templates (definition files on disk), MCP server composes at runtime. The `compose_agent` MCP tool assembles: base agent definition + domain skill + mode skill + chain step context → final agent prompt.

### Q6: Can MCP Tools Maintain State Across Calls?

**Yes.** MCP servers are long-running processes with in-memory state. MCP spec (2025-06-18) explicitly supports session management with `Mcp-Session-Id`.

State patterns:
- **In-memory:** Map of chain/team states
- **File-persisted:** Write to `.ai/orchestration/state/`
- **MCP resources:** Expose state as subscribable resources

**Decision:** In-memory state + file persistence + MCP resource exposure. In-memory for fast access during a session, file persistence for recovery after crashes/restarts, MCP resources for state visibility.

### Q7: Boundary Between ai-setup and Orchestration Layer

| ai-setup (scaffolding) | Orchestration Layer (runtime) |
|------------------------|------------------------------|
| One-time setup | Every coding session |
| Files on disk | Running MCP server |
| Agent definitions, skill files | Composed agents, running chains |
| `.ai/`, tool directories | In-memory state, task lists |
| User invokes via CLI | LLM invokes via MCP tools |

**Decision:** Separate MCP server package (`@ai-setup/orchestrator`) that reads definitions scaffolded by ai-setup. ai-setup stays a scaffolding tool; the orchestrator is a runtime tool. They share a repo but are independent packages.

---

## 6. Composable Agent Model

### Current Agents (7 Base Definitions)

| Agent | Model | Role | Type |
|-------|-------|------|------|
| Orchestrator | opus | Coordinate agents, track progress | Meta/coordination |
| Scout | sonnet | Map codebase, find patterns | Research |
| Planner | opus | Turn research into plans | Planning |
| Builder | sonnet | Follow plan, write code | Execution |
| Reviewer | opus | Find issues, report them | Quality |
| Red-Team | opus | Break code, find holes | Security/adversarial |
| Documenter | sonnet | Write docs | Documentation |

These define ROLES, not domains. They remain unchanged (except the orchestrator gets enhanced with MCP tool awareness).

### Domain Skills (5) — Knowledge Injection

Domain specialization changes what the agent KNOWS and DOES:

| Domain Skill | What It Adds |
|-------------|-------------|
| `backend` | API patterns, database queries, auth, caching, scaling, queue/worker, REST/GraphQL |
| `frontend` | Components, state management, CSS, accessibility, UX, SSR/CSR, browser APIs |
| `devops` | CI/CD, Docker/containers, IaC, monitoring, deployment, secrets management |
| `data` | SQL, ETL, pipelines, schemas, migrations, data modeling, analytics |
| `security` | OWASP, threat modeling, auth flows, secrets, encryption, audit logging |

**Why not mobile and fullstack?** Revised from the original 7 down to 5. Mobile and fullstack were removed as they are either too niche for the default set or combinations of existing domains. Users can create custom domain skills via `ai-setup create domain`.

### Mode Skills (3) — Behavioral Modifiers

| Mode | Behavioral Modifier |
|------|-------------------|
| `senior` | Autonomous decisions, proactive improvements, question requirements, suggest alternatives |
| `junior` | Ask before every decision, stick to plan exactly, flag everything, request review |
| `autonomous` | Minimal confirmation, full auto, only stop on errors or budget limits |

### Counterargument Addressed: Seniority as Agents vs Skills

Initial proposal was to have separate "senior backend engineer" and "junior backend engineer" base agents. This was rejected because:

**Seniority is an LLM prompt engineering pattern, not an agent architecture pattern.** A "junior" vs "senior" agent is just a different system prompt — the model doesn't become dumber, you just tell it to be more cautious or less autonomous. This is better handled as a **skill modifier** than a base agent.

Example of senior-mode skill content:
```
"You are experienced. Make autonomous decisions. Only escalate blockers."
"Suggest improvements proactively. Question requirements when they seem wrong."
```

Example of junior-mode skill content:
```
"Ask for clarification before every non-trivial decision."
"Do not deviate from the plan. Flag everything."
```

This is simpler and more composable than creating separate agent types per seniority level.

### Composition Formula

```
Final Agent Context =
  Root File (AGENTS.md)              ← project-level CE (RPI, CoT, quality gates)
  + Agent Definition (builder.md)     ← role-level instructions
  + Domain Skill (backend.md)         ← domain knowledge injection
  + Mode Skill (senior.md)            ← behavioral modifier
  + Chain Step Context                ← what prior steps produced (from MCP)
```

The MCP server's `compose_agent` tool assembles this:

```json
{
  "tool": "compose_agent",
  "input": {
    "base": "builder",
    "skills": ["backend", "senior"],
    "chain_context": {
      "prior_steps": ["research findings...", "approved plan..."],
      "current_step": "implement auth middleware"
    }
  },
  "output": {
    "agent_prompt": "[composed system prompt with all layers]",
    "tools": ["Read", "Edit", "Bash", "Grep"],
    "model": "sonnet"
  }
}
```

---

## 7. Building Blocks: Chains, Teams, Workflows

### Design Decision: Keep Separate (Not Unified)

A key design debate was whether to unify chains, teams, and workflows into a single "workflow" concept, or keep them as independent composable pieces.

**The "unify" argument (from Note #22):** Chains are just steps inside a workflow. Teams are a strategy choice within a workflow step. Having 3 concepts creates confusion.

**The "keep separate" decision (from Note #23):** After deliberation, the decision was to keep them independent:

1. **Users need ad-hoc chains without full workflow overhead** — "Research the auth module then plan a fix" doesn't need a full workflow
2. **Teams can be assembled on-the-fly for any purpose** — "Get multiple perspectives on this PR" is a standalone team, not a workflow
3. **Workflows are PRESETS that compose chains + teams, not replacements** — workflows use chains and teams internally
4. **More granular control for power users** — simpler building blocks = more flexibility

### Hierarchy

```
Layer 4: Workflows    — presets that compose chains + teams + gates + recovery
Layer 3: Chains       — reusable sequential execution patterns
Layer 3: Teams        — reusable parallel execution patterns
Layer 2: Agents       — composable units (base + domain + mode skills)
Layer 1: Skills       — injectable knowledge/behavior modifiers
Layer 0: Definitions  — files on disk (scaffolded by ai-setup)
```

### Independence

- A chain can run without a workflow (ad-hoc sequential execution)
- A team can run without a workflow (ad-hoc parallel execution)
- A workflow composes chains + teams + user gates + error recovery into a full preset
- Any of these can use composed agents (base + domain + mode)

### Usage Examples

```
# Ad-hoc chain (no workflow needed)
User: "Research the auth module then plan a fix"
Orchestrator → start_chain("research-plan", {task: "auth module fix"})

# Ad-hoc team (no workflow needed)
User: "Get multiple perspectives on this PR"
Orchestrator → build_team("review-council", {task: "PR #42"})

# Full workflow (composes chains + teams + gates)
User: "Implement the new payment feature"
Orchestrator → start_workflow("rpi", {task: "payment feature"})
  → internally uses research-plan chain
  → then implement-review chain
  → with user gates between phases
```

### Chains Are State Machines, Not Pipelines

A critical insight from the research: chains should be **state machines** (dynamic transitions), not **pipelines** (fixed steps). Real engineering workflows aren't linear — bugs get found during implementation, plans change mid-flight, reviews reveal design flaws that need re-planning.

Each chain step has transitions that define multiple possible next states:

```json
"transitions": {
  "success": "implement",
  "failure": { "retry": 1, "then": "escalate_to:planner" }
}
```

Or for review:

```json
"transitions": {
  "pass": "document",
  "minor_issues": "fix",
  "design_issues": "plan"
}
```

The orchestrator (the LLM) decides which transition to take based on the step's output. The MCP server manages the state machine mechanics.

---

## 8. Workflow Definitions (8)

Workflows are the top-level orchestration unit. Each workflow composes chains, teams, user gates, and recovery patterns into a preset for a common software engineering process. Every workflow maps to content that ALREADY EXISTS in the ai-setup library (fragments, skills, templates).

### Workflow → Library Content Mapping

| Workflow | Fragments Used | Skills Used | Templates Used |
|----------|---------------|-------------|----------------|
| rpi | rpi-workflow, reasoning, quality-gates | research, plan, implement, anti-speculation | plan-template, spec-template, task |
| code-review | reasoning | extract-standards | code-review-template |
| bug-investigation | bug-resolution, reasoning | research, iterate, tdd-loop, memory-write | bugfix-rca-template |
| system-design | decision-protocol, reasoning | research, plan | adr, spec-template |
| refactor | rpi-workflow, reasoning, quality-gates | extract-standards, implement, tdd-loop | plan-template, adr |
| onboard | context-discipline | research, extract-standards | checklist-template |
| tdd | quality-gates | tdd-loop, implement, iterate | task |
| incident-response | bug-resolution, reasoning | research, implement, tdd-loop, memory-write | postmortem-template |

### 1. `rpi` — Research, Plan, Implement (Core)

Source: `fragments/rpi-workflow.md` + `specs-agents/workflows.md`

```
Trigger: Any non-trivial task that needs structured approach

Steps:
  research → scout + research skill
  ⛔ USER GATE
  plan → planner + plan skill
  ⛔ USER GATE
  implement → builder + implement skill + anti-speculation
  review → reviewer
    IF blockers → builder + iterate (fix) → reviewer (re-review)
    IF design_issues → planner (revise plan)
  document → documenter (if behavior changed)

Inherits: RPI workflow, quality gates, reasoning protocol from root file
```

### 2. `code-review` — Code Review

Source: `templates/code-review-template.md` + `rules/review.md`

```
Trigger: PR review, code assessment, pre-merge check

Strategy: TEAM (council pattern — multiple perspectives)
  parallel:
    reviewer (correctness + logic)
    reviewer (quality + patterns) + extract-standards skill
    red-team (security + edge cases)
  synthesize: merge findings, deduplicate, rank by severity

Inherits: review rule, quality gates
```

### 3. `bug-investigation` — Bug/Issue Investigation

Source: `fragments/bug-resolution.xml` + `specs-agents/workflows.md` (bugfix type)

```
Trigger: Bug report, failing test, production issue

Steps:
  reproduce → scout + research (trace the bug)
  diagnose → scout (identify root cause)
  ⛔ USER GATE (confirm root cause before fixing)
  fix → builder + iterate + tdd-loop
  verify → reviewer (regression check)
  document → documenter + memory-write (capture lesson)

Inherits: bug-resolution protocol
```

**Note on Planner in bugfix:** For simple bugs, the Planner step is intentionally skipped — going directly from diagnosis to fix is faster. However, for complex bugs (race conditions, data corruption), the orchestrator should conditionally invoke the Planner for root-cause analysis based on Scout's findings:

```
Scout(reproduce) → IF complexity > threshold THEN Planner(root-cause-analysis) → Builder(fix) → Reviewer(verify)
```

### 4. `system-design` — System Design / Architecture

Source: `fragments/decision-protocol.md` + `templates/adr.md` + `templates/spec-template.md`

```
Trigger: Architecture decision, new system component, major refactor

Steps:
  research → scout + research (map current architecture)
  assess → scout (analyze constraints, trade-offs)
  ⛔ USER GATE
  design → planner + plan (propose architecture with options)
    Uses: decision protocol (ToT — evaluate multiple approaches)
    Output: ADR + spec
  ⛔ USER GATE (approve architecture)
  implement → builder + implement (if proceeding)
  review → reviewer + red-team (architecture review)

Inherits: decision protocol, ADR enforcement
```

### 5. `refactor` — Refactoring

Source: `specs-agents/workflows.md` (refactor type)

```
Trigger: Tech debt, code smell, pattern migration

Steps:
  research → scout + extract-standards (map current vs desired state)
  plan → planner + plan (phased refactor plan — NO big bang)
  ⛔ USER GATE (approve phases)
  FOR EACH phase:
    implement → builder + implement + tdd-loop
    verify → reviewer (no regressions)
  document → documenter (update affected docs, write ADR)

Rule: ADR is MANDATORY for refactors
```

### 6. `onboard` — Codebase Onboarding

Source: `templates/checklist-template.md` + KNOWLEDGE_MAP concept

```
Trigger: New team member, new codebase, understanding request

Steps:
  architecture → scout + research (high-level structure)
  conventions → scout + extract-standards (patterns and rules)
  testing → scout + research (test patterns and coverage)
  documentation → documenter (generate onboarding guide)

Output: Comprehensive onboarding document with actionable sections
```

**Note:** This uses the SAME agent (Scout) three times with different prompts. Chains don't require different agents per step.

### 7. `tdd` — Test-Driven Development

Source: `skills/tdd-loop.md` + `skills/iterate.md`

```
Trigger: User asks for TDD approach, or feature requires high test confidence

Steps:
  plan-tests → planner + plan (define test cases from requirements)
  red → builder + tdd-loop (write failing tests)
  green → builder + implement (write minimum code to pass)
  review → reviewer (verify coverage and correctness)
  refactor → builder + iterate (clean up, extract patterns)
  verify → reviewer (final check)

Cycles: red → green → refactor can repeat
```

### 8. `incident-response` — Production Incident

Source: `templates/postmortem-template.md` + `fragments/bug-resolution.xml`

```
Trigger: Production issue, P0/P1 incident

Steps:
  triage → scout + research (identify scope and impact)
  ⛔ USER GATE (confirm severity)
  mitigate → builder + implement (immediate fix/rollback)
  verify → reviewer (confirm mitigation works)
  root-cause → scout + research (deep investigation)
  permanent-fix → builder + implement + tdd-loop
  postmortem → documenter + memory-write (using postmortem template)

Priority: Speed over perfection in mitigation step
```

---

## 9. Chain Definitions

Chains are standalone reusable sequential patterns. They can be used independently (ad-hoc) or composed inside workflows.

### Chain: `feature`

```
Scout(research) → Planner(plan) → [USER GATE] → Builder(implement) → Reviewer(review) → Builder(fix) → Documenter(docs)
```

Full JSON definition:

```json
{
  "name": "feature",
  "description": "New feature development: research → plan → implement → review → document",
  "steps": [
    {
      "id": "research",
      "agent": "scout",
      "skills": ["research"],
      "description": "Map affected code, patterns, and dependencies",
      "output": "Research findings document",
      "transitions": {
        "success": "plan",
        "failure": { "retry": 1, "then": "escalate" }
      }
    },
    {
      "id": "plan",
      "agent": "planner",
      "skills": ["plan"],
      "description": "Create phased implementation plan from research",
      "output": "Implementation plan with tasks",
      "gate": "user_approval",
      "transitions": {
        "approved": "implement",
        "rejected": "research"
      }
    },
    {
      "id": "implement",
      "agent": "builder",
      "skills": ["implement", "anti-speculation"],
      "description": "Execute plan tasks one at a time",
      "output": "Code changes with passing tests",
      "transitions": {
        "success": "review",
        "failure": { "retry": 2, "then": "escalate_to:planner" }
      }
    },
    {
      "id": "review",
      "agent": "reviewer",
      "skills": ["extract-standards"],
      "description": "Review implementation for correctness and quality",
      "output": "Review findings",
      "transitions": {
        "pass": "document",
        "minor_issues": "fix",
        "design_issues": "plan"
      }
    },
    {
      "id": "fix",
      "agent": "builder",
      "skills": ["iterate"],
      "description": "Address review findings",
      "transitions": {
        "success": "review",
        "failure": { "retry": 1, "then": "escalate" }
      }
    },
    {
      "id": "document",
      "agent": "documenter",
      "skills": [],
      "description": "Update documentation for changed behavior",
      "transitions": {
        "success": "done"
      }
    }
  ],
  "domain_skill_injection": "all_steps",
  "mode_skill_injection": "builder_steps_only"
}
```

### Chain: `bugfix`

```
Scout(reproduce+trace) → Builder(fix) → Reviewer(verify) → Builder(test)
```

Skills: `iterate`, `tdd-loop`. Note: Planner is intentionally SKIPPED for bugfixes (conditional inclusion for complex bugs).

### Chain: `review`

```
Scout(map changes) → Reviewer(review) → Red-Team(adversarial review) → [SYNTHESIZE]
```

This is a council pattern — two reviewers with different angles. The orchestrator synthesizes findings from both.

### Chain: `refactor`

```
Scout(map current state) → Planner(refactor plan) → [USER GATE] → Builder(refactor) → Reviewer(verify no regressions) → Documenter(update docs)
```

Skills: `extract-standards`, `tdd-loop`

### Chain: `onboard`

```
Scout(architecture) → Scout(conventions) → Scout(test patterns) → Documenter(generate guide)
```

Same agent (Scout) three times with different prompts. Chains don't require different agents per step.

### Chain: `tdd`

```
Planner(define test cases) → Builder(write failing tests) → Builder(implement to pass) → Reviewer(verify coverage) → Builder(refactor)
```

Skills: `tdd-loop`, `iterate`

### Standalone Reusable Chains (from Design Decision #23)

In addition to the chains above (which map 1:1 with workflows), there are smaller reusable chains designed for ad-hoc composition:

| Chain | Steps | Use Case |
|-------|-------|----------|
| `research-plan` | scout → planner | Ad-hoc "research then plan" |
| `implement-review` | builder → reviewer | Ad-hoc "implement then review" |
| `tdd-cycle` | builder:red → builder:green → builder:refactor | Single TDD iteration |
| `investigate-fix` | scout → builder → reviewer | Quick bug investigation |

---

## 10. Team Definitions

Teams are standalone reusable parallel execution patterns. They can be used independently (ad-hoc) or composed inside workflows.

### Team: `review-council`

```json
{
  "name": "review-team",
  "description": "Parallel code review from multiple angles, then synthesize",
  "parallel": [
    {
      "role": "correctness-reviewer",
      "agent": "reviewer",
      "skills": [],
      "focus": "Logic errors, missing edge cases, incorrect behavior"
    },
    {
      "role": "security-reviewer",
      "agent": "red-team",
      "skills": [],
      "focus": "Security vulnerabilities, injection, auth bypass, data leaks"
    },
    {
      "role": "quality-reviewer",
      "agent": "reviewer",
      "skills": ["extract-standards"],
      "focus": "Code quality, patterns, maintainability, test coverage"
    }
  ],
  "synthesize": {
    "agent": "orchestrator",
    "description": "Merge findings, deduplicate, prioritize by severity"
  },
  "budget_multiplier": 3,
  "user_confirmation_required": true
}
```

### Team: `research-sweep`

3 parallel scouts investigating different aspects of a codebase or problem.

### Team: `assessment`

4 parallel scouts assessing architecture, code quality, test coverage, and security. Planner synthesizes into a prioritized report.

### Counterargument: Teams Are Expensive

Claude Code's own docs note: "Agent teams use significantly more tokens than a single session. Each teammate has its own context window."

For most tasks, a chain (sequential) is better than a team (parallel). Teams should be reserved for:
- Large reviews where parallel reviewers save wall-clock time
- Multi-area research where different aspects are truly independent
- High-stakes decisions where you want multiple perspectives

**Rule: Don't default to teams. Default to chains. Escalate to teams when the user or orchestrator decides parallel work adds value.**

**Always ask user before spawning a team:**
> "This task could benefit from parallel review (3 agents). Estimated cost: ~3× a single review. Proceed? [y/N]"

---

## 11. Context Engineering Inheritance

### What Already Exists (Compiled by ai-setup into Root File)

ai-setup already compiles these fragments into root files (AGENTS.md, CLAUDE.md):

| Fragment | What It Does |
|----------|-------------|
| `rpi-workflow.md` | Research → Plan → Implement flow |
| `reasoning-protocol.md` | CoT: Affected → Plan → Risks → Verdict |
| `decision-protocol.md` | ToT: Options → Pros/Cons → Decision |
| `quality-gates.xml` | Lint, typecheck, test, build gates |
| `context-discipline.md` | Token budgeting, context management |
| `agent-harness.md` | Multi-agent coordination rules |
| `bug-resolution.xml` | Reproduce → diagnose → fix → verify |
| `git-conventions.xml` | Branch/commit patterns |
| `system-context.xml` | Project name, language, framework |

### The Key Insight: Agents Inherit from the Root File

When the CLI tool starts a session:
1. It loads the root file (AGENTS.md / CLAUDE.md)
2. This root file already contains RPI, CoT, ToT, quality gates
3. Every agent (orchestrator, scout, builder, etc.) runs IN THIS SESSION
4. Therefore every agent already has all context engineering baked in

**The orchestrator doesn't need its own copy of RPI workflow — it's already in the root file.**

### But Agents Need ROLE-SPECIFIC Techniques

The root file provides PROJECT-LEVEL context engineering. But each agent also needs ROLE-SPECIFIC instructions:

| Agent | What It Needs Beyond Root File |
|-------|-------------------------------|
| Scout | Research-specific CoT: "Map → Trace → Document → Summarize" |
| Planner | Planning-specific ToT: "List options → Evaluate → Phase → Sequence" |
| Builder | Implementation-specific: "Test first → Implement → Run gates → Commit" |
| Reviewer | Review-specific: "Read → Annotate → Classify → Report" |
| Red-Team | Adversarial-specific: "Assume hostile → Find holes → Prove exploitable" |

These role-specific techniques live in the **agent definition files** (`library/agents/*.md`) and **skill files** (`library/skills/*.md`). They COMPLEMENT the root file, they don't REPLACE it.

### Full Composition Stack

```
Final Agent Context =
  Root File (AGENTS.md)              ← project-level CE (RPI, CoT, quality gates)
  + Agent Definition (builder.md)     ← role-level instructions
  + Domain Skill (backend.md)         ← domain knowledge injection
  + Mode Skill (senior.md)            ← behavioral modifier
  + Chain Step Context                ← what prior steps produced (from MCP)
```

Each layer adds specificity without duplicating what's above it. The root file gives the project-level foundation. The agent definition gives the role. The domain skill gives knowledge. The mode skill gives behavioral style. The chain context gives situational awareness.

---

## 12. MCP Server Design

### Core Principle: No LLM Access

The MCP server is **pure TypeScript logic** — no LLM API calls:

- Reads definition files (JSON, Markdown)
- Manages chain/team state (state machine)
- Composes agent prompts (string concatenation from templates + skills)
- Tracks token budgets (arithmetic)
- Returns structured data

**The LLM reasoning happens in the CLI tool, not in the MCP server.** The orchestrator agent IN the CLI tool is already an LLM — it can do all the "thinking" about composition, evaluation, and recovery. The MCP server just provides the state management and definition catalog.

This keeps the architecture clean:
- MCP server = deterministic logic (state machine, file I/O, composition rules)
- CLI tool's LLM = reasoning (when to use which chain, how to recover from errors)
- Agent definitions = instructions (what each role does)

### Why Not Give the MCP Server LLM Access?

The alternative was considered: MCP server calls an LLM to analyze task complexity, dynamically compose prompts, evaluate step output quality, and generate recovery suggestions.

Rejected because:
- Expensive (doubles API calls)
- Slower (adds latency to every tool call)
- Dependency on API key in the MCP server
- Unpredictable behavior
- The orchestrator agent already IS an LLM — no need for two LLMs

### MCP Tools (10 + Support Tools)

| Tool | Layer | What It Does |
|------|-------|-------------|
| `compose_agent` | 2 | Combine base agent + domain skill + mode skill → prompt |
| `start_chain` | 3 | Begin a sequential chain |
| `advance_chain` | 3 | Move chain to next step |
| `build_team` | 3 | Spawn parallel agents for a task |
| `assign_task` | 3 | Assign task to team member |
| `complete_task` | 3 | Mark team task complete with result |
| `start_workflow` | 4 | Begin a full workflow (internally uses chains/teams) |
| `advance_workflow` | 4 | Advance workflow (handles transitions, gates, recovery) |
| `get_status` | all | Status for any running chain/team/workflow |
| `list_catalog` | all | Browse agents, skills, chains, teams, workflows, domains, modes |

Support tools:

| Tool | Purpose |
|------|---------|
| `get_budget` / `update_budget` | Check/update token budget |
| `retry_step` | Retry a failed step |
| `escalate_step` | Route to different agent |
| `handoff` | Full context dump for new session |

### State Management

```
orchestrator/src/state/
├── chain-machine.ts    ← Chain state machine (dynamic transitions)
├── team-state.ts       ← Team task list + assignments
├── budget-tracker.ts   ← Token budget tracking
├── error-journal.ts    ← Error history + lessons
└── persistence.ts      ← Save/load state to .ai/orchestration/state/
```

- **In-memory** for fast access during a session
- **File-persisted** to `.ai/orchestration/state/` for crash recovery and session resumption
- **MCP resources** for state visibility (subscribable by the client)

### Package Structure

```
orchestrator/                           @ai-setup/orchestrator
├── package.json
├── tsconfig.json
├── tsup.config.ts
│
└── src/
    ├── index.ts                        MCP server entry point
    ├── server.ts                       MCP server setup (tools, resources)
    │
    ├── tools/                          MCP tool implementations
    │   ├── compose-agent.ts            compose_agent(base, skills[])
    │   ├── start-chain.ts              start_chain(name, context)
    │   ├── advance-chain.ts            advance_chain(chainId, result)
    │   ├── build-team.ts               build_team(name, task)
    │   ├── assign-task.ts              assign_task(teamId, taskId, agent)
    │   ├── complete-task.ts            complete_task(teamId, taskId, result)
    │   ├── list-catalog.ts             list_agents, list_skills, list_chains, list_teams
    │   ├── get-status.ts               get_chain_status, get_team_status
    │   ├── budget.ts                   get_budget, update_budget
    │   └── recovery.ts                 retry_step, escalate_step, handoff
    │
    ├── state/                          State management
    │   ├── chain-machine.ts            Chain state machine (dynamic transitions)
    │   ├── team-state.ts               Team task list + assignments
    │   ├── budget-tracker.ts           Token budget tracking
    │   ├── error-journal.ts            Error history + lessons
    │   └── persistence.ts              Save/load state to .ai/orchestration/state/
    │
    ├── composer/                        Agent composition logic
    │   ├── compose.ts                   Merge base + domain + mode → prompt
    │   ├── loader.ts                    Read definition files from disk
    │   └── validator.ts                 Validate composed agent
    │
    └── __tests__/                       Tests
        ├── chain-machine.test.ts
        ├── compose.test.ts
        ├── team-state.test.ts
        └── server.test.ts
```

---

## 13. Token Budgets

### The Real Cost Problem

| Operation | Estimated Cost (Claude Sonnet) |
|-----------|-------------------------------|
| Single chain (5 steps) | ~$3 per run |
| Team (3 parallel agents) | ~$9 per run |
| Complex feature (chain + review team) | ~$15 per run |

### Counterargument: Can We Even Estimate Costs?

Token usage is unpredictable. A "simple" implementation might take 500 tokens or 50,000 tokens. We CAN'T reliably estimate costs before execution.

### Decision: Budget-Based, Not Estimate-Based

Set BUDGETS, not estimates:

```
ai-setup orchestrate --budget $5.00 --chain feature
→ "Budget: $5.00. If any step exceeds budget, I'll pause and ask."
```

The MCP server tracks actual spend per step and pauses when approaching the limit.

### What Users See

Before execution:
```
🔗 Chain: feature (5 steps)
   Budget: $5.00
   Continue? [y/N]
```

During execution:
```
Step 3/5: Builder (implement)
   Spent so far: $1.45 / budget: $5.00
   ████████░░░░░░░░ 29%
```

### Implementation

The `budget-tracker.ts` state module:
- Accepts a total budget at chain/team/workflow start
- Tracks actual token usage per step (reported by the orchestrator agent)
- Calculates percentage consumed
- Pauses execution when budget threshold is reached (configurable, e.g., 80%)
- Exposes `get_budget` and `update_budget` MCP tools

---

## 14. Error Handling & Recovery

### 4 Recovery Patterns

| Pattern | When | What Happens |
|---------|------|-------------|
| **Retry** | Transient failure | Same step runs again (max 2 retries) |
| **Fix & Resume** | User fixes issue | Chain resumes from failed step |
| **Escalate** | Wrong approach | Routes to different agent (e.g., Builder fails → Planner re-plans) |
| **Handoff** | Fundamental issue | Full context dump, new session |

**Escalation is the most powerful pattern** — it's what distinguishes a state machine from a pipeline. The orchestrator can route failures to a different agent that might have a different perspective.

### Error Report Format

When a chain step fails, the orchestrator shows:

```
❌ Chain 'feature' failed at Step 3/5

  Chain:     feature
  Step:      3 — Builder (implement)
  Agent:     Builder + backend + senior-mode
  Task:      Implement auth middleware
  Skill:     implement, anti-speculation

  Error:     Test suite failed (3 failures in auth.test.ts)
  Cause:     Builder created middleware but didn't handle expired tokens

  Context saved to: .ai/orchestration/chain-abc123/

  Recommended actions:
    1. Review the error: cat .ai/orchestration/chain-abc123/step-3-error.md
    2. Fix the issue and resume: orchestrate resume chain-abc123
    3. Restart from this step: orchestrate retry chain-abc123 --from-step 3
    4. Get help: orchestrate explain-error chain-abc123
```

After any failure:
1. Show: chain, step, agent, skills, exact error
2. Show: what succeeded so far
3. Show: recommended action
4. Write lesson to error journal via MCP

### Handoff Context Dump

When a handoff is needed, the orchestrator creates a handoff document:
- What was accomplished (steps 1-2 results)
- What failed (step 3 details)
- What remains (steps 4-5)
- Full error context
- Suggested approach for retry

### Error Journal & Lessons

The MCP server writes a **lesson** after each failure:

```json
{
  "chain": "feature",
  "step": 3,
  "agent": "builder",
  "error": "Test failures in auth middleware",
  "root_cause": "Missing expired token handling",
  "resolution": "Added token expiry check before validation",
  "lesson": "Auth middleware must handle: valid, expired, malformed, and missing tokens",
  "timestamp": "2026-04-08T15:30:00Z"
}
```

This feeds into the `memory-write` skill and gets incorporated into future chain runs. The `error-journal.ts` state module maintains this history and can surface relevant past lessons when similar errors occur.

---

## 15. CLI Commands

### What Already Exists

```bash
ai-setup create workflow <name>    # Creates workflow .md file (WorkflowGenerator)
ai-setup list                      # Lists agents, skills, etc.
ai-setup info <name>               # Shows details
```

ai-setup already has:
1. `ai-setup create workflow` command — generates workflow definition files
2. `WorkflowGenerator` class — generates markdown workflow definitions with agent/skill refs
3. `specs-agents/workflows.md` — defines RPI flow for Feature, Bugfix, Refactor, Tech Debt
4. `fragments/agent-harness.md` — defines agent coordination, handoff protocol
5. `fragments/rpi-workflow.md` — the core RPI workflow

### What to Add

**Decision: Extend existing `create` / `list` / `info` commands** (not a separate subcommand). Add new categories to the existing `ArtifactType`:

```typescript
// src/types.ts — currently:
type ArtifactType = 'agent' | 'skill' | 'command' | 'prompt' | 'template' | 'workflow'

// Becomes:
type ArtifactType = 'agent' | 'skill' | 'command' | 'prompt' | 'template' | 'workflow' | 'domain' | 'mode'
```

Then:

```bash
# Workflows
ai-setup create workflow <name>       # Already exists — enhance to support JSON format
ai-setup list workflows               # List available workflows
ai-setup info <workflow-name>         # Show workflow details (steps, agents, transitions)

# Domain skills
ai-setup create domain <name>         # Create custom domain skill
ai-setup list domains                 # List available domains
ai-setup info <domain-name>           # Show domain details

# Mode skills
ai-setup create mode <name>           # Create custom mode skill
ai-setup list modes                   # List available modes

# General
ai-setup list chains                  # List available chains
ai-setup list teams                   # List available teams
```

**Plus a thin `ai-setup orchestration` alias** for discoverability:

```bash
# These are equivalent:
ai-setup create domain backend-go
ai-setup orchestration create domain backend-go

ai-setup list workflows
ai-setup orchestration list workflows

# Plus orchestration-specific commands:
ai-setup orchestration enable    # Enable orchestrator MCP server
ai-setup orchestration disable   # Disable orchestrator MCP server
ai-setup orchestration status    # Show orchestration health
```

The user doesn't HAVE to learn a new command. But the namespace exists for discoverability.

### Existing Commands That Change

| File | Change | Size |
|------|--------|------|
| `src/types.ts` | Add orchestration-related types | ~10 lines |
| `src/presets.ts` | Add 'orchestration' to feature flags | ~5 lines |
| `src/store/schema.ts` | Add orchestration config to Zod schema | ~15 lines |
| `src/commands/init.ts` | Add `--enable-orchestration` flag | ~10 lines |
| `src/commands/list.ts` | Add chains/teams/domains/modes/workflows categories | ~30 lines |
| `src/commands/info.ts` | Show chain/team/domain/mode/workflow details | ~40 lines |
| `src/commands/completions.ts` | Add new categories | ~10 lines |
| `src/wizard/index.ts` | Call scaffoldOrchestration() | ~10 lines |
| `src/wizard/phase1-context.ts` | Add orchestration question | ~15 lines |

### MCP Catalog Entry

```json
{
  "orchestrator": {
    "description": "Agent orchestration — composable agents, chains, teams",
    "command": "npx",
    "args": ["-y", "@ai-setup/orchestrator"],
    "tools": ["compose_agent", "start_chain", "build_team", "list_agents"],
    "enabled": false,
    "requiresInstall": false
  }
}
```

User opts in during init:
```
? Enable orchestration (agent teams, chains, expert guidance)? [y/N]
```

Or via CLI:
```bash
ai-setup init --enable-servers orchestrator
```

---

## 16. Product Fit

### The Question

Does the orchestration layer belong INSIDE ai-setup as an optional feature, or does it need to be a separate product?

### Analysis

| Aspect | ai-setup (scaffolding) | Orchestrator (runtime) | Conflict? |
|--------|----------------------|----------------------|-----------|
| When it runs | Setup time | Session time | Different lifecycle |
| What it produces | Files on disk | Dynamic responses | Different outputs |
| Dependencies | Node.js, filesystem | Node.js, MCP SDK, possibly LLM API | Orchestrator has heavier deps |
| Package size | ~2MB (templates + CLI) | Unknown (MCP SDK + state management) | Could bloat ai-setup |
| User expectation | "Set up my AI tools" | "Coordinate my AI agents" | Different mental model |
| MCP catalog | Already has 11 servers | Orchestrator IS an MCP server | Perfect fit as catalog entry |

### Decision: Same Repo, Separate npm Package, Opt-In via MCP Catalog

The orchestrator is **another MCP server in ai-setup's catalog** — like `memory`, `filesystem`, `ripgrep` are MCP servers that ai-setup configures, the orchestrator is another one.

1. **ai-setup stays a scaffolding tool** — no runtime behavior added to the CLI
2. **Orchestrator is a separate npm package** (`@ai-setup/orchestrator`) — published independently, versioned independently
3. **ai-setup scaffolds the config** — writes MCP config entries, creates chain/team/workflow definition files
4. **User opts in** — not everyone wants agent orchestration (it costs tokens)
5. **Any CLI tool uses it** — Claude Code, Codex, OpenCode all connect via MCP
6. **Same repo** — the orchestrator package lives alongside ai-setup as a subdirectory

### Repository Structure (NOT Monorepo — Yet)

```
ai-setup/
├── src/                           ← existing CLI
├── library/                       ← existing library
├── orchestrator/                  ← NEW: separate package.json
│   ├── src/
│   ├── package.json               ← @ai-setup/orchestrator
│   └── tsconfig.json
├── package.json                   ← existing CLI package.json
└── README.md
```

**Why not monorepo NOW?** Moving to a monorepo (turborepo/nx) is a significant refactor — it changes build system, CI/CD, import paths, publishing, and developer workflow. The single-repo-with-subdirectory approach works for now. Move to monorepo later IF more packages are added, shared code grows, or CI/CD needs workspace features.

### What ai-setup Scaffolds When Orchestration Is Enabled

When user runs `ai-setup init --enable-servers orchestrator`:

```
<project>/
├── .ai/
│   ├── constitution/                       (already created)
│   ├── mcp.json                            (orchestrator entry added)
│   │
│   └── orchestration/                      NEW
│       ├── chains/
│       │   ├── feature.json
│       │   ├── bugfix.json
│       │   ├── review.json
│       │   ├── refactor.json
│       │   ├── onboard.json
│       │   └── tdd.json
│       ├── teams/
│       │   ├── review-team.json
│       │   ├── feature-team.json
│       │   └── assessment-team.json
│       ├── workflows/
│       │   ├── rpi.json
│       │   ├── code-review.json
│       │   ├── bug-investigation.json
│       │   ├── system-design.json
│       │   ├── refactor.json
│       │   ├── onboard.json
│       │   ├── tdd.json
│       │   └── incident-response.json
│       ├── skills/
│       │   ├── domains/
│       │   │   ├── backend.md
│       │   │   ├── frontend.md
│       │   │   ├── devops.md
│       │   │   ├── data.md
│       │   │   └── security.md
│       │   └── modes/
│       │       ├── senior.md
│       │       ├── junior.md
│       │       └── autonomous.md
│       └── state/                          (empty, created at runtime by MCP server)
```

Per-tool output also generated:

```
├── .opencode/agents/orchestrator.md        Enhanced with MCP tool instructions
├── .claude/agents/orchestrator.md          Enhanced with MCP tool instructions
├── .agents/skills/orchestrator/SKILL.md    Orchestrator as skill for Codex
├── .opencode/opencode.jsonc              orchestrator MCP server added
├── .mcp.json                               orchestrator MCP server added
└── .vscode/mcp.json                        orchestrator MCP server added
```

---

## 17. Design Decisions Log

All decisions made during research, with rationale:

### D1: MCP as Communication Protocol (not RPC, not SDK)

**Decision:** Use MCP as the primary protocol for orchestration communication.
**Rationale:** Universal adoption (Anthropic, OpenAI, Google, Microsoft), tool-agnostic, stateful servers, structured results displayed natively.
**Alternatives considered:** Pi RPC (too tool-specific), Claude Agent SDK (Python/TS only), direct CLI invocation (not stateful).

### D2: No LLM in MCP Server

**Decision:** The MCP server is pure logic — no LLM API calls.
**Rationale:** The orchestrator agent in the CLI tool is already an LLM. Adding another LLM layer is expensive, slow, and unpredictable. The MCP server handles deterministic work: state machines, file I/O, prompt composition.
**Alternatives considered:** Smart MCP server with LLM for task analysis, dynamic composition, output evaluation.

### D3: Seniority as Mode Skills (not Base Agents)

**Decision:** Senior/junior/autonomous are behavioral modifier skills, not separate agent types.
**Rationale:** Seniority is a prompt engineering pattern, not an architecture pattern. The model doesn't become dumber — you just change instructions. Skills are more composable than separate agents.
**Alternatives considered:** Separate senior/junior agent definitions, seniority as agent property.

### D4: Domain Skills Reduced to 5

**Decision:** Ship 5 domain skills: backend, frontend, devops, data, security.
**Rationale:** Mobile and fullstack removed — too niche for defaults or combinations of existing domains. Users can create custom domains via `ai-setup create domain`.
**Alternatives considered:** 7 domains (including mobile, fullstack).

### D5: Chains, Teams, Workflows as Separate Concepts

**Decision:** Keep chains, teams, and workflows as independent composable pieces.
**Rationale:** Users need ad-hoc chains without workflow overhead. Teams can be assembled on-the-fly. Workflows are presets that compose chains + teams. Simpler building blocks = more flexibility.
**Alternatives considered:** Unifying everything under "workflow" with strategy parameter.

### D6: State Machines (not Pipelines)

**Decision:** Chains are state machines with dynamic transitions, not fixed pipelines.
**Rationale:** Real engineering workflows aren't linear. Reviews can reveal design flaws that need re-planning. The orchestrator decides which transition to take based on step output.
**Alternatives considered:** Fixed step sequences with optional skip.

### D7: Expert as Orchestrator Skill (not Separate Agent)

**Decision:** The orchestrator IS the expert — no separate expert agent.
**Rationale:** The orchestrator already knows the catalog (needs it for dispatching). Adding a "guidance" skill means no extra agent to maintain, natural flow (ask → recommend → approve → execute), single point of contact.
**Alternatives considered:** Expert as separate agent, expert as MCP tool only.

### D8: Budget-Based (not Estimate-Based) Token Management

**Decision:** Set budgets, track actual spend, pause at limits — don't try to estimate costs upfront.
**Rationale:** Token usage is unpredictable. A "simple" implementation might take 500 or 50,000 tokens. Budget limits are reliable; estimates aren't.
**Alternatives considered:** Pre-execution cost estimates per step.

### D9: Same Repo, Separate Package (not Monorepo)

**Decision:** Orchestrator lives as a subdirectory with its own package.json inside the ai-setup repo. Not a formal monorepo.
**Rationale:** Monorepo refactor (turborepo/nx) is heavyweight for just 2 packages. Simple subdirectory works now. Convert to monorepo if/when more packages are added.
**Alternatives considered:** Formal monorepo with turbo/nx, completely separate repo.

### D10: Extend Existing CLI Commands (not New Subcommand)

**Decision:** Add `domain`, `mode`, `chain`, `team`, `workflow` to existing `create`/`list`/`info` commands. Plus a thin `orchestration` alias for discoverability.
**Rationale:** Consistent with existing patterns, no new commands to learn. The alias provides a namespace without forcing users into it.
**Alternatives considered:** Dedicated `ai-setup orchestration` subcommand for everything.

### D11: Hybrid Build-Time + Runtime Composition

**Decision:** ai-setup generates templates (definition files), MCP server composes at runtime.
**Rationale:** Pure build-time leads to combinatorial explosion (every base × skill combo). Pure runtime requires a running server for everything. Hybrid gives the best of both: static files for inspection/customization, dynamic composition for execution.

### D12: Mix of File-Based + MCP Resource Context Passing

**Decision:** Files for simple chains, MCP resources for cross-tool chains. Session continuity as primary mechanism within a single CLI tool.
**Rationale:** Multiple proven patterns exist. Using session continuity (subagent results return to orchestrator) as primary, with MCP state for metadata, gives the best ergonomics.

---

## 18. What Does NOT Change

**Everything that exists today in ai-setup continues to work exactly as it does:**

- All existing fragments (RPI, CoT, ToT, quality gates) → UNCHANGED
- All existing constitution files → UNCHANGED
- All existing skills (9: anti-speculation, extract-standards, implement, iterate, memory-write, parallel-execution, plan, research, tdd-loop) → UNCHANGED
- All existing agents (7 base definitions) → UNCHANGED (except orchestrator enhancement with MCP tool awareness)
- All existing templates (10) → UNCHANGED
- All existing rules (10) → UNCHANGED
- All existing MCP servers (11 in catalog) → UNCHANGED
- Compile flow → UNCHANGED
- Migration engine → UNCHANGED
- Update/doctor/status commands → UNCHANGED
- Root file compilation → UNCHANGED
- Import/eject commands → UNCHANGED

**The orchestration is purely additive. Nothing breaks if you don't enable it.**

### Modification Summary

| Category | Count |
|----------|-------|
| Existing files modified | 17 files (~400 lines total) |
| New files in ai-setup | ~20 files (definitions + scaffold) |
| New package (orchestrator/) | ~35 files (MCP server) |
| Existing files unchanged | Everything else |

### New Files in ai-setup

| File | Purpose |
|------|---------|
| `src/scaffold/orchestration.ts` | Scaffold `.ai/orchestration/` directory |
| `src/__tests__/orchestration.test.ts` | Test orchestration scaffolding |
| `library/orchestration/chains/*.json` | 6 chain definitions |
| `library/orchestration/teams/*.json` | 3 team definitions |
| `library/orchestration/workflows/*.json` | 8 workflow definitions |
| `library/orchestration/skills/domains/*.md` | 5 domain skills |
| `library/orchestration/skills/modes/*.md` | 3 mode skills |

### Modified Files in ai-setup

| File | Change | Lines |
|------|--------|-------|
| `src/types.ts` | Add orchestration-related types | ~10 |
| `src/presets.ts` | Add 'orchestration' to feature flags | ~5 |
| `src/store/schema.ts` | Add orchestration config to Zod schema | ~15 |
| `src/commands/init.ts` | Add `--enable-orchestration` flag | ~10 |
| `src/commands/list.ts` | Add chains/teams/domains/modes/workflows | ~30 |
| `src/commands/info.ts` | Show details for new categories | ~40 |
| `src/commands/completions.ts` | Add new categories | ~10 |
| `src/wizard/index.ts` | Call scaffoldOrchestration() | ~10 |
| `src/wizard/phase1-context.ts` | Add orchestration question | ~15 |
| `src/adapters/claude-code.ts` | Enhanced orchestrator agent file | ~20 |
| `src/adapters/opencode.ts` | Enhanced orchestrator agent file | ~20 |
| `src/adapters/codex.ts` | Generate orchestrator.toml | ~25 |
| `src/adapters/gemini.ts` | Orchestrator as skill | ~10 |
| `src/adapters/copilot.ts` | Orchestrator as prompt | ~10 |
| `library/mcp/catalog.json` | Add orchestrator server entry | ~8 |
| `library/agents/orchestrator.md` | Enhance with MCP tool awareness | ~60 |
| `README.md` | Add orchestration documentation | ~100 |

---

## 19. Open Questions for Planning Phase

### Architecture & Implementation

1. **Chain definition format finalization** — The JSON structure shown in the blueprint needs formal schema validation. Should chains support JSONSchema for step outputs?
2. **Workflow-to-chain-to-team wiring** — How exactly does a workflow definition reference chains and teams? Inline or by name reference?
3. **MCP server session management** — How does the server handle multiple concurrent chains? Multiple users? Session isolation?
4. **Testing chains/teams without burning tokens** — How do we test the state machine, composition, and recovery logic without calling real LLMs? Mock agents? Recorded sessions?
5. **MCP server TypeScript implementation** — Confirm MCP SDK compatibility, server setup patterns, tool registration API.

### Agent Composition

6. **Compose output format** — What exactly does `compose_agent` return? Full system prompt text? Or structured parts the CLI tool assembles?
7. **Tool-specific agent format differences** — Claude `.md` vs Codex `.toml` vs OpenCode `.md` — how does the MCP server handle format translation?
8. **Skill injection depth** — How many skills can stack before the context window is overwhelmed? Is there a practical limit?

### User Experience

9. **When does the orchestrator auto-suggest workflows vs wait for explicit request?** Should it proactively say "This looks like a feature task, want me to run the RPI workflow?"
10. **How verbose should chain progress be?** Show every step transition? Only gates? Configurable verbosity?
11. **Offline mode** — Can the orchestrator work without the MCP server (degraded, manual mode)?

### Operational

12. **Versioning strategy** — How do `@ai-setup/cli` and `@ai-setup/orchestrator` versions relate?
13. **Definition file migration** — When chain/team/workflow formats change, how do we migrate user-customized definitions?
14. **Community definitions** — Can users share custom chains/teams/workflows? Import from a registry?

### From Earlier Research

15. Should the orchestrator decide when to escalate from chain to team, or is that always the user's call?
16. What's the minimum viable set of domain skills to ship in v1?
17. How does the orchestrator handle tool-specific subagent limitations (Gemini has no subagents)?

---

## Sources

| # | Source | Title |
|---|--------|-------|
| 15 | Obsidian Note | Agent Teams & External Orchestration Research |
| 16 | Obsidian Note | Composable Agent Orchestration Research |
| 17 | Obsidian Note | Seven Questions Deep Research |
| 18 | Obsidian Note | Orchestration Design Research (Pre-Planning) |
| 19 | Obsidian Note | Orchestrator Product Fit Analysis |
| 20 | Obsidian Note | Orchestrator Agent + MCP Architecture |
| 21 | Obsidian Note | Orchestrator Blueprint (Visual/Summary) |
| 22 | Obsidian Note | Workflows, CLI Commands & Domain Design Research |
| 23 | Obsidian Note | Design Decision — Separate Composable Pieces |
| — | Blueprint | `docs/orchestrator-blueprint.md` |

---

## Appendix A: Gap Research (G1-G7)

> These gaps were identified by comparing our research against a comprehensive design doc structure.
> All gaps have been researched and resolved.

### G1: Compilation Model — Definitions → Execution Plans

**Gap:** We assumed MCP server reads definitions directly. The prompt asks for a separate "compilation" stage.

**Finding:** Mastra uses `.commit()` to compile workflow definitions into executable plans. We should do the same.

**Design:** Definitions (static JSON) → Compile at `start_chain` time → Execution Plan (in-memory). Compilation injects:
- CLI tool context (which dispatch mechanism)
- Project stack (detected language/framework)
- Token budget constraints
- Available tools for this CLI

### G2: Output Contracts — Structured Step Outputs

**Gap:** Steps had `output: "description"` but no typed schemas.

**Design:** Each step declares required output fields:

| Step Type | Required Output Fields |
|---|---|
| research | `findings`, `files_examined`, `patterns` |
| plan | `plan`, `tasks`, `risks` |
| implement | `files_changed`, `tests_passed`, `lines_changed` |
| review | `verdict` (pass/minor/blocking), `findings[]` |
| document | `files_created`, `sections` |

Orchestrator validates output against contract before advancing.

### G3: Deterministic Composition — Merge Algorithm

**Gap:** No defined merge order or conflict resolution.

**Design:** 5-layer merge with strict precedence (highest wins):

1. Root file context (inherited, project-level CE)
2. Base agent definition (role identity)
3. Domain skill (knowledge — EXTENDS)
4. Mode skill (behavior — MODIFIES)
5. Chain step instructions (OVERRIDES)

Merge rules:
- **Prompts:** concatenate in order
- **Tool allowlist:** intersection only (step can restrict, never add)
- **Model:** most specific wins (step > mode > base)
- **Constraints:** union (all apply)
- **Output contract:** step is authoritative

### G4: Security and Trust Boundaries

**Gap:** Almost no security analysis.

**Design:** Three trust levels:

| Entity | Can | Cannot | Enforced By |
|---|---|---|---|
| MCP server | Read definitions, manage state, compose prompts | Execute tools, write files, call LLMs | Architecture (no tool access) |
| Orchestrator agent | Call MCP tools, dispatch subagents | Bypass user gates, exceed budget | Agent prompt + CLI permissions |
| Subagents | Use composed tool allowlist only | Use disallowed tools | CLI tool's subagent system |

Key insight: MCP server DECLARES tool restrictions. CLI tool ENFORCES them. No enforcement burden on our code.

### G5: Formal State Machine Model

**Gap:** Conceptual transitions but no formal state model.

**Design:**

Chain states: `CREATED → RUNNING → GATED → PAUSED → COMPLETED | ABANDONED | HANDOFF`

Step states: `PENDING → RUNNING → COMPLETED | FAILED → RETRYING | ESCALATED | ABANDONED`

Terminal states: `COMPLETED`, `ABANDONED`, `HANDOFF`

Full TypeScript interfaces defined for `ChainState`, `StepState`, `TeamState`.

### G6: Structured Error Taxonomy

**Gap:** 4 recovery patterns but no error classification.

**Design:** 6 error categories with default recovery actions:

| Category | Example | Default Recovery |
|---|---|---|
| `transient` | Network timeout, rate limit | Retry (max 2) |
| `logical` | Wrong approach | Escalate to different agent |
| `budget` | Token limit | Pause + ask user |
| `permission` | Tool denied | Abort step + report |
| `validation` | Output contract violation | Retry with guidance |
| `fatal` | Unrecoverable | Handoff to new session |

### G7: Tool Visibility Per Agent Role

**Gap:** Mentioned conceptually but not designed.

**Design:** Tool allowlists per agent role:

| Agent | Tools | Rationale |
|---|---|---|
| Scout | Read, Grep, Glob, LSP | Research only |
| Planner | Read, Grep, Glob, Write (specs/) | Plans only |
| Builder | Read, Edit, Write, Bash, Grep, Glob | Full access |
| Reviewer | Read, Grep, Glob, LSP | Review only |
| Red-Team | Read, Bash (tests), Grep, Glob | Test + audit |
| Documenter | Read, Write, Edit, Grep, Glob | Docs only |
| Orchestrator | Agent, MCP tools, Read | Coordination |

Enforcement: CLI tool's native subagent system (Claude Code frontmatter, Codex TOML sandbox_mode).

