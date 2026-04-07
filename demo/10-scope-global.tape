# Full interactive wizard — Global scope
# Placement: near Global Scope section
Output demo/10-scope-global.gif
Set FontSize 14
Set Width 1200
Set Height 700
Set Theme "Catppuccin Mocha"
Set Padding 20
Set TypingSpeed 30ms
Set Shell "bash"

Hide
Type "cd /Users/ricardo/projects/teachable/ai-setup && npm run build 2>/dev/null"
Enter
Sleep 3s
Type "alias ai-setup='/Users/ricardo/projects/teachable/ai-setup/bin/ai-setup.js'"
Enter
Sleep 300ms
# Mock a clean HOME so we don't mess with real global config
Type "export MOCK_HOME=$(mktemp -d) && export HOME=$MOCK_HOME && cd $MOCK_HOME"
Enter
Sleep 500ms
Show

Type "ai-setup init"
Enter
Sleep 2s

# Scope → Global (1st option, already selected)
Enter
Sleep 2s

# Tools → OpenCode + Claude Code (only supported for global)
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

# Planning dir → accept default
Enter
Sleep 2s

# Preset → Minimal (1st, default for global)
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
Type "echo '📁 ~/.ai/:' && ls -1 .ai/"
Enter
Sleep 2s

Type "echo '📁 ~/.config/opencode/:' && find .config/opencode -type f | sort"
Enter
Sleep 3s

Type "echo '📁 ~/.claude/:' && find .claude -type f | sort"
Enter
Sleep 3s

Sleep 2s
