#!/usr/bin/env bash
# Shared discovery helper for sidecar skill
# Sourced by all sidecar scripts. Walks up from CWD to find .sidecar.yml,
# parses it, and exports variables.
#
# Usage: source "$(dirname "$0")/_common.sh"
# Requires: yq, sqlite3, bash 3.2+

set -euo pipefail

# ---------------------------------------------------------------------------
# validate_sidecar_config
# Validates the loaded .sidecar.yml against the schema.
# Exits with error if validation fails.
# ---------------------------------------------------------------------------
validate_sidecar_config() {
    if [[ -z "${SIDECAR_YML:-}" ]]; then
        echo "Error: SIDECAR_YML not set. Call discover_sidecar first." >&2
        exit 1
    fi

    local errors=0

    # version must be numeric and equal to 1
    local version version_type
    version="$(yq eval '.version' "$SIDECAR_YML")"
    version_type="$(yq eval '.version | type' "$SIDECAR_YML")"
    if [[ "$version_type" == "!!null" ]]; then
        echo "Error: missing required field: version" >&2
        errors=$((errors + 1))
    elif [[ "$version_type" != "!!int" ]]; then
        echo "Error: version must be numeric" >&2
        errors=$((errors + 1))
    elif [[ "$version" != "1" ]]; then
        echo "Error: version must be 1, got: $version" >&2
        errors=$((errors + 1))
    fi

    # sidecar must be a non-empty relative string
    local sidecar sidecar_type
    sidecar="$(yq eval '.sidecar' "$SIDECAR_YML")"
    sidecar_type="$(yq eval '.sidecar | type' "$SIDECAR_YML")"
    if [[ "$sidecar_type" != "!!str" ]]; then
        echo "Error: 'sidecar' must be a string" >&2
        errors=$((errors + 1))
    elif [[ -z "$sidecar" || "$sidecar" == "null" ]]; then
        echo "Error: 'sidecar' must be a non-empty string" >&2
        errors=$((errors + 1))
    elif [[ "$sidecar" == /* || "$sidecar" == ~* ]]; then
        echo "Error: 'sidecar' must be a relative path, got: $sidecar" >&2
        errors=$((errors + 1))
    fi

    # repos must be a non-empty array
    local repos_type
    repos_type="$(yq eval '.repos | type' "$SIDECAR_YML")"
    if [[ "$repos_type" == "!!null" ]]; then
        echo "Error: missing required field: repos" >&2
        errors=$((errors + 1))
    elif [[ "$repos_type" != "!!seq" ]]; then
        echo "Error: repos must be an array" >&2
        errors=$((errors + 1))
    elif [[ "$(yq eval '.repos | length' "$SIDECAR_YML")" -eq 0 ]]; then
        echo "Error: repos must be non-empty" >&2
        errors=$((errors + 1))
    else
        # Validate each repos[] entry is either !!str or !!map (with required name field)
        local repo_count
        repo_count="$(yq eval '.repos | length' "$SIDECAR_YML")"
        local i=0
        while [[ "$i" -lt "$repo_count" ]]; do
            local entry_type
            entry_type="$(yq eval ".repos[$i] | type" "$SIDECAR_YML")"
            if [[ "$entry_type" == "!!str" ]]; then
                : # valid scalar
            elif [[ "$entry_type" == "!!map" ]]; then
                # Reject unknown keys in repo objects
                local entry_keys
                entry_keys="$(yq eval ".repos[$i] | keys | .[]" "$SIDECAR_YML" 2>/dev/null || true)"
                for key in $entry_keys; do
                    case "$key" in
                        name|path) ;;
                        *) echo "Error: unknown repos[$i] field: $key" >&2; errors=$((errors + 1)) ;;
                    esac
                done

                local entry_name
                entry_name="$(yq eval ".repos[$i].name" "$SIDECAR_YML")"
                local entry_name_type
                entry_name_type="$(yq eval ".repos[$i].name | type" "$SIDECAR_YML")"
                if [[ "$entry_name_type" != "!!str" || -z "$entry_name" || "$entry_name" == "null" ]]; then
                    echo "Error: repos[$i] object must have a non-empty string 'name' field" >&2
                    errors=$((errors + 1))
                fi
                local entry_path_type entry_path
                entry_path_type="$(yq eval ".repos[$i].path | type" "$SIDECAR_YML")"
                if [[ "$entry_path_type" == "!!null" ]]; then
                    echo "Error: repos[$i] object missing required 'path' field" >&2
                    errors=$((errors + 1))
                elif [[ "$entry_path_type" != "!!str" ]]; then
                    echo "Error: repos[$i].path must be a string" >&2
                    errors=$((errors + 1))
                else
                    entry_path="$(yq eval ".repos[$i].path" "$SIDECAR_YML")"
                    if [[ -z "$entry_path" || "$entry_path" == "null" ]]; then
                        echo "Error: repos[$i].path must be a non-empty string" >&2
                        errors=$((errors + 1))
                    elif [[ "$entry_path" == /* || "$entry_path" == ~* ]]; then
                        echo "Error: repos[$i].path must be relative to SIDECAR_ROOT, got: $entry_path" >&2
                        errors=$((errors + 1))
                    fi
                fi
            else
                echo "Error: repos[$i] must be a string or an object with 'name' field" >&2
                errors=$((errors + 1))
            fi
            i=$((i + 1))
        done
    fi

    # settings — apply defaults for missing values, validate types for present ones
    local settings_type
    settings_type="$(yq eval '.settings | type' "$SIDECAR_YML")"
    if [[ "$settings_type" != "!!null" && "$settings_type" != "!!map" ]]; then
        echo "Error: settings must be an object" >&2
        errors=$((errors + 1))
    else
        # spec_dir — default to "specs/" if missing, must be relative
        local spec_dir spec_dir_type
        spec_dir="$(yq eval '.settings.spec_dir' "$SIDECAR_YML")"
        spec_dir_type="$(yq eval '.settings.spec_dir | type' "$SIDECAR_YML")"
        if [[ "$spec_dir_type" != "!!null" && -n "$spec_dir" ]]; then
            if [[ "$spec_dir_type" != "!!str" ]]; then
                echo "Error: settings.spec_dir must be a string" >&2
                errors=$((errors + 1))
            elif [[ "$spec_dir" == /* || "$spec_dir" == ~* ]]; then
                echo "Error: settings.spec_dir must be a relative path, got: $spec_dir" >&2
                errors=$((errors + 1))
            fi
        fi

        # auto_index — default to true if missing
        local auto_index_type
        auto_index_type="$(yq eval '.settings.auto_index | type' "$SIDECAR_YML")"
        if [[ "$auto_index_type" != "!!null" && "$auto_index_type" != "!!bool" ]]; then
            echo "Error: settings.auto_index must be boolean" >&2
            errors=$((errors + 1))
        fi

        # index_strategy — default to "content" if missing
        local index_strategy index_strategy_type
        index_strategy="$(yq eval '.settings.index_strategy' "$SIDECAR_YML")"
        index_strategy_type="$(yq eval '.settings.index_strategy | type' "$SIDECAR_YML")"
        if [[ "$index_strategy_type" != "!!null" && -n "$index_strategy" && "$index_strategy" != "content" ]]; then
            echo "Error: settings.index_strategy must be 'content'" >&2
            errors=$((errors + 1))
        fi

        # reject unknown settings fields (only if settings is present)
        if [[ "$settings_type" != "!!null" ]]; then
            local known_settings="spec_dir auto_index index_strategy"
            local settings_keys
            settings_keys="$(yq eval '.settings | keys | .[]' "$SIDECAR_YML" 2>/dev/null || true)"
            for key in $settings_keys; do
                if ! echo "$known_settings" | grep -qw "$key"; then
                    echo "Error: unknown settings field: $key" >&2
                    errors=$((errors + 1))
                fi
            done
        fi
    fi

    # reject unknown top-level fields
    local known_fields="version sidecar repos settings"
    local all_fields
    all_fields="$(yq eval 'keys | .[]' "$SIDECAR_YML" 2>/dev/null || true)"
    while IFS= read -r field; do
        [[ -z "$field" ]] && continue
        if [[ " $known_fields " != *" $field "* ]]; then
            echo "Error: unknown top-level field: $field" >&2
            errors=$((errors + 1))
        fi
    done <<< "$all_fields"

    if [[ "$errors" -gt 0 ]]; then
        echo "Error: $errors validation error(s) in $SIDECAR_YML" >&2
        exit 1
    fi
}

# ---------------------------------------------------------------------------
# discover_sidecar
# Walk up from CWD looking for .sidecar.yml.
# Exports: SIDECAR_ROOT, SIDECAR_YML
# ---------------------------------------------------------------------------
discover_sidecar() {
    local dir="$PWD"

    while [[ "$dir" != "/" ]]; do
        if [[ -f "$dir/.sidecar.yml" ]]; then
            SIDECAR_ROOT="$(cd "$dir" && pwd)"
            SIDECAR_YML="$SIDECAR_ROOT/.sidecar.yml"
            export SIDECAR_ROOT SIDECAR_YML
            validate_sidecar_config
            return 0
        fi
        dir="$(dirname "$dir")"
    done

    # Check root directory as well (dirname of / is /)
    if [[ -f "/.sidecar.yml" ]]; then
        SIDECAR_ROOT="/"
        SIDECAR_YML="/.sidecar.yml"
        export SIDECAR_ROOT SIDECAR_YML
        validate_sidecar_config
        return 0
    fi

    echo "No .sidecar.yml found. Run 'sidecar init' first." >&2
    exit 1
}

# ---------------------------------------------------------------------------
# resolve_sidecar_dir
# Reads 'sidecar:' field from .sidecar.yml.
# Exports: SIDECAR_DIR (absolute path to sidecar directory)
# ---------------------------------------------------------------------------
resolve_sidecar_dir() {
    if [[ -z "${SIDECAR_YML:-}" ]]; then
        discover_sidecar
    fi

    local raw_dir
    raw_dir="$(yq eval '.sidecar' "$SIDECAR_YML")"

    if [[ -z "$raw_dir" || "$raw_dir" == "null" ]]; then
        echo "Error: 'sidecar' field not found in $SIDECAR_YML" >&2
        exit 1
    fi

    # Expand ~ to $HOME
    raw_dir="${raw_dir/#\~/$HOME}"

    # Resolve relative to SIDECAR_ROOT
    SIDECAR_DIR="$(cd "$SIDECAR_ROOT/$raw_dir" 2>/dev/null && pwd)" || {
        echo "Error: sidecar directory does not exist: $SIDECAR_ROOT/$raw_dir" >&2
        exit 1
    }

    export SIDECAR_DIR
}

# ---------------------------------------------------------------------------
# resolve_spec_dir
# Reads 'settings.spec_dir' from .sidecar.yml (default: specs/).
# Exports: SPEC_DIR (absolute path)
# ---------------------------------------------------------------------------
resolve_spec_dir() {
    if [[ -z "${SIDECAR_DIR:-}" ]]; then
        resolve_sidecar_dir
    fi

    local spec_dir
    spec_dir="$(yq eval '.settings.spec_dir // "specs/"' "$SIDECAR_YML")"

    # Remove trailing slash for consistency
    spec_dir="${spec_dir%/}"

    SPEC_DIR="$SIDECAR_DIR/$spec_dir"

    # Resolve to absolute path
    SPEC_DIR="$(cd "$SPEC_DIR" 2>/dev/null && pwd)" || {
        echo "Error: spec directory does not exist: $SPEC_DIR" >&2
        exit 1
    }

    export SPEC_DIR
}

# ---------------------------------------------------------------------------
# get_repos
# Reads 'repos:' array from .sidecar.yml.
# Returns: newline-separated list of repo names
# ---------------------------------------------------------------------------
get_repos() {
    if [[ -z "${SIDECAR_YML:-}" ]]; then
        discover_sidecar
    fi

    yq eval '.repos[] | select(kind == "scalar") // .name' "$SIDECAR_YML"
}

# ---------------------------------------------------------------------------
# get_index_db
# Returns path to SQLite index. Creates .sidecar/ directory if needed.
# ---------------------------------------------------------------------------
get_index_db() {
    if [[ -z "${SIDECAR_DIR:-}" ]]; then
        resolve_sidecar_dir
    fi

    local index_dir="$SIDECAR_DIR/.sidecar"

    if [[ ! -d "$index_dir" ]]; then
        mkdir -p "$index_dir"
    fi

    echo "$index_dir/index.db"
}
