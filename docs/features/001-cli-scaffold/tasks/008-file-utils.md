# Task 008: Create File Utilities

**Phase:** 3
**User Story:** US-1
**Status:** TODO
**Depends on:** T007

---

## Objective

Create utility functions for file operations: copy, write, hash, exists, directory creation, template rendering.

## Subtasks

- [ ] Create `src/utils/files.ts`
- [ ] `copyFile(src, dest)` — copy file, create parent dirs if needed
- [ ] `copyDir(src, dest)` — recursive directory copy
- [ ] `writeFile(dest, content)` — write with parent dir creation
- [ ] `fileExists(path)` — check existence
- [ ] `fileHash(path)` — MD5 or SHA256 for comparison
- [ ] `readFile(path)` — read to string
- [ ] `ensureDir(path)` — create directory recursively
- [ ] All functions handle errors gracefully with clear messages

## Files to Touch

| File | Action |
|------|--------|
| `src/utils/files.ts` | Create |

## Done When

- [ ] All utility functions work
- [ ] Error messages are clear ("Cannot write to X: permission denied")
- [ ] Parent directories are created automatically
