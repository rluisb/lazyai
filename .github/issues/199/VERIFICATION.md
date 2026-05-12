# #199 Verification — Manual Test Scenarios

This document contains all commands you need to run manually to verify the #199 fixes.
The interactive wizard tests require a real terminal — these commands are for you to run directly in your terminal.

---

## Prerequisites

```bash
cd ~/projects/teachable/lazyai/packages/cli
git pull --ff-only origin main
make build
```

---

## Scenario 1: OpenCode — Project Scope (Non-Interactive)

```bash
rm -rf /tmp/test-opencode-project
mkdir /tmp/test-opencode-project
cd /tmp/test-opencode-project

~/projects/teachable/lazyai/packages/cli/lazyai-cli init \
  --no-interactive \
  --scope project \
  --tools opencode \
  --preset standard \
  --name test \
  --no-reversa
```

**Expected:**
- Scaffold completes with 107 files
- Warning about opencode debug agent is expected (no real OpenCode installation)
- `.opencode/agents/planner.md` frontmatter shows `model: openai/gpt-5.5` (not `opencode/`)

**Verify frontmatter:**
```bash
head -15 /tmp/test-opencode-project/.opencode/agents/planner.md
```

**Assert 1 — No `opencode/` prefix:**
```bash
grep -E '^model: opencode/' /tmp/test-opencode-project/.opencode/agents/*.md \
  || echo "✓ OK — no opencode/ prefix"
```

**Assert 2 — No `## Model` in body:**
```bash
grep -l '^## Model' /tmp/test-opencode-project/.opencode/agents/*.md \
  || echo "✓ OK — body cleaned"
```

**Assert 3 — mode distribution:**
```bash
grep '^mode:' /tmp/test-opencode-project/.opencode/agents/*.md | sort | uniq -c
```
Expected: 7 agents all with `mode: subagent` (no orchestrator in OpenCode by default)

**Compile MCP:**
```bash
cd /tmp/test-opencode-project
~/projects/teachable/lazyai/packages/cli/lazyai-cli compile --tool opencode
ls -la .opencode/
```

---

## Scenario 2: OpenCode — Workspace Scope (Non-Interactive)

```bash
rm -rf /tmp/test-opencode-workspace
mkdir /tmp/test-opencode-workspace
cd /tmp/test-opencode-workspace

~/projects/teachable/lazyai/packages/cli/lazyai-cli init \
  --no-interactive \
  --scope workspace \
  --tools opencode \
  --preset standard \
  --name test \
  --no-reversa
```

**Verify frontmatter:**
```bash
head -15 /tmp/test-opencode-workspace/.opencode/agents/planner.md
```

**Assert 1 — No `opencode/` prefix:**
```bash
grep -E '^model: opencode/' /tmp/test-opencode-workspace/.opencode/agents/*.md \
  || echo "✓ OK — no opencode/ prefix"
```

**Assert 2 — No `## Model` in body:**
```bash
grep -l '^## Model' /tmp/test-opencode-workspace/.opencode/agents/*.md \
  || echo "✓ OK — body cleaned"
```

---

## Scenario 3: Copilot — Project Scope (Non-Interactive)

```bash
rm -rf /tmp/test-copilot-project
mkdir /tmp/test-copilot-project
cd /tmp/test-copilot-project

~/projects/teachable/lazyai/packages/cli/lazyai-cli init \
  --no-interactive \
  --scope project \
  --tools copilot \
  --preset standard \
  --name test \
  --no-reversa
```

**Verify model distribution:**
```bash
grep -h '^model:' /tmp/test-copilot-project/.github/agents/*.yaml | sort | uniq -c
```

**Expected:**
```
  3 model: claude-opus-4.7      ← orchestrator/planner/reviewer (Frontier)
 30 model: claude-sonnet-4.6    ← all skill-derived (Balanced default)
```

**Zero `claude-sonnet-4.5`:**
```bash
grep -c 'model: claude-sonnet-4.5' /tmp/test-copilot-project/.github/agents/*.yaml \
  || echo "✓ OK — zero 4.5 models"
```

