# Minimality report

LazyAI includes a report-only minimality check for spotting growth before it becomes a hard gate. Run it from the repository root:

```bash
go run ./packages/cli/internal/minimality/cmd/minimality-report
```

Use `--root <path>` when running from another directory.

## What the report measures

The report is deterministic and reads repository files only. It prints:

- Go files over 500 lines, sorted by line count descending and then path ascending.
- Top-level CLI command registrations from `packages/cli/cmd` source files, split into visible, hidden, and total registered commands.
- Canonical library byte count and the configured token-rent budget for `packages/cli/library/canonical/`.
- Canonical asset counts for agents, skills, hooks, prompts, templates, and rules when those directories are present.

Retired snapshots, worktrees, temporary directories, vendored dependencies, and hidden directories are excluded from Go-file line scanning so the report stays focused on active repository code.

## Advisory thresholds

These thresholds are advisory in this report:

| Signal | Advisory threshold | How to read it |
| --- | --- | --- |
| Go file length | Files over 500 lines are listed. | Treat listed files as review candidates for future simplification or seam extraction. |
| Top-level CLI commands | No hard maximum yet. | Watch trend growth; this report does not classify or block command categories. |
| Canonical asset counts | No hard maximum yet. | Watch growth in embedded agents, skills, hooks, prompts, templates, and rules. |

The report exits 0 even when advisory thresholds are exceeded. A future issue may promote specific thresholds to gates after the team agrees on budgets.

## Token-rent remains the hard gate

Canonical library bytes are informational in the minimality report. The existing token-rent check remains the hard gate for the canonical byte budget:

```bash
go run ./packages/cli/internal/tokenrent/cmd/token-rent-check
```

If token-rent fails, fix the canonical library size or add the documented `.lazyai/token-rent-override` with a non-empty `reason:`. Do not treat the minimality report as a replacement for token-rent enforcement.
## Issue #305 split report

The following files still exceed 500 lines after this split and are explicit exceptions for this issue:

- `packages/cli/internal/adapter/copilot.go` (~570 lines, adapter implementation boundary)
- `packages/cli/internal/adapter/claudecode.go` (~546 lines, adapter implementation boundary)
- `packages/cli/internal/adapter/mcp_compiler.go` (~720 lines, implementation-heavy compiler orchestration)
- `packages/cli/internal/adapter/mcp_compiler_test.go` (~574 lines, existing integration-style test scope)
- `packages/cli/internal/adapter/opencode_adapter_test.go` (~598 lines, integration surface for external plugin behavior)
