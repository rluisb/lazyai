# Task 005: Copy Templates, Rules, and Context Files to Library

**Phase:** 2
**User Story:** US-1
**Status:** TODO
**Depends on:** T003
**Parallel with:** T004, T006

---

## Objective

Copy all document templates, baseline rules, and AGENTS.md context files into the library.

## Subtasks

- [ ] Create `library/templates/` — copy all 9 template files (prd, techspec, tasks, task, adr, tech-debt, standard, progress, AGENTS.md)
- [ ] Create `library/rules/` — copy baseline rules (workflow.md, security.md, review.md, cost.md)
- [ ] Create `library/context/` — copy all 10 AGENTS.md context files (docs, rules, standards, templates, memory, adrs, features, bugfixes, refactors, tech-debt)
- [ ] Verify file counts match Templates directory

## Files to Touch

| File | Action |
|------|--------|
| `library/templates/*.md` | Create (copy 9 files) |
| `library/rules/*.md` | Create (copy 4 files) |
| `library/context/*.md` | Create (copy 10 files) |

## Done When

- [ ] 9 template files in `library/templates/`
- [ ] 4 rule files in `library/rules/`
- [ ] 10 context files in `library/context/`
