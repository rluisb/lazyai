# Task 014: Implement Update Command

**Phase:** 5
**User Story:** US-4
**Status:** TODO

## Objective
Implement the `update` command to refresh library files without overwriting user customizations.

## Subtasks
- [ ] Read `.ai-setup.json` to get original file hashes
- [ ] Compare current file hashes against original
- [ ] Only overwrite files where hashes match (i.e. user hasn't modified them)
- [ ] Add new files from library that didn't exist before
- [ ] Update `.ai-setup.json` with new hashes
