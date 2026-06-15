#!/usr/bin/env bash
# drift-check.sh — Compare spec claims vs actual code, report violations
# Usage: ./drift-check.sh --spec <file> [--root <dir>] [--json]

set -euo pipefail

SPEC_FILE=""
ROOT_DIR="."
JSON_OUTPUT=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --spec) SPEC_FILE="$2"; shift 2 ;;
        --root) ROOT_DIR="$2"; shift 2 ;;
        --json) JSON_OUTPUT=true; shift ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

if [[ -z "$SPEC_FILE" ]]; then
    echo "Usage: $0 --spec <file> [--root <dir>] [--json]"
    exit 1
fi

if [[ ! -f "$SPEC_FILE" ]]; then
    echo "❌ Spec file not found: $SPEC_FILE"
    exit 1
fi

echo "🔍 Drift Scope — Spec vs Implementation Check"
echo "   Spec: $SPEC_FILE"
echo "   Root: $ROOT_DIR"
echo ""

# Parse spec sections
INVARIANTS=()
INTERFACES=()
CONSTRAINTS=()
TASKS=()

current_section=""
while IFS= read -r line; do
    if [[ "$line" =~ ^##\ §V ]]; then current_section="V"; continue; fi
    if [[ "$line" =~ ^##\ §I ]]; then current_section="I"; continue; fi
    if [[ "$line" =~ ^##\ §C ]]; then current_section="C"; continue; fi
    if [[ "$line" =~ ^##\ §T ]]; then current_section="T"; continue; fi
    if [[ "$line" =~ ^##\ ]]; then current_section=""; continue; fi

    if [[ -n "$current_section" && "$line" =~ ^\| ]]; then
        case "$current_section" in
            V) INVARIANTS+=("$line") ;;
            I) INTERFACES+=("$line") ;;
            C) CONSTRAINTS+=("$line") ;;
            T) TASKS+=("$line") ;;
        esac
    fi
done < "$SPEC_FILE"

# Check invariants
echo "📋 Checking §V Invariants..."
VIOLATIONS=()
for inv in "${INVARIANTS[@]}"; do
    # Skip header row
    [[ "$inv" =~ ^\|\ *id\ *\| ]] && continue

    # Extract rule text
    rule=$(echo "$inv" | cut -d'|' -f3 | xargs)
    [[ -z "$rule" ]] && continue

    # Search for rule keywords in codebase
    keyword=$(echo "$rule" | grep -oP '\w+' | head -3 | tr '\n' ' ')
    if grep -rq "$keyword" "$ROOT_DIR" --include="*.go" --include="*.ts" --include="*.js" --include="*.py" --include="*.rb" 2>/dev/null; then
        echo "   ✅ $rule"
    else
        echo "   ❌ $rule (not found in codebase)"
        VIOLATIONS+=("§V|$rule|missing from codebase")
    fi
done

echo ""
echo "📋 Checking §I Interfaces..."
for iface in "${INTERFACES[@]}"; do
    [[ "$iface" =~ ^\|\ *id\ *\| ]] && continue
    sig=$(echo "$iface" | cut -d'|' -f3 | xargs)
    [[ -z "$sig" ]] && continue

    # Extract function name
    func=$(echo "$sig" | grep -oP '\w+\(' | head -1 | tr -d '(')
    if [[ -n "$func" ]] && grep -rq "$func" "$ROOT_DIR" --include="*.go" --include="*.ts" --include="*.js" --include="*.py" --include="*.rb" 2>/dev/null; then
        echo "   ✅ $sig"
    else
        echo "   ❌ $sig (not found)"
        VIOLATIONS+=("§I|$sig|missing from codebase")
    fi
done

echo ""
echo "📋 Checking §C Constraints..."
for constraint in "${CONSTRAINTS[@]}"; do
    [[ "$constraint" =~ ^\|\ *id\ *\| ]] && continue
    rule=$(echo "$constraint" | cut -d'|' -f3 | xargs)
    [[ -z "$rule" ]] && continue

    # Search for constraint keywords
    keyword=$(echo "$rule" | grep -oP '\w+' | head -3 | tr '\n' ' ')
    if grep -rq "$keyword" "$ROOT_DIR" --include="*.go" --include="*.ts" --include="*.js" --include="*.py" --include="*.rb" 2>/dev/null; then
        echo "   ✅ $rule"
    else
        echo "   ⚠️  $rule (not verified)"
    fi
done

echo ""
echo "📊 Drift Summary"
echo "   Total violations: ${#VIOLATIONS[@]}"
if [[ ${#VIOLATIONS[@]} -gt 0 ]]; then
    echo ""
    echo "   Violations:"
    for v in "${VIOLATIONS[@]}"; do
        echo "   - $v"
    done
    echo ""
    echo "💡 Run ricochet to backpropagate violations as new invariants"
else
    echo "   ✅ No drift detected"
fi
