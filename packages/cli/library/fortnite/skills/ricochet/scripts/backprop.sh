#!/usr/bin/env bash
# backprop.sh — Parse test failures and update spec invariants
# Usage: ./backprop.sh --test-output <file> --spec <file> [--dry-run]
#
# Classifies failures and updates spec with §V (invariants) or §B (bugs)

set -euo pipefail

TEST_OUTPUT=""
SPEC_FILE=""
DRY_RUN=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --test-output) TEST_OUTPUT="$2"; shift 2 ;;
        --spec) SPEC_FILE="$2"; shift 2 ;;
        --dry-run) DRY_RUN=true; shift ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

if [[ -z "$TEST_OUTPUT" || -z "$SPEC_FILE" ]]; then
    echo "Usage: $0 --test-output <file> --spec <file> [--dry-run]"
    exit 1
fi

if [[ ! -f "$TEST_OUTPUT" ]]; then
    echo "❌ Test output file not found: $TEST_OUTPUT"
    exit 1
fi

if [[ ! -f "$SPEC_FILE" ]]; then
    echo "❌ Spec file not found: $SPEC_FILE"
    echo "   Create spec first with: compact-blueprint skill"
    exit 1
fi

echo "🪨 Ricochet — Backpropagation Protocol"
echo "   Test output: $TEST_OUTPUT"
echo "   Spec file: $SPEC_FILE"
echo ""

# Parse test failures
FAILURES=()
while IFS= read -r line; do
    # Match common test failure patterns
    if [[ "$line" =~ (FAIL|Error|panic|exception|AssertionError) ]]; then
        FAILURES+=("$line")
    fi
done < "$TEST_OUTPUT"

if [[ ${#FAILURES[@]} -eq 0 ]]; then
    echo "✅ No test failures found. Nothing to backpropagate."
    exit 0
fi

echo "📊 Found ${#FAILURES[@]} failure(s) to classify"
echo ""

# Classify and generate updates
NEW_INVARIANTS=()
NEW_BUGS=()

for failure in "${FAILURES[@]}"; do
    # Extract key info
    file=$(echo "$failure" | grep -oP '[\w/]+\.(\w+):\d+' | head -1 || echo "unknown")
    message=$(echo "$failure" | sed 's/.*: //' | head -c 100)

    # Classify: if it's a null/nil/undefined check → invariant
    # Otherwise → bug entry
    if [[ "$failure" =~ (null|nil|undefined|NPE|NullPointerException|TypeError) ]]; then
        NEW_INVARIANTS+=("| V$(date +%s) | guard @ $file: $message | test: $file |")
    else
        NEW_BUGS+=("| B$(date +%s) | $message | TODO | open |")
    fi
done

# Generate updates
if [[ "$DRY_RUN" == true ]]; then
    echo "🔍 [DRY RUN] Would update spec with:"
    echo ""
    if [[ ${#NEW_INVARIANTS[@]} -gt 0 ]]; then
        echo "## §V Invariants (new)"
        for inv in "${NEW_INVARIANTS[@]}"; do
            echo "  $inv"
        done
        echo ""
    fi
    if [[ ${#NEW_BUGS[@]} -gt 0 ]]; then
        echo "## §B Bugs (new)"
        for bug in "${NEW_BUGS[@]}"; do
            echo "  $bug"
        done
        echo ""
    fi
else
    # Update spec file
    echo "📝 Updating spec..."

    # Check if §V section exists
    if grep -q "## §V Invariants" "$SPEC_FILE"; then
        # Append to existing section
        for inv in "${NEW_INVARIANTS[@]}"; do
            sed -i '' "/## §V Invariants/a\\
$inv" "$SPEC_FILE"
        done
    elif [[ ${#NEW_INVARIANTS[@]} -gt 0 ]]; then
        # Create new section
        echo "" >> "$SPEC_FILE"
        echo "## §V Invariants" >> "$SPEC_FILE"
        echo "| id | rule | evidence |" >> "$SPEC_FILE"
        echo "|----|------|----------|" >> "$SPEC_FILE"
        for inv in "${NEW_INVARIANTS[@]}"; do
            echo "$inv" >> "$SPEC_FILE"
        done
    fi

    # Check if §B section exists
    if grep -q "## §B Bugs" "$SPEC_FILE"; then
        for bug in "${NEW_BUGS[@]}"; do
            sed -i '' "/## §B Bugs/a\\
$bug" "$SPEC_FILE"
        done
    elif [[ ${#NEW_BUGS[@]} -gt 0 ]]; then
        echo "" >> "$SPEC_FILE"
        echo "## §B Bugs" >> "$SPEC_FILE"
        echo "| id | symptom | fix | status |" >> "$SPEC_FILE"
        echo "|----|---------|-----|--------|" >> "$SPEC_FILE"
        for bug in "${NEW_BUGS[@]}"; do
            echo "$bug" >> "$SPEC_FILE"
        done
    fi

    echo "✅ Spec updated"
fi

echo ""
echo "📊 Summary:"
echo "   New invariants: ${#NEW_INVARIANTS[@]}"
echo "   New bugs: ${#NEW_BUGS[@]}"
echo ""
echo "💡 Next: Run drift-scope to verify invariants are enforced"
