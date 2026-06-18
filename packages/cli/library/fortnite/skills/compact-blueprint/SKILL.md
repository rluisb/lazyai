---
name: compact-blueprint
description: Caveman-style spec templates with compressed encoding. ~75% fewer tokens than prose specs. Sections: §G goal, §C constraints, §I interfaces, §V invariants, §T tasks, §B bugs.
trigger: /compact-blueprint
skill_path: skills/compact-blueprint
scripts:
  - name: spec-encode.sh
    description: Convert prose spec to caveman-encoded compact format
    path: scripts/spec-encode.sh
---

# Compact Blueprint — Caveman-Style Spec Templates

## Purpose
Specs cost tokens every time they're read. Compact Blueprint encodes specs in caveman format — ~75% fewer tokens than prose, same technical accuracy. Brain still big. Spec still clear.

**Use when:**
- Creating a new spec — "compact-blueprint for this feature"
- Converting existing spec — "encode this spec"
- Before implementation — "load compact spec"

## Spec Format

```markdown
## §G Goal
feature: auth mw refresh path. scope: token expiry handling.

## §C Constraints
| id | rule | priority |
|----|------|----------|
| C1 | no cloud calls | critical |
| C2 | < 150ms latency | high |

## §I Interfaces
| id | signature | returns |
|----|-----------|---------|
| I1 | refresh(token) | {valid, expiry} |
| I2 | validate(session) | bool |

## §V Invariants
| id | rule | evidence |
|----|------|----------|
| V1 | auth mw throws 401 @ expired | test: auth_test.go:42 |
| V2 | null guard @ user profile | test: profile_test.go:15 |

## §T Tasks
| id | desc | done | files |
|----|------|------|-------|
| T1 | add refresh path | ☐ | auth.go |
| T2 | update mw check | ☐ | middleware.go |

## §B Bugs
| id | symptom | fix | status |
|----|---------|-----|--------|
| B1 | null user @ profile | add guard L42 | fixed |
```

## Scripts

| Script | Purpose | Key Flags |
|--------|---------|-----------|
| `spec-encode.sh` | Convert prose spec to caveman format | `--input <file>`, `--output <file>` |

## Workflow

### Step 1: Create Spec
Use the template above or run:
```bash
./scripts/spec-encode.sh --input prose-spec.md --output SPEC.md
```

### Step 2: Fill Sections
- §G: One-line goal, scope boundaries
- §C: Constraints as pipe table
- §I: Function signatures, API contracts
- §V: Invariants from ricochet backpropagation
- §T: Tasks with completion status
- §B: Bug history from ricochet

### Step 3: Use in Workflow
- Battle-bus templates reference compact specs
- Build-mode implements against §T tasks
- Zero-point verifies §V invariants
- Drift-scope checks spec compliance

## Integration with Other Skills

- **battle-bus**: Uses compact specs in all templates
- **ricochet**: Adds §V/§B entries from test failures
- **drift-scope**: Validates spec claims vs code
- **build-mode**: Implements §T tasks one by one
- **sidecar**: On task start, if `.sidecar.yml` is discoverable in the active repo, run `sidecar query` to find related specs. Load any returned specs before encoding to cross-reference related specs and avoid duplicate or conflicting §G/§C/§I claims. If no sidecar config exists or the query returns no specs, proceed normally without failing.
- **zero-point**: Verifies §V invariants pass

## Tips

- Keep §G to one line — if it's longer, scope is too big
- §V invariants are permanent — they become quality gates
- Use pipe tables for repeating records (tasks, bugs, constraints)
- Code blocks, URLs, paths preserved byte-for-byte
- Specs live at repo root or `specs/<slug>/SPEC.md`
