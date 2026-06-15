# LazyAI CLI · UI kit

A web recreation of the LazyAI TUI in plain HTML — useful for mocking new wizard screens, marketing diagrams, or docs visuals without spinning up a real terminal.

## Files

- `wizard.html` — the full `lazyai-cli init` wizard, six steps in a clickable stepper. Mirrors the upstream `huh` form: select, multi-select, text input, run progress, summary box.
- `surfaces.html` — four side-by-side surfaces: `doctor`, `status` (file tree), `lazyai-diffviewer`, and `compile --plan` (table view).
- `colors_and_type.css` — design tokens (copy of the system root file).

## Conventions used

- All text is JetBrains Mono.
- Color is **always** paired with a glyph (`✓ ⚠ ✗ ⚡ ○`) — never status-by-color-alone.
- Bordered "boxes" map to Lip Gloss `RoundedBorder()` with `Padding(0, 1)` — translated as `border: 1px solid` + `padding: 0 8px`.
- Key hints sit at the bottom of every panel, separated by a dashed rule. Same as the wizard chrome upstream.
- Form prompts: H1-style title in primary purple, dimmed secondary description below, options with a `▸` cursor (single-select) or `[•]/[ ]` markers (multi-select).

## What this is not

- Not a runnable CLI — these are static mockups with light JS for click-through behaviour.
- Not a Figma library — there's no design tool counterpart yet. Edit the HTML directly.
- Not exhaustive — only the surfaces currently rendered by `packages/cli/tui/`. New screens should be added here when added upstream.
