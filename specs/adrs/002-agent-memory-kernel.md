# ADR-002: Agent Memory Kernel — Architecture Decisions

**Date:** 2026-05-11
**Status:** Proposed
**Deciders:** lazyai maintainers
**Constitution:** [`.specify/memory/constitution.md`](../../.specify/memory/constitution.md)

> **Purpose.** Answer the 15 open questions from issue #202 before any implementation begins. This ADR defines the architectural boundaries, storage strategy, and Phase 1 scope for the Agent Memory Kernel.

---

## Context

Issue #202 proposes a durable memory module for long-running agent workflows. The PRD and tech spec are located in `docs/lazyai-agent-memory-docs-feature/`. The proposal identifies 15 open questions that block implementation. This ADR answers them.

The orchestrator already maintains a SQLite database with 12 migrations covering `chain_runs`, `team_runs`, `workflow_runs`, `handoffs`, `run_events`, `error_journal`, `execution_plans`, `definitions`, `definition_versions`, and `queue_jobs`. The Agent Memory Kernel must decide its relationship to this existing store.

---

## Constitution Alignment

| Article | Bearing | Note |
|---|---|---|
| I — Library-First | bears | Memory module is a self-contained Go package with embedded schema. |
| IV — Anti-Speculation | bears | Phase 1 is bounded to deterministic continuity only. Vector search, MCP exposure, and QMD integration are explicitly deferred. |
| V — Simplicity | bears | Separate DB, separate module. No shared tables with orchestrator. Clear ownership boundary. |
| VI — Anti-Overengineering | bears | No compaction, retention, or summarization in Phase 1. Add only what Phase 1 requires. |

This ADR does not amend the constitution.

---

## Decisions: 15 Open Questions Answered

### Q1: Ownership — Is `agentmemory` the canonical store or an adapter over existing orchestrator tables?

**Decision: Standalone canonical store, separate DB.**

`agentmemory` is its own SQLite database at `.ai/agent-memory/agent-memory.sqlite`. It does **not** share tables with the orchestrator's `orchestrator.db`. The orchestrator's `handoffs` and `run_events` tables serve the orchestrator's lifecycle needs (chain/team/workflow state transitions). The Agent Memory Kernel serves agent-level continuity (task state, checkpoints, memories, artifacts).

**Rationale:**
- Separate DB = separate lifecycle, separate backup, separate corruption domain.
- Orchestrator tables track *execution* state (which chain is running, which step failed). Memory tables track *agent* state (what did the agent decide, what checkpoint can it resume from).
- Future integration: the orchestrator can write events to both stores, or a thin adapter can sync run IDs. But the stores remain independent.
- If we share tables now, we couple two modules with different evolution rates and different consistency requirements.

