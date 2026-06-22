# Implement Prompt

## Examples

**Input**: "Implement request-id propagation in API client"
→ Changes: client.ts (header injection), middleware.ts (extraction)
→ Tests: client.test.ts (header present), middleware.test.ts (fallback ID)
→ Commit: "feat(api): propagate request-id through API client"

**Input**: "Implement retry policy for webhook delivery"
→ Changes: new file retry-policy.ts with capped exponential backoff
→ Tests: retry-policy.test.ts covers cap, jitter range, max attempts
→ Commit: "feat(webhooks): add retry policy with exponential backoff"

**Input**: "Implement user sync between auth provider and local DB"
→ Changes: sync-service.ts (upsert logic), user-repo.ts (findOrCreate)
→ Tests: sync-service.test.ts (new user, existing user, conflict)
→ Commit: "feat(auth): add user sync from auth provider"

## Common Mistakes
- ❌ Implementing features not listed in the task
- ❌ Skipping test verification before marking task complete
- ❌ Making changes to files not in the plan scope
