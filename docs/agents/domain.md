# Domain Docs

How the engineering skills should consume this repo's domain documentation when exploring the codebase.

This repo is **multi-context**: `GLOSSARY-MAP.md` at the root points at one `GLOSSARY.md` per package.

## Before exploring, read these

- **`GLOSSARY-MAP.md`** at the repo root: it points at one `GLOSSARY.md` per context. Read each one relevant to the topic.
- **`specs/adrs/`**: system-wide ADRs (this repo keeps them here, not in `docs/adr/`). Read ADRs that touch the area you're about to work in.
- **`packages/<context>/docs/adr/`**: context-scoped decisions, if present.

If any of these files don't exist, **proceed silently**. Don't flag their absence; don't suggest creating them upfront. The `/domain-modeling` skill (reached via `/grill-with-docs` and `/improve-codebase-architecture`) creates them lazily when terms or decisions actually get resolved.

## File structure

```
/
├── GLOSSARY-MAP.md
├── specs/adrs/                        ← system-wide decisions (001-…)
└── packages/
    ├── cli/
    │   ├── GLOSSARY.md
    │   └── docs/adr/                  ← cli-specific decisions
    └── diffviewer/
        ├── GLOSSARY.md
        └── docs/adr/                  ← diffviewer-specific decisions
```

New system-wide ADRs continue the existing `specs/adrs/NNN-slug.md` numbering.

## Use the glossary's vocabulary

When your output names a domain concept (in an issue title, a refactor proposal, a hypothesis, a test name), use the term as defined in the relevant `GLOSSARY.md`. Don't drift to synonyms the glossary explicitly avoids.

If the concept you need isn't in the glossary yet, that's a signal: either you're inventing language the project doesn't use (reconsider) or there's a real gap (note it for `/domain-modeling`).

## Flag ADR conflicts

If your output contradicts an existing ADR, surface it explicitly rather than silently overriding:

> _Contradicts ADR-006 (manifest-driven compile), but worth reopening because…_
