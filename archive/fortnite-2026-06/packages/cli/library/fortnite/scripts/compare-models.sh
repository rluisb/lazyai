#!/usr/bin/env bash
# scripts/compare-models.sh — Compare two models on the same eval suite
# Usage: ./scripts/compare-models.sh --suite NAME --model-a PROVIDER/MODEL --model-b PROVIDER/MODEL [OPTIONS]
#
# Runs (or loads pre-computed) eval results for two models and produces a
# comparison report with winner, deltas, and recommendation.

set -euo pipefail
IFS=$'\n\t'

# ─── Configuration ───
REPO_ROOT="${OPENCODE_WORKSPACE:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
SUITE_NAME=""
DATASET_OVERRIDE=""
MODEL_A_SPEC=""
MODEL_B_SPEC=""
RESULT_A_PATH=""
RESULT_B_PATH=""
OUTPUT_JSON=false
DRY_RUN=false

# ─── Helpers ───
die() {
    echo "Error: $1" >&2
    exit 1
}

usage() {
    cat << 'EOF'
Usage: compare-models.sh --suite NAME --model-a PROVIDER/MODEL --model-b PROVIDER/MODEL [OPTIONS]

Compare two models on the same eval suite.

Required:
  --suite NAME              Name of the eval suite
  --model-a PROVIDER/MODEL  First model (e.g. openai/gpt-4o)
  --model-b PROVIDER/MODEL  Second model (e.g. ollama/llama3)

Optional:
  --dataset PATH            Override dataset path from suite definition
  --result-a PATH           Pre-computed eval result JSON for model A
  --result-b PATH           Pre-computed eval result JSON for model B
  --json                    Output raw JSON instead of formatted report
  --dry-run                 Simulate results (no actual model execution)
  --help, -h                Show this help message

Modes:
  1. Live eval:   omit --result-a and --result-b. The script will call
                  run-evals.sh for each model (requires model execution infra).
  2. Pre-computed: provide --result-a and --result-b with eval JSON files.
  3. Dry-run:      use --dry-run to simulate results for both models.

Examples:
  ./scripts/compare-models.sh --suite agent-boundary \
      --model-a openai/gpt-4o --model-b ollama/llama3 --dry-run

  ./scripts/compare-models.sh --suite dispatch-accuracy \
      --model-a openai/gpt-4o --model-b openai/gpt-3.5-turbo \
      --result-a /tmp/eval-a.json --result-b /tmp/eval-b.json --json
EOF
}

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
        --model-a)
            MODEL_A_SPEC="$2"
            shift 2
            ;;
        --model-b)
            MODEL_B_SPEC="$2"
            shift 2
            ;;
        --result-a)
            RESULT_A_PATH="$2"
            shift 2
            ;;
        --result-b)
            RESULT_B_PATH="$2"
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
        --help|-h)
            usage
            exit 0
            ;;
        *)
            die "Unknown option: $1. Use --help for usage information."
            ;;
    esac
done

# ─── Validation ───
[[ -n "$SUITE_NAME" ]] || die "--suite is required. Use --help for usage information."
[[ -n "$MODEL_A_SPEC" ]] || die "--model-a is required. Use --help for usage information."
[[ -n "$MODEL_B_SPEC" ]] || die "--model-b is required. Use --help for usage information."

# Validate model specs contain a slash
[[ "$MODEL_A_SPEC" == */* ]] || die "--model-a must be in PROVIDER/MODEL format (e.g. openai/gpt-4o). Got: $MODEL_A_SPEC"
[[ "$MODEL_B_SPEC" == */* ]] || die "--model-b must be in PROVIDER/MODEL format (e.g. openai/gpt-4o). Got: $MODEL_B_SPEC"

# Determine mode
MODE="live"
if [[ "$DRY_RUN" == "true" ]]; then
    MODE="dry-run"
elif [[ -n "$RESULT_A_PATH" && -n "$RESULT_B_PATH" ]]; then
    MODE="precomputed"
    [[ -f "$RESULT_A_PATH" ]] || die "Result file not found: $RESULT_A_PATH"
    [[ -f "$RESULT_B_PATH" ]] || die "Result file not found: $RESULT_B_PATH"
elif [[ -n "$RESULT_A_PATH" || -n "$RESULT_B_PATH" ]]; then
    die "Both --result-a and --result-b are required for pre-computed mode."
fi

# ─── Parse model specs ───
PROVIDER_A="${MODEL_A_SPEC%%/*}"
MODEL_A="${MODEL_A_SPEC#*/}"
PROVIDER_B="${MODEL_B_SPEC%%/*}"
MODEL_B="${MODEL_B_SPEC#*/}"

