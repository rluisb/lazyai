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

## Closed divergences (fixed in the same commit that elevates PARITY.md)

The harness originally surfaced five divergences. All are now closed:

| Path | Fix |
|---|---|
| `opencode.jsonc` (project root) | Go adapter now writes the default config at project root for project/workspace scope; global still targets `~/.config/opencode/`. Matches TS. |
| `specs/{adrs,bugfixes,features,memory,prompts,rules,standards,templates}/AGENTS.md` | Go's `ScaffoldSpecs` now copies `library/specs-agents/<category>.md` → `specs/<category>/AGENTS.md` for each specs dir the preset scaffolds. Matches TS. |
| `specs/refactors/`, `specs/tech-debt/` directories | TS non-interactive default preset is now `standard` (matching Go), which excludes these dirs. The `full` preset still creates them via `--preset full`. |
| `specs/rules/*.md` (extra rule files) | Closed by the preset alignment above — `standard` preset's rule manifest is identical across runtimes. |
| `specs/templates/*.md` (extra template files) | Same — closed by preset alignment. |

## Known remaining divergences (surfaced by the harness on subsequent runs)

Now that the most visible divergences are closed, the harness exposes deeper systemic gaps. These warrant their own follow-up specs:

| Path | Go | TS | Investigation |
|---|---|---|---|
| `.ai/housekeeping/` | always created | only when `cliOverrides.housekeeping` is set | Go's `ScaffoldHousekeeping` runs unconditionally when `ctx.Housekeeping != nil`; TS wires a gated call at the wizard level. Either gate Go the same way or change TS to always run with a default config. |
| `.ai/orchestration/` | created | absent | TS wizard only calls `scaffoldOrchestration` when `enableServers` includes `orchestrator`; Go scaffolds unconditionally. Align. |
| `.opencode/opencode.jsonc` (MCP-scoped) | absent | created | TS MCP compiler runs during init; Go MCP compiler runs only on explicit `compile` command. Audit the scaffold pipeline and add the MCP compile step to Go's init flow. |
| `.ai/mcp.json` | content differs | content differs | Probably ordering/formatting only; verify and align via `marshalSortedJson` on the TS side. |
| `.opencode/agents/*.md` | content differs | content differs | Frontmatter transforms produce different output. Audit `stripFrontmatterAndInjectModel` vs Go's equivalent. |
| `.ai-setup.json` (manifest) | absent (uses `.ai-setup.db`) | present | Fundamental state-storage divergence: Go uses SQLite, TS uses lowdb JSON. Structural comparison would require its own tool — document as an intentional divergence. |
| `AGENTS.md` (project root compiled) | content differs | content differs | Template compiler output differs. Audit `scaffoldCompiledRoot` / Go's compiler output for placeholder interpolation differences. |

Each row is follow-up work. None are blocking — the harness is still advisory (`continue-on-error: true`).

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
