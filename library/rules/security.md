# Rule: Security

**Category:** Security
**Status:** Active

---

## Rule

All code must be secure by default.

## Rationale

Protects user data and system integrity.

## Guidelines

- **Input Validation:** Validate all external input.
- **Authentication:** Require authentication for all protected resources.
- **Authorization:** Enforce least-privilege access control.
- **Data Protection:** Encrypt sensitive data at rest and in transit.
- **Dependency Management:** Keep dependencies updated and scan for vulnerabilities.

## Enforcement

- Static analysis tools (SAST)
- Dynamic analysis tools (DAST)
- Security reviews
