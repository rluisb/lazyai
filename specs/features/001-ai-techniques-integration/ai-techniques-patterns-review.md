# AI Techniques & Patterns Review — Current State

**Scope:** Agentic coding, harness engineering, prompt engineering, context engineering, and related AI workflow practices in this repository.  
**Mode:** Research and report only.  
**Date:** 2026-05-01

---

## Summary

This repository has a strong AI workflow foundation: multi-agent orchestration, role-specialized agents, chain/team workflows, context discipline, quality gates, anti-speculation controls, and memory protocols are all present and documented.

The largest gaps are not in basic agentic workflow design, but in **automation and enforcement**:

1. The root project contract still contains placeholders like `[YOUR_PROJECT_OVERVIEW]`, `[YOUR_LANGUAGE]`, and `[YOUR_TEST_FRAMEWORK]`.
2. `specs/standards/` is effectively unpopulated, which weakens pattern-consistency review.
3. Retrieval systems exist as available tools/skills, but there is no automatic RAG-style context retrieval pipeline wired into agent startup.
4. Cost, observability, compaction, and agent progression are documented but mostly advisory.

---

## MUST HAVE — Implemented & Active

These techniques exist as documented patterns, skills, rules, agents, chains, workflows, or operational protocols.

| # | Technique | Evidence | Category |
|---|-----------|----------|----------|
| 1 | Multi-Agent Orchestration | `orchestrator.md`, `agents/AGENTS.md`, orchestration chains, teams, workflows | Agentic Coding |
| 2 | Role-Specialized Agents | Scout, planner, builder, implementor, reviewer, red-team, documenter, orchestrator | Agentic Coding |
| 3 | Sequential Chain Execution | `feature`, `bugfix`, `review`, `refactor`, `tdd`, `onboard`, `new-package` chains | Agentic Coding |
| 4 | Parallel Team Execution | `assessment-team`, `feature-team`, `review-team` | Agentic Coding |
| 5 | Dual-Agent Verification | Harness Protocol Rule 2: generator must not approve its own artifact | Harness |
| 6 | 5-Gate Quality Ladder | `quality-gates.xml`: static integrity, contract compliance, behavioral validation, pattern consistency, observability readiness | Harness |
| 7 | Feed Forward Blueprint Chain | `constitution.md → spec.md → plan.md → tasks.md → task-harness.md → code` | Harness |
| 8 | Memory & State Persistence | Ledgers, last-known-state files, handoff protocol, memory-write/update-memory skills | Harness |
| 9 | Anti-Slope Protocol | Regression tests, PoC discard, standards-as-memory, versioned constitution amendments | Harness |
| 10 | Composed Prompt Layering | Orchestrator composition: base agent + domain skill + mode skill | Prompt Engineering |
| 11 | LLM-as-Judge Verification | Review skill/workflow with multi-lens review and synthesis | Prompt Engineering |
| 12 | Contract-Based Prompting | Generator names contract; verifier evaluates against contract | Prompt Engineering |
| 13 | Skill Frontmatter Declarations | Skills declare harness/workspace/output-schema metadata | Prompt Engineering |
| 14 | Progressive Context Loading | Root `AGENTS.md` decision tree for what to load per task type | Context Engineering |
| 15 | Token Discipline Protocol | Targeted reads, summaries, avoid speculative loading, context budget guidance | Context Engineering |
| 16 | Context Budget Gates | 70% normal operation, 85% warning, 95% mandatory compaction | Context Engineering |
| 17 | One Task = One Session | Explicit session hygiene rule | Context Engineering |
| 18 | Self-Consistency Verification | Complexity-scaled verification rounds | Verification |
| 19 | Confidence Gate | High/medium/low confidence behavior rules | Verification |
| 20 | ReAct Trace Protocol | Thought → Action → Observation → Decision format | Reasoning |
| 21 | Tree-of-Thoughts Decision Protocol | Alternatives evaluated by complexity, consistency, reversibility, performance, familiarity | Reasoning |
| 22 | Recovery Patterns | Retry, fix-resume, escalate, handoff | Recovery |
| 23 | Agent Security Rules | Prompt injection, privilege escalation, secret exposure, config tampering, context poisoning mitigations | Security |
| 24 | Anti-Speculation Guards | Purpose Gate and halt protocol integration | Quality |
| 25 | Self-Improvement Protocol | Impact check, knowledge update checklist, session handoffs | Knowledge Management |

