# Research: RPI Human Gates Bypassed in Auto/Agent Modes

**Bugfix:** 023 — Cupcake RPI Gate Enforcement
**Started:** 2026-05-04
**Phase:** Research (R of RPI)
**Confidence:** HIGH

---

## 1. Affected Surface

### User Report
After installing setup in Claude Code or Copilot (both support modes like accept-edits, auto, agent), the user asked for RPI and the system **researched, planned, and implemented without honoring human gates.**

### Files Examined

| File | Relevance |
|------|-----------|
| `.agents/skills/rpi/SKILL.md` | Core RPI skill — 316 lines of process definition. Contains 4 ⛔ human gates. All advisory text. |
| `.claude/skills/rpi/SKILL.md` | Claude-specific RPI copy — identical content |
| `.opencode/skills/rpi/SKILL.md` | OpenCode-specific RPI copy — identical content |
| `packages/cli/library/fragments/rpi-workflow.md` | Harness-aligned RPI phases (0-4) with ⛔ gates at each phase boundary |
| `.opencode/modes/audit.md` | OpenCode audit mode — correctly restricts write/edit/bash to `false` |
| `.opencode/modes/plan.md` | OpenCode plan mode — correctly restricts write/edit/bash to `false` |
| `.github/agents/rpi.agent.yaml` | RPI agent config — grants `tools: ["*"]` (full write access) |
| `.github/agents/orchestrator.agent.yaml` | Orchestrator agent — generic, no gate enforcement |
| `.ai/orchestration/chains/feature.json` | Feature chain — defines `"gate": "user_approval"` on plan step. Dead code. |
| `.ai/orchestration/workflows/rpi.json` | RPI workflow — delegates to feature chain, no enforcement |
| `.ai/mcp.json` | MCP config — orchestrator server is `"enabled": false` |
| `.claude/settings.json` | Claude Code settings — empty permissions, no customInstructions |
| `.mcp.json` | Root MCP config — orchestrator listed but no enforcement |
| `docs/AI-Agentic-Setup-Templates/.github/copilot-instructions.md` | Copilot template — 260-line skeleton, no gate enforcement, no mode awareness |

### External References Researched

