#!/usr/bin/env bash
# agent-msg.sh — Root compatibility wrapper for inter-agent message bus
# Delegates to battle-bus implementation with correct SESSION_DB resolution

set -euo pipefail

# Resolve to the root session-db.sh
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SESSION_DB="$SCRIPT_DIR/session-db.sh"
BATTLE_BUS_AGENT_MSG="$SCRIPT_DIR/../skills/battle-bus/scripts/agent-msg.sh"

# Export environment for battle-bus script
export OPENCODE_WORKSPACE="${OPENCODE_WORKSPACE:-.}"
export SESSION_ID="${SESSION_ID:-}"

# Delegate to battle-bus implementation
if [[ -f "$BATTLE_BUS_AGENT_MSG" ]]; then
    exec "$BATTLE_BUS_AGENT_MSG" "$@"
else
    echo "❌ Battle-bus agent-msg.sh not found at $BATTLE_BUS_AGENT_MSG" >&2
    exit 1
fi
