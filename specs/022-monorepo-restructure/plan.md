# Spec 022 — Monorepo Restructure

**Status:** ✅ Complete (retrospective documentation; implemented in commit `0527034`)
**Date:** 2026-04-24

## Goal

Move the repo from a flat polyglot layout (Go + TS interleaved at the root) to a proper monorepo with:
- `packages/ai-setup-go/` — Go runtime
- `packages/ai-setup-ts/` — TypeScript runtime
- `packages/orchestrator/` — separate TS package
- pnpm workspaces for JS/TS, Go workspaces (`go.work`) for Go

## Why

The runtime-parity rule ("every feature in one runtime must be in the other") is easier to enforce and review when the two runtimes live side-by-side as explicit peer packages. The flat layout conflated root-level Go and TS config (`go.mod` next to `package.json`, `tsconfig.json` next to `cmd/`).

## Layout

```
/
├── packages/
│   ├── ai-setup-go/      cmd/ internal/ tui/ main.go go.mod go.sum migrations/ Makefile
│   │   └── library/      canonical library assets (required by //go:embed all:library)
│   ├── ai-setup-ts/      src/ bin/ package.json tsconfig.json tsup.config.ts vitest.config.ts biome.json
│   └── orchestrator/     moved from /orchestrator
├── library → packages/ai-setup-go/library   # symlink; single source of truth
├── specs/ docs/ examples/ demo/ scripts/   # shared, unchanged
├── go.work               # links packages/ai-setup-go
├── pnpm-workspace.yaml   # globs packages/ai-setup-ts + packages/orchestrator
├── package.json          # root pnpm workspace; shared dev deps; orchestration scripts
├── Makefile              # delegates to per-package Makefiles / pnpm scripts
└── README.md, CLAUDE.md, AGENTS.md, etc.   (unchanged)
```

## Key design decisions

### `library/` at root is a symlink

The canonical location is `packages/ai-setup-go/library/` because Go's `//go:embed all:library` directive requires the embedded tree to live under the Go module root (embed doesn't follow symlinks that escape the module).

A symlink at the repo root (`library → packages/ai-setup-go/library`) preserves the pre-monorepo convention where walk-up resolution from anywhere in the tree finds `library/` at the root. TS's `resolveLibraryDir` walk-up (looking for the `library/mcp/catalog.json` sentinel) still works because POSIX symlinks resolve transparently.

**Windows note:** requires `git config core.symlinks true` or developer mode.

### Tooling: pnpm + Go workspaces

- **pnpm via corepack** — `corepack enable pnpm` activates it from `engines.pnpm` in root `package.json`. No global install required.
- **Go workspaces (`go.work`)** — native multi-module support, zero external tooling.
- No Nx/Turborepo/Bazel — overkill for 3 workspace packages.

### Publish flow preserves single-file shipping

`packages/ai-setup-ts/package.json` uses `prepublishOnly` to copy `../../library` into the package before publish, and `postpublish` to remove it. This keeps the published npm tarball self-contained without bloating git history or fragmenting the source.

`.gitignore` excludes the temporary `packages/ai-setup-ts/library/` copy.

## Breakage + fixes (all in the same commit)

| Location | Change |
|---|---|
| `packages/ai-setup-ts/src/utils/files.ts#resolveLibraryDir` | Walks up for the `library/mcp/catalog.json` sentinel instead of `package.json + library`. Matches Go's `internal/library/embed.go#walkUpFromLibrary`. |
| `packages/ai-setup-ts/src/__tests__/test-helpers.ts` | New shared `findMonorepoLibraryDir()` for tests that used to rely on `process.cwd()`. |
| `packages/orchestrator/src/loader.ts#getDefaultLibraryRoots` | Walks up for the library sentinel instead of assuming `../../library`. |
| `packages/ai-setup-go/tui/wizard/phase1_cli_mcp.go#loadMcpCatalog` | Walks up for the library sentinel instead of a fixed `../../` depth. |
| `packages/ai-setup-ts/package.json` | Added `picocolors` dependency (was hoisted before). Added `prepublishOnly` / `postpublish` scripts. |

## CI changes

- `.github/workflows/ci.yml` — rewritten to use pnpm; runs TS + orchestrator jobs in parallel matrices.
- `.github/workflows/go-ci.yml` — paths filter widened to `packages/ai-setup-go/**` and `library/**`; working-directory set.
- `.github/workflows/publish.yml` — publishes `packages/ai-setup-ts` specifically via pnpm.

## Verification

- `pnpm install`: 229 packages installed cleanly
- `pnpm -r run typecheck`: clean
- `pnpm -r run test`: 375 (ai-setup-ts) + 98 (orchestrator) = 473 tests passing (1 skipped)
- `go test ./...` from `packages/ai-setup-go`: all packages passing
- `go build .` from `packages/ai-setup-go`: produces working binary

## Follow-up notes

- If a future contributor works on Windows, `git config core.symlinks true` may be needed to properly checkout the `library` symlink at root.
- The `packages/ai-setup-ts/prepublishOnly` copy approach means publishing from a dirty working tree will generate a stray `packages/ai-setup-ts/library/` directory that gets cleaned on `postpublish`. If publish fails mid-flight, this dir may linger — safely removable via `rm -rf packages/ai-setup-ts/library`.
