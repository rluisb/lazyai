#!/usr/bin/env bash
# End-to-end user-interaction harness for lazyai-cli V2 lifecycle.
# Drives real CLI commands against throwaway fake projects with an isolated HOME.
set -u
BIN="${BIN:?set BIN to the built lazyai-cli}"
ROOT="$(mktemp -d)"
export HOME_ISO="$ROOT/home"
mkdir -p "$HOME_ISO"
PASS=0
FAIL=0
run() { HOME="$HOME_ISO" "$BIN" "$@"; }
ok() { printf '  PASS  %s\n' "$1"; PASS=$((PASS + 1)); }
bad() { printf '  FAIL  %s\n' "$1"; FAIL=$((FAIL + 1)); }
check_file() { [ -e "$1" ] && ok "exists: ${2:-$1}" || bad "missing: ${2:-$1}"; }
check_nofile() { [ ! -e "$1" ] && ok "absent: ${2:-$1}" || bad "present(unexpected): ${2:-$1}"; }
expect_ok() {
  local n="$1"
  shift
  if "$@" >/tmp/lazyai-e2e.out 2>&1; then
    ok "$n"
  else
    bad "$n (exit $?)"
    sed 's/^/      /' /tmp/lazyai-e2e.out | tail -3
  fi
}
expect_fail() {
  local n="$1"
  shift
  if "$@" >/tmp/lazyai-e2e.out 2>&1; then
    bad "$n (expected non-zero)"
  else
    ok "$n"
  fi
}
grep_ok() { grep -q "$1" "$2" && ok "$3" || bad "$3 (pattern '$1' not in $2)"; }

echo "== Scenario 1: init project scope =="
P="$ROOT/p1"
mkdir -p "$P"
if (cd "$P" && run init --no-interactive --scope project --tools opencode --enable-servers filesystem --name demo) >/tmp/lazyai-e2e.out 2>&1; then
  ok "init exits 0"
else
  bad "init failed"
  tail -5 /tmp/lazyai-e2e.out
fi
check_file "$P/.ai/lazyai.json" "canonical manifest"
check_file "$P/.ai/mcp.json" "canonical mcp"
check_file "$P/.mcp.json" "native mcp"
check_file "$P/AGENTS.md" "AGENTS.md"
check_file "$P/opencode.json" "opencode native"
check_file "$P/.ai-setup.db" "setup db"
python3 - "$P/.ai/mcp.json" <<'PY' && ok "allowlist: filesystem enabled, others disabled" || bad "allowlist incorrect"
import json, sys
with open(sys.argv[1], encoding="utf-8") as f:
    d = json.load(f)
s = d.get("servers", d.get("mcpServers", {}))
fs = s.get("filesystem", {}).get("enabled") is True
extra = [k for k, v in s.items() if k != "filesystem" and (v.get("enabled") is True)]
sys.exit(0 if (fs and not extra) else 1)
PY

echo "== Scenario 2: init rejects unknown tool =="
expect_fail "init --tools bogus rejected" bash -c "cd '$ROOT' && HOME='$HOME_ISO' '$BIN' init --no-interactive --scope project --tools bogus --name x"

echo "== Scenario 3: compile dry-run writes nothing =="
(cd "$P" && run compile --dry-run) >/tmp/lazyai-e2e.out 2>&1
check_nofile "$P/.ai/lock.json" "lock absent after dry-run"

echo "== Scenario 4: compile writes lock v1.0 =="
expect_ok "compile exits 0" bash -c "cd '$P' && HOME='$HOME_ISO' '$BIN' compile"
check_file "$P/.ai/lock.json" "lock.json"
grep_ok '"version": "1.0"' "$P/.ai/lock.json" "lock frozen at 1.0"
python3 - "$P/.ai/lock.json" <<'PY' && ok "lock records lazyaiVersion and compiledAt" || bad "lock missing version metadata"
import json, sys
from datetime import datetime
with open(sys.argv[1], encoding="utf-8") as f:
    d = json.load(f)
if not d.get("lazyaiVersion") or not d.get("compiledAt"):
    sys.exit(1)
datetime.fromisoformat(d["compiledAt"].replace("Z", "+00:00"))
PY

echo "== Scenario 5: compile is stable except timestamp =="
cp "$P/.ai/lock.json" /tmp/lazyai-e2e-lock1.json
(cd "$P" && run compile) >/tmp/lazyai-e2e.out 2>&1
python3 - /tmp/lazyai-e2e-lock1.json "$P/.ai/lock.json" <<'PY' && ok "lock content stable across recompile except compiledAt" || bad "lock content changed beyond compiledAt"
import json, sys
with open(sys.argv[1], encoding="utf-8") as f:
    a = json.load(f)
with open(sys.argv[2], encoding="utf-8") as f:
    b = json.load(f)
a.pop("compiledAt", None)
b.pop("compiledAt", None)
sys.exit(0 if a == b else 1)
PY

echo "== Scenario 6: validate --all clean =="
expect_ok "validate --all passes on clean scaffold" bash -c "cd '$P' && HOME='$HOME_ISO' '$BIN' validate --all"

echo "== Scenario 7: validate --all enforces schema freeze =="
python3 - "$P/.ai/lazyai.json" <<'PY'
import json, sys
f = sys.argv[1]
with open(f, encoding="utf-8") as fh:
    d = json.load(fh)
d["version"] = "2.0"
with open(f, "w", encoding="utf-8") as fh:
    json.dump(d, fh, indent=2)