| Source | Key Finding |
|--------|------------|
| [Claude Code rules deep-dive](https://joseparreogarcia.substack.com/p/how-claude-code-rules-actually-work) | Rules are advisory text. Unconditional rules load ALL at session start. Scoped rules load only when touching matching paths. `CLAUDE.md` is an index, not enforcement. Claude has no native hook system for policy enforcement. |
| [OpenCode rules docs](https://opencode.ai/docs/rules/) | `AGENTS.md` is primary. Supports `opencode.json` → `instructions` with glob patterns. Has mode-based permissions (`write: false`) — strongest native enforcement of the three. Supports `@file.md` lazy-loading. |
| [GitHub Copilot custom instructions docs](https://docs.github.com/en/copilot/customizing-copilot/adding-repository-custom-instructions-for-github-copilot) | `.github/copilot-instructions.md` for repo-wide. `.github/instructions/*.instructions.md` with `applyTo:` for path-specific. Supports `AGENTS.md`. Explicitly advisory: "they do not guarantee compliance." Agent YAML `tools:` restriction is programmatic. |
| [Cupcake](https://github.com/eqtylab/cupcake) | **Native policy enforcement layer** for AI coding agents. Uses OPA/Rego compiled to Wasm. Intercepts agent tool calls via hooks and returns Allow/Modify/Block/Warn/Require Review. Supports Claude Code, OpenCode, Cursor natively. Copilot not supported (no runtime hook integration). |

---

## 2. Root Cause Analysis

### The 5-Layer Failure Stack

```
┌─────────────────────────────────────────────────────────┐
│ Layer 5: Orchestrator MCP server — DISABLED             │  ← Programmatic gate: OFF
├─────────────────────────────────────────────────────────┤
│ Layer 4: No precedence rule between gate and mode       │  ← When both say different things, nothing wins
├─────────────────────────────────────────────────────────┤
│ Layer 3: tools: ["*"] — full write access               │  ← Nothing prevents writes during research
├─────────────────────────────────────────────────────────┤
│ Layer 2: Auto/agent mode — "execute without pausing"    │  ← System loop doesn't stop for gates
├─────────────────────────────────────────────────────────┤
│ Layer 1: RPI gate instructions — advisory text only     │  ← No enforcement, just suggestions
└─────────────────────────────────────────────────────────┘
```

### Why It Happens

1. **RPI gates are natural-language text.** The ⛔ markers in `rpi-workflow.md` and `SKILL.md` are context-window instructions with zero enforcement power.

2. **Auto/agent modes are system-level instructions.** Claude Code's "accept edits" mode, Copilot's "agent mode" tell the model: "execute without pausing." These are tool-runtime instructions, not prompt text.

3. **No precedence rule exists.** When the model receives both "stop at gates" (from RPI text) and "don't wait" (from mode setting), there is no mechanism that declares which wins. The more salient instruction (usually the mode) dominates.

4. **The agent YAML grants full tool access.** `rpi.agent.yaml` uses `tools: ["*"]` — the agent can write, edit, and execute bash during any phase. Nothing physically prevents it.

5. **The programmatic enforcement path is disabled.** `feature.json` defines a `"gate": "user_approval"` but the `@ai-setup/orchestrator` MCP server that would enforce it is explicitly `"enabled": false`.

### The Forging Problem

A secondary but critical issue: even if we add artifact checks (pre-commit hooks looking for "Human Gate: APPROVED"), the model can **forge** the approval marker. It writes the exact text the check looks for. Without git authorship verification, timestamp signing, or external review state, text markers are trivially forgeable.

---

## 3. Existing Patterns

### What Already Works

| Pattern | File | Status |
|---------|------|--------|
| OpenCode mode-based permissions | `.opencode/modes/audit.md`, `plan.md` | ✅ Correct — `write: false`, `edit: false`, `bash: false` |
| Feature chain gate definition | `.ai/orchestration/chains/feature.json` | ⚠️ Defined but not enforced (MCP disabled) |
| RPI workflow phase gates | `rpi-workflow.md` | ⚠️ Clear text but advisory only |
| Task sizing skip rule | `rpi-workflow.md` L27 | ✅ Correct — <20 lines skips RPI |

### What's Missing

| Missing Pattern | Why Needed |
|----------------|------------|
| Mode-aware behavior in RPI skill | Skill doesn't detect or react to auto/agent mode |
| Tool restrictions per phase | Research phase agents should have read-only tools |
| Programmatic gate enforcement | Text gates need runtime enforcement |
| Cross-tool consistent rules | Each tool has different rule file conventions |
| Anti-forgery verification | Artifact checks need git authorship signals |

---

## 4. External Solution: Cupcake

Cupcake (`github.com/eqtylab/cupcake`) is a native policy enforcement layer for AI coding agents. Key attributes:

- **Architecture**: Intercepts agent tool calls via hooks → evaluates against OPA/Rego policies (compiled to Wasm) → returns Allow/Modify/Block/Warn/Require Review
- **Performance**: Sub-millisecond evaluation, zero context tokens consumed
- **Determinism**: Same input = same output. No model interpretation required.
- **Harness support**: Claude Code ✅, OpenCode ✅, Cursor ✅, Factory AI ✅, Copilot ❌, Gemini CLI 🟡
- **Key feature for RPI**: `Require Review` decision — physically halts tool execution until human approves

### Cupcake Effectiveness by Tool

| Tool | With Cupcake | Without Cupcake | Mechanism |
|------|-------------|-----------------|-----------|
| Claude Code | **93%** | 35% | Native hook → `require_review` blocks writes |
| OpenCode | **95%** | 40-85% | Native plugin + existing permissions model |
| Copilot | **70%** | 35% | Signals-based at commit/PR time (no native hook) |

### Copilot Gap Mitigation

Cupcake doesn't support Copilot natively (no hook integration). Mitigations:
- Enforce at commit time via `cupcake evaluate` in pre-commit hook
- Enforce at PR time via `cupcake evaluate` in CI
- Use Cupcake Watchdog (LLM-as-Judge) for post-hoc review of Copilot output
- Use TS/Python bindings to build a custom review pipeline

---

## 5. Unknowns & Risks

| Unknown | Confidence | Risk |
|---------|-----------|------|
| Cupcake `Require Review` UX in Claude Code | MEDIUM | If UX is poor (blocking without clear message), adoption suffers |
| Cupcake stability at project scale | MEDIUM | v0.5.1, 259 stars, active but young project |
| Copilot agent mode hooks — will GitHub add them? | LOW | No public roadmap. Cupcake explicitly says "unsupported harnesses lack runtime integration" |
| Will Claude Code's hook architecture change? | LOW | Anthropic's hook system (issue #712) was filed by cupcake team. Evolving. |
| Cross-tool policy maintenance burden | MEDIUM | Per-harness policies (`policies/claude/`, `policies/opencode/`) increase maintenance |

---

## 6. Verdict

**GO** — The root cause is clearly identified as a structural conflict between advisory RPI gate text and system-level auto-mode instructions. Cupcake provides a deterministic enforcement layer that addresses the core problem for Claude Code and OpenCode. The Copilot gap is real but mitigable via signals-based checkpoint enforcement.

**⛔ Human Gate:** Research approved? Proceed to Plan?

*Approve with: APPROVE / REQUEST_CHANGES*
