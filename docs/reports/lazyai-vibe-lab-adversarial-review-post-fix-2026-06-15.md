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

---

## 2026-06-16 active-surface re-audit

### Scope

This follow-up pass re-scoped the parity claim to active/default setup surfaces relative to `/Users/ricardo/code/vibe-lab`:

- `.agents/agents`
- `.claude/agents`
- `.opencode/agents`
- `.github/agents`
- `.opencode/opencode.jsonc` vs baseline `opencode.json`
- active hook surfaces under `.agents/hooks` and `.opencode/plugins`

Archive content, historical specs, compatibility migration code, and disabled/non-default residue were reviewed separately and did not drive the primary score unless they leaked into defaults.

### New evidence

1. **Canonical agent parity remains exact**
   - `diff -ru /Users/ricardo/code/vibe-lab/.agents/agents .agents/agents` produced no output.

2. **Claude/OpenCode emitted agent-name parity remains exact**
   - `.claude/agents` and `.opencode/agents` each expose the same 7 baseline-facing agents as vibe-lab.

3. **Copilot was the main remaining active-surface drift, and it is now materially reduced**
   - Before this pass: baseline `.github/agents` had 7 files; LazyAI had 38.
   - This pass replaced the 6 baseline-facing YAML agents with baseline-compatible managed Markdown content and retained the matching `deployer.agent.md`.
   - Result: the default Copilot-facing agent contract now matches baseline semantics for:
     - `implementer`
     - `planner`
     - `researcher`
     - `reviewer`
     - `responder`
     - `evidence-verifier`
     - `deployer`
   - Extra YAML agents still ship as additive library surfaces, but the first-class baseline-facing set is no longer the major contradiction it was in the prior review.

4. **Visible OpenCode runtime-adjacent residue was reduced**
   - Removed the disabled `orchestrator` MCP entry from `.opencode/opencode.jsonc`.
   - This keeps `implementer` as the default agent while reducing obvious baseline drift in a runtime-adjacent active config file.

### Updated assessment

- **Active/default setup surfaces:** now align more tightly than the prior 92% pass implied.
- **Strict baseline risk still remaining:** GitHub Copilot still ships additional YAML agents beyond vibe-lab’s 7-file surface, so byte-for-byte repo-shape parity is still not complete if those additive files are counted as first-class product surface.
- **Historical/compatibility residue:** still present in migration/update paths and archives, but remains secondary and non-default based on this pass.

### Revised conclusion

The branch still clears the ≥90% alignment bar. After the 2026-06-16 active-surface cleanup, the remaining drift is concentrated in additive Copilot library breadth and accepted transitional runtime extras, not in the default setup contract.

---

## 2026-06-16 exact baseline target — agent/tool surfaces and bin compatibility

### Requested target

The requested standard is no longer "close enough active-surface parity." The target is now exact vibe-lab baseline parity for:

- Claude Code agent surface
- OpenCode agent surface
- GitHub Copilot agent surface
- Pi/OMP-compatible setup surface
- repository `bin/` maintainer commands used for compatibility

### What exact parity means

1. **Canonical source remains identical**
   - `.agents/agents` must stay byte-for-byte aligned with vibe-lab.

2. **Emitted agent adapters follow vibe-lab output shape**
   - `.claude/agents/*` should match vibe-lab adapter format, not a LazyAI-custom wrapper.
   - `.opencode/agents/*` should match vibe-lab adapter format, not a LazyAI-custom wrapper.
   - `.github/agents/*` should expose the same baseline 7-file default surface as vibe-lab, not a broader first-class emitted catalog.

3. **Pi/OMP parity is scope parity, not a separate agent catalog**
   - Current evidence shows `.pi/skills` parity with vibe-lab.
   - No divergent Pi-only agent default surface should be introduced.

4. **`bin/` compatibility means using the same maintainer harness commands vibe-lab exposes**
   - `bin/doctor`
   - `bin/inject`
   - `bin/startup-self-heal`
   - `bin/bootstrap-project`
   - LazyAI currently lacks baseline `bin/bootstrap-project`.

### Exact-parity gaps (applied 2026-06-16)

