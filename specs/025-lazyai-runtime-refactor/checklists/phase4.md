# Checklist: Phase 4 Handoff Implementation

**Phase:** Phase 4  
**Exit:** Runtime writes handoffs matching the approved markdown schema.

## Preconditions

- [x] Phase 3 migration and restore evidence is recorded.
- [x] Human gate approves Phase 3 evidence.
- [x] V2 `handoff` table exists and is available to session-close integration.
- [x] Handoff schema in `schema-handoff.md` is amended to reflect the single-file writer/reader implementation.

## RED Tests

- [x] Writer test failed before `packages/cli/internal/handoff/writer.go` existed.
- [x] Round-trip parser test failed before handoff frontmatter/sections were emitted.
- [x] Repeated close test failed before atomic replace/update behavior existed.
- [x] Session-close integration test failed before handoff writing was wired.

## Implementation Checks

- [x] Handoff path follows `specs/memory/handoffs/YYYY-MM-DD-[topic].md`.
- [x] Required frontmatter keys are emitted: `goal`, `constraints`, `progress`, `decisions`, `critical_context`, `next_steps`.
- [x] Required sections are emitted: Goal, Constraints & Preferences, Progress, Key Decisions, Critical Context, Next Steps, Open Assumptions/Questions, Risks/Watchouts.
- [x] Writer performs atomic replace/update for repeated closes of the same session.
- [x] Writer does not append duplicate sections mid-session.
- [x] V2 handoff metadata stores `session_id`, `path`, `created_at`, and `status`.
- [x] Session-close lifecycle writes one handoff per session end.

## Verification

- [x] `go test ./packages/cli/internal/handoff -count=1` passes.
- [x] Session-close integration test passes.
- [x] Repeated close replacement test passes.
- [x] Handoff markdown can be read back and parsed with all required fields.


**Observed verification commands**

- `go test ./packages/cli/internal/handoff ./packages/cli/cmd -run 'Handoff|SessionEndWritesHandoff' -count=1` — passed after implementation; captured RED failure before implementation.
- `go test ./packages/cli/internal/handoff -count=1` — passed.
- `go test ./packages/cli/...` — passed.
- `go build ./packages/cli/...` — passed.

## Rollback Record

- [x] Equivalent tracked rollback point exists: commit `98e83d24cf5aefb64a51b3a6e3370bfac5369d4d` is the coarse pre-implementation reference until a human-reviewed Phase 4 commit/tag exists.
- [x] Reverting handoff integration restores previous session-close behavior via the updated developer rollback command in `rollback.md`.
- [x] Phase 4 verification output is recorded.


**Rollback boundary note:** no clean tracked Phase 4 commit/tag exists yet. Use coarse tracked point `98e83d24cf5aefb64a51b3a6e3370bfac5369d4d` until a human-reviewed Phase 4 commit/tag is created.

**RED evidence note:** pre-implementation RED coverage was captured with `go test ./packages/cli/internal/handoff ./packages/cli/cmd -run 'Handoff|SessionEndWritesHandoff' -count=1`, which failed on missing `Document` / `Write` / `Read` symbols before `writer.go` and session-close wiring existed.

## Gate

- [x] Human approves handoff round-trip evidence before Phase 5 library curation begins.
