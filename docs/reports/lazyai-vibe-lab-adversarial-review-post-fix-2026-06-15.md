# Adversarial Review — LazyAI vs vibe-lab Alignment (Post-fix)

**Date:** 2026-06-15  
**Branch reviewed:** `chore/lazyai-native-alignment-cleanup`  
**Method:** 3 rounds, paired advocate/skeptic subagents, forced cross-examination, main-agent evidence verification  
**Verification baseline:**
- `go test ./packages/cli/...` → 35 packages ok, 10 no tests
- `bash -n bin/inject bin/doctor` → passed
- `tests/scripts/drift-scan.sh` → clean
- `git diff --check` → clean

---

## Executive summary

The branch now **credibly clears the user’s ≥90% target** for LazyAI-to-vibe-lab alignment while preserving the approved product boundary: setup-core is the default product, runtime-adjacent capabilities remain transitional extras outside setup-core, and no retired task/workflow/orchestration/eval surfaces were reintroduced.

### Final publishable scores

- **Alignment:** **92%**
- **Confidence:** **90%**

### Why the score moved above 90

The prior blockers are now closed:
- emitted adapters no longer depend on repo-only `bin/startup-self-heal`
- startup-self-heal and caveman-memory-promotion are narrowed to policy/harness-only, not falsely advertised as emitted runtime behavior
- Antigravity emitted assets now use `.gemini/hooks/lazyai/*`
- Speckit curation now matches default adapter emission
- top-level docs and root help now frame ops-runtime-extra as transitional extras outside setup-core
- README Supported Tools includes Pi and Antigravity
- drift scan now catches the old Antigravity path regression
- `bin/inject` and `bin/doctor` no longer regenerate or validate the retired emitted startup/caveman hook surface

---

## Round 1 — refreshed independent review

### Advocate
- **Score:** 91%
- **Confidence:** 88%
- **Summary:** Active emitted hook surfaces now match the narrowed contract; the remaining issues are planning-doc drift, command-section ordering, and low-severity wording debt.

### Skeptic
- **Score:** 88% before final harness fix, 91% if harness generator drift is excluded
- **Confidence:** 90%
- **Summary:** The only remaining serious blocker was that `bin/inject`/`bin/doctor` could reintroduce the retired hook/path surface; once fixed, the remaining issues become transitional debt rather than active contradictions.

---

## Round 2 — blocker closure re-check

Main-agent patch set closed the final skeptic blocker:
- `bin/inject` now skips `startup-self-heal` and `caveman-memory-promotion` during adapter emission and emits Antigravity hooks under `.gemini/hooks/lazyai`
- `bin/doctor` skips policy-only startup/caveman hooks, checks Antigravity under `.gemini/hooks/lazyai`, and no longer expects startup hooks in Claude/Antigravity settings
- generated root copies were refreshed/removed so stale `.github/hooks/startup-self-heal*` and `.github/hooks/caveman-memory-promotion*` no longer remain

### Advocate revised
- **Score:** 93%
- **Confidence:** 90%
- **Verdict:** clears ≥90

### Skeptic revised
- **Score:** 92%
- **Confidence:** 90%
- **Verdict:** clears ≥90

---

## Round 3 — convergence

Both reviewers agreed the fair synthesis score is:

- **92% alignment**
- **90% confidence**

Convergence rationale: the remaining issues are no longer contradictions in the default emitted setup-core surface. They are stale planning docs, presentation polish, dead/no-op harness branches, and explicitly accepted transitional-module debt.

---

## Strong evidence for the ≥90 claim

1. **Emitted hook contract is now truthful and self-consistent**
   - OpenCode plugin only emits shell-blocking + objective gate behavior
   - Claude/Copilot/Antigravity no longer emit startup-self-heal or caveman-memory-promotion hooks
   - Evidence: `packages/cli/library/opencode/plugins/lazyai-hooks.js`, `packages/cli/internal/adapter/claudecode.go`, `packages/cli/library/antigravity/settings.json`, deleted startup/caveman hook assets under `packages/cli/library/{claudecode,copilot,antigravity}`

2. **Antigravity native path contract is consistent end-to-end**
   - Evidence: `packages/cli/internal/adapter/antigravity.go`, `packages/cli/library/antigravity/settings.json`, `packages/cli/library/antigravity/hooks/lazyai/*`, `docs/concepts/tools.md`, `docs/concepts/library-manifests.md`, `tests/scripts/drift-scan.sh`

3. **Speckit default emission and manifest provenance agree**
   - Evidence: `packages/cli/internal/types/types.go`, `packages/cli/library/manifests/curation.yaml`

4. **Docs/root surface now tell the truth about runtime extras**
   - Evidence: `README.md`, `docs/cli/reference.md`, `docs/concepts/product-boundaries.md`, `packages/cli/cmd/root.go`

5. **Generator/checker drift path is closed**
   - Evidence: `bin/inject`, `bin/doctor`

---

## Remaining non-blocking issues

1. **Planning-doc drift**
   - `specs/refactors/026-vibe-lab-alignment/plan.md` and `tasks.md` still describe the older four-hook emitted-runtime plan.
   - Non-blocking because the live spec (`spec.md: FR-005`), emitted assets, tests, and docs now agree.

2. **README command-section polish**
   - The top-level framing is correct, but command examples can still be refined further to keep setup-core even more visually dominant.
   - Non-blocking because setup-core now leads in title, intro, and section framing.

3. **Dead no-op branches in `bin/inject`**
   - Startup-self-heal no-op branches remain in helper functions even though generation skips them.
   - Non-blocking because they are no longer emitted or validated.

4. **Runtime extras still physically ship**
   - This remains the main reason the score is 92% rather than mid-90s.
   - Non-blocking because ADR-005 explicitly accepts the transitional state.

---

## Final verdict

**Yes — this branch now credibly clears 90%.**

Published score:
- **Alignment:** **92%**
- **Confidence:** **90%**

The remaining debt is transitional and documented. The default emitted setup-core surface no longer contradicts the alignment contract.