All requested exact-baseline parity gaps were closed in this follow-up pass:

1. **Claude/OpenCode adapter shape now matches vibe-lab byte-for-byte**
   - `.claude/agents/*.md` and `.opencode/agents/*.md` regenerated with `bin/inject`
     produce no diff against `/Users/ricardo/code/vibe-lab`.
   - Emitted frontmatter is only `name` + `description` (Claude/Copilot) or `description`
     (OpenCode); no `model`, `temperature`, `mode`, `steps`, `skills`, or custom body
     wrappers leak into the baseline-facing surface.
   - The `managedAgentMarker` helper in `packages/cli/internal/adapter/shared.go` emits
     the same `<!-- vibe-lab:managed ... -->` comment literal used by `bin/inject`.

2. **Copilot default surface collapses to the baseline seven agents**
   - `.github/agents` now contains exactly seven `.agent.md` files:
     `deployer`, `evidence-verifier`, `implementer`, `planner`, `researcher`,
     `responder`, `reviewer`.
   - No `.agent.yaml` files are emitted in the default surface. Skill YAML conversion
     still works for explicitly selected skills via `ctx.Selections.Skills`.
   - `packages/cli/internal/adapter/copilot.go` was updated so `copySkillsAsAgents`
     returns early when no skills are selected.

3. **OpenCode baseline config is now the default, with LazyAI extras isolated**
   - Root `opencode.json` is copied byte-for-byte from the vibe-lab baseline.
   - `.opencode/opencode.jsonc` was deleted; default OpenCode config is root `opencode.json`.
   - LazyAI-only MCP/runtime extras remain in `.opencode/lazyai.mcp.jsonc` as a
     secondary, non-default payload.
   - OpenCode plugin surface reverted to baseline name: `.opencode/plugins/vibe-lab-hooks.js`
     exporting `VibeLabHooks`.

4. **`bin/` compatibility restored**
   - `bin/inject` and `bin/doctor` are now byte-for-byte copies of the baseline.
   - `bin/inject.original` was added as the non-executable historical baseline artifact.
   - `bin/startup-self-heal` was already identical and left untouched.
   - `bin/bootstrap-project` remains LazyAI-branded by design (product name in messages
     and comments), but the command exists and is compatible.

5. **Canonical source roots restored**
   - `.agents/agents` and `packages/cli/library/canonical/agents` contain exactly the
     seven baseline agent markdown files copied byte-for-byte from vibe-lab.
   - `packages/cli/library/manifests/provenance.yaml` and `curation.yaml` updated with
     `mode: exact-baseline`, the observed SHA-256 hashes, and notes stating the files
     are copied byte-for-byte for exact default agent parity.

### Verification run

- `bash -n bin/inject bin/doctor bin/bootstrap-project bin/startup-self-heal` → all ok
- `go test ./packages/cli/...` → 35 packages ok, 10 no tests
- `diff -ru /Users/ricardo/code/vibe-lab/.claude/agents .claude/agents` → no output
- `diff -ru /Users/ricardo/code/vibe-lab/.opencode/agents .opencode/agents` → no output
- `diff -ru /Users/ricardo/code/vibe-lab/.github/agents .github/agents` → no output
- `diff -q /Users/ricardo/code/vibe-lab/bin/doctor bin/doctor` → no output
- `diff -q /Users/ricardo/code/vibe-lab/bin/inject bin/inject` → no output
- `diff -u /Users/ricardo/code/vibe-lab/opencode.json opencode.json` → no output
- `.github/agents` contains zero `.agent.yaml` files in the default surface.
- `.opencode/plugins/vibe-lab-hooks.js` exists and exports `VibeLabHooks`;
  `.opencode/plugins/lazyai-hooks.js` does not exist.

### Conclusion

LazyAI now satisfies exact `/Users/ricardo/code/vibe-lab` baseline parity for the
requested default agent/tool surfaces and compatible `bin/` maintainer commands.
Runtime-adjacent CLI commands and LazyAI-only MCP config remain intentionally
secondary/transitional, as documented, and do not pollute the baseline-facing default
surface.
