# Release Process

LazyAI is released as Go submodules from `github.com/rluisb/lazyai`. It is not published to npm.

## Go module tags

Each active package is an independent Go module and must be tagged with its module directory prefix:

| Module | Command | Tag format |
|---|---|---|
| `packages/cli` | `lazyai-cli` | `packages/cli/vX.Y.Z` |
| `packages/diffviewer` | `lazyai-diffviewer` | `packages/diffviewer/vX.Y.Z` |

Root `vX.Y.Z` tags do not version these submodules for `go install`.

## Homebrew install

`lazyai-cli` is distributed through a dedicated Homebrew tap:

- **Tap:** `rluisb/lazyai` (https://github.com/rluisb/homebrew-lazyai)
- **Formula:** `lazyai-cli`
- **Install:** `brew install rluisb/lazyai/lazyai-cli`

The canonical formula template lives at `packaging/homebrew/lazyai-cli.rb.tmpl` in this repository.
See `packaging/homebrew/README.md` for the release maintenance workflow.

Only macOS (arm64 and amd64) is supported for Homebrew installation. Linux users should use `go install` or the raw release binaries.

## Install contracts

LazyAI is published as Go modules and GitHub release binaries. The Homebrew tap
lives in a separate repository (`rluisb/homebrew-lazyai`); formula updates are a
manual post-release step.

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

The Homebrew tap (`rluisb/homebrew-lazyai`) is a separate repository. This repository
emits the rendered formula as a release artifact (`lazyai-cli.rb`) and documents the
release maintenance workflow in `packaging/homebrew/README.md`.

Tap publication is kept out of this repository; formula updates are a manual
post-release step performed in the tap repository.

## Recommended release preparation steps

1. Confirm the package module(s) being released and the target `packages/<module>/vX.Y.Z` tag(s).
2. Review `CHANGELOG.md` and package-specific release notes.
3. Run package tests, vet, and local builds for affected modules.
4. Run `mkdocs build --strict` so release documentation stays publishable.
5. **Cross-platform verification**: Consult the [Release Verification Matrix](../reports/release-verification-matrix.md) and confirm the target platform(s) pass all checks. At minimum, verify Linux (CI) and the local development platform.
6. Merge the release preparation branch through the normal PR process.
7. Create the prefixed submodule tag(s) and let release automation build `lazyai-*` assets.

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
