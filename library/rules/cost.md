# LLM & API Cost Management

## Purpose
Manage AI model token consumption and API costs throughout development workflows.

## Rules

### Model Selection
- Use the cheapest model that meets quality requirements for each task
- Scout/Research: fast models (Sonnet, GPT-4o-mini) — low cost, high throughput
- Planning/Review: capable models (Opus, GPT-4) — higher cost justified by quality
- Builder/Implementation: balanced models (Sonnet) — cost-effective for code generation
- Never use premium models for trivial tasks (formatting, renaming, simple edits)

### Context Management
- Keep context under 70% of model's window to avoid degraded output quality
- Compress/compact sessions proactively — every token in context costs money
- Use sub-agents with minimal context instead of loading everything into one session
- Remove stale files, logs, and exploration artifacts from context before generation

### Token Budget Awareness
- Estimate token cost before starting expensive operations (large file reads, multi-file edits)
- Prefer targeted file reads over full-file reads
- Batch related changes in single sessions rather than opening new sessions per file
- Use `--dry-run` to preview operations before committing to expensive scaffold runs

### API Cost Monitoring
- Track monthly API spend per project
- Set budget alerts at 50%, 75%, 90% of monthly allocation
- Review cost-per-task periodically — spikes indicate inefficient prompting
- Prefer local tools (linters, formatters) over LLM for mechanical tasks

### Anti-Patterns to Avoid
- ❌ Using Opus/GPT-4 for simple grep-like searches
- ❌ Loading entire codebases into context "just in case"
- ❌ Running the same prompt repeatedly without adjusting approach
- ❌ Generating code without a plan (leads to expensive rework loops)
- ❌ Ignoring compaction warnings until context window is full

## Enforcement
- Agents should log model selection rationale in task journals
- Reviewers should flag unnecessary premium model usage
- Cost rule violations are advisory, not blocking
