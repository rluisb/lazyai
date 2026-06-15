---
name: lazyai-design-system
description: |
  Visual + voice design system for LazyAI — a tool-agnostic CLI (lazyai-cli) that
  scaffolds AI development environments and compiles them into native formats
  for OpenCode, Claude Code, and GitHub Copilot. Use when designing anything
  that lives around the LazyAI product: docs pages, slide decks, marketing
  pages, terminal mockups, or recreating the TUI itself.

  Built from rluisb/lazyai @ packages/cli/tui/theme/theme.go. Catppuccin
  Mocha-derived dark palette, JetBrains Mono for everything terminal, Inter
  for prose. Status communication is always glyph + colour together
  (✓ ⚠ ✗ ⚡ ○) — never colour alone.
---
## Quick Reference

| | |
|---|---|
| **Use when** | [When to use this skill] |
| **Do not use when** | [When NOT to use this skill] |
| **Primary agent** | [Which agent uses this] |
| **Runtime risk** | [Low/Medium/High] |
| **Outputs** | [What this skill produces] |
| **Validation** | [How to validate output] |
| **Deep mode trigger** | [How to trigger full mode] |



# LazyAI design system

Read **`README.md`** first. It is the source of truth for content, voice, and
visual rules; this file just tells you how to use what's there.

## Files at a glance

| Path | Use it for |
|---|---|
| `README.md` | Brand context, content fundamentals, visual foundations, iconography. **Read first.** |
| `colors_and_type.css` | All design tokens as CSS custom properties. **Always import this first.** |
| `assets/lazyai-mark.svg` `assets/lazyai-wordmark.svg` | Logo files. The mark is invented from the CLI's own visual vocabulary; flag if a real one exists. |
| `assets/demos/*.gif` | Real `lazyai-cli` screen captures from the upstream repo. Prefer these over hand-built mockups. |
| `preview/` | Small HTML cards demonstrating each token / component in isolation. |
| `ui_kits/cli/index.html` | Full TUI recreation — wizard, summary, tree, compile progress, diff viewer. **Use as the component library when mocking new CLI screens.** |

## Workflow

1. **Always start by importing the tokens.** From a sibling folder:
   ```html
   <link rel="stylesheet" href="../colors_and_type.css">
   ```
   Anything you draw inherits the palette, type scale, semantic element styles
   (`h1`–`h4`, `code`, `kbd`, `pre`), and TUI primitives (`.tui-box`,
   `.tui-line`, `.glyph-ok`, `.chip`, `.btn`, `.spinner`).

2. **For terminal mockups, copy from `ui_kits/cli/index.html`.** Each section
   there is a complete, working component — wizard, multi-select, tree,
   summary, progress, diff. Paste a section and rewrite the copy; do not
   rebuild the chrome from scratch.

3. **For docs / slides / marketing**, the same tokens apply, but you can use
   Inter for prose. The visual rules in `README.md → Visual foundations` still
   hold: dark canvas, no gradients, no emoji, status communicated with glyphs.

## Hard rules (do not break without flagging)

- **No emoji in product copy.** Status uses Unicode glyphs (`✓ ⚠ ✗ ⚡ ○ •`).
- **No exclamation marks.** Even success states say `✓ Build success in 38ms`.
- **Sentence case** for UI labels and headings. Not Title Case.
- **lowercase** for binaries, flags, file paths: `lazyai-cli init`,
  `--no-interactive`, `.ai-setup.json`.
- **Color is always paired with a glyph.** Never communicate state with colour
  alone — accessibility *and* it's how the TUI does it.
- **No new accent colours.** The palette is fixed at 7 brand/semantic colours
  + 2 surfaces. If you need a new state, propose adding it to `theme.go`
  upstream first.
- **No gradients, glass, blur, drop shadows.** A single subtle radial primary
  glow at 8% behind a hero element is the only allowed accent.

## Voice cheat sheet

> Engineer-to-engineer. Reads like good `--help` output.

- ✅ "Scaffold a canonical, multi-tool AI development environment from one CLI."
- ✅ "Edit one tool-agnostic layer; `lazyai-cli` generates the rest."
- ❌ "Supercharge your AI workflow! 🚀"
- ❌ "We're excited to introduce…"

Em-dash (`—`) is the connective tissue. Backticks for everything code-shaped.

## Caveats — flag these in your output

1. **JetBrains Mono Nerd Font is official, Inter is substituted.** Mono ships
   locally in `fonts/` (the actual face the Bubble Tea TUI renders against).
   Inter is a Google Fonts fallback for prose — replace when an official
   sans is provided.
2. **Brand mark is invented.** The repo ships no logo. The wordmark in
   `assets/` is derived from the CLI's own visual vocabulary; replace if/when
   an official mark exists.
3. **No marketing UI exists upstream.** If asked to design a marketing site,
   note that the docs are auto-generated MkDocs and you're inventing the
   marketing surface from this terminal-first foundation.
