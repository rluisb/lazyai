#!/usr/bin/env bash
# hotctl-context.sh — Show current AWS context (profile, region, EKS, ECR)
# Usage: ./hotctl-context.sh [--profile PROFILE]
set -euo pipefail

PROFILE="${1:-default}"
REGION="${2:-us-east-1}"

echo "=== AWS Context ==="
echo "Profile: $PROFILE"
echo "Region:  $REGION"
echo ""

echo "--- Current EKS Context ---"
hotctl eks show --profile "$PROFILE" --region "$REGION" --silent 2>&1 || echo "No EKS context"
echo ""

echo "--- ECR Login Status ---"
hotctl ecr login --profile "$PROFILE" --region "$REGION" --silent 2>&1 && echo "ECR: Authenticated" || echo "ECR: Not authenticated"
echo ""

echo "--- Available EKS Clusters ---"
hotctl eks list --profile "$PROFILE" --region "$REGION" --silent 2>&1 | head -20 || echo "Failed to list clusters"
echo ""

echo "=== Context Complete ==="
