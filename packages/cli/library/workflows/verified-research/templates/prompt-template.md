# Prompt Template — Verified Research

Drop this into a new session. Replace `{{placeholders}}` with your task specifics.

---

````markdown
# Research Task: {{topic}}

## What I need to know

{{specific questions, e.g. "Does X support Y natively?" "How do we migrate from A to B?"}}

## What I want as output

A unified report + supporting findings files in `{{output_path}}` with:
- TL;DR answering my questions directly
- Three viewpoints (internal code, source system, target system)
- Mapping table
- Recommendation
- Open questions for design phase

## How to run this

Apply the `verified-research` skill. Execute all 9 phases.

### Phase 0 — PLAN (do this first, BEFORE any tool calls)

1. Restate my questions in your own words. Confirm understanding.
2. Define three research tracks:
   - Track A — Internal codebase
   - Track B — Source/current system official docs
   - Track C — Target/proposed system official docs
3. **Before dispatching anything: search the team's existing knowledge base** (wiki, design docs, ticket tracker) for prior work on this topic. If existing designs or epics exist, read them first and scope the research tracks accordingly.
4. Show me the plan and the existing-work findings. Wait for approval.

### Phase 1 — DISPATCH (parallel)

Run all tracks concurrently as separate sub-agent calls in one message.

Each dispatch must include:
- Explicit mode/depth
- Explicit token or step budget (don't assume unlimited)
- Concrete search anchors (file paths, queries, URLs)
- Concrete deliverable path: `{{output_path}}/findings-{track}.md`
- Required sections in the findings file
- Constraints: read-only, cite source for every claim, distinguish verified vs inferred

### Phase 2 — HANDLE FAILURES GRACEFULLY

When a researcher hits step limits or read-only constraints block file writes:
- Don't discard the work. Capture what they found.
- Either re-dispatch with prior findings as seed, OR compile the file yourself if enough content exists.
- If read-only blocks them from writing, do the write yourself (one file only, no other edits).

### Phase 3 — SYNTHESIZE

Write the unified report combining all findings. Required sections:
- TL;DR (direct answers, 3–5 bullets)
- Three viewpoints (one section each)
- Mapping table (concept ↔ source ↔ target ↔ gap)
- Recommendation with explicit reasoning
- Implementation sketch (high-level only)
- Open questions

### Phase 4 — VERIFY (drift-scope) — DO NOT SKIP

1. Dispatch a verifier to check every claim in the internal-codebase findings against actual source.
2. Require line-level citations.
3. Mark each claim: `[VERIFIED]`, `[DRIFT: ...]`, or `[UNVERIFIED]`.
4. Apply drift corrections in place. Make them visible.
5. Add a drift-scope appendix.

### Phase 5 — VALIDATE AGAINST AUTHORITATIVE DOCS — DO NOT SKIP

1. Ask explicitly: is there a team-authored doc on this topic already?
2. If yes, fetch it and cross-reference.
3. Write a separate alignment file. Do NOT silently update prior work.
4. If the team doc is more authoritative, re-baseline (Phase 6).

### Phase 6 — RE-BASELINE IF NEEDED

If a more authoritative source is found:
1. Write a new primary document with the authoritative doc as structural backbone.
2. Keep the original synthesis for context — don't overwrite.
3. Mark which doc is now primary.

### Phase 7 — ADVERSARIAL REVIEW

Before publishing a recommendation:
1. Frame Advocate vs Skeptic roles.
2. Run 3+ rounds covering architectural assumption, operational risk, silent failure.
3. Synthesize resolutions per round.
4. Carry unresolved tensions into open questions.

### Phase 8 — CONTRIBUTE BACK (if applicable)

1. Match the existing doc's tone (quote 1–2 sentences to anchor).
2. Append-only — never change, contradict, or delete.
3. Drop-in content + rationale provided separately.
4. Suggest a review path.
5. List open questions for the reviewer.

## Constraints (read carefully)

- Plan before researching.
- Search existing work BEFORE dispatching researchers.
- Parallelize independent tracks.
- Cite source for every claim (file:line for code, exact URL for docs).
- Distinguish verified, inferred, and assumed.
- Never skip verification (Phase 4).
- Never skip cross-reference (Phase 5).
- Never silently rewrite when re-baselining.
- Append-only when contributing to others' docs.

## Output locations

- Planning artifacts: `{{output_path}}`
- Durable knowledge (if applicable): `{{vault_path}}`
- Do NOT modify code in this task.

## Pre-flight gate

Before declaring "done," run the validation checklist at:
`~/.config/opencode/skills/verified-research/templates/checklist.md`

Every box must be ticked.
````
