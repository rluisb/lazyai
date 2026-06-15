# Retired Orchestrator Blueprint

> Historical only. Spec 025 removed this runtime from the active product surface. This blueprint is preserved to explain earlier design intent; it is not current setup guidance.

---

## Current Structure → With Orchestrator

```
ai-setup/
├── src/                                    ✅ UNCHANGED (mostly)
│   ├── cli.ts                              ✅
│   ├── index.ts                            ✅
│   ├── types.ts                            🟡 Add orchestration types
│   ├── presets.ts                           🟡 Add 'orchestration' feature flag
│   │
│   ├── commands/
│   │   ├── init.ts                         🟡 Add --enable-orchestration flag
│   │   ├── compile.ts                      ✅
│   │   ├── add.ts                          ✅
│   │   ├── update.ts                       ✅
│   │   ├── doctor.ts                       ✅
│   │   ├── status.ts                       ✅
│   │   ├── create.ts                       ✅
│   │   ├── list.ts                         🟡 Add 'chains', 'teams', 'domains' categories
│   │   ├── info.ts                         🟡 Show chain/team/domain details
│   │   ├── import.ts                       ✅
│   │   ├── migrate.ts                      ✅
│   │   ├── eject.ts                        ✅
│   │   └── completions.ts                  🟡 Add new categories
│   │
│   ├── adapters/
│   │   ├── claude-code.ts                  🟡 Install orchestrator agent with MCP-aware frontmatter
│   │   ├── opencode.ts                     🟡 Install orchestrator agent with MCP-aware frontmatter
│   │   ├── codex.ts                        🟡 Generate .codex/agents/orchestrator.toml
│   │   ├── gemini.ts                       🟡 Install orchestrator as skill (no subagent support)
│   │   ├── copilot.ts                      🟡 Install orchestrator as prompt (limited)
│   │   ├── mcp-compiler.ts                 🟡 Handle orchestrator MCP server entry
│   │   ├── registry.ts                     ✅
│   │   ├── shared.ts                       ✅
│   │   ├── types.ts                        ✅
│   │   └── pi.ts                           ✅ (dormant)
│   │
│   ├── scaffold/
│   │   ├── agents-skills-prompts.ts        ✅
│   │   ├── compiled-root.ts                ✅
│   │   ├── constitution.ts                 ✅
│   │   ├── mcp.ts                          ✅ (catalog entry handles it)
│   │   ├── orchestration.ts                🔴 NEW — scaffold .ai/orchestration/ directory
│   │   ├── specs.ts                        ✅
│   │   ├── templates-rules.ts              ✅
│   │   └── ...                             ✅
│   │
│   ├── wizard/
│   │   ├── index.ts                        🟡 Call scaffoldOrchestration() when enabled
│   │   ├── phase1-context.ts               🟡 Add orchestration prompt/question
│   │   ├── phase2-features.ts              ✅
│   │   └── ...                             ✅
│   │
│   ├── store/
│   │   └── schema.ts                       🟡 Add orchestration config to schema
│   │
│   └── __tests__/
│       ├── adapters-files.test.ts          🟡 Add orchestration test cases
│       └── orchestration.test.ts           🔴 NEW — test scaffolding
│
├── library/                                ✅ UNCHANGED (mostly)
│   ├── agents/
│   │   ├── builder.md                      ✅
│   │   ├── documenter.md                   ✅
│   │   ├── orchestrator.md                 🟡 Enhance with MCP tool awareness
│   │   ├── planner.md                      ✅
│   │   ├── red-team.md                     ✅
│   │   ├── reviewer.md                     ✅
│   │   └── scout.md                        ✅
│   │
│   ├── skills/
│   │   ├── anti-speculation.md             ✅
│   │   ├── extract-standards.md            ✅
│   │   ├── implement.md                    ✅
│   │   ├── iterate.md                      ✅
│   │   ├── memory-write.md                 ✅
│   │   ├── parallel-execution.md           ✅
│   │   ├── plan.md                         ✅
│   │   ├── research.md                     ✅
│   │   └── tdd-loop.md                     ✅
│   │
│   ├── orchestration/                      🔴 NEW — orchestration definitions
│   │   ├── chains/
│   │   │   ├── feature.json                🔴 Feature development chain
│   │   │   ├── bugfix.json                 🔴 Bug resolution chain
│   │   │   ├── review.json                 🔴 Code review chain (council pattern)
│   │   │   ├── refactor.json               🔴 Refactoring chain
│   │   │   ├── onboard.json                🔴 Codebase onboarding chain
│   │   │   └── tdd.json                    🔴 Test-driven development chain
│   │   │
│   │   ├── teams/
│   │   │   ├── review-team.json            🔴 Parallel reviewers + red-team
│   │   │   ├── feature-team.json           🔴 Parallel research → sequential build
│   │   │   └── assessment-team.json        🔴 Parallel codebase assessment
│   │   │
│   │   └── skills/
│   │       ├── domains/
│   │       │   ├── backend.md              🔴 Backend development knowledge
│   │       │   ├── frontend.md             🔴 Frontend development knowledge
│   │       │   ├── fullstack.md            🔴 Full-stack patterns
│   │       │   ├── mobile.md               🔴 Mobile development patterns
│   │       │   ├── devops.md               🔴 Infrastructure & CI/CD knowledge
│   │       │   ├── data.md                 🔴 Data engineering knowledge
│   │       │   └── security.md             🔴 Security engineering knowledge
│   │       └── modes/
│   │           ├── senior.md               🔴 Autonomous, proactive, opinionated
│   │           ├── junior.md               🔴 Cautious, asks before acting
│   │           └── autonomous.md           🔴 Full autonomy, minimal confirmation
│   │
│   ├── mcp/
│   │   └── catalog.json                    🟡 Add orchestrator server entry
│   │
│   ├── fragments/                          ✅ ALL UNCHANGED
│   ├── constitution/                       ✅ ALL UNCHANGED
│   ├── prompts/                            ✅
│   ├── root/                               ✅
│   ├── rules/                              ✅
│   ├── specs-agents/                       ✅
│   ├── templates/                          ✅
│   ├── tool-agents/                        ✅
│   ├── tool-templates/                     ✅
│   └── infra/                              ✅
│
├── orchestrator/                           🔴 NEW — separate npm package
│   ├── package.json                        🔴 @ai-setup/orchestrator
│   ├── tsconfig.json                       🔴
│   ├── tsup.config.ts                      🔴
│   │
│   └── src/
│       ├── index.ts                        🔴 MCP server entry point
│       ├── server.ts                       🔴 MCP server setup (tools, resources)
│       │
│       ├── tool-handlers.ts                🔴 MCP tool implementations (all 9 tools)
│       │                                      list_catalog, compose_agent,
│       │                                      start_chain, advance_chain,
│       │                                      get_status, get_budget,
│       │                                      retry_step, escalate_step, handoff
│       │
│       ├── state/                          🔴 State management
│       │   ├── chain-machine.ts            🔴 Chain state machine (dynamic transitions)
│       │   ├── team-state.ts               🔴 Team task list + assignments
│       │   ├── budget-tracker.ts           🔴 Token budget tracking
│       │   ├── error-journal.ts            🔴 Error history + lessons
│       │   └── persistence.ts             🔴 Save/load state to .ai/orchestration/state/
│       │
│       ├── composer/                       🔴 Agent composition logic
│       │   ├── compose.ts                  🔴 Merge base + domain + mode → prompt
│       │   ├── loader.ts                   🔴 Read definition files from disk
│       │   └── validator.ts                🔴 Validate composed agent
│       │
│       └── __tests__/                      🔴 Tests
│           ├── chain-machine.test.ts       🔴
│           ├── compose.test.ts             🔴
│           ├── team-state.test.ts          🔴
│           └── server.test.ts              🔴
│
├── demo/                                   ✅
├── docs/                                   ✅
├── scripts/                                ✅
├── specs/                                  ✅
├── bin/                                    ✅
├── package.json                            ✅
├── tsconfig.json                           ✅
├── tsup.config.ts                          ✅
└── README.md                               🟡 Add orchestration docs
```

