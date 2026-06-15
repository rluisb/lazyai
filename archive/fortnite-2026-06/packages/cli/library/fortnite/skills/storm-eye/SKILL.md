---
name: storm-eye
description: AI evaluation framework for the Fortnite multi-agent system. Defines eval objectives per workflow, manages datasets, runs LLM-as-judge scoring, pairwise comparisons, and tracks performance trends. Measures agent/skill/workflow quality across model updates.
trigger: /storm-eye
triggers:
  - "evaluate this"
  - "run evals"
  - "benchmark"
  - "measure quality"
  - "eval this workflow"
  - "compare models"
  - "grade this output"
skill_path: skills/storm-eye
---

## Quick Reference

| | |
|---|---|
| **Use when** | AI evaluation, model comparison, quality tracking |
| **Do not use when** | Normal implementation, routine review |
| **Primary agent** | shield-audit |
| **Runtime risk** | Low — eval framework |
| **Outputs** | Eval scores, trend reports, dataset metrics |
| **Validation** | Runnable eval suites, threshold checks |
| **Deep mode trigger** | `/storm-eye` or explicit eval request |

# Storm Eye — AI Evaluation Framework

**Tagline**: Watch everything. Measure everything. Improve everything.

## Purpose

Multi-agent systems are nondeterministic. The same input can produce different outputs across model updates, prompt changes, or even identical runs. Storm Eye provides structured evaluation to measure quality, detect degradation, and drive improvement across the entire Fortnite system.

**Use when:**
- After model updates — "did quality improve or degrade?"
- Before shipping a workflow change — "is the new version better?"
- Comparing agents — "which model performs better for this task?"
- Tracking trends — "are we getting better over time?"
- Debugging failures — "where exactly did quality break down?"

## Architecture

```
storm-eye (eval framework)
  ├── Define eval objectives per workflow
  ├── Manage eval datasets (golden cases, edge cases)
  ├── Run evaluators (LLM-as-judge, pairwise, metric-based)
  ├── Track performance trends over time
  └── Feed results back into improvement loop
       ├── Low scores → lesson-loot
       ├── Persistent failures → ricochet (§V invariants)
       ├── Model degradation → health-check alert
       └── All runs → truth-chain ledger
```

## Eval Objectives Per Workflow

### rpi (Research → Plan → Implement → Verify)
| Objective | Metric | Target | Evaluator |
|-----------|--------|--------|-----------|
| Spec compliance | §V invariant pass rate | >= 95% | zero-point MODE=eval |
| Instruction following | Phase output matches feedforward | >= 90% | LLM-as-judge |
| Phase transition accuracy | Correct agent dispatched | >= 98% | Metric-based |
| Implementation quality | shield-wall + build-fort pass rate | >= 85% | shield-wall / build-fort |

### bugfix (Diagnose → Fix → Verify)
| Objective | Metric | Target | Evaluator |
|-----------|--------|--------|-----------|
| Root cause accuracy | Diagnosed cause matches actual | >= 80% | LLM-as-judge |
| Fix correctness | Bug resolved, no regressions | >= 90% | zero-point MODE=eval |
| No workarounds | Zero workaround patterns | 100% | shield-wall gate functions |

### review-others (Review Someone Else's PR)
| Objective | Metric | Target | Evaluator |
|-----------|--------|--------|-----------|
| Finding relevance | Findings match actual issues | >= 85% | Human calibration |
| Severity accuracy | Rating matches impact | >= 80% | LLM-as-judge |
| Format quality | Output ready to copy-paste | >= 95% | Metric-based |

### review-mine (Process Feedback on Your PR)
| Objective | Metric | Target | Evaluator |
|-----------|--------|--------|-----------|
| Classification accuracy | Question vs change-request correct | >= 90% | LLM-as-judge |
| Evaluation validity | Assessment matches spec/docs | >= 85% | zero-point MODE=eval |
| Dispatch correctness | Right workflow triggered | >= 95% | Metric-based |

## Eval Datasets

### Dataset Types
| Type | Source | Use Case |
|------|--------|----------|
| **Golden cases** | Human-curated ideal inputs/outputs | Baseline quality measurement |
| **Production samples** | truth-chain ledger replay | Real-world performance |
| **Edge cases** | Generated from anti-patterns | Robustness testing |
| **Adversarial cases** | Deliberately tricky inputs | Security/stress testing |
| **Historical regressions** | Past failures from lesson-loot | Regression prevention |

### Dataset Management
```
.specify/evals/
  datasets/
    rpi-golden.jsonl        # Golden test cases for rpi workflow
    bugfix-edge.jsonl       # Edge cases for bugfix workflow
    review-others-prod.jsonl # Production samples for review
  results/
    2026-05-18-rpi-v4.json  # Eval run results
    trends.json             # Aggregated performance trends
```

## Evaluator Types

### 1. LLM-as-Judge (via zero-point MODE=eval)
```
Input: Agent output + eval rubric
Model: GPT-4.1 or o3 (strongest available judge)
Output: Score (0-1) + reasoning + specific issues

Rubric example for rpi clarify phase:
- Did the agent identify all critical unknowns? (0-1)
- Did it ask focused, non-redundant questions? (0-1)
- Did it search existing knowledge before asking? (0-1)
- Was output structured and actionable? (0-1)
```

