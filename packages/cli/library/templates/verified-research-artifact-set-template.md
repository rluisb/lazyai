# Standard Artifact Set — Verified Research

Every application of this methodology produces a folder with a predictable artifact set. Keep the layout consistent across investigations so future researchers know where to look.

---

## File Inventory

| File | Purpose | When required |
|------|---------|---------------|
| `findings-*.md` | Per-track raw findings | Always |
| `report.md` | Unified synthesis with TL;DR | Always |
| `alignment.md` | Cross-reference against source of truth | Always |
| `analysis.md` | Re-baselined doc after verification | When source of truth exists |
| `contribution.md` | Proposed edits back to source of truth | When contribution is intended |

---

## Naming Conventions

- Use kebab-case for filenames.
- Prefix findings files with `findings-` followed by the track name.
- Use the names above exactly for the shared artifacts.

---

## Folder Layout

```
[topic]/
  findings-codebase.md
  findings-docs.md
  findings-interviews.md
  report.md
  alignment.md
  analysis.md
  contribution.md
```

---

## File Structure Conventions

### `findings-*.md` (per-track findings)

Required sections:
1. Header (research date, source, method)
2. Raw observations
3. Claim inventory

After verification pass:
- Inline markers: `[VERIFIED]`, `[DRIFT: <correction>]`, `[UNVERIFIED]`
- Append-only — drift corrections are visible inline, not silently rewritten.

### `report.md` (unified synthesis)

Required sections:
1. TL;DR (3-5 bullets, direct answers)
2. Findings summary grouped by confidence
3. Risks and open questions
4. Recommendations

### `alignment.md` (cross-reference)

Required sections:
1. Source of truth (link to the authoritative doc)
2. Point-by-point alignment table
3. Deltas and gaps

### `analysis.md` (re-baselined doc)

Required front matter:
- `Baseline: <link to authoritative doc>`
- `Date: YYYY-MM-DD`

Body structured around the authoritative doc's framing, not the original synthesis.

### `contribution.md` (proposed edits)

Required sections:
1. Target page (link)
2. Proposed changes (diff-style)
3. Rationale per change

---

## Invariants

- **Findings files are append-only after verification.** Drift corrections are visible inline, not silently rewritten.
- **Report always has a TL;DR.** If the TL;DR cannot be written, the research is not done.
- **Alignment requires a named source of truth.** If none exists, state that explicitly and skip `analysis.md`.
- **Contribution is optional.** Not every investigation feeds back upstream.
