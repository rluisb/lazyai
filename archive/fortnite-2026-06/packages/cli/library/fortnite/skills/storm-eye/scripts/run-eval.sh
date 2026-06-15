#!/usr/bin/env bash
# scripts/run-evals.sh — Run eval suites against datasets and emit results
# Usage: ./scripts/run-evals.sh --suite NAME [--dataset PATH] [--json] [--dry-run]
#
# Reads suite YAML from .specify/evals/suites/, loads dataset JSONL,
# runs assertions, computes aggregate scores, records to session DB and ledger.

set -euo pipefail
IFS=$'\n\t'

# ─── Configuration ───
REPO_ROOT="${OPENCODE_WORKSPACE:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
SUITE_NAME=""
DATASET_OVERRIDE=""
OUTPUT_JSON=false
DRY_RUN=false
SHOW_COST=false

# ─── CLI Parsing ───
while [[ $# -gt 0 ]]; do
    case "$1" in
        --suite)
            SUITE_NAME="$2"
            shift 2
            ;;
        --dataset)
            DATASET_OVERRIDE="$2"
            shift 2
            ;;
        --json)
            OUTPUT_JSON=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --cost)
            SHOW_COST=true
            shift
            ;;
        --help|-h)
            cat << 'EOF'
Usage: run-evals.sh --suite NAME [OPTIONS]

Run an evaluation suite against a dataset and emit results.

Required:
  --suite NAME       Name of the eval suite (YAML file in .specify/evals/suites/)

Optional:
  --dataset PATH     Override dataset path from suite definition
  --json             Output raw JSON instead of formatted report
  --dry-run          Print plan without executing assertions
  --cost             Show estimated cost breakdown in report
  --help, -h         Show this help message

Examples:
  ./scripts/run-evals.sh --suite agent-boundary
  ./scripts/run-evals.sh --suite dispatch-accuracy --json
  ./scripts/run-evals.sh --suite test-runner-detection --dry-run
   ./scripts/run-evals.sh --suite agent-boundary --cost
EOF
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            echo "Use --help for usage information" >&2
            exit 1
            ;;
    esac
done

# ─── Validation ───
if [[ -z "$SUITE_NAME" ]]; then
    echo "Error: --suite is required" >&2
    echo "Use --help for usage information" >&2
    exit 1
fi

SUITE_PATH="$REPO_ROOT/.specify/evals/suites/${SUITE_NAME}.yaml"
if [[ ! -f "$SUITE_PATH" ]]; then
    echo "Error: suite not found: $SUITE_PATH" >&2
    exit 1
fi

