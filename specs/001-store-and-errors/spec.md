# Spec: Store & Error Handling (001)

> Technical specification for lowdb v7 + zod state management and structured error handling in `@rluisb/lazyai`.

---

## Table of Contents

1. [Store Subsystem](#1-store-subsystem)
2. [Error Handling Subsystem](#2-error-handling-subsystem)
3. [Migration Strategy](#3-migration-strategy)
4. [File-by-File Change Map](#4-file-by-file-change-map)
5. [Testing Strategy](#5-testing-strategy)
6. [Acceptance Criteria](#6-acceptance-criteria)
7. [Implementation Priority](#7-implementation-priority)
8. [Open Questions](#8-open-questions)

---

## 1. Store Subsystem

### 1.1 New Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `lowdb` | `^7.0.0` | ESM-native JSON database with atomic writes via steno, typed generics, Memory adapter for tests |
| `zod` | `^3.23.0` | Runtime schema validation, type derivation via `z.infer`, migration target definitions |

Both are **runtime** dependencies (added to `dependencies` in `package.json`).

### 1.2 Store Schema (v1)

The store schema defines the complete shape of `.ai-setup.json` under the new system. All types are derived from zod schemas via `z.infer<>`.

#### Top-level structure

```
StoreSchema {
  meta: MetaSchema
  config: ConfigSchema
  selections: WizardSelectionsSchema
  files: TrackedFileSchema[]
  sync: SyncSchema
  operations: OperationSchema[]   // capped at 50, oldest-first eviction
}
```

#### `meta` — Store metadata

| Field | Type | Description |
|-------|------|-------------|
| `schemaVersion` | `number` | Current value: `1`. Incremented on breaking schema changes. Used by migration system to determine which transforms to apply. |
| `cliVersion` | `string` | Package version at time of last write (e.g., `"0.2.0"`) |
| `installedAt` | `string` (ISO 8601) | Timestamp of first `init` |
| `lastUpdatedAt` | `string` (ISO 8601) | Timestamp of most recent write |

#### `config` — Setup configuration

| Field | Type | Description |
|-------|------|-------------|
| `setupType` | `'project' \| 'workspace'` | Same as current `SetupType` |
| `tools` | `ToolId[]` | Same as current |
| `projectName` | `string` | Same as current |
| `targetDir` | `string` | Absolute path to project root |

#### `selections` — Wizard selections

Same shape as current `WizardSelections` interface. All fields required (empty arrays for unselected categories).

| Field | Type |
|-------|------|
| `docsDirs` | `DocsDirId[]` |
| `docsAgents` | `DocsDirId[]` |
| `templates` | `TemplateId[]` |
| `rules` | `RuleId[]` |
| `agents` | `AgentId[]` |
| `skills` | `SkillId[]` |
| `prompts` | `PromptId[]` |
| `infra` | `InfraId[]` |

#### `files` — Tracked file records (enhanced)

| Field | Type | Description |
|-------|------|-------------|
| `path` | `string` | Relative path from project root (unchanged) |
| `hash` | `string` | SHA-256 truncated to 16 chars (unchanged) |
| `source` | `string` | Library-relative source path (unchanged) |
| `status` | `'installed' \| 'modified' \| 'missing' \| 'conflict'` | **New.** Computed on `doctor`/`update`, persisted for fast `status` command |
| `installedAt` | `string` (ISO 8601) | **New.** When this file was first installed |
| `lastCheckedAt` | `string` (ISO 8601) | **New.** When status was last verified |

#### `sync` — Sync tracking

| Field | Type | Description |
|-------|------|-------------|
| `lastSyncAt` | `string \| null` | ISO 8601 timestamp of last successful sync, or `null` if never synced |
| `dirty` | `boolean` | `true` if store has changes since last sync |

#### `operations` — Append-only operation log

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Unique operation ID (e.g., `op_<timestamp>_<random>`) |
| `type` | `'init' \| 'add' \| 'update' \| 'doctor' \| 'eject' \| 'create'` | Which command triggered this operation |
| `timestamp` | `string` (ISO 8601) | When the operation occurred |
| `filesAffected` | `string[]` | Relative paths of files touched |
| `result` | `'success' \| 'partial' \| 'failed'` | Outcome |
| `rollback` | `string[]` (optional) | Paths of backup files created, for potential rollback |

**Cap:** Maximum 50 entries. When adding entry 51, remove the oldest entry first.

### 1.3 File Structure

```
src/store/
  schema.ts      — Zod schemas, exported types, CURRENT_SCHEMA_VERSION, defaultStore()
  index.ts       — createStore(targetDir), createTestStore(), read/write with auto-migration
  migrations.ts  — isLegacyFormat(), v0→v1 migration, migration registry
```

### 1.4 Public API

#### `src/store/schema.ts`

**Exports:**

| Export | Kind | Description |
|--------|------|-------------|
| `CURRENT_SCHEMA_VERSION` | `const number` | Current schema version (1) |
| `MetaSchema` | `z.ZodObject` | Zod schema for `meta` |
| `ConfigSchema` | `z.ZodObject` | Zod schema for `config` |
| `WizardSelectionsSchema` | `z.ZodObject` | Zod schema for `selections` |
| `TrackedFileSchema` | `z.ZodObject` | Zod schema for individual tracked file |
| `SyncSchema` | `z.ZodObject` | Zod schema for `sync` |
| `OperationSchema` | `z.ZodObject` | Zod schema for individual operation |
| `StoreSchema` | `z.ZodObject` | Top-level store schema |
| `type StoreData` | `z.infer<typeof StoreSchema>` | Inferred TypeScript type |
| `type TrackedFile` | `z.infer<typeof TrackedFileSchema>` | Inferred TypeScript type |
| `type Operation` | `z.infer<typeof OperationSchema>` | Inferred TypeScript type |
| `type OperationType` | `z.infer<...>` | Union of operation type strings |
| `defaultStore(overrides?)` | `function` | Returns a valid `StoreData` with sensible defaults |

**ID type schemas:** The string literal union types (`DocsDirId`, `AgentId`, `SkillId`, etc.) are defined as zod enums in this file. The existing `src/types.ts` union types are **replaced** by `z.infer<>` exports from the zod enums.

**`src/types.ts` evolution:** The file remains but re-exports types from `src/store/schema.ts` for backward compatibility. Any consumer importing `SetupType`, `ToolId`, `WizardSelections`, etc. from `src/types.ts` continues to work. The `AiSetupConfig` type is deprecated with a JSDoc `@deprecated` tag pointing to `StoreData`.

#### `src/store/index.ts`

**Exports:**

| Export | Kind | Description |
|--------|------|-------------|
| `createStore(targetDir)` | `async function` | Creates a lowdb instance with `JSONFile` adapter pointing to `<targetDir>/.ai-setup.json`. Performs auto-migration if legacy format detected on first `read()`. Returns typed `Low<StoreData>` instance. |
| `createTestStore(initial?)` | `function` | Creates a lowdb instance with `Memory` adapter. Accepts optional initial data (uses `defaultStore()` if omitted). For vitest only. |
| `readStore(targetDir)` | `async function` | Convenience: creates store, reads, returns data. Throws `AiSetupError` with `MANIFEST_NOT_FOUND` if file missing, `MANIFEST_CORRUPT` if validation fails. |
| `writeStore(store, data)` | `async function` | Validates data against `StoreSchema`, writes via lowdb. Automatically updates `meta.lastUpdatedAt`. |
| `appendOperation(store, operation)` | `async function` | Appends operation to log, enforces 50-entry cap, writes store. |

**Auto-migration behavior:**
1. `createStore()` reads the file via lowdb
2. If file doesn't exist → returns store with `defaultStore()` data (not written to disk yet)
3. If file exists → checks `isLegacyFormat(data)`
4. If legacy → runs `migrateV0toV1(data)` → validates with `StoreSchema` → writes migrated data back
5. If current → validates with `StoreSchema` → returns
6. If validation fails post-migration → throws `Errors.migrationFailed()`

#### `src/store/migrations.ts`

**Exports:**

| Export | Kind | Description |
|--------|------|-------------|
| `isLegacyFormat(data)` | `function` | Returns `true` if data lacks `meta.schemaVersion` AND has top-level `version` field (string) |
| `migrateV0toV1(legacy)` | `function` | Transforms `AiSetupConfig` (v0) shape to `StoreData` (v1) shape |
| `migrate(data)` | `function` | Entry point: detects version, runs all needed migrations in sequence |

**v0 → v1 migration mapping:**

| v0 field | v1 field | Transform |
|----------|----------|-----------|
| `version` | `meta.cliVersion` | Direct copy |
| (absent) | `meta.schemaVersion` | Set to `1` |
| `installedAt` | `meta.installedAt` | Direct copy |
| (absent) | `meta.lastUpdatedAt` | Set to current ISO timestamp |
| `setupType` | `config.setupType` | Direct copy |
| `tools` | `config.tools` | Direct copy |
| `projectName` | `config.projectName` | Direct copy |
| (absent) | `config.targetDir` | Set to `process.cwd()` (best guess) |
| `selections` | `selections` | Direct copy if present; empty arrays for all fields if absent |
| `files[]` | `files[]` | Copy `path`, `hash`, `source`; set `status: 'installed'`, `installedAt` from `meta.installedAt`, `lastCheckedAt` to current timestamp |
| (absent) | `sync` | `{ lastSyncAt: null, dirty: false }` |
| (absent) | `operations` | `[]` (empty) |

---

## 2. Error Handling Subsystem

### 2.1 File Structure

```
src/errors/
  types.ts      — ErrorCode enum, AiSetupError class, Errors factory
  boundary.ts   — handleError() function
  operation.ts  — OperationTracker class
```

### 2.2 ErrorCode Enum

String enum with the following members:

| Code | Category | When thrown |
|------|----------|------------|
| `FILE_NOT_FOUND` | Filesystem | File expected but absent |
| `FILE_PERMISSION` | Filesystem | Permission denied on read/write |
| `FILE_CORRUPT` | Filesystem | File exists but content is invalid |
| `DIR_NOT_FOUND` | Filesystem | Directory expected but absent |
| `MANIFEST_NOT_FOUND` | Store | `.ai-setup.json` missing when required |
| `MANIFEST_CORRUPT` | Store | `.ai-setup.json` fails zod validation |
| `MANIFEST_VERSION` | Store | Schema version newer than CLI supports |
| `MIGRATION_FAILED` | Store | v0→v1 migration threw |
| `CONFLICT_UNRESOLVED` | Conflict | File conflict without resolution strategy |
| `PARTIAL_WRITE` | Operation | Some files written, some failed |
| `HASH_MISMATCH` | Integrity | File hash doesn't match expected |
| `USER_CANCELLED` | User | User pressed Ctrl+C or cancelled prompt |
| `INVALID_INPUT` | Validation | Invalid CLI argument or prompt response |
| `MISSING_DEPENDENCY` | Environment | Required tool/binary not found |
| `UNKNOWN` | Fallback | Unclassified error |

### 2.3 AiSetupError Class

```
class AiSetupError extends Error {
  readonly code: ErrorCode
  readonly context: Record<string, unknown>
  override readonly cause?: Error

  get isUserError(): boolean    // true for USER_CANCELLED, INVALID_INPUT
  get exitCode(): number        // 0 for USER_CANCELLED, 1 for all others
}
```

**Behavioral contract:**
- `isUserError === true` → show only message (no stack trace, no "unexpected error" framing)
- `isUserError === false` → show message, and show stack trace if DEBUG mode enabled
- `exitCode === 0` → user intentionally cancelled (not an error)
- `exitCode === 1` → actual error

### 2.4 Errors Factory

Convenience constructors that pre-fill `code`, `message`, and `context`:

| Factory Method | ErrorCode | Message Pattern |
|---------------|-----------|-----------------|
| `Errors.fileNotFound(path)` | `FILE_NOT_FOUND` | `"File not found: {path}"` |
| `Errors.filePermission(path, op)` | `FILE_PERMISSION` | `"Permission denied: cannot {op} {path}"` |
| `Errors.fileCorrupt(path, cause?)` | `FILE_CORRUPT` | `"File is corrupt or invalid: {path}"` |
| `Errors.dirNotFound(path)` | `DIR_NOT_FOUND` | `"Directory not found: {path}"` |
| `Errors.manifestNotFound(dir)` | `MANIFEST_NOT_FOUND` | `"No .ai-setup.json found in {dir}. Run 'ai-setup init' first."` |
| `Errors.manifestCorrupt(dir, cause?)` | `MANIFEST_CORRUPT` | `"Invalid .ai-setup.json in {dir}"` |
| `Errors.manifestVersion(found, expected)` | `MANIFEST_VERSION` | `"Schema version {found} is newer than supported ({expected}). Update ai-setup."` |
| `Errors.migrationFailed(cause?)` | `MIGRATION_FAILED` | `"Failed to migrate .ai-setup.json to current format"` |
| `Errors.userCancelled()` | `USER_CANCELLED` | `"Operation cancelled"` |
| `Errors.invalidInput(field, reason)` | `INVALID_INPUT` | `"Invalid {field}: {reason}"` |
| `Errors.partialWrite(succeeded, failed)` | `PARTIAL_WRITE` | `"Partially completed: {succeeded.length} succeeded, {failed.length} failed"` |
| `Errors.hashMismatch(path, expected, actual)` | `HASH_MISMATCH` | `"Hash mismatch for {path}"` |
| `Errors.missingDependency(name)` | `MISSING_DEPENDENCY` | `"Required dependency not found: {name}"` |

### 2.5 handleError(err) Boundary

Located in `src/errors/boundary.ts`. Single function that:

1. Receives any thrown value (`unknown`)
2. If `AiSetupError`:
   - `USER_CANCELLED` → `p.cancel(message)`, exit with code 0
   - `isUserError` → `p.log.error(message)`, exit with code 1
   - Otherwise → `p.log.error(message)`, show `context` if DEBUG, show stack trace if DEBUG, exit with `exitCode`
3. If `@clack/prompts` cancel symbol (`p.isCancel(err)`) → treat as `USER_CANCELLED`
4. If plain `Error` → wrap in `AiSetupError` with `UNKNOWN` code, show message + stack if DEBUG
5. If non-Error → stringify, wrap, show

**DEBUG mode detection:**
```
isDebug = process.env.AI_SETUP_DEBUG === '1' || process.argv.includes('--verbose')
```

**Exit behavior:** `handleError` calls `process.exit()`. It is the ONLY place in the codebase that calls `process.exit()` (besides the entry point wrapper in `src/index.ts`).

### 2.6 OperationTracker Class

Located in `src/errors/operation.ts`.

```
class OperationTracker {
  constructor(type: OperationType)

  trackSuccess(filePath: string): void
  trackFailure(filePath: string, error: Error): void
  registerBackup(originalPath: string, backupPath: string): void

  get succeeded(): string[]
  get failed(): Array<{ path: string; error: Error }>
  get backups(): Map<string, string>
  get result(): 'success' | 'partial' | 'failed'

  toOperation(): Operation   // Generates Operation record for store log
}
```

**Behavioral contract:**
- If all tracked files succeed → `result === 'success'`
- If some succeed and some fail → `result === 'partial'`
- If all fail or no files tracked → `result === 'failed'`
- `toOperation()` produces a complete `Operation` object ready for `appendOperation()`
- Backups map: `originalPath → backupPath` for potential rollback

---

## 3. Migration Strategy

### 3.1 Detection

```
isLegacyFormat(data) = 
  typeof data === 'object' &&
  data !== null &&
  typeof data.version === 'string' &&
  (!data.meta || typeof data.meta.schemaVersion !== 'number')
```

This distinguishes:
- **Legacy v0:** Has `version: "0.1.0"` at top level, no `meta` field
- **Current v1:** Has `meta.schemaVersion: 1`
- **Future v2+:** Has `meta.schemaVersion: N` where `N > CURRENT_SCHEMA_VERSION` → throws `Errors.manifestVersion()`

### 3.2 Migration Execution

1. `createStore(targetDir)` reads raw JSON from disk
2. Calls `migrate(rawData)` which:
   a. If `isLegacyFormat(rawData)` → runs `migrateV0toV1(rawData)`
   b. If `meta.schemaVersion === CURRENT_SCHEMA_VERSION` → returns as-is
   c. If `meta.schemaVersion > CURRENT_SCHEMA_VERSION` → throws `Errors.manifestVersion()`
   d. If `meta.schemaVersion < CURRENT_SCHEMA_VERSION` → runs sequential migrations (v1→v2, v2→v3, etc.) — infrastructure for future migrations
3. Validates result against `StoreSchema.parse()`
4. If validation fails → throws `Errors.migrationFailed(zodError)`
5. Writes migrated data back to disk (one-time upgrade)

### 3.3 Migration Safety

- Original file is NOT backed up automatically (lowdb's steno handles atomic writes)
- If power loss during write → steno's temp-file-then-rename ensures either old or new data, never corrupt
- If migration throws → error propagates to `handleError` boundary with clear message

---

## 4. File-by-File Change Map

### New Files

| File | Purpose |
|------|---------|
| `src/store/schema.ts` | Zod schemas, types, `CURRENT_SCHEMA_VERSION`, `defaultStore()` |
| `src/store/index.ts` | `createStore()`, `createTestStore()`, `readStore()`, `writeStore()`, `appendOperation()` |
| `src/store/migrations.ts` | `isLegacyFormat()`, `migrateV0toV1()`, `migrate()` |
| `src/errors/types.ts` | `ErrorCode` enum, `AiSetupError` class, `Errors` factory |
| `src/errors/boundary.ts` | `handleError()` function |
| `src/errors/operation.ts` | `OperationTracker` class |
| `src/__tests__/store.test.ts` | Store creation, read/write, defaults |
| `src/__tests__/store-migrations.test.ts` | Legacy detection, v0→v1 migration, validation |
| `src/__tests__/errors.test.ts` | AiSetupError, Errors factory, handleError |

### Modified Files

| File | Change Description |
|------|-------------------|
| `package.json` | Add `lowdb` and `zod` to `dependencies` |
| `src/types.ts` | Re-export types from `src/store/schema.ts`. Deprecate `AiSetupConfig` with `@deprecated` JSDoc. Keep all existing exports for backward compatibility. |
| `src/index.ts` | Replace inline catch block with `handleError()` import. Remove `process.exit(1)`. |
| `src/cli.ts` | Add `--verbose` global option to Commander program |
| `src/utils/manifest.ts` | Refactor `readManifest()` to use `readStore()`. Keep function signature for backward compat. `extractSelections()` unchanged (operates on any object with `selections` and `files`). |
| `src/utils/files.ts` | Replace `throw new Error(...)` with `throw Errors.fileNotFound(...)`, `Errors.filePermission(...)`, etc. |
| `src/utils/conflicts.ts` | Replace `process.exit(0)` with `throw Errors.userCancelled()`. Remove `@clack/prompts` cancel-to-exit pattern. |
| `src/commands/doctor.ts` | Replace `JSON.parse + process.exit` with `readStore()` + `AiSetupError`. Update file status tracking in store. |
| `src/commands/update.ts` | Replace `JSON.parse + process.exit` with `readStore()` + `AiSetupError`. Use `OperationTracker`. Write operation to store log. |
| `src/commands/add.ts` | Replace `JSON.parse + process.exit` with `readStore()` + `AiSetupError`. |
| `src/commands/eject.ts` | Replace `catch (err: any)` with typed error handling. |
| `src/commands/create.ts` | Replace `throw new Error(...)` with `Errors.invalidInput()`. Replace `process.exit(0)` with `throw Errors.userCancelled()`. |
| `src/wizard/index.ts` | Replace `process.exit(1)` in catch block with re-throw (let boundary handle). Replace `process.exit(0)` cancel handling with `throw Errors.userCancelled()`. |

### Unchanged Files

| File | Why unchanged |
|------|---------------|
| `src/commands/status.ts` | Stub only ("coming soon") — will use store in future |
| `src/commands/init.ts` | Delegates to wizard; no direct config access |
| `src/adapters/*` | Adapter install logic unchanged; receives `fileRecords` array |
| `src/generators/*` | Generator logic unchanged; no config access |
| `src/scaffold/*` | Scaffold functions unchanged; receive params from wizard |
| `src/wizard/phase*.ts` | Phase logic unchanged; wizard index handles config |
| `src/prompts.ts` | Output formatting unchanged |
| `src/utils/diff.ts` | Diff utility unchanged |
| `src/utils/frontmatter.ts` | Frontmatter parsing unchanged |
| `src/utils/validation.ts` | Validation helpers unchanged |

---

## 5. Testing Strategy

### 5.1 Store Tests (`src/__tests__/store.test.ts`)

| Test Case | What it verifies |
|-----------|-----------------|
| `createTestStore()` returns valid default data | Memory adapter works, `defaultStore()` passes schema validation |
| `writeStore()` validates data before write | Invalid data throws `MANIFEST_CORRUPT` |
| `writeStore()` updates `meta.lastUpdatedAt` | Timestamp auto-update |
| `appendOperation()` adds to log | Operation appears in `store.data.operations` |
| `appendOperation()` caps at 50 | 51st entry evicts oldest |
| `readStore()` throws on missing manifest | `MANIFEST_NOT_FOUND` error code |
| `readStore()` throws on corrupt JSON | `MANIFEST_CORRUPT` error code |

### 5.2 Migration Tests (`src/__tests__/store-migrations.test.ts`)

| Test Case | What it verifies |
|-----------|-----------------|
| `isLegacyFormat()` returns true for v0 shape | Detection logic |
| `isLegacyFormat()` returns false for v1 shape | No false positives |
| `migrateV0toV1()` maps all fields correctly | Field-by-field mapping |
| `migrateV0toV1()` handles absent `selections` | Empty arrays for all selection fields |
| `migrateV0toV1()` result passes `StoreSchema.parse()` | Migrated data is valid |
| `migrate()` is idempotent on v1 data | Running migrate on already-migrated data is a no-op |
| `migrate()` throws on future schema version | `MANIFEST_VERSION` error |
| Round-trip: real v0 JSON → migrate → validate | End-to-end migration |

### 5.3 Error Tests (`src/__tests__/errors.test.ts`)

| Test Case | What it verifies |
|-----------|-----------------|
| `AiSetupError` has correct code, message, context | Constructor behavior |
| `AiSetupError.isUserError` for cancel and input codes | Getter logic |
| `AiSetupError.exitCode` is 0 for cancel, 1 for errors | Getter logic |
| Each `Errors.*` factory produces correct code | Factory methods |
| `handleError(AiSetupError)` with user cancel | Calls `p.cancel`, exit 0 |
| `handleError(AiSetupError)` with file error | Calls `p.log.error`, exit 1 |
| `handleError(plain Error)` wraps as UNKNOWN | Fallback behavior |
| `handleError` shows stack in DEBUG mode | `AI_SETUP_DEBUG=1` behavior |

### 5.4 Existing Test Updates

| Test File | Change |
|-----------|--------|
| `manifest.test.ts` | Update `buildManifest()` helper if `AiSetupConfig` shape changes. Add migration-aware tests. |
| `update-doctor.test.ts` | Update to use `createTestStore()` instead of raw JSON fixtures |
| `cli.e2e.test.ts` | Verify `--verbose` flag is accepted |

### 5.5 Test Principles

- All store tests use `createTestStore()` (Memory adapter) — zero filesystem I/O
- Error boundary tests mock `process.exit` and `@clack/prompts` output functions
- Migration tests use inline JSON fixtures representing known v0 shapes
- No test depends on another test's state (each test creates its own store instance)

---

## 6. Acceptance Criteria

### AC-1: Store reads/writes validated data
- [ ] `createStore(targetDir)` returns a `Low<StoreData>` instance
- [ ] `StoreSchema.parse()` validates all data before write
- [ ] Invalid data throws `AiSetupError` with `MANIFEST_CORRUPT` code
- [ ] `meta.lastUpdatedAt` is automatically updated on every write

### AC-2: Legacy migration works silently
- [ ] Existing `.ai-setup.json` (v0 format) auto-migrates to v1 on first read
- [ ] No interactive prompts during migration
- [ ] Migrated file passes `StoreSchema.parse()`
- [ ] All original data is preserved (no data loss)
- [ ] Running `ai-setup doctor` after upgrade shows no integrity issues for files that were healthy before

### AC-3: All errors are typed
- [ ] No `throw new Error(string)` in any modified file (only `throw new AiSetupError` or `throw Errors.*()`)
- [ ] No `process.exit()` calls outside `src/index.ts` and `src/errors/boundary.ts`
- [ ] Every error has an `ErrorCode`
- [ ] `handleError()` is the single error formatting/exit point

### AC-4: User experience unchanged
- [ ] `ai-setup init` works identically (interactive and non-interactive)
- [ ] `ai-setup add <tool>` works identically
- [ ] `ai-setup update` works identically (including conflict resolution)
- [ ] `ai-setup doctor` works identically (plus writes file status to store)
- [ ] `ai-setup create` works identically
- [ ] `ai-setup eject` works identically
- [ ] User cancel (Ctrl+C) still exits cleanly with code 0

### AC-5: Quality gates pass
- [ ] `npm run typecheck` passes with zero errors
- [ ] `npm run test` passes with all existing + new tests green
- [ ] `npm run build` produces valid dist output
- [ ] No TypeScript `any` in new code

### AC-6: Operation tracking
- [ ] `update` command writes operation to store log
- [ ] Operation log never exceeds 50 entries
- [ ] Each operation records `type`, `timestamp`, `filesAffected`, `result`

### AC-7: DEBUG mode
- [ ] `AI_SETUP_DEBUG=1` shows stack traces on error
- [ ] `--verbose` flag shows stack traces on error
- [ ] Normal mode shows only user-friendly messages

---

## 7. Implementation Priority

| Priority | Deliverable | Depends On | Rationale |
|----------|------------|------------|-----------|
| **P0** | `src/errors/types.ts` + `src/errors/boundary.ts` | Nothing | Foundation — everything else throws typed errors |
| **P0** | Refactor `src/index.ts` to use `handleError()` | P0 errors | Wire the boundary into the entry point |
| **P1** | `src/store/schema.ts` + `src/store/index.ts` + `src/store/migrations.ts` | P0 errors (for error types) | Store layer depends on error types for `Errors.manifest*()` |
| **P1** | `src/types.ts` re-exports from store schemas | P1 store schema | Backward compat bridge |
| **P2** | Refactor `src/utils/files.ts` to throw `AiSetupError` | P0 errors | Low-level file ops get typed errors |
| **P2** | Refactor `src/utils/conflicts.ts` to throw instead of `process.exit` | P0 errors | Remove exit from utility code |
| **P2** | Refactor `src/commands/doctor.ts` to use store | P1 store | Doctor uses `readStore()` and updates file statuses |
| **P2** | Refactor `src/commands/update.ts` to use store | P1 store | Update uses `readStore()` / `writeStore()` |
| **P2** | Refactor `src/commands/add.ts` to use store | P1 store | Add uses `readStore()` / `writeStore()` |
| **P2** | Refactor `src/commands/eject.ts`, `create.ts` | P0 errors | Typed errors, remove `process.exit` |
| **P2** | Refactor `src/wizard/index.ts` | P0 errors + P1 store | Write store instead of raw JSON; throw instead of `process.exit` |
| **P2** | Refactor `src/utils/manifest.ts` | P1 store | `readManifest()` delegates to `readStore()` |
| **P3** | `src/errors/operation.ts` (OperationTracker) | P0 errors + P1 store schema | Needs `Operation` type from schema |
| **P3** | Wire OperationTracker into `update` and `init` commands | P3 tracker + P2 command refactors | Final integration |
| **P3** | Add `--verbose` global flag to Commander | P0 boundary | Exposes DEBUG mode via CLI |

---

## 8. Open Questions

| # | Question | Impact | Proposed Resolution |
|---|---------|--------|-------------------|
| Q1 | Should `config.targetDir` be stored as absolute path or relative? | Portability — absolute paths break if project moves | **Propose:** Store as absolute. It's for this machine's reference. Moving a project is a re-init anyway. |
| Q2 | Should migration back up the original `.ai-setup.json` before migrating? | Safety vs simplicity | **Propose:** No backup. lowdb/steno atomic writes prevent corruption. Users can `git checkout` if needed. |
| Q3 | Should `readManifest()` in `src/utils/manifest.ts` be deprecated or kept as a thin wrapper? | API surface size | **Propose:** Keep as thin wrapper over `readStore()` for backward compat. Deprecate with JSDoc. |
| Q4 | Should the `force` field remain in `SetupConfig` / `WizardConfig` or move to a separate concern? | `force` is a runtime option, not persisted config | **Propose:** Keep in `WizardConfig` (not stored). It's a command-time flag, not part of `StoreData`. |
| Q5 | Should `TrackedFile.status` be recomputed on every `readStore()` or only on explicit `doctor`/`update`? | Performance vs freshness | **Propose:** Only on `doctor`/`update`. Recomputing on every read means hashing every file on every command, which is too slow for large setups. |
| Q6 | Are there any external tools or scripts that parse `.ai-setup.json` directly? | Migration could break them | **Propose:** Treat as internal. Package is v0.1.0 (pre-1.0 semver allows breaking changes). Document the schema change in CHANGELOG. |

---

## Requirement Traceability

| Requirement | Source | Spec Section |
|-------------|--------|-------------|
| lowdb v7 with JSONFile adapter | User decision (context) | 1.1, 1.4 |
| zod for runtime validation | User decision (context) | 1.2, 1.4 |
| Schema version tracking | User spec (meta.schemaVersion) | 1.2 (meta) |
| Operation log capped at 50 | User decision (context) | 1.2 (operations), D3 |
| Legacy auto-migration | User requirement (backward compat) | 3.x, AC-2 |
| ErrorCode enum | User spec (error handling) | 2.2 |
| AiSetupError class | User spec (error handling) | 2.3 |
| Errors factory | User spec (error handling) | 2.4 |
| handleError boundary | User spec (error handling) | 2.5 |
| OperationTracker | User spec (error handling) | 2.6 |
| Memory adapter for tests | User requirement (testing) | 1.4, 5.5 |
| Same `.ai-setup.json` filename | User constraint | C3, P3 |
| No breaking command behavior | User constraint | C6, P8, AC-4 |

---

## Ready-for-Planning Statement

This spec is **ready for planning** with the following conditions:

1. **Open questions Q1-Q6** have proposed resolutions. If the proposed resolutions are accepted, no further clarification is needed.
2. **Assumption A6** (no external tools read `.ai-setup.json`) is marked pending. Low risk given pre-1.0 semver, but worth confirming.
3. All new dependencies (`lowdb`, `zod`) are well-established, ESM-compatible, and have been verified against the project's Node >= 18 constraint.
4. The implementation priority (P0 → P3) provides a safe incremental path where each layer builds on the previous and quality gates can be run at each step.
