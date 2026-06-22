# Skills

A **skill** in LazyAI is a reusable `SKILL.md` procedure that captures a bounded workflow:
its inputs, decision flow, and expected outputs. Skills are consumed by adapters and then emitted as tool-native skill assets.

The emitted set is defined in `packages/cli/library/manifests/curation.yaml`.

Per-tool output paths (including where those emitted skills land) are described in [Tool Outputs](tool-outputs.md). The emitted set is broad, with targets covering `claude-code`, `opencode`, and `copilot`; many also target `pi`.

## Emitted skills

### Engineering and workflow practices

| Skill | Purpose | Emitted to |
|---|---|---|
| `adhd-engineer` | ADHD-optimized cognitive scaffolding for senior/staff software engineers with ~10-minute focus windows. | `claude-code`, `opencode`, `copilot`, `pi` |
| `architecture-review` | Use before structural changes to make a lightweight ADR-style decision with constraints, trade-offs, and consequences. | `claude-code`, `opencode`, `copilot`, `pi` |
| `caveman` | Use when a specification, plan, or assistant message is too verbose and needs a compact working summary without losing links to source context or replacing durable ai-memory. | `claude-code`, `opencode`, `copilot`, `pi` |
| `codebase-exploration` | Locate the real entry points, relevant files, and existing patterns before making a change. | `claude-code`, `opencode`, `copilot` |
| `create-skill` | Use when asked to create or revise a LazyAI skill source file while keeping generated tool-native skill surfaces derived, not hand-edited. | `claude-code`, `opencode`, `copilot`, `pi` |
| `create-workflow` | Use when asked to create or revise a LazyAI workflow artifact as documentation/catalog guidance without reintroducing a runtime workflow engine. | `claude-code`, `opencode`, `copilot`, `pi` |
| `diagnose` | Reproduce the failure, rank hypotheses, instrument the right seam, and add a regression test for the fix. | `claude-code`, `opencode`, `copilot` |
| `doc-backed-clarify` | Use at task intake when requirements or repository context are unclear. Supports lightweight, grill-me, and grill-me-with-docs clarification levels while always preserving the four-point pattern. | `claude-code`, `opencode`, `copilot`, `pi` |
| `fast-feedback` | Use during implementation to run the smallest meaningful verification command after each focused change. | `claude-code`, `opencode`, `copilot`, `pi` |
| `four-point-vibe-coding` | Use when starting or steering an agent task with four-point communication, project constitution, and fast feedback. | `claude-code`, `opencode`, `copilot`, `pi` |
| `handoff` | Use when a session is ending, context needs to be preserved for a future session, or when transferring work between agents. | `claude-code`, `opencode`, `copilot`, `pi` |
| `memory-promotion` | Use at task closeout to propose durable ai-memory or documentation updates without writing silently, especially when caveman summaries, diagnoses, triage, or issue extraction reveal reusable knowledge. | `claude-code`, `opencode`, `copilot`, `pi` |
| `no-workarounds` | Use during review or debugging to reject temporary patches and require root-cause fixes for workaround-shaped changes. | `claude-code`, `opencode`, `copilot`, `pi` |
| `project-guardrails-init` | Use when onboarding into a project to discover existing stack, architecture, commands, and conventions before proposing rules or memory. | `claude-code`, `opencode`, `copilot`, `pi` |
| `pr-review` | Review a change set against requirements, tests, regressions, and repository conventions. | `claude-code`, `opencode`, `copilot` |
| `create-agent` | Use when asked to create or revise a LazyAI agent source definition while keeping generated tool-native agent files derived, not hand-edited. | `claude-code`, `opencode`, `copilot`, `pi` |
| `create-hook` | Use when asked to create or revise a LazyAI hook policy so runtime enforcement stays objective, generated, and tool-specific. | `claude-code`, `opencode`, `copilot`, `pi` |
| `issue-triage` | Use when a bug report, error message, alert, or issue needs classification, deduplication, severity, ownership, refinement, and reusable triage learning before implementation. | `claude-code`, `opencode`, `copilot`, `pi` |
| `test-first-change` | Drive behavior changes through a failing test, then make the smallest code change that turns it green. | `claude-code`, `opencode`, `copilot` |
| `zoom-out` | Use when stuck in implementation details and losing sight of the bigger picture, or when a bug suggests an architectural problem rather than a local fix. Forces a structured step back to re-evaluate assumptions and design. | `claude-code`, `opencode`, `copilot`, `pi` |

### Spec-Kit workflow

