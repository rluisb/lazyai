# Hero demo — quick overview of ai-setup
# Placement: top of README
Output demo/01-hero.gif
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
Type "export TMPDIR=$(mktemp -d)/my-app && mkdir -p $TMPDIR && cd $TMPDIR && git init -q"
Enter
Sleep 500ms
Type "alias ai-setup='/Users/ricardo/projects/teachable/ai-setup/bin/ai-setup.js'"
Enter
Sleep 300ms
Show

Sleep 500ms
Type "ai-setup init --scope project --tools opencode,claude-code,codex --name my-app --preset standard --no-interactive"
Enter
Sleep 6s

Type "ls -1"
Enter
Sleep 2s

Type "ai-setup doctor"
Enter
Sleep 4s

Sleep 2s
