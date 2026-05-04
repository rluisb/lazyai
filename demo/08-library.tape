# Library demo — list and info commands
# Placement: near list/info and library content sections
Output demo/08-library.gif
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

Type "# Browse bundled agents"
Enter
Sleep 1s
Type "ai-setup list agents"
Enter
Sleep 3s

Type "# Browse bundled skills"
Enter
Sleep 1s
Type "ai-setup list skills"
Enter
Sleep 3s

Type "# Browse MCP servers"
Enter
Sleep 1s
Type "ai-setup list mcp"
Enter
Sleep 3s

Type "# Get details on a specific item"
Enter
Sleep 1s
Type "ai-setup info builder"
Enter
Sleep 4s

Sleep 2s
