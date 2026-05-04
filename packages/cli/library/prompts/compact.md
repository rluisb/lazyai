# Compact Prompt

**Topic:** [Compaction Topic]
**Goal:** [Compaction Goal]

---

## Instructions

0. **Think Step-by-Step (CoT):** Privately reason step-by-step to decide what to keep/drop, but output only the final compressed handoff.
1. **Preserve Objective & Scope:** Capture the current objective, scope boundaries, and constraints.
2. **Preserve Decisions:** Capture key decisions made so far and rationale.
3. **Preserve Assumptions & Confidence:** Capture active assumptions/unknowns and current confidence level (high/medium/low).
4. **Preserve Progress:** Capture current progress and the next immediate action.
5. **Drop Noise:** Remove redundant narrative and stale exploration details.
6. **Produce a Structured Handoff:** Output a concise compaction report for seamless continuation.

## Few-Shot Mini Example (Generic)

Use this pattern as a guide:

```
Input (summary): Multiple threads discussing API auth, retries, and unrelated UI tweaks.
Output (shape):
- Objective/Scope: auth + retries only
- Decisions: token refresh interval = 10m
- Next Action: implement retry backoff test
- Dropped Context: unrelated UI brainstorming
```

**Example 2: Mid-task compaction**
- Input: 40 exchanges covering auth research + partial plan draft
- Output:
  ```
  ## Objective
  Implement OAuth2 PKCE flow for mobile app

  ## Decisions & Rationale
  - Chose PKCE over implicit grant (security, RFC 7636 compliance)
  - Using passport.js with custom strategy (team familiarity)

  ## Key Findings (Research Phase — Complete)
  - Current auth: session-based, 3 files (auth.ts, session.ts, middleware.ts)
  - No existing OAuth support; passport.js already in deps
  - Mobile team expects /auth/callback endpoint

  ## In Progress (Planning Phase — 60%)
  - TechSpec drafted through Phase 1 (auth routes)
  - Phase 2 (token refresh) needs edge case analysis
  - Blocked: confirm token storage approach with mobile team

  ## Next Action
  Complete TechSpec Phase 2, then proceed to implementation
  ```

**Example 3: Session handoff (end of day)**
- Input: Full implementation session with 3 completed tasks
- Output:
  ```
  ## Objective
  Migrate payment webhooks from v1 to v2 API

  ## Completed
  - [x] Task 1: Created v2 webhook handler (src/webhooks/v2.ts)
  - [x] Task 2: Added signature verification (HMAC-SHA256)
  - [x] Task 3: Updated 12 tests, all passing

  ## Remaining
  - [ ] Task 4: Add retry logic for failed deliveries
  - [ ] Task 5: Update API docs + changelog

  ## Decisions
  - Kept v1 handler as fallback (delete after 30-day migration window)
  - Used crypto.timingSafeEqual for signature comparison (timing attack prevention)

  ## Risks & Unknowns
  - [CONFIDENCE: MEDIUM] v2 payload format change may affect 2 downstream consumers
  - Need to verify staging webhook URL with DevOps before Task 5
  ```

## Common Mistakes to Avoid
- ❌ Losing critical file paths or decision rationale during compaction
- ❌ Summarizing too aggressively — preserve exact values and specifics
- ❌ Dropping open questions or unresolved items

## Output Format

```
## Compaction: [Topic]

### Context Summary
- [file:line] — [description] — [proposal]

### Objective, Scope, Constraints
- Objective: [current objective]
- In Scope: [explicit in-scope items]
- Out of Scope: [explicit exclusions]
- Constraints: [key constraints]

### Decisions & Rationale
- [decision] — [rationale]

### Assumptions, Unknowns, Confidence
- Assumptions: [verified/unverified items]
- Unknowns: [open questions]
- Confidence: [high|medium|low] — [why]

### Progress & Next Action
- Completed: [what is done]
- Next Action: [immediate next step]

### Dropped Context
- [redundant/stale detail removed]
```
