# Troubleshooting

Canonical troubleshooting docs live at <https://rluisb.github.io/lazyai/troubleshooting/faq/>.

## `lazyai-cli` command not found

Make sure Go's binary directory is on your `PATH`:

```bash
go env GOPATH
export PATH="$(go env GOPATH)/bin:$PATH"
```

Then reinstall and verify:

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@latest
lazyai-cli --help
```

## `go install ...@latest` installs an old version

LazyAI uses submodule tags. Confirm the package has a prefixed tag such as `packages/cli/vX.Y.Z`, `packages/orchestrator/vX.Y.Z`, or `packages/diffviewer/vX.Y.Z`.

If needed, install a pinned version:

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@v0.1.0
```

## Orchestrator MCP server does not start

Install the runtime and verify the generated MCP command:

```bash
go install github.com/rluisb/lazyai/packages/orchestrator/cmd/lazyai-orchestrator@latest
lazyai-orchestrator --help
```

MCP configs should invoke:

```json
{
  "command": "lazyai-orchestrator",
  "args": ["connect"]
}
```

Then run:

```bash
lazyai-cli compile
lazyai-cli doctor
```

## npm or npx commands no longer work

LazyAI is Go-only in this restructure. Replace npm/npx usage with the Go install commands on [[Installation]].

## Existing project still has `.ai-setup.json`

That local manifest name remains part of the current managed project state. The package restructure does not automatically rename `.ai-setup.json`, `.ai-setup.db`, `.ai-setup.toml`, or `.ai-setup-backup/`.

## Files drift after upgrade

Preview and apply updates:

```bash
lazyai-cli update --check
lazyai-cli update
lazyai-cli doctor --skills-check
```

Use `lazyai-cli eject` only if you want to stop LazyAI management while leaving generated files in place.
