# P0-7: Token-Rent Enforcement Design

**Status:** Complete — unified contract  
**Owner:** Ricardo Conceicao  
**Date:** 2026-06-14  
**Linked from:** `plan.md` Phase 0, P0-7

---

## Purpose

Design for enforcing a 50KB project-wide token-rent budget on the canonical library. Per Clarify resolution: single budget, CI/pre-commit enforcement, documented override.

## Budget

| Aspect | Design |
|---|---|
| Threshold | 50,000 bytes (50KB) |
| Scope | `packages/cli/library/canonical/` — all agents, skills, hooks, commands |
| Measurement | `wc -c` sum of all files in scope (byte count, not token count) |
| Check frequency | Pre-commit hook + CI pipeline |

## Enforcement

| Aspect | Design |
|---|---|
| CI integration | GitHub Actions workflow: `token-rent.yml`, runs on PR, fails if total > 50,000 bytes |
| Pre-commit hook | Hook in `.githooks/pre-commit` or `pre-commit` config |
| Failure output | `Library budget exceeded: X / 50000 bytes. Override: add .lazyai/token-rent-override with justification.` |
| Override path | `.lazyai/token-rent-override` — single file at project root |
| Override format | YAML or plain text with `reason:` field |
| Override semantics | Documented exception; auditable; does not silently bypass |

## Override Format

```yaml
# .lazyai/token-rent-override
budget: 76800  # 75KB
reason: "Additional agent definitions required for enterprise compliance workflows"
approved_by: "security-team"
expires: "2026-09-01"
```

## Byte-Counting Rule

- Count: `wc -c` on all files in `packages/cli/library/canonical/` (recursive)
- Exclude: `.gitkeep` and other non-content files
- Include: all markdown body text, code blocks, tool definitions, instruction text

## Gate

⛔ Human must approve this design before Phase 5 library curation enforcement begins.
