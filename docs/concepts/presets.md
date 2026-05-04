# Feature Presets

Presets control how much guidance and structure `ai-setup` installs by default.

## Preset levels

| Preset | What it enables | Typical use |
|---|---|---|
| `minimal` | `qualityGates` | Lightweight setup, global defaults, cheaper/faster workflows |
| `standard` | `rpiWorkflow`, `chainOfThought`, `qualityGates`, `bugResolution` | Recommended team baseline |
| `full` | All features | Maximum guidance and process structure |

## Feature meanings

| Feature | What it does |
|---|---|
| `contextEngineering` | Adds context discipline, file budget, and session hygiene guidance |
| `rpiWorkflow` | Adds Research → Plan → Implement workflow structure |
| `chainOfThought` | Adds structured reasoning protocol guidance |
| `treeOfThoughts` | Encourages evaluating multiple approaches before choosing one |
| `adrEnforcement` | Prompts ADR usage for significant architecture changes |
| `qualityGates` | Adds lint, typecheck, test, and build verification expectations |
| `agentHarness` | Adds multi-agent coordination and handoff patterns |
| `bugResolution` | Adds reproduce → diagnose → fix → verify debugging structure |
| `pivotHandling` | Adds guidance for requirement changes mid-implementation |

## Customize presets

Disable specific features:

```bash
ai-setup init --preset full --disable-features treeOfThoughts,agentHarness
```

Start from nothing and re-enable only what you want:

```bash
ai-setup init \
  --disable-features all \
  --features rpiWorkflow,qualityGates \
  --scope project \
  --tools opencode \
  --name minimal-custom \
  --no-interactive
```

## Preset and tool selection

Presets are independent of tool selection. You can use `minimal` with all three tools, or `full` with a single tool. The preset controls the content of generated rules and agent instructions; the tool selection controls which directories are created.
