# Handoff: MCP Teams and Workflows Implementation

**Date:** 2026-04-29
**Branch:** `feat/mcp-teams-workflows`

## 1. Current Objective and Status
- **Objective:** Expose the existing Agent Swarms (Teams) and Workflows functionality through the `@ai-setup/orchestrator` MCP server, making generic tools like `get_status` and `get_budget` polymorphic.
- **Status:** **Done**. The implementation is complete, thoroughly tested (all tests passed), and committed to the `feat/mcp-teams-workflows` branch.

## 2. Decisions Made
- **Polymorphism via Schema Updates:** Instead of creating separate MCP tools for teams/workflows status and budgets, we replaced the hardcoded `CHAIN_KIND_SCHEMA` with `RUN_KIND_SCHEMA` (`'chain' | 'team' | 'workflow'`). This maintains a clean and unified API surface for tracking any run type.
- **Explicit MCP Tool Registrations:** Added explicit registrations for `build_team`, `assign_team_task`, `complete_team_task`, `start_workflow`, and `advance_workflow` in `src/server.ts` to finally allow these operations to be triggered from host CLIs (e.g. Claude Code).
- **Handler Updates:** Extended `retryStep`, `escalateStep`, and `handoff` in `src/tool-handlers.ts` to fully support teams and workflows. We opted for a thin normalization layer in `server.ts` to keep the orchestration semantics securely in the handlers/machines.

## 3. Open Assumptions/Questions
- Will host CLIs like Claude Code automatically pick up the new MCP tools without additional configuration updates, or do we need to trigger a catalog refresh?
- Should we define a more concrete `user_approval` protocol specifically for Swarms (Teams) given their `budget_multiplier`, or is the current mechanism (relying on `build_team` to surface the budget) sufficient?

## 4. Next Concrete Actions
1. **Create Pull Request:** Push the `feat/mcp-teams-workflows` branch to remote and open a Pull Request against `main`.
2. **Knowledge Base Synchronization:** Update `specs/KNOWLEDGE_MAP.md` and possibly `README.md` to reflect that the orchestrator now officially supports multi-agent swarms via the MCP protocol.

## 5. Risks/Watchouts for the Next Agent
- **Unrelated Working Tree Changes:** The working tree still has uncommitted modifications to `../../CLAUDE.md`, `docs/`, and `specs/`. Be careful not to accidentally commit them alongside the MCP changes if they belong to a different task.
- **Token Burning (Budget):** Teams inherently execute tasks in parallel, multiplying token usage rapidly. Any testing of the new `build_team` tool should be done carefully to avoid burning through the budget.
