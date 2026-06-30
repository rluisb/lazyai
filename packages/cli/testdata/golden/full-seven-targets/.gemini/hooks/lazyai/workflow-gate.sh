#!/usr/bin/env bash
# LazyAI workflow-gate hook for Antigravity.
# PreInvocation slot: reserved for workflow step injection.
# Exits 0 (allow) unconditionally; replace with step-injection logic
# when workflow orchestration requires pre-model context injection.

if ! command -v python3 >/dev/null 2>&1; then
  echo '{"decision": "allow"}'
  exit 0
fi
input="$(cat)"

JSON_INPUT="$input" python3 - <<'PY'
import json, os, sys

try:
    data = json.loads(os.environ.get("JSON_INPUT", ""))
except Exception:
    # On parse failure, allow — PreInvocation failures must not block normal operation.
    print(json.dumps({"decision": "allow"}))
    sys.exit(0)

# Reserved slot: emit allow unconditionally.
# Future workflow step injection: inspect data["workflowStep"] or
# data["agentName"] here and return {"decision": "allow", "inject": [...]}
# to prepend context before the model call.
print(json.dumps({"decision": "allow"}))
sys.exit(0)
PY
