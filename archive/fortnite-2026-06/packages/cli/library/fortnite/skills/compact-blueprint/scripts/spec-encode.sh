#!/usr/bin/env bash
# spec-encode.sh — Convert prose spec to caveman-encoded compact format
# Usage: ./spec-encode.sh --input <file> --output <file>

set -euo pipefail

INPUT_FILE=""
OUTPUT_FILE=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --input) INPUT_FILE="$2"; shift 2 ;;
        --output) OUTPUT_FILE="$2"; shift 2 ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

if [[ -z "$INPUT_FILE" || -z "$OUTPUT_FILE" ]]; then
    echo "Usage: $0 --input <file> --output <file>"
    exit 1
fi

if [[ ! -f "$INPUT_FILE" ]]; then
    echo "❌ Input file not found: $INPUT_FILE"
    exit 1
fi

echo "🪨 Spec Encoder — Prose → Caveman Format"
echo "   Input: $INPUT_FILE"
echo "   Output: $OUTPUT_FILE"
echo ""

# Create compact spec template
cat > "$OUTPUT_FILE" << 'EOF'
## §G Goal
TODO: one-line goal, scope boundaries

## §C Constraints
| id | rule | priority |
|----|------|----------|
| C1 | TODO | critical |

## §I Interfaces
| id | signature | returns |
|----|-----------|---------|
| I1 | TODO | TODO |

## §V Invariants
| id | rule | evidence |
|----|------|----------|
| V1 | TODO | TODO |

## §T Tasks
| id | desc | done | files |
|----|------|------|-------|
| T1 | TODO | ☐ | TODO |

## §B Bugs
| id | symptom | fix | status |
|----|---------|-----|--------|
| B1 | TODO | TODO | open |
EOF

echo "✅ Compact spec template created: $OUTPUT_FILE"
echo ""
echo "📝 Next: Fill in sections from input spec"
echo "   §G: Extract goal from title/description"
echo "   §C: Extract constraints from requirements"
echo "   §I: Extract interfaces from API docs"
echo "   §V: Extract invariants from test cases"
echo "   §T: Extract tasks from implementation plan"
echo "   §B: Extract bugs from issue tracker"
