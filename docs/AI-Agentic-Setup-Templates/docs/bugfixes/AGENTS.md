<rule>
  <scope>auto</scope>
  <globs>docs/bugfixes/**</globs>
  <description>Bugfix workflow — two paths based on severity: HOTFIX (P0/P1) and SCHEDULED (P2/P3)</description>
</rule>

# Bugfix Workflow Rules

## Severity Triage — Human Decision First

⛔ **HUMAN GATE:** Before any work begins, classify severity:

| Severity | Description | Path |
|----------|-------------|------|
| **P0** | Production is down or data is corrupted | HOTFIX path ↓ |
| **P1** | Major feature broken, significant user impact | HOTFIX path ↓ |
| **P2** | Feature degraded, workaround exists | SCHEDULED path ↓ |
| **P3** | Minor impact, cosmetic, edge case | SCHEDULED path ↓ |

The human classifies severity. The AI does not override this classification.

---

## HOTFIX Path (P0 / P1)

For production emergencies. Speed is prioritized. Gates are compressed but not skipped.

### Directory Structure

```
docs/bugfixes/NNN-bug-name/
├── techspec.md          ← Abbreviated RCA (use bugfix-rca-template.md)
└── progress.md          ← Trace log + CR log
```

### Flow

1. **Reproduce** — confirm bug is reproducible before writing any code
   - ⛔ HUMAN GATE: confirm reproduction before proceeding
2. **RCA** (Scout + Planner) — root cause + blast radius → abbreviated techspec.md
   - Use `docs/templates/bugfix-rca-template.md`
   - ⛔ HUMAN GATE: confirm root cause + fix scope
3. **Implement** (Builder) — minimal fix + regression test
   - ⛔ CHECKPOINT: regression test passes, existing tests pass
4. **Expedited Review** (Reviewer) — single-pass, focused on fix correctness
5. **Post-mortem** — MANDATORY within 24h after deploy
   - Use `docs/templates/postmortem-template.md`
   - Output: `docs/bugfixes/NNN-bug-name/postmortem.md`

### Hotfix Principles
- Every P0/P1 gets a post-mortem. No exceptions.
- Fix the production issue first. Refactor later.
- Minimum viable fix — revert-friendly if it goes wrong.

---

## SCHEDULED Path (P2 / P3)

For non-critical bugs that go through the normal sprint cycle.

### Directory Structure

```
docs/bugfixes/NNN-bug-name/
├── research.md          ← Investigate the bug (Scout)
├── techspec.md          ← RCA + fix approach (Planner)
├── tasks/               ← Only if fix is > 20 lines or > 3 files
│   ├── tasks.md
│   └── 001-name.md
└── progress.md          ← Trace log + CR log
```

### Flow

1. **Research** (Scout) — reproduce bug, map affected code → research.md
   - ⛔ HUMAN GATE: confirm root cause before planning fix
2. **RCA / TechSpec** (Planner) — root cause + fix approach → techspec.md
   - Use `docs/templates/bugfix-rca-template.md`
   - **No PRD needed for bugfixes**
   - ⛔ HUMAN GATE: approve fix approach
3. **Tasks** — only if fix touches > 20 lines or > 3 files
   - Small fixes: implement directly from techspec
4. **Implement** (Builder) — fix the bug, add regression test
   - ⛔ CHECKPOINT: regression test passes, all tests pass
5. **Review** (Reviewer) — standard review pass

---

## Principles (Both Paths)

- **Fix the bug, nothing else.** No drive-by refactors.
- **Always add a regression test.** No regression test = fix is incomplete.
- **Reproduce before fixing.** A bug you cannot reproduce is one you do not understand.
- Root cause must be specific — "bad code" is not a root cause.

---

## Self-Improvement — After Every Bugfix

Before ending the session, run the Impact Check from root AGENTS.md.

Additionally for bugfixes:
- Bug caused by missing rule? → Flag `docs/rules/` update
- Bug caused by unclear standard? → Flag `docs/standards/` improvement
- Regression test uses new testing pattern? → Flag `docs/standards/testing/` update
- Bug reveals architectural weakness? → Consider ADR + `docs/standards/architecture/` update
- Write a memory note to `docs/memory/` explaining what went wrong and how to prevent it
