# ai-setup Demo
# Run with: cd /Users/ricardo/projects/teachable/ai-setup && vhs demo/demo.tape

# --- Settings ---
Output demo/demo.gif
Set FontSize 14
Set Width 1200
Set Height 700
Set Theme "Catppuccin Mocha"
Set Padding 20
Set TypingSpeed 35ms
Set Shell "bash"

# --- Prep: build CLI, create temp project ---
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

# --- Title ---
Sleep 500ms
Type "# 🚀 ai-setup — one command to scaffold AI-ready projects"
Enter
Sleep 2s

# --- Step 1: Init ---
Type "ai-setup init --scope project --tools opencode,claude-code,codex --name my-app --preset standard --no-interactive"
Enter
Sleep 8s

# --- Step 2: Explore what was created ---
Sleep 500ms
Type ""
Enter
Type "# What did it create? Let's look..."
Enter
Sleep 1.5s

Type "ls -1"
Enter
Sleep 2s

# Show OpenCode structure
Type "echo '📂 OpenCode structure:' && find .opencode -type f | sort"
Enter
Sleep 3s

# Show Claude Code structure
Type "echo '📂 Claude Code structure:' && find .claude -type f | sort"
Enter
Sleep 3s

# Show Codex structure
Type "echo '📂 Codex (AgentSkills) structure:' && find .agents -type f | sort"
Enter
Sleep 3s

# --- Step 3: Status ---
Type ""
Enter
Type "ai-setup status"
Enter
Sleep 4s

# --- Step 4: Doctor ---
Type "ai-setup doctor"
Enter
Sleep 4s

# --- Step 5: List library ---
Type "ai-setup list agents"
Enter
Sleep 3s

Type "ai-setup list skills"
Enter
Sleep 3s

# --- End ---
Sleep 500ms
Type ""
Enter
Type "# ✅ 3 AI tools configured, tracked, and verified — from one command"
Enter
Sleep 3s