**Consequence:** `agentmemory` has its own migration engine (matching the orchestrator's hand-rolled style with inline Go constants). Run IDs can be correlated via a shared `run_id` foreign key convention, but no FK constraint across databases.

---

### Q2: Package name — Do we commit to `packages/agentmemory`?

**Decision: Yes.**

`packages/agentmemory` is the module path. Import path: `github.com/rluisb/lazyai/packages/agentmemory`.

**Rationale:** `memory` is too generic and conflicts with existing concepts (workspace memory, agent memory, semantic memory). `agentmemory` is specific and unambiguous.

---

### Q3: Runtime integration point — Which orchestrator/run lifecycle boundaries should call memory APIs?

**Decision: Defer to Phase 2.**

Phase 1 is library-only. The memory module exposes a Go API. Wiring it into the orchestrator's run lifecycle is Phase 2 work. Phase 1 success is measured by: the module exists, tests pass, a test program can write/read task state, checkpoints, and memories.

**Rationale:** Integration decisions require understanding the orchestrator's event bus and run boundaries. Doing this in Phase 1 expands scope beyond deterministic continuity.

---

### Q4: DB location — `.ai/agent-memory/` vs `.lazyai/memory/`?

**Decision: `.ai/agent-memory/agent-memory.sqlite`.**

Matches the orchestrator convention (`.ai/orchestrator.db`). The `.ai/` directory is the canonical LazyAI data directory.

**Rationale:** Consistency. Users and tooling already expect `.ai/` as the data root. Auto-generate `.gitignore` in `.ai/agent-memory/` to prevent accidental commits.

---

### Q5: Namespace model — Per workspace, project, user, task, or all?

**Decision: Per-project namespace in Phase 1.**

All tables include a `namespace TEXT NOT NULL` column. Phase 1 uses the project root path (or a derived project ID) as the namespace value. Workspace-level and user-level namespaces are future work.

**Rationale:** Project is the natural isolation boundary for agent work. Workspace-level namespaces require understanding multi-repo layouts. User-level namespaces require auth context. Both are out of scope for Phase 1.

---

### Q6: Phase 1 acceptance — What is the first success metric?

**Decision: Resumability + handoff continuity.**

Phase 1 is successful when:
1. A test program can write a task, checkpoint, and handoff to the memory store.
2. A second test program (simulating a resumed agent) can read the task, latest checkpoint, and unconsumed handoff.
3. FTS5 lexical search returns correct results for stored memories and artifacts.
4. Context builder assembles a deterministic prompt context from: task + latest checkpoint + active handoff + recent events + FTS results.

Retrieval quality (semantic recall, reranking) is Phase 3.

---

### Q7: Vector feasibility — Does `modernc.org/sqlite` support `sqlite-vector`?

**Decision: No. Vector is deferred to Phase 3 as opt-in CGO build.**

`modernc.org/sqlite` is a pure-Go SQLite implementation compiled from C via GopherJS. It does **not** support loading native extensions (`.so`/`.dylib`). `sqlite-vector` is a native C extension.

**Phase 1:** FTS5-only. `modernc.org/sqlite` supports FTS5 out of the box. This is sufficient for lexical recall.

**Phase 3:** If vector search is required, the build must switch to `github.com/mattn/go-sqlite3` (CGO) for the vector-enabled build. The module will expose a build tag `sqlite_vector` that gates vector functionality. Default build (no tag) uses `modernc.org/sqlite` with FTS5-only.

**Rationale:** Pure-Go cross-compilation is a hard requirement for the LazyAI CLI. Vector search is an enhancement, not a core requirement. FTS5 provides adequate recall for Phase 1 and Phase 2.

---

### Q8: Schema compatibility — Should table names be prefixed?

**Decision: No prefix needed.**

Tables live in a separate database file (`agent-memory.sqlite`) from the orchestrator (`orchestrator.db`). No collision risk. Table names are: `tasks`, `task_events`, `checkpoints`, `artifacts`, `memories`.

**Rationale:** Prefixes add verbosity without solving a real problem when databases are separate.

---

### Q9: Redaction boundary — Enforced inside memory stores, at runtime call sites, or both?

**Decision: Both, with store-level enforcement as the canonical boundary.**

The memory store's `Write` methods apply redaction before persisting. This is the canonical enforcement point. Runtime call sites should also redact before calling the store as a defense-in-depth measure, but the store must not trust callers.

**Rationale:** Store-level enforcement guarantees that no unredacted data reaches disk, regardless of caller behavior. Runtime redaction catches bugs early.

**Redaction rules (Phase 1 minimum):**
- API keys (pattern: `sk-*`, `Bearer *`, `Authorization: *`)
- Private keys (pattern: `-----BEGIN * PRIVATE KEY-----`)
- Environment variable values matching `*_KEY`, `*_SECRET`, `*_TOKEN`
- JSON redaction must maintain valid JSON structure after traversal

---

### Q10: Artifact storage — Full content, paths only, or both?

**Decision: Paths + bounded content preview (max 4 KB).**

Artifacts store:
- `path TEXT` — file path or URI
- `content_preview TEXT` — first 4 KB of content (redacted)
- `size_bytes INTEGER` — full content size
- `content_hash TEXT` — SHA256 of full content

Full content is not stored in the database. The path points to the actual file on disk.

**Rationale:** Storing full content in SQLite causes unbounded database growth. 4 KB preview is sufficient for context assembly and FTS5 indexing. SHA256 hash enables integrity verification.

---

### Q11: CLI scope — Library-only for Phase 1, or CLI inspection commands too?

**Decision: Library-only for Phase 1.**

Phase 1 exposes a Go API. CLI inspection commands (`lazyai-cli memory inspect`, `lazyai-cli memory list`) are Phase 2.

**Rationale:** CLI commands require understanding the user-facing information model. Phase 1 should prove the storage and retrieval mechanics first.

---

### Q12: Migration policy — Match existing hand-rolled style exactly?

**Decision: Yes.**

Match the orchestrator's migration pattern:
- Inline Go constants in `migrations.go`
- `001_` prefix numbering
- Each migration is a `Migration{ID, SQL}` struct
- `RunMigrations()` applies pending migrations in order
- Multi-statement migrations wrapped in transactions

**Rationale:** Consistency across the codebase. No new migration framework to learn or maintain.

---

### Q13: Embedding provider — Which provider, offline/local required?

**Decision: Deferred to Phase 3.**

Phase 1 has no embedding. Phase 3 will define an `EmbeddingProvider` interface with at least two implementations:
- OpenAI-compatible (remote API)
- Local/offline (TBD — depends on `sqlite-vector` feasibility)

**Rationale:** Embedding is a Phase 3 concern. Defining the interface now is premature without knowing the vector extension constraints.

---

### Q14: Retention defaults — What kept indefinitely, what pruned?

**Decision: Deferred to Phase 4.**

Phase 1 has no retention or pruning. All data is retained indefinitely. Retention controls are a Phase 4 feature.

**Rationale:** Retention requires understanding usage patterns and storage growth rates. Implementing retention before understanding real-world data volume is speculative.

---

### Q15: Security posture — Encryption-at-rest out of scope or required?

**Decision: Out of scope for Phase 1–3. File permissions are the Phase 1 minimum.**

Phase 1 security:
- Database directory: `0700` permissions
- Database file: `0600` permissions
- Auto-generate `.gitignore` in data directory
- FTS5 input sanitization (escape FTS5 reserved characters)
- Size limits on `state_json`, `payload_json`, `content` to prevent disk exhaustion

Encryption-at-rest is a future consideration if the memory store contains project-sensitive data that warrants it.

**Rationale:** Application-level encryption adds significant complexity. File permissions + `.gitignore` prevent the most likely accidental exposure (committed database file). Encryption can be added later if threat modeling requires it.

---

## Options Considered

### Option A — Standalone module, separate DB *(chosen)*

- **Summary:** `packages/agentmemory` with its own SQLite database at `.ai/agent-memory/`.
- **Complexity:** Medium. New module, new DB, new migrations.
- **Reversibility:** High. Module can be removed without affecting orchestrator.
- **Constitution fit:** Strong fit for Simplicity and Anti-Overengineering.

### Option B — Extend orchestrator DB with new tables

- **Summary:** Add memory tables to `orchestrator.db` within the existing orchestrator package.
- **Complexity:** Low initially, high long-term.
- **Reversibility:** Low. Coupling two modules with different evolution rates.
- **Constitution fit:** Weak fit. Violates separation of concerns.

### Option C — Adapter over existing orchestrator tables

- **Summary:** `agentmemory` reads/writes orchestrator's `handoffs` and `run_events` tables.
- **Complexity:** Medium. Requires schema changes to orchestrator tables.
- **Reversibility:** Medium.
- **Constitution fit:** Moderate. Reuses existing data but couples memory semantics to orchestrator lifecycle.

**Why B and C were rejected:**
- Option B couples agent memory to orchestrator lifecycle. If the orchestrator changes its schema or migration strategy, the memory module breaks.
- Option C assumes orchestrator tables are sufficient for agent memory needs. They are not: orchestrator tracks execution state, not agent decision state, checkpoints, or semantic memories.

---

## Phase 1 Scope Summary

| Included | Excluded |
|----------|----------|
| `packages/agentmemory` Go module | `sqlite-vector` (Phase 3) |
| SQLite open/config/migration (hand-rolled style) | QMD integration (Phase 5) |
| Tables: tasks, task_events, checkpoints, artifacts, memories | MCP exposure (Phase 5) |
| FTS5 lexical search | Compaction/summarization (Phase 4) |
| Minimal context builder | Reranking (Phase 3) |
| Redaction (API keys, private keys, env patterns) | Retention/pruning (Phase 4) |
| File permissions (0700/0600) + `.gitignore` | Encryption-at-rest |
| Namespace: per-project | Workspace/user namespaces |
| Pure Go (`modernc.org/sqlite`) | CGO builds |
| Library-only API | CLI commands |
| Tests: CGO_ENABLED=0, FTS-only | Vector-enabled tests |

---

## Consequences

**Positive:**
- Clear ownership boundary between orchestrator and memory.
- Pure-Go cross-compilation preserved for Phase 1 and Phase 2.
- Phase 1 is bounded and testable.
- FTS5 provides adequate recall for deterministic continuity.

**Negative / accepted trade-offs:**
- Two SQLite databases to manage (orchestrator.db + agent-memory.sqlite).
- No semantic recall until Phase 3.
- No CLI inspection until Phase 2.
- No retention controls until Phase 4.

**Neutral:**
- Run IDs can be correlated across databases via shared convention, but no FK constraints.
- Embedding provider interface is undefined until Phase 3.

---

## Reversal Conditions

Re-open this ADR if any of the following become true:

- FTS5 lexical recall proves insufficient for agent continuity (e.g., agents cannot find relevant prior context without semantic search).
- The orchestrator's event model changes in a way that makes shared tables more practical than separate databases.
- `modernc.org/sqlite` gains native extension support, making `sqlite-vector` available in pure-Go builds.
- Project-sensitive memory data requires encryption-at-rest before Phase 4.

---

## Implementation Pointer

Planned implementation sequence (Phase 1):

1. Create `packages/agentmemory` module with `go.mod` using `modernc.org/sqlite`.
2. Implement `DB` struct with open, close, and migration (hand-rolled style matching orchestrator).
3. Create tables: `tasks`, `task_events`, `checkpoints`, `artifacts`, `memories` with FTS5 virtual tables.
4. Implement store structs: `TaskStore`, `CheckpointStore`, `HandoffStore`, `ArtifactStore`, `MemoryStore`.
5. Implement FTS5 lexical search over memories and artifacts.
6. Implement minimal context builder: task + latest checkpoint + active handoff + recent events + FTS results.
7. Implement redaction layer.
8. Implement file permissions and `.gitignore` generation.
9. Write tests: CGO_ENABLED=0, FTS-only, namespace isolation, redaction, FTS5 edge cases.

- Plan: follow-up implementation plan/spec required before code changes.
- PRs: pending.
- Standards updated: pending.

---

## Assumptions

| Assumption | Status |
|---|---|
| `modernc.org/sqlite` supports FTS5 out of the box. | Accepted (verified in documentation) |
| `modernc.org/sqlite` does not support native extensions. | Accepted (verified in documentation) |
| `.ai/` directory is the canonical LazyAI data directory. | Accepted (matches orchestrator convention) |
| Project root path is a suitable namespace identifier for Phase 1. | Accepted |
| 4 KB content preview is sufficient for context assembly. | Accepted (revisitable in Phase 2) |
| Redaction patterns cover the most common secret formats. | Accepted (expandable in Phase 2) |

---

## Risks

| Risk | Level | Mitigation |
|---|---|---|
| FTS5 insufficient for agent continuity. | Medium | Phase 3 adds vector search; monitor Phase 1/2 usage. |
| Database growth without retention controls. | Medium | Phase 4 adds retention; monitor disk usage in Phase 1/2. |
| `modernc.org/sqlite` performance vs. `go-sqlite3`. | Low | Benchmark in Phase 1; switch if unacceptable. |
| Namespace collision if project root is not unique. | Low | Use SHA256 of project root path as namespace if needed. |
| Redaction misses secret formats. | Medium | Add patterns as discovered; log redaction failures. |

---

## Ready for Planning

Ready for implementation planning once maintainers confirm this ADR. Phase 1 scope is bounded, testable, and does not depend on unresolved questions.