---

## What ai-setup Scaffolds When Orchestration Is Enabled

When user runs `ai-setup init --enable-servers orchestrator` (or checks the box in the wizard):

```
<project>/
├── .ai/
│   ├── constitution/                       ✅ (already created)
│   ├── mcp.json                            🟡 (orchestrator entry added)
│   │
│   └── orchestration/                      🔴 NEW
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
│       ├── skills/
│       │   ├── domains/
│       │   │   ├── backend.md
│       │   │   ├── frontend.md
│       │   │   ├── fullstack.md
│       │   │   ├── mobile.md
│       │   │   ├── devops.md
│       │   │   ├── data.md
│       │   │   └── security.md
│       │   └── modes/
│       │       ├── senior.md
│       │       ├── junior.md
│       │       └── autonomous.md
│       └── state/                          (empty, created at runtime by MCP server)
│
├── .ai-setup.json                          🟡 (orchestration files tracked)
│
│   ── Per-tool output ──
│
├── .opencode/
│   └── agents/
│       └── orchestrator.md                 🟡 Enhanced with MCP tool instructions
│
├── .claude/
│   └── agents/
│       └── orchestrator.md                 🟡 Enhanced with MCP tool instructions
│
├── .agents/                                (Codex)
│   └── skills/
│       └── orchestrator/
│           └── SKILL.md                    🟡 Orchestrator as skill for Codex
│
├── .opencode/opencode.jsonc               🟡 orchestrator MCP server added
├── .mcp.json                              🟡 orchestrator MCP server added
└── .vscode/mcp.json                       🟡 orchestrator MCP server added
```

