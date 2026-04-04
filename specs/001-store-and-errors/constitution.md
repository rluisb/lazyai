# Constitution: Store & Error Handling (001)

> Governing principles and non-negotiables for the lowdb/zod state management and structured error handling additions to `@ricardoborges-teachable/ai-setup`.

---

## 1. Scope

This constitution covers two tightly coupled subsystems:

| Subsystem | What | Why |
|-----------|------|-----|
| **Store** | Replace raw `JSON.parse`/`writeFileSync` config handling with lowdb v7 + zod validated state | Current config has no validation, no migration path, no file-status tracking, no operation history |
| **Errors** | Replace ad-hoc `throw new Error(string)` / `process.exit(1)` with typed error hierarchy and single boundary | Current error handling is scattered across 10+ files with 6 distinct patterns, making recovery and user messaging impossible |

### Out of scope

- New CLI commands (no new user-facing commands added)
- Changes to the `library/` bundled content or adapter installation logic
- Interactive prompt flow changes (wizard phases stay the same)
- CI/CD or publishing changes

---

## 2. Principles

### P1: Zod is the single source of truth for types

All store-related TypeScript types **must** be derived via `z.infer<>` from zod schemas. No parallel `interface` definitions. The zod schema IS the type definition, the runtime validator, and the migration target.

**Rationale:** Duplicate type definitions drift. Zod eliminates the class of bugs where runtime data doesn't match compile-time types.

### P2: Forward-only, silent migrations

Legacy `.ai-setup.json` files (v0 format, no `meta.schemaVersion`) **must** auto-migrate to v1 on first read. Migration is silent (no interactive prompts, no user confirmation). Migration is forward-only (no downgrade path).

**Rationale:** Users run `ai-setup update` or `ai-setup doctor` after upgrading the CLI. The store must "just work" — any migration friction would break the one-command promise.

### P3: Same file, new structure

The store file remains `.ai-setup.json`. lowdb writes to the same path. The filename is part of the public contract (documented, referenced in `.gitignore` patterns, used by `eject`).

**Rationale:** Changing the filename would orphan existing installations and break user gitignore configs.

### P4: Errors carry codes, not just messages

Every error thrown by ai-setup **must** be an `AiSetupError` with a typed `ErrorCode`. String-only errors are forbidden in new code. The error code determines user-facing behavior (message format, exit code, whether to show stack trace).

**Rationale:** Programmatic error handling requires codes. User-facing CLI tools need to distinguish "you cancelled" from "filesystem permission denied" to give appropriate feedback.

### P5: Single exit point

All errors flow through one `handleError()` boundary. No `process.exit()` calls outside the entry point (`src/index.ts`). Commands throw; the boundary catches and formats.

