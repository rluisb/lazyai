# Process Audit: [###-audit-slug]

**Audit ID:** ###
**Date:** YYYY-MM-DD
**Period covered:** YYYY-MM-DD to YYYY-MM-DD
**Auditor:** [name or `process-audit` agent]
**Scope:** [workflow types audited — e.g., bugfix / SDD / housekeeping / all]
**Repos audited:** [list]

> **Purpose.** Measure how well the team and its agents followed the constitution and the workflow protocols during the period. Audits are **objective** — every finding cites evidence in `ledger.md`, the git history, or the artifacts. The output feeds `self-improve` and may amend standards.

---

## 1. Method

How the audit was conducted. Self-Consistency requires multiple passes.

- **Sources read:**
  - [ ] `.specify/memory/repos/<repo>/ledger.md` entries in period.
  - [ ] Git log + commit messages in period.
  - [ ] Active spec / plan / tasks artifacts.
  - [ ] PR descriptions + review comments.
- **Reasoning passes:**
  - [ ] Pass 1 — "did the workflow happen?" (presence check).
  - [ ] Pass 2 — "did the workflow happen *correctly*?" (quality check).
  - [ ] Pass 3 — "did the workflow improve over time?" (trend check).

---

## 2. Workflow Adherence

For each workflow type in scope, score adherence against the protocol fragments.

| Workflow | Runs in period | Followed protocol | Skipped phases | Sample evidence |
|---|---|---|---|---|
| `speckit-*` SDD | [N] | [M of N] | [list phases skipped, e.g., "no clarify"] | [ledger date or PR link] |
| `rpi` | [N] | [M of N] | [list] | [link] |
| `bugfix` | [N] | [M of N] | [list — e.g., "missing regression test"] | [link] |
| `spike` | [N] | [M of N] | [list — e.g., "single-path"] | [link] |
| `poc` | [N] | [M of N] | [list — e.g., "code shipped to main"] | [link] |
| `housekeeping` | [N] | [M of N] | [list] | [link] |
| `review` | [N] | [M of N] | [list] | [link] |

---

## 3. Constitutional Conformance

Per-article rate of compliance, period-wide.

| Article | Compliance | Evidence | Notes |
|---|---|---|---|
| I — Library-First | x / y | [PR samples] | [observations] |
| II — Test-First (NON-NEGOTIABLE) | x / y | [PR samples] | [observations] |
| III — Docs as Source of Truth | x / y | [doc-vs-code drift incidents] | [...] |
| IV — Anti-Speculation | x / y | [incidents of speculative code] | [...] |
| V — Simplicity | x / y | [Complexity Tracking entries unjustified] | [...] |
| VI — Anti-Overengineering (NON-NEGOTIABLE) | x / y | [violations of 30/300 rule, DRY-too-early] | [...] |

> Non-negotiable articles (II, VI) below 100% require an immediate corrective action — record it in §6.

---

## 4. Harness Engineering Adherence

Per the 5 harness rules.

| Rule | Adherence | Notable gaps |
|---|---|---|
| 1 — Feed Forward | [PASS / GAPS] | [examples — e.g., "tasks created before plan approved"] |
| 2 — The Contract | [PASS / GAPS] | [examples — e.g., "self-approved review"] |
| 3 — Feedback & Sensors | [PASS / GAPS] | [examples — e.g., "tests skipped"] |
| 4 — Memory & State | [PASS / GAPS] | [examples — e.g., "ledger entries missing for N PRs"] |
| 5 — Anti-Slope | [PASS / GAPS] | [examples — e.g., "PoC code merged"] |

---

## 5. Findings

Each finding cites evidence and proposes an action. Severity guide:

- 🔴 **Critical** — non-negotiable article violated (II or VI), or a Harness rule absent.
- 🟡 **Major** — recurring negotiable-article slip, or single Harness rule slip across multiple runs.
- 🟢 **Minor** — one-off lapse, easy to correct.

### 🔴 Critical
- [ ] **F-1 — [title]** — [evidence: ledger date / PR link]. **Action:** [proposed correction; deadline].

### 🟡 Major
- [ ] **F-2 — [title]** — [evidence]. **Action:** [...].

### 🟢 Minor
- [ ] **F-3 — [title]** — [evidence]. **Action:** [...].

---

## 6. Corrective Actions

Concrete, owned items that close the findings above.

| Finding | Action | Owner | Due | Status |
|---|---|---|---|---|
| F-1 | [action] | [owner] | YYYY-MM-DD | open |
| F-2 | [action] | [owner] | YYYY-MM-DD | open |

---

## 7. Standards & Constitution Amendments

Audits often surface missing or outdated rules.

- [ ] **Propose standard:** [pointer to new entry in `specs/standards/`].
- [ ] **Update standard:** [pointer to existing entry needing amendment].
- [ ] **Amend constitution:** [article + proposed change + ADR link].

---

## 8. Trend (compare with previous audit)

| Metric | Previous period | This period | Δ |
|---|---|---|---|
| Article II compliance | x% | y% | ±N% |
| Article VI compliance | x% | y% | ±N% |
| Ledger completeness | x% | y% | ±N% |
| Reviewer-rejected PRs | N | M | ±k |

---

## 9. Verdict

```
OVERALL:  HEALTHY / WATCH / DEGRADING
Critical findings: N
Major findings:    N
Minor findings:    N
Auditor:           [name or agent]
Date:              YYYY-MM-DD
```

---

## 10. Memory Update

- [ ] Append audit summary to `.specify/memory/repos/<repo>/ledger.md` (date, verdict, link).
- [ ] If trend is DEGRADING: trigger `self-improve` with this audit as input.
- [ ] If a corrective action involves a constitution amendment: open an ADR and link from §7.
