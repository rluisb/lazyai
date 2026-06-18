# Output Schemas

## Dispatch Format (non-negotiable)

Every dispatch MUST include this block before the task description:

```
## Dispatch Parameters
AGENT: <target-agent-name>
MODE: <agent-specific-mode>
THINK: <true|xhigh|omit>
MAX_ATTEMPTS: <N>
DRY_RUN: <true|false>

## Task
<detailed task description with spec references and file paths>
```

If you omit parameters, the agent defaults to its safest mode.

- **TOKEN_BUDGET**: Maximum context tokens (default: 50K). Self-manage compression when approaching limit.

## Parameter Catalog

### turbo-crank (specify/plan)
| Parameter | Values | Default | Effect |
|-----------|--------|---------|--------|
| MODE | `clarify` / `research` / `plan` / `full` | `full` | Which storm-scout phase(s) to run |
| THINK | `true` / `xhigh` | `true` | Reasoning depth |
| MAX_QUESTIONS | 3–5 | 3 | Grill Me interrogation depth |
| SKIP_CONSTITUTION | `true` / `false` | `false` | Skip constitution for small changes |

### wall-builder (implementor)
| Parameter | Values | Default | Effect |
|-----------|--------|---------|--------|
| MODE | `junior` / `standard` / `senior` / `tdd` | `standard` | Preflight depth, escalation threshold |
| THINK | `true` / `xhigh` | `true` | Reasoning effort |
| MAX_ATTEMPTS | 3–8 | 5 | Escalation ceiling |
| DRY_RUN | `true` / `false` | `false` | Plan changes only, no writes |

### shield-audit (verifier)
| Parameter | Values | Default | Effect |
|-----------|--------|---------|--------|
| MODE | `quick` / `review` / `security` / `adversarial` | `review` | Verification depth and scope |
| THINK | `xhigh` | `xhigh` | Frontier reasoning for bug detection |
| FOCUS | `spec` / `security` / `performance` / `all` | `all` | Narrow review scope |
| MAX_ATTEMPTS | 3–8 | 5 | Gate re-check ceiling |

### loot-hawk (scout)
| Parameter | Values | Default | Effect |
|-----------|--------|---------|--------|
| MODE | `shallow` / `deep` / `exhaustive` | `deep` | Exploration depth |
| THINK | `true` / `false` | `true` | Reasoning on findings |
| DOMAIN | string | (none) | Focus exploration on one area |
| OUTPUT | `findings` / `map` / `paths` | `findings` | Output format |

### rift-deploy (ops)
| Parameter | Values | Default | Effect |
|-----------|--------|---------|--------|
| MODE | `dry-run` / `staging` / `production` | `dry-run` | Deployment target |
| THINK | `true` / `false` | `true` | Reasoning depth |
| PREFLIGHT | `true` / `false` | `true` | Run quality gates before deploy |

### respawn-crew (sre)
| Parameter | Values | Default | Effect |
|-----------|--------|---------|--------|
| MODE | `triage` / `diagnose` / `mitigate` / `post-mortem` | `triage` | Incident phase |
| THINK | `true` / `xhigh` | `true` | Diagnostic reasoning |
| SEVERITY | `P1` / `P2` / `P3` / `P4` | `P3` | Escalation urgency |
| TIMEBOX | minutes | 5 | Max investigation time before escalate |
| TOKEN_BUDGET | token count | 40K | Advisory context limit |

### Global (all agents)
| Parameter | Values | Effect |
|-----------|--------|--------|
| REPORT_ONLY | flag | Read-only mode — no mutations |
| PLAN_ONLY | flag | Produce plan, no execution |
| WORKTREE | path | Override worktree location |
| SILENT | `true` / `false` | Suppress verbose output |
| TOKEN_BUDGET | token count | Advisory context limit; agent self-compresses when approaching |