**Rationale:** Scattered `process.exit()` calls (currently in 6 files) prevent cleanup, make testing impossible (vitest can't catch `process.exit`), and create inconsistent user experiences.

### P6: Tests never touch the filesystem for store operations

All store tests use lowdb's `Memory` adapter. Filesystem tests (for `src/utils/files.ts`) remain in their own test files with temp directories.

**Rationale:** Fast, deterministic, parallelizable tests. No cleanup. No cross-test contamination.

### P7: Operation log is append-only and bounded

The operations log is capped at 50 entries, oldest-first eviction. Operations are never mutated after creation. The log exists for diagnostics, not undo.

**Rationale:** Unbounded logs in a JSON file would grow forever. 50 entries provides sufficient history for `doctor` diagnostics without bloating the store file.

### P8: Backward compatibility is non-negotiable

Every existing command (`init`, `add`, `update`, `doctor`, `create`, `eject`) **must** continue to work identically from the user's perspective after this change. Existing `.ai-setup.json` files from v0.1.0 **must** be readable.

**Rationale:** This is infrastructure, not a feature. Users must not notice the change except through improved error messages.

---

## 3. Decisions

| # | Decision | Alternatives Considered | Rationale |
|---|----------|------------------------|-----------|
| D1 | lowdb v7 with JSONFile adapter | conf, configstore, plain fs | ESM-native, typed generics, atomic writes via steno, Memory adapter for tests, minimal API surface |
| D2 | zod for schema validation | io-ts, typebox, ajv | Best DX for TypeScript (z.infer), runtime + compile-time in one definition, excellent error messages, widely adopted |
| D3 | Operation log capped at 50 | Unlimited, separate log file, no log | 50 entries ≈ 2-3 months of typical usage; JSON stays small; sufficient for `doctor` diagnostics |
| D4 | `AiSetupError` extends `Error` | Custom base class, result types | Extends Error for compatibility with existing catch blocks and stack traces; adds `code` and `context` for programmatic handling |
| D5 | `ErrorCode` as string enum (not numeric) | Numeric codes, symbol-based | String codes are self-documenting in logs and error messages; no lookup table needed |
| D6 | `handleError()` as function, not class | ErrorBoundary class, middleware pattern | Single function is simpler; no state needed; aligns with Commander's action-based flow |
| D7 | `TrackedFile.status` field added | Compute status on read only | Pre-computed status enables fast `doctor` checks and `status` command without re-hashing every file |
| D8 | DEBUG mode via both `AI_SETUP_DEBUG=1` and `--verbose` | Only env var, only flag | Env var works for piped commands; flag works for interactive use; both are standard CLI conventions |

---

## 4. Constraints

| ID | Constraint | Source |
|----|-----------|--------|
| C1 | ESM-only — no CommonJS | `package.json` `"type": "module"`, lowdb v7 requirement |
| C2 | Node >= 18 | `package.json` `engines` field |
| C3 | `.ai-setup.json` filename unchanged | Public contract (see P3) |
| C4 | Zero new runtime deps beyond lowdb and zod | Keep install footprint small for a scaffold CLI |
| C5 | All quality gates must pass: `npm run typecheck`, `npm run test`, `npm run build` | Existing CI contract |
| C6 | No breaking changes to existing command behavior | P8 |
| C7 | Wizard phases (1-8) remain structurally unchanged | Wizard is well-tested; changes are to plumbing, not UX |

---

## 5. Assumptions

| # | Assumption | Status | Notes |
|---|-----------|--------|-------|
| A1 | lowdb v7 `JSONFile` adapter uses steno for atomic writes | **Accepted** | Verified in lowdb v7 source |
| A2 | lowdb v7 `Memory` adapter API is compatible with `JSONFile` for testing | **Accepted** | Same `Low` class, different adapter |
| A3 | Existing `.ai-setup.json` files always have the shape of `AiSetupConfig` from `src/types.ts` | **Accepted** | Only one writer exists (wizard `index.ts` line 208-218) |
| A4 | `selections` field may be absent in older manifests (pre-wizard) | **Accepted** | Type has `selections?: WizardSelections` (optional) |
| A5 | The `version` field in current manifest is CLI version, not schema version | **Accepted** | Written as `'0.1.0'` — CLI package version |
| A6 | No external tools read `.ai-setup.json` programmatically | **Pending** | If external tools depend on the exact shape, migration could break them |
| A7 | `process.exit(0)` on user cancel is acceptable behavior | **Accepted** | Standard CLI convention; `handleError` will preserve this |

---

## 6. Risks

| # | Risk | Severity | Mitigation |
|---|------|----------|------------|
| R1 | Legacy manifest detection false-positive (new-format file without `meta` field) | **Low** | `isLegacyFormat()` checks for absence of `meta.schemaVersion` AND presence of top-level `version` string |
| R2 | lowdb v7 atomic write failure corrupts store | **Low** | steno writes to temp file then renames; OS-level atomic. Test with `Memory` adapter eliminates this in tests |
| R3 | zod validation rejects valid legacy data during migration | **Medium** | Migration transforms data BEFORE validation. Migration tests cover all known legacy shapes |
| R4 | Refactoring `process.exit()` out of commands breaks cancel behavior | **Medium** | `AiSetupError` with `USER_CANCELLED` code + `exitCode: 0` preserves exact behavior |
| R5 | Bundle size increase from zod (~13KB min) | **Low** | Acceptable for a CLI tool; zod is tree-shakeable for the subset used |
| R6 | Breaking change if `AiSetupConfig` type is imported by external consumers | **Low** | Package is `v0.1.0` — semver allows breaking changes. Types will be re-exported from zod schemas |
