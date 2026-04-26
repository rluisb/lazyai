# Plan: Setup Engine Hardening

## Overview
Hardening the `ai-setup` engine to transition from a simple script-based approach to a resource-managed lifecycle. Focus on safety, absorption of existing setups, and target-specific adapter stability.

## Phased Implementation

### Phase 1: Inventory & Scanning
- Implement the scanning logic for Tool Targets.
- Build the Inventory system to track current state vs. desired state.
- **Acceptance Criteria**: `setup --scan` returns a JSON report of all detected tool configs and their versions/origins.

### Phase 2: Safe Absorption Workflow
- Implement the backup mechanism for config files.
- Build the conflict detection logic (Match, Sensitive, Conflict).
- Implement the `--adopt` and `--import` flags.
- **Acceptance Criteria**: Running `setup --adopt` on a pre-existing config creates a backup and adds the config to the managed registry without data loss.

### Phase 3: Target Adapter Expansion
- Implement the **Pi Go adapter** to bring Pi into the supported targets.
- Refactor existing adapters (Claude, Gemini, etc.) to follow the new `Install/Uninstall/Adopt` interface.
- **Acceptance Criteria**: All targets (including Pi) can be initialized via the new engine.

### Phase 4: Directory-Based Agent Definitions
- Implement the `.ai/agents/` directory structure following the Compozy shape.
- Create logic to parse `AGENT.md` and optional `mcp.json` from folders.
- Support agent-local MCP configuration using the standard `mcpServers` shape.
- **Acceptance Criteria**: An agent created in `.ai/agents/my-agent/` effectively configures the AI assistant using its `AGENT.md` and restricts/adds tools via `mcp.json`.

### Phase 5: Command CLI Hardening
- Implement the full suite of proposed commands: `--list`, `--scan`, `--dry-run`, `--tool`, `--all`, `--global`, `--adopt`.
- Integrate the dry-run output into the CLI.
- **Acceptance Criteria**: User can perform a full-cycle setup (Scan $\rightarrow$ Dry-run $\rightarrow$ Apply) for any target.

### Phase 6: Validation & Quality Gates
- Implement the automated validation suite for both Go and TS components.
- Build the orchestrator build check.
- **Acceptance Criteria**: All CI/CD gates pass (Lint, Typecheck, Test, Build).

## Phase 7: Non-Goals (Out-of-Scope)
- $\times$ **Runtime Execution**: No `ai-setup exec`.
- $\times$ **Session Management**: No daemon or prompt state.
- $\times$ **ACP Execution**: No running of agents within the setup tool.

## Risks & Mitigations
| Risk | Severity | Mitigation |
|---|---|---|
| Config Corruption | High | Mandatory `.bak` files and checksums before any write operation. |
| Token Bloat | Medium | Strict enforcement of agent-local tool whitelists. |
| Adapter Drift | Medium | Shared interface for all adapters; unified test suite. |
| Pathing Issues | Low | Use of absolute paths and environment variable resolution in `~/.ai-setup/`. |
