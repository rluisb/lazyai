# P0-1: CLI Command Import Audit

**Status:** Complete — 73 files audited  
**Owner:** Ricardo Conceicao  
**Date:** 2026-06-14  
**Linked from:** `plan.md` Phase 0, P0-1

---

## Purpose

Matrix covering every `packages/cli/cmd/*.go` file (73 files total: 50 non-test, 23 test). Each file classified by its dependency on packages marked for removal.

## Classification Key

| Classification | Meaning |
|---|---|
| `breakage` | Direct import of a package marked for removal; command will break if package is removed without rewrite |
| `rewrite` | References orchestrator/Fortnite concepts but can be restructured to use primary-agent path |
| `keep` | No dependency on removed packages; no change needed |
| `remove` | Entire command is obsolete; delete file + tests with migration note |
| `defer-with-owner` | Has dependency but rewrite is deferred; MUST remain functional; owner + deadline required |

## Target Imports to Audit

- `runtime/workflow`
- `runtime/taskqueue`
- `runtime/dispatch`
- `internal/orchestrator`
- `packages/orchestrator`
- `library/fortnite`
- `FortniteMode`
- `loop-driver`
- `orchestrator` (string references)

## Audit Matrix — Non-Test Files (50 files)

| File | Classification | Evidence | Disposition |
|---|---|---|---|
| `task.go` | **breakage** | `import runtime/taskqueue` (line 10) | Rewrite to primary-agent path |
| `workflow.go` | **breakage** | `import runtime/dispatch` (line 11), `import runtime/workflow` (line 12) | Rewrite to primary-agent path |
| `orchestration.go` | **breakage** | `import internal/orchestrator` (line 13), 15.3KB of orchestrator catalog commands | Rewrite or remove; depends on P0-2 survey |
| `session.go` | **rewrite** | `agentName = "loop-driver"` (line 107) | Replace with `"primary-agent"` |
| `helpers.go` | **rewrite** | `FortniteMode` computation (lines 173-183), `FortniteMode: fortniteMode` (line 204) | Remove FortniteMode; wire to primary-agent |
| `config.go` | **rewrite** | `DefaultAgent: "orchestrator"` (lines 92, 171) | Replace with `"primary-agent"` |
| `mcp_setup.go` | **rewrite** | Orchestrator MCP server setup (lines 27-174), `enableServers: ["orchestrator"]` | Rewrite to primary-agent MCP or remove |
| `add.go` | **rewrite** | `huh.NewOption("Orchestrator", "orchestrator")` (line 85) | Replace with primary-agent option |
| `message.go` | **rewrite** | `fromAgent = "orchestrator"` (lines 40, 145), agent list includes orchestrator (line 163) | Replace with primary-agent |
| `validate_input.go` | **rewrite** | `validAgents` includes `"orchestrator"` (line 57) | Replace with primary-agent |
| `update.go` | **rewrite** | `.opencode/agents/orchestrator.md` reference (line 182) | Replace with primary-agent path |
| `server.go` | **rewrite** | `name == "orchestrator"` special case (line 764) | Replace with primary-agent |
| `doctor.go` | **rewrite** | ai-setup orchestrator binary check comment (lines 111-112) | Update comment; no code change |
| `doctor_health.go` | **rewrite** | `lazyai-orchestrator` PATH check (lines 115-122) | Remove or replace with primary-agent check |
| `doctor_mcp.go` | **rewrite** | Legacy orchestrator MCP detection (entire file, 136 lines) | Remove legacy detection; orchestrator MCP no longer exists |
| `init.go` | **rewrite** | `--enable-servers` flag mentions orchestrator (line 32) | Update help text |
| `helpers_selection_test.go` | **rewrite** | `EnableServers: ["filesystem", "orchestrator"]` (line 21) | Replace with primary-agent server |
| `helpers_store_test.go` | **rewrite** | `EnableServers: ["filesystem", "orchestrator"]` (line 27) | Replace with primary-agent server |
| `init_test.go` | **rewrite** | `FortniteMode` tests (lines 515-560) | Remove FortniteMode tests |
| `doctor_mcp_test.go` | **rewrite** | Orchestrator MCP test fixtures (entire file) | Remove or rewrite |
| `workflow_test.go` | **breakage** | References `internal/runtime/workflow` (line 13, 145) | Rewrite or remove with workflow.go |
| `task_test.go` | **keep** | No orchestrator/Fortnite references | Keep; test may need update if task.go changes |
| `session_test.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `update_test.go` | **keep** | No orchestrator/Fortnite references | Keep; test may need update |
| `runtime_helper.go` | **keep** | No orchestrator/Fortnite references | Keep; will gain SchemaV2 migration logic in Phase 3 |
| `runtime_helper_test.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `backup.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `auth.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `auth_test.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `build_helpers.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `build_plugin.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `build_plugin_test.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `compile.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `compile_test.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `completion.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `completions.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `cost.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `create.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `doctor_health_test.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `doctor_test.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `eject.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `eval.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `git.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `import.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `import_test.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `info.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `ledger.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `ledger_integration.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `ledger_test.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `list.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `log.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `main_test.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `mcp_hints.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `mcp_hints_test.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `memory.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `metrics.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `migrate.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `models_sync.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `notify.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `quality_metrics.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `root.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `root_test.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `secret.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `setup.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `setup_test.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `sidecar.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `status.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `status_test.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `test_helpers_test.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `update-self.go` | **keep** | No orchestrator/Fortnite references | Keep; `--version` flag added for rollback |
| `validate.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `workspace.go` | **keep** | No orchestrator/Fortnite references | Keep |
| `restore_runtime_db.go` | **keep** | New file; no orchestrator/Fortnite references | New; rollback command |

## Summary

| Classification | Count | Files |
|---|---|---|
| `breakage` | 4 | `task.go`, `workflow.go`, `orchestration.go`, `workflow_test.go` |
| `rewrite` | 17 | `session.go`, `helpers.go`, `config.go`, `mcp_setup.go`, `add.go`, `message.go`, `validate_input.go`, `update.go`, `server.go`, `doctor.go`, `doctor_health.go`, `doctor_mcp.go`, `init.go`, `helpers_selection_test.go`, `helpers_store_test.go`, `init_test.go`, `doctor_mcp_test.go` |
| `keep` | 52 | All remaining files |
| `remove` | 0 | None identified as fully obsolete |
| `defer-with-owner` | 0 | None deferred |

## Evidence

- Full grep audit of all 73 `.go` files in `packages/cli/cmd/` executed 2026-06-14
- Search patterns: `runtime/workflow`, `runtime/taskqueue`, `runtime/dispatch`, `internal/orchestrator`, `packages/orchestrator`, `library/fortnite`, `FortniteMode`, `loop-driver`, `orchestrator`
- 4 files have direct import breakage; 17 files have string/reference rewrites needed; 52 files are clean

## Gate

⛔ Human must approve this audit before Phase 1 begins.
