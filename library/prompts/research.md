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
