# Installation

## Prerequisites

- **Go 1.26+**
- `$(go env GOPATH)/bin` on your `PATH` so installed binaries are discoverable

## Install released commands

LazyAI does not currently ship a repository-local Homebrew tap or formula. Use the Go install paths below for supported releases, or build from source when working in the repository.

Install the CLI:

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@latest
```

Install the diff viewer utility:

```bash
go install github.com/rluisb/lazyai/packages/diffviewer/cmd/lazyai-diffviewer@latest
```

## Clone for development

If you are working on LazyAI itself:

```bash
git clone git@github.com:rluisb/lazyai.git
cd lazyai
cd packages/cli && go install ./cmd/lazyai-cli
cd ../diffviewer && go install ./cmd/lazyai-diffviewer
```

After linking, the binary is available as:

```bash
lazyai-cli --help
```

## Upgrade from an earlier Go install

Re-run the relevant `go install ...@latest` command. For example:

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@latest
```

After upgrading the binary, refresh managed files:

```bash
lazyai-cli update --check
lazyai-cli update
lazyai-cli doctor
```

## Binary names

LazyAI currently ships these Go command packages:

```bash
lazyai-cli init
lazyai-cli compile
lazyai-cli doctor
lazyai-cli status
lazyai-diffviewer --help
```

## Symlink mode

By default, `lazyai-cli` copies library files into tool directories. You can use symlinks instead to keep all tools in sync when the library updates:

```bash
lazyai-cli init --install-mode symlink --no-interactive
```

Symlinked files are tracked in `.ai-setup.json` with `kind: "symlink"`.

## Next steps

- [Quick Start](quick-start.md)
- [How It Works](../concepts/how-it-works.md)
- [CLI Reference](../cli/reference.md)
- [Migration note for removed runtime surfaces](../migration/fortnite-orchestrator-removal.md)
