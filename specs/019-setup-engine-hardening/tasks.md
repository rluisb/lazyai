# Tasks: Setup Engine Hardening

## Phase 1: Inventory & Scanning
- [ ] **Task 1.1**: Implement Tool Target scanning logic (Go).
- [ ] **Task 1.2**: Create the Inventory data structure to track `CurrentState` vs `DesiredState`.
- [ ] **Task 1.3**: Implement `setup --scan` command.
- **Validation**: Run `ai-setup setup --scan` and verify JSON output matches filesystem reality.

## Phase 2: Safe Absorption Workflow
- [ ] **Task 2.1**: Implement the backup utility (creates `.bak` with timestamps).
- [ ] **Task 2.2**: Implement conflict detection logic (Match/Sensiitve/Conflict/Adoptable/UserOwned/Missing).
- [ ] **Task 2.3**: Implement `--adopt` flag to mark external configs as managed.
- [ ] **Task 2.4**: Implement `--import` flag to copy external configs into `~/.ai-setup/`.
- **Validation**: Perform a setup on a directory with pre-existing config and verify backup creation.

## Phase 3: Target Adapter Expansion
- [ ] **Task 3.1**: Implement the **Pi Go adapter** (Target: Pi).
- [ ] **Task 3.2**: Standardize `ClaudeCode`, `Gemini`, `Copilot`, `Codex`, `OpenCode` adapters to the new `Sourcing` interface.
- **Validation**: `go build` and `go test ./... -v -count=1` for all adapters.

## Phase 4: Resource Selection Flow
- [ ] **Task 4.1**: Implement the sequential selection wizard: Tool Targets $\rightarrow$ Skills $\rightarrow$ Agents $\rightarrow$ MCP Servers.
- [ ] **Task 4.2**: Implement MCP preset logic (Minimal, Recommended, Full).
- [ ] **Task 4.3**: Create the MCP resource selector for individual tool toggling.
- **Validation**: Verify the user can select a preset and then manually adjust specific tools.

## Phase 5: Orchestrator MCP Setup
- [ ] **Task 5.1**: Implement orchestrator package build/install logic.
- [ ] **Task 5.2**: Implement wiring of orchestrator MCP into AI CLI tool configs.
- [ ] **Task 5.3**: Implement binary/path verification and optional smoke test.
- [ ] **Task 5.4**: Implement logic to preserve/adopt existing orchestrator entries.
- [ ] **Task 5.5**: Implement ownership recording in the setup manifest.
- **Validation**: Verify that the orchestrator MCP is correctly configured across selected targets and functional.

## Phase 6: Directory-Based Agent Definitions
- [ ] **Task 6.1**: Implement folder-based agent detection in `.ai/agents/`.
- [ ] **Task 6.2**: Create parser for `AGENT.md`.
- [ ] **Task 6.3**: Implement agent-local MCP configuration logic using `mcp.json` (standard `mcpServers` shape).
- **Validation**: Create a test agent folder (`.ai/agents/test-agent/`) with `AGENT.md` and `mcp.json` and verify they are correctly parsed and applied.

## Phase 7: Command CLI Hardening
- [ ] **Task 7.1**: Implement `setup --list`.
- [ ] **Task 7.2**: Implement `setup --dry-run` (simulation mode).
- [ ] **Task 7.3**: Implement `setup --tool <name>` and `setup --all`.
- [ ] **Task 7.4**: Implement `setup --global`.
- **Validation**: Verify all commands produce correct help text and expected output.

## Phase 8: Validation & Quality Gates
- [ ] **Task 8.1**: Configure Go gates: `go vet ./...`, `go test ./... -v -count=1`, `go build`.
- [ ] **Task 8.2**: Configure TS gates: `npm run typecheck`, `npm test`, `npm run lint`, `npm run build`.
- [ ] **Task 8.3**: Configure Orchestrator gate: `npm --prefix orchestrator run build`.
- **Validation**: Run all gates and ensure zero failures.

## Dependency Order
1. $\rightarrow$ 1.1 $\rightarrow$ 1.2 $\rightarrow$ 1.3
2. 1.3 $\rightarrow$ 2.1 $\rightarrow$ 2.2 $\rightarrow$ 2.3 $\rightarrow$ 2.4
3. 2.4 $\rightarrow$ 3.1 $\rightarrow$ 3.2
4. 3.2 $\rightarrow$ 4.1 $\rightarrow$ 4.2 $\rightarrow$ 4.3
5. 4.3 $\rightarrow$ 5.1 $\rightarrow$ 5.2 $\rightarrow$ 5.3 $\rightarrow$ 5.4 $\rightarrow$ 5.5
6. 5.5 $\rightarrow$ 6.1 $\rightarrow$ 6.2 $\rightarrow$ 6.3
7. 6.3 $\rightarrow$ 7.1 $\rightarrow$ 7.2 $\rightarrow$ 7.3 $\rightarrow$ 7.4
8. 7.4 $\rightarrow$ 8.1 $\rightarrow$ 8.2 $\rightarrow$ 8.3
