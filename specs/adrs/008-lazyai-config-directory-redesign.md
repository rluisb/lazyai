# ADR-008: LazyAI Config Directory Redesign — Positional Discovery, Field-Level Merge, Registry Deletion

**Date:** 2026-07-01
**Status:** Accepted — human approved 2026-07-01, proceeding to Implementation
**Deciders:** LazyAI maintainers
**Constitution:** [`.specify/memory/constitution.md`](../../.specify/memory/constitution.md)

> **Purpose.** Record why LazyAI's operational sidecar config (`docs_dir`/`specs_dir`/`plans_dir` overrides) moves from a central `~/.lazyai/workspaces.yaml` registry with an `active`-name pointer to positional discovery of `.lazyai/sidecar.yaml` files by directory walk-up, why resolution becomes a genuine per-field merge instead of a whole-struct overwrite, and why the registry-native CLI surface (`sidecar attach`/`detach`, the entire `workspace` command group) is deleted outright rather than reinterpreted.

---

## Context

Issue [#578](https://github.com/rluisb/lazyai/issues/578) flagged that `sidecar init` silently writes a flat `<repo>/.lazyai-sidecar.yaml` into the current project with no warning when no workspace is active — coupling operational state to a directory a user may not want polluted, and doing so invisibly. Issue [#579](https://github.com/rluisb/lazyai/issues/579) (which supersedes #578 and is the vehicle for this ADR) generalizes the fix: give LazyAI's operational config one predictable home (`.lazyai/sidecar.yaml`) at every scope, and replace the central "active workspace" registry with directory-position discovery, matching the mental model already used by `git` (`.git/`), `.vscode/`, and `.idea/`.

Research (`specs/refactors/579-lazyai-config-directory/research.md`) found the registry/active-pointer idiom more pervasive than either issue anticipated:

- **6 CLI commands** are 100% registry-native with zero positional-discovery code: `sidecar attach`, `sidecar detach`, `workspace add`, `workspace switch`, `workspace list`, `workspace status` (research.md:16, :64).
- **Resolution today is whole-tuple overwrite, not layering.** `resolver.go`'s `apply()` (resolver.go:32-44) does `result = resolved` — "precedence" is achieved purely by *order of `apply()` calls*, not per-field merge (research.md:55).
- **`determineScope`** (`cmd/sidecar.go:492-509`) defaults to `ScopeProject` whenever no workspace is active, and `runSidecarInit` writes into the repo with no warning — confirming #578's exact complaint (research.md:65).
- **Zero ancestor-walking/positional-discovery code exists anywhere in the repo today** (grep-confirmed independently by 3 research subagents; research.md:21, :60). This is net-new implementation, not an extension of existing logic.
- **~35 of ~68 existing tests** across `internal/sidecar/*_test.go` and `cmd/{sidecar,workspace}_test.go` are registry- or precedence-order-coupled and will break or need deletion (research.md:18, :73-74).
- **One "Approved" governing spec**, `specs/001-internal-sidecar/SPEC.md` (v1.1.0, 2026-06-17), codifies today's registry model — including "Workspace scope is the primary configuration surface" and workspace-over-project precedence — as normative acceptance criteria (SPEC.md:20,25). It must be formally superseded, not quietly diverged from.

The standard doc (`specs/standards/architecture/lazyai-config-directory.md`, Status: Draft) fixes the target shape — `.lazyai/sidecar.yaml` at exactly three scopes (global/workspace/project), layered resolution with global always the base and project > workspace > global precedence per setting — but leaves three questions unresolved that this ADR exists to close (research.md's "Open Questions," items 1–3, all ranked **[BLOCKING]**):

1. When multiple ancestor directories each contain a `.lazyai/`, which one is "the workspace"?
2. Is per-field merge actually required, or is reordering the existing whole-tuple-replace mechanism sufficient?
3. Does the registry-native command surface (`sidecar attach`/`detach`, `workspace {add,switch,list,status}`) survive in any form?

These are irreversible-ish design calls with CLI, on-disk-format, and test-suite consequences — exactly the class of decision `specs/refactors/AGENTS.md` requires an ADR for.

**Related artifacts:**
- Research: [`specs/refactors/579-lazyai-config-directory/research.md`](../refactors/579-lazyai-config-directory/research.md)
- Standard: [`specs/standards/architecture/lazyai-config-directory.md`](../standards/architecture/lazyai-config-directory.md)
- Spec: [`specs/refactors/579-lazyai-config-directory/spec.md`](../refactors/579-lazyai-config-directory/spec.md)
- Plan: [`specs/refactors/579-lazyai-config-directory/plan.md`](../refactors/579-lazyai-config-directory/plan.md)
- Issues: [#579](https://github.com/rluisb/lazyai/issues/579) (supersedes [#578](https://github.com/rluisb/lazyai/issues/578))

---

## Constitution Alignment

| Article | Bearing | Note |
|---|---|---|
| Article I — Library-First | neutral | Walk-up discovery and field merge are both plain stdlib (`filepath`, struct assignment); no new dependency. |
| Article II — Test-First | bears | Plan phase mandates new fixtures for multi-level nesting before implementation, since research.md confirms no existing test exercises this (research.md:75). |
| Article III — Docs as Source of Truth | bears | The Draft standard doc already specifies target behavior; this ADR closes the gaps the standard left open and both must agree before implementation starts. |
| Article IV — Anti-Speculation | bears | Rejects the N-layer walk-up option specifically because the standard fixes exactly 3 scopes — this ADR does not invent capability beyond what's asked. |
| Article V — Simplicity | bears | One directory name, one filename, one resolution algorithm, no registry file, no name-based indirection — strictly fewer surfaces than today. |
| Article VI — Anti-Overengineering | bears | Rejects a generic reflection-based field-merge utility in favor of explicit per-field assignment for a 3-field struct; rejects a deprecation window in favor of a direct breaking change consistent with repo's actual release pace. |

This ADR does not amend the constitution.

---

## Options Considered (Tree of Thoughts)

### Decision 1 — Walk-up algorithm (workspace discovery)

**The question:** given `projectDir = cwd`, how do we find the single ancestor directory that counts as "workspace," when multiple ancestors could each contain a `.lazyai/`?

#### Option 1A — Continue walking past the project layer for a second, more distant layer (chosen)
- **Summary:** Check `projectDir/.lazyai/sidecar.yaml` first (the "project" layer, if present). Then walk starting at `filepath.Dir(projectDir)` — never re-checking `projectDir` itself — one directory at a time toward the filesystem root, checking each for `.lazyai/sidecar.yaml`. The walk must stop **before** reaching `$HOME` (exclusive) or the filesystem root (inclusive), whichever comes first. The first hit found during that walk is the single "workspace" layer; the walk stops immediately after — it does not keep climbing for a third, even-more-distant layer.
- **Complexity:** Medium — one bounded loop, two boundary conditions (`$HOME`, fs-root), symlink resolution.
- **Reversibility:** Medium — on-disk format (`.lazyai/sidecar.yaml` per directory) is identical regardless of which discovery rule wins, so switching discovery rules later does not require a file-format migration, only a resolver-logic change.
- **Performance impact:** Negligible — walk is bounded by real filesystem depth between cwd and `$HOME`/root, typically single-digit directories; one `os.Stat` per level.
- **Team familiarity:** Medium — no existing ancestor-walk code in the repo to imitate (research.md:21), but the shape mirrors `git`'s `.git`-discovery walk, a broadly known idiom.
- **Constitution fit:** Best fit for Article III — matches the standard doc's fixed 3-scope model (global/workspace/project) exactly, and matches the worked examples in `lazyai-config-directory.md`'s "Outcomes" table (project+workspace both present is an explicit supported case, standard:50).

#### Option 1B — Nearest-ancestor-wins, stop at first `.lazyai/` hit
- **Summary:** Walk from `projectDir` upward (including `projectDir` itself) and stop at the very first `.lazyai/` found — that single hit is treated as "the" local layer, with no distinction between "project" and "workspace."
- **Complexity:** Low — simplest possible loop, no distinction logic needed.
- **Reversibility:** High — trivially generalizable to 1B→1A later since 1B is a strict subset of 1A's cases.
- **Performance impact:** Negligible, same order as 1A.
- **Team familiarity:** High — this is the naive/obvious first idea, closest to how most tools (e.g. `.editorconfig` lookup) behave.
- **Constitution fit:** Fails Article III — collapses the standard's documented "project **and** workspace both present" case (standard:50) into a single layer, silently dropping one config file's settings with no way to have both simultaneously. Contradicts the human's explicit locked-in instruction to continue past the first hit.
- **Rejected because:** it cannot produce the standard's own worked example (`global → workspace → project`, three distinct layers) when project and workspace configs coexist — it structurally allows at most one local layer, not two.

#### Option 1C — Unbounded walk collecting every `.lazyai/` from cwd to filesystem root as N layers
- **Summary:** Treat every ancestor `.lazyai/` as its own layer, in position order, with no fixed count.
- **Complexity:** High — variable-length layer list changes `ResolvedPaths`/`ConfigLevel` from a fixed enum to an open-ended provenance list, and downstream `sidecar status`/`doctor` output must handle arbitrary N.
- **Reversibility:** Low — once `ConfigLevel` becomes a list rather than one of 3 fixed values, external tooling or docs built against "exactly 3 scopes" breaks; hard to un-generalize later without a second breaking change.
- **Performance impact:** Negligible in practice (same bounded walk to `$HOME`/root) but the *output shape* complexity is unbounded.
- **Team familiarity:** Low — no precedent in the repo or the standard doc for N-way layering.
- **Constitution fit:** Fails Article IV (Anti-Speculation) — the standard doc explicitly fixes 3 scopes total (`lazyai-config-directory.md:10,42`); building N-layer support speculatively adds capability nobody asked for and the standard doesn't sanction.
- **Rejected because:** over-generalizes past a locked requirement (research.md's own framing calls this option "contradicting the doc's fixed 3-scope model," research.md:85).

**Chosen: Option 1A.** See "Decision" below for the exact algorithm and edge-case resolutions.

---

### Decision 2 — Resolution semantics (field-level merge vs. whole-tuple replace)

**The question:** when multiple layers are present, does the more specific layer replace the *entire* resolved struct, or does it override only the fields it actually sets?

#### Option 2A — Real per-field merge (chosen)
- **Summary:** For each field in `ResolvedPaths` (`DocsDir`, `SpecsDir`, `PlansDir`), scan layers in precedence order (project > workspace > global > built-in default) and take the first non-empty value found, independently per field. A layer setting only `docs_dir` inherits `specs_dir`/`plans_dir` from the next lower-precedence layer that sets them.
- **Complexity:** Medium — replaces `apply()`'s single `result = resolved` line (resolver.go:32-44) with an explicit per-field scan, but the struct only has 3 mergeable fields (plus `Path`, which isn't merged — it's the anchor), so the merge function stays small and explicit (no generic/reflection-based merge utility needed).
- **Reversibility:** Medium — reverting to whole-tuple-replace later is a pure logic change (no on-disk format impact) but would re-break any config that relies on partial overrides, which is a user-visible regression once shipped.
- **Performance impact:** Negligible — 3 field comparisons per resolution call, no measurable cost.
- **Team familiarity:** Medium — new pattern for this package, but a standard "coalesce first non-empty across ordered sources" idiom common elsewhere in Go codebases.
- **Constitution fit:** Best fit — the standard's own worked example (`lazyai-config-directory.md:117-124`, "Compliant resolution... merge(global, workspace, project)") is written as a field-level merge, and Article V favors this because it removes the surprising "setting one field wipes your other overrides" behavior a whole-struct model has.

#### Option 2B — Keep whole-tuple-replace, just reorder which layer applies last
- **Summary:** Preserve `apply()`'s `result = resolved` behavior; achieve project > workspace > global precedence purely by calling `apply()` in that literal order, same mechanism as today (research.md:55).
- **Complexity:** Low — smallest possible diff; only the *order* of existing `apply()` calls changes.
- **Reversibility:** High — trivial to swap order again.
- **Performance impact:** Negligible, identical to current code.
- **Team familiarity:** High — this is exactly today's existing mechanism (research.md confirms current "precedence" is achieved purely by call order, not real layering).
- **Constitution fit:** Fails Article III — the standard's prose reads as per-field merge and its worked example is explicitly a field-level `merge(...)` call, not "replace." `TestResolve_HigherPrioritySidecarReplacesWholeTuple` (research.md:86) locks in exactly the semantics this option preserves, which the standard's own examples contradict.
- **Rejected because:** a project-scope config that sets only `docs_dir` would silently discard a workspace-scope `specs_dir` override — directly contradicting the standard's documented behavior and the primary user-facing motivation for layering (global defaults "never lost," standard:85).

**Chosen: Option 2A.**

---

### Decision 3 — Command-surface deletion

**The question:** does the registry-native CLI surface (`sidecar attach`/`detach`, all of `workspace {add,switch,list,status}`) get deleted, kept-but-reinterpreted, or deprecated with a transition window?

#### Option 3A — Full deletion, no reinterpretation (chosen)
- **Summary:** Delete `sidecar attach`, `sidecar detach`, and the entire `workspace` command group (`add`, `switch`, `list`, `status`) outright. The `workspace` top-level command itself is removed since it would have zero subcommands left. Remaining surface: `sidecar {init, status, doctor}`. Creating a workspace- or project-scope config becomes `cd` to the intended directory and run `sidecar init` there, with scope inferred positionally (cwd is the target) and an explicit `--scope` flag available to confirm intent when ambiguous.
- **Complexity:** Medium — deletion touches `cmd/workspace.go` (full file), `cmd/workspace_test.go` (full file, all 4 tests, research.md:74), the `attach`/`detach` subtrees of `cmd/sidecar.go`, and the dead-code `requireWorkspaceFallback` parameter (research.md:66) and `WriteWorkspaceSidecar`/`UpdateWorkspaceConfig` (writer.go:14-34,46-62).
- **Reversibility:** Low — once deleted, restoring name-based registration means re-adding both the registry file format and the commands; not a config-file-compatible rollback.
- **Performance impact:** Neutral to positive — removes a locked read-modify-write registry file (`files.WithFileLock`, 5s/30s timeouts, writer.go:36) from the hot path entirely.
- **Team familiarity:** Medium — no precedent for a full command-surface deletion this size in the repo's history, but the repo has an established practice of shipping breaking changes at release-note granularity (CHANGELOG.md:163, "**Breaking:** Project renamed...").
- **Constitution fit:** Best fit for Article V — a name-based registry *and* a directory-position system are two mental models for the same concept; keeping both means users must learn which one actually took effect. One model only.

#### Option 3B — Keep the commands, reinterpret their semantics
- **Summary:** Keep `workspace add`/`switch`/`list` etc. as commands, but redefine what they do under positional discovery (e.g. `workspace add <path>` creates a `.lazyai/` there; `workspace switch` becomes a no-op or errors).
- **Complexity:** High — every reinterpreted command needs new help text, new flag semantics, and a mapping from "what used to happen" to "what happens now" that users must relearn anyway — effectively a rename exercise with all the cost of deletion and none of the clarity.
- **Reversibility:** Medium — command names survive, but their contracts change anyway, so "reversibility" here is illusory: any caller relying on old semantics still breaks.
- **Performance impact:** Neutral.
- **Team familiarity:** Low — the repo has one clean precedent for a hidden backward-compatible alias (`completions` → `completion`, `cmd/completions.go:10-18`, via Cobra's `Deprecated` field with `Hidden: true`), but that case is a pure rename with *identical* semantics — not applicable here since `workspace switch`'s entire premise (name-based activation) has no positional equivalent (research.md:74, "their entire premise... has no positional equivalent").
- **Constitution fit:** Fails Article V/VI — reinterpreting commands to mean something unrelated to their name is *more* confusing than deleting them, and risks silent behavior changes for any script currently calling them.
- **Rejected because:** there is no positional equivalent for "activate workspace X by name" (research.md:74) — reinterpretation would have to invent new semantics from nothing, which is worse than a clean deletion with a clear error message ("unknown command").

#### Option 3C — Deprecate with a transition window (hidden alias, prints warning, removed in a future release)
- **Summary:** Mirror the `completions`→`completion` pattern: keep the commands functional but `Hidden: true` with `Deprecated: "..."` messages, removing them in a subsequent minor/major release.
- **Complexity:** Medium — requires maintaining both the old registry-native code path and the new positional path simultaneously for at least one release cycle.
- **Reversibility:** High while the window is open.
- **Performance impact:** Neutral.
- **Team familiarity:** High — this is a well-known pattern and the repo has one precedent for it.
- **Constitution fit:** Weakens Article VI (Anti-Overengineering) — maintaining two parallel config models (registry + positional) simultaneously, even temporarily, reintroduces exactly the "two mental models" problem Option 3A eliminates, and the repo's actual release cadence (multiple 0.x/1.x releases with breaking changes shipped directly, e.g. the ai-setup→LazyAI rename, CHANGELOG.md:163) shows deprecation windows are not this project's established practice for structural renames.
- **Rejected because:** the registry model isn't being renamed, it's being replaced by a structurally different discovery mechanism (name lookup vs. directory position) — there is nothing meaningful to alias, and keeping the old registry file format alive during a window contradicts Decision 2's need to drop `WorkspaceConfig`/`Active` entirely to get real layering.

**Chosen: Option 3A.**

---

## Decision

**1. Walk-up algorithm** (Option 1A): up to two local layers may exist above the always-present global base.

```
projectDir := cwd
home, homeErr := os.UserHomeDir()  // resolved via filepath.EvalSymlinks if resolvable; fall back to raw path on error
resolvedProjectDir := resolveSymlinks(projectDir)  // fall back to raw projectDir on error
resolvedHome := resolveSymlinks(home)              // fall back to raw home on error

// Project layer: check cwd itself, exactly once, never re-checked during the walk below.
projectLayer := nil
if exists(join(resolvedProjectDir, ".lazyai/sidecar.yaml")):
    projectLayer := resolvedProjectDir

// Workspace layer: walk strictly ABOVE projectDir, never re-entering projectDir.
workspaceLayer := nil
current := filepath.Dir(resolvedProjectDir)
loop:
    if current == resolvedHome (when homeErr == nil):       # exclusive — never treat $HOME as workspace
        break loop
    if filepath.Dir(current) == current:                     # fs root reached (POSIX "/" or Windows drive root) — inclusive stop, root itself is still checked below then walk ends
        if exists(join(current, ".lazyai/sidecar.yaml")):
            workspaceLayer := current
        break loop
    if exists(join(current, ".lazyai/sidecar.yaml")):
        workspaceLayer := current
        break loop                                            # stop at FIRST hit — do not keep climbing for a third layer
    current := filepath.Dir(current)

// Global layer is always present (built-in defaults if ~/.lazyai/sidecar.yaml absent) — never part of this walk.
```

Edge cases resolved explicitly:
- **`cwd == $HOME`:** the project-layer check still runs against `$HOME/.lazyai/sidecar.yaml` (degenerate but harmless — it happens to be the same directory as the global layer's file, so "project" and "global" would read identical content, not double-counted as two *different* layers with conflicting precedence, since the walk-up for workspace starts at `filepath.Dir($HOME)` and never re-enters `$HOME`). Workspace walk proceeds above `$HOME` toward root exactly as any other case.
- **`cwd` outside `$HOME` entirely (e.g. `/tmp/foo`):** `$HOME` is never encountered as an ancestor of `/tmp/foo`, so the "stop before `$HOME`" condition simply never triggers; the walk runs to filesystem root unimpeded.
- **Symlinks:** resolve `cwd` and `$HOME` (and each candidate ancestor) via `filepath.EvalSymlinks` before identity comparison, falling back to the unresolved path if resolution errors (e.g. broken symlink) — prevents a symlinked cwd from either falsely matching or falsely missing the `$HOME` boundary.
- **Filesystem root / Windows drive roots:** detected portably via `filepath.Dir(x) == x` (true for POSIX `/` and Windows drive roots like `C:\`), not a hardcoded `"/"` check.

**2. Resolution semantics** (Option 2A): field-by-field merge, not whole-struct replace.

```
resolved.DocsDir  = firstNonEmpty(project.DocsDir,  workspace.DocsDir,  global.DocsDir,  builtinDefault.DocsDir)
resolved.SpecsDir = firstNonEmpty(project.SpecsDir, workspace.SpecsDir, global.SpecsDir, builtinDefault.SpecsDir)
resolved.PlansDir = firstNonEmpty(project.PlansDir, workspace.PlansDir, global.PlansDir, builtinDefault.PlansDir)
```

Precedence order is **project > workspace > global > built-in default**, applied independently per field — a layer that sets only `docs_dir` still inherits `specs_dir`/`plans_dir` from the next lower-precedence layer that sets them. This replaces `resolver.go`'s `apply()` whole-tuple overwrite (`result = resolved`) at resolver.go:32-44.

**3. Command-surface deletion** (Option 3A): `sidecar attach`, `sidecar detach`, `workspace add`, `workspace switch`, `workspace list`, `workspace status` are deleted outright, with no hidden alias and no reinterpretation. The `workspace` command group is removed entirely (zero subcommands would remain). Remaining CLI surface: `sidecar {init, status, doctor}`. `sidecar init` selects its write target positionally (cwd is the scope being initialized) with an explicit `--scope` flag to disambiguate/confirm intent (`project` vs. `workspace` vs. `global`) — no name argument, no new `--at <path>` flag (cwd-anchoring is sufficient, consistent with how `project` scope already anchors on cwd today). The `Scope` enum (`ScopeGlobal`/`ScopeWorkspace`/`ScopeProject`) survives in `types.go`, but narrows to a single purpose: parsing `sidecar init`'s `--scope` flag and feeding a new `ScopeRoot(scope Scope, cwd string) (string, error)` (replacing `scopeDefaultRoot`) that returns the home directory for `Global`, and **cwd itself for both `Project` and `Workspace`** — the two are identical write targets from the same directory; `--scope project` vs. `--scope workspace` at `init` time is a label/confirmation-message distinction only, not a different file path, because which role a given `.lazyai/sidecar.yaml` plays is determined later, at *resolve* time, by its directory position relative to whoever is resolving — not by which flag it was created with. `Resolve()`/`Doctor()` drop the `Scope` parameter entirely, becoming `Resolve(cwd string) (ResolvedPaths, error)` / `Doctor(cwd string) ([]Issue, error)` — there is one fully-layered result now, not one result per scope — and the `sidecar status`/`sidecar doctor` CLI commands drop their `--scope` flags for the same reason (always show the complete discovered picture, never a partial/depth-limited one). Discovery itself lives in a new `DiscoverLayers(cwd string) (*Layers, error)` (`internal/sidecar/discovery.go`): a "hit" requires the `.lazyai/sidecar.yaml` **file** to be present, not just a bare `.lazyai/` directory — an empty `.lazyai/` with no `sidecar.yaml` inside contributes no layer, keeping discovery and merge simple (nothing to merge from an empty directory).

**4. Migration data (workspaces.yaml):** this is a **breaking change with no automated migration.** `~/.lazyai/workspaces.yaml` and the `Active` pointer are deleted with no data-preserving upgrade path. `Doctor(cwd)` MUST detect a stale `~/.lazyai/workspaces.yaml` via a plain `os.Stat` existence check — **detection-only, no YAML parsing of the file's contents** — and surface it as a returned `Issue` carrying a one-line deprecation notice pointing the user to manually recreate any needed workspace config via `sidecar init` at the correct directory. No conversion tool reads, parses, or rewrites the old registry format.

**5. Supersession:** `specs/001-internal-sidecar/SPEC.md` (v1.1.0, Status: Approved, 2026-06-17) is **formally superseded** by this ADR. That spec's Constitution explicitly states "Workspace scope is the primary configuration surface" (SPEC.md:25) and its acceptance criteria assume the registry/`Active`-pointer model this ADR deletes. See "Related" below.

---

## Rationale

- **Walk-up (1A over 1B/1C):** the standard doc's own worked examples require project and workspace to coexist as two distinct, simultaneously-active layers (`lazyai-config-directory.md`'s outcomes table, row 3: "Project `.lazyai/` **and** workspace `.lazyai/`" → `global → workspace → project`). 1B cannot produce that outcome by construction. 1C could produce it (and more), but invents unbounded layering the standard explicitly forecloses (fixed 3 scopes) — Article IV forbids building capability nobody asked for. 1A is the minimal design that satisfies the standard's documented cases exactly.
- **Merge semantics (2A over 2B):** 2B is strictly the status quo with reordering — it does not fix the actual bug (research.md's "Bug found" at resolver.go:90 notes a related side effect already caused by whole-tuple semantics), and it directly contradicts the standard's own merge(...) example (`lazyai-config-directory.md:117-124`). Shipping 2B would mean the standard doc and the implementation disagree on day one.
- **Command deletion (3A over 3B/3C):** 3B requires inventing new meanings for verbs (`switch`, `add`) that describe a mental model (name-based registration/activation) which no longer exists; that's more confusing than an "unknown command" error. 3C's deprecation window only makes sense when the *old* and *new* behavior can coexist under the same name with compatible semantics (true for the `completions`→`completion` rename, false here) — the registry file itself must be deletable to unlock real per-field layering, so a coexistence window blocks Decision 2. The repo's own changelog shows direct breaking changes are already normal practice here (e.g. the ai-setup→LazyAI rename, CHANGELOG.md:163), so 3A is consistent with established team behavior, not a departure from it.
- **Migration (breaking change, doctor hint only):** a real migration tool would need to map `WorkspaceEntry.Path` + `Sidecar` fields onto newly-created `.lazyai/sidecar.yaml` files at those exact paths — technically possible, but speculative given research.md found `WriteWorkspaceSidecar` already has zero callers (dead code) and no test suite exercises this path today. A one-line `doctor` hint gives affected users a clear, low-cost next action ("re-run `sidecar init` at the right directory") without building and maintaining a one-time conversion tool for a config file whose schema (`WorkspaceEntry{Name, Path, Sidecar}`) is being deleted anyway. This is the "simplest" option the shared context explicitly floated, and nothing in research.md's usage evidence justifies more.
- **Supersession:** SPEC.md v1.1.0 is "Approved" and its acceptance criteria (workspace-primary, registry-backed) are the literal thing this ADR replaces; leaving it un-superseded would leave two contradictory "governing" documents active simultaneously, violating Article III (Docs as Source of Truth).

**Why the rejected options were rejected:** see the per-option "Rejected because" lines above (Options 1B, 1C, 3B, 3C) and the 2B paragraph — each ties back to either failing to reproduce a standard-doc-documented case, or over-building past a locked scope boundary (Article IV).

---

## Consequences

**Positive:**
- One discovery mechanism, one resolution algorithm, one filename at every scope — matches the standard doc exactly, closing all three of its previously-open blocking questions.
- Removes a locked, read-modify-write registry file (`files.WithFileLock`, 5s/30s timeouts) from the write path entirely — fewer failure modes (lock contention, stale lock files, corrupt registry) than today.
- Fixes the resolver.go:90 anchor bug as a side effect: a discovered ancestor `.lazyai/` becomes its own natural anchor rather than being anchored against `globalDir` (research.md:56).
- Directly satisfies #578's original complaint: project scope becomes presence-based/opt-in (no `.lazyai/` present → no repo-local write happens implicitly), with no separate #578-specific implementation needed once #579 lands (research.md:101).

**Negative / accepted trade-offs:**
- Users who relied on `sidecar attach`/`workspace switch` lose that exact workflow with no automated equivalent; they must learn the new `cd` + `sidecar init [--scope]` pattern.
- `~/.lazyai/workspaces.yaml` data is not migrated; any workspace registered there must be manually recreated by re-running `sidecar init` at the correct directory. Users get a one-line `doctor` nudge, not a guided migration.
- ~35 of ~68 existing tests break or need deletion (research.md:18, :73-74); `cmd/workspace_test.go`'s 4 tests are deleted outright since their premise no longer exists.
- `specs/001-internal-sidecar/SPEC.md` (Approved, v1.1.0) becomes stale documentation the moment this ADR is accepted, until it is marked Superseded.

**Neutral:**
- The `Scope` enum survives structurally (3 values unchanged) but its role narrows to a single purpose — parsing `sidecar init`'s `--scope` flag via `ScopeRoot(scope, cwd)`; `Resolve()`/`Doctor()` drop the parameter entirely, and `Project`/`Workspace` resolve to the identical write path (cwd) since the distinction is a label, not a location.
- `ResolvedPaths.ConfigLevel` (a single `"workspace"`/`"project"`/`"global"`/`"default"` string) is deleted, replaced by a per-field `Provenance map[string]string` (e.g. `{"docs_dir": "project", "specs_dir": "workspace", "plans_dir": "global"}`) — a public API shape change that `sidecar status`/`doctor` output must render field-by-field rather than as one scope label, since a single resolved result can now legitimately draw different fields from different layers.
- The `--path` flag on `sidecar init` (content root for docs/specs/plans) is orthogonal to this ADR and does not change (research.md:67).

---

## Reversal Conditions

Re-open this ADR if any of the following becomes true:

- Real-world usage reveals users routinely need more than 2 local layers (a scenario the standard doc's fixed 3-scope model does not anticipate) — would require revisiting Decision 1's stop-after-first-hit rule.
- Field-level merge (Decision 2) proves insufficient because a future config field has interdependent validation across layers that per-field coalescing can't express (e.g. a field whose valid values depend on another field's resolved layer) — would require a more structured merge strategy than `firstNonEmpty`.
- Deleted commands (Decision 3) turn out to have been load-bearing for an integration or script pattern discovered post-release, severe enough that a compatibility shim becomes necessary — would require reintroducing a scoped, time-boxed Option 3C instead of the clean Option 3A cutover.
- User feedback shows the no-migration breaking change (Decision 4) caused enough support burden that a real conversion tool becomes justified — would require building the `WorkspaceEntry`→`.lazyai/sidecar.yaml` migration this ADR explicitly declined to build.

---

## Implementation Pointer

- Spec: [`specs/refactors/579-lazyai-config-directory/spec.md`](../refactors/579-lazyai-config-directory/spec.md) — exact function signatures for the new walk-up discovery and field-merge resolver, and the finalized `sidecar init --scope` flag surface.
- Plan: [`specs/refactors/579-lazyai-config-directory/plan.md`](../refactors/579-lazyai-config-directory/plan.md) — phased task breakdown (positional-discovery + layered-resolve core first, landed alongside old code; CLI cutover including command deletion second; registry/dead-code removal + doc/spec supersession third), each phase leaving the codebase working per the refactor rule.
- New surfaces cited by spec.md: `DiscoverLayers(cwd string) (*Layers, error)` in a new `internal/sidecar/discovery.go` (a "hit" = `.lazyai/sidecar.yaml` file present, not a bare `.lazyai/` dir); `ScopeRoot(scope Scope, cwd string) (string, error)` replacing `scopeDefaultRoot`; `Resolve(cwd string) (ResolvedPaths, error)` and `Doctor(cwd string) ([]Issue, error)` dropping their `Scope` parameter; `ResolvedPaths.Provenance map[string]string` replacing `ResolvedPaths.ConfigLevel`.
- Primary touch points cited by research.md: `packages/cli/internal/sidecar/resolver.go` (`Resolve`, `apply`, `scopeDefaultRoot`), `packages/cli/internal/sidecar/doctor.go` (`Doctor`), `packages/cli/internal/sidecar/loader.go` (`getWorkspacesConfigPath`, `LoadWorkspaceConfig`), `packages/cli/internal/sidecar/writer.go` (`WriteWorkspaceSidecar`, `UpdateWorkspaceConfig`, `WriteProjectSidecar`), `packages/cli/cmd/sidecar.go` (`determineScope`, attach/detach subcommands), `packages/cli/cmd/workspace.go` (entire file, deleted).
- Docs to rewrite: `docs/concepts/scopes.md:100-143`, `docs/cli/reference.md:1429-1503`, `docs/getting-started/commands.md:122-129`, `README.md:139` (cosmetic).

---

## Related

- **Supersedes:** [`specs/001-internal-sidecar/SPEC.md`](../001-internal-sidecar/SPEC.md) (v1.1.0, Status: Approved, 2026-06-17) — its "Workspace scope is the primary configuration surface" constraint and registry/`Active`-pointer-based acceptance criteria are replaced by this ADR's Decisions 1–3. `SPEC.md`'s own status field should be updated to `Superseded by ADR-008` as part of this refactor's implementation (tracked as a plan task).
- **Related standard:** [`specs/standards/architecture/lazyai-config-directory.md`](../standards/architecture/lazyai-config-directory.md) — this ADR closes that standard's three previously-unresolved blocking questions (walk-up rule, merge semantics, command surface) and the standard's "Status: Draft" flips to "Active" once enforcement tests land (tracked as a plan task, per the standard's own "Enforcement" section).
- **Related issues:** [#579](https://github.com/rluisb/lazyai/issues/579) (supersedes [#578](https://github.com/rluisb/lazyai/issues/578)).
- **Prior ADRs:** none directly overlapping; this is the first ADR to touch the `internal/sidecar` package.

---

## Memory Update

- [ ] Update `specs/KNOWLEDGE_MAP.md`'s "Key Architecture Decisions" table with this ADR (done as part of this same change).
- [ ] Update `specs/001-internal-sidecar/SPEC.md`'s status header to `Superseded by ADR-008` once this ADR is accepted (tracked in plan.md).
- [ ] Flip `specs/standards/architecture/lazyai-config-directory.md`'s status from Draft to Active once enforcement tests land, recording the date + PR under that standard's "Enforcement" section (tracked in plan.md, last implementation phase).
- [ ] If future usage evidence triggers any Reversal Condition above, update or supersede this ADR with the revised decision.
