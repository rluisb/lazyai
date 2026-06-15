# Package Layout

LazyAI is a Go-only monorepo with active Go modules under `packages/`.

| Package | Module path | Command package | Binary |
|---|---|---|---|
| CLI | `github.com/rluisb/lazyai/packages/cli` | `github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli` | `lazyai-cli` |
| Diff viewer | `github.com/rluisb/lazyai/packages/diffviewer` | `github.com/rluisb/lazyai/packages/diffviewer/cmd/lazyai-diffviewer` | `lazyai-diffviewer` |

## Source tree

```text
packages/
├── cli/
│   ├── go.mod
│   └── cmd/lazyai-cli/
└── diffviewer/
    ├── go.mod
    └── cmd/lazyai-diffviewer/
```

Each active package can be tested, built, and versioned independently. Root `vX.Y.Z` tags are not the Go install contract for these submodules.

## What changed from ai-setup

- `packages/ai-setup-go` became `packages/cli`.
- The legacy TypeScript/Node workspace is removed from the Go-only restructure.
- The old dedicated orchestration runtime was retired from the active workspace.
- Public binary names are `lazyai-cli` and `lazyai-diffviewer`.

See [[Migration from ai-setup to LazyAI|Migration-ai-setup-to-LazyAI]] for user-facing rename steps.
