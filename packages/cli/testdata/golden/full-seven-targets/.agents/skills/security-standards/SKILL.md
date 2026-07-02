---
name: security-standards
description: Use when performing a security audit, threat-modeling a feature, reviewing code for security issues, or prioritizing vulnerability remediation. Provides OWASP checklists, scanning commands, and severity SLAs.
---

# Security Standards

## When to Use

Use this skill when:
- The user asks for a security audit or review of code, a module, or a dependency.
- Designing a feature that touches auth, secrets, user input, or external calls.
- A dependency scan surfaced advisories and remediation must be prioritized.
- Reviewing a PR with security-relevant surface (endpoints, queries, credentials).

## Rule

Audit first, report findings with severity, then fix in priority order. Never skip the report to jump straight to fixing — the audit trail matters.

## OWASP Top 10 Checklist (2021)

| # | Category | Check |
|---|----------|-------|
| A01 | Broken Access Control | Authorization on every endpoint, RBAC enforcement, CORS config |
| A02 | Cryptographic Failures | TLS config, password hashing, sensitive data at rest/transit |
| A03 | Injection | Parameterized queries, input validation, ORM usage |
| A04 | Insecure Design | Threat model exists, abuse cases documented, rate limiting |
| A05 | Security Misconfiguration | No default credentials, sanitized errors, security headers |
| A06 | Vulnerable Components | Dependency audit clean, no known CVEs, lock files current |
| A07 | Auth Failures | MFA available, secure sessions, credential-stuffing protection |
| A08 | Data Integrity Failures | CI/CD secured, dependency verification, signed artifacts |
| A09 | Logging Failures | Security events logged, no sensitive data in logs, alerting |
| A10 | SSRF | Server-side request validation, allowlists for outbound calls |

## Threat Model Template

```markdown
## Scope           — what is being audited (flow, module, API)
## Attack Surface  — entry points, data flows, trust boundaries
## Threat Categories — relevant OWASP categories
## Risk Assessment — likelihood × impact per threat
## Security Controls — existing controls to verify or implement
## Non-Goals       — explicitly out of scope
```

## Scanning Commands

Adapt to the repo stack:

```bash
# Dependency vulnerabilities
cargo audit          # Rust
npm audit            # Node (or pnpm audit)
pip-audit            # Python
trivy fs . / grype . # stack-agnostic

# Secrets scanning
gitleaks detect
trufflehog filesystem .
```

## Remediation Priority

| Severity | SLA | Action |
|----------|-----|--------|
| CRITICAL | Block merge immediately | Fix before any other work |
| HIGH | Current sprint | Prioritize over feature work |
| MEDIUM | Within 2 sprints | Schedule alongside regular work |
| LOW | Opportunistic | Track in backlog |

## Code-Review Security Checklist

- [ ] No hardcoded secrets, API keys, or tokens
- [ ] No sensitive data in logs or error messages
- [ ] Input validation on all user-facing endpoints
- [ ] Authorization checks on every protected resource
- [ ] Parameterized queries (no string-built SQL)
- [ ] Dependencies scanned for known CVEs
- [ ] Rate limiting on authentication endpoints
- [ ] CORS configured restrictively
- [ ] Security headers present (CSP, HSTS, X-Frame-Options)
- [ ] Error responses do not leak internal details

## Constraints

- Findings without severity are incomplete; every finding gets CRITICAL/HIGH/MEDIUM/LOW.
- Never disable a security check (e.g. secret scanners) to make a pipeline pass.
- Threat models list non-goals explicitly to keep audits scoped.

## Verification Checklist

- [ ] Scope and non-goals stated before auditing.
- [ ] Relevant OWASP categories checked, not just the easy ones.
- [ ] Dependency + secrets scans ran (or documented why not).
- [ ] Findings reported with severity before any fix landed.
- [ ] Remediation ordered by the SLA table.

## Related Skills

- `architecture-review` — security-relevant design decisions.
