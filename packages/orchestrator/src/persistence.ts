import fs from 'node:fs'
import path from 'node:path'
import type { ChainState, ErrorJournalEntry, ExecutionPlan, HandoffDocument, MaintenanceContractRecord, SyncStateSnapshot, TeamState, WorkflowState } from './types.js'
import { openDatabase } from './db/index.js'
import type { Db } from './db/index.js'
import { runMigrations } from './db/migrations.js'
import { getDatabasePath } from './config/paths.js'

// ---------------------------------------------------------------------------
// DB singleton
// ---------------------------------------------------------------------------

let _db: Db | null = null

function resolveDbPath(): string {
  if (process.env.AI_SETUP_ORCHESTRATOR_DB) return process.env.AI_SETUP_ORCHESTRATOR_DB
  if (process.env.NODE_ENV === 'test') return ':memory:'
  return getDatabasePath()
}

function getDb(): Db {
  if (!_db) {
    _db = openDatabase(resolveDbPath())
    runMigrations(_db)
  }
  return _db
}

export function getPersistenceDb(): Db {
  return getDb()
}

export function initPersistenceDb(dbPath: string): Db {
  _db = openDatabase(dbPath)
  runMigrations(_db)
  return _db
}

export function closePersistenceDb(): void {
  _db?.close()
  _db = null
}

// ---------------------------------------------------------------------------
// Legacy path helpers — kept for backward compat and the importer
// ---------------------------------------------------------------------------

export function getStateRoot(projectRoot: string): string {
  return path.join(projectRoot, '.ai', 'orchestration', 'state')
}

export function getChainStatePath(projectRoot: string, chainId: string): string {
  return path.join(getStateRoot(projectRoot), 'chains', `${chainId}.json`)
}

export function getExecutionPlanPath(projectRoot: string, planId: string): string {
  return path.join(getStateRoot(projectRoot), 'plans', `${planId}.json`)
}

export function getTeamStatePath(projectRoot: string, teamId: string): string {
  return path.join(getStateRoot(projectRoot), 'teams', `${teamId}.json`)
}

export function getWorkflowStatePath(projectRoot: string, workflowId: string): string {
  return path.join(getStateRoot(projectRoot), 'workflows', `${workflowId}.json`)
}

export function getHandoffPath(projectRoot: string, handoffId: string): string {
  return path.join(getStateRoot(projectRoot), 'handoffs', `${handoffId}.json`)
}

export function getErrorJournalPath(projectRoot: string): string {
  return path.join(getStateRoot(projectRoot), 'error-journal.jsonl')
}

export function ensureStateDir(_projectRoot: string): string {
  return getStateRoot(_projectRoot)
}

export function getHousekeepingRoot(projectRoot: string): string {
  return path.join(projectRoot, '.ai', 'housekeeping')
}

export function getSyncStatePath(projectRoot: string): string {
  return path.join(getHousekeepingRoot(projectRoot), 'sync-state.json')
}

export function getContractsRoot(projectRoot: string): string {
  return path.join(getHousekeepingRoot(projectRoot), 'contracts')
}

export function readSyncState(projectRoot: string): SyncStateSnapshot | null {
  const filePath = getSyncStatePath(projectRoot)
  if (!fs.existsSync(filePath)) return null
  return JSON.parse(fs.readFileSync(filePath, 'utf-8')) as SyncStateSnapshot
}

export function writeSyncState(projectRoot: string, state: SyncStateSnapshot): void {
  fs.mkdirSync(getHousekeepingRoot(projectRoot), { recursive: true })
  fs.writeFileSync(getSyncStatePath(projectRoot), JSON.stringify(state, null, 2), 'utf-8')
}

