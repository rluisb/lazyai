# Contributing

## Repository structure

```text
lazyai/
├── packages/
│   ├── cli/             # Go CLI module and lazyai-cli command
│   └── diffviewer/      # Go diff viewer utility and lazyai-diffviewer command
├── docs/                # MkDocs documentation site
├── .github/workflows/   # Go CI, release, docs
└── README.md
```

See [Package Layout](../concepts/package-layout.md) for module paths and install commands.

## Requirements

- **Go 1.26+**
- **Python 3.12+** for local docs builds

Node, npm, npx, and pnpm are not required for LazyAI development.

## Local development

```bash
git clone git@github.com:rluisb/lazyai.git
cd lazyai
go test ./packages/cli/... ./packages/diffviewer/...
```

## Useful Go commands

```bash
go test ./packages/cli/... ./packages/diffviewer/...
go vet ./packages/cli/... ./packages/diffviewer/...
```

Build local binaries:

```bash
go build -o /tmp/lazyai-cli ./packages/cli/cmd/lazyai-cli
go build -o /tmp/lazyai-diffviewer ./packages/diffviewer/cmd/lazyai-diffviewer
```

## Install local commands

```bash
go install ./packages/cli/cmd/lazyai-cli
go install ./packages/diffviewer/cmd/lazyai-diffviewer
lazyai-cli --help
```

## Docs

```bash
python -m pip install -r docs/requirements.txt
mkdocs build --strict
```

The docs site deploys to GitHub Pages at <https://rluisb.github.io/lazyai/>.

## Tests

- Tests live next to Go packages as `*_test.go` files.
- Run package tests from each Go module directory with `go test ./...`.
- Run `go vet ./...` for static checks.

## Code style

- **Go**: standard `gofmt`
- Commit messages should follow the project commit pattern configured in `.ai-setup.json` when present.

## CI

- `.github/workflows/ci-cli.yml`, `.github/workflows/ci-diffviewer.yml` — per-module Go tests, vet, and builds
- `.github/workflows/test.yml` — combined test suite, coverage, and smoke test
- `.github/workflows/docs.yml` — MkDocs strict build and GitHub Pages deploy
- `.github/workflows/release-cli.yml`, `.github/workflows/release-diffviewer.yml` — GoReleaser automation triggered by `packages/<module>/vX.Y.Z` tags

## Opening issues and PRs

- Use the provided issue templates when available.
- Ensure relevant `go test`, `go vet`, and docs checks pass before submitting.
- For large changes, open a discussion or draft PR first.

## Security

- Never commit `.env` or files containing secrets.
- Follow the agent-security and security rules in the bundled library.
- Report security concerns privately to the maintainers.
