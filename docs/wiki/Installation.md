# Installation

Canonical installation docs live at <https://rluisb.github.io/lazyai/getting-started/installation/>. This Wiki page is the short-form install checklist.

## Prerequisites

- Go 1.26 or newer (for `go install` path)
- `$(go env GOPATH)/bin` on your `PATH`

Check the install target with:

```bash
go env GOPATH
```

## Homebrew (macOS)

```bash
brew install rluisb/lazyai/lazyai-cli
```

Or tap first:

```bash
brew tap rluisb/lazyai
brew install lazyai-cli
```

Only macOS (arm64 and amd64) is supported for Homebrew installation.

## Go install

Install the CLI:

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@latest
```

Install the diff viewer utility:

```bash
go install github.com/rluisb/lazyai/packages/diffviewer/cmd/lazyai-diffviewer@latest
```


## Verify installation

```bash
lazyai-cli --help
lazyai-cli doctor --help
lazyai-diffviewer --help
```

If a command is not found, make sure `$(go env GOPATH)/bin` is on your shell `PATH`.

## Install from a local clone

```bash
git clone git@github.com:rluisb/lazyai.git
cd lazyai
go install ./packages/cli/cmd/lazyai-cli
go install ./packages/diffviewer/cmd/lazyai-diffviewer
```

## Upgrade

Re-run the relevant `go install ...@latest` command, then refresh managed project files when needed:

```bash
lazyai-cli update --check
lazyai-cli update
lazyai-cli doctor
```
