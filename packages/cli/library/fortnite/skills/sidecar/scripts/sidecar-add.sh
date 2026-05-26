#!/usr/bin/env bash
# sidecar-add.sh — Add a repo to the sidecar workspace
#
# Usage:
#   sidecar-add.sh <repo>              Add a repo by name
#   sidecar-add.sh --interactive       Interactive multi-select of unregistered siblings
#   sidecar-add.sh --path <path>       Add repo at explicit path
#   sidecar-add.sh --help              Show usage
#
# Conventions: set -euo pipefail, source _common.sh, use yq for YAML manipulation.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/_common.sh"

# ---------------------------------------------------------------------------
# Usage
# ---------------------------------------------------------------------------
usage() {
    cat <<'EOF'
Usage: sidecar-add.sh [OPTIONS] [<repo>]

Add a repo to the sidecar workspace.

Arguments:
  <repo>              Directory name relative to SIDECAR_ROOT

Options:
  --interactive       Interactive multi-select of unregistered sibling dirs
  --path <path>       Explicit path (for repos not in parent dir)
  --help              Show this help message

Examples:
  sidecar-add.sh fedora
  sidecar-add.sh --interactive
EOF
}

# ---------------------------------------------------------------------------
# Parse arguments
# ---------------------------------------------------------------------------
INTERACTIVE=false
EXPLICIT_PATH=""
REPO=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --interactive)
            INTERACTIVE=true
            shift
            ;;
        --path)
            if [[ $# -lt 2 ]]; then
                echo "Error: --path requires an argument" >&2
                usage >&2
                exit 1
            fi
            EXPLICIT_PATH="$2"
            shift 2
            ;;
        --help|-h)
            usage
            exit 0
            ;;
        --*)
            echo "Error: Unknown option $1" >&2
            usage >&2
            exit 1
            ;;
        *)
            if [[ -n "$REPO" ]]; then
                echo "Error: Only one repo name allowed" >&2
                usage >&2
                exit 1
            fi
            REPO="$1"
            shift
            ;;
    esac
done

# ---------------------------------------------------------------------------
# Discover workspace
# ---------------------------------------------------------------------------
discover_sidecar

# ---------------------------------------------------------------------------
# Helper: compute a path relative to a base directory
# Tries python3/python os.path.relpath first, then realpath --relative-to,
# then falls back to returning the target unchanged.
# ---------------------------------------------------------------------------
_relpath() {
    local target="$1"
    local base="$2"

    if command -v python3 &>/dev/null; then
        python3 -c "import os.path; print(os.path.relpath('$target', '$base'))"
        return 0
    elif command -v python &>/dev/null; then
        python -c "import os.path; print(os.path.relpath('$target', '$base'))"
        return 0
    fi

    if realpath --relative-to="$base" "$target" 2>/dev/null; then
        return 0
    fi

    echo "$target"
}

