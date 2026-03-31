<rule>
  <scope>auto</scope>
  <globs>docs/refactors/**</globs>
  <description>Refactor workflow — MICRO (<50 lines) and MACRO (≥50 lines) paths. ADR mandatory for MACRO.</description>
</rule>

# Refactor Workflow Rules

## Size Gate — Run This First

Before any other step, count the lines of production code that will change.

| Size | Threshold | Path |
|---|---|---|
| **MICRO** | < 50 lines changed | Lightweight — no PRD/TechSpec |
| **MACRO** | ≥ 50 lines changed | Full RPI + ADR + Compatibility Matrix |

> If you are unsure, default to MACRO. You can always descope; you cannot undo a skipped ADR.

---

## MICRO Path (< 50 lines)

Use for: rename, extract method, clean up dead code, fix naming, improve readability.  
Does NOT require PRD, TechSpec, or dedicated tasks file.

```
docs/refactors/NNN-refactor-name/
└── progress.md    ← intent + session log
```

### Steps

1. **Document intent** — Add a `## Refactor: [name]` entry to `progress.md`.
   - What is being changed and why (1–3 sentences)
   - What must NOT change (behavior, interface, API contract)
2. **Implement** — Apply the change in one focused commit.
3. **Review** — Run tests. Confirm behavior unchanged. No new test debt introduced.
4. **Impact check** — Grep for all callers of renamed/moved symbols. Update references.
5. **Log** — Append result and any standards updated to `progress.md`.

### Constraints
- Behavior must not change. If behavior changes, this is a feature — not a refactor.
- Tests must pass before and after.
- If you discover the scope is larger than 50 lines mid-way: stop, escalate to MACRO path.

---

## MACRO Path (≥ 50 lines)

Use for: module restructure, pattern migration, interface redesign, cross-file changes.

```
docs/refactors/NNN-refactor-name/
├── prd.md               ← Why refactor, goals, scope
├── techspec.md          ← New design + migration path + compatibility matrix
├── tasks/
│   ├── tasks.md
│   └── NNN-name.md
├── progress.md          ← Trace log
└── adr.md               ← MANDATORY — link to docs/adrs/NNN-adr-name.md
```

### Steps

1. **Research** — Map the current state. Use Scout to trace callers, consumers, and cross-module dependencies. Build the compatibility matrix (see below).
2. **PRD** — Why this refactor. What goal it achieves. What the scope boundary is. What is explicitly OUT of scope.
3. **TechSpec** — New design, migration path, rollback plan. Include the compatibility matrix. **ADR is MANDATORY.**
4. **Tasks** — Phased approach. Each phase must leave the codebase in a working state. Never big-bang.
5. **Implement** — One phase at a time. Tests pass after each phase.
6. **Review** — Full reviewer pass against TechSpec. Compatibility matrix re-checked.
7. **Standards update** — If a new pattern is introduced, update `docs/standards/`. Mark old pattern deprecated.

### Compatibility Matrix (required in TechSpec)

Add this section to `techspec.md` before implementation begins.

```markdown
## Compatibility Matrix

| Symbol / Interface | Current callers | Breaking after refactor? | Migration required? |
|---|---|---|---|
| [MethodName()] | [files:lines] | Yes / No | Yes / No |
| [InterfaceName] | [files:lines] | Yes / No | Yes / No |
| [ExportedConst] | [files:lines] | Yes / No | Yes / No |
```

> Every row where "Breaking" = Yes must have a corresponding migration task in tasks/tasks.md before implementation starts. No exceptions.

### Constraints
- Refactors MUST be phased. No "rewrite everything in one go."
- Each phase must leave the codebase in a working state.
- ADR captures WHY the architecture changed — link it from progress.md.
- Existing tests must pass at every phase. Add tests for new patterns.
- If the refactor reveals a bug: do NOT fix it inline. Open a separate bugfix ticket.

---

## Post-Refactor Checklist (Both Paths)

- [ ] All callers of renamed/moved symbols updated
- [ ] Tests pass
- [ ] `docs/standards/` updated if a new pattern was introduced
- [ ] Root `AGENTS.md` codebase map updated if module structure changed
- [ ] ADR exists and is linked (MACRO only)
- [ ] `KNOWLEDGE_MAP.md` updated (MACRO only)

## Self-Improvement

Refactors are the highest-impact changes to project knowledge.

- Module structure changed? → Update root AGENTS.md codebase map immediately
- Code patterns changed? → Update ALL affected docs/standards/ files
- Old pattern replaced? → Mark old pattern deprecated in standards
- Cross-module boundaries changed? → Update docs/standards/architecture/
