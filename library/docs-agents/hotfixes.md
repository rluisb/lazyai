<rule>
  <scope>auto</scope>
  <globs>docs/hotfixes/**</globs>
  <description>Hotfix workflow — P0/P1 production emergency fast-track with mandatory post-mortem</description>
</rule>

# Hotfix Workflow Rules

## What This Flow Is For

Production is broken right now. This is the P0/P1 emergency track.
It is intentionally leaner than the standard bugfix flow — but gates are NOT optional.

---

## Directory Structure

```
docs/hotfixes/NNN-hotfix-name/
├── techspec.md          ← Abbreviated RCA (use bugfix-rca-template.md)
├── progress.md          ← Trace log
└── postmortem.md        ← MANDATORY — due within 24h of deploy
```

---

## Flow

### 1. Reproduce — Confirm the Bug
- Get a reproduction case before writing any code
- If you cannot reproduce: escalate to human, do not guess
- ⛔ **HUMAN GATE:** Confirm reproduction + severity classification (P0 or P1)

### 2. RCA — Root Cause Analysis
- Map blast radius: what else could be affected?
- Write abbreviated `techspec.md` using `docs/templates/bugfix-rca-template.md`
- Answer: what caused it, what the fix is, and why the fix is safe
- ⛔ **HUMAN GATE:** Confirm root cause + fix scope before implementation

### 3. Implement — Minimal Fix
- Write the smallest change that resolves the root cause
- Add regression test — non-negotiable even in emergencies
- ⛔ **CHECKPOINT:** regression test passes, existing tests pass

### 4. Expedited Review
- Single-pass Reviewer focused on: correctness, blast radius, rollback safety
- Review ONLY the hotfix — no other feedback scope
- Output: APPROVE or REQUEST_CHANGES (no COMMENT-only for hotfixes)

### 5. Deploy + Monitor
- Human deploys — AI does not trigger deployments
- Monitor for 30 min post-deploy (human responsibility)

### 6. Post-Mortem — MANDATORY
- Due within 24h of confirmed resolution
- Use `docs/templates/postmortem-template.md`
- Output: `docs/hotfixes/NNN-hotfix-name/postmortem.md`
- No exceptions — every P0/P1 gets a post-mortem

---

## Hotfix Principles

- **Speed, but not recklessness.** Gates compress, not disappear.
- **Minimum viable fix.** Make it revert-friendly.
- **No scope creep during an incident.** Fix the incident. Refactor later.
- **Post-mortem is part of the fix.** Prevention is the second half of the work.

---

## Rollback Plan

Before implementing, state the rollback:
- Can this be feature-flagged? → Prefer flags over deploys
- What is the revert command?
- Who has deploy access?

---

## Self-Improvement — After Every Hotfix

- Post-mortem filed? → Confirm before closing session
- Prevention actions identified? → Create follow-up tickets
- New rule needed? → Flag `docs/rules/` update
- Pattern revealed? → Write memory note to `docs/memory/`
