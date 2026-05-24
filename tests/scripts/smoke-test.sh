#!/usr/bin/env bash
# Smoke Tests — LazyAI Agent Runtime
# Usage: bash tests/scripts/smoke-test.sh

set -euo pipefail

PASS=0
FAIL=0

assert_exists() {
    if [ -e "$1" ]; then
        PASS=$((PASS + 1))
        echo "  ✅ PASS: $2"
    else
        FAIL=$((FAIL + 1))
        echo "  ❌ FAIL: $2"
    fi
}

assert_command() {
    if command -v "$1" >/dev/null 2>&1; then
        PASS=$((PASS + 1))
        echo "  ✅ PASS: $2"
    else
        FAIL=$((FAIL + 1))
        echo "  ❌ FAIL: $2"
    fi
}

assert_json_valid() {
    if echo "$1" | python3 -c 'import sys,json; json.load(sys.stdin)' 2>/dev/null; then
        PASS=$((PASS + 1))
        echo "  ✅ PASS: $2"
    else
        FAIL=$((FAIL + 1))
        echo "  ❌ FAIL: $2"
    fi
}

echo "═══════════════════════════════════════════════════════════════"
echo "  LazyAI Smoke Tests"
echo "  $(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo "═══════════════════════════════════════════════════════════════"
echo ""

# ─── Agents ───
echo "📁 Agents"
for agent in orchestrator builder documenter implementor planner red-team reviewer scout; do
    assert_exists ".opencode/agents/${agent}.md" "Agent: ${agent}"
done

# ─── Skills ───
echo ""
echo "📚 Skills"
for skill in anti-speculation bugfix diagnose extract-standards housekeeping impact-check implement investigate iterate jira-grooming memory-write orchestrate parallel-execution plan process-audit; do
    assert_exists ".opencode/skills/${skill}/SKILL.md" "Skill: ${skill}"
done

# ─── Commands ───
echo ""
echo "⌨️  Commands"
for cmd in commit review test speckit.analyze speckit.checklist speckit.clarify speckit.constitution speckit.implement speckit.plan speckit.specify speckit.tasks; do
    assert_exists ".opencode/commands/${cmd}.md" "Command: ${cmd}"
done

# ─── Dependencies ───
echo ""
echo "🔧 Dependencies"
assert_command "git" "git available"
assert_command "npx" "npx available"
assert_command "node" "node available"

# ─── MCP Servers ───
echo ""
echo "🔌 MCP Servers"
assert_exists ".opencode/opencode.jsonc" "opencode.jsonc config"

# ─── Config Validation ───
echo ""
echo "⚙️  Config"
if [ -f ".opencode/opencode.jsonc" ]; then
    if python3 -c 'import sys,json; json.load(sys.stdin)' < ".opencode/opencode.jsonc" 2>/dev/null; then
        PASS=$((PASS + 1))
        echo "  ✅ PASS: opencode.jsonc is valid JSON"
    else
        FAIL=$((FAIL + 1))
        echo "  ❌ FAIL: opencode.jsonc is invalid JSON"
    fi
else
    FAIL=$((FAIL + 1))
    echo "  ❌ FAIL: opencode.jsonc not found"
fi

# ─── Specify Directory ───
echo ""
echo "📂 Specify"
assert_exists ".specify/" "Specify directory"
assert_exists ".specify/templates/" "Templates directory"

# ─── Tool Schemas ───
echo ""
echo "📖 Tool Schemas"
assert_exists ".opencode/TOOL-SCHEMAS.md" "Tool schemas reference"

# ─── Summary ───
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✅ ${PASS} passed  ❌ ${FAIL} failed"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
