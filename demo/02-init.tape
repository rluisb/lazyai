# Init command demo — project setup with different presets
# Placement: near init command section
Output demo/02-init.gif
Set FontSize 14
Set Width 1200
Set Height 600
Set Theme "Catppuccin Mocha"
Set Padding 20
Set TypingSpeed 30ms
Set Shell "bash"

Hide
Type "cd /Users/ricardo/projects/lazyai && npm run build 2>/dev/null"
Enter
Sleep 3s
Type "alias ai-setup='/Users/ricardo/projects/lazyai/bin/ai-setup.js'"
Enter
Sleep 300ms
Show

# --- Minimal preset ---
Hide
Type "export T1=$(mktemp -d)/minimal-app && mkdir -p $T1 && cd $T1 && git init -q"
Enter
Sleep 500ms
Show

Sleep 300ms
Type "# Minimal preset — lightweight, cheap models"
Enter
Sleep 1s
Type "ai-setup init --scope project --tools opencode --name minimal-app --preset minimal --no-interactive"
Enter
Sleep 6s
Type "find . -maxdepth 2 -not -path './.git/*' -not -path './.git' | sort | head -20"
Enter
Sleep 3s

# --- Full preset ---
Hide
Type "export T2=$(mktemp -d)/full-app && mkdir -p $T2 && cd $T2 && git init -q"
Enter
Sleep 500ms
Show

Sleep 300ms
Type "# Full preset — everything enabled"
Enter
Sleep 1s
Type "ai-setup init --scope project --tools opencode,claude-code --name full-app --preset full --no-interactive"
Enter
Sleep 6s
Type "find . -maxdepth 2 -not -path './.git/*' -not -path './.git' | sort | head -25"
Enter
Sleep 3s

Sleep 2s
