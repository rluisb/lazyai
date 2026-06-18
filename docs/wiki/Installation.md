# Installation

Canonical installation docs live at <https://rluisb.github.io/lazyai/getting-started/installation/>. This Wiki page is the short-form install checklist.

## Prerequisites

- Go 1.26 or newer
- `$(go env GOPATH)/bin` on your `PATH`

Check the install target with:

```bash
go env GOPATH
```

## Install released commands

Install the CLI:

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@latest
```

Install the optional orchestrator MCP runtime:

```bash
go install github.com/rluisb/lazyai/packages/orchestrator/cmd/lazyai-orchestrator@latest
```

Install the diff viewer utility:

```bash
go install github.com/rluisb/lazyai/packages/diffviewer/cmd/lazyai-diffviewer@latest
```

## Verify installation

```bash
lazyai-cli --help
lazyai-cli doctor --help
lazyai-orchestrator --help
lazyai-diffviewer --help
```

If a command is not found, make sure `$(go env GOPATH)/bin` is on your shell `PATH`.

## Install from a local clone

```bash
git clone git@github.com:rluisb/lazyai.git
cd lazyai
cd packages/cli && go install ./cmd/lazyai-cli
cd ../orchestrator && go install ./cmd/lazyai-orchestrator
cd ../diffviewer && go install ./cmd/lazyai-diffviewer
```

## Upgrade

Re-run the relevant `go install ...@latest` command, then refresh managed project files when needed:

```bash
lazyai-cli update --check
lazyai-cli update
lazyai-cli doctor
```
