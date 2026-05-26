---
name: zero-point
description: "The ultimate reality check. Two-phase quality gate: pre-flight YAGNI enforcement (anti-speculation) and post-flight independent verification. Nothing passes without judgment. Read-only checker — finds problems, doesn't fix them."
trigger: /zero-point
skill_path: skills/zero-point
scripts:
  - name: quality-gate.sh
    description: Run repo-aware quality gates (lint, test, build)
    path: scripts/quality-gate.sh
  - name: skill-audit.sh
    description: Cross-reference skills ↔ agents for consistency
    path: scripts/skill-audit.sh
  - name: check-models.sh
    description: Validate all configured models are available
    path: scripts/check-models.sh
  - name: mcp-health.sh
    description: Verify MCP servers are healthy
    path: scripts/mcp-health.sh
  - name: token-budget.sh
    description: Monitor token usage against budget
    path: scripts/token-budget.sh
  - name: contract-check.sh
    description: Pre/post condition assertions for implementation tasks
    path: scripts/contract-check.sh
---

## Quick Reference

| | |
|---|---|
| **Use when** | pre-flight YAGNI check, post-flight verification, quality gates |
| **Do not use when** | implementation, research |
| **Primary agent** | shield-audit |
| **Runtime risk** | Low — read-only checker |
| **Outputs** | Structured JSON gate results, violation reports |
| **Validation** | Contract checks, model health |
| **Deep mode trigger** | `/zero-point` or MODE=security/adversarial |

# Zero Point


## Tool Selection

For running verification tests on containerized services, use `dev exec`:
```bash
dev exec <service> --non-interactive -- <test-command>
```
See `skills/dev-cli/SKILL.md` for full reference.



Use the right tool for each job. See skills/_tool-hierarchy.md for full decision tree.

| Task | Tool |
|------|------|
| Read known file | OpenCode Read |
| Find code by description | morph codebase_search |
| Symbol analysis | codegraph MCP |
| Vault search | qmd MCP |
| Architecture overview | graphify CLI |


## Purpose
The Zero Point is the ultimate arbiter of quality. Two phases, one skill:

1. **Pre-flight: YAGNI Gate** — Prevent speculative implementation before code is written
2. **Post-flight: Verification** — Independent spec compliance and quality gate check after implementation

**Role:** Read-only throughout. You produce evidence and a verdict. You do not fix.

**Separation of concerns:**
- `build-mode` writes code
- `zero-point` judges it
- If zero-point rejects, hand back to `build-mode` or `reboot-van` with specific findings — never fix inline

---

## Scripts

This skill owns the following scripts:

| Script | Purpose |
|--------|---------|
| `quality-gate.sh` | Run repo-aware quality gates (lint, typecheck, test, build) |
| `skill-audit.sh` | Cross-reference skills ↔ agents for consistency |
| `check-models.sh` | Validate all configured models are available |
| `mcp-health.sh` | Verify MCP servers are healthy |
| `token-budget.sh` | Monitor token usage against budget |
| `contract-check.sh` | Pre/post condition assertions for implementation tasks |

Run from skill directory: `./scripts/<script-name>.sh`

---

# Phase 1: Pre-flight — YAGNI Gate

Detect and prevent speculative implementation — code that goes beyond what the spec or task requires. Run this before or during implementation to prevent scope creep.

## The 5 Speculation Patterns

### Pattern 1 — Feature Creep
Adding behaviors not listed in the spec or task description.

**Detection**: Implementation includes endpoints, UI elements, or business logic with no corresponding entry in the spec's scope or behaviors section.

### Pattern 2 — Premature Abstraction
Creating generic frameworks, base classes, or plugin systems for a single use case.

**Detection**: Abstract classes, strategy patterns, factory methods, or configuration-driven behavior where the spec describes exactly one concrete behavior.

### Pattern 3 — Future-Proofing
Adding parameters, flags, or extension points "for later."

**Detection**: Optional parameters not in the data contract. Feature flags for features not in scope. Database columns not referenced in any behavior.

### Pattern 4 — Gold Plating
Over-engineering error handling, logging, or monitoring beyond spec requirements.

**Detection**: Retry logic, circuit breakers, or caching where the spec doesn't mention performance or reliability concerns for that path.

### Pattern 5 — Scope Leakage
Touching files or services not listed in the task scope.

**Detection**: Diff includes files in services not assigned to the current task. Changes to shared libraries not mentioned in the plan.

## YAGNI Checkpoint

