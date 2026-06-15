# Retired Dashboard — Manual QA Checklist (v1)

> Historical only. The dedicated dashboard runtime was removed from the active workspace in Spec 025. This checklist applies only to archived snapshots; current verification uses CLI tests and builds.

**Issue:** https://github.com/rluisb/lazyai/issues/176
**Scope:** v1 cut shipped via PRs #180, #181, #183, #184, #185
**Use:** historical reference for archived dashboard changes.

---

## How to run

1. Build and start the daemon locally:
   ```bash
   cd packages/orchestrator
   go run ./cmd/lazyai-orchestrator serve
   ```
2. Open `http://127.0.0.1:<port>/dashboard/` (port is announced in stdout, default `:57372`).
3. Walk each section below. Tick items in your PR description or paste a copy of this file.

If you cannot reproduce a step (no failing run, no run with steps, etc.), seed data via the orchestrator MCP tools or use existing fixtures from your local SQLite DB.

---

## Load + theming

- [ ] Page loads without a console error.
- [ ] Both fonts (Fraunces serif + JetBrains Mono Nerd Font) render — headings are serif, run ids and timestamps are monospace.
- [ ] Status pill in topbar transitions from "Loading dashboard…" to "Dashboard data loaded." within ~1s.
- [ ] Open Tweaks (top-right cog) — toggling `theme` flips the palette without flicker.
- [ ] Toggling `density` resizes paddings/gaps; `comfortable` is visibly looser than `compact`.
- [ ] Toggling `nav` between `top tabs` and `sidebar` shows/hides the corresponding chrome and the other one disappears entirely (no double nav).
- [ ] Reload the page — Tweaks selections persist.

## Navigation + hash routing

- [ ] Clicking each primary nav item activates only that section (others have `hidden`).
- [ ] Browser URL hash updates (`#/overview`, `#/runs`, `#/catalog`, `#/errors`).
- [ ] Direct visit to each hash route activates the matching section on first paint.
- [ ] Browser back/forward moves between sections without page reload.
- [ ] Unknown hash (e.g. `#/nope`) falls back to Overview without a console error.
- [ ] **Run detail** nav item is disabled before any run is opened, with hint text "Select a run first".
- [ ] After opening a run from Runs, the Run detail item enables and remembers the selection across other-section navigation.
- [ ] **Planned block** items (Live activity / Logs / Settings) are visually disabled, do nothing on click, and show their gating-plan rationale text.

## Overview screen

- [ ] All four cards (Health, Active runs, Run states, Catalog) populate.
- [ ] Recent runs list renders compact rows.
- [ ] Recent errors list renders or shows "No recent errors." empty state.
- [ ] **Live activity** panel shows status "live · listening for events across all runs" within ~1s.
- [ ] When you trigger a run from another terminal/tool, new events animate into the activity feed.
- [ ] Clicking the activity pause toggle: status flips to "paused", feed border turns warning-yellow, no new events arrive. Clicking again resumes.
- [ ] Refreshing the page with the toggle paused remembers the paused state.
- [ ] A **failed** run produces a red bottom-right toast within ~1s.
- [ ] After a state-changing event (run started/completed/failed), Overview cards and Runs list refresh within ~1s.

## Runs screen

- [ ] Attention chips: clicking `running`, `failed`, `gated`, `recent`, `has errors` filters the list and updates the count summary.
- [ ] `budget` chip filters client-side using `budgetHealth` and shows the "client-side" note in the summary.
- [ ] Search box filters by id substring and definition name (case-insensitive), debounced.
- [ ] Kind dropdown narrows results.
- [ ] If `nextCursor` is present, a **Load more →** button appears and appends the next page.
- [ ] Empty filter combos render the empty state without breaking the chip bar.
- [ ] Clicking a run row opens Run detail and updates the URL to `#/runs/{kind}/{id}`.

## Run detail screen

- [ ] Hero header shows kind/id, definition + version, started + updated timestamps, state chip.
- [ ] Copy-id button copies `kind/id` to clipboard (verify by pasting elsewhere) and surfaces an "ok" status pill.
- [ ] **Budget cards** (tokens / cost / wall clock / retries) show progress bars; warning/critical states colour the border + bar.
- [ ] **Timeline** renders one node per step (chains) / task (teams) / phase (workflows). Failed nodes inline the structured error message.
- [ ] Events nest under the matching node by `stepId`/`taskId`/`phaseId`; unscoped events fall to the bottom.
- [ ] Per-run live events still arrive via the per-run SSE (open the Events collapsible, run something, see new events).
- [ ] All collapsibles open and close (Summary, Events, Budget raw, Raw state JSON, Execution plan, Handoffs).
- [ ] Reload on `#/runs/{kind}/{id}` re-fetches and renders without the user re-selecting.

## Catalog screen

- [ ] Definitions render grouped by kind.
- [ ] Search and kind filter narrow results without changing groupings.
- [ ] Sort dropdown reorders correctly.
- [ ] Clicking a definition with an active version opens the right-side detail with frontmatter, version pills, and body.
- [ ] Definitions with no active version show the warning chip and a disabled "No active version" button.

## Errors screen

- [ ] List renders or shows "No errors. Smooth sailing." empty state.
- [ ] Refresh button re-fetches.

## Accessibility

- [ ] Tab through the page — first tab focuses the **Skip to main content** link, which is visible (peach pill) and jumps to `#main-content` when activated.
- [ ] All primary nav buttons receive a visible focus ring (Tab cycles through them).
- [ ] Compact run rows on Overview are reachable by Tab; pressing Enter or Space opens the run.
- [ ] Activity pause toggle's accessible name is announced (use VoiceOver / NVDA: "Pause or resume live activity, button, pressed/not pressed").
- [ ] In macOS Settings → Accessibility → Display → "Reduce motion" enabled: activity feed items appear without slide-in, toasts appear without slide-in.
- [ ] Screen reader announces section changes via the `status-message` aria-live region after data loads.

## Responsive

- [ ] Resize to ~1024px desktop → no horizontal scrollbar, sidebar visible (when `data-nav="sidebar"`).
- [ ] Resize to ~768px tablet → grid collapses to single column, top-tabs scroll horizontally if needed.
- [ ] Resize to ~375px mobile → cards stack, toolbar inputs full width, tweaks panel takes full viewport width.

## Cross-cutting

- [ ] No `/api/dashboard/events` errors in browser console under normal operation.
- [ ] Browser DevTools → Network: no requests to `/api/dashboard/logs`, `/api/dashboard/settings`, `/admin/*` (those are explicitly out of scope).
- [ ] `prefers-color-scheme: dark` does not override the explicit `data-theme` attribute set via Tweaks.

---

## When to extend this checklist

Add a row whenever:

- A new dashboard section appears in primary nav.
- A new endpoint is consumed by the UI.
- A new interaction pattern is added that would not be caught by Go contract tests.

When a step becomes load-bearing enough that humans repeatedly find regressions in it, promote it to a Go contract test if possible.
