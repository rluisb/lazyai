# Interactive Diff Review (RPI UX)

## Goal
Improve the UI/UX for the Read-Print-Iterate (RPI) and AI-Orchestrator workflows when presenting diffs to users. Currently, presenting large files or multiple file diffs simultaneously results in "infinite scrolls," making it difficult for the user to understand, review, and approve changes.

## Current State
- **Go Implementation**: `packages/ai-setup-go/tui/diffviewer/viewer.go` provides a robust, interactive side-by-side diff viewer built with Bubble Tea for resolving template conflicts.
- **TS Implementation**: `packages/ai-setup-ts/src/utils/diff.ts` and `renderDiffPreview` print diffs directly to the terminal stdout.
- **Orchestration / RPI**: When an AI agent proposes changes across multiple files or large code blocks, the diffs are logged all at once. This leads to terminal flooding.

## Proposed UX
Instead of dumping all diffs at once, the system will switch to a **Paginated / Step-by-Step Interactive Review** when reviewing changes:

1. **Threshold Detection**: 
   - If changes are small (e.g., 1 file, < 20 lines), a standard inline diff can still be printed for speed.
   - If changes exceed the threshold (multiple files or large diffs), the interactive diff viewer is triggered.

2. **One-by-One Review Flow**:
   - The user is presented with the diff for **File 1**.
   - **Hunk Navigation**: The user can navigate back and forth between individual diff items (hunks/changes) within the file, quickly jumping between changes instead of just line-by-line scrolling.
   - The user chooses an action per file: `[a] Accept`, `[d] Deny/Reject`, `[s] Skip/Next`, `[p] Previous File`, `[q] Quit/Abort`.
   - The UI supports moving backward to re-evaluate a previous file if needed.

3. **Final Review Screen**:
   - Once all individual files have been processed, the user is presented with a **Final Review Summary**.
   - This screen lists all files and the user's chosen resolutions (Accepted, Denied, Skipped).
   - The user confirms the final batch application of the accepted changes before any writes or commits occur.

4. **Architecture Choice:**
   - We will extract and implement the diffviewer as a **standalone Go package** in the monorepo (`packages/diffviewer`).
   - This package will provide a Bubble Tea-based TUI library and a standalone CLI binary.
   - `ai-setup-go` can import this package to use the TUI components natively or call the CLI.
   - `ai-setup-ts` will delegate to this Go binary via IPC, passing a JSON input describing the diffs and awaiting the JSON output with user decisions.
   - We will **skip implementing this in the MCP `orchestrator` package** for now, leaving it specifically for the `ai-setup-go` and `ai-setup-ts` CLI tools.

## Requirements
- **No Infinite Scrolling**: Diffs must be contained within the terminal viewport height, allowing manual scrolling and hunk-jumping.
- **Granular Control**: User must be able to approve or reject changes per file, rather than all-or-nothing for the entire task.
- **Navigation**: Ability to step backward to a previous file's diff and jump between hunks inside a diff.
- **Final Confirmation**: A summary screen at the end showing all pending actions for final confirmation before applying.
- **Clear Context**: Each diff view must clearly indicate what file is being reviewed, the progress (e.g., "File 2 of 5"), and the stats (+X, -Y).

## Approved Task Breakdown

| ID | Title | Phase | Depends On |
|---|---|---|---|
| T001 | Define diffviewer JSON v1 contract | P1 - Contracts | — |
| T002 | Create standalone Go module scaffold | P1 - Contracts | T001 |
| T003 | Port reusable diff and viewer foundations | P2 - Go package | T002 |
| T004 | Add review request/response model types | P2 - Go package | T001, T002 |
| T005 | Refactor viewer state for revisitable file decisions | P2 - Go package | T003, T004 |
| T006 | Implement hunk navigation | P2 - Go package | T005 |
| T007 | Implement final summary and confirmation state | P2 - Go package | T005, T006 |
| T008 | Build diffviewer CLI wrapper | P3 - CLI wrapper | T004, T007 |
| T009 | Replace old Go diffviewer imports with delegation seam | P4 - Go integration | T008 |
| T010 | Add Go fallback inline review path | P4 - Go integration | T009 |
| T011 | Add TS diffviewer delegation utility | P4 - TS integration | T001, T008 |
| T012 | Integrate TS delegation into Phase 3 conflicts | P4 - TS integration | T011 |
| T013 | Add TS delegation and fallback tests | P5 - Verification | T011, T012 |
| T014 | Add Go package model and CLI tests | P5 - Verification | T004, T007, T008 |
| T015 | Add Go wizard integration tests | P5 - Verification | T009, T010 |
| T016 | Remove or deprecate old Go viewer package | P6 - Cleanup | T009, T015 |
| T017 | Document quickstart and run full verification | P6 - Cleanup | T013, T014, T015, T016 |

## Out of Scope
- MCP orchestrator package integration
- Actually applying file writes, commits, or changing migration executor semantics
- Remote/web diff review
- Configurable threshold flags or environment settings
- Replacing the diff algorithm with a third-party engine
- Publishing/prebuilding binaries for all platforms
- Database/store schema changes
- Changing non-conflict wizard phases