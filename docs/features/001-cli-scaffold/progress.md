# Progress: 001-cli-scaffold

**Last Updated:** 2026-03-30
**Phase:** 4 (MVP Complete)
**Overall:** 12/19 tasks complete

---

## Status

| Task | Name | Status | Notes |
|------|------|--------|-------|
| T001 | Init Project | ✅ DONE | Built with tsup, strict TS |
| T002 | CLI Entry | ✅ DONE | Commander stubbed out |
| T003 | Prompts | ✅ DONE | @clack/prompts interactive + flags |
| T004 | Lib: Agents | ✅ DONE | Copied scout, planner, etc. |
| T005 | Lib: Templates/Rules | ✅ DONE | Templates and context files |
| T006 | Lib: Prompts/Infra | ✅ DONE | Prompts and git hooks |
| T007 | AGENTS.md Template | ✅ DONE | Root template with placeholders |
| T008 | File Utils | ✅ DONE | Typed fs wrappers |
| T009 | Scaffold Engine | ✅ DONE | docs/ layout + shared files |
| T010 | Pi Adapter | ✅ DONE | Configured for .pi/ |
| T011 | OpenCode Adapter | ✅ DONE | Configured for .opencode/ |
| T012 | Wire Adapters | ✅ DONE | Setup config drives adapter run |
| T013 | Command: Add | ⏳ TODO | Phase 5 |
| T014 | Command: Update | ⏳ TODO | Phase 5 |
| T015 | Command: Doctor | ⏳ TODO | Phase 5 |
| T016 | Unit Tests | ⏳ TODO | Phase 6 |
| T017 | Integration Tests | ⏳ TODO | Phase 6 |
| T018 | Smoke Tests | ⏳ TODO | Phase 6 |
| T019 | Write README | ⏳ TODO | Phase 6 |

## Current Focus

MVP implementation (T001-T012) is complete and tested successfully in a temporary directory.

## Next Step

Begin Phase 5: Implement `add`, `update`, and `doctor` commands (T013-T015) or move to testing.

## Blockers

None.

## Decisions Made

- 2026-03-30: Extracted adapter interface into a dedicated types file to allow parallel development of Pi and OpenCode adapters.