---

## NICE TO HAVE — Defined but Lightly Implemented

These techniques are present in documentation or concepts, but enforcement, automation, or completeness is limited.

| # | Technique | Current State | Gap |
|---|-----------|---------------|-----|
| 1 | Few-Shot Prompting | Mentioned in reasoning protocol; local examples exist | No systematic few-shot catalog or retrieval mechanism |
| 2 | Generated Knowledge | Listed as a reasoning technique | No dedicated skill or structured process |
| 3 | Observability Readiness | Gate 5 defined in quality gates | No automated enforcement or CI integration |
| 4 | Coverage Thresholds | 85% client / 90% server listed in quality gates | Root AGENTS.md still contains `[YOUR_COVERAGE_THRESHOLD]%` placeholder |
| 5 | API Cost Monitoring | Cost rule documents tracking and alerts | No actual API/model spend integration |
| 6 | Cross-Repo Workflows | Workspace protocol defines the pattern | No strong evidence of active cross-repo workflow runs |
| 7 | Compaction Automation | Compaction protocol defined | Advisory only; no automated compaction enforcement |
| 8 | Constitution Population | Constitution template and protocol exist | Current root AGENTS.md has unresolved placeholders |
| 9 | Knowledge Graph Integration | Graphify, codegraph, qmd, memoria, knowledge-stack tools/skills available | Not automatically updated or injected into every workflow |
| 10 | Agent Progression Levels | L1-L4 supervision model documented | No enforcement or tracking mechanism |
| 11 | Standards-as-Code | Standards directory and rules exist | `specs/standards/` is mostly empty; few concrete pattern standards |
| 12 | TillDone Protocol | Defined in workflow rule | No automated detection if agents stop early |
| 13 | Multi-CLI Support | Claude Code, Copilot, Gemini, Codex, OpenCode plans/research exist | OpenCode appears primary; others are partial/WIP |

---

## DO NOT HAVE — Missing or Not Evident

These are established AI/agent techniques that are not meaningfully implemented in the current state.

| # | Technique | Why It Matters | Current Gap |
|---|-----------|----------------|-------------|
| 1 | Retrieval-Augmented Generation (RAG) | Automatically injects relevant project context | Tools exist, but no automatic retrieval pipeline is wired into agent prompts |
| 2 | Automated Model Selection/Routing | Reduces cost and improves fit by task type | Cost rules are advisory; routing is manual |
| 3 | Multi-Agent Debate/Consensus | Resolves disagreement through structured adversarial consensus | Reviewer/red-team exist, but no formal debate protocol |
| 4 | Chain-of-Verification | Validates consistency across every workflow artifact | Per-artifact checks exist, but no chain-wide verifier |
| 5 | Automated Agentic Error Recovery | Enables safe autonomous recovery from known failures | Recovery usually waits for human confirmation |
| 6 | Execution Plan Validation | Checks plan completeness, feasibility, and consistency before implementation | Plan approval is mostly human/manual |
| 7 | Guardrails/Railings System | Enforces output/input safety constraints in code | Security rules are documented, not enforced by a guardrails runtime |
| 8 | Explicit Agent State Machine | Improves observability and recovery | Chain state exists; individual agent lifecycle is implicit |
| 9 | Structured Human-in-the-Loop Feedback | Captures more than approve/reject decisions | Human gates exist, but feedback loop is not structured |
| 10 | Dynamic Prompt Optimization | Improves prompts based on failures and success metrics | Prompts are static files |
| 11 | Structured Evaluation Benchmark | Measures agent quality across sessions | No benchmark/eval suite for agent performance |
| 12 | Tool Creation by Agents | Lets agents create bounded tools for repetitive tasks | Agents only use existing tools |
| 13 | Formal Causal Reasoning Framework | Improves bug diagnosis rigor | Bugfix chain has RCA-like phases, but no formal 5 Whys/fault-tree protocol |
| 14 | Environment-Aware Planning | Incorporates token budget, CI latency, model cost, and operational limits into plans | Budget tracking is separate from planning |
| 15 | Multi-Model Ensemble | Improves high-stakes judgment via model diversity | No ensemble pattern beyond role-specialized agents |
| 16 | Streaming/Progressive Output | Improves visibility on long-running tasks | Outputs are batch-style |
| 17 | Adversarial Self-Play During Design | Finds design flaws before implementation | Red-team is mainly post-implementation/review-oriented |
| 18 | Continuous Learning from Human Corrections | Turns repeated corrections into prompt/rule improvements automatically | Self-improve exists, but extraction across sessions is not automated |

