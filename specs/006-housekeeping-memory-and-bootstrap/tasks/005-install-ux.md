# Spec 006 — Task 005: Install-Time UX for Optional Tooling

> Scope: Specify install-time UX for optional Obsidian (`ob`), qmd, and codegraph integrations. This task is documentation-only and preserves approval-first behavior for config writes, sync-state writes, and index mutations.

---

## Objective

Define how `ai-setup init` and related setup surfaces present optional tooling choices so users can enable richer retrieval and code context without blurring read-only discovery, approval-gated mutations, or runtime responsibilities.

---

## Normative Inputs and Defaults

- Research: `specs/006-housekeeping-memory-and-bootstrap/research.md`
- Plan: `specs/006-housekeeping-memory-and-bootstrap/plan.md`
- Data model: `specs/006-housekeeping-memory-and-bootstrap/data-model.md`
- Metadata migration checklist: `specs/006-housekeeping-memory-and-bootstrap/tasks/001-metadata-migration.md`
- Bootstrap workflow: `specs/006-housekeeping-memory-and-bootstrap/tasks/002-bootstrap-workflow.md`
- Housekeeping workflow: `specs/006-housekeeping-memory-and-bootstrap/tasks/003-housekeeping-workflow.md`
- Sync state path: `.ai/housekeeping/sync-state.json`
- Metadata enforcement: strict for new artifacts; warn/migrate for legacy artifacts
- Approval model default: task-scoped approval; optional session-scoped; standing approvals hard-expire at 30 days while allowing an in-flight task to finish
- qmd index path: project-local by default
- Default memory path: `specs/memory`

---

## Install-Time UX Principles

The install experience for optional tooling should follow these rules:

1. **Optional by default**
   - `ob`, qmd, and codegraph are opt-in integrations.
   - A project may enable none, one, two, or all three.

2. **Read-only discovery first**
   - Detection of installed tools, vault paths, candidate config, and drift status is read-only.
   - Discovery never creates indexes, writes config, or updates sync-state by itself.

3. **Approval-first mutation model**
   - Config writes, sync-state writes, index builds/rebuilds, and repair actions remain explicit mutations.
   - Setup UX must describe these as follow-up actions that require user approval or later contract coverage.

4. **Prefer smallest useful scope**
   - The wizard should recommend task-scoped approval as the default maintenance contract choice.
   - Session-scoped approval is optional.
   - Persistent standing approval is an explicit advanced choice that hard-expires at 30 days.

5. **Project-local defaults first**
   - Memory defaults to `specs/memory`.
   - qmd index location is project-local by default.
   - sync-state remains project-local at `.ai/housekeeping/sync-state.json`.

---

## Ownership Boundaries for Phase 6

| Layer | Owns | Does Not Own |
|---|---|---|
| `ai-setup` CLI / wizard | install prompts, opt-in choices, config persistence prompts, explanation of approval model, initial path capture | runtime bootstrap execution, housekeeping execution, silent maintenance writes |
| Orchestrator runtime | bootstrap and housekeeping behavior after install, contract evaluation, approved context loading, approved sync execution | wizard UI, tool installation detection UX copy |
| Spec conventions | defaults, allowed behaviors, approval contract semantics, path conventions, artifact boundaries | implementation details of CLI screens or runtime adapters |

This task defines the contract between those layers; it does not implement any layer.

---

## 1. Wizard Sequence Overview

Phase 6 assumes `ai-setup init` offers optional integration prompts after core project setup is known but before final config is written.

Recommended order:

1. choose or confirm the project memory path
2. offer Obsidian (`ob`) discovery/configuration
3. offer qmd markdown retrieval/index support
4. offer codegraph structural analysis support
5. summarize the selected integrations and clearly label which actions are:
   - read-only discovery
   - config writes
   - future runtime reads
   - future approval-gated maintenance mutations

The wizard should not treat tool enablement as blanket approval for future sync/index writes.

---

## 2. Memory Path Configuration UX

### 2.1 Prompt

`ai-setup init` should ask:

> Where should project memory be stored?

### 2.2 Default Behavior

- Default answer: `specs/memory`
- The path is project-local and remains the canonical default for bootstrap and housekeeping.
- If the user does not override it, downstream workflows should resolve memory from `specs/memory` first.

### 2.3 Persistence Behavior

