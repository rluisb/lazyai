#!/usr/bin/env bash
# test-skills.sh — Validate skills, scripts, and agent wiring
# Usage: ./test-skills.sh [--verbose]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SKILLS_DIR="$ROOT_DIR/skills"
AGENTS_DIR="$ROOT_DIR/agents"

VERBOSE="${1:-}"
PASS=0
FAIL=0
WARN=0

pass() { PASS=$((PASS + 1)); [ "$VERBOSE" = "--verbose" ] && echo "  ✅ $1"; }
fail() { FAIL=$((FAIL + 1)); echo "  ❌ $1"; }
warn() { WARN=$((WARN + 1)); [ "$VERBOSE" = "--verbose" ] && echo "  ⚠️  $1"; }

echo "🧪 OpenCode Skill & Agent Validation"
echo "====================================="
echo ""

# ── 1. Validate SKILL.md YAML frontmatter ──
echo "📋 Checking SKILL.md YAML frontmatter..."
for skill_dir in "$SKILLS_DIR"/*/; do
    [ -d "$skill_dir" ] || continue
    skill_name=$(basename "$skill_dir")

    # Skip archived and meta directories
    case "$skill_name" in
        _*) continue ;;
    esac

    skill_file="$skill_dir/SKILL.md"

    if [ ! -f "$skill_file" ]; then
        fail "$skill_name: Missing SKILL.md"
        continue
    fi

    # Check YAML frontmatter exists and is closed
    if ! head -1 "$skill_file" | grep -q '^---'; then
        fail "$skill_name: No YAML frontmatter"
        continue
    fi

    # Check frontmatter is closed (second ---)
    fm_close=$(grep -n '^---' "$skill_file" | head -2 | tail -1 | cut -d: -f1)
    if [ -z "$fm_close" ] || [ "$fm_close" -lt 2 ]; then
        fail "$skill_name: Unclosed frontmatter"
        continue
    fi

    # Extract frontmatter and validate required fields
    fm=$(sed -n "2,$((fm_close - 1))p" "$skill_file")

    # Check for required fields
    if echo "$fm" | grep -q '^name:'; then
        # Check name has a value
        name_val=$(echo "$fm" | grep '^name:' | head -1 | sed 's/^name:[[:space:]]*//')
        if [ -n "$name_val" ] && [ "$name_val" != "''" ] && [ "$name_val" != '""' ]; then
            pass "$skill_name: name='$name_val'"
        else
            fail "$skill_name: Empty name field"
        fi
    else
        fail "$skill_name: Missing 'name' field"
    fi

    if echo "$fm" | grep -q '^description:'; then
        desc_val=$(echo "$fm" | grep '^description:' | head -1 | sed 's/^description:[[:space:]]*//')
        if [ -n "$desc_val" ] && [ "$desc_val" != "''" ] && [ "$desc_val" != '""' ]; then
            pass "$skill_name: description present"
        else
            fail "$skill_name: Empty description field"
        fi
    else
        fail "$skill_name: Missing 'description' field"
    fi

    # Check for unquoted colons in description (common YAML error)
    if echo "$fm" | grep '^description:' | grep -qE ':[^"'"'"'].*:'; then
        warn "$skill_name: Description may have unquoted colons"
    fi
done

echo ""

# ── 2. Validate script syntax ──
echo "🔧 Checking script syntax (bash -n)..."
for skill_dir in "$SKILLS_DIR"/*/; do
    [ -d "$skill_dir" ] || continue
    skill_name=$(basename "$skill_dir")
    scripts_dir="$skill_dir/scripts"

    [ -d "$scripts_dir" ] || continue

    for script in "$scripts_dir"/*.sh; do
        [ -f "$script" ] || continue
        script_name=$(basename "$script")

        if bash -n "$script" 2>/dev/null; then
            pass "$skill_name/$script_name: syntax OK"
        else
            fail "$skill_name/$script_name: syntax error"
            bash -n "$script" 2>&1 | head -3 | sed 's/^/    /'
        fi
    done
done

# Also check shared scripts
echo ""
echo "🔧 Checking shared scripts..."
if [ -d "$ROOT_DIR/scripts" ]; then
    for script in "$ROOT_DIR/scripts"/*.sh; do
        [ -f "$script" ] || continue
        script_name=$(basename "$script")

        if bash -n "$script" 2>/dev/null; then
            pass "scripts/$script_name: syntax OK"
        else
            fail "scripts/$script_name: syntax error"
            bash -n "$script" 2>&1 | head -3 | sed 's/^/    /'
        fi
    done
fi

echo ""

# ── 3. Validate agent files ──
echo "🤖 Checking agent files..."
for agent_file in "$AGENTS_DIR"/*.md; do
    [ -f "$agent_file" ] || continue
    agent_name=$(basename "$agent_file" .md)

    # Check for required sections
    if grep -q "permission.task" "$agent_file"; then
        pass "$agent_name: has permission.task"
    else
        warn "$agent_name: missing permission.task"
    fi

    if grep -qi "skill" "$agent_file"; then
        pass "$agent_name: references skills"
    else
        warn "$agent_name: no skill references"
    fi
done

echo ""

# ── 4. Cross-reference: skills mentioned in agents vs actual skills ──
echo "🔗 Cross-referencing skills in agents..."
for skill_dir in "$SKILLS_DIR"/*/; do
    [ -d "$skill_dir" ] || continue
    skill_name=$(basename "$skill_dir")

    case "$skill_name" in
        _*) continue ;;
    esac

    found_in=$(grep -rl "$skill_name" "$AGENTS_DIR"/*.md 2>/dev/null | wc -l | tr -d ' ')
    if [ "$found_in" -eq 0 ]; then
        warn "$skill_name: not referenced by any agent"
    else
        pass "$skill_name: referenced by $found_in agent(s)"
    fi
done

echo ""

# ── 5. Check for orphaned scripts ──
echo "📦 Checking script references in SKILL.md..."
for skill_dir in "$SKILLS_DIR"/*/; do
    [ -d "$skill_dir" ] || continue
    skill_name=$(basename "$skill_dir")
    scripts_dir="$skill_dir/scripts"

    [ -d "$scripts_dir" ] || continue

    for script in "$scripts_dir"/*.sh; do
        [ -f "$script" ] || continue
        script_name=$(basename "$script")

        # Check if script is referenced in SKILL.md
        if grep -q "$script_name" "$skill_dir/SKILL.md" 2>/dev/null; then
            pass "$skill_name/$script_name: referenced in SKILL.md"
        else
            warn "$skill_name/$script_name: not referenced in SKILL.md"
        fi
    done
done

echo ""

# ── Summary ──
echo "====================================="
echo "📊 Results: ✅ $PASS passed | ❌ $FAIL failed | ⚠️  $WARN warnings"

if [ "$FAIL" -gt 0 ]; then
    echo "❌ Validation failed — fix errors before deploying"
    exit 1
else
    echo "✅ All checks passed"
    exit 0
fi
