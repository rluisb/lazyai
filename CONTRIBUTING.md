# Contributing to LazyAI

Thanks for your interest in contributing!

## Code of Conduct

We follow the [Contributor Covenant 2.1](CODE_OF_CONDUCT.md). Be respectful.

## Before You Contribute

### Is this the right place?
- **Bug report?** → Open an issue using the bug report template
- **Feature idea?** → Start a [GitHub Discussion](https://github.com/rluisb/lazyai/discussions) first, then an issue if a maintainer confirms it
- **Question or help?** → Use [GitHub Discussions](https://github.com/rluisb/lazyai/discussions)

### Before you write code
1. Open an issue describing what you plan to do
2. Wait for a maintainer to confirm the change is wanted (prevents wasted effort)
3. Fork the repository
4. Create a branch: `feat/short-description` or `fix/short-description`

## Development Setup

```bash
git clone https://github.com/YOUR_USERNAME/lazyai.git
cd lazyai
git config core.hooksPath .githooks
# Requires Go 1.26+
go work sync
make build
```

The repo-local hooks then:
- run the token-rent pre-commit check
- append a `Signed-off-by:` trailer automatically when missing

### Project Structure
```
packages/
├── cli/              # TUI CLI app (Bubble Tea + Cobra)
├── orchestrator/     # MCP-based multi-agent orchestrator
└── diffviewer/       # Terminal diff viewer
```

### Running Tests
```bash
# All packages
make test

# Single package
cd packages/cli && go test ./... -v
```

## Pull Request Process

1. **One PR = one logical change.** No mega-PRs please.
2. PR title follows the convention: `feat:`, `fix:`, `docs:`, `chore:`, `refactor:`
3. Link the related issue with `Closes #NN`
4. Fill out the PR template completely
5. All CI checks must pass (tests, vet, build, cross-compile, DCO)
6. At least one maintainer approval is required
7. You are responsible for resolving merge conflicts — not the maintainers
8. Commits must include `Signed-off-by: Your Name <email>` (DCO)

## What We Typically Reject
- Formatting-only changes (unless part of an automated linting pass)
- Large refactors without prior discussion in an issue
- Changes that break existing behavior without a clear justification
- PRs that skip the PR template
- AI-generated PRs that haven't been reviewed by a human
- New dependencies without justification

## Developer Certificate of Origin (DCO)

All commits must be signed off.

Recommended one-time setup:

```bash
git config core.hooksPath .githooks
```

With the repo hooks installed, the `commit-msg` hook appends the `Signed-off-by:` trailer automatically when it is missing. You can still use `git commit -s` explicitly if you prefer.

Expected trailer:

```
Signed-off-by: Your Name <your@email.com>
```

This certifies that you wrote the code or have the right to contribute it under the MIT license.

## Getting Help

Stuck? Open a [Discussion](https://github.com/rluisb/lazyai/discussions) — we're happy to help.