**Compile MCP:**
```bash
cd /tmp/test-copilot-project
~/projects/teachable/lazyai/packages/cli/lazyai-cli compile --tool copilot
cat .vscode/mcp.json | head -30
python3 -c "import json; json.load(open('.vscode/mcp.json'))" && echo "✓ Valid JSON"
```

---

## Scenario 4: Copilot — Workspace Scope (Non-Interactive)

```bash
rm -rf /tmp/test-copilot-workspace
mkdir /tmp/test-copilot-workspace
cd /tmp/test-copilot-workspace

~/projects/teachable/lazyai/packages/cli/lazyai-cli init \
  --no-interactive \
  --scope workspace \
  --tools copilot \
  --preset standard \
  --name test \
  --no-reversa
```

**Verify model distribution:**
```bash
grep -h '^model:' /tmp/test-copilot-workspace/.github/agents/*.yaml | sort | uniq -c
```

**Expected:** Same as project scope — 3 opus-4.7 + 30 sonnet-4.6

---

## Scenario 5: Claude Code — Project Scope (Non-Interactive)

```bash
rm -rf /tmp/test-claude-code-project
mkdir /tmp/test-claude-code-project
cd /tmp/test-claude-code-project

~/projects/teachable/lazyai/packages/cli/lazyai-cli init \
  --no-interactive \
  --scope project \
  --tools claude-code \
  --preset standard \
  --name test \
  --no-reversa
```

**Expected:**
- Scaffold completes with 110 files
- Shows MCP server registration messages
- `installed: 110`

**Compile MCP:**
```bash
cd /tmp/test-claude-code-project
~/projects/teachable/lazyai/packages/cli/lazyai-cli compile --tool claude-code
```

---

## Scenario 6: All Three Tools — Project Scope (Non-Interactive)

```bash
rm -rf /tmp/test-all-tools
mkdir /tmp/test-all-tools
cd /tmp/test-all-tools

~/projects/teachable/lazyai/packages/cli/lazyai-cli init \
  --no-interactive \
  --scope project \
  --tools opencode,claude-code,copilot \
  --preset standard \
  --name test \
  --no-reversa
```

**Expected:**
- All three tool adapters install
- Shows separate summaries per tool

---

## Scenario 7: Interactive Wizard — OpenCode (Manual Terminal Test)

> **Requires a real terminal.** Run this in your terminal, not via AI.

```bash
rm -rf /tmp/test-interactive-opencode
mkdir /tmp/test-interactive-opencode
cd /tmp/test-interactive-opencode

~/projects/teachable/lazyai/packages/cli/lazyai-cli init
```

**Follow the prompts:**
1. **Project name**: `test-interactive`
2. **Scope**: Select `project` (type `1` or use arrow keys)
3. **Tools**: Select `opencode` (type `1` or use arrow keys)
4. **Preset**: Select `standard` (type `2` or use arrow keys)
5. **Enable servers**: Choose `n` (no orchestrator)
6. **Reversa**: Choose `n` (no Scout analysis)

**Expected:** Completes without cancellation, produces same output as `--no-interactive`

---

## Scenario 8: Interactive Wizard — Copilot Workspace (Manual Terminal Test)

> **Requires a real terminal.**

```bash
rm -rf /tmp/test-interactive-copilot-ws
mkdir /tmp/test-interactive-copilot-ws
cd /tmp/test-interactive-copilot-ws

~/projects/teachable/lazyai/packages/cli/lazyai-cli init
```

**Follow the prompts:**
1. **Project name**: `test-interactive`
2. **Scope**: Select `workspace` (type `2` or use arrow keys)
3. **Tools**: Select `copilot` (type `3` or use arrow keys)
4. **Preset**: Select `standard` (type `2` or use arrow keys)
5. **Enable servers**: Choose `n`
6. **Reversa**: Choose `n`

**Expected:** Completes, model distribution matches `3 opus + 30 sonnet`

---

## Scenario 9: Interactive Wizard — All Steps Full Workflow (Manual Terminal Test)

> **Full workflow test — most comprehensive.**

