# Plan: #573 OMP — Add OMP-Native Subagent Frontmatter Transform

**Issue:** [#573](https://github.com/rluisb/lazyai/issues/573)  
**Epic:** #568 (Cross-CLI agent-tools alignment)  
**Research:** `specs/issues/573-omp-agent-frontmatter/research.md`  
**Depends on:** #569 (canonical capability model) — **DO NOT IMPLEMENT until #569 merges**  
**Status:** DRAFT — pending human gate

---

## What

Replace the verbatim `CopyLibraryDirectory` copy of canonical agents in `omp.go` with a transform-based emission that:

1. Drops LazyAI-only frontmatter fields (`role`, `mode`, `temperature`, `steps`).
2. Emits OMP-native subagent frontmatter: `tools` (lowercase allowlist), `thinkingLevel`, `autoloadSkills` (from canonical `skills:`).
3. Adds the vibe-lab managed marker.
4. Preserves the agent body verbatim.

## How

### Constraints

- Do not start implementation before #569 merges. The `tools` allowlist values and `thinkingLevel` enum must come from the canonical capability model that #569 delivers; any pre-#569 hardcoding of those values would be superseded.
- The OMP adapter status is **stable** — no change to the destination path (`.omp/agents/<name>.md`) or directory structure. The agent surface itself (flat, `canonical/agents` → `.omp/agents/`) is unchanged.
- `read-only` agents (`researcher`, `reviewer`, `evidence-verifier`) **must** have `tools` restricted to `read` and `search` only. This is the primary correctness requirement.
- `autoloadSkills` must be omitted (not emitted as empty) when the canonical source has no `skills:` list.
- Must not regress existing `TestOmpAdapter_Install_AgentsAndSkills` or any other passing OMP test.

### Dependencies on #569

#569 will deliver:
- A canonical per-agent capability list using target-neutral tokens such as `read`, `edit`, `shell`, `search`, `web`, `mcp`, and `spawn`.
- Read-only agents will be tagged with `read`/`search` only.
- This plan derives the OMP-native `tools` allowlist from that merged #569 parser/API, not from role-name hardcoding.

---

## Tasks

> These tasks are ordered and gated. Do not begin a task until its predecessor is complete and verified.

### Task 0 — Verify #569 is merged (gate)

- Confirm `#569` is closed and its canonical capability model is available in `packages/cli/library/canonical/agents/*.md` (or wherever #569 places the output).
- Read the #569 implementation to understand the exact token names and which agents receive which tokens.
- **DO NOT proceed to Task 1 until this is done.**

### Task 1 — Add `RewriteAgentForOMP` (RED: write failing test first)

**File:** `packages/cli/internal/adapter/agent_transform.go`

**Test file:** `packages/cli/internal/adapter/omp_adapter_test.go` (or a new `omp_frontmatter_test.go` following the `claudecode_frontmatter_test.go` pattern)

**Red test — write before implementation:**

```go
// TestRewriteAgentForOMP_ResearcherIsReadOnly verifies that the researcher agent,
// which is described as read-only, gets only read+search in its OMP tools list.
func TestRewriteAgentForOMP_ResearcherIsReadOnly(t *testing.T) {
    source := []byte(`---
name: researcher
description: Scout agent — read-only codebase explorer.
role: researcher
mode: all
temperature: 0.2
steps: 10
skills:
  - tdd-planning
---

# System Prompt
You are a research specialist.
`)
    out, err := RewriteAgentForOMP(source, &AdapterContext{})
    if err != nil {
        t.Fatalf("RewriteAgentForOMP: %v", err)
    }
    fm, _, err := frontmatter.ExtractFrontmatter(out)
    if err != nil {
        t.Fatalf("parse emitted frontmatter: %v", err)
    }
    // OMP-native fields must be present
    tools := frontmatter.ExtractField(fm, "tools")
    if !strings.Contains(tools, "read") {
        t.Errorf("tools missing 'read': %s", tools)
    }
    if !strings.Contains(tools, "search") {
        t.Errorf("tools missing 'search': %s", tools)
    }
    // Mutable tools must be absent for read-only agents
    for _, forbidden := range []string{"bash", "edit", "write", "task"} {
        if strings.Contains(tools, forbidden) {
            t.Errorf("tools must not contain %q for read-only agent: %s", forbidden, tools)
        }
    }
    // LazyAI-only fields must not leak
    for _, leaked := range []string{"role", "mode", "temperature", "steps"} {
        if frontmatter.ExtractField(fm, leaked) != "" {
            t.Errorf("LazyAI-only field %q must not appear in OMP output", leaked)
        }
    }
    // autoloadSkills from canonical skills:
    autoload := frontmatter.ExtractField(fm, "autoloadSkills")
    if !strings.Contains(autoload, "tdd-planning") {
        t.Errorf("autoloadSkills must include tdd-planning: %s", autoload)
    }
}

// TestRewriteAgentForOMP_DeployerHasShell verifies that a non-read-only agent
// (deployer) includes bash in its tools list.
func TestRewriteAgentForOMP_DeployerHasShell(t *testing.T) {
    source := []byte(`---
name: deployer
description: Infrastructure, deployment, and CI/CD operations agent.
role: deployer
mode: all
temperature: 0.2
steps: 14
---

# System Prompt
You are a deployment specialist.
`)
    out, err := RewriteAgentForOMP(source, &AdapterContext{})
    if err != nil {
        t.Fatalf("RewriteAgentForOMP: %v", err)
    }
    fm, _, err := frontmatter.ExtractFrontmatter(out)
    if err != nil {
        t.Fatalf("parse: %v", err)
    }
    tools := frontmatter.ExtractField(fm, "tools")
    if !strings.Contains(tools, "bash") {
        t.Errorf("deployer tools must include bash: %s", tools)
    }
    // No autoloadSkills for agents with no skills:
    if autoload := frontmatter.ExtractField(fm, "autoloadSkills"); autoload != "" {
        t.Errorf("autoloadSkills must be absent when no skills: got %q", autoload)
    }
}

// TestRewriteAgentForOMP_ManagedMarker verifies the managed marker is present.
func TestRewriteAgentForOMP_ManagedMarker(t *testing.T) { /* ... */ }

// TestOmpAdapter_Install_AgentFrontmatterContent verifies the integration:
// after Install, the on-disk researcher.md has OMP-native fields, not LazyAI fields.
func TestOmpAdapter_Install_AgentFrontmatterContent(t *testing.T) { /* ... */ }
```

**Implementation — `RewriteAgentForOMP`:**

```go
// RewriteAgentForOMP transforms a library agent into an OMP-native subagent file.
// Output frontmatter: name, description (required), tools (OMP allowlist derived
// from the canonical capability token delivered by #569), thinkingLevel, and
// autoloadSkills (from canonical skills:, omitted when empty).
// LazyAI-only fields (role, mode, temperature, steps) are dropped.
// Body is preserved verbatim after the managed marker.
func RewriteAgentForOMP(source []byte, ctx *AdapterContext) ([]byte, error) {
    fm, body, err := frontmatter.ExtractFrontmatter(source)
    if err != nil {
        return nil, err
    }
    name := strings.TrimSpace(frontmatter.ExtractField(fm, "name"))
    if name == "" {
        return nil, fmt.Errorf("omp adapter: agent source missing name")
    }
    description := strings.TrimSpace(frontmatter.ExtractField(fm, "description"))
    skills := parseYAMLList(frontmatter.ExtractField(fm, "skills"))

    tools := ompToolsForRole(name)          // derives from #569 canonical token
    thinkingLevel := ompThinkingLevel(name) // derives from role posture

    var b strings.Builder
    b.WriteString("---\n")
    b.WriteString("name: ")
    b.WriteString(name)
    b.WriteByte('\n')
    b.WriteString("description: ")
    b.WriteString(yamlDoubleQuote(description))
    b.WriteByte('\n')
    b.WriteString("tools: [")
    b.WriteString(strings.Join(quoteEach(tools), ", "))
    b.WriteString("]\n")
    b.WriteString("thinkingLevel: ")
    b.WriteString(thinkingLevel)
    b.WriteByte('\n')
    if len(skills) > 0 {
        b.WriteString("autoloadSkills: [")
        b.WriteString(strings.Join(quoteEach(skills), ", "))
        b.WriteString("]\n")
    }
    b.WriteString("---\n\n")
    b.WriteString(managedAgentMarker("omp", name))
    b.WriteString("\n\n")
    b.Write(trimLeadingNewlines(body))
    return []byte(b.String()), nil
}
```

**Helper — `ompToolsForGrants(grants []frontmatter.AgentToolGrant) []string`:**

Derives the OMP `tools` allowlist from the canonical capability grants delivered by #569. Read-only grants (`read`, `search`) map to `["read", "search"]`. Full-capability grants map to OMP-native read/search/edit/write/bash/web/MCP/spawn equivalents only where OMP exposes a documented tool name. Do not infer capability from role name once #569 is available.

**Helper — `ompThinkingLevel(name string) string`:**

`"low"` for read-only, `"high"` for planner, `"auto"` for all others.

### Task 2 — Replace verbatim copy in `omp.go` with transform

**File:** `packages/cli/internal/adapter/omp.go`

Replace the `CopyLibraryDirectory` block for agents (lines 48–58) with a transform-based loop that:

1. Reads each selected agent file from the library.
2. Calls `RewriteAgentForOMP(content, ctx)`.
3. Writes the transformed output to the destination path.

Follow the same pattern used by the Claude Code adapter: if `TransformLibraryDirectory` or an equivalent helper exists (check `adapter/shared.go`), use it with a transform function parameter. If not, inline a range loop.

### Task 3 — Validate no LazyAI-only fields in integration test

Extend `TestOmpAdapter_Install_AgentsAndSkills` (or add a sibling test `TestOmpAdapter_Install_AgentFrontmatterContent`) to:

1. After `Install`, read `.omp/agents/researcher.md` from disk.
2. Parse frontmatter.
3. Assert `tools` is present and excludes `bash`/`edit`/`write`.
4. Assert `role`, `mode`, `temperature`, `steps` are absent.
5. Assert `autoloadSkills` contains `tdd-planning`.

This closes the test-coverage gap identified in research §6.

### Task 4 — Update `docs/adapters/omp.md` agent behavior section

In `docs/adapters/omp.md`, update the **Agent Behavior** section (currently line 54–56) to document:

- The frontmatter transform (not verbatim copy).
- Which fields are emitted (`name`, `description`, `tools`, `thinkingLevel`, `autoloadSkills`).
- Which fields are dropped (`role`, `mode`, `temperature`, `steps`).
- The read-only agent restriction.

Update the **Test Coverage** table to include the new frontmatter content test.

### Task 5 — Update `docs/ai-cli-tools/tool-systems/agent-tools-matrix.md` gap column

Change the OMP row gap from:
> ❌ no OMP-native `tools`/`spawns`/`thinkingLevel`. LazyAI-only fields leak; native features unused.

To reflect the closed gap (after implementation passes).

---

## Acceptance Criteria

From issue #573:

1. **[ ] OMP-native frontmatter** — emitted `.omp/agents/<name>.md` contains `tools`, `thinkingLevel`, `autoloadSkills` (when applicable); does not contain `role`, `mode`, `temperature`, `steps`.
2. **[ ] Read-only restriction** — `researcher`, `reviewer`, `evidence-verifier` have `tools: ["read", "search"]` only; `bash`, `edit`, `write` are absent.
3. **[ ] `skills:` → `autoloadSkills`** — canonical `skills:` list maps directly to OMP `autoloadSkills`.
4. **[ ] LazyAI-only fields do not leak** — verified by a test that parses the emitted frontmatter.
5. **[ ] Existing tests pass** — `TestOmpAdapter_Install_AgentsAndSkills`, `TestOmpAdapter_GlobalScope_InstallsAgentsAndSkills`, `TestOmpOutputMapping_AgentsEmitted`, and all MCP tests continue to pass.
6. **[ ] Body preserved** — agent system prompt body is unchanged after the transform.

---

## TDD Plan

- **Mode:** medium (2 behavioral tests + 1 integration test; transform is isolated)
- **Red tests (Task 1):** `TestRewriteAgentForOMP_ResearcherIsReadOnly`, `TestRewriteAgentForOMP_DeployerHasShell`, `TestRewriteAgentForOMP_ManagedMarker` — must fail before any implementation change.
- **Green (Tasks 2–3):** implement `RewriteAgentForOMP`, swap `omp.go` copy, extend integration test.
- **Refactor:** extract helper functions (`ompToolsForRole`, `ompThinkingLevel`) if they grow beyond 10 lines each; keep them in `agent_transform.go` beside the other `Rewrite*` functions.

---

## Rollback Criteria

Revert Tasks 1–2 (restore verbatim `CopyLibraryDirectory`) if:

- Any OMP-installed agent file fails to parse in OMP at runtime.
- The `tools` field causes OMP to reject the agent definition.
- A regression surfaces in any test that was passing before this change.

The verbatim-copy behaviour (current main) is the safe fallback.

---

## Non-goals

- Do not emit `spawns`, `read-summarize`, `output`, `blocking`, or `model` fields — these require design decisions not in scope for #573.
- Do not change the destination path or directory structure.
- Do not change the `isCanonicalAgentFile` filter or `SelectionKey`.
- Do not start before #569 merges.

---

Human Gate: APPROVED by rluisb at 2026-06-30T09:30:00-03:00

<!-- The human approver records approval by replacing the line above with:
     Human Gate: APPROVED — <approver> <date>
     AI-generated approvals are rejected (see .github/copilot-instructions.md §Gate Attestation Integrity). -->
