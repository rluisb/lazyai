# Local Example Prompt

**Topic:** [Example Topic]
**Goal:** [Example Goal]

---

## Instructions

0. **CoT — Frame the task before acting:** Briefly restate the topic, goal, assumptions, and constraints. Identify what “good” looks like for the example.
1. **Identify Existing Patterns:** Look for existing patterns, APIs, and conventions.
2. **Create a Local Example:** Create a local example that demonstrates the pattern.
3. **Document the Example:** Document the example with comments and explanations.
4. **Verify the Example:** Verify that the example works as expected.
5. **Produce a Report:** Summarize your findings in a structured report.

## Few-Shot Mini Example

### Input
- **Topic:** Error handling wrapper
- **Goal:** Show how to wrap async operations with typed error handling

### Expected Approach (condensed)
1. Reuse existing error utility patterns
2. Implement a tiny wrapper with one happy-path and one failure-path example
3. Explain why this pattern is preferred in this codebase

### Example Output Snippet

```ts
type Result<T> = { ok: true; data: T } | { ok: false; error: string }

export async function safeCall<T>(fn: () => Promise<T>): Promise<Result<T>> {
  try {
    return { ok: true, data: await fn() }
  } catch (err) {
    return { ok: false, error: err instanceof Error ? err.message : 'Unknown error' }
  }
}
```

## Common Mistakes to Avoid

- Skipping step 0 and generating code before clarifying constraints
- Creating examples that ignore existing local conventions
- Providing code without explaining why the pattern is appropriate
- Leaving the example unverified (no execution or validation notes)
- Using broken or nested code fences that make output unreadable

## Output Format

~~~markdown
## Local Example: [Topic]

### Example
```[language]
[Example code]
```

### Explanation
[Explanation of the example]
~~~
