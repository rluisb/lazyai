> **SUPERSEDED (2026-06-20) — the `packages/ai-setup-ts` TypeScript bootstrap package was removed; the CLI is now Go-only (`packages/cli`). Retained for historical context.**
>
> This document is historical; `packages/ai-setup-ts` is no longer present.

# ADR-001: TypeScript CLI Bootstrap Delegates to Go Binary

**Date:** 2026-05-02
**Status:** Superseded (originally Accepted)
**Deciders:** ai-setup maintainers
**Constitution:** [`.specify/memory/constitution.md`](../../.specify/memory/constitution.md)

> **Purpose.** Capture the decision to end dual CLI implementations while preserving the published `@ai-setup/cli` npm entry point and `npx` user experience.

---

## Context

The ai-setup project currently maintains two independent implementations of the same CLI surface:

1. [`packages/ai-setup-go/`](../../packages/ai-setup-go/) — the canonical Go CLI, with a native binary, SQLite store, embedded library, migrations, adapters, scaffold pipeline, and command implementations.
2. [`packages/ai-setup-ts/`](../../packages/ai-setup-ts/) — the published npm package `@ai-setup/cli`, with a TypeScript CLI, lowdb JSON store, copied library at publish time, adapters, migrations, and scaffold logic intended to mirror Go behavior.

The documentation already identifies Go as the canonical implementation and TypeScript as a mirror. Parity work has shown that mirroring is expensive and imperfect: every feature or bug fix must be repeated across command handling, adapter behavior, migration engines, persistence, scaffold output, and release packaging. The graph report reinforces that the core abstractions cluster around filesystem reads/writes, command execution, adapters, and scaffold behavior; duplicating those abstractions in two runtimes increases drift risk.

The orchestrator package ([`packages/orchestrator/`](../../packages/orchestrator/)) is a separate MCP server package and is outside this decision.

**Related artifacts:**
- Spec(s): [`specs/004-go-migration/`](../004-go-migration/), [`specs/020-go-ts-setup-parity/`](../020-go-ts-setup-parity/), [`specs/021-parity-verification/`](../021-parity-verification/)
- Knowledge maps: [`KNOWLEDGE_MAP.md`](../../KNOWLEDGE_MAP.md), [`specs/KNOWLEDGE_MAP.md`](../KNOWLEDGE_MAP.md)
- Graph report: [`graphify-out/GRAPH_REPORT.md`](../../graphify-out/GRAPH_REPORT.md)
- Prior ADR(s): none in `specs/adrs/`

---

## Constitution Alignment

| Article | Bearing | Note |
|---|---|---|
| I — Library-First | bears | Keeps one embedded library and one implementation of adapter/scaffold behavior instead of copying library assets and logic into TypeScript. |
| III — Docs as Source of Truth | bears | Records the architecture decision before the refactor and aligns docs with the declared Go source of truth. |
| IV — Anti-Speculation | bears | Removes speculative dual-runtime parity work when the required user-facing contract can be served by one implementation. |
| V — Simplicity | bears | Replaces a full second CLI with a small bootstrap whose only responsibilities are platform detection, verified download, and exec. |
| VI — Anti-Overengineering | bears | Eliminates duplicate migration engines, adapter layers, and scaffold pipelines. |

This ADR does not amend the constitution.

---

## Options Considered (Tree of Thoughts)

### Option A — TypeScript as thin wrapper delegating to Go binary *(chosen)*

- **Summary:** Strip `packages/ai-setup-ts/` down to a small bootstrap that detects platform, downloads the Go binary from GitHub Releases with checksum verification, and execs it.
- **Complexity:** Medium initially, low after migration.
- **Reversibility:** Medium; the npm package can be republished with a different bootstrap strategy if release constraints change.
- **Performance impact:** Fast steady-state startup through the Go binary; first `npx` invocation pays a one-time binary download of approximately 15 MB.
- **Team familiarity:** High for existing Go CLI behavior; moderate for release/checksum bootstrap mechanics.
- **Constitution fit:** Strong fit for Simplicity, Anti-Speculation, and Anti-Overengineering.

### Option B — Go only; drop TypeScript package entirely

