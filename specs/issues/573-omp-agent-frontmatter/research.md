# Research: #573 OMP — Canonical Agents Copied Verbatim; No Native Subagent Frontmatter Transform

**Issue:** [#573](https://github.com/rluisb/lazyai/issues/573)  
**Epic:** #568 (Cross-CLI agent-tools alignment)  
**Blocker:** #569 (canonical capability model — must merge first)  
**Status:** Research complete; implementation blocked on #569

---

## 1. Problem Statement (from issue)

The OMP adapter copies canonical agents **verbatim** into `.omp/agents/<name>.md`.

- LazyAI-only frontmatter (`role`, `mode`, `temperature`, `steps`) is passed through — OMP ignores these silently.
- OMP-native subagent fields (`tools`, `spawns`, `thinkingLevel`, `autoloadSkills`, `read-summarize`) are never emitted.
- Read-only agents (`researcher`, `reviewer`, `evidence-verifier`) are unrestricted at runtime; OMP uses its default tool surface.

---

## 2. Code Evidence

### 2a. Verbatim copy path (`omp.go:48–58`)

```go
// packages/cli/internal/adapter/omp.go:48-58
if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
    Ctx:          ctx,
    SourceSubdir: "canonical/agents",
    SelectionKey: "agents",
    ToDestPath: func(file string) string {
        return filepath.Join(ompDir, "agents", filepath.Base(file))
    },
    IncludeFile: isCanonicalAgentFile,
}); err != nil {
    return nil, err
}
```

`CopyLibraryDirectory` writes the file bytes unchanged. No transform function is applied. This is the root cause.

### 2b. Canonical agent frontmatter (verbatim from sources)

**`researcher.md`** (`packages/cli/library/canonical/agents/researcher.md:1-10`):
```yaml
---
name: researcher
description: Scout agent — read-only codebase explorer. Gathers evidence, maps existing tests, and identifies TDD planning constraints before implementation.
role: researcher
mode: all
temperature: 0.2
steps: 10
skills:
  - tdd-planning
---
```

**`reviewer.md`** (`packages/cli/library/canonical/agents/reviewer.md:1-10`):
```yaml
---
name: reviewer
description: Universal verifier. Quality gates, spec traceability, adversarial testing, security audits. Read-only.
role: reviewer
mode: all
temperature: 0.1
steps: 14
skills:
  - zero-point
---
```

**`implementer.md`** (`packages/cli/library/canonical/agents/implementer.md:1-12`):
```yaml
---
name: implementer
description: Universal implementer — builds from specs, writes tests first, preserves existing tests, and follows the selected TDD mode.
role: implementer
mode: all
temperature: 0.1
steps: 25
skills:
  - build-mode
  - tdd-planning
  - refresh-dev-containers
---
```

**`guide.md`** (`packages/cli/library/canonical/agents/guide.md:1-8`):
```yaml
---
name: guide
description: Front-door default agent. Answers directly, chats naturally, and only suggests or delegates specialists when that improves the outcome.
role: guide
mode: all
temperature: 0.2
steps: 18
---
```

**`deployer.md`** — no `skills:`.  
**`responder.md`** — no `skills:`.  
**`evidence-verifier.md`** — `skills: [architecture-review]`.  
**`planner.md`** — `skills: [tdd-planning]`.

Fields `role`, `mode`, `temperature`, `steps` appear in every canonical agent but are LazyAI-internal. OMP does not define or document any of these.

### 2c. Installed agent is identical to canonical source

`.agents/agents/researcher.md` (current OMP-installed file, checked in the worktree) is byte-for-byte identical to `packages/cli/library/canonical/agents/researcher.md`. Fields `role: researcher`, `mode: all`, `temperature: 0.2`, `steps: 10` are present in the installed output.

### 2d. Existing transform precedents (`agent_transform.go`)

Two prior transforms exist as the pattern to follow:

- `RewriteAgentForClaudeCode` (`agent_transform.go:101-126`): strips to `name` + `description` only; preserves body; adds managed marker.
- `RewriteAgentForOpenCode` (`agent_transform.go:134-151`): strips to `description`; delegates to `BuildOpenCodeAgentFrontmatter`; adds managed marker.

Both use `frontmatter.ExtractFrontmatter` + `frontmatter.ExtractField` to parse, then build a fresh frontmatter block. The OMP transform will follow the same pattern.

The managed agent marker function (`shared.go:17-19`):
```go
func managedAgentMarker(surface, name string) string {
    return fmt.Sprintf(
        "<!-- vibe-lab:managed kind=agent surface=%s name=%s source=.agents/agents/%s.md -->",
        surface, name, name,
    )
}
```

---

## 3. OMP Native Subagent Frontmatter

### 3a. Authoritative specification

Source: `docs/ai-cli-tools/tool-systems/omp.md:64` (verified 2026-06-29):

> **Subagents**: `<scope>/agents/<name>.md`. Frontmatter `name`/`description` (req), `tools`, `model`, `spawns`, `thinkingLevel`, `output`, `blocking`, `autoloadSkills`, `read-summarize`. Resolution: project → user → plugin → bundled.

### 3b. Per-agent tool model (agent-tools-matrix.md, verified 2026-06-29)

| Target | Native restriction | Built-in tool names | Semantics |
|---|---|---|---|
| **OMP** | `tools:` (CSV/YAML subset) + `spawns`, `thinkingLevel`, `autoloadSkills`, `read-summarize` | lowercase: `read`, `bash`, `edit`, `write`, `search`, `task`, … | allowlist (subset of built-ins) |

**Current gap** (`agent-tools-matrix.md:51`):
> OMP: canonical agents **copied verbatim** (`omp.go`). ❌ no OMP-native `tools`/`spawns`/`thinkingLevel`. LazyAI-only fields leak; native features unused.

### 3c. OMP tool names (canonical → OMP)

From `agent-tools-matrix.md:34-40`:

| Canonical capability | OMP tool name |
|---|---|
| file read | `read` |
| file write/edit | `write`, `edit` |
| shell/exec | `bash` |
| search/grep | `search` |
| web fetch/search | `web_search` |
| MCP tools | `mcp__<srv>_<tool>` |
| subagent spawn | `task` / `spawns` |

---

## 4. Canonical Role → OMP Tool Mapping

Using role descriptions and `mode:` as the capability signal (until #569 delivers the canonical model):

| Agent | Canonical role | Posture | OMP `tools` | `thinkingLevel` | `autoloadSkills` |
|---|---|---|---|---|---|
| `guide` | Front door | Conversational | `read`, `search`, `bash`, `edit`, `write`, `task` | `auto` | _(none)_ |
| `researcher` | Scout | **Read-only** | `read`, `search` | `low` | `tdd-planning` |
| `planner` | Planner | Read + write artifacts | `read`, `search`, `edit`, `write` | `high` | `tdd-planning` |
| `implementer` | Implementer | Full | `read`, `search`, `bash`, `edit`, `write` | `auto` | `build-mode`, `tdd-planning`, `refresh-dev-containers` |
| `reviewer` | Verifier | **Read-only** | `read`, `search` | `low` | `zero-point` |
| `deployer` | Release | Shell + read | `read`, `search`, `bash` | `auto` | _(none)_ |
| `responder` | SRE | Full | `read`, `search`, `bash`, `edit`, `write` | `auto` | _(none)_ |
| `evidence-verifier` | Claim checker | **Read-only** | `read`, `search` | `low` | `architecture-review` |

**Read-only agents** (`researcher`, `reviewer`, `evidence-verifier`): `tools` must exclude `bash`, `edit`, `write`, `task`. This directly satisfies the issue acceptance criterion.

**`autoloadSkills`**: derived from canonical `skills:` list. Agents without `skills:` (`guide`, `deployer`, `responder`) get no `autoloadSkills` field.

**`thinkingLevel` mapping** (heuristic, to be confirmed by #569):
- Read-only scout/verifier roles → `low`
- Planning roles → `high`
- Everything else → `auto`

---

## 5. OMP Adapter Surface (docs/adapters/omp.md)

- Agents surface: `.omp/agents/<name>.md` (flat)
- Status: **stable** (verified 2026-06-23, #486)
- No change to directory structure is needed — the transform only affects frontmatter content, not the destination path.
- `isCanonicalAgentFile` filter and `SelectionKey: "agents"` are preserved; only the copy function changes to a transform.

---

## 6. Existing Tests

| Test | File | What it covers |
|---|---|---|
| `TestOmpAdapter_Install_AgentsAndSkills` | `omp_adapter_test.go:13-33` | File presence (`researcher.md`, `reviewer.md`, skills) |
| `TestOmpAdapter_GlobalScope_InstallsAgentsAndSkills` | `omp_adapter_test.go:96-116` | Global scope path (`~/.omp/agent/agents/`) |
| `TestOmpOutputMapping_AgentsEmitted` | `omp_adapter_test.go:180-188` | Output target shape = flat |

**Gap:** No existing test validates frontmatter content of installed agent files. The current test suite only calls `assertExists` — it does not open the file or parse frontmatter. A new test must verify that:

1. OMP-native fields are present (`tools`, `thinkingLevel`, and `autoloadSkills` when applicable).
2. LazyAI-only fields are absent (`role`, `mode`, `temperature`, `steps`).
3. Read-only agents (`researcher`, `reviewer`, `evidence-verifier`) have only `read` and `search` in `tools`.

The pattern for this test is established in `claudecode_frontmatter_test.go` (uses `frontmatter.ExtractFrontmatter` on the emitted file and asserts field presence/absence).

---

## 7. Implementation Shape (pre-plan sketch)

A new `RewriteAgentForOMP(source []byte, ctx *AdapterContext) ([]byte, error)` function in `agent_transform.go` (or `omp.go`):

1. `frontmatter.ExtractFrontmatter(source)` → parse `name`, `description`, `role`, `skills`.
2. Derive `tools` allowlist from `role` using the mapping table above.
3. Derive `thinkingLevel` from `role`.
4. Derive `autoloadSkills` from `skills:` list (if any).
5. Emit OMP-native YAML frontmatter: `name`, `description`, `tools`, `thinkingLevel`, `autoloadSkills` (when non-empty).
6. Append managed marker + preserved body.

The `CopyLibraryDirectory` call in `omp.go:48-58` must be replaced by a loop that reads each file, calls `RewriteAgentForOMP`, and writes the output — following the same pattern as Claude Code's `TransformLibraryDirectory` (if it exists) or an inline loop.

**Dependency:** The exact `tools` allowlist values and `thinkingLevel` enum values must come from the canonical capability model delivered by #569. The mapping table above is a research-time best-effort; #569 may refine or change it.

---

## 8. Summary of Findings

| Finding | Evidence |
|---|---|
| Verbatim copy is the root cause | `omp.go:48-58` (`CopyLibraryDirectory`, no transform) |
| LazyAI-only fields leak | Every canonical agent has `role`, `mode`, `temperature`, `steps`; OMP docs do not define them |
| OMP native fields available | `omp.md:64`: `tools`, `thinkingLevel`, `autoloadSkills`, `spawns`, `read-summarize`, `model`, `output`, `blocking` |
| Read-only agents misidentified | `researcher`, `reviewer`, `evidence-verifier` describe "read-only" in text but `mode: all` in frontmatter |
| `skills:` → `autoloadSkills` mapping is direct | `researcher.skills=[tdd-planning]` → `autoloadSkills: [tdd-planning]` |
| Existing tests do not check frontmatter | `assertExists` only; no content validation |
| Transform pattern is established | `RewriteAgentForClaudeCode` and `RewriteAgentForOpenCode` in `agent_transform.go` |
| Destination path unchanged | `.omp/agents/<name>.md` flat — only content changes |
