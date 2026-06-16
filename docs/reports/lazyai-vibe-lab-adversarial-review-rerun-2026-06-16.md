# Adversarial Review — LazyAI vs vibe-lab Alignment (Rerun)

**Date:** 2026-06-16  
**Branch reviewed:** `chore/baseline-parity-90`  
**Branch tip at review start:** `87011fb`  
**Review target:** current working tree relative to `/Users/ricardo/code/vibe-lab` baseline  
**Method:** evidence-only rerun against the exact-baseline criteria recorded in `docs/reports/lazyai-vibe-lab-adversarial-review-post-fix-2026-06-15.md`  

---

## Executive summary

**Verdict:** the current working tree does **not** preserve the prior exact-baseline parity claim.

### Score

- **Alignment:** **88%**
- **Confidence:** **94%**

### Why it regressed

Most baseline-facing surfaces still match vibe-lab exactly:
- `.claude/agents`
- `.opencode/agents`
- `.github/agents`
- `bin/doctor`
- `bin/inject`
- `opencode.json`

But the OpenCode hook/plugin surface is now split across two incompatible truths:
1. checked-in generated output still uses baseline `vibe-lab-hooks.js`
2. library source, curation, and adapter tests now expect `lazyai-hooks.js` / `LazyAIHooks`

That means future installs no longer target the same baseline surface the repo currently claims to match.

---

## Observed evidence

### Checks that still pass

- `cd packages/cli && go test ./... -count=1` → `35 packages ok, 10 no tests`
- `tests/scripts/drift-scan.sh` → `drift-scan: clean (no banned tokens in active scope)`
- `diff -ru /Users/ricardo/code/vibe-lab/.claude/agents .claude/agents` → no output
- `diff -ru /Users/ricardo/code/vibe-lab/.opencode/agents .opencode/agents` → no output
- `diff -ru /Users/ricardo/code/vibe-lab/.github/agents .github/agents` → no output
- `diff -q /Users/ricardo/code/vibe-lab/bin/doctor bin/doctor` → no output
- `diff -q /Users/ricardo/code/vibe-lab/bin/inject bin/inject` → no output
- `diff -u /Users/ricardo/code/vibe-lab/opencode.json opencode.json` → no output
- `.github/agents` contains `7` `.agent.md` files and `0` `.agent.yaml` files

### Checks that contradict exact-baseline parity

1. **Checked-in OpenCode plugin output no longer matches the baseline byte-for-byte**
   - `diff -ru /Users/ricardo/code/vibe-lab/.opencode/plugins .opencode/plugins` reports a content diff in `.opencode/plugins/vibe-lab-hooks.js`
   - observed diff: `Startup policy hash` comment differs
   - checksums:
     - baseline `vibe-lab-hooks.js`: `6128984b2e409f99cd030d7ab4394ca1b9de7a497ce6b665ce65049ee6d1065f`
     - repo `.opencode/plugins/vibe-lab-hooks.js`: `16764d8fb4f24bc5930c7aeb3578e6665c21388c086b35b4c987e1de5e42fc84`

2. **Future OpenCode installs now target a non-baseline plugin name/export**
   - `packages/cli/library/opencode/plugins` contains only `lazyai-hooks.js`
   - `packages/cli/library/manifests/curation.yaml:965-973` now curates `packages/cli/library/opencode/plugins/lazyai-hooks.js` with `LazyAIHooks`
   - `packages/cli/internal/adapter/adapter_test.go:147-148` fixture now uses `opencode/plugins/lazyai-hooks.js`
   - `packages/cli/internal/adapter/adapter_test.go:549-555` now asserts emitted `lazyai-hooks.js` contains `LazyAIHooks`
   - `cd packages/cli && go test ./internal/adapter -run TestOpenCodeAdapter_Install_CopiesHookPlugin -count=1 -v` → pass
   - prior exact-baseline report required the opposite contract:
     - `.opencode/plugins/vibe-lab-hooks.js`
     - export `VibeLabHooks`

3. **The drift guard now conflicts with the exact-baseline target**
   - `tests/scripts/drift-scan.sh:37-45` bans both `vibe-lab-hooks` and `VibeLabHooks`
   - `docs/reports/lazyai-vibe-lab-adversarial-review-post-fix-2026-06-15.md:241-247,274-275` states exact-baseline parity requires `.opencode/plugins/vibe-lab-hooks.js` exporting `VibeLabHooks`
   - Result: the active CI-style guard currently pushes the repo away from the documented baseline target

### Working tree state during review

`git status --short`:
- `M Makefile`
- `M docs/concepts/library-manifests.md`
- `M docs/concepts/tools.md`
- `M packages/cli/internal/adapter/adapter_test.go`
- `M packages/cli/library/manifests/curation.yaml`
- `D packages/cli/library/opencode/plugins/vibe-lab-hooks.js`
- `?? packages/cli/library/opencode/plugins/lazyai-hooks.js`
- plus local SQLite sidecar deletions: `.ai-setup.db-shm`, `.ai-setup.db-wal`

---

## Findings

### 1. Split-brain OpenCode plugin contract

**Status:** contradicted

The repo is currently saying two different things:
- checked-in baseline-facing output still centers `vibe-lab-hooks.js`
- source-of-truth install assets now center `lazyai-hooks.js`

This is not a documentation nit. It means the next generated install can drift away from the currently reviewed default surface.

### 2. Exact-baseline claim is no longer publishable for the working tree

**Status:** contradicted

The prior report claimed exact parity for baseline-facing default surfaces. That claim does not survive the current working tree because:
- the committed `.opencode/plugins/vibe-lab-hooks.js` already differs by checksum from the baseline
- the source library no longer emits the same filename/export as the baseline

### 3. Guardrails are enforcing the wrong direction

**Status:** contradicted

The drift scan is now green only because the source library was renamed away from the baseline token set. As written, the guardrail cannot coexist with the previous exact-parity requirement for the OpenCode plugin surface.

---

## Smallest safe recommendation

If the target remains **exact vibe-lab baseline parity**, do not publish the prior exact-parity conclusion against the current working tree.

Smallest safe path:
1. Restore a single OpenCode truth: baseline `vibe-lab-hooks.js` + `VibeLabHooks` if exact parity still wins.
2. Align `tests/scripts/drift-scan.sh` with that decision so the guardrail stops fighting the target.
3. Regenerate/re-verify `.opencode/plugins` and rerun the baseline diffs.

If the target has changed from exact-baseline parity to LazyAI-native naming, then the docs and prior post-fix report must be rewritten to stop claiming exact vibe-lab parity.

---

## Final verdict

**Current working tree:** **not aligned enough to honestly claim exact vibe-lab baseline parity.**

Fair synthesis:
- **88% alignment**
- **94% confidence**

The core agent/bin surfaces remain strong. The blocker is concentrated and concrete: OpenCode plugin naming/content drift plus a drift guard that now rewards divergence from the documented baseline target.
