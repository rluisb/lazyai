# Implement Prompt

**Task:** [Task Name]
**Spec:** [Link to Task Spec]

---

## Instructions

0. **Think Step-by-Step (CoT):** Privately reason step-by-step before edits/tests, but output only concise implementation outcomes.
1. **Read Context First:** Understand existing conventions before writing code.
2. **Test as You Go:** Write tests for every non-trivial function.
3. **One Task at a Time:** Complete and verify before moving to the next.
4. **Follow the Plan:** Don't expand scope; create a new task if you find more work.
5. **Commit Atomically:** One logical change per commit.

## Few-Shot Mini Example (Generic)

Use this pattern as a guide:

```
Input (summary): Add request-id propagation to API client.
Output (shape):
- Changes: client.ts (header injection), middleware.ts (context extraction)
- Tests: client.test.ts (header present), middleware.test.ts (context fallback)
```

```
Input (summary): Create retry policy helper for webhook delivery.
Output (shape):
- Changes: new file retry-policy.ts with capped exponential backoff
- Tests: retry-policy.test.ts covers cap, jitter range, and max attempts
```

```
Input (summary): Modify user sync to skip deactivated records.
Output (shape):
- Changes: sync-service.ts adds active-status guard before upsert
- Tests: sync-service.test.ts verifies deactivated records are ignored
```

## Common Mistakes to Avoid
- ❌ Implementing features not explicitly requested (speculation)
- ❌ Skipping test verification before marking complete
- ❌ Making changes to files not listed in the plan

## Output Format

```
## Implementation: [Task Name]

### Changes
- [file] — [change description]

### Tests
- [test file] — [test description]
```
