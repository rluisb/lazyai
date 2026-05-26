#!/usr/bin/env bash
# task-init.sh — Scaffold a new speckit task with spec/tasks templates
# Usage: ./task-init.sh <slug> [title]

set -euo pipefail

SLUG="${1:-}"
TITLE="${2:-$SLUG}"

if [ -z "$SLUG" ]; then
    echo "Usage: ./task-init.sh <slug> [title]"
    echo "  Creates bee-gone/specs/<slug>/ with templates from bee-gone/templates/"
    exit 1
fi

SPEC_DIR="bee-gone/specs/${SLUG}"
TODAY=$(date +%Y-%m-%d)

if [ -d "$SPEC_DIR" ]; then
    echo "⚠️  $SPEC_DIR already exists"
    read -rp "Overwrite? [y/N] " yn
    case $yn in
        [Yy]*) rm -rf "$SPEC_DIR" ;;
        *) echo "Aborted."; exit 1 ;;
    esac
fi

mkdir -p "$SPEC_DIR"
mkdir -p "$SPEC_DIR/contracts"
mkdir -p "$SPEC_DIR/checklists"
mkdir -p "bee-gone/.specify/memory"

# Helper: copy template and substitute placeholders
copy_template() {
    src="$1"
    dest="$2"
    if [ -f "$src" ]; then
        sed -e "s/{{TODAY}}/$TODAY/g" \
            -e "s/{{SLUG}}/$SLUG/g" \
            -e "s/{{TITLE}}/$TITLE/g" \
            "$src" > "$dest"
    else
        echo "Warning: template $src not found, skipping $dest"
    fi
}

# Copy core templates
copy_template "bee-gone/templates/spec.md" "$SPEC_DIR/spec.md"
copy_template "bee-gone/templates/tasks.md" "$SPEC_DIR/tasks.md"
copy_template "bee-gone/templates/research.md" "$SPEC_DIR/research.md"
copy_template "bee-gone/templates/plan.md" "$SPEC_DIR/plan.md"

# Copy harness extension templates
copy_template "bee-gone/templates/decisions.md" "$SPEC_DIR/decisions.md"
copy_template "bee-gone/templates/validation.md" "$SPEC_DIR/validation.md"
copy_template "bee-gone/templates/handoff.md" "$SPEC_DIR/handoff.md"

# Copy optional templates if they exist
copy_template "bee-gone/templates/data-model.md" "$SPEC_DIR/data-model.md"
copy_template "bee-gone/templates/quickstart.md" "$SPEC_DIR/quickstart.md"

# Ensure .gitkeep in empty directories
touch "$SPEC_DIR/contracts/.gitkeep"
touch "$SPEC_DIR/checklists/.gitkeep"
touch "bee-gone/.specify/memory/.gitkeep"

echo ""
echo "✅ Task scaffolded: $SLUG"
echo ""
echo "Files created:"
for f in "$SPEC_DIR/spec.md" \
         "$SPEC_DIR/tasks.md" \
         "$SPEC_DIR/research.md" \
         "$SPEC_DIR/plan.md" \
         "$SPEC_DIR/decisions.md" \
         "$SPEC_DIR/validation.md" \
         "$SPEC_DIR/handoff.md" \
         "$SPEC_DIR/data-model.md" \
         "$SPEC_DIR/quickstart.md"; do
    if [ -f "$f" ]; then
        echo "  $f"
    fi
done
echo ""
echo "Directories created:"
echo "  $SPEC_DIR/contracts/"
echo "  $SPEC_DIR/checklists/"
echo "  bee-gone/.specify/memory/"
echo ""
echo "Next: run 'storm-scout' skill on this spec to clarify, research, and plan."
