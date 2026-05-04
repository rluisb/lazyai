# Full interactive wizard — Workspace scope (planning repo separate from code repos)
# Placement: near Workspace Scope section
Output demo/11-scope-workspace-separate.gif
Set FontSize 14
Set Width 1200
Set Height 700
Set Theme "Catppuccin Mocha"
Set Padding 20
Set TypingSpeed 25ms
Set Shell "bash"

Hide
Type "cd /Users/ricardo/projects/lazyai && npm run build 2>/dev/null"
Enter
Sleep 3s
Type "alias ai-setup='/Users/ricardo/projects/lazyai/bin/ai-setup.js'"
Enter
Sleep 300ms

# Set up workspace structure: planning repo is separate from code repos
Type "export WS=$(mktemp -d)"
Enter
Sleep 200ms
Type "mkdir -p $WS/planning-repo $WS/frontend $WS/backend-api"
Enter
Sleep 200ms
# Init git repos in each
Type "cd $WS/planning-repo && git init -q"
Enter
Sleep 200ms
Type `cd $WS/frontend && git init -q && echo '{"name":"frontend"}' > package.json`
Enter
Sleep 200ms
Type `cd $WS/backend-api && git init -q && echo 'module backend-api' > go.mod`
Enter
Sleep 200ms
Type "cd $WS"
Enter
Sleep 300ms
Show

Type "# Workspace with separate planning repo"
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

# Planning repo → select "planning-repo" (2nd option: current dir is 1st, planning-repo is 2nd)
Down
Enter
Sleep 2s

# Referenced repos → select both frontend and backend-api
Space
Down
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
Type "echo '📁 Planning repo (planning-repo/):' && find planning-repo -maxdepth 2 -not -path '*/\.git/*' -not -name '.git' -type f | sort | head -15"
Enter
Sleep 3s

Type "echo '📁 Referenced repo (frontend/):' && find frontend -not -path '*/\.git/*' -not -name '.git' -type f | sort"
Enter
Sleep 3s

Type "echo '📁 Referenced repo (backend-api/):' && find backend-api -not -path '*/\.git/*' -not -name '.git' -type f | sort"
Enter
Sleep 3s

Sleep 2s
