# Go Install Versions

LazyAI packages are Go submodules. Use the command package path plus either `@latest` or a pinned semantic version.

## Latest installs

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@latest
go install github.com/rluisb/lazyai/packages/orchestrator/cmd/lazyai-orchestrator@latest
go install github.com/rluisb/lazyai/packages/diffviewer/cmd/lazyai-diffviewer@latest
```

## Pinned installs

Use the same command path with a version suffix:

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@v0.1.0
go install github.com/rluisb/lazyai/packages/orchestrator/cmd/lazyai-orchestrator@v0.1.0
go install github.com/rluisb/lazyai/packages/diffviewer/cmd/lazyai-diffviewer@v0.1.0
```

The version suffix is the module version selected by Go, but the Git tags in this repository must include the package directory prefix.

## Required tag formats

| Module | Git tag format | User install suffix |
|---|---|---|
| `packages/cli` | `packages/cli/vX.Y.Z` | `@vX.Y.Z` |
| `packages/orchestrator` | `packages/orchestrator/vX.Y.Z` | `@vX.Y.Z` |
| `packages/diffviewer` | `packages/diffviewer/vX.Y.Z` | `@vX.Y.Z` |

Examples for a coordinated `v0.1.0` release:

```bash
git tag packages/cli/v0.1.0
git tag packages/orchestrator/v0.1.0
git tag packages/diffviewer/v0.1.0
```

Root tags such as `v0.1.0` do not version these submodules for `go install`.

## When `@latest` does not update

- Confirm the relevant prefixed submodule tag exists.
- Clear stale module cache only when needed: `go clean -modcache`.
- Check that the command path matches the package you are installing.
- Use a pinned version temporarily if you need a specific release.
