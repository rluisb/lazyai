# Plan: Kiro Agent Format — JSON vs .md Reconciliation

**Issue:** #574 (part of #568)
**Date:** 2026-06-29
**Status:** draft — awaiting human gate
**Depends on:** Research: `research.md` (this directory)
**Blocked:** runtime format verification (auth); tool-capability values (#569)

---

## Decision Summary

Current evidence **favors JSON** as the required format: `kiro.md` (verified 2026-06-29, official docs) explicitly states custom agents are JSON at `.kiro/agents/<name>.json`. The `spec.md:35` "tolerated `.md`" finding predates this research, has no recorded source URL, and was a peripheral assertion in a hooks-focused spec.

However, runtime verification remains blocked (`kiro-cli agent validate` requires auth). Two implementation paths are therefore defined; the correct path is selected after verification.

---

## Path A — Docs-Only Reconciliation (if `.md` is confirmed tolerated)

**Trigger:** Runtime test with Kiro CLI login confirms `.kiro/agents/<name>.md` is discovered as a custom agent.

**Evidence requirement:** A logged-in `kiro-cli agent validate .kiro/agents/probe.md` succeeds (exit 0, agent recognized) before Path A may be selected.

### Tasks

| # | Task | Scope |
|---|---|---|
| A-1 | Update `specs/030-kiro-cli-v3-output-gaps/spec.md` A-003 to add caveat: "`.md` tolerated confirmed; `.json` also accepted; adapter may continue emitting `.md`." Cite the test result as evidence. | docs |
| A-2 | Update `docs/ai-cli-tools/tool-systems/kiro.md` custom-agents section to note `.md` is discovered alongside `.json` with a gotcha entry. Add source URL or note as [INFERENCE] if only derived from CLI test. | docs |
| A-3 | Update `kiro.go:10-13` inline comment to say "`.md` confirmed tolerated; `.json` also valid." | code (comment only) |
| A-4 | Add `tools`/`allowedTools` emit to the existing `.md` agent output **if Kiro CLI parses those fields from YAML frontmatter** (separate verify). If not parsed from `.md`, mark as a gap and defer to post-#569. | code or gap-note |

**Net change:** Minimal; no adapter format change required.

---

## Path B — JSON Transform (if `.json` is required)

**Trigger:** Runtime test with Kiro CLI login confirms `.kiro/agents/<name>.md` is NOT discovered (agent not found, or CLI ignores `.md`), OR official docs are updated to explicitly state `.md` is not supported.

**Evidence requirement:** `kiro-cli agent validate .kiro/agents/probe.md` fails to recognize agent (agent unknown / not found) before Path B may be selected. An auth error alone does NOT select Path B.

### Sub-path B1 — Format only (JSON emit, no capability mapping) [gated]

> This sub-path does NOT depend on #569, but it still requires either verified evidence that JSON is required or an explicit human decision to treat the current official Kiro docs as authoritative despite the unavailable runtime probe. An auth error alone does NOT authorize this code path.

| # | Task | Scope | Depends on |
|---|---|---|---|
| B1-1 | Add `AgentProfile` Go struct (or equivalent inline map) for the Kiro agent JSON schema fields: `name`, `description`, `tools`, `allowedTools`, `resources`, `prompt`, `model`. | code | — |
| B1-2 | Replace `copyCanonicalDefaultAgent` call in `kiro.go:39-44` with a new `emitKiroAgentJSON` helper that reads canonical `.md` frontmatter (name, description, system prompt) and marshals to `.kiro/agents/<name>.json`. Emit `tools: []` and `allowedTools: []` as empty arrays until #569 lands. | code | B1-1 |
| B1-3 | Update `kiro.go:10-13` inline comment to state "`.json` required; `.md` format not recognized by Kiro CLI v3." | code (comment) | B1-2 |
| B1-4 | Update `specs/030-kiro-cli-v3-output-gaps/spec.md` A-003 to retract "no transform needed"; state ".json required per kiro.dev docs; agent transform added in #574." | docs | B1-2 |
| B1-5 | Update `docs/ai-cli-tools/tool-systems/kiro.md` to remove any ambiguity; confirm `.json` is the only recognized format. | docs | B1-2 |
| B1-6 | Regenerate golden fixtures (`testdata/golden/`) to include `.kiro/agents/<name>.json` output. | tests | B1-2 |
| B1-7 | Add/update unit test asserting emitted file is valid JSON and contains required fields (`name`, `description`). | tests | B1-2 |

### Sub-path B2 — Tool capability mapping (requires #569)

> Blocked until #569 delivers the machine-readable capability model.

| # | Task | Scope | Depends on |
|---|---|---|---|
| B2-1 | Wire capability parser from #569 into `emitKiroAgentJSON` to populate `tools` (canonical tool whitelist) and `allowedTools` (auto-approved subset). | code | #569 + B1-2 |
| B2-2 | Update golden fixtures for agents that declare tool capability. | tests | B2-1 |
| B2-3 | Add `allowedTools` cross-verify: assert no entry in `allowedTools` is absent from `tools`. | tests | B2-1 |

---

## Shared Tasks (both paths)

| # | Task | Notes |
|---|---|---|
| S-1 | Obtain Kiro CLI login for runtime format verification, or record an explicit human decision to treat official docs as authoritative | Prerequisite for selecting Path A or Path B; must precede any code change |
| S-2 | After path selection, update `specs/031-cross-cli-agent-tools-alignment/research.md:37` to resolve the "contested" note | Links back to the cross-CLI epic |

---

## Sequencing

```
S-1 (auth + verify)
  │
  ├── confirmed .md tolerated → Path A (A-1 → A-2 → A-3 → A-4)
  │
  └── .json required          → Path B
                                  └── B1 tasks (now, parallel-safe)
                                        └── B2 tasks (after #569)
```

---

## Acceptance Criteria

- A single, unambiguous statement of Kiro's custom agent format is present in both `kiro.md` and `specs/030` A-003, citing evidence.
- `kiro.go` emits files with the verified extension (`.md` or `.json`).
- If JSON: emitted files are valid JSON; golden fixtures updated; unit test green.
- If JSON with capability: `tools`/`allowedTools` populated from canonical model; no `allowedTools` entry absent from `tools`.
- `specs/031-cross-cli-agent-tools-alignment/research.md:37` contradiction note resolved.

---

## Out of Scope

- Kiro steering, specs, hooks, or MCP changes — covered by spec 030 (done).
- Multi-agent emission from the library — only `defaultAgentID` agent is in scope for this issue.
- `includeMcpJson`, `mcpServers`, `resources`, or `model` field population — deferred; not in the issue's acceptance criteria.

---

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Auth remains unavailable, runtime verification impossible | Medium | Accept `kiro.md` docs as authoritative; default to Path B (JSON) since it is source-verified |
| #569 delayed, blocking B2 | High | B1 is independent; emit empty `tools`/`allowedTools` arrays as a safe stub |
| Golden test regeneration produces large diff | Low | Scope to kiro-only fixtures; diff is mechanical |

---

## Human Gate

<!-- The human approver records approval here. Do NOT let an AI author this line. -->

Human Gate: PENDING

<!-- When approving: confirm whether runtime verification is available (selects Path A or B),
     and whether B2 (capability mapping) should be deferred until #569 or attempted in parallel. -->
