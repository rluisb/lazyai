# Release Process

LazyAI is released as Go submodules from `github.com/rluisb/lazyai`. It is not published to npm.

## Go module tags

Each active package is an independent Go module and must be tagged with its module directory prefix:

| Module | Command | Tag format |
|---|---|---|
| `packages/cli` | `lazyai-cli` | `packages/cli/vX.Y.Z` |
| `packages/diffviewer` | `lazyai-diffviewer` | `packages/diffviewer/vX.Y.Z` |

Root `vX.Y.Z` tags do not version these submodules for `go install`.

## Install contracts

Released users install commands with:

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@latest
go install github.com/rluisb/lazyai/packages/diffviewer/cmd/lazyai-diffviewer@latest
```

Pinned installs should use the same module path with the matching submodule version, for example:

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@v0.1.0
```

## Recommended release preparation steps

1. Confirm the package module(s) being released and the target `packages/<module>/vX.Y.Z` tag(s).
2. Review `CHANGELOG.md` and package-specific release notes.
3. Run package tests, vet, and local builds for affected modules.
4. Run `mkdocs build --strict` so release documentation stays publishable.
5. Merge the release preparation branch through the normal PR process.
6. Create the prefixed submodule tag(s) and let release automation build `lazyai-*` assets.

## GitHub release assets

GoReleaser (`packages/<module>/.goreleaser.yaml`) publishes one archive per OS/arch plus a checksum file. `<Os>` is title-cased and `amd64` renders as `x86_64`:

- `lazyai-cli_<Os>_<Arch>.tar.gz` — e.g. `lazyai-cli_Linux_x86_64.tar.gz`, `lazyai-cli_Darwin_arm64.tar.gz`
- `lazyai-diffviewer_<Os>_<Arch>.tar.gz` — e.g. `lazyai-diffviewer_Linux_x86_64.tar.gz`
- Windows archives use `.zip` — e.g. `lazyai-cli_Windows_x86_64.zip`
- `checksums.txt`

## Upgrading commands

End users upgrade by re-running the relevant Go install command:

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@latest
```

After upgrading the CLI binary, refresh managed files:

```bash
lazyai-cli update --check
lazyai-cli update
lazyai-cli doctor
```

## Store migration

Existing managed projects may still contain `.ai-setup.json` and `.ai-setup.db`. Those file names are part of the current local state contract and are not automatically renamed in the LazyAI package restructure.

See [Migration from ai-setup to LazyAI](../migration/ai-setup-to-lazyai.md) for user-facing rename guidance.
