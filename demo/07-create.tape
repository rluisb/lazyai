# Create demo — scaffold new agents, skills, workflows
# Placement: near create command section
Output demo/07-create.gif
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
Type "ai-setup init --scope project --tools opencode --name my-app --preset standard --no-interactive 2>&1 | tail -2"
Enter
Sleep 5s
Show

Type "# Create a new agent"
Enter
Sleep 1s
Type "ai-setup create agent security-auditor --description 'Reviews code for security vulnerabilities' --mode autonomous --no-interactive"
Enter
Sleep 4s

Type "# Create a new skill"
Enter
Sleep 1s
Type "ai-setup create skill code-review --description 'Structured code review workflow' --no-interactive"
Enter
Sleep 4s

Type "# Create a new workflow"
Enter
Sleep 1s
Type "ai-setup create workflow deploy --description 'Build, test, deploy pipeline' --no-interactive"
Enter
Sleep 4s

Type "# See what was created"
Enter
Sleep 1s
Type "find specs -name '*.md' -newer .ai-setup.json | sort"
Enter
Sleep 3s

Sleep 2s
