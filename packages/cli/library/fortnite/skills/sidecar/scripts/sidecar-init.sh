#!/usr/bin/env bash
# sidecar-init.sh — Interactive init wizard for the sidecar skill
# Creates .sidecar.yml, sets up the sidecar directory, and runs initial index.
#
# Usage: sidecar init [options]
#   --dir <path>      Parent directory containing repos (default: parent of CWD)
#   --sidecar <dir>   Directory to use as the sidecar (skips prompt)
#   --help            Show usage

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMMON_SH="$SCRIPT_DIR/_common.sh"

# ---------------------------------------------------------------------------
# Defaults
# ---------------------------------------------------------------------------
PARENT_DIR=""
SIDECAR_DIR_ARG=""

# ---------------------------------------------------------------------------
# Usage
# ---------------------------------------------------------------------------
usage() {
    cat <<EOF
Usage: sidecar init [options]

Initialize a new sidecar workspace. Creates .sidecar.yml in the current
or specified parent directory, prompts for repo selection, and runs the
initial index.

Options:
  --dir <path>      Parent directory containing repos (default: parent of CWD)
  --sidecar <dir>   Directory to use as the sidecar (skips prompt)
  --help            Show this message and exit
EOF
}

# ---------------------------------------------------------------------------
# Parse arguments
# ---------------------------------------------------------------------------
while [[ $# -gt 0 ]]; do
    case "$1" in
        --dir)
            shift
            if [[ $# -eq 0 ]]; then
                echo "Error: --dir requires an argument" >&2
                usage
                exit 1
            fi
            PARENT_DIR="$1"
            ;;
        --sidecar)
            shift
            if [[ $# -eq 0 ]]; then
                echo "Error: --sidecar requires an argument" >&2
                usage
                exit 1
            fi
            SIDECAR_DIR_ARG="$1"
            ;;
        --help|-h)
            usage
            exit 0
            ;;
        *)
            echo "Error: unknown option: $1" >&2
            usage
            exit 1
            ;;
    esac
    shift
done

# ---------------------------------------------------------------------------
# 1. Check if .sidecar.yml already exists (walk up from CWD)
# ---------------------------------------------------------------------------
find_existing_config() {
    local dir="$PWD"
    while [[ "$dir" != "/" ]]; do
        if [[ -f "$dir/.sidecar.yml" ]]; then
            echo "$dir/.sidecar.yml"
            return 0
        fi
        dir="$(dirname "$dir")"
    done
    if [[ -f "/.sidecar.yml" ]]; then
        echo "/.sidecar.yml"
        return 0
    fi
    return 1
}

EXISTING_CONFIG=""
REINIT=false
if EXISTING_CONFIG=$(find_existing_config); then
    echo "Found existing sidecar config: $EXISTING_CONFIG"
    read -rp "Reinitialize? [y/N] " answer
    if [[ "$answer" != "y" && "$answer" != "Y" ]]; then
        echo "Aborted."
        exit 0
    fi
    REINIT=true
fi

# ---------------------------------------------------------------------------
# Helper: count git repos (dirs with .git/) at a given path
# ---------------------------------------------------------------------------
_count_git_repos() {
    local path="$1"
    local count=0
    for d in "$path"/*/; do
        [[ -d "$d/.git" ]] && count=$((count + 1))
    done
    echo "$count"
}

# ---------------------------------------------------------------------------
# 2. Determine parent directory
# ---------------------------------------------------------------------------
if [[ -z "$PARENT_DIR" ]]; then
    # Walk up from CWD looking for a level with 2+ git repos as siblings.
    # If CWD itself has 2+ git repos, use CWD. Otherwise walk up.
    # Fallback to parent of CWD if no such level is found.
    PARENT_DIR=""
    dir="$PWD"
    while [[ "$dir" != "/" ]]; do
        repo_count="$(_count_git_repos "$dir")"
        if [[ "$repo_count" -ge 2 ]]; then
            PARENT_DIR="$dir"
            break
        fi
        dir="$(dirname "$dir")"
    done
    if [[ -z "$PARENT_DIR" ]]; then
        PARENT_DIR="$(dirname "$PWD")"
    fi
fi

# Resolve to absolute path
PARENT_DIR="$(cd "$PARENT_DIR" 2>/dev/null && pwd)" || {
    echo "Error: parent directory does not exist: $PARENT_DIR" >&2
    exit 1
}

# ---------------------------------------------------------------------------
# 3. List ALL subdirectories in the parent dir
# ---------------------------------------------------------------------------
ALL_DIRS=()
while IFS= read -r line; do
    ALL_DIRS+=("$line")
done < <(find "$PARENT_DIR" -maxdepth 1 -mindepth 1 -type d ! -name '.*' | sort)

if [[ ${#ALL_DIRS[@]} -eq 0 ]]; then
    echo "Error: no sibling directories found in $PARENT_DIR" >&2
    echo "       Create some repos first, or use --dir to point to a different parent." >&2
    exit 1
fi

# ---------------------------------------------------------------------------
# 4. Present numbered list for multi-select
# ---------------------------------------------------------------------------
echo ""
echo "Available directories in $PARENT_DIR:"
for i in "${!ALL_DIRS[@]}"; do
    printf '  %d. %s\n' "$((i+1))" "$(basename "${ALL_DIRS[$i]}")"
done
echo ""

# If reinit, show current repos and check if all are already tracked
CURRENT_REPOS=()
if [[ "$REINIT" == true ]]; then
    CURRENT_REPOS=()
    while IFS= read -r line; do
        CURRENT_REPOS+=("$line")
    done < <(yq eval '.repos[]' "$EXISTING_CONFIG")
    if [[ ${#CURRENT_REPOS[@]} -gt 0 ]]; then
        echo "Currently tracked repos: ${CURRENT_REPOS[*]}"
    fi

    all_tracked=true
    for d in "${ALL_DIRS[@]}"; do
        dir_name=$(basename "$d")
        found=false
        for repo in "${CURRENT_REPOS[@]}"; do
            if [[ "$repo" == "$dir_name" ]]; then
                found=true
                break
            fi
        done
        if [[ "$found" == false ]]; then
            all_tracked=false
            break
        fi
    done

    if [[ "$all_tracked" == true ]]; then
        echo "All sibling directories are already tracked in the current config."
        read -rp "Continue with reinitialization? [y/N] " answer
        if [[ "$answer" != "y" && "$answer" != "Y" ]]; then
            echo "Aborted."
            exit 0
        fi
    fi
fi

read -rp "Select repos (comma-separated numbers, or 'all'): " selection

# Parse selection
SELECTED_DIRS=()
if [[ "$selection" == "all" ]]; then
    SELECTED_DIRS=("${ALL_DIRS[@]}")
else
    IFS=',' read -ra indices <<< "$selection"
    for idx in "${indices[@]}"; do
        idx=$(echo "$idx" | tr -d ' ') # trim spaces
        if ! [[ "$idx" =~ ^[0-9]+$ ]]; then
            echo "Error: invalid selection: $idx" >&2
            exit 1
        fi
        if [[ $idx -lt 1 || $idx -gt ${#ALL_DIRS[@]} ]]; then
            echo "Error: selection out of range: $idx" >&2
            exit 1
        fi
        SELECTED_DIRS+=("${ALL_DIRS[$((idx-1))]}")
    done
fi

# Remove duplicates (Bash 3.2 compatible — no associative arrays)
UNIQUE_SELECTED=()
for d in "${SELECTED_DIRS[@]}"; do
    found=false
    for u in "${UNIQUE_SELECTED[@]}"; do
        if [[ "$u" == "$d" ]]; then
            found=true
            break
        fi
    done
    if [[ "$found" == false ]]; then
        UNIQUE_SELECTED+=("$d")
    fi
done
SELECTED_DIRS=("${UNIQUE_SELECTED[@]}")

if [[ ${#SELECTED_DIRS[@]} -eq 0 ]]; then
    echo "Error: must select at least one repo" >&2
    exit 1
fi

# ---------------------------------------------------------------------------
# 5. Ask which directory is the sidecar
# ---------------------------------------------------------------------------
if [[ -n "$SIDECAR_DIR_ARG" ]]; then
    # Expand ~ to $HOME
    SIDECAR_DIR_ARG="${SIDECAR_DIR_ARG/#\~/$HOME}"
    SIDECAR_DIR="$(cd "$SIDECAR_DIR_ARG" 2>/dev/null && pwd)" || {
        echo "Error: sidecar directory does not exist: $SIDECAR_DIR_ARG" >&2
        exit 1
    }
else
    echo ""
    echo "Which directory is the sidecar (specs/docs)?"
    for i in "${!ALL_DIRS[@]}"; do
        printf '  %d. %s\n' "$((i+1))" "$(basename "${ALL_DIRS[$i]}")"
    done
    read -rp "Enter number: " sidecar_idx
    if ! [[ "$sidecar_idx" =~ ^[0-9]+$ ]]; then
        echo "Error: invalid selection: $sidecar_idx" >&2
        exit 1
    fi
    if [[ $sidecar_idx -lt 1 || $sidecar_idx -gt ${#ALL_DIRS[@]} ]]; then
        echo "Error: selection out of range: $sidecar_idx" >&2
        exit 1
    fi
    SIDECAR_DIR="${ALL_DIRS[$((sidecar_idx-1))]}"
fi

# Validate: sidecar cannot also be a workspace repo
for d in "${SELECTED_DIRS[@]}"; do
    if [[ "$d" == "$SIDECAR_DIR" ]]; then
        echo "Error: sidecar cannot also be a workspace repo" >&2
        exit 1
    fi
done

# ---------------------------------------------------------------------------
# 6. Generate .sidecar.yml with yq
# ---------------------------------------------------------------------------
CONFIG_FILE="$PARENT_DIR/.sidecar.yml"

sidecar_name="$(basename "$SIDECAR_DIR")"

# Build YAML using yq v4 native syntax (mikefarah/yq does not support --arg/--argjson)
yq -n ".version = 1" > "$CONFIG_FILE"
yq -i ".sidecar = \"$sidecar_name\"" "$CONFIG_FILE"
yq -i '.settings.spec_dir = "specs/"' "$CONFIG_FILE"
yq -i '.settings.auto_index = true' "$CONFIG_FILE"
yq -i '.settings.index_strategy = "content"' "$CONFIG_FILE"

# Add repos array
yq -i '.repos = []' "$CONFIG_FILE"
for d in "${SELECTED_DIRS[@]}"; do
    yq -i ".repos += [\"$(basename "$d")\"]" "$CONFIG_FILE"
done

echo "Created $CONFIG_FILE"

# ---------------------------------------------------------------------------
# 6b. Validate generated YAML
# ---------------------------------------------------------------------------
validate_config() {
    local file="$1"
    local errors=0

    # version must be present and equal 1
    local version_type
    version_type=$(yq eval '.version | type' "$file")
    if [[ "$version_type" == "!!null" ]]; then
        echo "Error: missing required field: version" >&2
        errors=$((errors + 1))
    elif [[ "$version_type" != "!!int" ]]; then
        echo "Error: version must be an integer" >&2
        errors=$((errors + 1))
    else
        local version
        version=$(yq eval '.version' "$file")
        if [[ "$version" != "1" ]]; then
            echo "Error: version must be 1, got: $version" >&2
            errors=$((errors + 1))
        fi
    fi

    # sidecar must be present and a string
    local sidecar_type
    sidecar_type=$(yq eval '.sidecar | type' "$file")
    if [[ "$sidecar_type" == "!!null" ]]; then
        echo "Error: missing required field: sidecar" >&2
        errors=$((errors + 1))
    elif [[ "$sidecar_type" != "!!str" ]]; then
        echo "Error: sidecar must be a string" >&2
        errors=$((errors + 1))
    else
        local sidecar
        sidecar=$(yq eval '.sidecar' "$file")
        if [[ -z "$sidecar" ]]; then
            echo "Error: sidecar must be non-empty" >&2
            errors=$((errors + 1))
        fi
    fi

    # repos must be present and a non-empty array
    local repos_type
    repos_type=$(yq eval '.repos | type' "$file")
    if [[ "$repos_type" == "!!null" ]]; then
        echo "Error: missing required field: repos" >&2
        errors=$((errors + 1))
    elif [[ "$repos_type" != "!!seq" ]]; then
        echo "Error: repos must be an array" >&2
        errors=$((errors + 1))
    elif [[ "$(yq eval '.repos | length' "$file")" -eq 0 ]]; then
        echo "Error: repos must be non-empty" >&2
        errors=$((errors + 1))
    fi

    # settings must be present with required subfields
    local settings_type
    settings_type=$(yq eval '.settings | type' "$file")
    if [[ "$settings_type" == "!!null" ]]; then
        echo "Error: missing required field: settings" >&2
        errors=$((errors + 1))
    elif [[ "$settings_type" != "!!map" ]]; then
        echo "Error: settings must be an object" >&2
        errors=$((errors + 1))
    else
        for subfield in spec_dir auto_index index_strategy; do
            if [[ "$(yq eval ".settings.$subfield | type" "$file")" == "!!null" ]]; then
                echo "Error: missing required settings subfield: $subfield" >&2
                errors=$((errors + 1))
            fi
        done
    fi

    # reject unknown top-level fields
    local known_fields="version sidecar repos settings"
    local all_fields
    all_fields=$(yq eval 'keys | .[]' "$file" 2>/dev/null || true)
    while IFS= read -r field; do
        [[ -z "$field" ]] && continue
        if [[ " $known_fields " != *" $field "* ]]; then
            echo "Error: unknown top-level field: $field" >&2
            errors=$((errors + 1))
        fi
    done <<< "$all_fields"

    return $errors
}

if ! validate_config "$CONFIG_FILE"; then
    echo "Error: generated config failed validation" >&2
    exit 1
fi

# ---------------------------------------------------------------------------
# 7. Create <sidecar>/.sidecar/ directory
# ---------------------------------------------------------------------------
mkdir -p "$SIDECAR_DIR/.sidecar"

# ---------------------------------------------------------------------------
# 8. Run initial index
# ---------------------------------------------------------------------------
# Source _common.sh now that .sidecar.yml exists so discover_sidecar works
if [[ -f "$COMMON_SH" ]]; then
    # shellcheck source=_common.sh
    source "$COMMON_SH"
    discover_sidecar
    resolve_sidecar_dir
    resolve_spec_dir

    INDEX_SCRIPT="$SCRIPT_DIR/sidecar-index.sh"
    if [[ -x "$INDEX_SCRIPT" ]]; then
        echo "Running initial index..."
        "$INDEX_SCRIPT"
    else
        echo "Warning: sidecar-index.sh not found or not executable at $INDEX_SCRIPT" >&2
        echo "         Run it manually after setup." >&2
    fi
else
    echo "Warning: _common.sh not found at $COMMON_SH" >&2
fi

echo ""
echo "Sidecar workspace initialized!"
echo "  Config: $CONFIG_FILE"
echo "  Sidecar: $SIDECAR_DIR"
echo -n "  Repos: "
for d in "${SELECTED_DIRS[@]}"; do
    echo -n "$(basename "$d") "
done
echo ""
