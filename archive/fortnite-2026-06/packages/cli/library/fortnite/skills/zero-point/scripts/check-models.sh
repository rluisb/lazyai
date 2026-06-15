#!/usr/bin/env bash
# check-models.sh — Validate that all agent primary models are available via opencode CLI
# Usage: ./check-models.sh

set -euo pipefail

# Map of agent → expected primary model
declare -A EXPECTED_MODELS=(
    ["loop-driver"]="ollama-cloud/minimax-m2.7"
    ["loot-hawk"]="ollama-cloud/deepseek-v4-flash"
    ["turbo-crank"]="ollama-cloud/deepseek-v4-pro"
    ["wall-builder"]="ollama-cloud/kimi-k2.6"
    ["shield-audit"]="openai/gpt-5.5"
    ["rift-deploy"]="ollama-cloud/nemotron-3-super"
    ["respawn-crew"]="ollama-cloud/kimi-k2.6"
)

AGENTS_DIR="${OPENCODE_CONFIG_DIR:-$HOME/.config/opencode}/agents"

echo "=== Model Availability Check ==="
echo ""

all_ok=true

check_model_available() {
    local agent="$1"
    local expected="$2"

    local provider="${expected%%/*}"
    local model="${expected#*/}"

    if opencode models "$provider" 2>/dev/null | grep -q "$model"; then
        echo "  ✅ $agent → $expected"
    else
        echo "  ❌ $agent → $expected (NOT FOUND in $provider)"
        all_ok=false
    fi
}

echo "Checking primary models..."
for agent in "${!EXPECTED_MODELS[@]}"; do
    check_model_available "$agent" "${EXPECTED_MODELS[$agent]}"
done

echo ""
if [ "$all_ok" = true ]; then
    echo "✅ All models available"
    exit 0
else
    echo "❌ Some models are unavailable"
    exit 1
fi