Before writing any production code, answer these four questions:

1. **Is this behavior in the spec?** → If NO, don't build it.
2. **Is this the simplest way to satisfy the spec?** → If NO, simplify.
3. **Would removing this code cause a spec'd test to fail?** → If NO, remove it.
4. **Does this belong in the current task's scope?** → If NO, flag it.

## Detection Heuristics

When reviewing implementation against spec:

- [ ] Every new public function/method maps to a spec behavior
- [ ] Every new database column maps to a data contract field
- [ ] Every new endpoint maps to a spec behavior
- [ ] No abstract base classes for single implementations
- [ ] No optional parameters beyond the data contract
- [ ] Diff stays within the scope of the task

## Halt Protocol

When speculation is detected:

1. **IDENTIFY** the speculation pattern (1–5)
2. **QUOTE** the spec section that should govern this code
3. **SHOW** the speculative code or design
4. **SUGGEST** the minimal spec-compliant alternative
5. **ASK** the user:
   - (A) Remove this and implement minimally
   - (B) Amend the spec to include this behavior
   - (C) Keep with explicit approval (noted in plan)

### Halt Output Format

```
## ⚠️ Speculation Detected — Pattern {N}: {Name}

**Spec says**: "{quote from spec}"
**Implementation adds**: "{description of speculative code}"

**Minimal alternative**: {what spec-compliant code would look like}

Options:
- (A) Remove and implement minimally
- (B) Amend the spec to include this behavior
- (C) Keep with explicit approval
```

---

# Phase 2: Post-flight — Verification

Act as an independent checker. Validate that implementation matches the spec — no more, no less.

## Tooling — Use `codebase_search` (WarpGrep)

Use `codebase_search` (WarpGrep) during **Step 3: Spec Compliance Check** to locate implementation evidence for each requirement.

Best for: "Find the implementation of requirement X", "Where is this acceptance criterion handled?", locating test coverage for spec items.
Not for: running quality gates — use the bash commands specified in the Gates table for that, reading known files — use OpenCode `Read`.

## Process

### Step 1: Load Source of Truth
Before checking anything, locate and read:
1. The spec (`.specify/spec.md` or inline spec from the task)
2. The acceptance criteria
3. The task's done-condition

If the spec is missing: report `FAIL — No spec found. Cannot verify without a source of truth.` Stop here.

### Step 2: Run Quality Gates
Execute the repo-appropriate quality gates using `./scripts/quality-gate.sh` or manually:

| Repo Profile | Gates |
|-------------|-------|
| `fedora` | `bundle exec rubocop`, `bundle exec rspec` |
| `creator-checkout` | `npm run quality`, `npm run build` |
| `mono-frontend` | `yarn lint`, `yarn typecheck`, `yarn test`, `yarn build` |
| `school-plan-service` | `go test ./...`, `go vet ./...` |
| Unknown | Ask human which gates apply before proceeding |

After quality gates, run post-condition assertions:
```bash
./scripts/contract-check.sh --mode post --spec-dir bee-gone/specs/<NNN-slug> --repo-profile <profile>
```
Validates: no new lint errors, spec files exist, tests still pass, no orphaned tests.

Record each gate result via `session-db.sh metric-record <sid> <repo> <gate-type> <passed> <duration-ms> <errors> <warnings>` for historical tracking.

### Step 3: Spec Compliance Check
For each requirement in the spec:
- Is it implemented? (yes / no / partial)
- Is the implementation correct per the acceptance criteria?

### Gate Result Output Format

zero-point emits a structured JSON result that captures the outcome of all gates, findings, and a summary. This JSON is the machine-readable contract for downstream consumers (e.g., `truth-chain`, `session-db.sh`, CI pipelines).

```json
{
  "status": "PASS|FAIL|WARN",
  "gates": [
    {
      "name": "lint",
      "status": "PASS|FAIL|WARN",
      "errors": 0,
      "warnings": 2,
      "duration_ms": 1500
    }
  ],
  "findings": [
    {
      "severity": "error|warning|info",
      "message": "...",
      "file": "path/to/file",
      "line": 42
    }
  ],
  "summary": "2/3 gates passed, 1 warning"
}
```

**Field definitions:**
- `status` — Overall verdict. `PASS` if all gates pass and no errors; `FAIL` if any gate fails or any finding has `severity: error`; `WARN` if all gates pass but there are warnings.
- `gates` — Array of per-gate results. Each gate must report `name`, `status`, `errors`, `warnings`, and `duration_ms`.
- `findings` — Array of individual findings from all gates. Each finding must include `severity`, `message`, and may include `file` and `line`.
- `summary` — Human-readable summary string for quick inspection.