PY
expect_fail "validate --all rejects unsupported manifest version" bash -c "cd '$P' && HOME='$HOME_ISO' '$BIN' validate --all"
expect_fail "compile rejects unsupported manifest version" bash -c "cd '$P' && HOME='$HOME_ISO' '$BIN' compile"
python3 - "$P/.ai/lazyai.json" <<'PY'
import json, sys
f = sys.argv[1]
with open(f, encoding="utf-8") as fh:
    d = json.load(fh)
d["version"] = "1.0"
with open(f, "w", encoding="utf-8") as fh:
    json.dump(d, fh, indent=2)
PY

echo "== Scenario 8: validate evals =="
expect_fail "validate evals fails when dir missing" bash -c "cd '$P' && HOME='$HOME_ISO' '$BIN' validate evals"
mkdir -p "$P/.ai/evals/cases" "$P/.ai/evals/holdouts" "$P/.ai/evals/rubrics"
cat >"$P/.ai/evals/cases/c1.yaml" <<'Y'
id: case-1
title: Demo case
input:
  prompt: hello
expected:
  answer: hi
Y
printf '# Rubric\nScore clarity.\n' >"$P/.ai/evals/rubrics/r1.md"
expect_ok "validate evals passes on valid case" bash -c "cd '$P' && HOME='$HOME_ISO' '$BIN' validate evals"
cat >"$P/.ai/evals/cases/bad.yaml" <<'Y'
id: case-2
title: missing expected
input:
  prompt: hello
Y
expect_fail "validate evals fails on invalid case" bash -c "cd '$P' && HOME='$HOME_ISO' '$BIN' validate evals"
rm "$P/.ai/evals/cases/bad.yaml"

echo "== Scenario 9: build-plugin all targets =="
for t in claude copilot-cli omp pi; do
  OUT="$ROOT/bundle-$t"
  expect_ok "build-plugin $t" bash -c "cd '$ROOT' && HOME='$HOME_ISO' '$BIN' build-plugin --target $t --out '$OUT'"
  case $t in
  claude)
    check_file "$OUT/.claude-plugin/plugin.json" "claude plugin.json"
    check_file "$OUT/agents/guide.md" "claude agent"
    ;;
  copilot-cli)
    check_file "$OUT/plugin.json" "copilot plugin.json"
    check_file "$OUT/skills/implement/SKILL.md" "copilot skill"
    check_file "$OUT/hooks.json" "copilot hooks.json"
    check_file "$OUT/.mcp.json" "copilot .mcp.json"
    ;;
  omp)
    check_file "$OUT/skills/implement/SKILL.md" "omp skill"
    check_file "$OUT/mcp.json" "omp mcp.json"
    ;;
  pi)
    check_file "$OUT/agents/guide.md" "pi agent"
    check_file "$OUT/skills/implement/SKILL.md" "pi skill"
    check_file "$OUT/prompts/plan.md" "pi prompt"
    ;;
  esac
done
expect_fail "build-plugin rejects bad target" bash -c "cd '$ROOT' && HOME='$HOME_ISO' '$BIN' build-plugin --target nope --out '$ROOT/bn'"
expect_fail "build-plugin refuses non-empty out without --force" bash -c "cd '$ROOT' && HOME='$HOME_ISO' '$BIN' build-plugin --target claude --out '$ROOT/bundle-claude'"
expect_ok "build-plugin overwrites with --force" bash -c "cd '$ROOT' && HOME='$HOME_ISO' '$BIN' build-plugin --target claude --out '$ROOT/bundle-claude' --force"

echo "== Scenario 10: import preserves originals + canonical extract =="
S="$ROOT/imp"
mkdir -p "$S/.opencode/agents"
printf '{"$schema":"x","mcp":{"filesystem":{"type":"local","command":["echo"]}}}\n' >"$S/opencode.json"
printf -- '---\ndescription: guide\n---\nbody\n' >"$S/.opencode/agents/guide.md"
printf '# imported\n' >"$S/AGENTS.md"
if (cd "$S" && run import --no-interactive) >/tmp/lazyai-e2e.out 2>&1; then
  ok "import exits 0"
else
  bad "import failed"
  tail -5 /tmp/lazyai-e2e.out
fi
check_file "$S/.ai/lazyai.json" "import: canonical manifest"
check_file "$S/.ai/migration-report.md" "import: migration report"
check_file "$S/.ai/agents/guide.md" "import: canonical agent extracted"
check_file "$S/opencode.json" "import: original opencode.json preserved"
ls "$S/.ai/adapters"/*/raw >/dev/null 2>&1 && ok "import: raw adapter assets preserved" || bad "import: no raw assets"

echo "== Scenario 11: eject removes metadata, keeps native =="
if (cd "$P" && run eject --no-interactive) >/tmp/lazyai-e2e.out 2>&1; then
  ok "eject exits 0"
else
  bad "eject failed"
  tail -5 /tmp/lazyai-e2e.out
fi
check_nofile "$P/.ai/lazyai.json" "eject: manifest removed"
check_nofile "$P/.ai/lock.json" "eject: lock removed"
check_nofile "$P/.ai-setup.db" "eject: setup db removed"
check_file "$P/opencode.json" "eject: native opencode kept"
check_file "$P/.ai/mcp.json" "eject: canonical mcp kept"

echo
echo "==================== RESULT: $PASS passed, $FAIL failed ===================="
[ "$FAIL" -eq 0 ]