| Skill | Purpose | Emitted to |
|---|---|---|
| `speckit-analyze` | Verify plan completeness via architectural & requirement traceability analysis. | `claude-code`, `opencode`, `copilot`, `pi` |
| `speckit-checklist` | Gate implementation against spec and constitution. Verify all requirements met and all gates passed. | `claude-code`, `opencode`, `copilot`, `pi` |
| `speckit-clarify` | Resolve ambiguity in spec.md through sequential, focused questioning. Records answers in the spec's Clarifications section. | `claude-code`, `opencode`, `copilot`, `pi` |
| `speckit-constitution` | Establish or amend the project constitution — the governing contract every workflow is judged against. | `claude-code`, `opencode`, `copilot`, `pi` |
| `speckit-implement` | Execute a task following its harness contract and 5-gate loop. | `claude-code`, `opencode`, `copilot`, `pi` |
| `speckit-plan` | Author an implementation plan from a validated spec. Defines approach, phases, and measurable milestones. | `claude-code`, `opencode`, `copilot`, `pi` |
| `speckit-specify` | Author a feature specification from a user description. Output is the contract every downstream artifact is judged against. | `claude-code`, `opencode`, `copilot`, `pi` |
| `speckit-tasks` | Break a plan into executable task files with dependency graph and parallelization. | `claude-code`, `opencode`, `copilot`, `pi` |
| `tdd-planning` | Use during research or planning before implementation to choose a TDD mode, define red tests, preserve existing tests, and produce the test-first artifact for a feature, bugfix, or refactor. | `claude-code`, `opencode`, `copilot`, `pi` |
| `task-to-issues` | Use when extracting actionable tasks from meeting notes, Slack threads, PR descriptions, specs, or other unstructured text. Converts ephemeral notes into tracked issues with context, acceptance criteria, deduplication, and learning capture. | `claude-code`, `opencode`, `copilot`, `pi` |

### Memory, agents, and governance helpers

| Skill | Purpose | Emitted to |
|---|---|---|
| `skill-authoring` | Use when creating, renaming, or modifying LazyAI skill source files so the canonical source and generated tool-native outputs stay aligned. | `claude-code`, `opencode`, `copilot`, `pi` |

## Library & spec-workflow skills


The following skills are retained in the LazyAI library only. They are used to drive the `.ai/` and specs workflow, and are **not copied into tool-specific skill directories**.

| Skill | Purpose |
|---|---|
| `anti-speculation` | Prevent scope creep and speculative implementation. |
| `bugfix` | Execute a bugfix workflow — reproduce, root-cause, fix, verify regression. |
| `chain-verify` | Read-only verification trace across spec, plan, tasks, implementation, and test evidence. |
| `extract-standards` | Derive standards from code patterns. Workspace-aware, prevents duplication across scopes. |
| `housekeeping` | Execute planned tech-debt, dependency, or code cleanup as a bounded task. |
| `impact-check` | Impact-check automation — identify and propose updates to CLAUDE.md, standards, and constitution after work completes. |
| `implement` | Implement requested changes safely with test-first workflow. |
| `improve-codebase-architecture` | Surface architectural friction and propose deepening opportunities — refactors that turn shallow modules into deep ones. Uses Ousterhout deep-module principle. Trigger when user wants to improve architecture, find refactoring opportunities, or make a codebase more testable and AI-navigable. |
| `iterate` | Iterate on implementation based on feedback or new requirements. |
| `memory-write` | Write persistent context and decisions for future sessions. |
| `parallel-execution` | Execute independent sub-tasks concurrently. |
| `plan` | Plan implementation approach before writing code. |
| `process-audit` | Audit workflow adherence — verify RPI phases, gates, constitution compliance. |
| `proof-of-concept` | Validate feasibility via a minimal, time-boxed PoC. Discard after learning is captured. |
| `red-team-plan` | Read-only adversarial design review of plan/spec artifacts before implementation approval. |
| `research` | Research codebase patterns, dependencies, and conventions. |
| `review` | Conduct rigorous code review with Constitutional alignment + Article VI audit. |
| `rpi` | Orchestrate full Spec-Driven Development workflow — chain all phases with human gates. |
| `self-improve` | Analyze human interventions to identify process failures and improve library instructions to reduce future corrections. |
| `spike` | Execute a time-boxed research spike to answer a single sharp question. |
| `tdd-loop` | Execute test-driven development loop: red, green, refactor. |
| `update-memory` | Capture and update persistent knowledge, entity relationships, and lessons. |

All listed source files were found in the canonical skill directories under `packages/cli/library/`; none were omitted.