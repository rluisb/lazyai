---
description: Run tests and summarize failures
argument-hint: "[package-or-file]"
allowed-tools: Bash Read
---

Run the project's test suite (or a specific target) and summarize results.

Steps:
1. Detect the test command from the project (CLAUDE.md / AGENTS.md, or infer from package.json, Cargo.toml, go.mod, etc).
2. If $ARGUMENTS provided, scope to that package/file (e.g., `go test ./path -count=1`).
3. Run the test command. Capture exit code, output, and failures.
4. Extract failure summary: test name, file:line, and first meaningful error line.

Report format:
- ✅ `<suite>` — N passing, 0 failing (exit 0), **or**
- ❌ `<suite>` — N passing, M failing (exit X):
  - `pkg.TestFoo` (file.go:42) — short reason
  - ...

Do not attempt fixes unless the user explicitly asks.
