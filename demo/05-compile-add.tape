# Compile + Add demo — re-compile and add tools
# Placement: near compile and add command sections
Output demo/05-compile-add.gif
Set FontSize 14
Set Width 1200
Set Height 600
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
Type "export TD=$(mktemp -d)/my-app && mkdir -p $TD && cd $TD && git init -q"
Enter
Sleep 300ms
Type "ai-setup init --scope project --tools opencode --name my-app --preset standard --no-interactive 2>&1 | tail -3"
Enter
Sleep 5s
Show

Type "# Start with OpenCode only"
Enter
Sleep 1s
Type "ls -1"
Enter
Sleep 2s

Type "# Add Claude Code to existing setup"
Enter
Sleep 1s
Type "ai-setup add claude-code"
Enter
Sleep 5s

Type "ls -1"
Enter
Sleep 2s

Type "# Re-compile all tools from .ai/"
Enter
Sleep 1s
Type "ai-setup compile --force"
Enter
Sleep 4s

Sleep 2s
