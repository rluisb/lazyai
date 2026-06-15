#!/usr/bin/env bash
# dev-status.sh — Quick service status check
# Usage: ./dev-status.sh [--json]
set -euo pipefail

JSON_FLAG=""
if [[ "${1:-}" == "--json" ]]; then
    JSON_FLAG="--json"
fi

echo "=== Dev CLI Service Status ==="
echo ""

dev list $JSON_FLAG 2>&1

echo ""
echo "=== Docker Status ==="
docker info --format '{{.ServerVersion}}' 2>/dev/null && echo "Docker: OK" || echo "Docker: NOT RUNNING"

echo ""
echo "=== Colima Status ==="
colima status 2>&1 | head -5 || echo "Colima: NOT RUNNING"

echo ""
echo "=== Status Complete ==="
