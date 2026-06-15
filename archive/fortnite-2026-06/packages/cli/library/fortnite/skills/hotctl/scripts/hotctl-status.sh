#!/usr/bin/env bash
# hotctl-status.sh — Quick health check of all AWS services via hotctl
# Usage: ./hotctl-status.sh [--profile PROFILE] [--region REGION]
set -euo pipefail

PROFILE="${1:-default}"
REGION="${2:-us-east-1}"

echo "=== Hotctl AWS Status ==="
echo "Profile: $PROFILE | Region: $REGION"
echo ""

echo "--- SSO ---"
hotctl sso login --silent 2>&1 && echo "SSO: OK" || echo "SSO: NEEDS LOGIN"
echo ""

echo "--- EKS Clusters ---"
hotctl eks list --profile "$PROFILE" --region "$REGION" --silent 2>&1 || echo "EKS: Failed to list"
echo ""

echo "--- RDS Clusters ---"
hotctl rds list --profile "$PROFILE" --region "$REGION" --silent 2>&1 || echo "RDS: Failed to list"
echo ""

echo "--- ECR Login ---"
hotctl ecr login --profile "$PROFILE" --region "$REGION" --silent 2>&1 && echo "ECR: OK" || echo "ECR: Failed"
echo ""

echo "--- Secrets ---"
hotctl secret list --profile "$PROFILE" --region "$REGION" --silent 2>&1 || echo "Secrets: Failed to list"
echo ""

echo "=== Status Complete ==="
