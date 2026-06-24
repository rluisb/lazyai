# Tasks: 030 — Kiro CLI v3 output gaps

## Phase 1: Verify Kiro hook contract and author source assets
- [ ] **Task 1.1**: Add a focused failing test (`internal/adapter/kiro_hooks_test.go`) that loads every file in `packages/cli/library/kiro/hooks/` and asserts the Kiro v3 schema: top-level `version: "v1"`, non-empty `hooks[]`, valid `trigger`, valid `action.type`, required `command`/`prompt` field.
- [ ] **Task 1.2**: Create `packages/cli/library/kiro/hooks/` and author the initial Kiro-native hook JSON asset(s) using only source-verified triggers.
- [ ] **Validation**: Run the focused Go test for `kiro_hooks_test.go` and show the assets parse + validate. ⛔ checkpoint

## Phase 2: Emit Kiro hooks from the adapter
- [ ] **Task 2.1**: Update `packages/cli/internal/adapter/kiro.go` to create `.kiro/hooks/` and copy selected hook assets from `library/kiro/hooks/` during `Install`.
- [ ] **Task 2.2**: Update `packages/cli/library/embed.go` to embed the new `all:kiro` asset root if needed.
- [ ] **Task 2.3**: Extend `packages/cli/internal/adapter/kiro_adapter_test.go` to assert `.kiro/hooks/<name>.json` emission and remove the existing `assertMissing(.kiro/hooks)` expectation.
- [ ] **Validation**: Run the focused Kiro adapter tests proving the hook files are emitted. ⛔ checkpoint

## Phase 3: Align capability/conformance tests
- [ ] **Task 3.1**: Flip `KiroAdapter.Capabilities().Hooks` to `true` in `capabilities.go` and update the comment to remove the stale "instruction-only" statement.
- [ ] **Task 3.2**: Update `packages/cli/internal/adapter/capabilities_test.go` to assert Kiro hooks are supported while `Specs` and `Steering` remain false.
- [ ] **Task 3.3**: Update `packages/cli/internal/adapter/adapter_adapters_test.go` to remove/replace the `assertMissing(.kiro/hooks)` expectation.
- [ ] **Validation**: Run the focused adapter capability/conformance tests. ⛔ checkpoint

## Phase 4: Regenerate goldens
- [ ] **Task 4.1**: Regenerate the `kiro-only` and `full-seven-targets` goldens so `.kiro/hooks/` appears in the expected output.
- [ ] **Task 4.2**: Inspect the golden diff and confirm only the intended Kiro hook additions (plus capability-driven incidental changes, if any).
- [ ] **Validation**: Run `TestCompilerGolden` (or its narrow package target) with and without `UPDATE_GOLDEN=true` as appropriate. ⛔ checkpoint

## Phase 5: Docs and explicit non-goals
- [ ] **Task 5.1**: Update `docs/README.md` to drop the stale Kiro hook limitation.
- [ ] **Task 5.2**: Update `docs/adapters/capability-matrix.md`, `docs/reference/hooks.md`, and `docs/reference/tool-outputs.md` to mark Kiro hooks supported and record the non-goals: specs not emitted, repo-local permissions forbidden, direct `.kiro/powers/` not emitted.
- [ ] **Task 5.3**: Keep `Permissions: true` documented as host-support metadata rather than a repo-emitted file.
- [ ] **Validation**: Run `mkdocs build --strict`. ⛔ checkpoint

## Parallelization notes
- Phases are sequential.
- Within a phase, tasks touching the **same file** are single-owner.
- Candidate `[P]` lanes if implementation is approved:
  - `[P]` Phase 1 assets + schema test (`kiro_hooks_test.go`, `library/kiro/hooks/`, `embed.go`)
  - `[P]` Phase 5 docs (`docs/README.md`, `docs/adapters/capability-matrix.md`, `docs/reference/hooks.md`, `docs/reference/tool-outputs.md`)
- Phase 2/3 Kiro adapter Go edits share files and should stay with one owner.

## Dependency order
1. 1.1 → 1.2 ✅ once schema is verified
2. 1.2 → 2.1 → 2.2 → 2.3
3. 2.3 → 3.1 → 3.2 → 3.3
4. 3.3 → 4.1 → 4.2
5. 3.3 → 5.1 → 5.2 → 5.3

---

**Human gate:** ⛔ Per repo RPI rules, stop after plan approval. Do not implement until the human explicitly approves this spec/plan/tasks set.