# ─── Run or load evals ───
run_eval_for_model() {
    local provider="$1"
    local model="$2"
    local result_file="$3"

    local extra_args=()
    [[ -n "$DATASET_OVERRIDE" ]] && extra_args+=("--dataset" "$DATASET_OVERRIDE")

    # Note: run-evals.sh does not yet accept --model, but we document the intent.
    # For now we call it without model args and the caller is expected to have
    # model-specific results or use --dry-run.
    "$REPO_ROOT/scripts/run-evals.sh" \
        --suite "$SUITE_NAME" \
        "${extra_args[@]}" \
        --json > "$result_file"
}

RESULT_A_TMP=""
RESULT_B_TMP=""

if [[ "$MODE" == "live" ]]; then
    # Live mode: run evals for both models
    RESULT_A_TMP=$(mktemp)
    RESULT_B_TMP=$(mktemp)
    # shellcheck disable=SC2064
    trap "rm -f '$RESULT_A_TMP' '$RESULT_B_TMP'" EXIT

    echo "Running eval for model A: $PROVIDER_A/$MODEL_A ..." >&2
    run_eval_for_model "$PROVIDER_A" "$MODEL_A" "$RESULT_A_TMP"

    echo "Running eval for model B: $PROVIDER_B/$MODEL_B ..." >&2
    run_eval_for_model "$PROVIDER_B" "$MODEL_B" "$RESULT_B_TMP"

    RESULT_A_PATH="$RESULT_A_TMP"
    RESULT_B_PATH="$RESULT_B_TMP"

elif [[ "$MODE" == "dry-run" ]]; then
    # Dry-run: simulate results
    RESULT_A_TMP=$(mktemp)
    RESULT_B_TMP=$(mktemp)
    # shellcheck disable=SC2064
    trap "rm -f '$RESULT_A_TMP' '$RESULT_B_TMP'" EXIT

    TIMESTAMP=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

    # Simulate model A results (slightly better baseline)
    python3 -c '
import json, sys, random
random.seed(sys.argv[1])
provider, model, suite, ts = sys.argv[2:6]
total = random.randint(80, 120)
passed = random.randint(int(total*0.55), int(total*0.85))
score = round(passed / total, 4)
result = {
    "suite": suite,
    "timestamp": ts,
    "dataset_path": "/tmp/simulated-dataset.jsonl",
    "total_cases": total,
    "passed_cases": passed,
    "failed_cases": total - passed,
    "primary_metric": "accuracy",
    "primary_score": score,
    "target_threshold": 0.7,
    "threshold_met": score >= 0.7,
    "per_assertion": [{"name": "assertion-1", "passed": passed, "failed": total - passed}],
    "duration_ms": random.randint(800, 3500),
    "_meta": {"provider": provider, "model": model, "mode": "dry-run"}
}
print(json.dumps(result, indent=2))
' "${MODEL_A_SPEC}" "$PROVIDER_A" "$MODEL_A" "$SUITE_NAME" "$TIMESTAMP" > "$RESULT_A_TMP"

    # Simulate model B results
    python3 -c '
import json, sys, random
random.seed(sys.argv[1])
provider, model, suite, ts = sys.argv[2:6]
total = random.randint(80, 120)
passed = random.randint(int(total*0.50), int(total*0.80))
score = round(passed / total, 4)
result = {
    "suite": suite,
    "timestamp": ts,
    "dataset_path": "/tmp/simulated-dataset.jsonl",
    "total_cases": total,
    "passed_cases": passed,
    "failed_cases": total - passed,
    "primary_metric": "accuracy",
    "primary_score": score,
    "target_threshold": 0.7,
    "threshold_met": score >= 0.7,
    "per_assertion": [{"name": "assertion-1", "passed": passed, "failed": total - passed}],
    "duration_ms": random.randint(800, 3500),
    "_meta": {"provider": provider, "model": model, "mode": "dry-run"}
}
print(json.dumps(result, indent=2))
' "${MODEL_B_SPEC}" "$PROVIDER_B" "$MODEL_B" "$SUITE_NAME" "$TIMESTAMP" > "$RESULT_B_TMP"

    RESULT_A_PATH="$RESULT_A_TMP"
    RESULT_B_PATH="$RESULT_B_TMP"
fi

# ─── Extract metrics ───
extract_json() {
    python3 -c 'import json,sys; d=json.load(open(sys.argv[1])); print(json.dumps(d.get(sys.argv[2], {})))' "$1" "$2"
}

SCORE_A=$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get("primary_score", 0.0))' "$RESULT_A_PATH")
LATENCY_A=$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get("duration_ms", 0))' "$RESULT_A_PATH")
COST_A=$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get("_meta",{}).get("cost_usd", 0.0))' "$RESULT_A_PATH")

SCORE_B=$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get("primary_score", 0.0))' "$RESULT_B_PATH")
LATENCY_B=$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get("duration_ms", 0))' "$RESULT_B_PATH")
COST_B=$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get("_meta",{}).get("cost_usd", 0.0))' "$RESULT_B_PATH")

