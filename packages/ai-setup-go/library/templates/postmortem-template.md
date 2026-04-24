<template>
  <id>postmortem</id>
  <version>1.0</version>
  <scope>hotfixes</scope>
  <trigger>Every P0/P1 hotfix. Mandatory. Complete within 24 hours of resolution.</trigger>
  <outputs>postmortem.md in hotfix dir</outputs>
</template>

# Post-Mortem: [TICKET] — [Short Incident Title]

> **Severity:** P0 / P1  
> **Status:** Resolved  
> **Incident Date:** YYYY-MM-DD  
> **Resolution Date:** YYYY-MM-DD  
> **Total Duration:** [X hours Y minutes]  
> **Author:** [name]  
> **Review Date:** [date — within 24h of resolution]

---

## 1. Incident Summary

> One paragraph. What broke, who was affected, what the immediate impact was, and how it was resolved. No blame. Facts only.

---

## 2. Impact

| Dimension | Detail |
|---|---|
| **Users affected** | [count or %] |
| **Services affected** | [list] |
| **Data loss** | Yes / No |
| **Revenue impact** | [estimate if known, or N/A] |
| **SLA breach** | Yes / No |

---

## 3. Timeline

> UTC timestamps. Be precise. Include discovery, escalation, diagnosis, mitigation, and resolution.

| Time (UTC) | Event |
|---|---|
| HH:MM | Incident detected [by whom / how] |
| HH:MM | On-call notified |
| HH:MM | Initial diagnosis |
| HH:MM | Mitigation applied |
| HH:MM | Incident resolved |
| HH:MM | Post-mortem initiated |

---

## 4. Root Cause Analysis (5-Why)

> Drill to the systemic cause, not just the proximate cause. Stop when you reach a process or system gap you can actually fix.

**What broke:**  
[Concise technical description of the failure]

| # | Why? | Answer |
|---|---|---|
| 1 | Why did [symptom] occur? | [answer] |
| 2 | Why did [answer 1] happen? | [answer] |
| 3 | Why did [answer 2] happen? | [answer] |
| 4 | Why did [answer 3] happen? | [answer] |
| 5 | Why did [answer 4] happen? | [root cause] |

**Root Cause:**  
[One sentence. The systemic gap that enabled this incident.]

---

## 5. Contributing Factors

> Additional conditions that made the incident worse or harder to detect. Not the root cause — the amplifiers.

- [ ] Missing monitoring / alerting
- [ ] Insufficient test coverage
- [ ] Lack of runbook / playbook
- [ ] Knowledge silos
- [ ] Time pressure / deployment velocity
- [ ] Other: ___

---

## 6. Resolution

**What was done to stop the bleeding:**  
[Immediate mitigation — rollback, feature flag, hotfix deployed, etc.]

**Permanent fix:**  
[What was implemented as the durable solution — reference TICKET and PR]

**Rollback available:** Yes / No  
**Rollback exercised:** Yes / No

---

## 7. Prevention — Action Items

> Concrete, owned, time-bound actions. Not vague resolutions.

| # | Action | Owner | Due | Status |
|---|---|---|---|---|
| 1 | [What specifically will be done] | [name] | YYYY-MM-DD | Open |
| 2 | [e.g., Add alert for X] | [name] | YYYY-MM-DD | Open |
| 3 | [e.g., Write runbook for Y] | [name] | YYYY-MM-DD | Open |

---

## 8. What Went Well

> Honest credit. What detection, response, or collaboration worked better than expected?

- 
- 

---

## 9. What Could Be Better

> Honest critique. What slowed resolution down? No blame — focus on systems and process.

- 
- 

---

## 10. Follow-Up Tasks

> Ticket references for work that can't be done immediately but must be tracked.

- [ ] [TICKET] — [description]
- [ ] [TICKET] — [description]

---

*Post-mortem is blameless. The goal is systemic improvement, not individual accountability.*