**Rules:**
- The JSON must be emitted to stdout as the final output of the zero-point run.
- If a gate cannot be executed (e.g., missing tool), it is recorded as `FAIL` with a finding explaining the failure.
- Findings are cumulative across all gates; a single gate may produce zero or more findings.

### Step 4: Scope Check
- Does the implementation do **more** than the spec requires? (scope creep — flag it)
- Does the implementation do **less** than the spec requires? (gaps — flag them)

### Step 5: Drift-Scope Check
Run drift-scope to validate implementation hasn't drifted from spec claims:
```bash
./skills/drift-scope/scripts/drift-check.sh --spec <spec-file> --repo <repo-path>
```
This checks:
- **§V invariants** — are all spec invariants enforced in code?
- **§I interfaces** — do function signatures match the spec contract?
- **§C constraints** — are constraints (performance, security, etc.) respected?
- **§T tasks** — are all tasks completed and marked done?

Drift violations are reported as additional findings in the Verification Report. A drift violation is a FAIL — the implementation has diverged from the spec.

### Step 6: Report

```markdown
## Verification Report

**Overall:** PASS / FAIL / PARTIAL

### Quality Gates
| Gate | Status | Notes |
|------|--------|-------|
| [gate name] | ✅ PASS | [details] |
| [gate name] | ❌ FAIL | [error output] |

### Spec Compliance
| Requirement | Status | Evidence |
|-------------|--------|----------|
| [req text] | ✅ Met | [where/how it's implemented] |
| [req text] | ❌ Missing | [what's absent] |
| [req text] | ⚠️ Partial | [what's there, what's not] |

### Scope Assessment
- **Over-scope:** [anything implemented beyond spec, or "none"]
- **Under-scope:** [anything in spec not implemented, or "none"]

### Verdict
[1–2 sentences: what passed, what failed, what's blocked]

### Blocked items (if FAIL or PARTIAL)
[Specific findings that must be addressed — no HOW, just WHAT is wrong]
```

---

## Integration
- **Pre-flight (YAGNI)**: Run before or during `build-mode` to prevent speculation
- **Post-flight (Verify)**: Run after `build-mode` completes to validate the output
- **Drift-scope**: Runs as Step 5 of post-flight — validates spec claims vs actual code
- **Ricochet**: Receives backprop data when verification finds test failures
- **Truth-chain**: Records verification gate results, drift violations, and quality gate outcomes as immutable ledger entries
- **sidecar**: On task start, if `.sidecar.yml` is discoverable in the active repo, run `sidecar query` to find related specs. Load any returned specs before pre-flight YAGNI gate and post-flight verification to ensure checks match the full spec context. If no sidecar config exists or the query returns no specs, proceed normally without failing.
- **Reviewer agent**: Runs detection heuristics during code review
- **Planner agent**: References YAGNI patterns when scoping tasks

## Rules
- Never suggest fixes — report findings, hand back to `build-mode` or `reboot-van`
- Never run mutations — read-only throughout
- If a gate fails: report the failure, do not attempt to fix it
- Compare against spec only — not against what "seems reasonable"
- A `PARTIAL` verdict is always honest; a false `PASS` is a failure of this skill
- If acceptance criteria are ambiguous, flag it — do not interpret charitably or harshly
- The Zero Point does not bend. Either the implementation passes or it doesn't.

## MODE=eval — LLM-as-Judge Scoring

When invoked with MODE=eval, zero-point acts as an LLM-as-judge evaluator for storm-eye.

### How It Works
1. storm-eye provides: agent output + eval rubric
2. zero-point scores each rubric criterion (0-1)
3. Returns: score, reasoning, specific issues found
4. Logs to truth-chain as eval_run entry

### Rubric Format
```yaml
rubric:
  - criterion: "Identified all critical unknowns"
    weight: 0.3
  - criterion: "Asked focused, non-redundant questions"
    weight: 0.3
  - criterion: "Searched existing knowledge before asking"
    weight: 0.2
  - criterion: "Output structured and actionable"
    weight: 0.2
```

### Integration
- **storm-eye**: Defines rubrics, collects results
- **truth-chain**: Logs eval_run entries
- **lesson-loot**: Low scores trigger lesson extraction
- **ricochet**: Persistent failures become §V invariants
