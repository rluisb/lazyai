# Full interactive wizard — Workspace scope (planning repo IS also a code repo)
# Placement: near Workspace Scope section (second scenario)
Output demo/12-scope-workspace-combined.gif
Set FontSize 14
Set Width 1200
Set Height 700
Set Theme "Catppuccin Mocha"
Set Padding 20
Set TypingSpeed 25ms
Set Shell "bash"

Hide
Type "cd /Users/ricardo/projects/teachable/ai-setup && npm run build 2>/dev/null"
Enter
Sleep 3s
Type "alias ai-setup='/Users/ricardo/projects/teachable/ai-setup/bin/ai-setup.js'"
Enter
Sleep 300ms

# Set up workspace: main-app is BOTH planning repo AND code repo
Type "export WS=$(mktemp -d)"
Enter
Sleep 200ms
Type "mkdir -p $WS/main-app $WS/api-service"
Enter
Sleep 200ms
Type `cd $WS/main-app && git init -q && echo '{"name":"main-app"}' > package.json`
Enter
Sleep 200ms
Type `cd $WS/api-service && git init -q && echo '{"name":"api-service"}' > package.json`
Enter
Sleep 200ms
Type "cd $WS"
Enter
Sleep 300ms
Show

Type "# Workspace where planning repo is also a code repo"
Enter
Sleep 1s
Type "ls -1"
Enter
Sleep 1.5s

Type "ai-setup init"
Enter
Sleep 2s

# Scope → Workspace (2nd option)
Sleep 500ms
Down
Enter
Sleep 2s

# Tools → OpenCode + Claude Code
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

# Workspace name → accept default
Enter
Sleep 2s

# Planning repo → select "main-app" (2nd option)
Down
Enter
Sleep 2s

# Referenced repos → select api-service (only option since main-app is already planning repo)
Space
Enter
Sleep 2s

# Planning dir → accept default
Enter
Sleep 2s

# Preset → Standard (2nd)
Down
Enter
Sleep 2s

# Branch pattern → Conventional
Enter
Sleep 1.5s

# Commit pattern → Conventional
Enter
Sleep 1.5s

# Require ticket → No
Enter
Sleep 2s

# Confirm → Yes
Enter
Sleep 8s

# Show results
Type "echo '📁 main-app/ (planning repo + code repo):' && find main-app -maxdepth 2 -not -path '*/\.git/*' -not -name '.git' -type f | sort | head -20"
Enter
Sleep 3s

Type "echo '📁 api-service/ (referenced repo):' && find api-service -not -path '*/\.git/*' -not -name '.git' -type f | sort"
Enter
Sleep 3s

Sleep 2s
