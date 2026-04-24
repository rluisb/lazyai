# Cross-Runtime Parity

The ai-setup repo ships two peer implementations: **Go** (`packages/ai-setup-go/`)
and **TypeScript** (`packages/ai-setup-ts/`). The project's core rule is:

> Any feature added to one runtime must be added to the other in the same change set.

## The harness

`scripts/parity-check.sh` is the enforcement mechanism. It:

1. Builds both runtimes from source
2. Runs `ai-setup init` with identical inputs through each binary
3. Diffs the resulting output trees
4. Exits non-zero on any unexpected divergence

### Local invocation

```bash
bash scripts/parity-check.sh          # quiet mode
bash scripts/parity-check.sh --verbose # shows steps
```

### CI

`.github/workflows/parity.yml` runs the harness on every PR targeting `main`.
The job is **currently advisory** (`continue-on-error: true`) because known
divergences exist — see below. Graduate it to blocking once those are closed.

## Expected (filtered) differences

The harness intentionally ignores:
- `.ai-setup.db` / `.ai-setup.db-journal` — Go uses SQLite, TS uses lowdb JSON. Both store the same manifest content in different physical formats; a structural comparison would require its own tool.
- Manifest timestamps — `installedAt`, `lastUpdatedAt`, `cliVersion` fields in `.ai-setup.json`.
- File hashes — both runtimes compute per-file content hashes, but since content can differ (e.g. trailing newlines), the 16-char hash strings often differ too.
- `.git/`, `node_modules/` — independently initialized in each temp dir.

## Known unexpected divergences (2026-04-24)

First run of the harness surfaces these real gaps. Each deserves a follow-up spec or bugfix:

| Path | Go | TS | Investigation |
|---|---|---|---|
| `opencode.jsonc` (project root) | absent | present | Adapter writes to different locations between runtimes after the spec 019 consolidation — verify both target `<target>/opencode.jsonc` for project scope. |
| `specs/{adrs,bugfixes,features,memory,prompts,rules,standards,templates}/AGENTS.md` | absent | present | TS scaffolds per-spec-dir AGENTS.md guides; Go does not. Decide whether Go should (add scaffold step) or TS shouldn't (remove — but users rely on them). |
| `specs/refactors/`, `specs/tech-debt/` (directories) | absent | present | TS creates these spec category dirs; Go does not. Minor scaffold parity gap. |
| `specs/rules/{agent-security,cost,review,tool-use}.md` | absent | present | TS copies these rule templates from the library; Go copies a different subset. Audit the rules copy step in both adapters. |
| `specs/templates/{code-review-template,postmortem-template,tech-debt-template}.md` | absent | present | Same pattern — TS copies more templates than Go. Align the template manifest. |

Each row is a small fix (~10–40 LOC in most cases). They're deferred from the harness-introduction commit because fixing them requires touching both runtimes and warrants individual review rather than a big bundled change.

## When to update this file

Update the "known divergences" table when:
- A new divergence appears (harness starts failing on a new file)
- An existing divergence is closed (remove the row; if all closed, flip the CI job to `continue-on-error: false`)

## Adding more scenarios

The harness currently only exercises `init --scope project --tools opencode`. To expand coverage:

1. Add a new `run_scenario()` function in `scripts/parity-check.sh` accepting the init args
2. Call it from `main` with each scenario
3. Fail the harness if any scenario diverges

Future scenarios to consider:
- `--scope workspace` with multiple repos
- `--scope global`
- `add` command after an initial init
- `compile` after modifying scaffold files