# ─── Compute deltas ───
SCORE_DELTA=$(python3 -c "print(round(float('$SCORE_A') - float('$SCORE_B'), 4))")
LATENCY_DELTA=$(python3 -c "print(int('$LATENCY_A') - int('$LATENCY_B'))")
COST_DELTA=$(python3 -c "print(round(float('$COST_A') - float('$COST_B'), 4))")

# ─── Determine winner ───
WINNER="tie"
RECOMMENDATION=""

if (( $(python3 -c "print(1 if float('$SCORE_DELTA') > 0.001 else 0)") )); then
    WINNER="model_a"
    RECOMMENDATION="Model A ($PROVIDER_A/$MODEL_A) outperforms Model B on primary score."
elif (( $(python3 -c "print(1 if float('$SCORE_DELTA') < -0.001 else 0)") )); then
    WINNER="model_b"
    RECOMMENDATION="Model B ($PROVIDER_B/$MODEL_B) outperforms Model A on primary score."
else
    # Score tie — consider latency
    if (( LATENCY_DELTA < -50 )); then
        WINNER="model_a"
        RECOMMENDATION="Scores are tied, but Model A is significantly faster."
    elif (( LATENCY_DELTA > 50 )); then
        WINNER="model_b"
        RECOMMENDATION="Scores are tied, but Model B is significantly faster."
    else
        WINNER="tie"
        RECOMMENDATION="Both models perform similarly on score and latency."
    fi
fi

# ─── Build comparison JSON ───
COMPARISON_JSON=$(python3 -c '
import json, sys
suite, dataset, ts = sys.argv[1:4]
provider_a, model_a, score_a, latency_a, cost_a = sys.argv[4:9]
provider_b, model_b, score_b, latency_b, cost_b = sys.argv[9:14]
winner, score_delta, latency_delta, cost_delta, recommendation = sys.argv[14:19]

result = {
    "suite": suite,
    "dataset": dataset,
    "comparison_timestamp": ts,
    "model_a": {
        "provider": provider_a,
        "model": model_a,
        "score": float(score_a),
        "latency_ms": int(latency_a),
        "cost_usd": float(cost_a)
    },
    "model_b": {
        "provider": provider_b,
        "model": model_b,
        "score": float(score_b),
        "latency_ms": int(latency_b),
        "cost_usd": float(cost_b)
    },
    "winner": winner,
    "score_delta": float(score_delta),
    "latency_delta_ms": int(latency_delta),
    "cost_delta_usd": float(cost_delta),
    "recommendation": recommendation
}
print(json.dumps(result, indent=2))
' \
    "$SUITE_NAME" \
    "${DATASET_OVERRIDE:-""}" \
    "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" \
    "$PROVIDER_A" "$MODEL_A" "$SCORE_A" "$LATENCY_A" "$COST_A" \
    "$PROVIDER_B" "$MODEL_B" "$SCORE_B" "$LATENCY_B" "$COST_B" \
    "$WINNER" "$SCORE_DELTA" "$LATENCY_DELTA" "$COST_DELTA" \
    "$RECOMMENDATION")

# ─── Output ───
if [[ "$OUTPUT_JSON" == "true" ]]; then
    echo "$COMPARISON_JSON"
else
    # Human-readable report using Python via stdin to avoid quoting issues
    echo "$COMPARISON_JSON" | python3 -c '
import json, sys
c = json.load(sys.stdin)

a = c["model_a"]
b = c["model_b"]

def fmt_delta(v):
    sign = "+" if v > 0 else ""
    return f"{sign}{v}"

print("╔══════════════════════════════════════════════════════════════╗")
print("║  MODEL COMPARISON REPORT                                    ║")
print("╠══════════════════════════════════════════════════════════════╣")
print("  Suite:        " + c["suite"])
print("  Dataset:      " + (c["dataset"] or "(suite default)"))
print("  Timestamp:    " + c["comparison_timestamp"])
print("")
print("  ── Model A ──────────────────────────────────────────────────")
print("    Provider:   " + a["provider"])
print("    Model:      " + a["model"])
print("    Score:      {:.4f}".format(a["score"]))
print("    Latency:    {} ms".format(a["latency_ms"]))
print("    Cost:       ${:.4f}".format(a["cost_usd"]))
print("")
print("  ── Model B ──────────────────────────────────────────────────")
print("    Provider:   " + b["provider"])
print("    Model:      " + b["model"])
print("    Score:      {:.4f}".format(b["score"]))
print("    Latency:    {} ms".format(b["latency_ms"]))
print("    Cost:       ${:.4f}".format(b["cost_usd"]))
print("")
print("  ── Deltas (A − B) ──────────────────────────────────────────")
print("    Score:      " + fmt_delta(c["score_delta"]))
print("    Latency:    " + fmt_delta(c["latency_delta_ms"]) + " ms")
print("    Cost:       $" + fmt_delta(c["cost_delta_usd"]))
print("")
print("  Winner:       " + c["winner"].upper())
print("  Recommendation: " + c["recommendation"])
print("╚══════════════════════════════════════════════════════════════╝")
'
fi

exit 0