- Persist the chosen memory path to project config during install.
- The persistence location may be the existing project config surface already used by `ai-setup`.
- This task does **not** define a new config API or new `compose_agent` fields.

### 2.4 Integration Effects

- If qmd is enabled, the configured memory path is included in qmd retrieval/index scope.
- If codegraph is enabled, the configured project memory path is added to the declared scan scope defined by setup config.
- If neither tool is enabled, bootstrap and housekeeping still use the memory path for direct file reads when approved.

### 2.5 Approval Notes

- Reading the currently configured path is read-only.
- Writing the chosen path to config during setup is a user-approved setup mutation.
- Later changes to the path after install are separate config mutations and must remain explicit.

---

## 3. Obsidian (`ob`) Integration UX

### 3.1 Purpose

Obsidian support exists for vault discovery and configuration awareness. It is not the primary retrieval engine for agent context.

### 3.2 Wizard Prompt

`ai-setup init` should offer a prompt equivalent to:

> Enable Obsidian integration for vault discovery and configuration awareness?

### 3.3 Read-Only Discovery Actions

The following Obsidian-related actions are read-only and may run without a maintenance contract:

- detect whether `ob` is installed and reachable
- inspect candidate vault paths
- inspect vault-related configuration that can be read without mutation
- report whether a vault appears available for optional linking

### 3.4 Mutating Actions

The following remain explicit write actions:

- persisting vault-path or Obsidian-related config into project setup files
- writing any new integration settings during install/update
- future repair or migration of Obsidian-related config

### 3.5 UX Requirements

- The wizard should explain that enabling Obsidian does **not** make the vault the canonical memory path unless the user explicitly chooses that path.
- The wizard should explain that Obsidian discovery is advisory and non-blocking.
- If `ob` is unavailable, the wizard should allow the user to continue without enabling it.

---

## 4. qmd Integration UX

### 4.1 Purpose

qmd is the preferred retrieval and indexing layer for markdown corpora when enabled.

### 4.2 Wizard Prompt

`ai-setup init` should offer a prompt equivalent to:

> Enable qmd for markdown retrieval?

### 4.3 Default Behavior

- qmd remains optional.
- qmd index storage is project-local by default.
- The configured memory path should participate in qmd retrieval resolution when qmd is enabled.

### 4.4 Read-Only Actions

The following qmd-related actions are always read-only and should be described that way in setup UX:

- detect whether qmd is installed and reachable
- inspect whether a project-local qmd index already exists
- inspect qmd freshness or drift using mtimes, hashes, timestamps, or tool state
- execute qmd searches against an existing index during later runtime operations

### 4.5 Mutating Actions

The following qmd-related actions are explicit maintenance mutations:

- creating the qmd index
- rebuilding or repairing the qmd index
- updating `.ai/housekeeping/sync-state.json` with qmd drift, acknowledgment, or repair results
- changing qmd-related project config after initial setup

### 4.6 Approval Model Description

The wizard should clearly state:

- enabling qmd allows future read-only retrieval when an index is present
- enabling qmd does **not** authorize silent index builds or rebuilds
- qmd sync/re-index later requires per-action approval or active contract coverage
- task-scoped approval is the default recommended contract when users want the system to keep qmd in sync during a task

### 4.7 Failure and Deferral Behavior

- If qmd is enabled but no index exists yet, setup may record the preference without forcing an immediate index build.
- If the user declines an initial index build, later bootstrap/housekeeping may report qmd as unavailable, unknown, or stale depending on actual state.
- Rejection of qmd sync later should follow `staleAcked` handling in `.ai/housekeeping/sync-state.json` when that write is approved or contract-covered.

---

## 5. Codegraph Integration UX

### 5.1 Purpose

codegraph provides structural code analysis and code-context preparation when enabled.

### 5.2 Wizard Prompt

`ai-setup init` should offer a prompt equivalent to:

> Enable codegraph for structural analysis?

### 5.3 Default Behavior

- codegraph remains optional.
- codegraph data stays project-local by default unless a later spec says otherwise.
- When enabled, bootstrap may use codegraph for code-relevant tasks after approval or contract evaluation.

### 5.4 Read-Only Actions

The following codegraph-related actions are read-only:

- detect whether codegraph is installed and reachable
- inspect whether project-local codegraph state already exists
- inspect codegraph freshness/drift against source state
- run read-only codegraph queries against existing indexed data

### 5.5 Mutating Actions

The following codegraph-related actions require approval or active contract coverage:

