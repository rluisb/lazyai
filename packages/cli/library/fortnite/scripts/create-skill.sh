#!/usr/bin/env bash
# create-skill.sh — Scaffold a new skill with proper structure
# Usage: ./create-skill.sh <fortnite-name> "<description>" [--scripts script1.sh script2.sh]
#
# Example:
#   ./create-skill.sh medkit "Environment health check" --scripts health-check.sh cleanup.sh
#   ./create-skill.sh pickaxe "Dev tool reference"

set -euo pipefail

SKILL_NAME="${1:-}"
DESCRIPTION="${2:-}"
SCRIPTS=()

shift 2 || true
while [[ $# -gt 0 ]]; do
    case "$1" in
        --scripts) shift; while [[ $# -gt 0 && ! "$1" =~ ^-- ]]; do SCRIPTS+=("$1"); shift; done ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

if [[ -z "$SKILL_NAME" || -z "$DESCRIPTION" ]]; then
    echo "Usage: $0 <fortnite-name> \"<description>\" [--scripts script1.sh script2.sh]"
    echo ""
    echo "Example:"
    echo "  $0 medkit \"Environment health check\" --scripts health-check.sh cleanup.sh"
    exit 1
fi

SKILL_DIR="skills/$SKILL_NAME"
SCRIPTS_DIR="$SKILL_DIR/scripts"

# Check if skill already exists
if [[ -d "$SKILL_DIR" ]]; then
    echo "❌ Skill already exists: $SKILL_DIR"
    exit 1
fi

echo "🏗️  Creating skill: $SKILL_NAME"
echo "   Description: $DESCRIPTION"
echo "   Trigger: /$SKILL_NAME"
echo ""

# Create directories
mkdir -p "$SCRIPTS_DIR"

# Create SKILL.md from template
cat > "$SKILL_DIR/SKILL.md" << EOF
---
name: $SKILL_NAME
description: "$DESCRIPTION"
trigger: /$SKILL_NAME
skill_path: skills/$SKILL_NAME
scripts:
EOF

# Add script entries
for script in "${SCRIPTS[@]}"; do
    cat >> "$SKILL_DIR/SKILL.md" << EOF
  - name: $script
    description: TODO: What this script does
    path: scripts/$script
EOF
    # Create empty script file
    cat > "$SCRIPTS_DIR/$script" << 'SCRIPT_EOF'
#!/usr/bin/env bash
# TODO: Add script description
# Usage: ./script-name.sh [options]

set -euo pipefail

echo "TODO: Implement script"
SCRIPT_EOF
    chmod +x "$SCRIPTS_DIR/$script"
    echo "   ✅ Created scripts/$script"
done

# If no scripts, add empty scripts section note
if [[ ${#SCRIPTS[@]} -eq 0 ]]; then
    echo "   (no scripts — skill-only)" >> "$SKILL_DIR/SKILL.md"
fi

# Add body from template
cat >> "$SKILL_DIR/SKILL.md" << 'EOF'
---

# <Display Name> — <One-line tagline>

## Purpose
What problem this skill solves. When agents or users should load it.

**Use when:**
- User says "..." or agent needs X capability
- Before/after specific workflow phases
- Specific trigger conditions met

## Scripts

| Script | Purpose | Key Flags |
|--------|---------|-----------|
| _(none — skill-only)_ | | |

## Workflow

### Step 1: <Action Name>
What happens first. What data is gathered.

### Step 2: <Action Name>
What happens next. How results are processed.

### Step 3: <Action Name>
Final output. What gets returned or saved.

## Integration with Other Skills

- **<skill-name>**: Uses for <purpose>
- **<skill-name>**: Feeds into <purpose>

## Tips

- Best practice or common pattern
- Gotcha to avoid
- When NOT to use this skill
EOF

echo ""
echo "✅ Skill created: $SKILL_DIR/"
echo "   - SKILL.md (with trigger: /$SKILL_NAME)"
if [[ ${#SCRIPTS[@]} -gt 0 ]]; then
    echo "   - scripts/ (${#SCRIPTS[@]} scripts)"
fi
echo ""
echo "Next steps:"
echo "1. Edit $SKILL_DIR/SKILL.md — fill in Purpose, Workflow, Integration"
echo "2. Implement scripts in $SCRIPTS_DIR/"
echo "3. Add to skills/_INDEX.md"
echo "4. Wire to agents in AGENTS.md"
echo "5. Update agent files if this is a primary skill"
