#!/usr/bin/env bash
# mcp-health.sh — Verify all MCP servers are responsive before agent dispatch
# Usage: ./mcp-health.sh

set -euo pipefail

CONFIG_FILE="${OPENCODE_CONFIG_DIR:-$HOME/.config/opencode}/dcp.jsonc"

echo "=== MCP Health Check ==="

if [ ! -f "$CONFIG_FILE" ]; then
    echo "No config file found at $CONFIG_FILE"
    exit 1
fi

# Parse enabled MCP servers from config
SERVERS=$(python3 -c "
import json, sys
with open('$CONFIG_FILE') as f:
    raw = f.read()
data = json.loads(raw)
servers = data.get('mcp-servers', data.get('mcp_servers', {}))
for name, cfg in servers.items():
    if cfg.get('enabled', True):
        cmd = cfg.get('command', '')
        print(f'{name}|{cmd}')
" 2>/dev/null)

if [ -z "$SERVERS" ]; then
    echo "No MCP servers configured or could not parse config"
    exit 0
fi

all_ok=true

while IFS='|' read -r name command; do
    echo -n "  $name: "
    if [ -z "$command" ]; then
        echo "✅ (remote/URL-based)"
    elif command -v "${command%% *}" &>/dev/null; then
        echo "✅ (${command%% *})"
    else
        echo "❌ binary not found: ${command%% *}"
        all_ok=false
    fi
done <<< "$SERVERS"

echo ""
if [ "$all_ok" = true ]; then
    echo "✅ All MCP servers available"
    exit 0
else
    echo "❌ Some MCP servers unavailable — agents may fail on tool calls"
    exit 1
fi
