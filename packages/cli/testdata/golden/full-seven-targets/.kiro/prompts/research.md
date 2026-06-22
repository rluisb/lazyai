# Research Prompt

## Examples

**Input**: "Why does checkout sometimes timeout after 30 seconds?"
→ Searched: payment/, middleware/, config/
→ Found: PaymentClient has hardcoded 30s timeout in client.ts:42
→ Related: ADR-005 chose synchronous payment flow (specs/adrs/005-sync-payments.md)

**Input**: "How does auth work in this project?"
→ Searched: auth/, middleware/, config/
→ Found: JWT middleware in middleware/auth.ts, token refresh in services/auth-service.ts
→ Patterns: all protected routes use `requireAuth()` middleware

**Input**: "What would be affected by changing the user model?"
→ Searched: models/user.ts imports, test files, migrations
→ Found: 12 files import User, 3 services depend on user.email, 2 migrations reference users table

## Common Mistakes
- ❌ Suggesting improvements during research (research observes, it doesn't judge)
- ❌ Reading the entire codebase instead of targeted searches
- ❌ Making assumptions about code behavior without reading the source
