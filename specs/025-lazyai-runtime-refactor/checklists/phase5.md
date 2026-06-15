# Checklist: Phase 5 Canonical Library and Token-Rent Enforcement

**Phase:** Phase 5  
**Exit:** Canonical library stays within 50KB or fails with documented override instructions.

## Preconditions

- [x] Phase 2 Fortnite archive/removal is complete.
- [x] Phase 4 handoff command behavior is implemented if canonical `handoff` command depends on it.
- [x] Human gate approves Phase 4 handoff evidence.
- [x] Canonical inventory in `library-canonical.md` is unchanged or explicitly amended.

## RED Tests

- [ ] Canonical library tests fail before required canonical agents/skills/hooks/commands are present.
- [ ] Embed tests fail before `all:canonical` is wired.
- [ ] Token-rent over-budget test fails before enforcement exists.
- [ ] Invalid override test fails before `reason:` validation exists.

## Implementation Checks

- [x] Canonical agents exist: `primary-agent`, `builder`, `planner`, `reviewer`, `scout`.
- [x] Canonical skills exist: `codebase-exploration`, `test-first-change`, `diagnose`, `pr-review`.
- [x] Canonical hooks exist: `session-start`, `pre-commit`.
- [x] Canonical commands exist: `graphify`, `handoff`.
- [x] `packages/cli/library/embed.go` embeds canonical assets.
- [x] Adapter file-generation code reads canonical assets.
- [x] Active embed path does not include Fortnite-only library content.
- [x] Token-rent byte count excludes non-content files such as `.gitkeep`.
- [x] `.lazyai/token-rent-override` requires a non-empty `reason:` field.
- [x] CI and pre-commit emit: `Library budget exceeded: X / 50000 bytes. Override: add .lazyai/token-rent-override with justification.`

## Verification

- [x] Under-budget canonical library passes token-rent check.
- [x] Over-budget canonical library fails token-rent check.
- [x] Valid override passes token-rent check.
- [x] Missing/empty override reason fails token-rent check.
- [x] `go test ./packages/cli/...` passes.
- [x] Root `go build ./...` was rechecked and remains invalid in the current `go.work` layout; `go build ./packages/cli/...` passes.
- [x] Token-rent CI workflow is syntactically valid.
- [x] Pre-commit hook runs the same budget logic as CI.

## Rollback Record

- [x] Equivalent tracked rollback point exists: commit `98e83d24cf5aefb64a51b3a6e3370bfac5369d4d` is the coarse pre-implementation reference until a human-reviewed Phase 5 commit/tag exists.
- [x] Override path is documented and verified.
- [x] Phase 5 verification output is recorded.

**RED evidence note:** no separate pre-green capture was recorded for the canonical/embed/token-rent tests before implementation. The implemented regression suite now covers the required over-budget, invalid override, embed, and canonical-library scenarios.

**Observed verification commands**

- `go test ./packages/cli/internal/tokenrent -run 'TestCheckPassesUnderBudget|TestCheckFailsOverBudgetWithoutOverride|TestCheckPassesWithValidOverride|TestCheckFailsWithInvalidOverrideReason|TestCheckExcludesGitkeepFromBudget|TestCurrentCanonicalLibraryWithinBudget' -count=1` — passed.
- `go run ./packages/cli/internal/tokenrent/cmd/token-rent-check` — passed with `Canonical library budget OK: 8448 / 50000 bytes`.
- `.githooks/pre-commit` — passed and reused the same token-rent command as CI.
- `ruby -e 'require \"yaml\"; YAML.load_file(\".github/workflows/token-rent.yml\"); puts \"ok\"'` — passed.
- `go test ./packages/cli/internal/tokenrent ./packages/cli/internal/adapter ./packages/cli/internal/library ./packages/cli/internal/compiler ./packages/cli/internal/types ./packages/cli/internal/db ./packages/cli/internal/setup ./packages/cli/tui/wizard ./packages/cli/cmd -count=1` — passed.
- `go test ./packages/cli/...` — passed.
- `go build ./packages/cli/...` — passed.
- `go build ./...` — still fails in the current `go.work` layout with `directory prefix . does not contain modules listed in go.work or their selected dependencies`.

## Completion Gate

- [ ] Human approves token-rent CI evidence.
- [x] Rollback records exist for every destructive phase.
- [ ] Refactor is declared complete only after all phase gates are checked.
