#!/usr/bin/env bash
# LazyAI Codex PreToolUse hook: blocks obviously destructive shell commands.
# Codex delivers the hook payload as a single JSON object on stdin (tool_name,
# tool_input.command) and honors a hookSpecificOutput.permissionDecision="deny"
# response on stdout. This is a usability guardrail, not a security boundary
# (Codex can reach equivalent work through other tool paths), so it fails open.
# See https://developers.openai.com/codex/hooks.

if ! command -v python3 >/dev/null 2>&1; then
  exit 0
fi
input="$(cat)"

JSON_INPUT="$input" python3 - <<'PY'
import json
import os
import sys


def deny(reason):
    print(json.dumps({
        "hookSpecificOutput": {
            "hookEventName": "PreToolUse",
            "permissionDecision": "deny",
            "permissionDecisionReason": reason,
        }
    }))
    sys.exit(0)


try:
    data = json.loads(os.environ.get("JSON_INPUT", ""))
except Exception:
    sys.exit(0)

if not isinstance(data, dict) or data.get("tool_name") not in ("Bash", "bash", "shell"):
    sys.exit(0)

tool_input = data.get("tool_input")
command = ""
if isinstance(tool_input, dict):
    command = tool_input.get("command", "")
elif isinstance(tool_input, str):
    command = tool_input
if not isinstance(command, str):
    sys.exit(0)

command = command.strip()
denied_prefixes = [
    "rm -rf /",
    "rm -rf /*",
    "mkfs",
    "dd if=/dev/zero of=",
    "dd if=/dev/zero of=/dev/",
    "> /dev/sd",
    "shutdown",
    "poweroff",
    "reboot",
    "halt",
]
if any(command == p or command.startswith(p + " ") for p in denied_prefixes):
    deny("Destructive shell command blocked by LazyAI policy")

sys.exit(0)
PY
