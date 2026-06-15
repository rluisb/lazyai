#!/usr/bin/env bash
# skill-audit.sh — Validate skill ↔ agent mapping consistency
# Usage: ./skill-audit.sh

set -euo pipefail

SKILLS_DIR="${OPENCODE_CONFIG_DIR:-$HOME/.config/opencode}/skills"
AGENTS_DIR="${OPENCODE_CONFIG_DIR:-$HOME/.config/opencode}/agents"

echo "=== Skill ↭ Agent Audit ==="
echo ""

# Build skill set
echo "Skills found:"
for skill in "$SKILLS_DIR"/*/SKILL.md; do
    [ -f "$skill" ] || continue
    SKILL_NAME=$(basename "$(dirname "$skill")")
    if [ "$SKILL_NAME" = "_archived" ]; then continue; fi
    echo "  📦 $SKILL_NAME ($(wc -l < "$skill" | tr -d ' ') lines)"
done

echo ""
echo "Agents → Skill references:"
for agent in "$AGENTS_DIR"/*.md; do
    [ -f "$agent" ] || continue
    AGENT_NAME=$(basename "$agent" .md)
    if [ "$AGENT_NAME" = "AGENTS" ]; then continue; fi

    SKILL_REF=$(grep -oP '(?<=load |Load |skill: |Skill: |loads )["\x60]?[a-z-]+["\x60]?' "$agent" 2>/dev/null | head -5 | tr '\n' ' ')
    SKILL_REF=$(grep -i "storm-scout\|build-mode\|zero-point\|battle-bus\|reboot-van\|slurp-juice\|the-vault\|supply-llama" "$agent" 2>/dev/null | head -3 | tr '\n' ' ')

    echo "  🤖 $AGENT_NAME → $SKILL_REF"

    # Check if referenced skill exists
    if echo "$SKILL_REF" | grep -q "storm-scout"; then
        [ -f "$SKILLS_DIR/storm-scout/SKILL.md" ] || echo "    ⚠️ storm-scout not found!"
    fi
    if echo "$SKILL_REF" | grep -q "build-mode"; then
        [ -f "$SKILLS_DIR/build-mode/SKILL.md" ] || echo "    ⚠️ build-mode not found!"
    fi
    if echo "$SKILL_REF" | grep -q "zero-point"; then
        [ -f "$SKILLS_DIR/zero-point/SKILL.md" ] || echo "    ⚠️ zero-point not found!"
    fi
done

echo ""
echo "Unused skills (not referenced by any agent):"
for skill in "$SKILLS_DIR"/*/SKILL.md; do
    [ -f "$skill" ] || continue
    SKILL_NAME=$(basename "$(dirname "$skill")")
    if [ "$SKILL_NAME" = "_archived" ]; then continue; fi

    if ! grep -rq "$SKILL_NAME" "$AGENTS_DIR"/*.md 2>/dev/null; then
        echo "  📭 $SKILL_NAME — no agent loads this skill"
    fi
done

echo ""
echo "✅ Audit complete"
