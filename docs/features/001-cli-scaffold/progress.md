# Progress: 001-cli-scaffold

**Last Updated:** 2026-03-30
**Phase:** 6 (Complete)
**Overall:** 19/19 tasks complete

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
| T013 | Command: Add | ✅ DONE | Phase 5 complete |
| T014 | Command: Update | ✅ DONE | Smart refresh with hash checks |
| T015 | Command: Doctor | ✅ DONE | Integrity verification + non-zero exit |
| T016 | Unit Tests | ✅ DONE | Adapter + file utils coverage |
| T017 | Integration Tests | ✅ DONE | Full non-interactive init path verified |
| T018 | README | ✅ DONE | Usage, commands, contribution guide |
| T019 | ADR-001 | ✅ DONE | TypeScript + @clack decision recorded |

## Current Focus

All planned tasks (T001-T019) are complete. Phase 5 added add/update/doctor commands. Phase 6 added unit/integration coverage, README documentation, and ADR-001.

## Next Step

Project is implementation-complete. Optional next work: broaden CLI negative-path coverage, add tool-matrix integration tests, or begin a new feature.

## Blockers

None.

## Decisions Made

- 2026-03-30: Extracted adapter interface into a dedicated types file to allow parallel development of Pi and OpenCode adapters.
- 2026-03-30: Chose TypeScript + @clack/prompts + commander for the CLI; documented in ADR-001.
- 2026-03-30: Implemented hash-based update/doctor flows using `.ai-setup.json` as the managed file inventory.