---

## Category Dashboard

| Category | Strongly Implemented | Lightly Implemented | Missing / Not Evident |
|----------|---------------------|---------------------|-----------------------|
| Agentic Coding | High | Medium | Medium |
| Harness Engineering | High | Medium | Low-Medium |
| Prompt Engineering | High | Medium | Medium |
| Context Engineering | High | Medium | Medium |
| Verification | Medium-High | Low-Medium | High |
| Reasoning | Medium-High | Low-Medium | Medium |
| Recovery | Medium | Low | Medium |
| Security | Medium-High | Low | Medium |
| Knowledge Management | Medium-High | Medium | Medium |

---

## Top Recommendations

### 1. Fill in the project constitution and root contract

The root `AGENTS.md` still contains placeholders. This weakens the entire harness because agents cannot enforce project-specific rules that have not been stated.

Priority fields to fill:

- Project overview
- Stack
- Codebase map
- Naming conventions
- Error handling conventions
- API response conventions
- Import conventions
- Test command
- Lint command
- Build command
- Coverage threshold
- Protected branch

### 2. Populate `specs/standards/`

Gate 4 depends on concrete standards. Right now the repository has strong rules, but few actual project-specific standards.

Recommended initial standards:

- `specs/standards/architecture/orchestration-patterns.md`
- `specs/standards/testing/setup-engine-test-patterns.md`
- `specs/standards/coding/error-handling.md`
- `specs/standards/security/agent-security-patterns.md`
- `specs/standards/context/context-loading-patterns.md`

### 3. Wire retrieval into agent startup

The repository has CodeGraph, QMD, memory, and knowledge-stack capabilities, but they are not automatically injected into task context.

A useful next step would be a lightweight retrieval preflight:

1. Determine task type.
2. Query relevant standards/rules/specs.
3. Query code symbols or files through CodeGraph.
4. Summarize top findings into a small context packet.
5. Pass only that packet to the next agent.

### 4. Add plan validation before implementation

The planner produces plans, but there is no automated plan verifier.

A plan validator should check:

- Every acceptance criterion has a task.
- Every task has verification.
- Every task has declared files/scope.
- No task includes speculative behavior.
- Risks and rollback path are documented.
- Human approval gate is explicit.

### 5. Track agent/process outcomes

To move from documentation to learning system, capture outcomes:

- Which agent ran?
- Which model/mode?
- Which skills?
- Did review find issues?
- Did tests fail?
- Was human correction needed?
- Which prompt/rule/standard should change?

This would enable structured evaluation, cost analysis, and prompt/process improvement over time.

---

## Bottom Line

The repository is already implementing many modern AI engineering patterns, especially around:

- Multi-agent decomposition
- Harness-based workflow control
- Prompt chaining
- Context discipline
- Human approval gates
- Review separation
- Anti-speculation controls
- Memory and handoff protocols

The biggest opportunity is to convert documented protocols into enforced, observable systems: populate the project contract and standards, automate retrieval, validate plans, and track agent outcomes.
