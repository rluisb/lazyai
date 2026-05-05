# Cupcake Setup Guide — RPI Gate Enforcement

Cupcake provides deterministic, real-time policy enforcement for AI coding
agents. It intercepts tool calls before execution and evaluates them against
Rego policies compiled to WebAssembly.

## Installation

### Via Nix (Recommended)

```bash
nix profile install github:eqtylab/cupcake#cupcake-cli
cupcake --version  # Should report 0.5.1+
```

### Manual Download

Download the latest binary from:
https://github.com/eqtylab/cupcake/releases

### Verify

```bash
cupcake --version
cupcake --help
```

## Quick Start

### 1. Clone the Repository

```bash
git clone <repo-url>
cd <project>
```

### 2. Review the Policies

The project ships with pre-authored Rego policies:

```
policies/
├── common/
│   └── security.rego        # Secret detection, SQL injection blocking
├── claude/
│   └── rpi_gates.rego       # RPI phase gates for Claude Code
└── opencode/
    └── rpi_gates.rego       # RPI phase gates for OpenCode
```

### 3. Start Cupcake

```bash
# For Claude Code:
cupcake watch --harness claude-code

# For OpenCode:
# The plugin is auto-wired via opencode.jsonc
```

### 4. Test the Enforcement

Try writing code without a plan:

```bash
# This should be BLOCKED by cupcake:
echo "console.log('test')" > src/test.ts
# → "⛔ GATE REQUIRED: plan.md with 'Human Gate: APPROVED' must exist"

# Create a valid plan with gate attestation:
mkdir -p specs/001-test
echo "Human Gate: APPROVED by your-name at $(date -u +%Y-%m-%d)" > specs/001-test/plan.md

# Now writes to src/ should proceed
```

## Policy Decisions

Cupcake returns one of five decisions:

| Decision | Meaning | Example |
|----------|---------|---------|
| **Allow** | Action proceeds normally | Writing docs, reading files |
| **Modify** | Action proceeds with changes | Auto-lint after write |
| **Block** | Action is stopped | Writing secrets, pushing to main |
| **Warn** | Action proceeds with warning | File exceeds 300 lines |
| **Require Review** | Action paused until human approves | Writing src/ without approved plan |

## Troubleshooting

### "Why was my write blocked?"

Cupcake enforces the RPI gate protocol. To write to `src/`:
1. Create `specs/<task>/research.md` with your findings
2. Get research approved by a human
3. Create `specs/<task>/plan.md` with `Human Gate: APPROVED`
4. The approval must be written by a human (verified via git authorship)

### "How do I temporarily disable cupcake?"

```bash
# Stop the cupcake daemon
cupcake stop

# Or bypass with env var (emergencies only):
CUPCAKE_DISABLE=1 <your-command>
```

Bypasses are logged and should be used sparingly. The pre-commit hook and
CI checks still enforce gate compliance even when cupcake is disabled.

### "Does cupcake work with Copilot?"

Copilot does not support native hook integration. For Copilot users:
- **Pre-commit hook** catches gate violations at commit time
- **CI gate check** provides a second checkpoint on pull requests
- **Watchdog** (LLM-as-Judge) can review Copilot output post-hoc
  (enable by setting `watchdog.enabled: true` in `cupcake.yml`)

### "Does cupcake consume context tokens?"

No. Policies run as compiled WebAssembly outside the model. Zero context
tokens are consumed by the enforcement layer.

## Updating Policies

1. Edit the appropriate `.rego` file in `policies/`
2. Run `cupcake validate` to check syntax
3. Commit the changes — cupcake picks them up on next evaluation
4. Update AGENTS.md if the policy change affects documented behavior

## Signal Reference

Signals are ground-truth data gathered at evaluation time:

| Signal | What It Checks | Used By |
|--------|---------------|---------|
| `plan_attested` | plan.md contains "Human Gate: APPROVED" | RPI gates |
| `gate_attested` | Git commits contain "Gate: APPROVED" | Commit checks |
| `author_is_human` | Git author is not a bot/AI agent | Anti-forgery |
| `lines_changed` | Lines in staged diff | Trivial change bypass |
| `branch` | Current git branch | Push protection |
| `has_tests` | Test files exist | TDD enforcement |
| `test_suite_passing` | Test command succeeds | Quality gate |
| `spec_exists` | spec.md exists | Doc-driven dev |
| `adr_exists` | ADR exists | Architecture governance |
| `file_too_large` | File exceeds 300 lines | Anti-overengineering |

## Getting Help

- [Cupcake Documentation](https://cupcake.eqtylab.io)
- [GitHub Issues](https://github.com/eqtylab/cupcake/issues)
- Policy examples: `examples/` directory in the cupcake repository
