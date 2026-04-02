# Research Prompt

**Topic:** [Research Topic]
**Goal:** [Research Goal]

---

## Instructions

0. **Think Step-by-Step (CoT):** Privately reason step-by-step about search strategy and evidence quality, but output only evidence-backed conclusions.
1. **Explore the Codebase:** Read relevant files, directories, and documentation.
2. **Identify Patterns:** Look for existing patterns, APIs, and conventions.
3. **Map Dependencies:** Identify relationships between different components.
4. **Identify Risks:** Look for potential edge cases, security vulnerabilities, and performance bottlenecks.
5. **Produce a Report:** Summarize your findings in a structured report with citations.

## Few-Shot Mini Example (Generic)

Use this pattern as a guide:

```
Input (summary): Investigate checkout timeout spikes.
Output (shape):
- Sources: [service.ts:120-220], [queue.ts:40-88]
- Findings: timeout linked to retry storm when provider returns 429
- Open Questions: should retries be capped per tenant?
```

```
Input (summary): Investigate API auth intermittently returning 401.
Output (shape):
- Sources: [auth-client.ts:30-92], [token-cache.ts:10-67]
- Findings: stale token reused after refresh race under concurrency
- Open Questions: lock per key or singleflight cache?
```

```
Input (summary): Explore how feature flags flow through UI rendering.
Output (shape):
- Sources: [flags/provider.tsx:1-70], [dashboard/page.tsx:18-140]
- Findings: server flag preload exists, client fallback bypasses default guard
- Open Questions: should fallback be removed for deterministic SSR?
```

## Common Mistakes to Avoid
- ❌ Listing findings without evidence (file paths, line numbers)
- ❌ Making assumptions instead of reading actual code
- ❌ Ignoring edge cases or error handling patterns

## Output Format

```
## Research: [Topic]

### Sources Examined
- [file or URL]

### Findings
[Structured findings with citations]

### Open Questions
[Questions for the team]
```
