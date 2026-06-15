---
name: build-mode
description: Context-engineered implementation. Builds against spec, not against vibe. Every line of code is traceable to a spec requirement. Supports standard and TDD (RED→GREEN→REFACTOR) modes. Switch to build mode and construct with materials — no harvesting without a blueprint.
trigger: /build-mode
skill_path: skills/build-mode
scripts:
  - name: worktree-manager.sh
    description: Manage parallel worktrees for isolated implementation
    path: scripts/worktree-manager.sh
---

## Quick Reference

| | |
|---|---|
| **Use when** | implementation against spec, TDD mode, code construction |
| **Do not use when** | research, review, planning without spec |
| **Primary agent** | wall-builder |
| **Runtime risk** | Medium — code changes |
| **Outputs** | Implemented code, test results, file manifest |
| **Validation** | Spec traceability, test gates — **done means tested** |
| **Deep mode trigger** | `/build-mode` or MODE=tdd |

# Build Mode

## Purpose
Write code that fulfils the spec — no more, no less. Every implementation decision is traceable to a spec requirement. Implementation without a spec is guesswork dressed up as work.

**Two modes:**
- **Standard mode** — planned implementation from spec tasks
- **TDD mode** — test-first implementation using RED→GREEN→REFACTOR cycle

---

## Scripts

This skill owns the following scripts:

| Script | Purpose |
|--------|---------|
| `worktree-manager.sh` | Manage parallel worktrees for isolated implementation branches |

Run from skill directory: `./scripts/worktree-manager.sh <command>`

---

## Prerequisites
Before writing a single line of code, confirm:

1. **Spec exists** — `.specify/spec.md` or inline spec with testable requirements. If missing → load `storm-scout` skill first.
2. **Current task is clear** — specific task with an explicit done-condition. If missing → load `storm-scout` skill first.
3. **Context is loaded** — relevant existing code is read and understood. Do not implement blind.
4. **(Optional) Contract assertions** — Run `./skills/zero-point/scripts/contract-check.sh --mode pre` to validate pre-conditions (spec exists, branch clean, tests pass baseline, no stale locks).

---

## Tooling

### Code Exploration — Use the Right Tool

| Task | Tool | When |
|------|------|------|
| Read a known file | OpenCode `Read` | You know the path |
| Find code by description | morph `codebase_search` | Semantic exploration |
| Symbol analysis | codegraph MCP | Callers, callees, impact |
| Architecture overview | graphify CLI | Path between concepts, communities |

Full decision tree and workflow patterns: see `skills/_tool-hierarchy.md`.

### Test Execution — Use dev-cli

For containerized services, run tests via `dev exec`:
```bash
dev exec <service> --non-interactive -- <test-command>
```

See `skills/dev-cli/SKILL.md` for full dev CLI reference.
| Architecture overview | graphify CLI | Path between concepts, communities |

### Code Editing — Use `edit_file` (Morph Fast Apply)

Use `edit_file` (morph-mcp Fast Apply) for all code changes in **Phase 1: Implement**:

- Provide only the sections changing with `// ... existing code ...` markers for unchanged code
- Make all related edits in a single `edit_file` call instead of multiple calls to the same file
- Speed: 10,500+ tok/s with 98% accuracy — keeps build cycles fast and precise

---

## Standard Mode

### Phase 0: Context Load (mandatory)
Read before writing:
1. The spec — what are you implementing?
2. The current task — what is your exact done-condition?
3. Existing code in the affected area — what patterns are already in use?
4. Quality gate requirements — what must pass when you're done?

**Announce your scope:** "I am implementing [task name]. Done when [done-condition]. NOT touching [explicit exclusions]."

This announcement is not ceremony — it prevents scope creep before it starts.

### Phase 1: Intent Declaration
State in one sentence what you are about to write and explicitly what you are not touching:

> "I am adding [X] to [Y], which will enable [Z]. I am NOT changing [A], [B], or [C]."

### Purpose Gate (non-negotiable)
Before writing any code, declare your exact intent in full. This gate is mandatory — NEVER skip it:

```
## Purpose Gate

**What I am building:** [exact behavior/feature from the spec]
**Why this task exists:** [the spec requirement or bug report that justifies it]
**What I am NOT touching:** [files, subsystems, or behaviors explicitly excluded]
**Done looks like:** [the acceptance criteria this implementation must meet]
**Risks I see:** [anything that could go wrong or be ambiguous]
```

If you cannot fill all five fields: STOP. Return to `storm-scout` Phase 0 Clarify for the missing pieces. Never proceed with a partial Purpose Gate.

### Phase 2: Implement
Write the code. Follow existing patterns unless the spec explicitly calls for a change.

