#!/usr/bin/env bash
# create-agent.sh — Scaffold a new agent with proper structure
# Usage: ./create-agent.sh <fortnite-name> <primary-model> <primary-skill>
#
# Example:
#   ./create-agent.sh scout-agent ollama-cloud/deepseek-v4-flash storm-scout

set -euo pipefail

AGENT_NAME="${1:-}"
MODEL="${2:-}"
SKILL="${3:-}"

if [[ -z "$AGENT_NAME" || -z "$MODEL" || -z "$SKILL" ]]; then
    echo "Usage: $0 <fortnite-name> <primary-model> <primary-skill>"
    echo ""
    echo "Example:"
    echo "  $0 scout-agent ollama-cloud/deepseek-v4-flash storm-scout"
    echo ""
    echo "Available models (from AGENTS.md):"
    echo "  ollama-cloud/deepseek-v4-pro"
    echo "  ollama-cloud/deepseek-v4-flash"
    echo "  ollama-cloud/kimi-k2.6:cloud"
    echo "  ollama-cloud/glm-5.1"
    echo "  ollama-cloud/nemotron-3-super"
    echo "  ollama-cloud/gemma4"
    echo "  openai/gpt-5.5"
    exit 1
fi

AGENT_FILE="agents/$AGENT_NAME.md"

# Check if agent already exists
if [[ -f "$AGENT_FILE" ]]; then
    echo "❌ Agent already exists: $AGENT_FILE"
    exit 1
fi

echo "🏗️  Creating agent: $AGENT_NAME"
echo "   Model: $MODEL"
echo "   Skill: $SKILL"
echo ""

# Create agent file from template
cat > "$AGENT_FILE" << EOF
---
name: $AGENT_NAME
model: $MODEL
skill: $SKILL
think: true
permission.task: allow
---

# $AGENT_NAME

## Role
<TODO: One sentence — what this agent does in the squad>

## Parameter Contract

| Parameter | Values | Default | Effect |
|-----------|--------|---------|--------|
| MODE | \`<value1>\` / \`<value2>\` | \`<default>\` | What mode controls |
| THINK | \`true\` / \`false\` / \`xhigh\` | \`true\` | Reasoning depth |
| TOKEN_BUDGET | number | \`40K\` | Max context tokens |

## Fallback Chain

\`$MODEL\` → \`<fallback-1>\` → \`<fallback-2>\` → escalate to loop-driver

## Capabilities

**Can do:**
- TODO: Capability 1
- TODO: Capability 2
- Dispatch to other agents (per dispatch matrix)

**Cannot do:**
- TODO: Restriction 1
- TODO: Restriction 2

## Delegation Rules

- Can dispatch to: \`<agent-list>\`
- Requires human approval for: \`<actions>\`
- Never delegates to: \`<agents>\` without loop-driver mediation

## Integration

- **Primary skill**: \`$SKILL\` — loaded automatically
- **Secondary skills**: \`<skill-list>\` — loaded as needed
- **CLI tools**: \`<tool-list>\` — available for this agent
EOF

echo "✅ Agent created: $AGENT_FILE"
echo ""
echo "Next steps:"
echo "1. Edit $AGENT_FILE — fill in Role, Parameter Contract, Capabilities"
echo "2. Add to AGENTS.md 'The Squad' table"
echo "3. Add to AGENTS.md 'Parameter Contracts' table"
echo "4. Add to AGENTS.md 'Fallback Chains' table"
echo "5. Update dispatch matrix if this agent can dispatch to others"
echo "6. Update skills/_INDEX.md Agent → Skill Mapping"
