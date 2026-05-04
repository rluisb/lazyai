# Implementation Plan: Store & Error Handling (001)

> Ordered execution plan for spec `001-store-and-errors` in `@rluisb/lazyai`.  
> **Stack:** TypeScript ESM, Commander, @clack/prompts, tsup, vitest  
> **Quality gates:** `npm run typecheck`, `npm run test`, `npm run build`

---

## Overview

| Tier | Theme | Tasks | AC Coverage |
|------|-------|-------|-------------|
| P0 | Error foundation | T01–T03 | AC-3, AC-7 |
| P1 | Store layer | T04–T06 | AC-1, AC-2 |
| P2 | Command refactors | T07–T12 | AC-3, AC-4 |
| P3 | Operation tracking + verbose | T13–T14 | AC-6, AC-7 |

**Parallelization:** Tasks marked `[P]` are safe to work simultaneously because they touch non-overlapping files. All others must execute in dependency order.

---

## Dependency Graph

```
T01 (errors/types.ts)
  └─► T02 (errors/boundary.ts)
        └─► T03 (src/index.ts wire boundary)
              └─► [P] T04 (store/schema.ts)
              └─► [P] T07 (files.ts + conflicts.ts)    ← depends on T01
T04 (store/schema.ts)
  └─► T05 (store/migrations.ts)
        └─► T06 (store/index.ts)
              └─► [P] T08 (doctor.ts)
              └─► [P] T09 (update.ts)
              └─► [P] T10 (add.ts)
              └─► [P] T11 (eject.ts + create.ts)
              └─► T12 (wizard/index.ts + manifest.ts)
                    └─► T13 (OperationTracker)
                          └─► T14 (wire tracker + --verbose)
```

---

## Milestones

| Milestone | After Task | Gate |
|-----------|------------|------|
| M1: Error Foundation | T03 | `npm run typecheck && npm run test` |
| M2: Store Layer | T06 | `npm run typecheck && npm run test` |
| M3: Commands Refactored | T12 | `npm run typecheck && npm run test && npm run build` |
| M4: Full Spec | T14 | All AC verified, full quality gate |

---

## Risks

| Risk | Tasks Affected | Mitigation |
|------|---------------|------------|
| R3: Zod rejects valid legacy data | T05, T06 | Run `migrateV0toV1` before `StoreSchema.parse()`. Cover all known legacy shapes in T05 tests. |
| R4: Removing `process.exit()` breaks cancel | T07, T08–T12 | `USER_CANCELLED` error code with `exitCode: 0`. Verify by running `ai-setup init` interactively and pressing Ctrl+C. |
| update-doctor.test.ts reads `.ai-setup.json` as `AiSetupConfig` | T09, T12 | After T12 the test must read via `readStore()` or remain compatible by keeping `readManifest()` as wrapper. |

---

## Tasks

---

### T01 — Create `src/errors/types.ts`
**Priority:** P0 | **Complexity:** M | **Depends on:** nothing

**Spec refs:** §2.2, §2.3, §2.4

**What to read before implementing:**
- `src/types.ts` (existing types that errors reference)
- `spec.md §2.2–2.4`

**Files to create:**
- `src/errors/types.ts`

**Implement:**
1. `ErrorCode` string enum — all 15 members from spec §2.2.
2. `AiSetupError` class extending `Error`:
   - Constructor: `(code: ErrorCode, message: string, context?: Record<string, unknown>, cause?: Error)`
   - `readonly code: ErrorCode`
   - `readonly context: Record<string, unknown>`
   - `override readonly cause?: Error`
   - Getter `isUserError`: `true` when code is `USER_CANCELLED` or `INVALID_INPUT`
   - Getter `exitCode`: `0` when code is `USER_CANCELLED`, `1` otherwise
3. `Errors` factory object with all 13 factory methods from spec §2.4.
   - Each method produces the correct `code`, message pattern, and `context`.

**Files to create (tests):**
- `src/__tests__/errors.test.ts` — test all factory methods, getters, and class properties (spec §5.3)

**Acceptance check:**
- `npm run typecheck` passes
- All `errors.test.ts` cases green: `AiSetupError` shape, `isUserError`, `exitCode`, each `Errors.*` factory

---

### T02 — Create `src/errors/boundary.ts`
**Priority:** P0 | **Complexity:** M | **Depends on:** T01

**Spec refs:** §2.5, AC-3, AC-7

**What to read before implementing:**
- `src/errors/types.ts` (T01 output)
- `spec.md §2.5`
- `src/index.ts` (current error boundary pattern to replace)

**Files to create:**
- `src/errors/boundary.ts`

