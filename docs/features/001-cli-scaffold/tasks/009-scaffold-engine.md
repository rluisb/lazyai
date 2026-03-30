# Task 009: Create Scaffold Engine

**Phase:** 3
**User Story:** US-1
**Status:** TODO
**Depends on:** T008

---

## Objective

Create the core scaffold engine that creates the docs/ structure and copies all shared (tool-agnostic) files. Does NOT handle tool-specific files — that's the adapters' job.

## Subtasks

- [ ] Create `src/scaffold.ts`
- [ ] `scaffoldDocs(targetDir)` — creates all docs/ subdirectories
- [ ] `copyTemplates(targetDir)` — copies library/templates/ → docs/templates/
- [ ] `copyRules(targetDir)` — copies library/rules/ → docs/rules/
- [ ] `copyContextFiles(targetDir)` — copies library/context/ → each docs/*/AGENTS.md
- [ ] `copyInfra(targetDir)` — copies CODEOWNERS, pre-commit, compliance.md, KNOWLEDGE_MAP
- [ ] `createRootFiles(targetDir, config)` — creates AGENTS.md + CLAUDE.md from template
- [ ] `writeConfig(targetDir, config)` — writes .ai-setup.json with file hashes
- [ ] `createStandardsDirs(targetDir)` — creates 8 empty standards category dirs
- [ ] Skip files that already exist (warn user)
- [ ] Wire into init command handler

## Files to Touch

| File | Action |
|------|--------|
| `src/scaffold.ts` | Create |
| `src/commands/init.ts` | Modify (wire scaffold) |

## Done When

- [ ] `init` creates full docs/ structure
- [ ] All shared files are copied
- [ ] AGENTS.md + CLAUDE.md created at root with placeholders
- [ ] .ai-setup.json written with file inventory
- [ ] Existing files are NOT overwritten
