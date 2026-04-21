# Research: Housekeeping, Memory, and Bootstrap Workflows

**Date:** 2026-04-17
**Status:** Research Phase
**Related Specs:** 001, 002, 003, 004, 005
**Core Documents:** `docs/orchestrator-design.md`, `docs/orchestration-usage.md`, `docs/orchestrator-blueprint.md`

## Problem Statement
The `ai-setup` ecosystem is transitioning from a simple scaffolding tool to a runtime coordination system via the `@ai-setup/orchestrator`. However, there is a gap between the runtime execution of agent chains and the "lifestyle" of the project: how context is bootstrapped at startup, how memory is maintained across sessions, and how the technical index (Codegraph/qmd) is kept in sync with the cognitive index (specs/memory). Without a formal housekeeping and memory workflow, the "durable state" goal of the orchestrator is limited to a single chain's lifecycle rather than the project's entire evolution.

## Goals and Non-Goals
### Goals
- Define a deterministic bootstrap process for the Orchestrator MCP when a user begins work.
- Establish a "housekeeping" ritual (pre- and post-task) to ensure project hygiene.
- Create a rigorous memory workflow for reading/updating the project's knowledge base.
- Integrate technical indexing tools (`qmd`, `codegraph`) with approval-gated human checkpoints.
- Clean up the proliferation of `AGENTS.md` files to a canonical structure.
- Define the roadmap for a persistent store for tasks and configurations.

### Non-Goals
- Implement the code for the MCP server or CLI tools.
- Define the internal logic of `qmd` or `codegraph`.
- Replace the host CLI tool's primary search engine (treat them as complementary).

## Current State Inventory

| Component | State | Evidence |
| :--- | :--- | :--- |
| **Orchestrator MCP Server** | Shipped (Phase 2) | `orchestrator/src/server.ts` (9 tools: `start_chain`, `advance_chain`, etc.) |
| **Orchestrator Agent** | Shipped | `.opencode/agents/orchestrator.md`, etc. |
| **RPI Workflow** | Shipped | `specs/rules/workflow.md` |
| **Memory Storage** | Partial | `specs/memory/` exists as a pattern, but no formal runtime workflow. |
| **Codegraph/qmd** | External, available locally, not integrated | Present as optional tooling and installable MCP surfaces, but not yet part of orchestrator-managed lifecycle/state. |
| **AGENTS.md** | Bloated | Found in nearly every subdirectory of `specs/`. |
| **Obsidian (`ob`)** | Installed locally, not integrated | Useful for vault discovery/configuration, but not currently wired into ai-setup defaults or orchestrator workflows. |

## Repo Evidence Summary

### 1. Orchestration Runtime
The current orchestrator is a "thin" layer. It manages state machines (chains/workflows) but does not have a "Project Lifecycle" concept. It starts a chain, advances it, and completes it. It lacks a "session start" or "session end" hook that triggers housekeeping.

### 2. Agent Definitions
The project is currently suffering from `AGENTS.md` sprawl. The intention was likely to provide context for agents in specific domains, but it has resulted in redundant files across `specs/features/`, `specs/bugfixes/`, etc., contradicting the "root-level" guidance principle.

### 3. Durable State
`docs/orchestrator-design.md` mentions "durable state" and "lessons" (Section 25), but these are currently treated as artifacts created at the end of a chain, not as a living memory graph that is checked *before* starting a new chain.

## External/Tooling Evidence Summary

- **qmd**: Preferred retrieval/index engine for markdown corpora. High-performance indexing of specs/docs.
- **Codegraph**: Provides structural code context.
- **`ob` (Obsidian CLI)**: Useful for vault discovery and high-level configuration, but not as a primary search engine for the LLM.
- **Second-Brain Principles**: Evidence suggests a "Fresh context per wave" approach, where the orchestrator explicitly loads only the necessary "memories" to avoid token bloat and hallucination.

## Requirements and Tensions

### The Sync vs. Approval Tension
**Conflict:** There is a requirement that `codegraph` and `qmd` stay synced (meaning indexing must happen when files change). Simultaneously, there is a requirement that reads, writes, and maintenance actions should require explicit user approval.

**Policy Options:**
1. **Strict Gating:** Every index update or memory load is a proposed task. *Result: High friction, user becomes an "indexing bot".*
2. **Implicit Sync / Explicit Report:** Indexing happens automatically on file write. *Result: Conflicts with the requirement for non-silent mutating actions.*
3. **Bounded Standing Approval (Contract):** The user grants a session-scoped or task-scoped "maintenance contract" (e.g., "Keep indexes synced for this task"). Under this contract, specific maintenance actions are pre-approved. *Result: Balanced friction and control.*