**Implement:**
1. `isDebug()` helper: `process.env.AI_SETUP_DEBUG === '1' || process.argv.includes('--verbose')`
2. `handleError(err: unknown): never` function:
   - If `p.isCancel(err)` → `p.cancel('Operation cancelled.')` + `process.exit(0)`
   - If `err instanceof AiSetupError` and `code === USER_CANCELLED` → `p.cancel(err.message)` + `process.exit(0)`
   - If `err instanceof AiSetupError` and `isUserError` → `p.log.error(err.message)` + `process.exit(1)`
   - If `err instanceof AiSetupError` → `p.log.error(err.message)`, show `err.context` if `isDebug()`, show stack if `isDebug()`, `process.exit(err.exitCode)`
   - If plain `Error` → wrap as `new AiSetupError(ErrorCode.UNKNOWN, err.message, {}, err)`, show stack if `isDebug()`, `process.exit(1)`
   - Else → stringify, wrap, `process.exit(1)`

**Test additions to `src/__tests__/errors.test.ts`:**
- Mock `process.exit` and `p.cancel`, `p.log.error`
- `handleError(AiSetupError USER_CANCELLED)` → calls `p.cancel`, exit 0
- `handleError(AiSetupError FILE_NOT_FOUND)` → calls `p.log.error`, exit 1
- `handleError(plain Error)` → wraps as UNKNOWN, exit 1
- `handleError(cancel symbol)` → treated as USER_CANCELLED, exit 0
- `AI_SETUP_DEBUG=1` → stack trace shown (spy on console or p.log)

**Acceptance check:**
- `npm run typecheck` passes
- All `handleError` tests green

---

### T03 — Wire `handleError` into `src/index.ts` + `src/errors/index.ts` barrel
**Priority:** P0 | **Complexity:** S | **Depends on:** T01, T02

**Spec refs:** §2.5 "Exit behavior", §4 modified files, AC-3

**What to read before implementing:**
- `src/index.ts` (10 lines — current catch block)
- `src/errors/boundary.ts` (T02 output)

**Files to modify:**
- `src/index.ts` — replace inline catch with `handleError(err)`

**Files to create:**
- `src/errors/index.ts` — barrel re-exporting `AiSetupError`, `ErrorCode`, `Errors` from `./types.js` and `handleError` from `./boundary.js`

**Implement `src/index.ts`:**
```typescript
import { run } from './cli.js'
import { handleError } from './errors/boundary.js'

run().catch(handleError)
```

**Acceptance check:**
- `npm run typecheck` passes
- `npm run test` (existing tests still green — no behavior change yet)
- `npm run build` compiles

> **MILESTONE M1:** Error foundation complete. Run `npm run typecheck && npm run test`.

---

### T04 — Create `src/store/schema.ts`
**Priority:** P1 | **Complexity:** M | **Depends on:** T01

**Spec refs:** §1.2 (all schema tables), §1.3, §1.4 schema.ts exports

**What to read before implementing:**
- `src/types.ts` (all existing type definitions to migrate)
- `spec.md §1.2` (field-by-field schema tables)
- `spec.md §1.4` (exports list for schema.ts)

**Files to create:**
- `src/store/schema.ts`

**Implement:**
1. `CURRENT_SCHEMA_VERSION = 1`
2. All ID type enums as zod enums: `SetupTypeSchema`, `ToolIdSchema`, `DocsDirIdSchema`, `AgentIdSchema`, `SkillIdSchema`, `PromptIdSchema`, `TemplateIdSchema`, `RuleIdSchema`, `InfraIdSchema`
3. `MetaSchema`, `ConfigSchema`, `WizardSelectionsSchema`, `TrackedFileSchema`, `SyncSchema`, `OperationSchema`, `StoreSchema`
4. Exported inferred types: `type StoreData`, `type TrackedFile`, `type Operation`, `type OperationType`
5. `defaultStore(overrides?: Partial<StoreData>): StoreData` — returns valid store with:
   - `meta`: `schemaVersion: 1`, `cliVersion: '0.1.0'`, `installedAt` + `lastUpdatedAt` as current ISO
   - `config`: empty strings, empty arrays
   - `selections`: all empty arrays
   - `files`: `[]`
   - `sync`: `{ lastSyncAt: null, dirty: false }`
   - `operations`: `[]`

**Critical:** `TrackedFileSchema` must include the new `status`, `installedAt`, `lastCheckedAt` fields. `status` is `z.enum(['installed', 'modified', 'missing', 'conflict'])`.

**No tests required** for this file (schemas are validated indirectly through store and migration tests).

**Acceptance check:**
- `npm run typecheck` passes (zod types resolve correctly)

---

### T05 — Create `src/store/migrations.ts`
**Priority:** P1 | **Complexity:** M | **Depends on:** T04

**Spec refs:** §1.4 migrations.ts exports, §3 (full migration strategy), AC-2

**What to read before implementing:**
- `src/store/schema.ts` (T04 output — `StoreSchema`, `CURRENT_SCHEMA_VERSION`)
- `src/types.ts` (legacy `AiSetupConfig` shape)
- `spec.md §3.1, §3.2, §1.4 migrations.ts, §1.2 v0→v1 mapping table`

**Files to create:**
- `src/store/migrations.ts`

**Implement:**
1. `isLegacyFormat(data: unknown): boolean`:
   - `typeof data === 'object' && data !== null`
   - `typeof (data as any).version === 'string'`
   - `!data.meta || typeof data.meta.schemaVersion !== 'number'`
