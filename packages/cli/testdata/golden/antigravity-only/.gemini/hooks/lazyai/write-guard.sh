#!/usr/bin/env bash
# LazyAI write-guard hook for Antigravity.
# Fires on PreToolUse for write_to_file, replace_file_content,
# multi_replace_file_content. Hard-denies writes that lack an explicit
# file argument or target protected paths.

if ! command -v python3 >/dev/null 2>&1; then
  echo '{"decision": "deny", "reason": "Python3 is required but not found"}'
  exit 2
fi
input="$(cat)"

JSON_INPUT="$input" python3 - <<'PY'
import json, os, sys

try:
    data = json.loads(os.environ.get("JSON_INPUT", ""))
except Exception:
    print(json.dumps({"decision": "deny", "reason": "write-guard: could not parse JSON input"}))
    sys.exit(2)

WRITE_TOOLS = {"write_to_file", "replace_file_content", "multi_replace_file_content"}
tool_name = data.get("toolName") or data.get("tool_name") or ""
if tool_name not in WRITE_TOOLS:
    print(json.dumps({"decision": "allow"}))
    sys.exit(0)

tool_input = data.get("toolInput") or data.get("tool_input") or {}
if isinstance(tool_input, str):
    try:
        tool_input = json.loads(tool_input)
    except Exception:
        tool_input = {}

path = None
if isinstance(tool_input, dict):
    path = tool_input.get("path") or tool_input.get("file_path") or tool_input.get("file")

if not path:
    print(json.dumps({"decision": "deny", "reason": "write-guard: write tool called without a resolvable file path"}))
    sys.exit(2)

# Block writes to protected system/config paths.
protected_prefixes = ["/etc/", "/usr/", "/bin/", "/sbin/", "/boot/"]
if any(path.startswith(p) for p in protected_prefixes):
    print(json.dumps({"decision": "deny", "reason": f"write-guard: write to protected path blocked: {path}"}))
    sys.exit(2)

print(json.dumps({"decision": "allow"}))
sys.exit(0)
PY
