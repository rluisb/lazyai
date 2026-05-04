# Standard: Agent Security

**Category:** Security
**Scope:** project
**Date:** 2026-05-01
**Owner:** AI Setup Maintainers
**Status:** Active
**Constitution article(s):** I, III, IV, VI

> **Purpose.** Ensure AI agents operate within explicit authority boundaries, protect secrets, and make risky actions visible before they happen.

---

## Scope Cascade

This starter standard is written at project scope and applies to agent prompts, commands, tools, workflows, and task handoffs until a more specific security standard replaces it.

```
global standards
        ↓
workspace standards
        ↓
project standards
        ↓
agent execution
```

**Where this standard lives:** project

**Overrides:** none — this is starter guidance for new projects.

---

## Rule

Agents MUST declare write authority, avoid secret material, and obtain explicit approval before destructive, credential-bearing, or remote-changing actions.

**Trigger:** Granting tool access, editing files, reading credentials, running migrations, pushing to remotes, or invoking external services.

---

## Rationale

Agent work is safest when authority is narrow and visible. Security review must be able to trace what an agent was allowed to read, write, or execute.

- **Article support:** Article I keeps reusable authority boundaries in library instructions; Article III makes approvals part of the task record; Article IV and VI prevent broad speculative automation.
- **Origin:** Wave 1 starter standards for concrete pattern review.

---

## Examples

**Compliant:**
```
Scope: edit only library/standards/starter/*.md.
Do not read .env files. Do not push. Ask before changing migrations.
```

**Non-compliant:**
```
Use any available tool to finish the task, including credentials and remote updates if needed.
```

**Why the non-compliant case fails:** It grants authority that is broader than the task and hides high-risk actions from review.

---

## Enforcement

| Mechanism | Where | When |
|---|---|---|
| Task scope review | Task file or user request | Before tool use |
| Secret-file exclusion | Git status and review checklist | Before commit or handoff |
| Explicit approval record | Conversation, PR, or handoff | Before destructive or remote-changing action |

---

## Exceptions

Exceptions require explicit human approval in the same task context and must name the risky action being authorized.

- Incident response — emergency actions require post-action documentation and owner review.

---

## Workspace Awareness

| Repo | Override | Reason |
|---|---|---|
| All repos | none | Starter guidance applies until repo-specific agent-security standards exist. |

---

## Related

- **Supersedes:** none
- **Related ADR(s):** none
- **Related standards:** orchestration-patterns, context-loading

---

## Memory

- [ ] Standard published at the declared scope.
- [ ] Security exceptions link to the approval and owner.
- [ ] Repeated approval patterns are considered for a dedicated security rule.
