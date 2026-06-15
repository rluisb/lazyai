# Release Process

Canonical release docs live at <https://rluisb.github.io/lazyai/development/release/>. This Wiki page summarizes the public package/version contract.

## Package tags

LazyAI has independent active Go modules under `packages/`, so release tags must be prefixed by module directory:

| Module | Command | Tag format |
|---|---|---|
| `packages/cli` | `lazyai-cli` | `packages/cli/vX.Y.Z` |
| `packages/diffviewer` | `lazyai-diffviewer` | `packages/diffviewer/vX.Y.Z` |

Root `vX.Y.Z` tags do not version these submodules for Go users.

## Release preparation checklist

1. Confirm which package modules are being released.
2. Update release notes or changelog entries.
3. Run tests, vet, and local builds for affected modules.
4. Run `mkdocs build --strict` so Pages remains publishable.
5. Merge through the normal PR process.
6. Create prefixed submodule tags, for example `packages/cli/v0.1.0`.
7. Let release automation build assets named with `lazyai-*`.

## User install contract

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@latest
go install github.com/rluisb/lazyai/packages/diffviewer/cmd/lazyai-diffviewer@latest
```

Pinned installs use the same paths with `@vX.Y.Z`.

## Release asset names

- `lazyai-cli-<os>-<arch>[.exe]`
- `lazyai-diffviewer-<os>-<arch>[.exe]`
- `checksums.txt`
