# Standard: [Standard Name]

**Category:** Code | Process | Security | Testing | Architecture
**Scope:** global | workspace | project *(see "Scope cascade" below)*
**Date:** YYYY-MM-DD
**Owner:** [team]
**Status:** Draft | Active | Deprecated | Superseded by `[pointer]`
**Constitution article(s):** [I-VI as applicable]

> **Purpose.** A standard is a durable, named rule that applies whenever a workflow encounters its trigger. Standards are the project's long-term memory of "we already decided this" — appealing to a standard is faster and more reliable than re-deriving the rule each time.

---

## Scope Cascade

A standard is written at exactly one scope. The cascade lets specific scopes override general ones.

```
global standards (~/.claude/, ~/.specify/)
        ↓ (default)
workspace standards (<workspace>/specs/standards/)
        ↓ (overrides for this team / org)
project standards (<repo>/specs/standards/)
        ↓ (overrides for this codebase)
code
```

**Where this standard lives:** [global / workspace / project]

**Overrides:** [pointer to a higher-scope standard this overrides, OR "none — this is the canonical scope"]

> If a standard is canonical at a higher scope, it MUST NOT be re-stated at a lower scope unless the lower scope has a different rule. Duplication invites drift.

---

## Rule

[The standard, in one clear, testable sentence. The reader can decide compliance from the rule alone.]

**Trigger:** [the situation in which the rule applies]

**Examples of triggers:**
- A new function is written.
- A new dependency is added to `package.json`.
- A new public API is exposed.

---

## Rationale

Why this rule exists. Cite the constitution article it operationalizes, and any incident that motivated it.

- **Article support:** [Article number + reasoning]
- **Origin:** [PR / incident / ADR that prompted this rule, if known]

---

## Examples

**Compliant:**
```
[example of code or behavior that follows the rule]
```

**Non-compliant:**
```
[example of code or behavior that violates the rule]
```

**Why the non-compliant case fails:** [one sentence].

---

## Enforcement

How this rule is checked. Prefer automated enforcement; fall back to review only when automation is impractical.

| Mechanism | Where | When |
|---|---|---|
| Linter rule / type-check | [config path] | Gate 1 (Static Integrity) |
| Test | [test path] | Gate 3 (Behavioral Validation) |
| Reviewer checklist | [pointer] | Gate 4 (Pattern Consistency) |
| Human review | [pointer] | Gate 4 |

---

## Exceptions

When may this rule be waived? Exceptions MUST be explicit (PR description note + reviewer approval). "Default deny" — silence is not an exception.

- [exception class] — requires [approver] sign-off.

---

## Workspace Awareness

If this standard differs by repo within the workspace, list overrides.

| Repo | Override | Reason |
|---|---|---|
| [repo] | [different rule or "exempt"] | [reason] |

---

## Related

- **Supersedes:** [pointer if this replaces a prior standard]
- **Related ADR(s):** [pointer]
- **Related standards:** [pointers]

---

## Memory

- [ ] Standard published at the declared scope.
- [ ] If a violation slipped past Gate 4 in the past, the relevant ledger entry references this standard.
- [ ] If this standard's enforcement mechanism is added later, the date and PR are recorded above under "Enforcement."