export function readMaintenanceContracts(projectRoot: string, now = new Date().toISOString()): MaintenanceContractRecord[] {
  const contractsRoot = getContractsRoot(projectRoot)
  if (!fs.existsSync(contractsRoot)) return []

  return fs
    .readdirSync(contractsRoot)
    .filter((entry) => entry.endsWith('.json'))
    .map((entry) => JSON.parse(fs.readFileSync(path.join(contractsRoot, entry), 'utf-8')) as MaintenanceContractRecord)
    .filter((contract) => !contract.approvalExpiresAt || contract.approvalExpiresAt >= now)
}

export function readJsonLines<T>(filePath: string): T[] {
  if (!fs.existsSync(filePath)) return []
  return fs
    .readFileSync(filePath, 'utf-8')
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => JSON.parse(line) as T)
}

// ---------------------------------------------------------------------------
// Chain runs
// ---------------------------------------------------------------------------

const insertChainRun = (db: Db) =>
  db.prepare<[string, string, string | null, string, string | null, string, string, string, string]>(`
    INSERT INTO chain_runs (id, definition_name, definition_version, state, current_step_id, project_root, state_json, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(id) DO UPDATE SET
      state = excluded.state,
      current_step_id = excluded.current_step_id,
      state_json = excluded.state_json,
      updated_at = excluded.updated_at
  `)

const selectChainRun = (db: Db) =>
  db.prepare<[string], { state_json: string }>('SELECT state_json FROM chain_runs WHERE id = ?')

export function saveChainState(projectRoot: string, state: ChainState): void {
  const db = getDb()
  insertChainRun(db).run(
    state.chainId,
    state.definitionName,
    state.definitionVersion ?? null,
    state.state,
    state.currentStepId ?? null,
    projectRoot,
    JSON.stringify(state),
    state.createdAt,
    state.updatedAt,
  )
}

export function loadChainState(projectRoot: string, chainId: string): ChainState {
  const row = selectChainRun(getDb()).get(chainId)
  if (!row) throw new Error(`Chain state not found: ${chainId} (project: ${projectRoot})`)
  return JSON.parse(row.state_json) as ChainState
}

// ---------------------------------------------------------------------------
// Team runs
// ---------------------------------------------------------------------------

const insertTeamRun = (db: Db) =>
  db.prepare<[string, string, string | null, string, string, string, string, string]>(`
    INSERT INTO team_runs (id, definition_name, definition_version, state, project_root, state_json, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(id) DO UPDATE SET
      state = excluded.state,
      state_json = excluded.state_json,
      updated_at = excluded.updated_at
  `)

const selectTeamRun = (db: Db) =>
  db.prepare<[string], { state_json: string }>('SELECT state_json FROM team_runs WHERE id = ?')

export function saveTeamState(projectRoot: string, state: TeamState): void {
  const db = getDb()
  insertTeamRun(db).run(
    state.teamId,
    state.definitionName,
    state.definitionVersion ?? null,
    state.state,
    projectRoot,
    JSON.stringify(state),
    state.createdAt,
    state.updatedAt,
  )
}

export function loadTeamState(projectRoot: string, teamId: string): TeamState {
  const row = selectTeamRun(getDb()).get(teamId)
  if (!row) throw new Error(`Team state not found: ${teamId} (project: ${projectRoot})`)
  return JSON.parse(row.state_json) as TeamState
}

// ---------------------------------------------------------------------------
// Workflow runs
// ---------------------------------------------------------------------------

const insertWorkflowRun = (db: Db) =>
  db.prepare<[string, string, string | null, string, string | null, string, string, string, string]>(`
    INSERT INTO workflow_runs (id, definition_name, definition_version, state, current_phase_id, project_root, state_json, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(id) DO UPDATE SET
      state = excluded.state,
      current_phase_id = excluded.current_phase_id,
      state_json = excluded.state_json,
      updated_at = excluded.updated_at
  `)

const selectWorkflowRun = (db: Db) =>
  db.prepare<[string], { state_json: string }>('SELECT state_json FROM workflow_runs WHERE id = ?')

