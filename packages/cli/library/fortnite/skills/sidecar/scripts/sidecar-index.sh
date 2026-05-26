#!/usr/bin/env bash
# sidecar-index.sh — Rebuild the spec-to-repo SQLite index
# Spec: R2 (SQLite Index), R6 (Script Responsibilities)
#
# Scans all spec directories under $SPEC_DIR, extracts metadata,
# performs whole-word case-insensitive repo name matching with
# confidence scoring, and populates the spec_repo_links table.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/_common.sh"

# ---------------------------------------------------------------------------
# CLI flags
# ---------------------------------------------------------------------------
FORCE=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --force)
            FORCE=true
            shift
            ;;
        --help|-h)
            cat <<'USAGE'
Usage: sidecar-index.sh [--force]

Rebuild the spec-to-repo SQLite index.

Options:
  --force    Rebuild even if index is fresh
  --help     Show this help message
USAGE
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            echo "Usage: sidecar-index.sh [--force]" >&2
            exit 1
            ;;
    esac
done

# ---------------------------------------------------------------------------
# Discovery
# ---------------------------------------------------------------------------
discover_sidecar
resolve_sidecar_dir

# Build spec dir path manually; if missing, exit gracefully
spec_dir_setting="$(yq eval '.settings.spec_dir // "specs/"' "$SIDECAR_YML")"
SPEC_DIR="$SIDECAR_DIR/$spec_dir_setting"

if [[ ! -d "$SPEC_DIR" ]]; then
    echo "Indexed 0 specs, found 0 links across 0 repos"
    exit 0
fi

INDEX_DB="$(get_index_db)"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

# Cross-platform mtime in seconds
_get_mtime() {
    local file="$1"
    if [[ "$(uname)" == "Darwin" ]]; then
        stat -f %m "$file" 2>/dev/null || echo 0
    else
        stat -c %Y "$file" 2>/dev/null || echo 0
    fi
}

# Cross-platform ISO-8601 to epoch seconds
_iso_to_epoch() {
    local ts="$1"
    if [[ "$(uname)" == "Darwin" ]]; then
        date -u -j -f "%Y-%m-%dT%H:%M:%SZ" "$ts" +%s 2>/dev/null || echo 0
    else
        date -d "$ts" +%s 2>/dev/null || echo 0
    fi
}

# Escape single quotes for SQLite string literals
_escape_sql() {
    echo "${1//\'/\'\'}"
}

# Escape ERE metacharacters so a repo name can be embedded safely in grep -E
_escape_regex() {
    sed 's/[][\\^$.*+?{}|()/]/\\&/g' <<< "$1"
}

# Return the path value as stored in .sidecar.yml (relative).
# For scalar entries, returns the repo name itself.
# For object entries, returns the .path field.
_get_repo_yaml_path() {
    local repo_name="$1"

    # Object format
    local obj_path
    obj_path="$(yq eval ".repos[] | select(.name == \"$repo_name\") | .path" "$SIDECAR_YML" 2>/dev/null || true)"
    if [[ -n "$obj_path" && "$obj_path" != "null" ]]; then
        echo "$obj_path"
        return 0
    fi

    # Scalar entry
    echo "$repo_name"
}