- **Summary:** Stop publishing `@ai-setup/cli` and require users to install or download the Go binary directly.
- **Complexity:** Low.
- **Reversibility:** Medium; npm discoverability could be restored later, but users would experience a distribution break.
- **Performance impact:** Fast native binary startup with no npm bootstrap.
- **Team familiarity:** High for maintainers, lower for Node.js-native users.
- **Constitution fit:** Strong simplicity fit, but weaker user adoption fit because it removes an existing distribution channel.

### Option C — Keep both implementations as status quo

- **Summary:** Continue maintaining Go and TypeScript as independent CLIs with parity checks and duplicated fixes.
- **Complexity:** High and ongoing.
- **Reversibility:** High because no migration is required now, but the maintenance burden compounds.
- **Performance impact:** Runtime-specific; TypeScript remains Node-dependent and Go remains native.
- **Team familiarity:** High with current structure.
- **Constitution fit:** Weak fit; continued duplication conflicts with Simplicity and Anti-Overengineering.

---

## Decision

Convert the existing TypeScript `@ai-setup/cli` package into a thin bootstrap that delegates all real behavior to the Go binary.

This is not a new package. The existing [`packages/ai-setup-ts/`](../../packages/ai-setup-ts/) package remains the npm distribution point for `@ai-setup/cli`, preserving the `npx` UX, but its implementation responsibility is reduced to:

1. Detect the user platform and architecture.
2. Download the matching Go binary from GitHub Releases if not already cached.
3. Verify the downloaded binary using published checksums.
4. Exec the Go binary with the original argv and stdio.

All command behavior, adapter behavior, migrations, scaffold logic, and store semantics live in Go only. The package must ship with a major version bump because runtime behavior and installation/update mechanics change materially.

---

## Rationale

- Go is already documented as the source of truth, while TypeScript is documented as a mirror. A mirror CLI creates permanent parity work without adding distinct product behavior.
- Recent parity verification found real drift between Go and TypeScript in list output, adopt/import behavior, wizard defaults, MCP merge behavior, and CLI reconciliation paths.
- Keeping `@ai-setup/cli` preserves npm and `npx` discoverability for Node.js-native teams while avoiding duplicate command implementation.
- A bootstrap-only TypeScript package has a narrow, auditable boundary: platform detection, verified release download, cache management, and exec.
- The one-time first-run download cost is acceptable compared with the ongoing cost and user-facing risk of divergent CLI behavior.

**Why the rejected options were rejected:**

- Option B was rejected because dropping npm entirely would remove an existing discovery and execution path for Node.js users.
- Option C was rejected because it preserves the highest maintenance cost and the highest probability of behavior drift.

---

## Consequences

**Positive:**

- Single source of truth for CLI behavior in Go.
- Guaranteed command parity across direct binary users and `npx @ai-setup/cli` users.
- Lower long-term maintenance cost: no duplicate adapters, migrations, scaffold pipeline, wizard defaults, or MCP compiler behavior.
- Faster steady-state CLI execution through the native Go binary.
- Existing npm package name and `npx` UX are preserved.

**Negative / accepted trade-offs:**

- First `npx` invocation must download a platform-specific binary, approximately 15 MB.
- CI/release must publish cross-platform Go binaries and checksums to GitHub Releases before the npm bootstrap can reliably resolve versions.
- Bootstrap logic introduces a new supply-chain boundary; checksum verification is mandatory.
- Offline first-run via `npx` will fail unless the binary is already cached.
- A major version bump is required to signal changed installation/update mechanics.

**Neutral:**

- Existing `.ai-setup.json` stores created by the TypeScript CLI are auto-imported into SQLite `.ai-setup.db` by the Go binary using the shared `types.StoreData` schema.
- Direct Go binary users continue using the Go CLI; they gain `ai-setup update-self` when implemented.
- Existing external setup migration/import commands remain responsible for raw Claude/OpenCode/Copilot/Gemini/Codex setup migration.
- The orchestrator MCP server remains a separate TypeScript package and is not consolidated by this ADR.

---

## Reversal Conditions

Re-open this ADR if any of the following become true:

