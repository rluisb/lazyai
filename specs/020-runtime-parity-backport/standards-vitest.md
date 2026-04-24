# Testing Standards — Vitest Configuration

**Applies to:** TypeScript test suite (`src/__tests__/*.test.ts`) run under `vitest`.

---

## Global setup file

All vitest runs load `src/__tests__/setup.ts` via `vitest.config.ts#test.setupFiles`.
This file is the single source of truth for global test-environment side effects.

**Rule:** Side effects needed by every test (env vars, global mocks, timezone pinning, etc.) go in `setup.ts`. Never duplicate them across individual test files — duplication drifts and creates order-dependent flakes.

## Environment variables

Tests that invoke user-facing install/scaffold flows must NOT trigger real CLI probes against externally-installed binaries.

### `AI_SETUP_SKIP_VALIDATION=1`

Skips the post-install `opencode debug config` / `opencode debug agent` probes in `validateOpenCodeInstall`. Set globally in `setup.ts`; real users never set it.

- **Set by:** `src/__tests__/setup.ts` globally
- **Honored by:** `src/adapters/opencode-validate.ts` (TS) and should be honored by the Go equivalent in `internal/adapter/opencode_validate.go` if/when a Go test exercises the same path
- **Why:** `opencode debug config` takes ~1.5s per call. With ~20 wizard-invoking tests × ~10 agents each, unshielded tests balloon the suite from 5s to 10+ min on any machine with opencode installed

Tests that specifically validate the probe behavior (e.g. `opencode-validate.test.ts`) must unset this flag in `beforeEach` and restore it in `afterEach`.

## Injectable command runners

Any adapter that shells out to a tool CLI must expose an injectable runner so tests can avoid real subprocess calls.

### Pattern

```typescript
export type CmdRunner = (command: string, args: string[]) => string

const defaultCmdRunner: CmdRunner = (command, args) =>
  execFileSync(command, args, { encoding: 'utf-8' })

export function myOperation(ctx: Ctx, runner: CmdRunner = defaultCmdRunner): Result {
  // ... uses runner() instead of exec directly
}
```

### Why

- Tests don't depend on the tool being installed locally
- Tests run in milliseconds instead of seconds
- Failure modes (`throw new Error(...)`) are controllable per-test
- No need for process mocking frameworks

### Where this lives

- `src/adapters/opencode-validate.ts#CmdRunner` — the canonical TS example
- `internal/adapter/opencode_validate.go#CmdRunner` — the canonical Go example (mirror)

Any new adapter that invokes external CLIs must follow this pattern in both runtimes.

## PATH-scoping for binary-presence checks

When testing code that calls `command -v <name>` or similar:

- Use `process.env.PATH = binDir` where `binDir` contains a fake executable, OR `process.env.PATH = ''` to force the absent path
- Restore original `PATH` in `afterEach` — vitest does not isolate env between tests
- Do NOT use `which <name>` in production code: it fails when PATH is restricted (the `which` binary itself becomes unreachable). Use `command -v <name>` instead — it's a POSIX shell builtin

## Temp directories

Standard pattern for tests needing isolated filesystem state:

```typescript
import { mkdtempSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'

const tempDirs: string[] = []

const makeTempDir = (prefix: string): string => {
  const dir = mkdtempSync(path.join(tmpdir(), prefix))
  tempDirs.push(dir)
  return dir
}

afterEach(() => {
  for (const dir of tempDirs) {
    rmSync(dir, { recursive: true, force: true })
  }
  tempDirs.length = 0
})
```

Don't use `t.TempDir()`-style APIs — vitest doesn't provide them. The array-and-afterEach pattern above is the standard in this repo.

## Cross-runtime parity for tests

When porting Go tests to TS (or vice versa), keep the test case names and assertion structure aligned so failures in one runtime point clearly to the equivalent case in the other. Example:

| Go | TS |
|---|---|
| `TestValidateOpenCodeInstall_BinaryAbsent` | `'returns no warnings when the opencode binary is absent'` |
| `TestValidateOpenCodeInstall_ConfigFails` | `'emits a config warning when opencode debug config fails'` |
| `TestMergeJSONFile_Idempotent` | `'is idempotent — re-run with same patch produces identical bytes; .bak never overwritten'` |

Prose naming is fine in TS, but the scenario ordering and assertion targets should match the Go file for easy side-by-side review.
