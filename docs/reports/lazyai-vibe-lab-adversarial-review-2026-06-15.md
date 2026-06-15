# Adversarial Review — LazyAI vs vibe-lab Alignment

**Date:** 2026-06-15  
**Branch reviewed:** `chore/lazyai-native-alignment-cleanup`  
**Branch tip at review start:** `d989823`  
**Method:** 3 rounds, paired advocate/skeptic subagents, forced cross-examination, main-agent evidence verification  
**Verification baseline:** `go test ./packages/cli/...` → 35 packages ok, 10 no tests

---

## Executive summary

This branch is **materially closer** to the user's requested target, but it is **not yet at ≥90% alignment** with vibe-lab-like default behavior.

### Final scores

- **Current alignment:** **78%**
- **Confidence in this review:** **92%**
- **Projected alignment after must-fix set:** **92%**
- **Projected confidence after must-fix set:** **87–89%**

### Bottom line

The branch now has the right high-level philosophy:
- setup-core is the headline story,
- vibe-lab is treated as an input, not a runtime dependency,
- retired task/workflow/orchestration/eval surfaces stay retired,
- neutral active agents replace Fortnite-era defaults,
- ADRs now record the alignment contract and the core-vs-optional split.

But several **active emitted surfaces** still contradict that story. The largest remaining problems are not abstract philosophy; they are concrete contract mismatches in shipped hooks, Antigravity path naming, memory-promotion behavior, and Speckit provenance/defaults.

Verdict: **Do not claim ≥90% alignment yet.** Fix the must-fix set below first.

---

## Review method

### Round 1 — Independent assessment

- **Advocate:** 86% alignment, 86% confidence
- **Skeptic:** 72% alignment, 88% confidence

### Round 2 — Cross-examination

After reviewing each other's strongest points:

- **Advocate revised:** 78% alignment, 89% confidence
- **Skeptic revised:** 76% alignment, 89% confidence

### Round 3 — Convergence

Both reviewers independently converged on a narrow must-fix set that would move the branch above 90%.

---

## What is already strong

1. **Top-level framing is much better.** README now leads with setup-core and explicitly classifies runtime-adjacent commands as optional modules.
   - Evidence: `README.md:3-15`, `README.md:20-26`

2. **The alignment contract is now explicit.** ADR-004 and ADR-005 define capability-first parity and core-vs-optional framing.
   - Evidence: `specs/adrs/004-vibe-lab-alignment-contract.md:20-25`, `specs/adrs/004-vibe-lab-alignment-contract.md:72-79`, `specs/adrs/005-core-vs-optional-modules.md:20-25`, `specs/adrs/005-core-vs-optional-modules.md:73-79`

3. **Retired runtime surfaces remain retired.** The branch does not reintroduce task/workflow/orchestration/eval CLI surfaces.
   - Evidence: `README.md:97-106`, `packages/cli/cmd/create.go:93-104`, `packages/cli/internal/validation/validation.go:48-59`

4. **Neutral active agents are now the default set.**
   - Evidence: `packages/cli/internal/types/types.go:397-404`

5. **OpenCode plugin naming is now native.** `vibe-lab-hooks.js` and `VibeLabHooks` were renamed to `lazyai-hooks.js` and `LazyAIHooks`.
   - Evidence: `packages/cli/library/opencode/plugins/lazyai-hooks.js:58`

6. **Architecture decisions are now actually versioned.** ADR markdown is committed instead of hidden behind `.gitignore`.
   - Evidence: `.gitignore:47-52`, tracked `specs/adrs/001-005`

---

## Findings and debate resolution

### 1) Startup self-heal hooks still depend on a repo-only script

- **Severity:** High
- **Status:** **Blocker**
- **Advocate:** Agrees this keeps alignment below 90%
- **Skeptic:** Agrees this is the highest-risk emitted-surface defect

**Evidence**
- `packages/cli/library/opencode/plugins/lazyai-hooks.js:73-76`
- `packages/cli/library/claudecode/hooks/startup-self-heal.sh:4-5`
- `packages/cli/library/copilot/hooks/startup-self-heal.sh:4-5`
- `packages/cli/library/antigravity/hooks/vibe-lab/startup-self-heal.sh:4-5`
- `README.md:25`

**Why it matters**

Generated hook surfaces call `bin/startup-self-heal`, but the README says `bin/startup-self-heal` is a **repository harness only** script, not a shipped CLI command. That means the branch advertises/parity-claims a startup repair behavior that is not self-contained in emitted repos.

**Resolution**

Must fix before any ≥90% claim. Either:
- ship/embed a real generated startup-repair implementation, or
- remove startup self-heal from active emitted behavior and docs.

---

### 2) Antigravity still emits `hooks/vibe-lab/*` while docs/spec moved to `lazyai`

- **Severity:** High
- **Status:** **Blocker**
- **Advocate:** Agrees this is a blocker
- **Skeptic:** Agrees this is a blocker

**Evidence**
- `packages/cli/library/antigravity/settings.json:7,19,32`
- `packages/cli/library/antigravity/hooks/vibe-lab/startup-self-heal.sh:1-5`
- `docs/concepts/tools.md:42-49`
- `docs/concepts/library-manifests.md:27-28`
- `specs/refactors/026-vibe-lab-alignment/spec.md:66`

**Why it matters**

This is an active emitted contract mismatch. The new native-alignment story says `lazyai`, but generated Antigravity output still ships `vibe-lab` path branding.

**Resolution**

Must fix before any ≥90% claim. Rename the full Antigravity hook surface to `.gemini/hooks/lazyai/*.sh`, or revert the docs/spec back to `vibe-lab` with an explicit rationale. Current split-brain state is not acceptable.

---

### 3) Memory-promotion runtime behavior is still not wired