2. `migrateV0toV1(legacy: unknown): StoreData` — apply all mappings from spec §3.2 v0→v1 table:
   - `version` → `meta.cliVersion`
   - set `meta.schemaVersion = 1`
   - copy `installedAt`, set `lastUpdatedAt` = `new Date().toISOString()`
   - `setupType`, `tools`, `projectName` → `config.*`
   - `config.targetDir = process.cwd()` (best guess)
   - `selections` → copy or default to empty arrays per field
   - `files[]` → copy `path/hash/source`, add `status: 'installed'`, `installedAt` from `meta.installedAt`, `lastCheckedAt = now`
   - `sync = { lastSyncAt: null, dirty: false }`
   - `operations = []`
3. `migrate(data: unknown): StoreData`:
   - if `isLegacyFormat(data)` → `migrateV0toV1(data)`
   - if `schemaVersion === CURRENT_SCHEMA_VERSION` → validate with `StoreSchema.parse()` and return
   - if `schemaVersion > CURRENT_SCHEMA_VERSION` → `throw Errors.manifestVersion(schemaVersion, CURRENT_SCHEMA_VERSION)`
   - if `schemaVersion < CURRENT_SCHEMA_VERSION` → run sequential chain (infrastructure for future)
   - Validate with `StoreSchema.parse()`. If throws → `throw Errors.migrationFailed(zodError)`

**Files to create (tests):**
- `src/__tests__/store-migrations.test.ts` — all 8 cases from spec §5.2

**Acceptance check:**
- `npm run typecheck` passes
- All `store-migrations.test.ts` cases green including round-trip test using a real v0 JSON fixture

---

### T06 — Create `src/store/index.ts` + `src/store/index-barrel.ts`
**Priority:** P1 | **Complexity:** M | **Depends on:** T04, T05

**Spec refs:** §1.4 store/index.ts exports, §1.4 auto-migration behavior, AC-1, AC-2

**What to read before implementing:**
- `src/store/schema.ts` (T04 output)
- `src/store/migrations.ts` (T05 output)
- `src/errors/types.ts` (T01 output — `Errors.manifest*()`)
- `spec.md §1.4 store/index.ts`

**Files to create:**
- `src/store/index.ts`