---

## Chain Definition Format

```json
// library/orchestration/chains/feature.json
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

---

## Team Definition Format

```json
// library/orchestration/teams/review-team.json
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

---

## Orchestrator Agent (Enhanced)

```markdown
<!-- library/agents/orchestrator.md — what changes -->
---
name: Orchestrator
model: opus
tools: list_catalog compose_agent start_chain advance_chain get_status get_budget retry_step escalate_step handoff
description: Coordinates agent chains. Use when tasks require
  multiple agents, structured workflows, or recoverable multi-step execution.
---

# Orchestrator Agent

## Identity
You coordinate agents through chains (sequential execution with recovery).
You do not write code, review code, or make architecture decisions.

## MCP Tools Available

You have access to these orchestration tools via MCP (9 total, matching `orchestrator/src/server.ts`):

| Tool | What it does |
|------|-------------|
| `list_catalog` | Browse available chains, teams, workflows, domains, modes |
| `compose_agent` | Combine base agent + domain skill + mode → specialized prompt |
| `start_chain` | Begin a predefined chain (feature, bugfix, review, etc.) |
| `advance_chain` | Move chain to next step after current step completes |
| `get_status` | Check chain progress and step history |
| `get_budget` | Check remaining token budget |
| `retry_step` | Retry a failed step |
| `escalate_step` | Route to a different agent for re-evaluation |
| `handoff` | Persist a resumable handoff document |

## How Chains Work

1. Call `start_chain("feature", {task: "..."})` → get chainId + first step
2. Dispatch the agent for the current step (as a subagent)
3. When step completes, call `advance_chain(chainId, {result: ...})`
4. MCP returns next step (may include user gate, retry, or escalation)
5. Repeat until chain reaches "done"

If a step fails, MCP returns recovery options:
- retry: run same step again
- escalate_to: route to different agent
- handoff: save context for new session

## When to Use Chains vs Teams

**Default to chains** (sequential). Use teams only when:
- Work is truly independent and parallel saves time
- Multiple perspectives on the SAME problem are valuable
- User explicitly asks for a team

**Always ask user before spawning a team:**
> "This task could benefit from parallel review (3 agents). 
>  Estimated cost: ~3× a single review. Proceed? [y/N]"

## Token Budget

Before starting any chain or team:
1. Call `get_budget()` to check remaining budget
2. Show user the estimated cost range
3. Get confirmation before proceeding
4. Track actual spend during execution

## Error Recovery (4 Patterns)

| Pattern | When | What happens |
|---------|------|-------------|
| **Retry** | Transient failure | Same step runs again (max 2 retries) |
| **Fix & Resume** | User fixes issue | Chain resumes from failed step |
| **Escalate** | Wrong approach | Routes to different agent (e.g., Planner) |
| **Handoff** | Fundamental issue | Full context dump, new session |

After any failure:
1. Show: chain, step, agent, skills, exact error
2. Show: what succeeded so far
3. Show: recommended action
4. Write lesson to error journal via MCP
```

