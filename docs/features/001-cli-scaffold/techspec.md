# TechSpec: AI Setup CLI

**Feature:** 001-cli-scaffold
**Date:** 2026-03-28
**Status:** Draft
**PRD:** [prd.md](./prd.md)
**Research:** [research.md](./research.md)

---

## Simplicity Gate

| Question | Answer |
|----------|--------|
| Can this be done without a new service? | YES — it's a CLI, no server |
| Can this be done without a database? | YES — files only |
| Can this be done without new dependencies? | Minimal — need a prompt library + file ops |
| Is there an existing pattern? | YES — npx CLI pattern (like create-next-app, skills CLI) |
| What is the simplest P1? | `init` command that copies files with interactive tool selection |

## Summary

A Node.js CLI package published to npm. Uses the adapter pattern to package tool-agnostic library content for each supported AI tool. Interactive prompts via `@clack/prompts`. Library content shipped inside the package (no git fetch at runtime).

## Approach Options

### Option A: TypeScript + @clack/prompts
- **How:** TypeScript CLI with `@clack/prompts` for beautiful interactive menus, `tsup` for build
- **Pros:** Type safety, great DX, @clack has the best terminal UX (used by Astro, SvelteKit)
- **Cons:** Build step required
- **Complexity:** Low

### Option B: Plain JavaScript + inquirer
- **How:** ESM JavaScript, `inquirer` for prompts
- **Pros:** No build step, widely used
- **Cons:** No types, inquirer UX is dated
- **Complexity:** Low

### Decision: Option A
**Why:** TypeScript gives us type safety for the adapter pattern. @clack/prompts is modern, beautiful, and lightweight. Build step is trivial with tsup.

## Architecture

```
User runs: npx @ricardoborges-teachable/ai-setup init
    │
    ▼
┌─────────────┐
│ CLI Entry   │ ← bin/ai-setup.js
│ (commander) │
└──────┬──────┘
       │
       ▼
┌──────────────┐     ┌──────────────┐
│ Prompts      │────▶│ Scaffold     │
│ (@clack)     │     │ Engine       │
│ - setup type │     │ - creates    │
│ - tools      │     │   docs/      │
│ - project    │     │ - copies     │
│   name       │     │   library    │
└──────────────┘     └──────┬───────┘
                            │
                    ┌───────┼───────┐
                    ▼       ▼       ▼
              ┌─────────┐ ┌─────────┐ ┌──────────┐
              │ Pi      │ │OpenCode │ │ Future   │
              │ Adapter │ │ Adapter │ │ Adapters │
              └─────────┘ └─────────┘ └──────────┘
```

### Components

| Component | Responsibility | Location |
|-----------|---------------|----------|
| CLI entry | Parse commands, route to handlers | `src/cli.ts` |
| Prompts | Interactive user input | `src/prompts.ts` |
| Scaffold engine | Create docs/ structure, copy shared files | `src/scaffold.ts` |
| Pi adapter | Format + copy to `.pi/` directories | `src/adapters/pi.ts` |
| OpenCode adapter | Format + copy to `.opencode/` directories | `src/adapters/opencode.ts` |
| Library content | All markdown files | `library/` (shipped in package) |
| File utils | Copy, write, check existence, hash comparison | `src/utils/files.ts` |

## Patterns to Follow

| Pattern | Reference |
|---------|-----------|
| CLI structure | `create-next-app`, `@clack/create-app` |
| Adapter pattern | One file per tool, common interface |
| Library-in-package | Content shipped with npm, no runtime fetch |

## Data Model

No database. All state is the filesystem.

**Config file (written after init):** `.ai-setup.json`
```json
{
  "version": "1.0.0",
  "setupType": "project",
  "tools": ["pi", "opencode"],
  "installedAt": "2026-03-28T00:00:00Z",
  "files": {
    "docs/rules/workflow.md": { "hash": "abc123", "customized": false },
    "docs/rules/code-style.md": { "hash": "def456", "customized": true }
  }
}
```

This enables:
- `update` knows which files are customized (skip) vs untouched (overwrite)
- `doctor` knows what should exist
- `status` shows what's installed
- `add` knows which tools are already configured

## API (CLI Commands)

| Command | Args | Flags | Action |
|---------|------|-------|--------|
| `init` | — | `--type project\|workspace` `--tools pi,opencode` `--name my-project` `--no-interactive` | Full scaffold |
| `add` | `<tool>` | — | Add tool-specific files to existing setup |
| `remove` | `<tool>` | — | Remove tool-specific files |
| `update` | — | `--force` (overwrite all) | Update library files, skip customized |
| `doctor` | — | — | Verify setup integrity |
| `status` | — | — | Show installed tools + file status |

**Non-interactive mode** (for scripts/CI):
```bash
npx @ricardoborges-teachable/ai-setup init --type project --tools pi,opencode --name my-api --no-interactive
```

## Dependencies

| Dependency | Why | Alternative Rejected |
|------------|-----|---------------------|
| `@clack/prompts` | Beautiful interactive prompts | inquirer (dated UX), gum (bash only) |
| `commander` | CLI argument parsing | yargs (heavier), minimist (too minimal) |
| `tsup` | TypeScript build | esbuild (lower level), tsc (slower) |
| `picocolors` | Terminal colors (already in @clack) | chalk (heavier) |

**Total production deps: 2** (commander + @clack/prompts). Minimal.

## Test Strategy

| Level | What | How |
|-------|------|-----|
| Unit | Adapter formatting, file hashing, config read/write | Vitest + temp directories |
| Integration | Full init flow → verify file tree | Vitest + temp git repos |
| Smoke | `npx` from published package works | Manual after publish |

## Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| OpenCode changes their directory structure | Low | Medium | Adapter is isolated — update one file |
| npm package name taken | Low | Low | Check availability before publishing |
| Large package size (29 library files) | Low | Low | Markdown is tiny. Total < 200KB. |

## ADRs

- [ ] ADR needed: TypeScript + @clack over alternatives → `docs/adrs/001-typescript-clack-cli.md`

---

<!-- PRINCIPLES CHECK
- [x] Simplicity gate passed — minimal deps, file-copy based
- [x] 2 approaches explored (TS+clack vs JS+inquirer)
- [x] Decision justified (type safety for adapters, better UX)
- [x] Architecture uses adapter pattern (extensible without modifying core)
- [x] Data model is minimal (one JSON config file)
- [x] Test strategy covers init flow end-to-end
- [x] YAGNI: only Pi + OpenCode adapters, others later
-->
