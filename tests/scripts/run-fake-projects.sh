#!/usr/bin/env bash
set -eo pipefail

# Scripts to setup fake projects, execute lazyai-cli scenarios, and collect evidence.

CLI_BIN="$(pwd)/packages/cli/lazyai-cli"
BASE_TMP_DIR="/tmp/lazyai-fake-projects"
EVIDENCES_DIR="$(pwd)/evidences"

if [[ ! -x "$CLI_BIN" ]]; then
  echo "Error: lazyai-cli binary not found or not executable at $CLI_BIN"
  exit 1
fi

TOOLS=("opencode" "claude-code" "copilot" "pi" "antigravity")
mkdir -p "$EVIDENCES_DIR"

PRESETS=("minimal" "standard" "full")
SCOPES=("project") # Testing global would pollute the host, restricting to project for safety in this script unless we isolate HOME

echo "Starting Fake Projects Matrix Test..."
echo "Metadata: $(date)" > "$EVIDENCES_DIR/run_metadata.txt"

# Counter for scenario ID
SCENARIO_ID=1

run_scenario() {
  local tool=$1
  local preset=$2
  local state=$3
  
  local scenario_name=$(printf "scenario_%03d_%s_%s_%s" "$SCENARIO_ID" "$tool" "$preset" "$state")
  local target_dir="$BASE_TMP_DIR/$scenario_name"
  local evidence_out="$EVIDENCES_DIR/$scenario_name"
  
  echo "Running $scenario_name..."
  
  mkdir -p "$target_dir"
  mkdir -p "$evidence_out/logs"
  mkdir -p "$evidence_out/configs"
  
  # Setup initial state
  if [[ "$state" == "git" ]]; then
    (cd "$target_dir" && git init >/dev/null 2>&1)
  fi

  # Record initial tree
  tree -a "$target_dir" > "$evidence_out/initial_tree.txt" 2>/dev/null || find "$target_dir" > "$evidence_out/initial_tree.txt"

  # We set up an isolated HOME so that global config doesn't pollute the user's actual home
  export HOME="$evidence_out/home"
  mkdir -p "$HOME"

  # 1. INIT
  echo "  - Init"
  set +e
  (cd "$target_dir" && "$CLI_BIN" init --tools "$tool" --preset "$preset" --scope project --no-interactive > "$evidence_out/logs/01_init.log" 2>&1)
  local init_exit=$?
  set -e
  echo "init_exit_code=$init_exit" > "$evidence_out/logs/01_init.status"

  # 2. DOCTOR
  echo "  - Doctor"
  set +e
  (cd "$target_dir" && "$CLI_BIN" doctor --json > "$evidence_out/logs/02_doctor.json" 2>&1)
  set -e

  # 3. COMPILE
  echo "  - Compile"
  set +e
  (cd "$target_dir" && "$CLI_BIN" compile > "$evidence_out/logs/03_compile.log" 2>&1)
  set -e

  # 4. VALIDATE
  echo "  - Validate"
  set +e
  (cd "$target_dir" && "$CLI_BIN" validate skills > "$evidence_out/logs/04_validate.log" 2>&1)
  (cd "$target_dir" && "$CLI_BIN" validate agents >> "$evidence_out/logs/04_validate.log" 2>&1)
  set -e

  # 5. CREATE ARTIFACTS
  echo "  - Create"
  set +e
  (cd "$target_dir" && "$CLI_BIN" create agent test-agent --description "Test Agent" --no-interactive > "$evidence_out/logs/05_create.log" 2>&1)
  (cd "$target_dir" && "$CLI_BIN" create skill test-skill --description "Test Skill" --no-interactive >> "$evidence_out/logs/05_create.log" 2>&1)
  set -e

  # 6. LEDGER
  echo "  - Ledger"
  set +e
  (cd "$target_dir" && "$CLI_BIN" ledger init > "$evidence_out/logs/06_ledger.log" 2>&1)
  (cd "$target_dir" && "$CLI_BIN" ledger append test "action=matrix_test" >> "$evidence_out/logs/06_ledger.log" 2>&1)
  (cd "$target_dir" && "$CLI_BIN" ledger verify >> "$evidence_out/logs/06_ledger.log" 2>&1)
  set -e

  # Post execution capture
  tree -a "$target_dir" > "$evidence_out/final_tree.txt" 2>/dev/null || find "$target_dir" > "$evidence_out/final_tree.txt"

  # Copy configs
  if [[ -d "$target_dir/.opencode" ]]; then cp -r "$target_dir/.opencode" "$evidence_out/configs/" 2>/dev/null || true; fi
  if [[ -d "$target_dir/.claude" ]]; then cp -r "$target_dir/.claude" "$evidence_out/configs/" 2>/dev/null || true; fi

  # Tarball project state
  (cd "$BASE_TMP_DIR" && tar -czf "$evidence_out/project_state.tar.gz" "$scenario_name")

  SCENARIO_ID=$((SCENARIO_ID + 1))
}
# Run all scenarios explicitly
for t in opencode claude-code copilot pi antigravity; do
  run_scenario "$t" "standard" "empty"
done
for p in minimal standard full; do
  run_scenario "opencode" "$p" "git"
done
run_scenario "claude-code" "full" "empty"
run_scenario "copilot" "minimal" "git"
echo "Test execution complete. Evidences stored in $EVIDENCES_DIR"