# ---------------------------------------------------------------------------
# Helper: list sibling directories not already in repos array
# ---------------------------------------------------------------------------
list_unregistered_siblings() {
    local repos
    repos="$(get_repos | sort -u)" || true

    for dir in "$SIDECAR_ROOT"/*/; do
        [[ -d "$dir" ]] || continue
        local basename
        basename="$(basename "$dir")"
        # Skip hidden dirs and the sidecar dir itself
        [[ "$basename" == .* ]] && continue
        [[ "$basename" == ".sidecar" ]] && continue
        # Skip if already in repos
        if echo "$repos" | grep -qx "$basename"; then
            continue
        fi
        echo "$basename"
    done
}

# ---------------------------------------------------------------------------
# Interactive mode
# ---------------------------------------------------------------------------
if [[ "$INTERACTIVE" == true ]]; then
    # Check for fzf
    if ! command -v fzf &>/dev/null; then
        echo "Error: fzf is required for interactive mode. Install it first." >&2
        exit 1
    fi

    choices=()
    while IFS= read -r line; do
        choices+=("$line")
    done < <(list_unregistered_siblings)

    if [[ ${#choices[@]} -eq 0 ]]; then
        echo "No unregistered sibling directories found."
        exit 0
    fi

    # Multi-select with fzf
    selected=()
    while IFS= read -r line; do
        selected+=("$line")
    done < <(printf '%s\n' "${choices[@]}" | fzf --multi --prompt="Select repos to add> ")

    if [[ ${#selected[@]} -eq 0 ]]; then
        echo "No repos selected."
        exit 0
    fi

    for repo in "${selected[@]}"; do
        # Validate directory exists
        if [[ ! -d "$SIDECAR_ROOT/$repo" ]]; then
            echo "Warning: directory does not exist: $SIDECAR_ROOT/$repo — skipping" >&2
            continue
        fi

        # Check if already in config (race condition / stale list)
        if get_repos | grep -qx "$repo"; then
            echo "$repo is already in the workspace"
            continue
        fi

        # Append to repos array
        yq eval -i ".repos += [\"$repo\"]" "$SIDECAR_YML"
        echo "Added $repo to workspace"
    done

    # Trigger re-index
    if [[ -x "$SCRIPT_DIR/sidecar-index.sh" ]]; then
        "$SCRIPT_DIR/sidecar-index.sh"
    else
        echo "Warning: sidecar-index.sh not found or not executable" >&2
    fi

    exit 0
fi

# ---------------------------------------------------------------------------
# Non-interactive mode
# ---------------------------------------------------------------------------

# Validate repo argument
if [[ -z "$REPO" && -z "$EXPLICIT_PATH" ]]; then
    echo "Error: No repo specified" >&2
    usage >&2
    exit 1
fi

# Handle explicit path
if [[ -n "$EXPLICIT_PATH" ]]; then
    if [[ -z "$REPO" ]]; then
        echo "Error: repo name is required when using --path" >&2
        usage >&2
        exit 1
    fi

    # Validate path exists and is a directory
    if [[ ! -e "$EXPLICIT_PATH" ]]; then
        echo "Error: path does not exist: $EXPLICIT_PATH" >&2
        exit 1
    fi
    if [[ ! -d "$EXPLICIT_PATH" ]]; then
        echo "Error: path is not a directory: $EXPLICIT_PATH" >&2
        exit 1
    fi

    # Resolve to absolute path for validation
    abs_path="$(cd "$EXPLICIT_PATH" && pwd)"

    # Check if already in config
    if get_repos | grep -qx "$REPO"; then
        echo "$REPO is already in the workspace"
        exit 0
    fi

    # Compute relative path from SIDECAR_ROOT for storage
    rel_path="$(_relpath "$abs_path" "$SIDECAR_ROOT")"

    # Append as object entry
    yq eval -i ".repos += [{\"name\": \"$REPO\", \"path\": \"$rel_path\"}]" "$SIDECAR_YML"
    echo "Added $REPO to workspace (path: $rel_path)"

    # Trigger re-index
    if [[ -x "$SCRIPT_DIR/sidecar-index.sh" ]]; then
        "$SCRIPT_DIR/sidecar-index.sh"
    else
        echo "Warning: sidecar-index.sh not found or not executable" >&2
    fi

    exit 0
fi

# Validate directory exists
if [[ ! -d "$SIDECAR_ROOT/$REPO" ]]; then
    echo "Error: directory does not exist: $SIDECAR_ROOT/$REPO" >&2
    echo "Hint: Run 'sidecar-add.sh --interactive' to see available repos" >&2
    exit 1
fi

# Check if already in config
if get_repos | grep -qx "$REPO"; then
    echo "$REPO is already in the workspace"
    exit 0
fi

# Append to repos array
yq eval -i ".repos += [\"$REPO\"]" "$SIDECAR_YML"
echo "Added $REPO to workspace"

# Trigger re-index
if [[ -x "$SCRIPT_DIR/sidecar-index.sh" ]]; then
    "$SCRIPT_DIR/sidecar-index.sh"
else
    echo "Warning: sidecar-index.sh not found or not executable" >&2
fi