- GitHub Releases cannot reliably publish signed or checksum-verifiable binaries for supported platforms.
- The first-run binary download creates unacceptable adoption friction that cannot be mitigated by caching, documentation, or release packaging.
- The npm bootstrap becomes more complex than its boundary responsibilities and starts re-implementing CLI behavior.
- A supported platform cannot run the Go binary and requires a native TypeScript implementation for reasons other than convenience.

---

## Implementation Pointer

Planned implementation sequence:

1. Add `ai-setup update-self` to the Go CLI for self-updates from GitHub Releases.
2. Add GitHub Release publishing to CI for cross-platform Go binaries and checksums.
3. Strip `packages/ai-setup-ts/` to the bootstrap boundary: platform detection, verified download, cache, exec.
4. Rely on existing Go-side `.ai-setup.json` to `.ai-setup.db` auto-import via `db.AutoImportJSON()` / shared `types.StoreData` compatibility.

- Plan: follow-up implementation plan/spec required before code changes.
- PRs: pending.
- Standards updated: pending.

---

## Store Compatibility

Existing TypeScript-created `.ai-setup.json` stores are migrated without user action. The Go binary auto-imports JSON data into SQLite `.ai-setup.db`, and both formats share the same `types.StoreData` schema. If both stores exist, Go store precedence and migration behavior must remain documented in the implementation plan and tests.

---

## Migration Path for Existing Users

- **TypeScript `npx` users:** the next major-version invocation of `@ai-setup/cli` downloads the Go binary, verifies it, and executes it; existing JSON stores are auto-imported.
- **Go binary users:** no required change; `update-self` becomes available once implemented.
- **Users with existing external setups:** existing `migrate` and `import` commands remain the supported path.

---

## Requirement Traceability

| Source | Requirement |
|---|---|
| User decision context | Record dual Go/TS CLI maintenance burden and Go source-of-truth status. |
| User decision | Select Option A: TypeScript bootstrap delegates to Go binary. |
| User options evaluated | Include Options A, B, and C with trade-offs. |
| User implementation plan | Capture update-self, release publishing, TS bootstrap stripping, and JSON-to-SQLite auto-import. |
| User store compatibility | Preserve no-action migration from `.ai-setup.json` to `.ai-setup.db`. |
| User migration path | Document TS `npx`, Go binary, and external setup user paths. |
| Constitution Article III | ADR precedes implementation/refactor work. |

---

## Assumptions

| Assumption | Status |
|---|---|
| `specs/adrs/` has no existing ADR files, so ADR-001 is the next available number. | Accepted |
| Maintainers intend this decision to be Accepted rather than Proposed. | Accepted |
| GitHub Releases will be the canonical binary distribution channel for the bootstrap. | Accepted |
| Checksum verification is required for downloaded binaries. | Accepted |
| The orchestrator package remains out of scope. | Accepted |
| Specific cache path, binary naming, and checksum manifest format will be defined during implementation planning. | Pending |

---

## Risks

| Risk | Level | Mitigation boundary |
|---|---|---|
| Release pipeline fails to publish complete binaries/checksums. | High | Gate npm publish on verified release artifacts. |
| Bootstrap supply-chain weakness. | High | Require checksum verification; consider signing in implementation planning. |
| First-run download friction. | Medium | Cache binary and document behavior clearly. |
| Store auto-import edge cases. | Medium | Cover `.ai-setup.json` import compatibility in implementation tests. |
| Bootstrap grows into a second implementation again. | Medium | Keep boundary explicit: detect, download, verify, exec only. |

---

## Ready for Planning

Ready for implementation planning once maintainers confirm supported platform matrix, release artifact naming, checksum manifest format, and bootstrap cache location.

---

## Memory Update

- [ ] Append ADR ratification to `.specify/memory/repos/<repo>/ledger.md` if/when that ledger exists.
- [ ] Update `KNOWLEDGE_MAP.md` and/or `specs/KNOWLEDGE_MAP.md` to point to ADR-001 as the canonical decision for TS bootstrap delegation.
- [ ] If future ADRs supersede this decision, mark this file **Superseded by ADR-[NNN]**.
