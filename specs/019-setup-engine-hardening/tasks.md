# Tasks: Setup Engine Hardening

## Phase 1: Inventory & Scanning
- [x] **Task 1.1**: Implement Tool Target scanning logic (Go).
- [x] **Task 1.2**: Create the Inventory data structure to track `CurrentState` vs `DesiredState`.
- [x] **Task 1.3**: Implement `setup --scan` command.
- **Validation**: Run `ai-setup setup --scan` and verify JSON output matches filesystem reality. ✅

## Phase 2: Safe Absorption Workflow
- [x] **Task 2.1**: Implement the backup utility (creates `.bak` with timestamps).
- [x] **Task 2.2**: Implement conflict detection logic (Match/Sensitive/Conflict/Adoptable/UserOwned/Missing).
- [x] **Task 2.3**: Implement `--adopt` flag to mark external configs as managed.
- [x] **Task 2.4**: Implement `--import` flag to copy external configs into `~/.ai-setup/`.
- **Validation**: Perform a setup on a directory with pre-existing config and verify backup creation. ✅

## Phase 3: Target Adapter Expansion
- [x] **Task 3.1**: Implement the **Pi Go adapter** (Target: Pi).
- [x] **Task 3.2**: Standardize `ClaudeCode`, `Gemini`, `Copilot`, `Codex`, `OpenCode` adapters to the new `Sourcing` interface.
- **Validation**: `go build` and `go test ./... -v -count=1` for all adapters. ✅

## Phase 4: Resource Selection Flow
- [x] **Task 4.1**: Implement the sequential selection wizard: Tool Targets → Skills → Agents → MCP Servers.
- [x] **Task 4.2**: Implement MCP preset logic (Minimal, Recommended, Full).
- [x] **Task 4.3**: Create the MCP resource selector for individual tool toggling.
- **Validation**: Verify the user can select a preset and then manually adjust specific tools. ✅

## Phase 5: Orchestrator MCP Setup
- [x] **Task 5.1**: Implement orchestrator package build/install logic.
- [x] **Task 5.2**: Implement wiring of orchestrator MCP into AI CLI tool configs.
- [x] **Task 5.3**: Implement binary/path verification and optional smoke test.
- [x] **Task 5.4**: Implement logic to preserve/adopt existing orchestrator entries.
- [x] **Task 5.5**: Implement ownership recording in the setup manifest.
- **Validation**: Verify that the orchestrator MCP is correctly configured across selected targets and functional. ✅

## Phase 6: Directory-Based Agent Definitions
- [x] **Task 6.1**: Implement folder-based agent detection in `.ai/agents/`.
- [x] **Task 6.2**: Create parser for `AGENT.md`.
- [x] **Task 6.3**: Implement agent-local MCP configuration logic using `mcp.json` (standard `mcpServers` shape).
- **Validation**: Create a test agent folder (`.ai/agents/test-agent/`) with `AGENT.md` and `mcp.json` and verify they are correctly parsed and applied. ✅

## Phase 7: Command CLI Hardening
- [x] **Task 7.1**: Implement `setup --list`.
- [x] **Task 7.2**: Implement `setup --dry-run` (simulation mode).
- [x] **Task 7.3**: Implement `setup --tool <name>` and `setup --all`.
- [x] **Task 7.4**: Implement `setup --global`.
- **Validation**: Verify all commands produce correct help text and expected output. ✅

## Phase 8: Validation & Quality Gates
- [x] **Task 8.1**: Configure Go gates: `go vet ./...`, `go test ./... -v -count=1`, `go build`.
- [x] **Task 8.2**: Configure TS gates: `npm run typecheck`, `npm test`, `npm run lint`, `npm run build`.
- [x] **Task 8.3**: Configure Orchestrator gate: `npm --prefix orchestrator run build`.
- **Validation**: Run all gates and ensure zero failures. ✅

## Dependency Order
1. → 1.1 → 1.2 → 1.3 ✅
2. 1.3 → 2.1 → 2.2 → 2.3 → 2.4 ✅
3. 2.4 → 3.1 → 3.2 ✅
4. 3.2 → 4.1 → 4.2 → 4.3 ✅
5. 4.3 → 5.1 → 5.2 → 5.3 → 5.4 → 5.5 ✅
6. 5.5 → 6.1 → 6.2 → 6.3 ✅
7. 6.3 → 7.1 → 7.2 → 7.3 → 7.4 ✅
8. 7.4 → 8.1 → 8.2 → 8.3 ✅

---

**Delivered by**: PR #147 (`feat(setup): add setup engine hardening`) and PR #148 (`feat(setup): add TypeScript setup parity`), with parity gaps closed post-merge.