### 2. Pairwise Comparison (via shield-audit MODE=judge)
```
Input: Output A (model X) vs Output B (model Y) + criteria
Output: Winner + confidence + reasoning

Use cases:
- deepseek-v4-pro vs kimi-k2.6 for plan phase
- New prompt vs old prompt for same task
- Agent A vs Agent B for same workflow step
```

### 3. Metric-Based (automated)
| Metric | How It's Measured | Tool |
|--------|-------------------|------|
| Spec compliance | §V invariant pass/fail | drift-check.sh |
| Anti-pattern count | shield-wall + build-fort violations | shield-wall / build-fort |
| Test pass rate | quality-gate.sh results | quality-gate.sh |
| Phase completion | workflow phase status | workflow-run.sh status |
| Token efficiency | tokens per phase | truth-chain query |

## Modes

| Mode | Behavior | Use Case |
|------|----------|----------|
| **define** | Set up eval objectives and datasets for a workflow | Initial setup |
| **run** | Execute eval suite against current system state | Regular measurement |
| **compare** | Run pairwise comparison between two configs | Model/prompt selection |
| **trend** | Show performance trends over time | Improvement tracking |
| **calibrate** | Human reviews eval results, adjusts scoring | Quality assurance |

## Continuous Evaluation Pipeline

```
On every model update or workflow change:
  1. storm-eye run --workflow rpi --dataset golden
  2. Compare scores against baseline
  3. If degraded > 5%: BLOCK deployment, alert
  4. If improved: update baseline, log to truth-chain
  5. Weekly: human calibration of top 10% and bottom 10%
```

## Integration

### With zero-point
- storm-eye defines eval rubrics -> zero-point MODE=eval executes them
- Eval results feed into verification gates
- Low eval scores -> zero-point flags for human review

### With shield-audit
- Pairwise comparison via MODE=judge
- Cross-model evaluation (deepseek-v4-pro vs kimi-k2.6)
- Adversarial eval cases

### With truth-chain
- Every eval run logged as `eval_run` entry
- Tracks: workflow, dataset, scores, model version, timestamp
- Enables trend analysis over time

### With lesson-loot
- Low eval scores -> auto-trigger lesson extraction
- Pattern: "rpi clarify phase scored 0.4 on instruction following"
- Lesson stored: what went wrong, how to improve

### With ricochet
- Persistent eval failures -> new §V invariants
- "rpi plan phase must produce spec with <= 3 ambiguous requirements"
- Backprop updates compact-blueprint spec

### With health-check.sh
- Eval scores as health metrics
- Alert when any workflow drops below threshold
- Trend dashboard in health report

### With workflow-engine
- Pre-deployment gate: evals must pass before shipping
- Post-deployment: continuous eval monitoring
- Degradation -> rollback trigger

## Rules

1. **Eval before ship** — never deploy a workflow change without running evals
2. **Golden dataset is sacred** — never modify golden cases to make scores look better
3. **Human calibrates machine** — periodic human review of LLM-as-judge scores
4. **Trends over snapshots** — a single score is noise; trends are signal
5. **Degradation = blocker** — if quality drops, stop and investigate
6. **Log everything** — every eval run to truth-chain
7. **Feed the loop** — low scores -> lesson-loot -> ricochet -> improved specs
8. **Compare fairly** — same dataset, same rubric, different models/configs

## Output Format

```
👁️ Storm Eye Eval Report
   Workflow: rpi
   Dataset: golden (50 cases)
   Model: deepseek-v4-pro
   Date: 2026-05-18T18:00:00Z

   Scores:
   ✅ Spec compliance:      0.96 (target: >= 0.95)
   ✅ Instruction following: 0.92 (target: >= 0.90)
   ✅ Phase transition:      0.99 (target: >= 0.98)
   ⚠️ Implementation quality: 0.83 (target: >= 0.85) — BELOW TARGET

   Trend (last 5 runs):
   Spec compliance:     0.94 -> 0.95 -> 0.96 -> 0.95 -> 0.96 ↗
   Instruction:         0.88 -> 0.90 -> 0.91 -> 0.92 -> 0.92 ↗
   Implementation:      0.87 -> 0.86 -> 0.85 -> 0.84 -> 0.83 ↘ DEGRADING

   Actions:
   - Implementation quality trending down — investigate
   - Trigger lesson-loot on implementation phase failures
   - Consider ricochet backprop for new §V invariants
```

## Anti-Patterns

| Pattern | Why Wrong | Fix |
|---------|-----------|-----|
| "It looks fine to me" | Vibe-based eval — no measurement | Run storm-eye with golden dataset |
| "We'll add evals later" | Without evals, you can't detect degradation | Define objectives before first workflow run |
| "The scores are always high" | Dataset may be too easy or overfitted | Add edge cases, adversarial cases, production samples |
| "One bad score means it's broken" | Single scores are noisy | Look at trends over 5+ runs |
| "LLM-as-judge is perfect" | Judges have bias (position, verbosity) | Periodic human calibration |
