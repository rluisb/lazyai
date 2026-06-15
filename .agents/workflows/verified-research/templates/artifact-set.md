# Standard Artifact Set — Verified Research

Every application of this methodology produces a folder with a predictable artifact set. Keep the layout consistent across investigations so future researchers know where to look.

---

## File inventory

| File | Purpose | When required |
|---|---|---|
| `findings-internal.md` | Internal-codebase evidence with line citations | Always |
| `findings-source.md` | Source/incumbent system docs evidence with URLs | Always |
| `findings-target.md` | Target/proposed system docs evidence with URLs | Always |
| `report.md` (or `RESEARCH.md`) | Unified synthesis (Phase 3 output) | Always |
| `alignment.md` (or `<author>-alignment.md`) | Cross-reference vs authoritative team doc | If Phase 5 finds external authoritative source |
| `analysis.md` (or `MIGRATION-ANALYSIS.md`) | Re-baselined primary doc | If Phase 6 re-baseline triggered |
| `contribution.md` (or `<target-doc>-contribution.md`) | Drop-in additions for existing docs | If Phase 8 contributing back |
| `research-playbook.md` | Methodology record for this investigation (worked example) | Optional; once per repo |

---

## Naming conventions

- Use kebab-case for filenames.
- When cross-referencing a specific authoring teammate, prefix with their name in lowercase (e.g., `kim-alignment.md`).
- When contributing back to a specific named page, suffix with the page slug (e.g., `addons-current-flow-contribution.md`).
- Keep `findings-*` filenames consistent so a glob picks them up cleanly.

---

## Folder layout example

```
specs/<topic-slug>/
├── findings-internal.md          ← Phase 1 (Track A)
├── findings-source.md            ← Phase 1 (Track B)
├── findings-target.md            ← Phase 1 (Track C)
├── report.md                     ← Phase 3 synthesis
├── alignment.md                  ← Phase 5 cross-reference (if applicable)
├── analysis.md                   ← Phase 6 re-baseline (if applicable)
└── contribution.md               ← Phase 8 contribution (if applicable)
```

---

## File structure conventions

### `findings-*.md` (per-track findings)

Required sections:
1. Header (research date, source, method)
2. Numbered sections (one per area investigated)
3. Citations inline (`path/to/file.ext:LINE` for code, exact URL for docs)
4. Citations appendix at the bottom

After Phase 4 verification:
- Inline markers: `[VERIFIED]`, `[DRIFT: <correction>]`, `[UNVERIFIED]`
- Drift-scope verification report appendix

### `report.md` (unified synthesis)

Required sections:
1. TL;DR (3–5 bullets, direct answers)
2. One section per viewpoint (internal, source, target)
3. Mapping table (concept ↔ source ↔ target ↔ gap)
4. Recommendation with reasoning
5. Implementation sketch (high-level only)
6. Open questions
7. Document map (links to sibling files)

### `alignment.md` (Phase 5 cross-reference)

Required sections:
1. Source of truth (link to the authoritative doc)
2. Where we aligned (table: claim · our research · their doc · match)
3. Where we missed (numbered gaps with migration impact)
4. Updated mental model
5. Updated scope (if applicable)
6. Updated open questions

### `analysis.md` (Phase 6 re-baselined doc)

Required front matter:
- "Baseline: <link to authoritative doc>"
- "Supersedes: <link to original synthesis>"

Body structured around the authoritative doc's framing, not the original synthesis.

### `contribution.md` (Phase 8 contribution back)

Required sections:
1. Target page (link)
2. What this document is (brief tone-setting intro)
3. Organized contribution sections, each with:
   - Drop-in content (in code-fence so it's directly copy-pasteable)
   - Why this helps (rationale)
4. Meta-rationale (why we're contributing)
5. Suggested review path
6. Open questions for reviewers

---

## Invariants

These hold across all investigations using this methodology:

- **Findings files are append-only after Phase 4.** Drift corrections are visible inline, not silently rewritten.
- **The original synthesis is never overwritten.** Re-baselining produces a new file.
- **Citations are non-optional.** Any claim without a citation is a quality bug.
- **Authoritative team docs always win.** If a team-authored doc contradicts independent research, the team doc is authoritative until proven otherwise.
- **Contribution files never modify the target doc directly.** They contain proposals for the target's author to apply.
