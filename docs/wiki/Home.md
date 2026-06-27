# LazyAI

LazyAI scaffolds a canonical, multi-tool AI development environment from one Go CLI. It keeps source-of-truth AI project files under `.ai/` and compiles them into native formats for OpenCode, Claude Code, GitHub Copilot, Pi, OMP, Antigravity, and Kiro.

> **Canonical docs:** GitHub Pages at <https://rluisb.github.io/lazyai/>. This Wiki is a concise bootstrap mirror for quick installation, package, release, migration, and troubleshooting notes.

## Install commands

### Homebrew (macOS)

```bash
brew install rluisb/lazyai/lazyai-cli
```

### Go install

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@latest
go install github.com/rluisb/lazyai/packages/diffviewer/cmd/lazyai-diffviewer@latest
```
The primary command is `lazyai-cli`. The diff viewer is an optional companion utility.

## Wiki map

- [[Installation]] — prerequisites, install commands, and local development installs
- [[Package Layout|Package-Layout]] — the active Go modules and command packages
- [[Go Install Versions|Go-Install-Versions]] — `@latest`, pinned versions, and submodule tags
- [[Migration from ai-setup to LazyAI|Migration-ai-setup-to-LazyAI]] — breaking rename checklist
- [[Troubleshooting]] — common install, PATH, MCP, and update issues
- [[Release Process|Release-Process]] — release preparation and asset naming notes

## Project identity

- Repository: <https://github.com/rluisb/lazyai>
- Pages docs: <https://rluisb.github.io/lazyai/>
- Go modules: `github.com/rluisb/lazyai/packages/{cli,diffviewer}`
- Distribution: Go installs and release assets only; no npm/npx package is planned for this restructure.
