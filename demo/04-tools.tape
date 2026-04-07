# Tool structures demo — what each tool generates
# Placement: near supported tools section
Output demo/04-tools.gif
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
Type "export TD=$(mktemp -d)/all-tools && mkdir -p $TD && cd $TD && git init -q"
Enter
Sleep 500ms
Type "ai-setup init --scope project --tools opencode,claude-code,gemini,copilot,codex --name my-app --preset standard --no-interactive 2>&1 | tail -5"
Enter
Sleep 6s
Show

Type "# OpenCode → .opencode/ + opencode.json + AGENTS.md"
Enter
Sleep 1s
Type "find .opencode -type f | sort"
Enter
Sleep 3s

Type "# Claude Code → .claude/ + CLAUDE.md"
Enter
Sleep 1s
Type "find .claude -type f | sort"
Enter
Sleep 3s

Type "# Gemini → .gemini/ + GEMINI.md"
Enter
Sleep 1s
Type "find .gemini -type f | sort"
Enter
Sleep 2.5s

Type "# Copilot → .github/ + AGENTS.md"
Enter
Sleep 1s
Type "find .github -type f | sort"
Enter
Sleep 2.5s

Type "# Codex → .agents/ (AgentSkills standard) + AGENTS.md"
Enter
Sleep 1s
Type "find .agents -type f | sort"
Enter
Sleep 2.5s

Sleep 2s
