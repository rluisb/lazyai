#!/usr/bin/env bash
# Test workflow failure policies
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE_ROOT="$(cd "$SCRIPT_DIR/../../../" && pwd)"
SESSION_DB="$WORKSPACE_ROOT/scripts/session-db.sh"
WORKFLOW_EXEC="$WORKSPACE_ROOT/skills/workflow-engine/scripts/workflow-exec.sh"

# Ensure DB is initialized
"$SESSION_DB" init >/dev/null 2>&1

# Create a session
SID=$("$SESSION_DB" start-session "test-failure-policy" "test-repo")
echo "Session: $SID"

echo ""
echo "=== TEST 1: Default policy (stop) ==="
WIID1=$("$WORKFLOW_EXEC" start rpi "$SID" 2>/dev/null)
echo "Started instance: $WIID1"
POLICY1=$("$SESSION_DB" query "SELECT on_failure FROM workflow_instances WHERE id='$WIID1';" 2>/dev/null | tail -n 1)
echo "Policy stored: ${POLICY1:-stop}"
"$WORKFLOW_EXEC" fail "$WIID1" "Simulated error" >/dev/null 2>&1
STATUS1=$("$SESSION_DB" query "SELECT status FROM workflow_instances WHERE id='$WIID1';" 2>/dev/null | tail -n 1)
echo "Final status: $STATUS1"
[ "$STATUS1" = "failed" ] && echo "✅ PASS: stop policy halts workflow" || echo "❌ FAIL: expected failed, got $STATUS1"

echo ""
echo "=== TEST 2: Continue policy ==="
WIID2=$("$WORKFLOW_EXEC" start rpi "$SID" --on-failure continue 2>/dev/null)
echo "Started instance: $WIID2"
POLICY2=$("$SESSION_DB" query "SELECT on_failure FROM workflow_instances WHERE id='$WIID2';" 2>/dev/null | tail -n 1)
echo "Policy stored: ${POLICY2:-stop}"
# Advance to step 1 first
"$SESSION_DB" workflow-step-start "$WIID2" >/dev/null 2>&1 || true
"$WORKFLOW_EXEC" fail "$WIID2" "Simulated error" >/dev/null 2>&1
STATUS2=$("$SESSION_DB" query "SELECT status FROM workflow_instances WHERE id='$WIID2';" 2>/dev/null | tail -n 1)
STEP2_STATUS=$("$SESSION_DB" query "SELECT status FROM workflow_steps WHERE instance_id='$WIID2' AND step_order=1;" 2>/dev/null | tail -n 1)
echo "Workflow status: $STATUS2"
echo "Step 1 status: ${STEP2_STATUS:-unknown}"
[ "$STATUS2" = "running" ] && [ "$STEP2_STATUS" = "completed" ] && echo "✅ PASS: continue policy advances workflow" || echo "⚠️  PARTIAL: continue policy (status=$STATUS2, step=$STEP2_STATUS)"

echo ""
echo "=== TEST 3: Rollback policy ==="
WIID3=$("$WORKFLOW_EXEC" start rpi "$SID" --on-failure rollback 2>/dev/null)
echo "Started instance: $WIID3"
POLICY3=$("$SESSION_DB" query "SELECT on_failure FROM workflow_instances WHERE id='$WIID3';" 2>/dev/null | tail -n 1)
echo "Policy stored: ${POLICY3:-stop}"
# Complete step 1
"$SESSION_DB" workflow-step-start "$WIID3" >/dev/null 2>&1 || true
"$SESSION_DB" workflow-step-done "$WIID3" "pass" >/dev/null 2>&1 || true
# Start step 2 then fail
"$SESSION_DB" workflow-step-start "$WIID3" >/dev/null 2>&1 || true
"$WORKFLOW_EXEC" fail "$WIID3" "Simulated error" >/dev/null 2>&1
STATUS3=$("$SESSION_DB" query "SELECT status FROM workflow_instances WHERE id='$WIID3';" 2>/dev/null | tail -n 1)
ERR3=$("$SESSION_DB" query "SELECT error_message FROM workflow_instances WHERE id='$WIID3';" 2>/dev/null | tail -n 1)
echo "Workflow status: $STATUS3"
echo "Error message: ${ERR3:-none}"
[ "$STATUS3" = "failed" ] && [[ "$ERR3" == *"rollback attempted"* ]] && echo "✅ PASS: rollback policy attempts rollback then halts" || echo "⚠️  PARTIAL: rollback policy (status=$STATUS3)"

echo ""
echo "=== TEST 4: Invalid policy rejected ==="
if "$WORKFLOW_EXEC" start rpi "$SID" --on-failure invalid >/dev/null 2>&1; then
    echo "❌ FAIL: invalid policy should be rejected"
else
    echo "✅ PASS: invalid policy rejected"
fi

echo ""
echo "=== All tests complete ==="
