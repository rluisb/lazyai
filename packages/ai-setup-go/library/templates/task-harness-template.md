# Task Harness: T### — [Task Name]

**Task ID:** T###
**Phase:** [N — phase name from `tasks.md`]
**User story:** [US1 / US2 / US3 — link to `spec.md`]
**Spec FRs covered:** FR-###, FR-###
**Plan reference:** [`plan.md#section`](./../plan.md#section)
**Status:** ☐ TODO · 🔄 IN PROGRESS · ✅ DONE · ⛔ BLOCKED
**Lifecycle label:** `loading_context` / `planning` / `awaiting_approval` / `executing` / `verifying` / `blocked` / `handoff` / `done` / `error`
**Depends on:** T###, T###
**Parallel with:** T### *(only if `[P]` in tasks.md)*

> **Purpose.** This is the **only** document the implementing agent reads at execution time. It snapshots the state of the world (tools, paths, patterns, permissions) so the agent does not improvise constraints. Harness in → code + tests out.

---

## 1. Objective

[One clear sentence: what this task accomplishes. No more, no less.]

---

## 2. Acceptance Criteria

Verifiable, observable outcomes. The task is **DONE** only when every box is checked with evidence.

- [ ] **AC-1:** [criterion]. Evidence: [test path or observable].
- [ ] **AC-2:** [criterion]. Evidence: [test path or observable].
- [ ] **AC-3:** [criterion]. Evidence: [test path or observable].

---

## 3. Environment Snapshot

The exact world the agent operates in. Treat as read-only — do not upgrade or mutate without an ADR.

| Aspect | Value | Notes |
|---|---|---|
| Repo | [repo name] | [planning / code / workspace] |
| Branch | [branch name] | created from [base] |
| Working directory | [absolute path] | |
| Language | [e.g., Go 1.26 / TypeScript 5.x] | |
| Package manager | [e.g., npm / pnpm / go mod] | |
| Test runner | [e.g., go test / vitest / rspec] | |
| Linter / formatter | [e.g., go vet / eslint / rubocop] | |
| Relevant env vars | [LIST_OR_NONE] | redact secrets |

---

## 4. Files in Play

| Path | Action | Why |
|---|---|---|
| `[path]` | create / modify / delete | [reason] |
| `[path]` | create / modify / delete | [reason] |

**Out of bounds (must NOT touch):**
- [path or pattern] — [reason: out of scope / different task / different repo]

---

## 5. Patterns to Follow

References to existing code or standards the implementation MUST mirror. Article I (Library-First) and Article IV (Pattern Consistency) live here.

| Pattern | Reference | Notes |
|---|---|---|
| [pattern name] | [file:line or `specs/standards/...`] | [what to copy / adapt] |
| [pattern name] | [file:line] | [notes] |

> If no existing pattern fits, **stop** and escalate. Inventing a novel pattern requires an ADR — not a task-level decision.

---

## 6. Testing Strategy (Article II — Test-First)

The test plan **before** the production code is written.

| Test | Type | Path | Initial state |
|---|---|---|---|
| [test name] | unit / integration / e2e | [path] | RED — fails on unimplemented code |
| [test name] | unit / integration / e2e | [path] | RED |

**TDD loop:**
1. **RED** — write the tests above; verify they fail for the right reason.
2. **GREEN** — write the smallest production code that turns them green.
3. **REFACTOR** — improve names, structure, duplication; keep the suite green.
4. **GATE** — run the 5-gate ladder (§ 8) and record evidence.
5. **COMMIT** — small, conventional message scoped to this task.

---

## 7. Permissions & Side Effects

What the agent is allowed to do during this task.

- [ ] May read: [paths or "anywhere in this repo"].
- [ ] May write: [paths]. **Default deny outside the list.**
- [ ] May run: [commands]. **Default deny everything else.**
- [ ] May install dependencies: YES / NO. *(NO unless declared in `plan.md`.)*
- [ ] May call network: YES / NO. *(Test fixtures must not require live network.)*
- [ ] May modify other repos: NO. *(Cross-repo work goes through the orchestrator.)*

---

## 8. Quality Gates (5-Gate Ladder)

Every gate runs locally before commit. Record evidence inline.

- [ ] **Gate 1 — Static Integrity:** lint, format, type-check. Evidence: `[command output / OK]`.
- [ ] **Gate 2 — Contract Compliance:** public API matches plan; no speculative options (Article IV); no extraction below 3 instances (Article VI).
- [ ] **Gate 3 — Behavioral Validation:** all tests green; new ACs covered; coverage threshold met.
- [ ] **Gate 4 — Pattern Consistency:** uses existing helpers; naming follows convention; no novel pattern without ADR.
- [ ] **Gate 5 — Observability Readiness:** logging / metrics present where applicable; rollback path noted if risky.

---

## 9. Anti-Overengineering Audit (Article VI)

Self-audit before the COMMIT step.

- [ ] No speculative parameters or options.
- [ ] No abstractions added with one caller.
- [ ] No helper extracted below the third instance.
- [ ] No function exceeds 30 lines without a documented reason.
- [ ] No file exceeds 300 lines without a documented reason.
- [ ] No `if/else` for impossible branches; trust internal callers.

---

## 10. Memory Update

On completion, update workspace memory.

- [ ] Append to `.specify/memory/repos/<repo>/ledger.md` (date, who, what, plan ref, verified).
- [ ] Rewrite `.specify/memory/repos/<repo>/last-known-state.md` with current branch + dirty-files state.

---

## 11. Notes / Pivots

*(Optional — fill only if the task deviated from plan.)*

- **Pivot:** [date] — [what changed and why]. ADR: [link if architecture-affecting].

### Lifecycle reporting

Use the **Lifecycle label** as status reports, handoff evidence, recovery summaries, and final completion evidence are updated:

- `loading_context` while reading this harness and required context.
- `planning` while defining touch map, risks, assumptions, and verification matrix.
- `awaiting_approval` when a human decision is required before continuing.
- `executing` while performing approved task actions.
- `verifying` while running checks and collecting evidence.
- `blocked` when safe progress is impossible without new information, tools, or approval.
- `handoff` when preparing resumable context for another agent/session.
- `done` only when the verdict evidence satisfies every acceptance criterion.
- `error` when a command, test, tool, or validation failure is being reported.

---

## 12. Verdict

```
DONE / BLOCKED — [evidence]
Implementor: [name or agent]
Reviewer:    [name or agent]
Date:        YYYY-MM-DD
Commit(s):   [SHA, SHA, …]
```