**Recommended Conclusion:** Adopt **Option 3**. All memory/context loading, indexing updates, cleanup, and housekeeping are approval-gated by default. The preferred model is a bounded standing approval granted by the user for a session or a specific task (for example: "For this task, you may load memory context and keep qmd/codegraph synced"). Drift and staleness checks may occur without mutation, but actual memory loads and repair/sync actions should only happen after explicit approval or within that standing contract. This preserves approval-first behavior while ensuring the system does not knowingly proceed on stale indexes when sync is required.

## Proposed Workflows

### 1. Bootstrap (Session Start)
When a user initiates a session:
- **Discovery:** Identify the memory path (default `specs/memory`, but user-configurable).
- **State/Drift Inspection:** Check whether qmd/codegraph indexes appear stale and whether relevant memory/spec sources exist.
- **Approval Request:** Unless the user has already granted a bounded standing approval for this session/task, ask before loading memory/search context and before performing any sync/index repair.
- **Context Loading:** After approval (or under standing approval), use `qmd` to find relevant specs/memories based on the initial task description and use `codegraph` for code-context preparation when relevant.
- **Bootstrap Report:** Summarize what was loaded, what remained stale, and what maintenance actions were performed or deferred.

### 2. Housekeeping (Pre- and Post-Task)
- **Pre-Task:**
    - Verify the "Contract" for the next step.
    - Check whether newer memories/specs may change the approach.
    - Request approval before any memory/context load unless covered by standing approval.
- **Post-Task:**
    - **Memory Extraction:** Identify new "lessons" or "decisions" from the task output.
    - **Cleanup Proposal:** Identify temporary artifacts, git-hygiene opportunities, and organization work.
    - **Sync Proposal:** Identify required `qmd`/`codegraph` updates and whether indexes are stale.
    - **Approval:** Perform writes/cleanup/sync only after explicit approval or within standing approval, then report what changed.

### 3. Memory Lifecycle
- **Read:** `Task Request` $\rightarrow$ `Memory Path Resolution` $\rightarrow$ `Approval / Standing Approval Check` $\rightarrow$ `qmd Search` $\rightarrow$ `Filtered Memory Load` $\rightarrow$ `Context Window`.
- **Write:** `Step Completion` $\rightarrow$ `Lesson Extraction` $\rightarrow$ `Proposed Memory Update/File` $\rightarrow$ `User Approval / Standing Approval Check` $\rightarrow$ `Write to chosen memory path`. 

## Roadmap and Strategy

### Database/Store Evolution
Move from flat-file memory to a structured store for:
- **Task Registry:** Tracking all historical chains, their budgets, and outcomes.
- **Config Store:** Centralized agents/skills/commands/hooks (removing the need for fragmented `AGENTS.md` files).

### `AGENTS.md` Migration
- **Action:** Delete all `AGENTS.md` files except those in the root or designated global config areas (e.g., `.opencode/agents/`).
- **Replacement:** Use the Orchestrator's `compose_agent` tool to dynamically inject role-specific instructions from the library instead of relying on local file presence.

### Install-Time UX
- **Optionality:** The `ai-setup init` wizard should offer optional integration for:
    - `Obsidian` discovery/configuration (via `ob`)
    - `qmd` (preferred markdown retrieval/index layer when enabled)
    - `Codegraph` (for structural analysis and code-context preparation)
- Each option should make its approval model clear: read-only discovery, drift checks, and mutating sync/index updates must be explicitly described.
- The wizard should also let the user choose the canonical memory path (default `specs/memory`).

## Risks and Constraints
- **Risk (High):** The "Approval Gate" could become a bottleneck if the orchestrator asks for approval for every minor index update.
- **Risk (Medium):** Headless `ob` may have environment-specific pathing issues across different OSs.
- **Constraint:** Must remain "least-privilege"; the orchestrator cannot modify files without a tool call, which must be visible to the user.

## Open Questions for Planning
1. Should the "Memory Extraction" phase be a separate step in every RPI chain, or a global "Post-Implementation" hook?
2. How do we handle "Conflicting Memories" (e.g., an old spec says X, a new lesson says Y)?
3. What is the minimal set of metadata required in the "Task Store" to reconstruct a session for a future agent?
