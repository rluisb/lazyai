# Task 002: Extract Shared Runtime Catalog Helper

## Goal

Avoid divergence between MCP handlers and the `init` CLI by centralizing runtime catalog loading.

## Files

- `src/catalog/runtime.ts` (new)
- `src/tool-handlers.ts`

## Implementation Notes

Create helper:

```ts
export interface RuntimeCatalogOptions {
  projectRoot: string
  libraryOrchestrationRoot?: string
  libraryAgentsRoot?: string
  hostCli?: HostCli
  db?: Db
}

export function loadRuntimeCatalog(options: RuntimeCatalogOptions): OrchestrationCatalog
```

Implementation should:

1. call `loadCatalog()`
2. resolve supported host CLI (`opencode`, `claude-code`, `codex`) for host scanning
3. call `resolveCatalog()`
4. default `db` to `getPersistenceDb()` if not provided

Refactor `OrchestratorToolHandlers.getCatalog()` to call this helper.

## Done When

- Existing handler behavior is unchanged.
- Tests pass.
- `init` can reuse helper directly.