---

## What Changes in Existing Files (Summary)

### Modifications to Existing Code

| File | Change | Size |
|------|--------|------|
| `src/types.ts` | Add orchestration-related types | ~10 lines |
| `src/presets.ts` | Add 'orchestration' to feature flags | ~5 lines |
| `src/store/schema.ts` | Add orchestration config to Zod schema | ~15 lines |
| `src/commands/init.ts` | Add `--enable-orchestration` flag | ~10 lines |
| `src/commands/list.ts` | Add chains/teams/domains categories | ~30 lines |
| `src/commands/info.ts` | Show chain/team/domain details | ~40 lines |
| `src/commands/completions.ts` | Add new categories | ~10 lines |
| `src/wizard/index.ts` | Call scaffoldOrchestration() | ~10 lines |
| `src/wizard/phase1-context.ts` | Add orchestration question | ~15 lines |
| `src/adapters/claude-code.ts` | Enhanced orchestrator agent file | ~20 lines |
| `src/adapters/opencode.ts` | Enhanced orchestrator agent file | ~20 lines |
| `src/adapters/codex.ts` | Generate orchestrator.toml | ~25 lines |
| `src/adapters/gemini.ts` | Orchestrator as skill | ~10 lines |
| `src/adapters/copilot.ts` | Orchestrator as prompt | ~10 lines |
| `library/mcp/catalog.json` | Add orchestrator server entry | ~8 lines |
| `library/agents/orchestrator.md` | Enhance with MCP tool awareness | ~60 lines |
| `README.md` | Add orchestration documentation | ~100 lines |

**Total modifications: ~400 lines across 17 files**

### New Files in ai-setup

| File | Purpose |
|------|---------|
| `src/scaffold/orchestration.ts` | Scaffold .ai/orchestration/ directory |
| `src/__tests__/orchestration.test.ts` | Test orchestration scaffolding |
| `library/orchestration/chains/*.json` | 6 chain definitions |
| `library/orchestration/teams/*.json` | 3 team definitions |
| `library/orchestration/skills/domains/*.md` | 7 domain skills |
| `library/orchestration/skills/modes/*.md` | 3 mode skills |

**Total new in ai-setup: ~20 new files**

### New Package: orchestrator/

| Directory | Files | Purpose |
|-----------|-------|---------|
| `orchestrator/src/` | ~12 files | MCP server implementation |
| `orchestrator/src/tools/` | ~10 files | MCP tool handlers |
| `orchestrator/src/state/` | ~5 files | State management |
| `orchestrator/src/composer/` | ~3 files | Agent composition |
| `orchestrator/src/__tests__/` | ~4 files | Tests |

**Total new package: ~35 files**

---

## What Does NOT Change

- All existing fragments (RPI, CoT, ToT, quality gates) → UNCHANGED
- All existing constitution files → UNCHANGED  
- All existing skills (9) → UNCHANGED
- All existing agents (7 base definitions) → UNCHANGED (except orchestrator enhancement)
- All existing templates (10) → UNCHANGED
- All existing rules (10) → UNCHANGED
- All existing MCP servers (11) → UNCHANGED
- Compile flow → UNCHANGED
- Migration engine → UNCHANGED
- Update/doctor/status commands → UNCHANGED
- Root file compilation → UNCHANGED

**The orchestration is purely additive. Nothing breaks if you don't enable it.**
