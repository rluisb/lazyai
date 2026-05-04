# Scopes demo — project vs global
# Placement: near scopes section
Output demo/03-scopes.gif
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

# --- Project scope ---
Hide
Type "export T1=$(mktemp -d)/project-demo && mkdir -p $T1 && cd $T1 && git init -q"
Enter
Sleep 500ms
Show

Type "# Project scope — self-contained in current directory"
Enter
Sleep 1s
Type "ai-setup init --scope project --tools opencode --name my-project --preset standard --no-interactive"
Enter
Sleep 6s
Type "ls -1"
Enter
Sleep 2s

# --- Global scope ---
Hide
Type "export MOCK_HOME=$(mktemp -d) && export HOME=$MOCK_HOME && cd $MOCK_HOME"
Enter
Sleep 300ms
Show

Type "# Global scope — personal defaults at ~/.ai/"
Enter
Sleep 1s
Type "ai-setup init --scope global --tools opencode,claude-code --preset minimal --no-interactive"
Enter
Sleep 6s
Type "echo '~/.ai/:' && ls .ai/ && echo '' && echo '~/.config/opencode/:' && ls .config/opencode/ && echo '' && echo '~/.claude/:' && ls .claude/"
Enter
Sleep 4s

Sleep 2s
