# Package Layout

LazyAI is a Go-only monorepo with independent Go modules under `packages/`.

## Modules and commands

| Package | Module path | Command package | Binary |
|---|---|---|---|
| CLI | `github.com/rluisb/lazyai/packages/cli` | `github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli` | `lazyai-cli` |
| Orchestrator | `github.com/rluisb/lazyai/packages/orchestrator` | `github.com/rluisb/lazyai/packages/orchestrator/cmd/lazyai-orchestrator` | `lazyai-orchestrator` |
| Diff viewer | `github.com/rluisb/lazyai/packages/diffviewer` | `github.com/rluisb/lazyai/packages/diffviewer/cmd/lazyai-diffviewer` | `lazyai-diffviewer` |

## Install commands

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@latest
go install github.com/rluisb/lazyai/packages/orchestrator/cmd/lazyai-orchestrator@latest
go install github.com/rluisb/lazyai/packages/diffviewer/cmd/lazyai-diffviewer@latest
```

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

Each module can be tested and versioned independently. See [Release Process](../development/release.md) for submodule tag requirements.
