# Ownership prior art: ADR-013 vs LazyAI 1.x lockfile

Research for [#640](https://github.com/rluisb/lazyai/issues/640), part of map [#632](https://github.com/rluisb/lazyai/issues/632).
This document reports facts only and makes no design recommendations. LazyAI file references point at `main` @ `6b2b6f9d`.

## 1. ADR-013 (agent-context-broker)

Source: `~/projects/personal/agent-context-broker/specs/adrs/013-boundaries-and-ownership.md`. The title is "Runtime, installer, harness, and memory ownership".

- **Status: Proposed.** No broker code exists yet. The ADR describes the intended boundary, and the M0 inventory still has to confirm or refute it (lines 3-5, 16-17). ai-memory (`personal/agent-context-broker`) only mentions the "M0-05 ownership ADR review + manifest" as a planned ticket. It holds no decision record beyond the ADR.
- **Four components, one owner each** (lines 21-34): Runtime (it never edits harness config), Installer, Harness (planning, editing and approvals), and Providers (CodeGraph and ai-memory own their indexes and capture).
- **Installer contract** (lines 26-29): it owns only broker-added harness config, *marked* guidance blocks, and broker-owned skill/adapter assets. Every install is **additive, idempotent, preview-by-default, journaled and reversible**. It "writes ownership markers plus a manifest; **uninstall removes only owned content that is unchanged**."
- **Key-level scope for Codex** (lines 36-39): the installer owns only `mcp_servers.context-broker` in the chosen TOML scope, plus any explicitly marked `AGENTS.md` block. The rest of those files is user-owned.
- **Consequences** (lines 43-52): uninstall is bounded to unchanged owned content. User-modified owned content is *preserved and flagged for reconciliation*. Install-time conflicts become detectable. The cost is a **durable journal**, ownership markers and **crash-recovery paths**, plus per-harness adapters and conflict rules. Markers must survive harness self-rewrites, and M0 has to test this.
- **Unverified** (lines 54-57): it is not yet known whether harnesses tolerate third-party markers in their config.
- **Rejected options** (lines 61-71): a single binary that does both runtime and install, direct writes without a journal or markers ("uninstall would either leave residue or delete user content"), and the broker owning memory capture.

## 2. LazyAI 1.x today

### 2.1 Declared contract (spec 029 / ADR-006)

- ADR-006 makes `.ai/lock.json` "the authoritative generated-output lockfile for idempotency, drift-aware writes, adapter metadata, and output hashes" (`specs/adrs/006-manifest-driven-compile-and-seven-target-contract.md:33`).
- Spec 029 sets these requirements: FR-003 says the lock records `sourceHash`, `outputHash` and `managed` per output. FR-004 says to skip outputs whose hashes match. FR-005 says to preserve content outside managed regions (`specs/029-lazyai-v2/spec.md:85-87`). EC-003 says to refuse to overwrite drifted outputs without `--force` (`spec.md:131`). FR-016 says `eject` keeps native files and removes metadata (`spec.md:98`).
- Plan 029 names a `discover → … → plan → write` pipeline with `plan.Build` and `writer.Apply` (`specs/029-lazyai-v2/plan.md:107-108`). The human gate is recorded as "pending — implementation blocked until approved" (`plan.md:173`).

### 2.2 What the Go code does

**Lockfile schema.** `Lock{version, lazyaiVersion, compiledAt, adapters, generated[]}`, where `Generated{path, target, sourceHash, outputHash, managed}` (`packages/cli/internal/lockfile/lockfile.go:16-40`). Hashes are `sha256:` hex (`lockfile.go:99-102`). Save sorts entries and writes through `files.SafeWriteFile` (`lockfile.go:69-95`). The lock has no install or remove history, no operation IDs and no backup pointers.

**Who writes the lock in production.** Only `writeCompileLock` in `packages/cli/cmd/compile.go:390-439`, which `compile` calls (`compile.go:341`).
- It records every compiled record with `Target: "mcp"`, `Managed: false`, and `SourceHash` set to the hash of the whole MCP source file (`compile.go:395, 427-433`).
- It hashes whatever is on disk *after* writing, and skips unreadable files (`compile.go:423-426`).
- It never removes entries. `Upsert` only adds or replaces (`lockfile.go:116-124`).
- `compile` registers only the `--dry-run` flag (`compile.go:37`) and has no `--force`. In dry-run mode it does not compile at all; it only prints "Would compile" (`compile.go:261-266`).

**Plan/writer pipeline (drift-aware) exists but is not wired in.**
- `plan.Build` classifies each output as Create, Update, Skip or Drift. For untracked existing whole-file outputs and for whole-file outputs edited on disk, it returns Drift. For managed outputs it re-merges ("preserving out-of-region edits") (`packages/cli/internal/plan/plan.go:58-95`).
- `writer.Apply` refuses Drift unless `Force` is set and writes atomically with temp file, fsync and rename (`packages/cli/internal/writer/writer.go:49-114, 139-165`).
- Only `internal/writer` (and its test) imports `internal/plan`, and no package outside its own test imports `internal/writer`. `grep -rln 'internal/plan"\|internal/writer"'` returns only `writer.go` and `writer_test.go`.

**Managed markers.** There are two unrelated marker families.
1. **Region markers** `<!-- lazyai:managed:start root-agents v1 -->` / `…:end…` (`packages/cli/internal/adapter/shared.go:12-13`).
   - `MergeManagedBlock` replaces the text between the markers, or appends a block if the markers are missing (`internal/adapter/shared_copy.go:20-63`).
   - Its only production callers are the unwired `writer.writeOne` (`writer.go:128`) and `WriteManagedBlockToFile` (`shared_copy.go:68-88`), which has no callers.
   - The region markers are therefore not emitted by any wired compile path today.
2. **Whole-file ownership marker** `<!-- vibe-lab:managed kind=agent surface=… name=… source=.agents/agents/<name>.md -->`. The Go side (`shared.go:16-19`) states it "must match bin/inject exactly".
   - It is written into Claude, OpenCode, Copilot and OMP agent files (`agent_transform.go:136, 213, 609`, `copilot.go:327`, `shared_default_agent.go:17, 38`).
   - It is read in one place: `isLikelyLegacySkillAgentFile` treats a marker match as permission to delete a legacy Copilot skill-agent file (`copilot.go:457-460`, deletion at `copilot.go:451`).

**Legacy store (the actual ownership ledger today).**
- `TrackedFile{path, hash, source, owner, status, …}` lives in `.ai-setup.db` (SQLite) or `.ai-setup.json` (`internal/types/types.go:537-547`, `cmd/status.go:200-227`, `internal/db/db.go:86-88`, `internal/manifest/manifest.go:16`).
- `owner` is one of `library | user | migrated` (`types.go:332-338`).
- Adapters stamp most outputs `library`. Some MCP outputs are stamped `user`: OMP, Antigravity and Codex configs (`internal/adapter/mcp_compiler.go:512, 543, 572, 602`).
- `compile` appends new records to the store after writing the lock (`compile.go:347-352`).
- `StoreData.Operations []Operation{id, type, timestamp, filesAffected, result, backupPaths}` looks like a journal (`types.go:623-641`). However, `OperationTracker.ToOperation`/`NewOperationTracker` (`internal/error/operation.go:29, 77`) have no production callers.

**Overwrite protection on install/compile.**
- `CopyWithRecord` calls `conflict.ResolveConflictWithOptions` without a `TrackedHash` (`internal/adapter/shared_copy.go:112-115`).
- For an existing file that function returns Skip by default and Replace on `--force` or the `backup-and-replace`/`align` strategies. It returns Replace only when the on-disk hash equals a supplied tracked hash (`internal/conflict/strategy.go:139-172`).
- A Replace first copies the old file to `.ai-setup-backup/` (`shared_copy.go:128-132`, `internal/files/files.go:442-448`).
- Context docs (for example `GEMINI.md`) are appended to idempotently by checking for an import token, never overwritten (`internal/scaffold/root.go:314-336`).

**Drift detection (`lazyai-cli doctor`).**
- `doctor` hashes every store `TrackedFile` and marks it `missing` or `modified` (`cmd/doctor.go:69-91`).
- It reads the legacy store, **not** `.ai/lock.json`. `lockfile.` has no references in `cmd/` outside `compile.go`.

**`bin/inject` (repo-maintainer path).**
- It embeds `<!-- Source hash: … -->` in generated context files. The hash is a sha256 over sorted `canonical/*.md`, `.agents/{skills,agents,hooks,workflows}` sources (`bin/inject:134-160, 425-437`).
- It embeds `# Policy hash:` in hook scripts (`bin/inject:163-165`, e.g. `:639`).
- `bin/doctor` compares the embedded hashes with recomputed ones and warns "stale — run bin/inject" (`bin/doctor:571-594, 817-893`). These are *source-freshness* hashes, not output hashes. Hand edits to generated output are not detected by them.
- `ensure_agent_adapter_slot` refuses to overwrite an existing agent file unless it contains the `vibe-lab:managed` marker or exactly equals the legacy content ("Refusing to overwrite user-owned agent adapter") (`bin/inject:266-282`).
- Pruning removes skill adapters and docs whose source no longer exists. It checks neither ownership nor hashes (`bin/inject:489-519`).

**Uninstall / eject.**
- `lazyai-cli eject` deletes only metadata: `.ai/lazyai.json`, `.ai/lock.json`, `.ai/migration-report.md`, `.ai-setup.json` and `.ai-setup.db`. It leaves every generated native file in place (`internal/eject/eject.go:8-45`, `cmd/eject.go:18-94`).
- Go code has no `uninstall` command. `grep -rni uninstall --include=*.go` (non-test) finds nothing.
- The closest thing to "remove only unchanged owned content" is `removeLegacyAgents` in `update`. It deletes a legacy `.opencode/agents/*` file only if the file is tracked, has `Owner == library`, and its current hash equals the tracked hash (`cmd/update.go:221-265`).

## 3. Comparison

| Dimension | ADR-013 (proposed, unimplemented) | LazyAI 1.x (as coded) |
|---|---|---|
| **Marking** | Ownership markers on broker-added content: marked guidance blocks plus owned assets. Key-level ownership in shared config (`mcp_servers.context-broker`). | `vibe-lab:managed` whole-file marker in agent files (Go and `bin/inject`). `lazyai:managed:start/end` region markers are defined, but no wired path emits them. MCP/JSON config outputs carry no in-file marker. Ownership of a whole file is recorded out of band (`TrackedFile.owner`). |
| **Journal / lock** | Durable journal plus manifest, with crash-recovery paths. | `.ai/lock.json` is a current-state snapshot (no history, entries never pruned) that holds only compile/MCP outputs, all `managed:false`. The legacy store holds `TrackedFile` records. `Operations` journal type exists but nothing writes it. Atomic temp+rename exists only in the unwired writer. |
| **Drift detection** | Install-time conflict detection. Modified owned content is flagged for reconciliation. | `doctor` reports `modified`/`missing` against store hashes. `.ai/lock.json` is written but never read for drift. `plan.Build` drift classification is unwired. `bin/doctor` checks source-hash freshness, not output edits. |
| **User-edit preservation** | Owned content that the user changed is preserved and flagged. Non-owned content in shared files stays untouched. | Existing files are skipped by default. `--force`/strategy replaces them after a `.ai-setup-backup/` copy. Context docs are append-only. `bin/inject` refuses to overwrite unmarked agent files. Managed-region re-merge (FR-005) is implemented but unwired. |
| **Uninstall** | Removes only owned content that is unchanged, and is reversible. | No uninstall. `eject` deletes metadata and keeps all generated files. The only hash-gated, owner-gated deletion is `update`'s `removeLegacyAgents`. `bin/inject` prunes stale skill adapters without ownership or hash checks. |

## 4. Gaps

### In LazyAI 1.x, relative to ADR-013

1. No uninstall that removes owned, unchanged generated content. `eject` removes only metadata (`eject.go:8-14`).
2. No journal. `Operations`/`OperationTracker` are dead code (`operation.go:29, 77`). There are no crash-recovery or rollback records beyond `.ai-setup-backup/` copies.
3. Two ownership ledgers that are not reconciled. `doctor` reads the store and ignores `.ai/lock.json`, which ADR-006 calls authoritative (`doctor.go:58-91` vs `adrs/006…:33`).
4. The lock under-describes outputs. Every entry is `target:"mcp"` and `managed:false`, the source hash is the whole MCP file's hash, and stale entries are never removed (`compile.go:395, 427-433`, `lockfile.go:116-124`).
5. The drift refusal (EC-003) and managed-region preservation (FR-005) exist only in the unwired `plan`/`writer` packages. `compile` has no `--force`.
6. No in-file ownership marking for JSON/TOML/MCP config outputs. There is no key-level ownership of a shared config file comparable to ADR-013's `mcp_servers.context-broker`.
7. No preview-by-default. `compile --dry-run` is opt-in and does not compute a diff (`compile.go:261-266`).
8. `bin/inject` pruning deletes adapters without an ownership or hash check (`bin/inject:489-519`).

### In ADR-013, relative to LazyAI 1.x

1. It is not implemented, its status is Proposed, and marker tolerance per harness is listed as unverified (ADR lines 3-5, 54-57). LazyAI ships working markers, lock hashing, backups and drift reporting.
2. It does not define a hash scheme, a lock/manifest schema, or how an "unchanged" state is determined. LazyAI defines both `sourceHash` and `outputHash` (`lockfile.go:16-22`) plus `TrackedFile.hash`.
3. It defines no ownership classes for imported or adopted content. LazyAI has `library | user | migrated` (`types.go:336-338`) and a plan-level "adopt" path for untracked managed files (`plan.go:72-73`).
4. It has no notion of whole-file versus region ownership. LazyAI distinguishes whole-file outputs (`managed:false`) from managed-region outputs (`plan.go:30-33`).
5. It does not cover multi-repo/workspace propagation. LazyAI records propagated per-repo outputs under `<repoPath>/<path>` (`compile.go:403-422`).
6. It has no eject mode that keeps generated files. Its only exit path is uninstall.
7. It says nothing about source-freshness checks, the equivalent of `bin/inject`'s embedded source and policy hashes plus `bin/doctor` stale warnings.
