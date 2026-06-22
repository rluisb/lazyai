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

LazyAI is published as Go modules and GitHub release binaries. There is no
repository-local Homebrew formula or tap automation in this repo yet, so Homebrew
installation is not a supported release claim here.

## Supported install paths

Released users install commands with:

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@latest
go install github.com/rluisb/lazyai/packages/diffviewer/cmd/lazyai-diffviewer@latest
```

Pinned installs should use the same module path with the matching submodule version, for example:

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@v0.1.0
```

## Homebrew release contract

If a Homebrew tap/formula is added later, this repository should only emit
repository-local release artifacts and documentation that Homebrew can consume:

1. Tag and publish the Go release first.
2. Build or update the Homebrew formula in the tap repository from the published binary/version metadata.
3. Keep tap publication out of this repository unless the tap repo is added as a
   first-class, reviewed dependency.

Until that exists, do not advertise `brew install lazyai` or any equivalent.

## Recommended release preparation steps

1. Confirm the package module(s) being released and the target `packages/<module>/vX.Y.Z` tag(s).
2. Review `CHANGELOG.md` and package-specific release notes.
3. Run package tests, vet, and local builds for affected modules.
4. Run `mkdocs build --strict` so release documentation stays publishable.
5. Merge the release preparation branch through the normal PR process.
6. Create the prefixed submodule tag(s) and let release automation build `lazyai-*` assets.

## GitHub release assets

The release workflow (`.github/workflows/release-<module>.yml`) cross-compiles one raw binary per OS/arch with `go build` and publishes them with `softprops/action-gh-release`. Binaries are named `<command>-<os>-<arch>`; Windows binaries add `.exe`:

- `lazyai-cli-darwin-arm64`, `lazyai-cli-darwin-amd64`, `lazyai-cli-linux-amd64`, `lazyai-cli-linux-arm64`, `lazyai-cli-windows-amd64.exe`
- `lazyai-diffviewer-<os>-<arch>` for the diffviewer module (same OS/arch matrix)
- `checksums.txt` (SHA-256 of every binary in the release)

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
