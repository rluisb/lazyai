# Orchestration demo — enable orchestration, inspect catalog, create custom artifacts
# Placement: near orchestration docs / README orchestration section
Output demo/13-orchestration.gif
Set FontSize 14
Set Width 1280
Set Height 720
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
Type "export TD=$(mktemp -d)/orchestration-demo && mkdir -p $TD && cd $TD && git init -q"
Enter
Sleep 400ms
Show

Type "# Enable orchestrator during init"
Enter
Sleep 800ms
Type "ai-setup init --scope project --tools opencode,claude-code,codex,copilot,gemini --enable-servers orchestrator --name orchestration-demo --preset standard --no-interactive"
Enter
Sleep 7s

Type "# Inspect the scaffolded orchestration tree"
Enter
Sleep 800ms
Type "find .ai/orchestration -maxdepth 2 -type d | sort"
Enter
Sleep 3s

Type "# List built-in and project orchestration artifacts"
Enter
Sleep 800ms
Type "ai-setup list orchestration --json"
Enter
Sleep 4s

Type "# Create a custom domain and workflow"
Enter
Sleep 800ms
Type "ai-setup create domain payments --description 'Payments domain constraints' --no-interactive"
Enter
Sleep 3s
Type "ai-setup create workflow payments-review --chain feature --team review-team --no-interactive"
Enter
Sleep 3s

Type "# Inspect the generated workflow"
Enter
Sleep 800ms
Type "ai-setup info payments-review --json"
Enter
Sleep 4s

Sleep 2s
