---
name: ricochet
description: Auto-update specs from test failures — backpropagation protocol. Every test failure becomes a spec invariant the system never forgets.
trigger: /ricochet
skill_path: skills/ricochet
scripts:
  - name: backprop.sh
    description: Parse test failures and update spec invariants
    path: scripts/backprop.sh
---

## Quick Reference

| | |
|---|---|
| **Use when** | Test failure backpropagation, spec invariant updates |
| **Do not use when** | Normal test passing, unrelated spec changes |
| **Primary agent** | wall-builder / shield-audit |
| **Runtime risk** | Low — spec updates from failures |
| **Outputs** | Updated spec with §V/§B entries |
| **Validation** | Failure pattern matching, dry-run default |
| **Deep mode trigger** | `/ricochet` or test suite failure |

# Ricochet — Backpropagation Protocol

## Purpose
Test failures should make the spec smarter, not just fix the code. Ricochet captures failure patterns and bounces them back into the spec as permanent invariants (§V) or bug entries (§B).

**Use when:**
- Test suite fails — "ricochet the failures into the spec"
- Bug fix applied — "backprop this fix as an invariant"
- After zero-point verification — "capture any violations as spec updates"

## Scripts

| Script | Purpose | Key Flags |
|--------|---------|-----------|
| `backprop.sh` | Parse test output → update spec with §V/§B entries | `--spec <path>`, `--test-output <path>`, `--dry-run` |

## Workflow

### Step 1: Capture Test Failure
Run test suite and capture output:
```bash
./scripts/backprop.sh --test-output test-failures.log --spec SPEC.md
```

### Step 2: Classify Failure
The script classifies each failure:
- **New bug** → Add §B entry (bug table)
- **Known pattern** → Add §V invariant (never-break rule)
- **Spec gap** → Add §I interface or §C constraint

### Step 3: Update Spec
The spec is updated with caveman-encoded entries:
```markdown
## §V Invariants
| id | rule | evidence |
|----|------|----------|
| V1 | auth mw throws 401 @ expired token | test: auth_test.go:42 |

## §B Bugs
| id | symptom | fix | status |
|----|---------|-----|--------|
| B1 | null user @ profile load | add guard L42 | fixed |
```

### Step 4: Report Changes
Output summary of what changed:
- New invariants added
- Bug entries created
- Spec sections updated

## Integration with Other Skills

- **zero-point**: After verification fails, ricochet captures violations
- **build-mode**: After test failures, ricochet updates spec before retry
- **compact-blueprint**: Uses same §V/§B spec format
- **sidecar**: On task start, if `.sidecar.yml` is discoverable in the active repo, run `sidecar query` to find related specs. Load any returned specs before backpropagation to ensure new §V/§B entries are anchored to the correct repo context. If no sidecar config exists or the query returns no specs, proceed normally without failing.
- **drift-scope**: Validates that invariants are still enforced

## Tips

- Run ricochet after every test failure, not just after fixes
- §V invariants are permanent — they become quality gates
- §B entries track bug history — useful for post-mortems
- Use `--dry-run` to preview changes before applying
- Ricochet works with any test framework (rspec, jest, go test, etc.)