**Scope checks while building:**
- Every function/method: "Is this in the spec?"
- Every file touched: "Was this in the task's scope?"
- If you notice something useful to refactor: **note it, do NOT refactor now** — that's a separate task
- If you discover a missing requirement: **surface it, do NOT silently add it**

**Context engineering while building:**
- Explain what a block of code does before writing it (feedforward to yourself and to reviewers)
- Prefer small, named, well-scoped functions over large inline blocks
- Organise by domain concern, not by technical layer

### Phase 3: Self-Verify
Before declaring done:
1. Re-read the done-condition from the task
2. Re-read the acceptance criteria from the spec
3. Confirm: does what you wrote match what was asked?
4. **Run all tests — they MUST pass**
5. Run the quality gates (or mentally trace through them if tooling isn't available)

**Completion gate:** Implementation is NOT complete until all tests pass. If tests fail, return to Phase 2 (Implement) — do not mark the task as done.

### Phase 4: Ricochet Backprop (on test failure)
If tests fail after implementation:
1. **Run ricochet backprop** — `./skills/ricochet/scripts/backprop.sh` to parse test failures
2. **Update spec invariants** — backprop adds §V invariants and §B bugs to the compact spec
3. **Re-read updated spec** — the new invariants become part of the done-condition
4. **Retry implementation** — fix against the enriched spec, not against the old one

This closes the loop: test failures → spec invariants → implementation fixes. Never fix blindly — let ricochet tell you what the spec is missing.

---

## TDD Mode

When operating in TDD mode, use the RED→GREEN→REFACTOR inner loop. One test at a time.

### RED — Write a failing test
- One behavior per test
- Test should fail for the right reason (not a syntax error)
- If the test passes immediately, you misunderstood the requirement — STOP

### GREEN — Make it pass
- Write the minimum code to pass the test
- **All tests must pass, not just the new one** — regressions are failures
- No premature optimization, no extra features
- If the change is >30 lines, you're probably doing too much — break it down
- Use `edit_file` for the GREEN change — one call per cycle

### REFACTOR — Clean up
- Improve names, extract functions, remove duplication
- Tests must still pass after every refactor step
- No behavior changes during refactor
- Use `edit_file` for each REFACTOR step

### TDD Iteration Limits
- Max 10 RED→GREEN→REFACTOR cycles per session
- If stuck after 3 cycles on the same issue: STOP and escalate
- **On persistent test failure**: run ricochet `backprop.sh` to check if spec invariants are missing
- The TDD inner loop governs one test at a time; the outer loop (task completion) is governed by the standard mode phases

---

## Scope Discipline
If at any point you notice scope creep happening:
- Stop writing
- Note the scope boundary concern
- Surface it to the human before continuing

**Signs of scope creep:**
- "While I'm here, I'll also fix..."
- "This would be better if I also..."
- "The spec doesn't say, but I think..."

---

## Package Manager Safety

**Policy:** All package manager installations must use `--ignore-scripts` by default.

- **Forbidden:** `npm install` without `--ignore-scripts`, `yarn add` without `--ignore-scripts`, `pnpm add` without `--ignore-scripts`.
- **Logging:** Every package manager invocation must be recorded to `tool_calls`.
- **Override:** Explicit `--ignore-scripts` override requires justification in the implementation notes.
- **Rationale:** Prevents arbitrary package scripts from executing during agent operations.

This policy is enforced in `agents/wall-builder.md` forbidden tools and must be respected in all build-mode implementations.

---

## Integration with Other Skills

- **storm-scout**: Returns here when spec is unclear or missing
- **zero-point**: Runs post-implementation verification against spec
- **ricochet**: Runs backprop on test failures to update spec invariants (§V/§B)
- **drift-scope**: Validates implementation hasn't drifted from spec claims
- **truth-chain**: Records implementation decisions, changes, and quality gate results as immutable ledger entries
- **reboot-van**: Takes over when root cause of a bug is known
- **sidecar**: On task start, if `.sidecar.yml` is discoverable in the active repo, run `sidecar query` to find related specs. Load any returned specs during Phase 0 (Context Load) before proceeding with implementation. If no sidecar config exists or the query returns no specs, proceed normally without failing.
- **slurp-juice**: Checkpoints at phase transitions

## Output
- Working code
- Brief implementation note: what was done, what was explicitly skipped, any risks or follow-up tasks noted

## Rules
- Never implement what wasn't in the spec
- Never refactor while implementing — note it, ship it separately
- Never silently expand scope
- If the spec is unclear on a point: stop, return to `storm-scout` Phase 0 for that specific point
- Prefer existing patterns over inventing new ones
- Announce scope at start, self-verify at end — both are non-negotiable