```bash
rm -rf /tmp/test-full-workflow
mkdir /tmp/test-full-workflow
cd /tmp/test-full-workflow

~/projects/teachable/lazyai/packages/cli/lazyai-cli init
```

**Follow ALL prompts:**
1. **Project name**: `full-workflow-test`
2. **Scope**: `workspace`
3. **Tools**: Select `claude-code` (first, then `copilot` using multi-select)
4. **Preset**: `standard`
5. **Features**: Accept defaults
6. **Enable servers**: `y` to orchestrator
7. **Org**: Press Enter for default
8. **Team**: Press Enter for default
9. **Reversa**: `n`

**After completion, verify:**
```bash
# Check all tools were configured
ls -la

# Check MCP servers registered
cat .mcp.json | python3 -c "import json,sys; d=json.load(sys.stdin); print(f\"Servers: {len(d.get('mcpServers',{}))}\")"

# Check agents exist for both tools
ls .claude/agents/ 2>/dev/null | head -5
ls .github/agents/ 2>/dev/null | head -5
```

---

## Scenario 10: Minimal Preset — OpenCode (Non-Interactive)

```bash
rm -rf /tmp/test-opencode-minimal
mkdir /tmp/test-opencode-minimal
cd /tmp/test-opencode-minimal

~/projects/teachable/lazyai/packages/cli/lazyai-cli init \
  --no-interactive \
  --scope project \
  --tools opencode \
  --preset minimal \
  --name test \
  --no-reversa
```

**Expected:** Fewer files than `standard` (minimal preset)

---

## Scenario 11: Full Preset — Copilot (Non-Interactive)

```bash
rm -rf /tmp/test-copilot-full
mkdir /tmp/test-copilot-full
cd /tmp/test-copilot-full

~/projects/teachable/lazyai/packages/cli/lazyai-cli init \
  --no-interactive \
  --scope project \
  --tools copilot \
  --preset full \
  --name test \
  --no-reversa
```

**Expected:** More files than `standard` (full preset)

---

## Scenario 12: Tier Override Smoke Test (Copilot)

This test requires modifying a skill file temporarily. **Do NOT commit this change.**

```bash
# Create a temporary worktree or copy the skill file
cp ~/projects/teachable/lazyai/packages/cli/library/skills/red-team-plan.md \
   /tmp/red-team-plan-test.md

# Add tier: frontier to the copy (we'll use it directly)
```

**Edit `/tmp/red-team-plan-test.md` frontmatter to add `tier: frontier` and `risk: 5`:**

```yaml
---
name: red-team-plan
description: Read-only adversarial design review...
tier: frontier    # ADD THIS
risk: 5           # ADD THIS
argument-hint: "[feature-plan-path]"
...
```

**Now create a test and use that skill file:**

```bash
rm -rf /tmp/test-tier-override
mkdir /tmp/test-tier-override
cd /tmp/test-tier-override

# This would require the skill file to be in the library path
# For this smoke test, we verify via unit tests instead:

cd ~/projects/teachable/lazyai/packages/cli
go test -v -run "TestSkillSpecOrDefault_WithTierFrontier" ./internal/adapter/
```

**Expected:** Test passes — `tier: frontier` correctly parsed

---

## Quick Reference — All Assert Commands

```bash
# Run all non-interactive tests quickly
echo "=== OpenCode Project ===" && \
  grep -E '^model: opencode/' /tmp/test-opencode-project/.opencode/agents/*.md || echo "OK" && \
echo "=== Copilot Project ===" && \
  grep -h '^model:' /tmp/test-copilot-project/.github/agents/*.yaml | sort | uniq -c

# Go test suite
cd ~/projects/teachable/lazyai/packages/cli
go test ./internal/adapter/ ./internal/models/

# Catalog future-proof check
grep -A 4 'CopilotCatalog' internal/models/catalog.go
```

---

## After Running All Scenarios

Run the cleanup and investigation commands:

```bash
# Cleanup
rm -rf /tmp/test-*

# Investigate contract warnings
cd ~/projects/teachable/lazyai/packages/cli
~/projects/teachable/lazyai/packages/cli/lazyai-cli compile --strict-contracts 2>&1 | head -50
```
