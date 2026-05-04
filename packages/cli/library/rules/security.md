# Security Rules

## Purpose
Prevent security vulnerabilities in code generated or modified by AI agents.

## Secret Detection
- **API keys**: Patterns like `sk-`, `AKIA`, `ghp_`, `glpat-`, `xoxb-`
- **Tokens**: JWT patterns, OAuth tokens, bearer tokens in code
- **Passwords**: Hardcoded strings assigned to `password`, `secret`, `apiKey` variables
- **Connection strings**: Database URLs with credentials

### Enforcement
- Block commits containing detected secret patterns
- Flag `.env` files or config files with plaintext credentials
- Require environment variable references instead of hardcoded values

## Dependency Security
- Check `package-lock.json` / `Gemfile.lock` for known vulnerabilities
- Prefer dependencies with active maintenance (commits within 6 months)
- Audit transitive dependencies for critical CVEs
- Pin exact versions in production dependencies

## Authentication & Authorization
- All auth-related changes require security-focused review
- Session management must use httpOnly, secure, sameSite cookies
- Never expose user IDs in URLs without authorization checks
- Rate limiting required on all authentication endpoints

## Input Validation
- Validate and sanitize ALL user inputs at the boundary
- Use parameterized queries — never string-concatenate SQL
- Validate file uploads: type, size, content
- Sanitize HTML output to prevent XSS

## Security-Sensitive Code Patterns
Flag for mandatory review:
- Changes to authentication/authorization logic
- New API endpoints
- Database migration with data changes
- Third-party service integrations
- File system operations
- Shell command execution
- Cryptographic operations

## HTTPS & Transport
- All external API calls must use HTTPS
- Certificate validation must not be disabled
- No sensitive data in URL query parameters

## Logging & Monitoring
- Never log sensitive data (passwords, tokens, PII)
- Log authentication events (login, logout, failed attempts)
- Log authorization failures
- Ensure log injection is prevented (sanitize log inputs)

## Secure-by-Default Heuristics
- Deny by default on new endpoints until auth rules are explicit
- Require explicit allowlists for outbound hosts in sensitive integrations
- Treat all user-controlled strings as tainted until validated
- Prefer stable, maintained cryptography libraries over custom code

## Review Checklist
- [ ] Threat model updated for new auth or data-flow changes
- [ ] Secrets scan completed and no hardcoded credentials found
- [ ] Input validation exists at all external boundaries
- [ ] Authorization checks verified for every protected action
- [ ] Logs reviewed for sensitive-data leakage

## Escalation Triggers
- Any auth bypass or privilege escalation finding
- Any leaked credential or token in source, logs, or artifacts
- Any disabled TLS verification in production code paths
- Any unreviewed schema/data migration impacting user data
