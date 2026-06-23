**Epic — RPI Cycle 3: Hook lifecycle catalog & adapter capability mapping**

Goal: Define a neutral hook lifecycle and map current hook assets honestly to each supported adapter — without pretending every tool supports hooks equally, and without adding a runtime hook scheduler.

### Why
Repo ships 6 hook assets (`library/hooks/`: pre-commit, rpi-gate-check.yml, caveman-memory-promotion.md, startup-self-heal.md, block-destructive-shell.md, objective-workflow-gate.md) and 7 adapter surfaces, but no neutral lifecycle/capability model. Note: `internal/hooks/` does not yet exist; `internal/adapter/capabilities.go` does.

### Tasks (sub-issues)
- Author neutral hook lifecycle catalog (docs + library)
- Build hook capability matrix per adapter
- Classify the 6 existing hook assets
- Add hook-asset validation checks + tests (or docs-only fallback)

### Lifecycle vocabulary
before_agent, before_model, before_tool, after_tool, after_model, after_agent, on_error, on_compaction, on_handoff, on_human_gate

### Capability vocabulary
supported, partial, instruction_only, unsupported, not_applicable — for opencode, claude, copilot, pi, omp, antigravity, kiro.

### Boundary
Docs + validation + capability mapping preferred over runtime behavior. No hook scheduler. Capability matrix must match adapter code.
