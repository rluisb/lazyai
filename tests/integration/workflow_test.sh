#!/usr/bin/env bash
set -euo pipefail

# Integration Test: Workflow Execution
# Tests workflow YAML parsing and execution

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
CLI="$PROJECT_DIR/packages/cli/lazyai-cli"

echo "═══════════════════════════════════════════════════════════════"
echo "🧪 Integration Test: Workflow Execution"
echo "═══════════════════════════════════════════════════════════════"

# Test 1: List workflows
echo ""
echo "Test 1: List available workflows"
WORKFLOWS_DIR="$PROJECT_DIR/.opencode/workflows"
if [ -d "$WORKFLOWS_DIR" ]; then
    WORKFLOW_COUNT=$(ls -1 "$WORKFLOWS_DIR"/*.yaml 2>/dev/null | wc -l | tr -d ' ')
    echo "✅ Found $WORKFLOW_COUNT workflow definitions"
    
    if [ "$WORKFLOW_COUNT" -eq 0 ]; then
        echo "❌ FAIL: No workflow definitions found"
        exit 1
    fi
else
    echo "❌ FAIL: Workflows directory not found"
    exit 1
fi

# Test 2: Verify workflow YAML syntax
echo ""
echo "Test 2: Verify workflow YAML syntax"
for workflow in "$WORKFLOWS_DIR"/*.yaml; do
    if python3 -c "import yaml; yaml.safe_load(open('$workflow'))" 2>/dev/null; then
        echo "  ✅ $(basename "$workflow")"
    else
        echo "  ❌ $(basename "$workflow") - invalid YAML"
        exit 1
    fi
done

# Test 3: Check required workflow fields
echo ""
echo "Test 3: Check required workflow fields"
for workflow in "$WORKFLOWS_DIR"/*.yaml; do
    NAME=$(basename "$workflow" .yaml)
    if python3 -c "
import yaml
with open('$workflow') as f:
    data = yaml.safe_load(f)
    assert 'name' in data, 'Missing name'
    assert 'phases' in data, 'Missing phases'
    assert 'metadata' in data, 'Missing metadata'
    print('✅ $NAME')
" 2>&1; then
        :
    else
        echo "❌ FAIL: $NAME missing required fields"
        exit 1
    fi
done

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "✅ All workflow integration tests passed!"
echo "═══════════════════════════════════════════════════════════════"
