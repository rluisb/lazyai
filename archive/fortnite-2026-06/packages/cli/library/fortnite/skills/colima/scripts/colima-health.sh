#!/usr/bin/env bash
# colima-health.sh — Quick health check of Colima + Docker
# Usage: ./colima-health.sh [--profile PROFILE]
set -euo pipefail

PROFILE="${1:-default}"

echo "=== Colima Health Check ==="
echo "Profile: $PROFILE"
echo ""

echo "--- Colima Status ---"
colima status --profile "$PROFILE" 2>&1
STATUS_EXIT=$?
echo ""

if [[ $STATUS_EXIT -ne 0 ]]; then
    echo "❌ Colima is NOT running"
    echo "Run: colima start --profile $PROFILE"
    exit 1
fi

echo "--- Colima Instances ---"
colima list 2>&1
echo ""

echo "--- Docker Version ---"
docker --version 2>&1 || echo "❌ Docker not available"
echo ""

echo "--- Docker Info ---"
docker info --format 'Server Version: {{.ServerVersion}}' 2>&1 || echo "❌ Docker daemon not responding"
echo ""

echo "--- Docker Socket ---"
SOCKET_PATH="$HOME/.colima/$PROFILE/docker.sock"
if [[ -S "$SOCKET_PATH" ]]; then
    echo "✅ Docker socket: $SOCKET_PATH"
else
    echo "❌ Docker socket not found: $SOCKET_PATH"
fi
echo ""

echo "=== Health Check Complete ==="