# Return the absolute filesystem path for a repo name.
# Tries object format in .sidecar.yml first, then sibling directories.
_resolve_repo_path() {
    local repo_name="$1"

    # Object format: repos: [{name: "x", path: "../path/to/x"}, ...]
    local obj_path
    obj_path="$(yq eval ".repos[] | select(.name == \"$repo_name\") | .path" "$SIDECAR_YML" 2>/dev/null || true)"
    if [[ -n "$obj_path" && "$obj_path" != "null" ]]; then
        # Expand ~ to $HOME
        obj_path="${obj_path/#\~/$HOME}"
        # Resolve relative to SIDECAR_ROOT
        if [[ ! "$obj_path" == /* ]]; then
            obj_path="$SIDECAR_ROOT/$obj_path"
        fi
        if [[ -d "$obj_path" ]]; then
            echo "$(cd "$obj_path" && pwd)"
        else
            echo "$obj_path"
        fi
        return 0
    fi

    # Relative to the directory that holds .sidecar.yml
    if [[ -d "$SIDECAR_ROOT/$repo_name" ]]; then
        echo "$(cd "$SIDECAR_ROOT/$repo_name" && pwd)"
        return 0
    fi

    # Fallback
    echo "$repo_name"
}

# Whole-word case-insensitive match.
# Word characters = [a-zA-Z0-9-] so that "fedora" does NOT match inside "fedora-iac".
_repo_in_line() {
    local repo="$1"
    local line="$2"
    local escaped_repo
    escaped_repo="$(_escape_regex "$repo")"
    grep -qiE "(^|[^a-zA-Z0-9-])${escaped_repo}([^a-zA-Z0-9-]|$)" <<< "$line"
}

# ---------------------------------------------------------------------------
# Corruption check & freshness check (skip unless --force)
# ---------------------------------------------------------------------------
needs_rebuild=false

if [[ -f "$INDEX_DB" ]]; then
    # Always probe for corruption before any query
    if ! sqlite3 "$INDEX_DB" "SELECT 1 FROM sqlite_master LIMIT 1;" >/dev/null 2>&1; then
        echo "Warning: corrupt index.db detected, recreating..." >&2
        rm -f "$INDEX_DB"
        needs_rebuild=true
    elif [[ "$FORCE" == false ]]; then
        # Not corrupt — proceed with freshness check
        index_mtime="$(_get_mtime "$INDEX_DB")"

        newest_spec_mtime=0
        while IFS= read -r -d '' spec_file; do
            spec_mtime="$(_get_mtime "$spec_file")"
            if [[ "$spec_mtime" -gt "$newest_spec_mtime" ]]; then
                newest_spec_mtime="$spec_mtime"
            fi
        done < <(find "$SPEC_DIR" -type f -print0 2>/dev/null || true)

        if [[ "$newest_spec_mtime" -gt 0 && "$index_mtime" -ge "$newest_spec_mtime" ]]; then
            # Also check if .sidecar.yml is newer than the last index
            sidecar_mtime="$(_get_mtime "$SIDECAR_YML")"
            last_indexed_at="$(sqlite3 "$INDEX_DB" "SELECT MAX(indexed_at) FROM specs;" 2>/dev/null || true)"
            if [[ -n "$last_indexed_at" && "$last_indexed_at" != "null" ]]; then
                last_indexed_epoch="$(_iso_to_epoch "$last_indexed_at")"
                if [[ "$sidecar_mtime" -le "$last_indexed_epoch" ]]; then
                    echo "Index is up to date. Use --force to rebuild."
                    exit 0
                fi
            fi
        fi
    fi
else
    needs_rebuild=true
fi

# ---------------------------------------------------------------------------
# Initialise schema (idempotent)
# ---------------------------------------------------------------------------

sqlite3 "$INDEX_DB" <<'SQL'
CREATE TABLE IF NOT EXISTS repos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    path TEXT NOT NULL,
    added_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS specs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    path TEXT NOT NULL,
    title TEXT,
    indexed_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS spec_repo_links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    spec_id INTEGER NOT NULL REFERENCES specs(id) ON DELETE CASCADE,
    repo_id INTEGER NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    confidence REAL DEFAULT 1.0,
    match_source TEXT,
    UNIQUE(spec_id, repo_id)
);

CREATE INDEX IF NOT EXISTS idx_spec_repo_links_repo ON spec_repo_links(repo_id);
CREATE INDEX IF NOT EXISTS idx_spec_repo_links_spec ON spec_repo_links(spec_id);
CREATE INDEX IF NOT EXISTS idx_specs_slug ON specs(slug);
SQL

# ---------------------------------------------------------------------------
# Clear links for full rebuild
# ---------------------------------------------------------------------------
sqlite3 "$INDEX_DB" "DELETE FROM spec_repo_links;"

# ---------------------------------------------------------------------------
# Populate repos table
# ---------------------------------------------------------------------------
REPOS=()
while IFS= read -r repo; do
    REPOS+=("$repo")
done < <(get_repos)

for repo in "${REPOS[@]}"; do
    repo_path="$(_resolve_repo_path "$repo")"
    yaml_path="$(_get_repo_yaml_path "$repo")"

    # Defensive: reject absolute paths in yaml_path (should have been caught by validation)
    if [[ "$yaml_path" == /* || "$yaml_path" == ~* ]]; then
        echo "Error: repos path for '$repo' is absolute in .sidecar.yml: $yaml_path" >&2
        exit 1
    fi

    escaped_repo="$(_escape_sql "$repo")"
    escaped_path="$(_escape_sql "$yaml_path")"
    added_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

    sqlite3 "$INDEX_DB" <<SQL
INSERT INTO repos (name, path, added_at)
VALUES ('$escaped_repo', '$escaped_path', '$added_at')
ON CONFLICT(name) DO UPDATE SET
    path = excluded.path,
    added_at = excluded.added_at;
SQL
done

# ---------------------------------------------------------------------------
# Bash 3.2 compatible helpers for per-spec repo confidence tracking
# (replaces associative arrays unavailable in Bash 3.2)
# ---------------------------------------------------------------------------
_get_repo_index() {
    local target="$1"
    local i
    for i in "${!spec_repo_list[@]}"; do
        if [[ "${spec_repo_list[$i]}" == "$target" ]]; then
            echo "$i"
            return 0
        fi
    done
    echo "-1"
}

_set_repo_conf() {
    local repo="$1"
    local conf="$2"
    local src="$3"
    local idx
    idx="$(_get_repo_index "$repo")"
    if [[ "$idx" == "-1" ]]; then
        spec_repo_list+=("$repo")
        spec_repo_conf+=("$conf")
        spec_repo_src+=("$src")
    else
        spec_repo_conf[$idx]="$conf"
        spec_repo_src[$idx]="$src"
    fi
}

_get_repo_conf() {
    local repo="$1"
    local idx
    idx="$(_get_repo_index "$repo")"
    if [[ "$idx" == "-1" ]]; then
        echo "0"
    else
        echo "${spec_repo_conf[$idx]}"
    fi
}

_get_repo_src() {
    local repo="$1"
    local idx
    idx="$(_get_repo_index "$repo")"
    if [[ "$idx" == "-1" ]]; then
        echo ""
    else
        echo "${spec_repo_src[$idx]}"
    fi
}

# ---------------------------------------------------------------------------
# Index specs
# ---------------------------------------------------------------------------
spec_count=0
link_count=0

while IFS= read -r -d '' spec_dir; do
    slug="$(basename "$spec_dir")"
    spec_path="${spec_dir#$SIDECAR_DIR/}"

    # Extract title from first '# ' heading in spec.md
    title=""
    if [[ -f "$spec_dir/spec.md" ]]; then
        title="$(grep -m1 '^# ' "$spec_dir/spec.md" 2>/dev/null | sed 's/^# //' || true)"
    fi

    escaped_slug="$(_escape_sql "$slug")"
    escaped_path="$(_escape_sql "$spec_path")"
    escaped_title="$(_escape_sql "$title")"
    indexed_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

    # Insert or update spec record
    sqlite3 "$INDEX_DB" <<SQL
INSERT INTO specs (slug, path, title, indexed_at)
VALUES ('$escaped_slug', '$escaped_path', '$escaped_title', '$indexed_at')
ON CONFLICT(slug) DO UPDATE SET
    path = excluded.path,
    title = excluded.title,
    indexed_at = excluded.indexed_at;
SQL

    spec_count=$((spec_count + 1))

    # Retrieve the spec_id
    spec_id="$(sqlite3 "$INDEX_DB" "SELECT id FROM specs WHERE slug = '$escaped_slug';")"

    # Per-spec tracking: highest confidence and source per repo
    # Bash 3.2 compatible parallel indexed arrays (no associative arrays)
    spec_repo_list=()
    spec_repo_conf=()
    spec_repo_src=()

    # 1.0 — repo name appears in spec slug / directory name
    for repo in "${REPOS[@]}"; do
        if _repo_in_line "$repo" "$slug"; then
            _set_repo_conf "$repo" "1.0" "filename"
        fi
    done

    # 1.0 — repo name appears in spec title (exact match, same as slug)
    for repo in "${REPOS[@]}"; do
        # Already at max confidence from slug match
        [[ "$(_get_repo_conf "$repo")" == "1.0" ]] && continue

        if _repo_in_line "$repo" "$title"; then
            _set_repo_conf "$repo" "1.0" "title"
        fi
    done

    # Scan content files for repo mentions (file-level grep for performance)
    for repo in "${REPOS[@]}"; do
        # Already at max confidence from slug match
        [[ "$(_get_repo_conf "$repo")" == "1.0" ]] && continue

        escaped_repo="$(_escape_regex "$repo")"
        pattern="(^|[^a-zA-Z0-9-])${escaped_repo}([^a-zA-Z0-9-]|$)"

        for content_file in spec.md plan.md research.md; do
            file_path="$spec_dir/$content_file"
            [[ -f "$file_path" ]] || continue

            # Check headings first (0.8)
            if [[ "$(_get_repo_conf "$repo")" != "0.8" ]]; then
                if grep -qiE "$pattern" <(grep '^#' "$file_path" 2>/dev/null); then
                    _set_repo_conf "$repo" "0.8" "heading"
                    continue
                fi
            fi

            # Check full file for body mention (0.5)
            if [[ "$(_get_repo_conf "$repo")" == "0" ]]; then
                if grep -qiE "$pattern" "$file_path" 2>/dev/null; then
                    _set_repo_conf "$repo" "0.5" "content"
                fi
            fi
        done
    done

    # Insert links for this spec
    for repo in "${spec_repo_list[@]}"; do
        confidence="$(_get_repo_conf "$repo")"
        match_source="$(_get_repo_src "$repo")"
        escaped_repo="$(_escape_sql "$repo")"

        repo_id="$(sqlite3 "$INDEX_DB" "SELECT id FROM repos WHERE name = '$escaped_repo';")"
        if [[ -n "$repo_id" ]]; then
            sqlite3 "$INDEX_DB" <<SQL
INSERT INTO spec_repo_links (spec_id, repo_id, confidence, match_source)
VALUES ($spec_id, $repo_id, $confidence, '$match_source')
ON CONFLICT(spec_id, repo_id) DO UPDATE SET
    confidence = excluded.confidence,
    match_source = excluded.match_source;
SQL
            link_count=$((link_count + 1))
        fi
    done
done < <(find "$SPEC_DIR" -mindepth 1 -maxdepth 1 -type d -print0 2>/dev/null || true)

# ---------------------------------------------------------------------------
# Summary — derive repo count from SQLite so it is always correct
# ---------------------------------------------------------------------------
repo_link_count=0
if [[ "$link_count" -gt 0 ]]; then
    repo_link_count="$(sqlite3 "$INDEX_DB" "SELECT COUNT(DISTINCT repo_id) FROM spec_repo_links;" 2>/dev/null || true)"
    [[ -n "$repo_link_count" ]] || repo_link_count=0
fi
echo "Indexed $spec_count specs, found $link_count links across $repo_link_count repos"
