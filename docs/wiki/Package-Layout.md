# Package Layout

LazyAI is a Go-only monorepo with three independent Go modules under `packages/`.

| Package | Module path | Command package | Binary |
|---|---|---|---|
| CLI | `github.com/rluisb/lazyai/packages/cli` | `github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli` | `lazyai-cli` |
| Orchestrator | `github.com/rluisb/lazyai/packages/orchestrator` | `github.com/rluisb/lazyai/packages/orchestrator/cmd/lazyai-orchestrator` | `lazyai-orchestrator` |
| Diff viewer | `github.com/rluisb/lazyai/packages/diffviewer` | `github.com/rluisb/lazyai/packages/diffviewer/cmd/lazyai-diffviewer` | `lazyai-diffviewer` |

## Source tree

```text
packages/
├── cli/
│   ├── go.mod
│   └── cmd/lazyai-cli/
├── orchestrator/
│   ├── go.mod
│   └── cmd/lazyai-orchestrator/
└── diffviewer/
    ├── go.mod
    └── cmd/lazyai-diffviewer/
```

Each package can be tested, built, and versioned independently. Root `vX.Y.Z` tags are not the Go install contract for these submodules.

## What changed from ai-setup

- `packages/ai-setup-go` became `packages/cli`.
- `packages/orchestrator-go` became `packages/orchestrator`.
- The legacy TypeScript/Node workspace is removed from the Go-only restructure.
- Public binary names are `lazyai-cli`, `lazyai-orchestrator`, and `lazyai-diffviewer`.

See [[Migration from ai-setup to LazyAI|Migration-ai-setup-to-LazyAI]] for user-facing rename steps.