- initial codegraph index build
- codegraph rebuild or repair
- sync-state writeback for codegraph drift, `staleAcked`, or repair outcomes
- changes to persisted codegraph config after initial setup

### 5.6 Approval Model Description

The wizard should clearly state:

- codegraph queries are read-only when index data already exists
- codegraph indexing is a mutating maintenance action
- enabling codegraph does **not** grant silent ongoing indexing permission
- task-scoped approval remains the default recommendation for future sync behavior

### 5.7 Runtime Relationship

- codegraph is a runtime context source, not a wizard-owned execution engine.
- The wizard only captures preference and config intent.
- Bootstrap and housekeeping own later drift checks, approval evaluation, and approved repair proposals.

---

## 6. Discovery vs Mutation Matrix

| Action | Layer that initiates it | Read-only? | Approval / contract needed? |
|---|---|---:|---:|
| Detect `ob`, qmd, or codegraph availability | CLI / wizard or runtime | Yes | No |
| Discover candidate Obsidian vault path | CLI / wizard | Yes | No |
| Read existing project config for memory/tool settings | CLI / wizard or runtime | Yes | No |
| Inspect qmd/codegraph drift or freshness | Runtime | Yes | No |
| Run qmd search on an existing index | Runtime | Yes | No |
| Run codegraph query on existing indexed data | Runtime | Yes | No |
| Persist memory path during setup | CLI / wizard | No | Yes, as explicit setup write |
| Persist Obsidian/qmd/codegraph enablement config | CLI / wizard | No | Yes, as explicit setup write |
| Create or rebuild qmd index | Runtime or approved setup action | No | Yes |
| Create or rebuild codegraph index | Runtime or approved setup action | No | Yes |
| Write `.ai/housekeeping/sync-state.json` | Runtime | No | Yes |

This matrix is normative for approval-first behavior.

---

## 7. Contract and Approval UX Copy Requirements

When the wizard explains future maintenance behavior, it should preserve the contract distinctions defined in the data model:

- **Per-action approval**: one explicit mutation at a time
- **Task-scoped approval**: default recommended maintenance scope for one task/workflow run
- **Session-scoped approval**: optional broader approval for the active session
- **Standing approval**: explicit advanced option with hard expiry at 30 days; an in-flight task may finish, but the next task needs renewal

The wizard should not imply that install-time enablement automatically creates a standing contract.

---

## 8. Reporting Expectations After Install

Install-time summary output should clearly separate:

- **Enabled integrations**
- **Configured project-local paths**
- **Writes performed during setup**
- **Deferred actions not yet performed**, such as:
  - initial qmd index build not yet approved
  - initial codegraph index build not yet approved
  - future sync-state creation pending first maintenance write

This keeps setup honest about what is configured versus what is actually indexed or synchronized.

---

## 9. Out of Scope for Phase 6

This task does **not** define:

- wizard implementation details or screen layouts
- runtime bootstrap or housekeeping code
- a new `compose_agent` API surface
- silent background indexing behavior
- AGENTS.md cleanup mechanics from Phase 5
- quickstart verification/runbook work from Phase 7

---

## Acceptance Mapping

| Plan / AC Target | Coverage in this task |
|---|---|
| Phase 6 Task 6-1 | Obsidian install UX, discovery, and approval distinctions captured in Section 3 |
| Phase 6 Task 6-2 | qmd install UX, project-local indexing behavior, and approval model captured in Section 4 |
| Phase 6 Task 6-3 | codegraph install UX and runtime boundary captured in Section 5 |
| Phase 6 Task 6-4 | memory path configuration UX and default `specs/memory` behavior captured in Section 2 |
| AC-7 | Optional tooling install-time UX and approval-model descriptions specified |
| AC-9 | CLI/wizard vs runtime vs spec-convention ownership boundaries specified |

---

## Validation Notes

- Consistent with research guidance that `ob` is for discovery/config awareness, qmd is the preferred markdown retrieval layer, and codegraph is for structural code context.
- Consistent with the data model's approval-first rules: read-only drift/discovery allowed; sync/index/config writes require approval or contract coverage.
- Consistent with approved defaults: `.ai/housekeeping/sync-state.json`, strict-new/warn-legacy metadata handling, default task-scoped approval, 30-day standing approval expiry, project-local qmd behavior, and default memory path `specs/memory`.
- No implementation code or wizard/runtime behavior was added in this task.
