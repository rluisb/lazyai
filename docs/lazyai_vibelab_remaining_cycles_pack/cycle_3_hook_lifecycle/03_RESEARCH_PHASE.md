# Research Phase — Cycle 3

Inspect current hook and adapter behavior.

## Required inspection

```text
packages/cli/library/hooks/
packages/cli/library/claudecode/
packages/cli/library/opencode/
packages/cli/library/antigravity/
packages/cli/library/cupcake/
packages/cli/internal/adapter/
packages/cli/internal/adapter/capabilities.go
packages/cli/internal/adapter/output_mapping.go
docs/concepts/harness-principles.md
```

Answer:

```text
- What hook assets exist?
- Which hooks block actions?
- Which hooks require human approval?
- Which hooks capture evidence?
- Which adapters support executable hooks?
- Which adapters are instruction-only?
- Which adapters do not support hooks?
- Is there already a neutral hook lifecycle model?
```

Confirm current hook-related files emitted for:

```text
OpenCode
Claude Code
Gemini / Antigravity
Copilot
Pi
OMP
Kiro
```
