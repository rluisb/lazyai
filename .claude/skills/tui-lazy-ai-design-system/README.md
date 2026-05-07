# LazyAI Design System

> Visual + voice foundation for **LazyAI** — a tool-agnostic CLI that scaffolds AI development environments and compiles them into native formats for OpenCode, Claude Code, and GitHub Copilot.

LazyAI has no GUI. It is a **terminal-first product**: a Go binary (`lazyai-cli`) plus optional MCP runtime (`lazyai-orchestrator`) and diff viewer (`lazyai-diffviewer`). Its "interface" is a Charm Bracelet TUI — Bubble Tea + Lip Gloss + Huh — rendered against the Catppuccin Mocha palette. Anything we design *around* the product (docs site, README headers, slide decks, marketing pages) should feel like a friendly extension of that terminal world: dark canvas, monospaced rhythm, ASCII chrome, the same six accent colors used as semantic states.

This system codifies that look so anyone — including an AI agent — can produce on-brand artifacts without re-deriving it from scratch.

---

## Sources

Everything here was extracted from a single source repo (no Figma, no separate brand book). If you want to verify or extend, read these files directly:

- **Repo:** [`github.com/rluisb/lazyai`](https://github.com/rluisb/lazyai) (default branch `main`)
- **TUI palette + style functions:** `packages/cli/tui/theme/theme.go` — *the* canonical color and component vocabulary
- **Diff viewer palette (subset):** `packages/diffviewer/theme.go`
- **Reusable TUI components:** `packages/cli/tui/components/{spinner.go, summary.go, table.go, tree.go}`
- **Wizard flow + form labels (voice reference):** `packages/cli/tui/wizard/interactive.go`
- **Top-level voice + structure:** `README.md`, `docs/index.md`, `docs/getting-started/quick-start.md`, `docs/concepts/how-it-works.md`, `docs/cli/reference.md`
- **Recorded TUI demos (real screen captures):** `assets/demos/01-hero.gif` through `assets/demos/06-maintain.gif` — six canonical recordings from upstream `lazyai/demo/`

The TUI uses the [Charm Bracelet](https://charm.sh) stack. Charm's Lip Gloss is responsible for everything you see rendered in the terminal.

---

## Index

| File / Folder | What it is |
|---|---|
| `README.md` | This document. Brand context, content fundamentals, visual foundations, iconography. |
| `colors_and_type.css` | All design tokens as CSS custom properties — colors, type scale, spacing, radii, shadows, semantic element styles. **Import this first** in any new HTML artifact. |
| `SKILL.md` | Skill manifest so this folder is portable as an Agent Skill. |
| `fonts/` | Webfont files (JetBrains Mono + Inter, loaded from Google Fonts CDN — see "Fonts" caveat below). |
| `assets/` | Logos, glyphs, demo GIFs. |
| `assets/demos/` | Real `lazyai-cli` screen captures from the repo. Use these in marketing/docs over hand-built mockups. |
| `preview/` | Small HTML cards that populate the **Design System** tab. One sub-concept per card. |
| `ui_kits/cli/` | Recreation of the LazyAI TUI in HTML/JSX — wizard form, summary boxes, status tree, diff viewer, spinner. Use as a component library when mocking new CLI screens. |

---

## Brand at a glance

- **Name:** LazyAI · CLI binary `lazyai-cli` · also `lazyai-orchestrator`, `lazyai-diffviewer`
- **One-liner:** *Scaffold a canonical, multi-tool AI development environment from one CLI.*
- **Voice:** matter-of-fact, lowercase shell prompts, imperative verbs (`init`, `compile`, `doctor`), no exclamation marks, no emoji in product copy.
- **Distribution:** Go modules — `go install github.com/rluisb/lazyai/packages/...`
- **Tools targeted:** OpenCode · Claude Code · GitHub Copilot
- **Platforms:** macOS, Linux. License MIT.

---

## Content fundamentals

### Tone

Engineer-to-engineer. Reads like good `--help` output: short lines, no marketing fluff, an assumption that the reader knows what a CLI is. The README opens with `Quick Start` and a code block — there is no hero paragraph selling the idea.

> *Examples (verbatim from `docs/`):*
>
> - "Scaffold a canonical, multi-tool AI development environment from one CLI, with optional orchestration scaffolding and MCP runtime integration."
> - "Edit one tool-agnostic layer; `lazyai-cli` generates the rest."
> - "Local native agents are the intended execution path."
> - Wizard prompts: `Scope`, `AI Tools`, `MCP Preset`, `Project Name`, `Branch Pattern`.

### Person

- Use **imperative second person** for instructions: *"Run `lazyai-cli init`."* / *"Edit canonical files."*
- Use **third person** for what the tool does: *"`lazyai-cli` writes `.ai-setup.json`."*
- Avoid first person plural ("we", "our") in product copy. The README uses it sparingly in passing only.

### Casing

- **Sentence case** for all UI labels and headings. Not Title Case. (`"Branch Pattern"` is the rare exception — a proper-noun-feeling form label.)
- **lowercase** for command names, flags, file paths: `lazyai-cli init`, `--no-interactive`, `.ai-setup.json`.
- **kebab-case** for binaries and packages: `lazyai-cli`, `claude-code`, `lazyai-orchestrator`.
- **MCP, CLI, TUI, MIT, JSON, TOML** stay uppercase as proper acronyms.

### Punctuation & rhythm

- Em-dash (`—`) is the connective tissue. Used heavily in form descriptions: *"Standard (recommended) — +RPI, reasoning, bug resolution"*.
- Backticks for everything code-shaped. Inline `code` is the default emphasis, not bold.
- Periods at the end of full sentences; **no period** on bullet labels or table cells.
- No exclamation marks. None. Even in success states the tool just shows `✓ Build success in 38ms`.

### Emoji

**Effectively never.** The repo has zero emoji in product copy. Status communication uses Unicode glyphs (`✓ ✗ ⚠ ⚡ ○ •`) treated as iconography, not decoration. Don't introduce emoji in slides or docs unless explicitly asked.

### Vibe

A senior engineer writing internal tooling for other senior engineers — terse, accurate, slightly playful in naming (`doctor`, `eject`, `cupcake.yml`) but never cute in the body copy itself. The personality is in the *commands*, not the prose.

---

## Visual foundations

### Color

The palette is **Catppuccin Mocha-derived**, lifted directly from `packages/cli/tui/theme/theme.go`. Every CSS variable in `colors_and_type.css` corresponds 1:1 to a `lipgloss.Color` constant in Go.

| Token | Hex | Role |
|---|---|---|
| `--lz-primary` | `#7D56F4` | Brand purple. Titles, section headers, brand marks. |
| `--lz-secondary` | `#4ECDC4` | Teal accent. Subtitles, bullets, links. |
| `--lz-success` | `#2ECC71` | Installed / OK / done · always paired with `✓` |
| `--lz-warning` | `#F1C40F` | Modified / drift · always paired with `⚠` |
| `--lz-error` | `#E74C3C` | Missing / failed · always paired with `✗` |
| `--lz-orange` | `#E8912D` | Conflict · always paired with `⚡` |
| `--lz-highlight` | `#89B4FA` | Keyboard shortcuts, key badges, table headers. |
| `--lz-text` | `#CDD6F4` | Default foreground. |
| `--lz-dimmed` | `#6C7086` | De-emphasized text, separators, pending state. |
| `--lz-bg` | `#1E1E2E` | Default canvas. |
| `--lz-bg-code` | `#313244` | Inline code background, table row stripe. |

**Rules**

- Dark canvas is the default; light variants exist only as a courtesy for printed docs (not preferred).
- Never invent a new accent. If a state isn't listed, it doesn't exist yet — propose adding it to `theme.go` first.
- Color is **always** redundant with a glyph. `✓ green`, `⚠ yellow`, `✗ red`, `⚡ orange`. Never communicate state with color alone.

### Type

- **Display + body:** `Inter` (variable, substituted — loaded from Google Fonts CDN). For long-form prose, marketing pages, docs site headings. Replace with an official sans if/when the team ships one.
- **Mono (everything terminal):** `JetBrains Mono Nerd Font` — official, ships locally from `fonts/` (7 weights: Light, Regular, Italic, Medium, SemiBold, Bold, ExtraBold). Code blocks, command names, file paths, all TUI mockups, all status lines. The Nerd Font variant carries the Unicode glyphs (✓ ⚠ ✗ ⚡ ⠋…) the Bubble Tea TUI relies on. Licensed under SIL OFL 1.1 — see `fonts/OFL.txt`.
- **Caveat:** mono is local; Inter is still a Google-Fonts substitute. If the team prefers a different mono (Berkeley Mono, IBM Plex Mono, etc.), swap them and regenerate the type cards. A proportional (`Propo`) variant is *not* currently wired up — `--font-sans` falls back to standard Mono today; add Propo `@font-face` declarations and ship the `Propo-*` `.ttf` files alongside Mono if proportional metrics are wanted.

Type scale lives in `colors_and_type.css` — modular, tightly tracked, with semantic element styles (`h1`–`h4`, `p`, `code`, `kbd`, `.tui-line`).

### Spacing & rhythm

The TUI thinks in **monospace columns**, not pixels. Translate to web with an 8px base grid:

```
4 · 8 · 12 · 16 · 24 · 32 · 48 · 64 · 96
```

Box padding inside a Lip Gloss `Box()` is `Padding(0, 1)` — translate as `padding: 0 8px` for cards that mimic the TUI directly.

### Borders, radii, shadows

- **Box borders:** Lip Gloss `RoundedBorder()` — translate as `border: 1px solid var(--lz-dimmed)` and `border-radius: 6px`. Variants in `--lz-success`, `--lz-warning`, `--lz-error` (matches `SuccessBox()`, `WarningBox()`, `ErrorBox()`).
- **Corner radii:** `--radius-sm` 4px (chips), `--radius-md` 6px (cards/boxes), `--radius-lg` 12px (modal-feeling surfaces). Never larger than 16px.
- **Shadows:** the TUI has none. For web, use a single subtle elevation (`--shadow-1`) only — heavy drop shadows feel off-brand.

### Backgrounds

- **Default:** flat `--lz-bg` (`#1E1E2E`). No gradients, no noise, no full-bleed photography.
- **Acceptable accents:** a single subtle radial glow in `--lz-primary` at 8% opacity *behind* a hero element. Use sparingly.
- **Never:** bluish-purple gradients, glass/blur, gradient text on body copy, generic abstract Unsplash photos.

### Animation

- **Primary motion:** spinner dots (`⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`), cursor blink, progress bars filling left→right.
- **Easing:** linear for spinners, `ease-out` for state transitions. No bouncy springs, no scale-pop.
- **Durations:** 120ms (state flips), 200ms (panel reveals), 80ms (spinner frame). Anything longer feels sluggish in a CLI context.
- **No parallax. No scroll-jacking.** Decorative motion is off-brand.

### Hover & press

- **Hover:** brighten foreground by ~10% (`color-mix(in oklch, var(--lz-primary) 90%, white)`), or underline. Never glow, never scale.
- **Press / active:** invert — set background to the accent color, foreground to `--lz-bg`. Same pattern as a selected option in `huh.NewSelect`.
- **Focus ring:** 2px solid `--lz-highlight` outline, 2px offset. Always visible — this is a keyboard-first product.

### Transparency, blur, layering

- Transparency is reserved for **dimmed text** (effectively `--lz-dimmed`) and the optional 8% primary glow mentioned above. **No frosted glass, no backdrop-filter.**
- Layering is flat. Cards sit on the canvas with a 1px border, not stacked elevations.

### Cards

A "card" in this system = a Lip Gloss bordered box. Translate to:

```css
.card {
  background: var(--lz-bg);
  border: 1px solid var(--lz-dimmed);
  border-radius: var(--radius-md);
  padding: 12px 16px;
  box-shadow: none;
}
```

Status-bordered variants swap `border-color` to `--lz-success`, `--lz-warning`, `--lz-error`, or `--lz-orange`.

### Layout rules

- Max line length for prose: **72ch**. The TUI literally wraps at terminal width — respect that mental model.
- Gutters scale on the 8px grid. Two-column layouts gap at 32px minimum.
- Fixed elements: a header (binary name + version chip) and a footer (key hints like `↑↓ navigate · enter select · q quit`) — copying the wizard's own chrome.

### Imagery

- Real terminal screenshots / GIFs (in `assets/demos/`) trump illustrations.
- If you must add an illustration, prefer **ASCII-art-feeling** monochrome line drawings on dark, not full-color isometric scenes.
- Avoid stock photography entirely.

---

## Iconography

LazyAI does not use an icon font, an SVG sprite library, or emoji. **All iconography is Unicode glyphs rendered in the terminal palette.** This is unusual and worth honoring.

### Status glyphs (canonical — never substitute)

| Glyph | Token | Meaning |
|---|---|---|
| `✓` | success | installed, ok, done |
| `✗` | error | missing, failed |
| `⚠` | warning | modified, drift |
| `⚡` | orange | conflict |
| `○` | dimmed | pending |
| `•` | secondary | bullet |
| `─` | dimmed | horizontal rule, table separator |
| `├── └── │` | dimmed | tree connectors |
| `█ ░` | success / dimmed | progress bar fill / track |
| `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏` | primary | spinner frames (Bubbles `Dot` style) |

### Brand mark

The repo itself ships no logo image. We render a **wordmark** built from JetBrains Mono: `lazyai` lowercase, with the leading `>` shell prompt in `--lz-secondary` and `lazyai` in `--lz-text`, optional `_` cursor blinking in `--lz-primary`. See `assets/lazyai-wordmark.svg` and the `Logo` card under preview/.

### When you actually need an icon

Sometimes web content (a docs nav, a blog post intro) wants a real icon. In that case:

1. Use [**Lucide**](https://lucide.dev) at stroke-width 1.75, rendered in `currentColor`. It pairs cleanly with JetBrains Mono.
2. Load via the official CDN (`https://unpkg.com/lucide@latest`) — do not vendor copies.
3. **Flag it as a substitution** in the artifact, since this is not present in the upstream repo.

### Imagery in `assets/`

- `assets/lazyai-wordmark.svg` — primary wordmark
- `assets/lazyai-mark.svg` — square mark for favicons / chips
- `assets/demos/01-hero.gif` … `06-maintain.gif` — real terminal recordings, lifted from `lazyai/demo/`. Treat these as the canonical product imagery.

---

## Caveats

1. **JetBrains Mono Nerd Font is official, Inter is substituted.** The mono face ships locally in `fonts/` (Mono variant for code, Propo for prose) — this is the actual face the upstream Bubble Tea TUI renders against. Inter is loaded from Google Fonts as a fallback prose face; replace if/when the team ships an official sans.
2. **No traditional UI kit.** LazyAI has no app, no website beyond the auto-generated MkDocs docs, no Figma. The "UI kit" here recreates the **TUI** itself — wizard form, summary box, file tree, diff viewer.
3. **Brand mark is invented.** The repo ships no logo. The wordmark in `assets/` is a reasonable derivation from the CLI's own visual vocabulary; replace with an official mark if/when one exists.
4. **No marketing site exists yet** (the docs site is auto-generated MkDocs). If you want a marketing UI kit, that's a separate pass.
