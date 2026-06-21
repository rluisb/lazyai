> **SUPERSEDED — Go/TS parity is moot; the TS CLI was removed. Retained for historical context.**
>
> This document is historical; `packages/ai-setup-ts` is no longer present.

# Plan: Go/TS Setup Engine Parity

## Phase 1 — Contract foundation

Create language-agnostic contracts for setup-engine output, target registry, MCP presets, and resource states.

Acceptance criteria:

- Contract data matches Go behavior.
- Schemas cover setup list, dry-run, and inventory output.
- Normalization rules are explicit.

## Phase 2 — TS read-only setup surface

Implement TS `setup --list` and `setup --dry-run` from the shared contract and Go behavior.

Acceptance criteria:

- TS output matches Go-normalized fixtures.
- `--tool`, `--all`, and `--global` filters match Go semantics.

## Phase 3 — TS scan parity

Implement TS `setup --scan` including target detections and reusable `.ai/agents/<id>/AGENT.md` scan.

Acceptance criteria:

- TS inventory matches Go schema and fixtures.
- Invalid agents/configs are reported consistently.

## Phase 4 — TS adopt/import parity

Implement TS `setup --adopt` and `setup --import` with Go-compatible registry and state behavior.

Acceptance criteria:

- TS uses the same resource states and registry semantics.
- Existing user setup is preserved.

## Phase 5 — MCP compiler and orchestrator parity

Align TS MCP compilation and orchestrator MCP preparation with Go.

Acceptance criteria:

- Codex, Copilot global, OpenCode, Claude, Gemini, and Pi scope behavior follows Go.
- Orchestrator MCP preparation matches Go's local build/path/smoke-test semantics.

## Phase 6 — Wizard parity

Align TS wizard logic with Go's setup step order and defaults.

Acceptance criteria:

- TS wizard state matches Go for the same inputs.

## Validation gates

- Go: `go test ./... -v -count=1`, `go vet ./...`, `go build`
- TS: `npm run lint`, `npm test`, `npm run typecheck`, `npm run build`
- Orchestrator: `npm --prefix orchestrator run build`
