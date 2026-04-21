---
description: Run the project's tests and summarize failures
---

Run the repository's test suite and summarize the result.

Steps:

1. Detect the test command from the project (e.g. `go test ./... -count=1`,
   `pnpm test`, `pytest`, `cargo test`). Prefer the command declared in
   `AGENTS.md` / `CLAUDE.md` if present.
2. Run it. Capture exit code and failing-test output.
3. If anything failed, extract: test name, file:line, and the first useful
   stderr/stdout line per failure. Skip stack trace noise.

Report:

- ✅ `<suite>` — N passing, 0 failing (exit 0), **or**
- ❌ `<suite>` — N passing, M failing (exit X):
  - `pkg.TestFoo` (file.go:42) — short reason
  - ...

Do not attempt fixes unless the user explicitly asks.