- **Severity:** High
- **Status:** **Blocker**
- **Advocate:** Agrees this keeps alignment below 90%
- **Skeptic:** Agrees this keeps alignment below 90%

**Evidence**
- `packages/cli/library/opencode/plugins/lazyai-hooks.js:61-78`
- `packages/cli/internal/adapter/claudecode.go:60-94`
- `packages/cli/library/claudecode/hooks/caveman-memory-promotion.sh:1-5`
- `packages/cli/library/copilot/hooks/caveman-memory-promotion.sh:1-5`
- `.agents/hooks/caveman-memory-promotion/POLICY.md:5-11`

**Why it matters**

The policy claims a memory-promotion behavior, but the actual emitted surfaces are advisory/no-op or simply unwired. This makes the parity claim too strong relative to implementation.

**Resolution**

Must fix before any ≥90% claim. Either wire real behavior per adapter or narrow the spec/docs/manifests so they stop claiming it as approved runtime parity.

---

### 4) Speckit skills now emit by default, but curation still says they are not adapter targets

- **Severity:** Medium-High
- **Status:** **Blocker for ≥90%**
- **Advocate:** Agrees non-minimal default emission and manifest disagree
- **Skeptic:** Agrees this broadens default surface without truthful provenance

**Evidence**
- `packages/cli/internal/types/types.go:406-437`
- `packages/cli/library/manifests/curation.yaml:266-329`

**Why it matters**

The branch fixed the missing Speckit skill emission hole, but now the default emitted surface and the curation manifest disagree about intent. That is a provenance/contract problem, not just a docs nit.

**Resolution**

Must fix before any ≥90% claim. Pick one truth:
- keep Speckit in defaults and mark them as approved adapter targets, or
- remove them from default emission.

---

### 5) “Optional modules” are still mostly a prose boundary, not a surface boundary

- **Severity:** High
- **Status:** **Blocker for ≥90% unless wording/surface are tightened**
- **Advocate:** Calls it an accepted transitional state, but still below 90%
- **Skeptic:** Treats it as a must-fix because root-visible extras keep the default product broad

**Evidence**
- `README.md:22-25`
- `packages/cli/cmd/root.go:42-50`
- `docs/cli/reference.md:3-5`
- `docs/concepts/product-boundaries.md:45-85`

**Why it matters**

The branch now says runtime-adjacent commands are opt-in modules, but the shipped CLI/docs still present a broad root-visible runtime surface. The product framing is improved, but the lived surface is not yet as narrow as the claim suggests.

**Resolution**

Must fix before any ≥90% claim, but the minimum fix does **not** require full extraction. One of these is enough:
- hide/gate/de-emphasize ops-runtime-extra at the root/help/docs surface, or
- weaken the claim from “opt-in modules” to “shipped transitional extras outside setup-core”.

Full module lifecycle can defer.

---

### 6) Drift scan does not catch active Antigravity hook-path drift

- **Severity:** Medium
- **Status:** **Important guardrail gap, but not the main semantic blocker**
- **Advocate:** Can defer if concrete active drift is fixed
- **Skeptic:** Wants it included in the must-fix set

**Evidence**
- `tests/scripts/drift-scan.sh:37-44`
- current live drift is in `packages/cli/library/antigravity/settings.json:7,19,32`

**Why it matters**

The new CI guard is valuable, but it currently misses the exact hook-path drift that remains active. That reduces confidence that the cleanup will stay fixed.

**Resolution**

Recommended in the same patch set that fixes Antigravity path drift. Not the philosophical blocker by itself, but it should land with the path rename.

---

### 7) README Supported Tools omits Pi and Antigravity

- **Severity:** Low
- **Status:** **Cheap fix; not a stand-alone blocker**
- **Advocate:** Wants it fixed before ≥90%
- **Skeptic:** Views it as minor if deeper docs are correct

**Evidence**
- `README.md:289-294`
- `docs/concepts/tools.md:1-4`, `docs/concepts/tools.md:33-49`

**Why it matters**

Top-level docs still understate the setup-core surface the branch now claims to support.

**Resolution**

Fix it. Cheap, high-signal, low-risk.

---

## Minimum fix set to cross 90%

1. **Startup self-heal:** make emitted hook behavior self-contained or remove it from active parity claims.
2. **Antigravity:** rename the emitted hook surface to `lazyai` everywhere or revert the docs/spec back consistently.
3. **Memory-promotion:** wire it for real or downgrade the claim everywhere.
4. **Speckit:** resolve default-emission vs curation drift.
5. **Optional modules:** either narrow root/help/docs presentation or weaken the “opt-in” claim to a truthful transitional statement.
6. **README Supported Tools:** add Pi and Antigravity.
7. **Drift scan:** extend it to catch `hooks/vibe-lab` / `.gemini/hooks/vibe-lab` if the path rename lands.

Projected after completing 1-7: **~92% alignment**.

---

## Can defer while still claiming ≥90%

- Full physical module extraction / registry / enable-disable lifecycle
- Removal of old workflow/task/eval schema compatibility residue
- Full historical/provenance perfection for every old curated asset note
- Final status flips from Draft/Proposed to Accepted after human approval
- Broader archived-research cleanup

---

## Final recommendation

**Current state:** not yet at the requested ≥90% alignment threshold.

**Best honest score today:** **78% alignment, 92% review confidence**.

**Recommendation:**
- Accept the strategic direction.
- Do **not** call the branch “fully aligned” yet.
- Land the 5 core semantic fixes first:
  1. startup self-heal
  2. Antigravity path naming
  3. memory-promotion wiring/claim narrowing
  4. Speckit provenance/default coherence
  5. optional-module truthfulness at the default surface
- Then re-run this same 3-round adversarial review. Expected outcome after those fixes: **91–93% alignment**.