export function saveWorkflowState(projectRoot: string, state: WorkflowState): void {
  const db = getDb()
  insertWorkflowRun(db).run(
    state.workflowId,
    state.definitionName,
    state.definitionVersion ?? null,
    state.state,
    state.currentPhaseId ?? null,
    projectRoot,
    JSON.stringify(state),
    state.createdAt,
    state.updatedAt,
  )
}

export function loadWorkflowState(projectRoot: string, workflowId: string): WorkflowState {
  const row = selectWorkflowRun(getDb()).get(workflowId)
  if (!row) throw new Error(`Workflow state not found: ${workflowId} (project: ${projectRoot})`)
  return JSON.parse(row.state_json) as WorkflowState
}

// ---------------------------------------------------------------------------
// Execution plans
// ---------------------------------------------------------------------------

const insertPlan = (db: Db) =>
  db.prepare<[string, string, string, string | null, string, string, string]>(`
    INSERT OR IGNORE INTO execution_plans (id, kind, definition_name, definition_version, project_root, plan_json, created_at)
    VALUES (?, ?, ?, ?, ?, ?, ?)
  `)

const selectPlan = (db: Db) =>
  db.prepare<[string], { plan_json: string }>('SELECT plan_json FROM execution_plans WHERE id = ?')

export function saveExecutionPlan(projectRoot: string, plan: ExecutionPlan): void {
  insertPlan(getDb()).run(
    plan.id,
    plan.kind,
    plan.definition.name,
    plan.definition.version ?? null,
    projectRoot,
    JSON.stringify(plan),
    plan.createdAt,
  )
}

export function loadExecutionPlan(projectRoot: string, planId: string): ExecutionPlan {
  const row = selectPlan(getDb()).get(planId)
  if (!row) throw new Error(`Execution plan not found: ${planId} (project: ${projectRoot})`)
  return JSON.parse(row.plan_json) as ExecutionPlan
}

// ---------------------------------------------------------------------------
// Handoffs
// ---------------------------------------------------------------------------

const insertHandoff = (db: Db) =>
  db.prepare<[string, string, string, string, string]>(`
    INSERT OR IGNORE INTO handoffs (id, run_id, run_kind, doc_json, created_at)
    VALUES (?, ?, ?, ?, ?)
  `)

export function saveHandoff(projectRoot: string, handoff: HandoffDocument): string {
  insertHandoff(getDb()).run(
    handoff.id,
    handoff.runId,
    handoff.kind,
    JSON.stringify(handoff),
    handoff.createdAt,
  )
  return getHandoffPath(projectRoot, handoff.id)
}

// ---------------------------------------------------------------------------
// Error journal (DB-backed; kept here alongside the other save/load pairs)
// ---------------------------------------------------------------------------

const insertJournalEntry = (db: Db) =>
  db.prepare<[string, string | null, string | null, string, string | null, string, string, string, string, string]>(`
    INSERT OR IGNORE INTO error_journal
      (id, run_id, run_kind, definition_name, step_id, category, code, message, entry_json, created_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `)

const selectJournalByProject = (db: Db) =>
  db.prepare<[string], { entry_json: string }>('SELECT entry_json FROM error_journal WHERE definition_name = ? ORDER BY rowid ASC')

const selectAllJournal = (db: Db) =>
  db.prepare<[], { entry_json: string }>('SELECT entry_json FROM error_journal ORDER BY rowid ASC')

export function saveErrorJournalEntry(entry: ErrorJournalEntry): void {
  insertJournalEntry(getDb()).run(
    entry.id,
    entry.runId ?? null,
    entry.runKind ?? null,
    entry.definitionName,
    entry.stepId ?? null,
    entry.error.category,
    entry.error.code,
    entry.error.message,
    JSON.stringify(entry),
    entry.error.timestamp,
  )
}

export function loadErrorJournalEntries(projectRoot: string): ErrorJournalEntry[] {
  const db = getDb()
  void projectRoot
  const rows = selectAllJournal(db).all()
  return rows.map((r) => JSON.parse(r.entry_json) as ErrorJournalEntry)
}