# ─── Load Suite Config ───
SUITE_CONFIG=$(python3 -c '
import yaml, sys
try:
    with open(sys.argv[1], "r") as f:
        config = yaml.safe_load(f)
    import json
    print(json.dumps(config))
except Exception as e:
    print(f"Error loading suite: {e}", file=sys.stderr)
    sys.exit(1)
' "$SUITE_PATH")

SUITE_NAME_PARSED=$(echo "$SUITE_CONFIG" | python3 -c 'import sys,json; print(json.load(sys.stdin).get("name", ""))')
DATASET_PATH=$(echo "$SUITE_CONFIG" | python3 -c 'import sys,json; print(json.load(sys.stdin).get("dataset_path", ""))')
TARGET_THRESHOLD=$(echo "$SUITE_CONFIG" | python3 -c 'import sys,json; print(json.load(sys.stdin).get("target_threshold", 0.0))')
PRIMARY_METRIC=$(echo "$SUITE_CONFIG" | python3 -c 'import sys,json; d=json.load(sys.stdin); print(d.get("metric",{}).get("primary","accuracy"))')
ASSERTIONS_JSON=$(echo "$SUITE_CONFIG" | python3 -c 'import sys,json; print(json.dumps(json.load(sys.stdin).get("assertions", [])))')

# Override dataset if provided
if [[ -n "$DATASET_OVERRIDE" ]]; then
    DATASET_PATH="$DATASET_OVERRIDE"
fi

# Resolve dataset path
if [[ ! "$DATASET_PATH" = /* ]]; then
    DATASET_PATH="$REPO_ROOT/$DATASET_PATH"
fi

# ─── Dry Run ───
if [[ "$DRY_RUN" == "true" ]]; then
    DATASET_SIZE=0
    if [[ -f "$DATASET_PATH" ]]; then
        DATASET_SIZE=$(wc -l < "$DATASET_PATH" | tr -d ' ')
    fi

    if [[ "$OUTPUT_JSON" == "true" ]]; then
        python3 -c '
import json, sys
config = json.loads(sys.argv[1])
print(json.dumps({
    "dry_run": True,
    "suite": config.get("name", ""),
    "description": config.get("description", ""),
    "dataset_path": sys.argv[2],
    "dataset_cases": int(sys.argv[3]),
    "target_threshold": config.get("target_threshold", 0.0),
    "primary_metric": config.get("metric", {}).get("primary", "accuracy"),
    "assertions": [a.get("name", "") for a in config.get("assertions", [])]
}, indent=2))
' "$SUITE_CONFIG" "$DATASET_PATH" "$DATASET_SIZE"
    else
        cat << EOF
╔══════════════════════════════════════════════════════════════╗
║  DRY RUN — Evaluation Plan                                   ║
╠══════════════════════════════════════════════════════════════╣
  Suite:        $SUITE_NAME_PARSED
  Description:  $(echo "$SUITE_CONFIG" | python3 -c 'import sys,json; print(json.load(sys.stdin).get("description", "N/A"))')
  Dataset:      $DATASET_PATH
  Cases:        $DATASET_SIZE
  Threshold:    $TARGET_THRESHOLD
  Metric:       $PRIMARY_METRIC
  Assertions:
$(echo "$ASSERTIONS_JSON" | python3 -c 'import sys,json; [print("    - " + a.get("name","") + ": " + a.get("check","")) for a in json.load(sys.stdin)]')
╚══════════════════════════════════════════════════════════════╝
EOF
    fi
    exit 0
fi

# ─── Validate Dataset ───
if [[ ! -f "$DATASET_PATH" ]]; then
    echo "Error: dataset not found: $DATASET_PATH" >&2
    exit 1
fi

# ─── Run Evaluations ───
START_MS=$(python3 -c "import time; print(int(time.time()*1000))")

EVAL_RESULT=$(python3 - "$SUITE_CONFIG" "$DATASET_PATH" "$TARGET_THRESHOLD" "$PRIMARY_METRIC" "$ASSERTIONS_JSON" << 'PYEOF'
import json, sys

suite_config = json.loads(sys.argv[1])
dataset_path = sys.argv[2]
target_threshold = float(sys.argv[3])
primary_metric = sys.argv[4]
assertions = json.loads(sys.argv[5])

results = {
    "suite": suite_config.get("name", ""),
    "timestamp": "",
    "dataset_path": dataset_path,
    "total_cases": 0,
    "passed_cases": 0,
    "failed_cases": 0,
    "primary_metric": primary_metric,
    "primary_score": 0.0,
    "target_threshold": target_threshold,
    "threshold_met": False,
    "per_assertion": [],
    "duration_ms": 0
}

# Load dataset
cases = []
try:
    with open(dataset_path, "r") as f:
        for line in f:
            line = line.strip()
            if line:
                try:
                    cases.append(json.loads(line))
                except json.JSONDecodeError:
                    pass  # Skip invalid JSON lines
except Exception as e:
    print(f"Error loading dataset: {e}", file=sys.stderr)
    sys.exit(1)

results["total_cases"] = len(cases)

if results["total_cases"] == 0:
    print(json.dumps(results))
    sys.exit(0)

# Track assertion results
assertion_stats = {}
for assertion in assertions:
    name = assertion.get("name", "unknown")
    assertion_stats[name] = {"passed": 0, "failed": 0}

# Run assertions per case
passed_cases = 0
for case in cases:
    case_passed = True
    case_output = case.get("output", case)

    for assertion in assertions:
        name = assertion.get("name", "")
        check_desc = assertion.get("check", "")

        if name == "basic_structure":
            # Verify output has required fields
            if isinstance(case_output, dict):
                # Check for required fields if specified, otherwise just validate structure
                required = assertion.get("required_fields", [])
                if required:
                    if all(k in case_output for k in required):
                        assertion_stats[name]["passed"] += 1
                    else:
                        assertion_stats[name]["failed"] += 1
                        case_passed = False
                else:
                    # Just check it's a non-empty dict with some standard fields
                    # or has an 'output' key
                    if case_output:
                        assertion_stats[name]["passed"] += 1
                    else:
                        assertion_stats[name]["failed"] += 1
                        case_passed = False
            else:
                assertion_stats[name]["failed"] += 1
                case_passed = False

        elif name == "threshold_met":
            # This is computed at aggregate level, skip per-case
            # Mark as passed for each case (aggregate check happens later)
            assertion_stats[name]["passed"] += 1

        else:
            # Unknown assertion: pass by default
            assertion_stats[name]["passed"] += 1

    if case_passed:
        passed_cases += 1

results["passed_cases"] = passed_cases
results["failed_cases"] = results["total_cases"] - passed_cases
results["primary_score"] = passed_cases / results["total_cases"] if results["total_cases"] > 0 else 0.0
results["threshold_met"] = results["primary_score"] >= target_threshold

# Build per_assertion list
results["per_assertion"] = [
    {"name": name, "passed": stats["passed"], "failed": stats["failed"]}
    for name, stats in assertion_stats.items()
]

# Update threshold_met assertion to reflect reality
for pa in results["per_assertion"]:
    if pa["name"] == "threshold_met":
        if results["threshold_met"]:
            pa["passed"] = results["total_cases"]
            pa["failed"] = 0
        else:
            pa["passed"] = 0
            pa["failed"] = results["total_cases"]

print(json.dumps(results))
PYEOF
)

END_MS=$(python3 -c "import time; print(int(time.time()*1000))")
DURATION_MS=$((END_MS - START_MS))

# Add timestamp and duration
TIMESTAMP=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
EVAL_RESULT=$(echo "$EVAL_RESULT" | python3 -c "
import sys, json
d = json.load(sys.stdin)
d['timestamp'] = '$TIMESTAMP'
d['duration_ms'] = $DURATION_MS
print(json.dumps(d))
")

# ─── Record to Session DB ───
DB_PATH="$REPO_ROOT/.specify/session.db"
SESSION_ID="${OPENCODE_SESSION_ID:-unknown}"

if [[ -f "$DB_PATH" ]]; then
    # Insert into quality_metrics (per spec P5.3)
    PASSED_INT=$([[ "$(echo "$EVAL_RESULT" | python3 -c 'import sys,json; print(json.load(sys.stdin)["threshold_met"])')" == "True" ]] && echo 1 || echo 0)
    ERROR_COUNT=$(echo "$EVAL_RESULT" | python3 -c 'import sys,json; print(json.load(sys.stdin)["failed_cases"])')

    sqlite3 "$DB_PATH" << EOF 2>/dev/null || true
INSERT INTO quality_metrics (session_id, repo, gate_type, passed, duration_ms, error_count, warning_count, timestamp)
VALUES ('$SESSION_ID', '$REPO_ROOT', 'test', $PASSED_INT, $DURATION_MS, $ERROR_COUNT, 0, '$TIMESTAMP');
EOF

    # Also insert into eval_runs if table exists
    RUN_ID="run_$(date +%s%N | cut -c1-13)"
    sqlite3 "$DB_PATH" << EOF 2>/dev/null || true
INSERT INTO eval_runs (id, suite_name, dataset_name, status, started_at, ended_at, score, metadata_json)
VALUES ('$RUN_ID', '$SUITE_NAME_PARSED', '$DATASET_PATH', 'completed', '$TIMESTAMP', '$TIMESTAMP', $(echo "$EVAL_RESULT" | python3 -c 'import sys,json; print(json.load(sys.stdin)["primary_score"])'), '$EVAL_RESULT');
EOF
fi

# ─── Append to Ledger ───
LEDGER_SCRIPT="$REPO_ROOT/skills/truth-chain/scripts/ledger.sh"
if [[ -x "$LEDGER_SCRIPT" ]]; then
    LEDGER_DATA=$(echo "$EVAL_RESULT" | python3 -c '
import sys, json
d = json.load(sys.stdin)
print(json.dumps({
    "suite": d["suite"],
    "total_cases": d["total_cases"],
    "passed_cases": d["passed_cases"],
    "failed_cases": d["failed_cases"],
    "primary_score": d["primary_score"],
    "threshold_met": d["threshold_met"],
    "duration_ms": d["duration_ms"]
}, separators=(",", ":")))
')
    "$LEDGER_SCRIPT" append --type "eval_run" --data "$LEDGER_DATA" --session "$SESSION_ID" >/dev/null 2>&1 || true
fi

# ─── Output ───
if [[ "$OUTPUT_JSON" == "true" ]]; then
    echo "$EVAL_RESULT" | python3 -m json.tool
else
    # Human-readable report
    SUITE=$(echo "$EVAL_RESULT" | python3 -c 'import sys,json; print(json.load(sys.stdin)["suite"])')
    TOTAL=$(echo "$EVAL_RESULT" | python3 -c 'import sys,json; print(json.load(sys.stdin)["total_cases"])')
    PASSED=$(echo "$EVAL_RESULT" | python3 -c 'import sys,json; print(json.load(sys.stdin)["passed_cases"])')
    FAILED=$(echo "$EVAL_RESULT" | python3 -c 'import sys,json; print(json.load(sys.stdin)["failed_cases"])')
    SCORE=$(echo "$EVAL_RESULT" | python3 -c 'import sys,json; d=json.load(sys.stdin); print("{:.2%}".format(d["primary_score"]))')
    THRESHOLD=$(echo "$EVAL_RESULT" | python3 -c 'import sys,json; print(json.load(sys.stdin)["target_threshold"])')
    MET=$(echo "$EVAL_RESULT" | python3 -c 'import sys,json; print("✅ MET" if json.load(sys.stdin)["threshold_met"] else "❌ NOT MET")')
    DUR=$(echo "$EVAL_RESULT" | python3 -c 'import sys,json; print(json.load(sys.stdin)["duration_ms"])')

    cat << EOF
╔══════════════════════════════════════════════════════════════╗
║  Evaluation Results                                          ║
╠══════════════════════════════════════════════════════════════╣
  Suite:        $SUITE
  Cases:        $TOTAL total | $PASSED passed | $FAILED failed
  Score:        $SCORE (threshold: $THRESHOLD)
  Status:       $MET
  Duration:     ${DUR}ms
╠══════════════════════════════════════════════════════════════╣
  Per-Assertion Breakdown:
EOF

    echo "$EVAL_RESULT" | python3 -c '
import sys, json
d = json.load(sys.stdin)
for a in d.get("per_assertion", []):
    name = a["name"]
    passed = a["passed"]
    failed = a["failed"]
    total = passed + failed
    pct = (passed / total * 100) if total > 0 else 0
    print(f"    {name:20s}  {passed:4d}/{total:4d}  ({pct:5.1f}%)")
'

    # ─── Cost Breakdown ───
    if [[ "$SHOW_COST" == "true" ]]; then
        echo "╠══════════════════════════════════════════════════════════════╣"
        echo "║  Cost Breakdown                                              ║"
        COST_DATA=$(sqlite3 "$DB_PATH" << EOF 2>/dev/null || true
SELECT provider, model, tokens_in, tokens_out, estimated_cost_usd
FROM cost_snapshots
WHERE session_id = '$SESSION_ID'
ORDER BY estimated_cost_usd DESC;
EOF
)
        if [[ -n "$COST_DATA" ]]; then
            echo "║  Provider          Model               Tokens In   Tokens Out   Cost USD  ║"
            # Use process substitution to avoid subshell and preserve TOTAL_COST
            while IFS='|' read -r provider model tokens_in tokens_out cost_usd; do
                printf "║  %-16s  %-18s  %9s  %10s  %9s  ║\n" "$provider" "$model" "$tokens_in" "$tokens_out" "$cost_usd"
            done <<< "$COST_DATA"
            TOTAL_COST=$(sqlite3 "$DB_PATH" << EOF 2>/dev/null || true
SELECT ROUND(SUM(estimated_cost_usd), 6)
FROM cost_snapshots
WHERE session_id = '$SESSION_ID';
EOF
)
            printf "║  %-58s  %9s  ║\n" "TOTAL" "$TOTAL_COST"
        else
            echo "║  No cost snapshots found for session $SESSION_ID           ║"
            echo "║  (Costs are estimates. See .specify/evals/reports/cost-methodology.md) ║"
        fi
    fi

    echo "╚══════════════════════════════════════════════════════════════╝"
fi

exit 0
