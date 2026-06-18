# Full interactive wizard — Project scope
# Placement: near Project Scope section
Output demo/09-scope-project.gif
Set FontSize 14
Set Width 1200
Set Height 700
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
Type "export TD=$(mktemp -d)/my-project && mkdir -p $TD && cd $TD && git init -q"
Enter
Sleep 500ms
Show

Type "ai-setup init"
Enter
Sleep 2s

# Scope → Project (3rd option)
Sleep 500ms
Down
Down
Enter
Sleep 2s

# Tools → OpenCode + Claude Code + Gemini
Space
Down
Space
Down
Space
Enter
Sleep 2s

# CLI tools → skip
Enter
Sleep 2s

# Integrations → skip
Enter
Sleep 2s

# Project name → "my-project" (accept default)
Enter
Sleep 2s

# Planning dir → accept default ".planning"
Enter
Sleep 2s

# Preset → Standard (2nd)
Down
Enter
Sleep 2s

# Branch pattern → Conventional (1st)
Enter
Sleep 1.5s

# Commit pattern → Conventional (1st)
Enter
Sleep 1.5s

# Require ticket → No
Enter
Sleep 2s

# Confirm → Yes
Enter
Sleep 6s

# Show what was created
Type "ls -1"
Enter
Sleep 2s

Type "echo '📁 Generated structure:' && find . -maxdepth 3 -not -path './.git/*' -not -name '.git' | sort | head -30"
Enter
Sleep 4s

Sleep 2s
