---
name: drift-scope
description: Spec vs implementation drift detection. Checks if code still satisfies spec invariants, interfaces, constraints, and task completion.
trigger: /drift-scope
skill_path: skills/drift-scope
scripts:
  - name: drift-check.sh
    description: Compare spec claims vs actual code, report violations
    path: scripts/drift-check.sh
---

## Quick Reference

| | |
|---|---|
| **Use when** | Spec vs implementation drift detection, compliance checks |
| **Do not use when** | Initial implementation, research |
| **Primary agent** | shield-audit |
| **Runtime risk** | Low — read-only drift detection |
| **Outputs** | Drift report, violation list, compliance score |
| **Validation** | Spec claim coverage, code scan accuracy |
| **Deep mode trigger** | `/drift-scope` or release readiness check |

# Drift Scope — Spec vs Implementation Drift Detection

## Purpose
Specs decay as code evolves. Drift Scope scans the codebase and checks if the implementation still satisfies the spec's claims: invariants (§V), interfaces (§I), constraints (§C), and task completion (§T).

**Use when:**
- Before a release — "check for spec drift"
- After refactoring — "verify invariants still hold"
- During code review — "drift-check this PR"
- Periodic health check — "scope the drift"

## Scripts

| Script | Purpose | Key Flags |
|--------|---------|-----------|
| `drift-check.sh` | Parse spec → scan code → report violations | `--spec <path>`, `--root <dir>`, `--json` |

## Workflow

### Step 1: Parse Spec
Extract claims from spec file:
- §V invariants → code patterns that must exist
- §I interfaces → function signatures that must match
- §C constraints → rules that must be enforced
- §T tasks → completion status to verify

### Step 2: Scan Codebase
Search code for:
- Invariant enforcement (guards, validations, checks)
- Interface implementations (function signatures, types)
- Constraint compliance (rate limits, auth checks, etc.)
- Task completion (files exist, tests pass)

### Step 3: Generate Drift Report
Output violations:
```
## Drift Report
| Section | Claim | Status | Detail |
|---------|-------|--------|--------|
| §V1 | guard @ auth mw | ✅ enforced | middleware.go:42 |
| §V2 | rate limit 100rps | ❌ missing | no rate limiter found |
| §I1 | GetUser(id) → User | ✅ matches | user.go:15 |
```

## Integration with Other Skills

- **ricochet**: After drift found, ricochet adds new invariants
- **zero-point**: Drift check runs as part of verification
- **compact-blueprint**: Uses same spec format (§G/§C/§I/§V/§T/§B)
- **sidecar**: On task start, if `.sidecar.yml` is discoverable in the active repo, run `sidecar query` to find related specs. Load any returned specs before Step 1 (Parse Spec) to ensure drift-checking covers the complete spec set. If no sidecar config exists or the query returns no specs, proceed normally without failing.
- **build-mode**: Drift check before merging to main

## Tips

- Run drift-scope before every release
- §V violations are critical — they're spec invariants
- §I mismatches indicate API drift
- Use `--json` for CI/CD integration
- Drift report becomes input for ricochet backpropagation
