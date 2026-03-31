# AI Agentic Setup Playbook

> **Author:** Ricardo & Team
> **Created:** 2026-03-28 | **Updated:** 2026-03-28
> **Status:** Living Document — Concept Reference
> **Audience:** Engineering Team
> **Companion docs:** [`Implementation Plan`](./AI-Agentic-Setup-Implementation-Plan.md) · [`Templates`](./AI-Agentic-Setup-Templates/)

---

## How to Read This Document

**8 clusters**, each building on the previous. Every concept includes WHAT, WHY, WHERE, WHEN, HOW.

This is the **concepts document** (WHY/WHAT). For execution steps, see the Implementation Plan. For ready-to-copy files, see the Templates directory.

---

## Table of Contents

- [Cluster 1: Foundation — Prompt Engineering](#cluster-1-foundation--prompt-engineering)
- [Cluster 2: Context Engineering](#cluster-2-context-engineering)
- [Cluster 3: Agent Workflow Patterns](#cluster-3-agent-workflow-patterns)
- [Cluster 4: Sub-Agents & Multi-Agent](#cluster-4-sub-agents--multi-agent)
- [Cluster 5: Safety, Quality Gates & Guardrails](#cluster-5-safety-quality-gates--guardrails)
- [Cluster 6: RAG, MCPs & External Context](#cluster-6-rag-mcps--external-context)
- [Cluster 7: Team Setup & Adoption](#cluster-7-team-setup--adoption)
- [Cluster 8: Implementation Plan](#cluster-8-implementation-plan)
- [Quick Reference Card](#quick-reference-card)
- [Glossary](#glossary)
- [References & Sources](#references--sources)

---

## Cluster 1: Foundation — Prompt Engineering

> The 8 building blocks every agent, rule, and skill is built on.

### 1.1 System Message

The invisible instruction before every conversation. Tells the AI who it is and how to behave. In our setup → `AGENTS.md`, `CLAUDE.md`, `.pi/agents/*.md`.

> System prompt = the DNA of every agent.

### 1.2 The 9-Section Blueprint

| # | Section | Skip Risk |
|---|---------|-----------|
| ① | Identity & Role | AI acts generic |
| ② | Personality & Tone | Inconsistent voice |
| ③ | Knowledge & Specialties | Wanders into unknown territory |
| ④ | Response Style | Walls of unstructured text |
| ⑤ | Specific Guidelines | Invents its own rules |
| ⑥ | Limitations | Never says "I don't know" |
| ⑦ | Output Format | Unparseable responses |
| ⑧ | Examples (Few-Shot) | Ambiguous pattern interpretation |
| ⑨ | Fallback Behavior | Silent guessing |

**Minimum:** ①②⑤. **Production agents:** All 9.

### 1.3 Roles API

| Role | Purpose | Weight |
|------|---------|--------|
| **System** | Hidden instructions | Highest — AI *obeys* |
| **User** | What the human says | Medium — AI *serves* |
| **Assistant** | Conversation history | Context — AI *remembers* |

> System = law. User = request. Assistant = memory. Don't mix.

### 1.4 Temperature & Parameters

| Parameter | Code | Brainstorm |
|-----------|------|-----------|
| Temperature | 0.0–0.3 | 0.7–1.0 |
| Top-P | 0.9 | 0.9 |

> Never above 1.0 in production.

### 1.5 Delimiters & XML Tags

Use XML tags to separate instruction from data. Our setup uses `<rule>`, `<template>`, `<thinking>` tags throughout.

### 1.6 Specific Verbs

| ❌ Vague | ✅ Specific |
|----------|------------|
| "Help me with" | "Diagnose the root cause of" |
| "Improve" | "Refactor to reduce cyclomatic complexity" |
| "Check" | "Validate against the OpenAPI spec" |

> One specific verb per instruction.

### 1.7 Structured Output

Define the output shape. The AI fills it. No shape = no consistency.

### 1.8 Conditional Prompts

If/then logic in rules. One file, multiple behaviors based on file type, task size, or situation. Max 2 nesting levels.

---

## Cluster 2: Context Engineering

> The most important cluster. Everything else depends on this.

### 2.1 What Is Context?

Everything the AI can see: system prompt, conversation, files read, command output, tool results. Nothing else exists.

### 2.2 The Context Window

| Model | Raw | Practical |
|-------|-----|-----------|
| Claude Sonnet | 200K | ~120K |
| GPT-5 | 200K | ~130K |

Quality degrades gradually, not at a cliff.

### 2.3 What Eats Context

Every message, file, and command output consumes tokens. Silent killers: log dumps, unnecessary file reads, verbose AI responses.

### 2.4 Context Rot

Solved problems and wrong turns still taking up space. The AI can't tell old from new.

### 2.5 The 40-60% Rule

| Zone | Effect |
|------|--------|
| Below 40% | Starved — guesses |
| **40-60%** | **Sharp** |
| Above 70% | Drops early instructions |
| Above 90% | Unreliable |

### 2.6 Frequent Intentional Compaction (FIC)

Pause → write progress.md → fresh session → load progress.md. Clean context every time.

### 2.7 The Four Pillars of Context Quality

| Pillar | Question |
|--------|----------|
| Correctness | Is it true right now? |
| Completeness | Does the AI have everything it needs? |
| Size | Is it as small as possible? |
| Trajectory | Does it point toward the goal? |

### 2.8 Decision Tree (Progressive Documentation Loading)

A single section in root `AGENTS.md` that tells the AI **exactly what to read based on the task type.** Without it, the AI loads everything or guesses.

```
Writing code?     → docs/standards/coding/ + docs/rules/code-style.md
Writing a PRD?    → research.md + docs/templates/prd-template.md
Writing tests?    → docs/standards/testing/ + docs/rules/testing.md
Reviewing code?   → docs/rules/review.md + docs/standards/
Don't know yet?   → docs/KNOWLEDGE_MAP.md
```

**Why it's critical:** Second most important thing in the setup (after project overview). Controls what enters context. Prevents wrong-context contamination. Makes the system scale.

### 2.9 Documentation as Context

| Level | File | Scope |
|-------|------|-------|
| Project rules | `AGENTS.md` + `CLAUDE.md` | This repo |
| Modular rules | `docs/rules/*.md` | By topic |
| Standards | `docs/standards/[category]/*.md` | By pattern type |
| Personal prefs | `CLAUDE.local.md` | You only (gitignored) |
| Memory | `docs/memory/` | Agent learnings |

> If you explained it twice, it belongs in a rules file. That's the feedback loop.

---

## Cluster 3: Agent Workflow Patterns

> Where context engineering becomes a system.

### 3.1 The RPI Pattern (Research → Plan → Implement)

Three phases. Fresh sessions. Artifacts carry forward, noise doesn't.

**Research (Scout):** Map what exists → `docs/features/NNN/research.md`
**Plan (Planner):** PRD → TechSpec → Tasks → `docs/features/NNN/`
**Implement (Builder):** One task per session → code changes

### 3.2 The RPI Steps (Expanded)

```
Research ──GATE──▶ PRD ──GATE──▶ TechSpec ──GATE──▶ Tasks ──GATE──▶ Task Files ──GATE──▶ Implement
 (Scout)          (WHAT)         (HOW)           (ORDER)         (DETAIL)           (Builder)
```

Each gate = human approval. Each arrow = fresh session with artifact file.

### 3.3 PRD, TechSpec, Tasks — How They Relate

| Artifact | Job | Focus | Template |
|----------|-----|-------|----------|
| `prd.md` | WHAT and WHY | Problem, goals, user stories, scope | prd-template.md |
| `techspec.md` | HOW | Architecture, patterns, data model, tests | techspec-template.md |
| `tasks/tasks.md` | IN WHAT ORDER | Phases, dependencies, parallel markers | tasks-template.md |
| `tasks/NNN-*.md` | INDIVIDUAL CONTEXT | Subtasks, patterns to follow, done criteria | task-template.md |

### 3.4 ADRs (Architecture Decision Records)

Short docs that capture WHY a decision was made. Created during TechSpec when a non-obvious choice is made. Live permanently in `docs/adrs/`. Never edited — superseded.

### 3.5 The Ralph Loop

Worker builds → Reviewer judges → Context resets → Repeat until SHIP. No memory of past failures.

### 3.6 Work Types

**Proactive work** (you initiate):

| Type | Flow | PRD? | ADR? |
|------|------|------|------|
| **Feature** | Full RPI | Yes | If applicable |
| **Bugfix (P2/P3)** | Shortened: RCA → TechSpec → Tasks | No | Rarely |
| **Refactor MICRO** | Document intent → Implement → Review | No | No |
| **Refactor MACRO** | Full RPI + Compatibility Matrix | Yes | **Always** |
| **Tech Debt** | Shortened: Research → TechSpec → Tasks | No | If applicable |

**Reactive work** (something happened, you respond):

| Type | Trigger | Flow |
|------|---------|------|
| **Change Request** (`/change-request`) | PR review feedback received | Triage → Size Gate → Implement → CR Log |
| **Code Review** (`/code-review`) | Teammate's PR needs review | Context Assembly → Reviewer (External Mode) → Structured Output |
| **Hotfix** (`/hotfix`) | P0/P1 production incident | Reproduce → RCA → Implement → Expedited Review → Deploy → Post-mortem |
| **Bugfix (P0/P1)** | Severity-triaged bug | Same as Hotfix path |

### 3.7 The 7 Canonical Workflow Patterns

| # | Pattern | Default? |
|---|---------|----------|
| ① Chain | A → B → C → D |
| ② Parallel + Judge | Fan out → judge picks best |
| ③ Router | Decision node → different paths |
| **④ ReAct** | **Think → Tool → Observe → Repeat** | **Default** |
| ⑤ Orchestrator → Workers | Brain decomposes, specialists execute |
| ⑥ Dynamic Workers | Count decided at runtime |
| ⑦ Optimizer | Loop until quality threshold |

> 80% of tasks = ReAct inside RPI. Multi-agent when single-agent hits a wall.

---

## Cluster 4: Sub-Agents & Multi-Agent

> When one agent isn't enough — and when one IS enough.

### 4.1 Why Sub-Agents

Protect main agent's context from noise. Sub-agent searches → returns summary → dies.

### 4.2 Three Properties

Own system prompt. Own context window. Restricted tools.

### 4.3 The 6 Agent Roles

| Role | Job | CoT? | Key Feature |
|------|-----|------|-------------|
| **Scout** | Map codebase, find patterns | `<thinking>` — where to look | Read-only, neutral |
| **Planner** | PRD, TechSpec, Tasks | `<thinking>` — full ToT reasoning | Questions first, templates |
| **Builder** | Follow plan, write code | No (mechanical) | Standards before writing |
| **Reviewer** | Find issues, report | `<thinking>` — severity reasoning | Conformance check |
| **Red-Team** | Break code, find holes | `<thinking>` — attack planning | Systematic vectors |
| **Documenter** | Write docs | No (descriptive) | Creates new standards |

### 4.4 Chain of Thought in Agents

Planner, Reviewer, and Red-Team use `<thinking>` blocks before producing output. Forces reasoning before acting.

### 4.5 Tree of Thoughts in TechSpec

The Approach Options section forces the Planner to explore ≥2 approaches, evaluate each, and choose with justification. Rejected options feed the ADR's "Alternatives Considered."

### 4.6 Agent Progression

```
Level 1: One agent + good rules + RPI        → covers 80%
Level 2: + sub-agents for search/review       → covers 95%
Level 3: Agent chain (formalized RPI)         → covers 99%
Level 4: Agent team + orchestrator            → the remaining 1%
```

> One good agent + good context > five mediocre agents. Always.

---

## Cluster 5: Safety, Quality Gates & Guardrails

> The AI is confident and fast. Guardrails exist because confident + wrong + fast = damage.

### 5.1 The Guardrail Stack

```
Layer 1: PURPOSE GATE         → declare intent before work
Layer 2: TILLDONE             → list tasks before tools
Layer 3: PATH ACCESS          → can you touch this file?
Layer 4: DAMAGE CONTROL       → is this command safe?
Layer 5: HUMAN GATE           → should I continue?
Layer 6: LLM-AS-JUDGE         → is this output good?
```

Implementation priority: human gates → path access → purpose gate → TillDone → damage control → LLM-as-Judge.

### 5.2 Observability (progress.md)

Every agent, every step, appends an entry to `progress.md`. This is the audit trail:

```
### [YYYY-MM-DD HH:MM] — [Step] ([Agent])
- Agent: [name]
- Files read: [count]
- Files changed: [paths]
- Output: [artifact]
- Status: ✅ | ⏳ | 🚫
```

---

## Cluster 6: RAG, MCPs & External Context

> How the AI gets information not in your codebase.

### 6.1 MCPs — Priority

| Priority | MCP | When |
|----------|-----|------|
| 1 | Git | Recent changes |
| 2 | Jira/Linear | Ticket details |
| 3 | Database (read-only) | Schema questions |
| 4 | Browser/Docs | Library docs |

### 6.2 Context Decision Framework

```
1. Rules/AGENTS.md?  → Free
2. Codebase files?   → Cheap
3. Project docs?     → Moderate
4. MCP call?         → Watch tokens
5. Web search?       → Last resort
```

---

## Cluster 7: Team Setup & Adoption

> How to make it work for a whole team.

### 7.1 The Directory Split

| Directory | Contains | Survives tool switch? |
|-----------|----------|---------------------|
| `.pi/` | Agent mechanics (agents, skills, prompt templates) | No — pi-specific |
| `docs/` | Project knowledge (rules, standards, ADRs, features, memory) | **Yes** — project knowledge |

> Test: "If we switched tools tomorrow, would this file matter?" Yes → `docs/`. No → `.pi/`.

### 7.2 Rules vs Standards

| | Rules (`docs/rules/`) | Standards (`docs/standards/`) |
|---|---|---|
| **What** | Prescriptive: WHAT to do | Descriptive: HOW we do it |
| **Format** | ✅/❌ directives | Code examples from real codebase |
| **Example** | "Use Pydantic for validation" | "See src/auth/auth.service.ts for service pattern" |

### 7.3 Standards Categories (8)

```
docs/standards/
├── coding/          ← API, service, error handling patterns
├── architecture/    ← Module boundaries, cross-module communication
├── testing/         ← Unit, integration, e2e patterns
├── quality/         ← Naming, file organization, checklists
├── resilience/      ← Circuit breakers, timeouts, degradation
├── observability/   ← Logging, metrics, health checks
├── data/            ← Entity, migration, query patterns
└── security/        ← Auth, validation, secret handling
```

Standards are **extracted from real code, never invented.** Every standard references a real file.

### 7.4 Self-Improvement Protocol

Every agent runs an **Impact Check** after completing work:

```
Did my work change:
├── Project structure?     → update AGENTS.md codebase map
├── API contracts?         → update docs/standards/coding/
├── Architecture?          → create ADR
├── Testing patterns?      → update docs/standards/testing/
├── Dependencies?          → update AGENTS.md stack section
├── New pattern not in standards? → create new standard
└── Nothing?               → done
```

Output: `📋 Knowledge Updates Needed` — human decides whether to update now or later.

**Memory promotion path:**
```
docs/memory/ (advisory) → repeated 2x → promote to docs/standards/ or docs/rules/ → delete memory
```

### 7.5 Adoption Strategy

Phase 1: You alone (1 week). Phase 2: 2-3 devs (1-2 weeks). Phase 3: Full team (ongoing).

> Ship early, refine from real use. Never design in isolation.

### 7.6 KNOWLEDGE_MAP.md

Navigable index of all project docs with **bidirectional ADR ↔ Feature links.** AI reads this for orientation.

---

## Cluster 8: Implementation Plan

> 9 phases, each a checkpoint. See the full Implementation Plan document for execution steps.

| Phase | Goal | Who | Duration |
|-------|------|-----|----------|
| 1. Foundation | Rules, AGENTS.md, directory structure | You | Week 1 |
| 2. Workflow | Manual RPI practice (3 cycles) | You | Week 1-2 |
| 3. Guardrails | Safety layers tested | You | Week 2 |
| 4. Agent Roles | 6 agents tested in pipeline | You + 1 | Week 2-3 |
| 5. Standards Bootstrap | Extract patterns from codebase | You + 1 | Week 3 |
| 6. External Context | MCPs (Git, Jira, DB) | Small group | Week 3-4 |
| 7. Automation | Skills and commands | Small group | Week 4 |
| 8. Optimization | RTK, cost rules, metrics | Full team | Week 5 |
| 9. Team Rollout | Onboarding, demo, feedback loop | Full team | Week 5-6 |

---

## Quick Reference Card

```
CONTEXT
• 40-60% window = sharp AI
• New task = new session
• Compact at 15-20 exchanges
• Decision tree: load only what the task needs

RPI FLOW
• Research → PRD → TechSpec → Tasks → Task Files → Implement
• Human gate at every arrow
• Fresh session at every phase

FOUR PILLARS (before adding context)
• Correct? Complete? Small? Pointed?

GUARDRAIL CHECK
• Intent declared? Tasks listed? Path allowed? Command safe?

MODEL SELECTION
• Simple → Sonnet   Complex → Opus   Never Opus for trivial

STANDARDS
• Extracted from real code, never invented
• 8 categories: coding/arch/testing/quality/resilience/observability/data/security
• One pattern per file. Reference real files.

SELF-IMPROVEMENT
• Impact Check after every task
• New pattern? → create standard
• Bug from missing rule? → add rule
• Explained twice? → write it down

REACTIVE FLOWS (respond to events)
• PR feedback received?  → /change-request (triage → size gate → CR log)
• Teammate PR to review? → /code-review (context assembly → Reviewer External Mode)
• P0/P1 incident?        → /hotfix (reproduce → RCA → expedited → post-mortem)
• P2/P3 bug?             → /bugfix (RCA → techspec → scheduled)
• Refactor < 50 lines?   → MICRO path (no PRD/TechSpec)
• Refactor ≥ 50 lines?   → MACRO path (full RPI + compatibility matrix + ADR)
```

---

## Glossary

| Term | Definition |
|------|-----------|
| **AGENTS.md** | Project rules file read by Codex/other agents. Same content as CLAUDE.md. |
| **ADR** | Architecture Decision Record — permanent record of WHY a decision was made |
| **Chain of Thought (CoT)** | `<thinking>` blocks that force explicit reasoning before output |
| **CLAUDE.md** | Project rules file read by Claude/pi. Same content as AGENTS.md. |
| **Context Rot** | Degradation as solved problems and wrong turns accumulate |
| **Context Window** | Fixed-size box of tokens the AI can see |
| **Damage Control** | Tool-call interception: block, warn, or restrict |
| **Decision Tree** | Section in root AGENTS.md that maps task type → what to read |
| **FIC** | Frequent Intentional Compaction — summarize to file, start fresh |
| **Human-in-the-Loop** | Points where AI stops and waits for human approval |
| **Impact Check** | Post-task self-improvement protocol — flag what needs updating |
| **KNOWLEDGE_MAP.md** | Navigable project index with ADR ↔ Feature links |
| **LLM-as-Judge** | Second model/prompt evaluates first model's output |
| **MCP** | Model Context Protocol — standard for AI to access external tools |
| **Meta-Agent** | Agent that builds other agents |
| **PRD** | Product Requirements Document — WHAT/WHY |
| **Purpose Gate** | Intent declaration before work begins |
| **RAG** | Retrieval-Augmented Generation — search first, answer second |
| **Ralph Loop** | Worker + Reviewer cycle with total context reset |
| **ReAct** | Reasoning + Acting — think → tool → observe → repeat |
| **RPI** | Research → Plan → Implement |
| **RTK** | Rust Token Killer — proxy that reduces token consumption |
| **Self-Improvement Protocol** | System where agents flag knowledge updates after every task |
| **Standards** | Descriptive patterns extracted from real code (docs/standards/) |
| **Sub-Agent** | Separate AI session with own context, prompt, and tool restrictions |
| **TechSpec** | Technical Specification — HOW/architecture |
| **Tech Debt** | Planned risk reduction — separate work type with own flow |
| **Change Request** | Reactive workflow for responding to PR review feedback |
| **Code Review** | Reactive workflow for reviewing a teammate's PR (External PR Mode) |
| **Hotfix** | Expedited P0/P1 fix flow with mandatory post-mortem |
| **Post-mortem** | Blameless incident analysis artifact — 5-why, impact, prevention actions |
| **Compatibility Matrix** | Table of callers/consumers mapped to breaking-change risk (MACRO refactors) |
| **RCA** | Root Cause Analysis — 5-why drill to systemic cause, used in bugfix and hotfix flows |
| **TillDone** | Task discipline: list all tasks before using any tools |
| **Tree of Thoughts (ToT)** | Exploring multiple approaches before committing to one |
| **WORKSPACE.md** | Optional file declaring monorepo/multi-repo structure |

---

## References & Sources

### From Obsidian Vault

| Area | Notes |
|------|-------|
| ACE-FCA, FIC, RPI, Ralph Loop | `ai-agents/*.md` |
| Context Engineering | `PDF_Concepts/Context_Engineering.md`, `Context_Driven_Development.md` |
| Prompt Engineering (26 notes) | `resources/rhawk_prompt_engineering/PE-*.md` |
| Workflow Patterns | `_atoms/concepts/concept-Workflow-Patterns.md` |
| Multi-Agent | `_atoms/concepts/concept-Multi-Agent-Orchestration.md` |
| Pi Framework | `resources/pi-mono/` (13 notes) |
| Pi vs Claude Code | `resources/pi-vs-claude-code/` (15 notes) |
| Developer Toolkit AI | `resources/developer-toolkit-ai/` (130+ notes) |
| iadev Templates | `iadev-knowledge/` (PRD, TechSpec, Tasks commands) |
| TLC / Fakeflix | `fakeflix_repomix.md` (standards, skills, architecture) |
| Chain of Thought | `prompt-engineering/techniques/Chain-of-Thought Prompting.md` |
| Tree of Thoughts | `prompt-engineering/techniques/Tree of Thoughts.md` |
| Reflexion | `prompt-engineering/agents/Reflexion.md` |

### External

| Resource | URL |
|----------|-----|
| Pi Monorepo | github.com/badlogic/pi-mono |
| Developer Toolkit AI | developertoolkit.ai |
| Model Context Protocol | modelcontextprotocol.io |
| Anthropic Prompt Guide | docs.anthropic.com/en/docs/build-with-claude/prompt-engineering |

---

> **Living document.** Update via PR. If it's stale, it's [context rot](#24-context-rot).

