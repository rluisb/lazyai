#!/usr/bin/env bash
# colima-resources.sh — Show current resource allocation
# Usage: ./colima-resources.sh [--profile PROFILE]
set -euo pipefail

PROFILE="${1:-default}"

echo "=== Colima Resources ==="
echo "Profile: $PROFILE"
echo ""

echo "--- Instance Details ---"
colima list 2>&1
echo ""

echo "--- VM CPU ---"
colima ssh --profile "$PROFILE" -- nproc 2>&1 || echo "Failed to get CPU count"
echo ""

echo "--- VM Memory ---"
colima ssh --profile "$PROFILE" -- free -h 2>&1 || echo "Failed to get memory info"
echo ""

echo "--- VM Disk ---"
colima ssh --profile "$PROFILE" -- df -h / 2>&1 || echo "Failed to get disk info"
echo ""

echo "--- Docker Disk Usage ---"
docker system df 2>&1 || echo "Failed to get Docker disk usage"
echo ""

echo "=== Resources Complete ==="
