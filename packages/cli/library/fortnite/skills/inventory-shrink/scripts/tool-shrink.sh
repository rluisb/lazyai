#!/usr/bin/env bash
# tool-shrink.sh — Compress MCP tool descriptions
# Usage: ./tool-shrink.sh --input <json> --output <json> [--level lite|full|ultra]

set -euo pipefail

INPUT_FILE=""
OUTPUT_FILE=""
LEVEL="full"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --input) INPUT_FILE="$2"; shift 2 ;;
        --output) OUTPUT_FILE="$2"; shift 2 ;;
        --level) LEVEL="$2"; shift 2 ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

if [[ -z "$INPUT_FILE" || -z "$OUTPUT_FILE" ]]; then
    echo "Usage: $0 --input <json> --output <json> [--level lite|full|ultra]"
    exit 1
fi

if [[ ! -f "$INPUT_FILE" ]]; then
    echo "❌ Input file not found: $INPUT_FILE"
    exit 1
fi

echo "🪨 Inventory Shrink — Tool Description Compression"
echo "   Input: $INPUT_FILE"
echo "   Output: $OUTPUT_FILE"
echo "   Level: $LEVEL"
echo ""

# Check if jq is available
if ! command -v jq &>/dev/null; then
    echo "❌ jq is required but not installed."
    echo "   Install: brew install jq"
    exit 1
fi

# Count original tokens (rough estimate: words * 1.3)
ORIGINAL_TOKENS=$(jq -r '.tools[]?.description // ""' "$INPUT_FILE" 2>/dev/null | wc -w | awk '{print int($1 * 1.3)}')

# Compress descriptions based on level
case "$LEVEL" in
    lite)
        # Drop filler words: "the", "a", "an", "is", "are", "you", "your", etc.
        jq '
          .tools = [.tools[] |
            .description = (.description
              | gsub("^(The|This|That|These|Those) "; "")
              | gsub(" (the|a|an|is|are|was|were|you|your|youre|its|its) "; " ")
              | gsub("  +"; " ")
              | gsub("^ "; "")
              | gsub(" $"; ""))
          ]' "$INPUT_FILE" > "$OUTPUT_FILE"
        ;;
    full)
        # Fragments, no articles, direct statements
        jq '
          .tools = [.tools[] |
            .description = (.description
              | gsub("^(The|This|That|These|Those|A|An) "; "")
              | gsub(" (the|a|an|is|are|was|were|you|your|youre|its|its|will|would|should|can|could|may|might|must) "; " ")
              | gsub(" to "; " ")
              | gsub(" for "; " ")
              | gsub(" in order to "; " ")
              | gsub("  +"; " ")
              | gsub("^ "; "")
              | gsub(" $"; ""))
          ]' "$INPUT_FILE" > "$OUTPUT_FILE"
        ;;
    ultra)
        # Telegraphic, symbols only
        jq '
          .tools = [.tools[] |
            .description = (.description
              | gsub("^(The|This|That|These|Those|A|An) "; "")
              | gsub(" (the|a|an|is|are|was|were|you|your|youre|its|its|will|would|should|can|could|may|might|must|to|for|in|on|at|by|with|from|into|through|during|before|after|above|below|between|under|again|further|then|once|here|there|when|where|why|how|all|both|each|few|more|most|other|some|such|no|nor|not|only|own|same|so|than|too|very|just|because|as|until|while|of|and|but|or|if) "; " ")
              | gsub("  +"; " ")
              | gsub("^ "; "")
              | gsub(" $"; ""))
          ]' "$INPUT_FILE" > "$OUTPUT_FILE"
        ;;
    *)
        echo "❌ Unknown level: $LEVEL. Use lite, full, or ultra."
        exit 1
        ;;
esac

# Count compressed tokens
COMPRESSED_TOKENS=$(jq -r '.tools[]?.description // ""' "$OUTPUT_FILE" 2>/dev/null | wc -w | awk '{print int($1 * 1.3)}')

# Calculate savings
if [[ $ORIGINAL_TOKENS -gt 0 ]]; then
    SAVINGS=$(( (ORIGINAL_TOKENS - COMPRESSED_TOKENS) * 100 / ORIGINAL_TOKENS ))
else
    SAVINGS=0
fi

echo "✅ Compressed tool descriptions saved: $OUTPUT_FILE"
echo ""
echo "📊 Token Savings:"
echo "   Original: ~${ORIGINAL_TOKENS} tokens"
echo "   Compressed: ~${COMPRESSED_TOKENS} tokens"
echo "   Saved: ~${SAVINGS}%"
echo ""
echo "💡 Next: Replace original tool descriptions in MCP config"
