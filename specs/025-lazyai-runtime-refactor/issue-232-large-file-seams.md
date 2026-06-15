# Issue 232 Large Go File Seams

Wave 0 split only moved existing declarations/tests into focused files; behavior was not rewritten.

| Original hotspot | Before | After original | New seam file | New file lines | Seam |
| --- | ---: | ---: | --- | ---: | --- |
| `packages/cli/internal/adapter/adapter_test.go` | 1232 | 711 | `packages/cli/internal/adapter/opencode_adapter_test.go` (490), `packages/cli/internal/adapter/managed_block_test.go` (53) | 543 | OpenCode adapter install/scope/config migration tests, plus managed-block merge tests |
| `packages/cli/internal/scaffold/root.go` | 947 | 718 | `packages/cli/internal/scaffold/root_targeted_update.go` | 240 | Targeted AGENTS.md update patch contract and slot replacement policy |
| `packages/cli/internal/setupscan/absorb.go` | 835 | 689 | `packages/cli/internal/setupscan/absorb_mcp_entries.go` | 163 | MCP entry snapshotting, fingerprinting, and registry state comparison |

Recommended targeted verification for the main Wave 0 gate:

```sh
go test ./packages/cli/internal/adapter -run 'Test(OpenCodeAdapter|MergeManagedBlock)' -count=1
go test ./packages/cli/internal/scaffold -run 'Test.*Targeted|TestScaffoldCompiledRoot' -count=1
go test ./packages/cli/internal/setupscan -run 'Test.*(MCP|Absorb|Run)' -count=1
```