**Implement:**
1. Import `Low`, `JSONFile`, `Memory` from `lowdb` (v7 ESM imports)
2. `createStore(targetDir: string): Promise<Low<StoreData>>`:
   - Construct path `join(targetDir, '.ai-setup.json')`
   - Create `new Low(new JSONFile(path), defaultStore())`
   - Call `await store.read()`
   - If `store.data` is the default (file didn't exist) → return (not yet written)
   - Else → run `migrate(store.data)`, assign result to `store.data`, write back if migrated
   - Return `store`
3. `createTestStore(initial?: Partial<StoreData>): Low<StoreData>`:
   - `new Low(new Memory(), { ...defaultStore(), ...initial })`
   - No async needed (Memory adapter is sync)
4. `readStore(targetDir: string): Promise<StoreData>`:
   - `createStore(targetDir)`, check `store.data`
   - If file not found: throw `Errors.manifestNotFound(targetDir)`
   - If zod fails: throw `Errors.manifestCorrupt(targetDir, zodError)`
   - Return `store.data`
5. `writeStore(store: Low<StoreData>, data: StoreData): Promise<void>`:
   - Validate via `StoreSchema.parse(data)` — throw `Errors.manifestCorrupt()` on failure
   - Mutate `data.meta.lastUpdatedAt = new Date().toISOString()`
   - Assign `store.data = data`
   - `await store.write()`
6. `appendOperation(store: Low<StoreData>, op: Omit<Operation, 'id' | 'timestamp'>): Promise<void>`:
   - Generate id: `` `op_${Date.now()}_${Math.random().toString(36).slice(2, 7)}` ``
   - Set `timestamp = new Date().toISOString()`
   - If `store.data.operations.length >= 50` → remove oldest entry first
   - Push new operation
   - `await store.write()`

**Files to create (tests):**
- `src/__tests__/store.test.ts` — all 7 cases from spec §5.1 using `createTestStore()` (zero filesystem I/O)

**Note on lowdb v7 import:**
```typescript
import { Low } from 'lowdb'
import { JSONFile } from 'lowdb/node'
import { Memory } from 'lowdb'
```

**Acceptance check:**
- `npm run typecheck` passes
- All `store.test.ts` cases green
- All `store-migrations.test.ts` still green

> **MILESTONE M2:** Store layer complete. Run `npm run typecheck && npm run test`.

---

### T07 — `[P]` Refactor `src/utils/files.ts` and `src/utils/conflicts.ts`
**Priority:** P2 | **Complexity:** S | **Depends on:** T01 (errors/types.ts)

**Spec refs:** §4 modified files (files.ts, conflicts.ts), AC-3

**What to read before implementing:**
- `src/utils/files.ts` (current — throws plain `Error`)
- `src/utils/conflicts.ts` (current — `process.exit(0)` on cancel)
- `src/errors/types.ts` (T01 output — `Errors` factory)

**Files to modify:**
- `src/utils/files.ts`
- `src/utils/conflicts.ts`

**Implement `files.ts` changes:**
Replace each `throw new Error(...)` with typed equivalent:
- `findPackageRoot`: `throw Errors.dirNotFound(startDir)` (at bottom of loop)
- `ensureDir`: `throw Errors.filePermission(dirPath, 'create directory')`
- `readFile`: `throw Errors.filePermission(filePath, 'read')` (distinguish ENOENT → `fileNotFound` vs other → `filePermission`)
- `writeFile`: `throw Errors.filePermission(filePath, 'write')`
- `copyFile`: `throw Errors.filePermission(\`${src} → ${dest}\`, 'copy')`
- `copyDir` missing src: `throw Errors.dirNotFound(src)`
- `fileHash`: `throw Errors.fileNotFound(filePath)` (ENOENT) or `Errors.filePermission(filePath, 'hash')`

**Implement `conflicts.ts` changes:**
- Both `p.isCancel(replaceCustomized)` handlers: replace `p.cancel(...); process.exit(0)` with `throw Errors.userCancelled()`
- Both `p.isCancel(replaceExisting)` handlers: same pattern

**Acceptance check:**
- `npm run typecheck` passes
- Existing `conflict-strategy.test.ts` still green (behavior unchanged)
- `npm run test` passes

---

### T08 — `[P]` Refactor `src/commands/doctor.ts`
**Priority:** P2 | **Complexity:** S | **Depends on:** T06, T07

**Spec refs:** §4 modified files (doctor.ts), AC-3, AC-4

**What to read before implementing:**
- `src/commands/doctor.ts` (current — 70 lines, `JSON.parse`, `process.exit(1)`)
- `src/store/index.ts` (T06 — `readStore()`)
- `src/errors/types.ts` (T01 — `Errors`)

**Files to modify:**
- `src/commands/doctor.ts`

**Changes:**
1. Remove `readFileSync`, `JSON.parse` imports
2. Replace `fileExists(configPath)` guard + `process.exit(1)` with:
   ```typescript
   const store = await createStore(targetDir)
   const data = store.data
   if (!data.meta.installedAt) throw Errors.manifestNotFound(targetDir)
   ```
   Actually: use `readStore(targetDir)` directly — it throws `MANIFEST_NOT_FOUND` if absent
3. Access `data.config.files` instead of `config.files` (schema change: `files` is at top level `data.files`)
4. After computing `missing`/`modified`/`healthy`, update each file's `status` in the store:
   - files in `missing` → `status: 'missing'`, update `lastCheckedAt`
   - files in `modified` → `status: 'modified'`, update `lastCheckedAt`
   - remaining → `status: 'installed'`, update `lastCheckedAt`
5. Write updated statuses via `writeStore(store, data)` (or update `store.data` and `store.write()`)
6. Remove final `process.exit(1)` — let the command return normally (exit code managed by boundary); or keep `process.exit(1)` only if spec says so. Per spec P5, `handleError` is the only exit point — so re-throw an error instead:
   - `throw Errors.fileNotFound('setup integrity')` ... actually re-read spec: AC-4 says behavior must be identical. Doctor currently exits 1 with issues found. The boundary will do that for us if we throw. Use a named error or just log and return (currently it logs + exits 1). Safest: throw an `AiSetupError(HASH_MISMATCH)` or define a helper to preserve the list output + exit 1. Keep logs the same; after the display, throw `new AiSetupError(ErrorCode.FILE_CORRUPT, 'Setup integrity issues found')` so boundary exits 1.

**Acceptance check:**
- `npm run typecheck` passes
- `update-doctor.test.ts` green (doctor integration test)
- `npm run test` passes

---

### T09 — `[P]` Refactor `src/commands/update.ts`
**Priority:** P2 | **Complexity:** M | **Depends on:** T06, T07

**Spec refs:** §4 modified files (update.ts), AC-3, AC-4

**What to read before implementing:**
- `src/commands/update.ts` (current — 305 lines)
- `src/store/index.ts` (T06)
- `src/types.ts` (existing `FileRecord`)

**Files to modify:**
- `src/commands/update.ts`

**Changes:**
1. Remove `readFileSync`, `writeFileSync` imports
2. Replace `JSON.parse(readFileSync(...)) as AiSetupConfig` with `const store = await createStore(targetDir); const data = store.data`
3. Replace `process.exit(1)` on missing manifest with: rely on `readStore()` throwing (or use `createStore` and check)
4. `buildExpectedFiles`: signature changes from `(config: AiSetupConfig, ...)` to `(config: StoreData, ...)` — access `config.config.*` and `config.selections.*` instead of top-level fields. Note: `config.config.tools` not `config.tools`, `config.config.projectName` not `config.projectName`.
5. Replace `writeFileSync(configPath, JSON.stringify(config, null, 2))` with `await writeStore(store, data)`
6. `updatedRecords` now uses the new `TrackedFile` shape — include `status: 'installed'`, `installedAt` (preserved from existing record), `lastCheckedAt: new Date().toISOString()`
7. `FileRecord` → use `TrackedFile` type from store schema (keep imports compatible)

**Important:** `buildExpectedFiles` is a large function. The structural access pattern `config.config.tools` vs `config.tools` is the main change. Take care to update all references.

**Acceptance check:**
- `npm run typecheck` passes
- `update-doctor.test.ts` still passes (integration test runs full `update --force` scenario)
- `npm run test` passes

---

### T10 — `[P]` Refactor `src/commands/add.ts`
**Priority:** P2 | **Complexity:** S | **Depends on:** T06

**Spec refs:** §4 modified files (add.ts), AC-3, AC-4

**What to read before implementing:**
- `src/commands/add.ts` (current — 71 lines)
- `src/store/index.ts` (T06)

**Files to modify:**
- `src/commands/add.ts`

**Changes:**
1. Remove `readFileSync`, `writeFileSync` imports
2. Replace `JSON.parse(readFileSync(...)) as AiSetupConfig` with `const store = await createStore(targetDir)`
3. Replace `process.exit(1)` (3 occurrences: unknown tool, missing manifest, missing adapter) with throw `Errors.invalidInput(...)` or `Errors.manifestNotFound(...)` or `Errors.missingDependency(...)`
4. Replace `writeFileSync(configPath, JSON.stringify(config, null, 2))` with `await writeStore(store, store.data)`
5. `config.tools` → `store.data.config.tools`
6. `config.files` → `store.data.files`
7. New `FileRecord`s pushed into `store.data.files` must include `status: 'installed'`, `installedAt: new Date().toISOString()`, `lastCheckedAt: new Date().toISOString()`

**Acceptance check:**
- `npm run typecheck` passes
- `npm run test` passes (no direct test for `add` command, but typecheck validates)

---

### T11 — `[P]` Refactor `src/commands/eject.ts` and `src/commands/create.ts`
**Priority:** P2 | **Complexity:** S | **Depends on:** T01 (for errors), T06 (for eject store read)

**Spec refs:** §4 modified files (eject.ts, create.ts), AC-3, AC-4

**What to read before implementing:**
- `src/commands/eject.ts` (current — 51 lines, `catch(err: any)`)
- `src/commands/create.ts` (current — 368 lines, multiple `process.exit(0)`)
- `src/errors/types.ts` (T01 output)

**Files to modify:**
- `src/commands/eject.ts`
- `src/commands/create.ts`

**`eject.ts` changes:**
1. Replace `catch(err: any)` with `catch(err: unknown)`:
   ```typescript
   } catch(err: unknown) {
     throw err instanceof AiSetupError ? err : Errors.filePermission('.ai-setup.json', 'delete')
   }
   ```
2. Replace `readManifest(targetDir)` with `createStore(targetDir)` to get `StoreData | null`
3. The `!shouldEject || typeof shouldEject === 'symbol'` cancel check: change to `if (p.isCancel(shouldEject)) throw Errors.userCancelled()`; keep `!shouldEject` for normal "No" answer (just log and return, no error)

**`create.ts` changes:**
All 7 occurrences of:
```typescript
p.cancel('Create cancelled.')
process.exit(0)
```
Replace with:
```typescript
throw Errors.userCancelled()
```
Replace `throw new Error(`Invalid artifact type: ${value}`)` with `throw Errors.invalidInput('type', \`Invalid artifact type: ${value}\`)`
Replace `throw new Error('Prompt value cannot be empty')` with `throw Errors.invalidInput('prompt', 'value cannot be empty')`
Replace `throw new Error(\`A name is required...\`)` with `throw Errors.invalidInput('name', 'required in non-interactive mode')`
Replace `throw new Error('Type is required...')` with `throw Errors.invalidInput('type', 'required in non-interactive mode')`
Replace `throw new Error(\`No generator registered for type: ${type}\`)` with `throw Errors.invalidInput('type', \`no generator registered for type: ${type}\`)`
Replace `throw new Error(\`File already exists: ${file.path}...\`)` with `throw Errors.filePermission(file.path, 'overwrite (file exists, use --force)')`

**Acceptance check:**
- `npm run typecheck` passes — no `err: any`, no `throw new Error()` in modified files
- `npm run test` passes

---

### T12 — Refactor `src/wizard/index.ts` + `src/utils/manifest.ts` + update `src/types.ts`
**Priority:** P2 | **Complexity:** M | **Depends on:** T06, T07

**Spec refs:** §4 modified files (wizard/index.ts, manifest.ts, types.ts), AC-2, AC-3, AC-4

**What to read before implementing:**
- `src/wizard/index.ts` (full 230 lines — writes the manifest, handles cancel/error)
- `src/utils/manifest.ts` (readManifest + extractSelections)
- `src/types.ts` (all current exports — must remain backward compatible)
- `src/store/index.ts` (T06 — writeStore, readStore, createStore)
- `src/store/schema.ts` (T04 — StoreData, defaultStore)

**Files to modify:**
- `src/wizard/index.ts`
- `src/utils/manifest.ts`
- `src/types.ts`

**`wizard/index.ts` changes:**
1. Remove old `catch` block logic:
   ```typescript
   // BEFORE:
   } catch (error) {
     if (p.isCancel(error)) { p.cancel('Setup cancelled.'); process.exit(0) }
     const message = error instanceof Error ? error.message : String(error)
     p.cancel(`Setup failed: ${message}`)
     process.exit(1)
   }
   // AFTER:
   } catch (error) {
     if (p.isCancel(error)) throw Errors.userCancelled()
     throw error  // boundary handles it
   }
   ```
2. Replace manifest write block (lines 208–218):
   ```typescript
   // BEFORE: writes raw AiSetupConfig JSON
   // AFTER: write via writeStore()
   const store = await createStore(opts.targetDir)
   const storeData: StoreData = {
     meta: {
       schemaVersion: CURRENT_SCHEMA_VERSION,
       cliVersion: packageVersion,   // read from package.json or pass in
       installedAt: store.data.meta.installedAt || new Date().toISOString(),
       lastUpdatedAt: new Date().toISOString(),
     },
     config: {
       setupType,
       tools,
       projectName,
       targetDir: opts.targetDir,
     },
     selections,
     files: fileRecords.map(r => ({
       ...r,
       status: 'installed' as const,
       installedAt: new Date().toISOString(),
       lastCheckedAt: new Date().toISOString(),
     })),
     sync: { lastSyncAt: null, dirty: false },
     operations: store.data.operations,  // preserve existing if re-init
   }
   await writeStore(store, storeData)
   ```
3. Replace `readManifest(opts.targetDir)` at top with `createStore(opts.targetDir)` and read `store.data`
4. The `prior` object now reads from `store.data.config.*` and `store.data.selections`

**`manifest.ts` changes:**
1. `readManifest(targetDir)` → becomes a thin wrapper:
   ```typescript
   /** @deprecated Use readStore() from src/store/index.ts */
   export async function readManifest(targetDir: string): Promise<AiSetupConfig | null> {
     try {
       const store = await createStore(targetDir)
       if (!store.data.meta.installedAt) return null
       // convert StoreData back to AiSetupConfig shape for backward compat
       return {
         version: store.data.meta.cliVersion,
         setupType: store.data.config.setupType,
         tools: store.data.config.tools,
         projectName: store.data.config.projectName,
         installedAt: store.data.meta.installedAt,
         files: store.data.files.map(f => ({ path: f.path, hash: f.hash, source: f.source })),
         selections: store.data.selections,
       }
     } catch {
       return null
     }
   }
   ```
2. `extractSelections(manifest: AiSetupConfig)` — **unchanged** (operates on any object with the right shape)

**`types.ts` changes:**
1. Add imports from `src/store/schema.ts`
2. Re-export types that now come from zod schemas: `SetupType`, `ToolId`, `DocsDirId`, `AgentId`, `SkillId`, `PromptId`, `TemplateId`, `RuleId`, `InfraId`, `WizardSelections`
3. Mark `AiSetupConfig` as `@deprecated`:
   ```typescript
   /** @deprecated Use StoreData from src/store/schema.ts */
   export interface AiSetupConfig { ... }
   ```
4. Keep all existing interface exports (`SetupConfig`, `FileRecord`, `WizardConfig`, `ConflictStrategy`, `ArtifactType`) — **do not remove**

**Test updates:**
- `src/__tests__/manifest.test.ts`: `buildManifest()` helper creates a valid v1 store `StoreData` OR still creates legacy v0 format (readManifest wraps, so v0 JSON on disk will be auto-migrated). The tests using `fs.writeFileSync` of legacy JSON should still pass if readManifest wraps correctly.
- `src/__tests__/update-doctor.test.ts`: Reading `JSON.parse(fs.readFileSync(...)) as AiSetupConfig` in the test setup (line 43) — this will need updating to handle the new v1 format. The test reads the file directly. Options: (a) update test to read via `readStore()`, or (b) update assertion to match v1 shape. Prefer (a).

**Acceptance check:**
- `npm run typecheck` passes
- `npm run test` passes (including `manifest.test.ts`, `update-doctor.test.ts`, `wizard-integration.test.ts`)
- `npm run build` succeeds

> **MILESTONE M3:** All commands refactored. Run `npm run typecheck && npm run test && npm run build`.

---

### T13 — Create `src/errors/operation.ts` (OperationTracker)
**Priority:** P3 | **Complexity:** M | **Depends on:** T01, T04

**Spec refs:** §2.6, AC-6

**What to read before implementing:**
- `spec.md §2.6` (OperationTracker behavioral contract)
- `src/store/schema.ts` (T04 — `Operation`, `OperationType`)
- `src/errors/types.ts` (T01 — `AiSetupError`, `Errors`)

**Files to create:**
- `src/errors/operation.ts`

**Implement:**
```typescript
class OperationTracker {
  private _succeeded: string[] = []
  private _failed: Array<{ path: string; error: Error }> = []
  private _backups: Map<string, string> = new Map()

  constructor(private readonly type: OperationType) {}

  trackSuccess(filePath: string): void
  trackFailure(filePath: string, error: Error): void
  registerBackup(originalPath: string, backupPath: string): void

  get succeeded(): string[]
  get failed(): Array<{ path: string; error: Error }>
  get backups(): Map<string, string>
  get result(): 'success' | 'partial' | 'failed'
    // all succeed → 'success'
    // some succeed, some fail → 'partial'
    // all fail OR zero tracked → 'failed'

  toOperation(): Operation  // generate complete Operation record
}
```

**Update `src/errors/index.ts` barrel** to also export `OperationTracker`.

**Acceptance check:**
- `npm run typecheck` passes

---

### T14 — Wire OperationTracker into `update`/`init` + add `--verbose` flag
**Priority:** P3 | **Complexity:** S | **Depends on:** T12, T13

**Spec refs:** §7 P3 items, §2.5 DEBUG mode, AC-6, AC-7

**What to read before implementing:**
- `src/commands/update.ts` (T09 output — needs tracker wired in)
- `src/wizard/index.ts` (T12 output — init needs tracker)
- `src/cli.ts` (add `--verbose` global option)
- `src/errors/operation.ts` (T13 output)
- `spec.md §2.6, AC-6, AC-7`

**Files to modify:**
- `src/commands/update.ts` — wrap file write loop with OperationTracker
- `src/wizard/index.ts` — wrap `installFiles()` with OperationTracker, append to store
- `src/cli.ts` — add `--verbose` global option

**`update.ts` OperationTracker wiring:**
```typescript
const tracker = new OperationTracker('update')
// In the file write loop:
try {
  writeFile(absPath, entry.content)
  tracker.trackSuccess(entry.path)
} catch (err) {
  tracker.trackFailure(entry.path, err instanceof Error ? err : new Error(String(err)))
}
// After loop, if backup was created:
tracker.registerBackup(absPath, backupPath)
// At end, append operation:
await appendOperation(store, tracker.toOperation())
```

**`wizard/index.ts` OperationTracker wiring:**
```typescript
const tracker = new OperationTracker('init')
// wrap installFiles() in try/catch per file — or pass tracker down to scaffold functions
// At end: await appendOperation(store, tracker.toOperation())
```
Note: scaffold functions don't currently accept a tracker. The simplest approach is to track at the top level: `tracker.trackSuccess()` for each file in `fileRecords` after `installFiles()` succeeds.

**`cli.ts` `--verbose` flag:**
```typescript
program.option('--verbose', 'Enable verbose/debug output (shows stack traces on error)')
```
The `boundary.ts` `isDebug()` already checks `process.argv.includes('--verbose')`, so no additional wiring needed.

**Test additions:**
- `src/__tests__/cli.e2e.test.ts`: Add test `it('accepts --verbose flag without error', ...)` — run `node bin/ai-setup.js --verbose --help` and verify exit 0.

**Acceptance check:**
- `npm run typecheck` passes
- `npm run test` passes (new e2e test for `--verbose`)
- `npm run build` passes

> **MILESTONE M4:** Full spec complete. Run full quality gate.

---

## Quality Gate Checkpoints

| After | Command | Expected |
|-------|---------|----------|
| T03 (M1) | `npm run typecheck && npm run test` | All existing tests pass, new error tests pass |
| T06 (M2) | `npm run typecheck && npm run test` | + store and migration tests pass |
| T12 (M3) | `npm run typecheck && npm run test && npm run build` | Full suite green, dist built |
| T14 (M4) | `npm run typecheck && npm run test && npm run build` | All AC verified |

---

## Dependency Install (do first)

Before any task, add runtime dependencies:

```bash
npm install lowdb@^7.0.0 zod@^3.23.0
```

Verify `package.json` `"dependencies"` section includes both. Verify `import { Low } from 'lowdb'` and `import { JSONFile } from 'lowdb/node'` resolve correctly (lowdb v7 has separate node adapter).

**Test import compatibility:**
```typescript
// ESM check — this must work:
import { Low, Memory } from 'lowdb'
import { JSONFile } from 'lowdb/node'
import { z } from 'zod'
```

---

## Acceptance Criteria Mapping

| AC | Tasks | Verification |
|----|-------|-------------|
| AC-1: Store reads/writes validated data | T04, T06 | `store.test.ts` — all 7 cases green |
| AC-2: Legacy migration works silently | T05, T06, T12 | `store-migrations.test.ts` round-trip test; run `ai-setup doctor` against v0 JSON |
| AC-3: All errors are typed | T01–T03, T07–T12 | `npm run typecheck`; grep for `throw new Error` → zero hits in modified files; grep for `process.exit` → only in `src/index.ts` and `src/errors/boundary.ts` |
| AC-4: User experience unchanged | T07–T12 | `update-doctor.test.ts`, `wizard-integration.test.ts`, `cli.e2e.test.ts` all green |
| AC-5: Quality gates pass | All | `npm run typecheck && npm run test && npm run build` |
| AC-6: Operation tracking | T13, T14 | Inspect `.ai-setup.json` after `update` — `operations` array has one entry with correct shape; run `update` 51 times → array stays ≤ 50 |
| AC-7: DEBUG mode | T02, T14 | `AI_SETUP_DEBUG=1 ai-setup doctor` on broken setup → stack trace visible; normal run → no stack; `--verbose` equivalent |

---

## Open Questions Resolutions (from spec §8)

All 6 proposed resolutions are accepted as stated:
- **Q1:** `config.targetDir` stored as absolute path
- **Q2:** No backup of original `.ai-setup.json` before migration (lowdb/steno handles atomicity)
- **Q3:** `readManifest()` kept as thin wrapper, deprecated via JSDoc
- **Q4:** `force` stays in `WizardConfig` (not persisted in `StoreData`)
- **Q5:** `TrackedFile.status` only recomputed on `doctor`/`update`, not on every read
- **Q6:** `.ai-setup.json` schema change treated as internal (v0.1.0 pre-semver-1.0)

---

## File Touch Map (complete)

### New files
| File | Task |
|------|------|
| `src/errors/types.ts` | T01 |
| `src/errors/boundary.ts` | T02 |
| `src/errors/index.ts` | T03 |
| `src/store/schema.ts` | T04 |
| `src/store/migrations.ts` | T05 |
| `src/store/index.ts` | T06 |
| `src/errors/operation.ts` | T13 |
| `src/__tests__/errors.test.ts` | T01, T02 |
| `src/__tests__/store.test.ts` | T06 |
| `src/__tests__/store-migrations.test.ts` | T05 |

### Modified files
| File | Task | Nature of change |
|------|------|-----------------|
| `package.json` | Pre-work | Add `lowdb`, `zod` to `dependencies` |
| `src/index.ts` | T03 | Replace catch block with `handleError` |
| `src/types.ts` | T12 | Re-export from store schema, deprecate `AiSetupConfig` |
| `src/cli.ts` | T14 | Add `--verbose` global option |
| `src/utils/files.ts` | T07 | Replace `throw new Error` with `Errors.*` |
| `src/utils/conflicts.ts` | T07 | Replace `process.exit(0)` with `throw Errors.userCancelled()` |
| `src/utils/manifest.ts` | T12 | `readManifest()` as thin wrapper over `readStore()` |
| `src/commands/doctor.ts` | T08 | Use `readStore()`, update file statuses, remove `process.exit` |
| `src/commands/update.ts` | T09 | Use `createStore()`, `writeStore()`, remove `process.exit` |
| `src/commands/add.ts` | T10 | Use `createStore()`, `writeStore()`, remove `process.exit` |
| `src/commands/eject.ts` | T11 | Typed catch, `userCancelled()` |
| `src/commands/create.ts` | T11 | Replace all `process.exit(0)` with `userCancelled()`, `Error` → `Errors.*` |
| `src/wizard/index.ts` | T12 | Write via `writeStore()`, throw instead of `process.exit` |
| `src/__tests__/manifest.test.ts` | T12 | Update if `readManifest` wrapper changes test expectations |
| `src/__tests__/update-doctor.test.ts` | T09 | Update to read store via `readStore()` instead of raw JSON |
| `src/__tests__/cli.e2e.test.ts` | T14 | Add `--verbose` acceptance test |

### Unchanged files (confirmed)
`src/commands/status.ts`, `src/commands/init.ts`, `src/adapters/*`, `src/generators/*`, `src/scaffold/*`, `src/wizard/phase*.ts`, `src/prompts.ts`, `src/utils/diff.ts`, `src/utils/frontmatter.ts`, `src/utils/validation.ts`

---

## Complexity Summary

| Task | Complexity | Rationale |
|------|------------|-----------|
| T01 | M | 15 error codes, class, 13 factory methods |
| T02 | M | Error boundary with multiple branches, DEBUG mode |
| T03 | S | 2-line change to index.ts + barrel file |
| T04 | M | 9 zod enums + 7 schemas + defaultStore |
| T05 | M | Migration logic + field-by-field mapping + 8 tests |
| T06 | M | lowdb wiring, 5 exported functions, 7 tests |
| T07 | S | Mechanical replacement in 2 small files |
| T08 | S | 70-line file, small structural changes |
| T09 | M | 305-line file, structural access pattern change throughout |
| T10 | S | 71-line file, straightforward |
| T11 | S | 51+368 lines, mechanical process.exit replacement |
| T12 | M | Wizard (230 lines) + manifest wrapper + types.ts re-exports |
| T13 | M | New class with getters and toOperation() logic |
| T14 | S | Wire existing pieces, add --verbose to Commander |
