# LazyAI + vibe-lab — Implementation Roadmap

Verification date: 2026-06-21.

---

## Guiding sequence

1. Product boundaries and canonical model.
2. Official-doc adapter compliance.
3. Validation and safe compilation.
4. Security/trust/sandbox warnings.
5. Migration/eject.
6. Packaging/plugin generation.
7. Optional trace/eval improvement loop.

---

## Milestone 0 — Product boundary and docs alignment

Goal: lock the product identity before implementation expands.

Tasks:

- Write LazyAI/vibe-lab split doc.
- Write harness principles doc.
- Write official-doc compliance matrix.
- Create source registry.
- Mark unsupported/experimental surfaces clearly.
- Add `CLAUDE.md` compliance gap to tracker.
- Decide strict vs compatibility skill output defaults.

Exit criteria:

- Maintainers agree LazyAI is asset manager/compiler, not runtime.
- All target tools have adapter requirement docs.
- All official-doc sources have verification dates.

---

## Milestone 1 — Canonical source and core compile pipeline

Goal: compile a simple `.ai/` setup into native outputs for OpenCode, Claude, and Copilot.

Tasks:

- Implement manifest parser.
- Implement asset loader.
- Implement frontmatter parser.
- Implement canonical graph resolver.
- Implement diff planner.
- Implement safe writer.
- Implement lockfile.
- Implement `init`, `compile`, `validate`, `status`.
- Implement initial OpenCode adapter.
- Implement initial Claude adapter with `CLAUDE.md`.
- Implement initial Copilot adapter.

Exit criteria:

- Starter pack compiles for OpenCode/Claude/Copilot.
- Generated files pass golden tests.
- No unsafe overwrites.
- `validate` catches invalid skill names and inline secrets.

---

## Milestone 2 — Full adapter matrix

Goal: add Pi, Kiro, Antigravity, and OMP outputs.

Tasks:

- Implement Pi adapter.
- Implement Kiro adapter.
- Implement Antigravity adapter.
- Implement OMP adapter.
- Add adapter capabilities model.
- Add per-adapter experimental/beta/stable status.
- Add docs conformance fixtures.

Exit criteria:

- All seven target adapters generate starter outputs.
- Each target has golden tests.
- Unsupported features emit warnings.
- OMP/Antigravity beta caveats are documented.

---

## Milestone 3 — Validation hardening

Goal: make LazyAI safe and predictable in real repos.

Tasks:

- Skill validator.
- Agent validator.
- Hook validator.
- MCP validator.
- Secret scanner.
- Path/symlink safety validator.
- Generated-file drift detector.
- Official-doc conformance validator.
- `doctor` security report.

Exit criteria:

- CI can run `lazyai-cli validate --all`.
- Dangerous hooks fail validation.
- Inline secrets fail in team profile.
- `doctor` reports trust/sandbox caveats for Pi and Kiro.

---

## Milestone 4 — MCP and hooks depth

Goal: support high-value tool integrations safely.

Tasks:

- Implement `.ai/mcp.json` compiler for all target outputs.
- Implement MCP inventory report.
- Implement hook policy IR.
- Implement Claude hook adapter.
- Implement OpenCode plugin hook adapter.
- Implement Pi extension hook adapter.
- Implement Kiro hook adapter.
- Implement Copilot CLI hook adapter.
- Implement Antigravity/OMP hook adapters as beta.

Exit criteria:

- Protected-path hook compiles to supported targets.
- Unsupported targets get clear warnings.
- MCP configs use env var references.
- Generated hook commands are testable.

---

## Milestone 5 — Migration and eject

Goal: allow teams to adopt LazyAI without losing existing setup.

Tasks:

- Implement native file scanner.
- Import `AGENTS.md`, `CLAUDE.md`, `.opencode`, `.claude`, `.github`, `.pi`, `.omp`, `.kiro`, `.agents`, `.gemini`.
- Confidence scoring.
- Migration report.
- Adopt generated regions.
- Implement `eject`.

Exit criteria:

- Existing multi-tool repo can migrate to `.ai/`.
- No existing native config is deleted silently.
- Ejected repo remains usable by host tools.

---

## Milestone 6 — Packages/plugins

Goal: ship team-distributable setup bundles.

Tasks:

- `build-plugin` for Claude Code.
- `build-plugin` for Copilot CLI.
- OMP plugin bundle.
- Pi package bundle.
- Antigravity plugin bundle when schema confirmed.
- Asset pack provenance and checksum metadata.

Exit criteria:

- Team can build and install a LazyAI/vibe-lab plugin in supported tools.
- Plugin contents pass native validation.
- Precedence/conflict warnings exist.

---

## Milestone 7 — Trace/eval improvement loop

Goal: support trace-backed harness improvement without becoming a runtime.

Tasks:

- Trace taxonomy templates.
- Eval case schema.
- Holdout schema.
- Rubric templates.
- Harness change report template.
- Manual trace import.
- `validate evals`.

Exit criteria:

- A recurring failure can be converted into an eval case.
- A harness change can be documented and reviewed.
- Holdouts prevent overfitting.
- No cloud tracing dependency required.

---

## Milestone 8 — v1.0 hardening

Goal: stable product release.

Tasks:

- Freeze `.ai/` v1 schema.
- Complete documentation.
- Security review.
- Adapter smoke tests.
- Cross-platform file writer tests.
- Installer/release process.
- Upgrade/migration guarantee.

Exit criteria:

- All stable adapters pass conformance tests.
- All core commands documented.
- No known destructive write issues.
- Security warnings are explicit.
- Product boundary is clear.
