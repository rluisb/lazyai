# Maintenance demo — status, doctor, update
# Placement: near update/doctor/status sections
Output demo/06-maintain.gif
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
Type "export TD=$(mktemp -d)/my-app && mkdir -p $TD && cd $TD && git init -q"
Enter
Sleep 300ms
Type "ai-setup init --scope project --tools opencode,claude-code --name my-app --preset standard --no-interactive 2>&1 | tail -3"
Enter
Sleep 5s
Show

Type "# Check setup status"
Enter
Sleep 1s
Type "ai-setup status"
Enter
Sleep 4s

Type "# Verify file integrity"
Enter
Sleep 1s
Type "ai-setup doctor"
Enter
Sleep 4s

Type "# Simulate a drift: modify a tracked file"
Enter
Sleep 1s
Type "echo '# Modified!' >> AGENTS.md"
Enter
Sleep 500ms

Type "# Doctor detects the change"
Enter
Sleep 1s
Type "ai-setup doctor"
Enter
Sleep 4s

Type "# Update restores managed files"
Enter
Sleep 1s
Type "ai-setup update --force"
Enter
Sleep 5s

Sleep 2s
