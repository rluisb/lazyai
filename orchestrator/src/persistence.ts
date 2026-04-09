import fs from 'node:fs'
import path from 'node:path'
import type { ChainState, ExecutionPlan, HandoffDocument, TeamState, WorkflowState } from './types.js'

export function ensureStateDir(projectRoot: string): string {
  const stateRoot = getStateRoot(projectRoot)
  fs.mkdirSync(path.join(stateRoot, 'chains'), { recursive: true })
  fs.mkdirSync(path.join(stateRoot, 'teams'), { recursive: true })
  fs.mkdirSync(path.join(stateRoot, 'workflows'), { recursive: true })
  fs.mkdirSync(path.join(stateRoot, 'plans'), { recursive: true })
  fs.mkdirSync(path.join(stateRoot, 'handoffs'), { recursive: true })
  return stateRoot
}

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

export function saveChainState(projectRoot: string, state: ChainState): void {
  ensureStateDir(projectRoot)
  fs.writeFileSync(getChainStatePath(projectRoot, state.chainId), JSON.stringify(state, null, 2), 'utf-8')
}

export function loadChainState(projectRoot: string, chainId: string): ChainState {
  return JSON.parse(fs.readFileSync(getChainStatePath(projectRoot, chainId), 'utf-8')) as ChainState
}

export function saveTeamState(projectRoot: string, state: TeamState): void {
  ensureStateDir(projectRoot)
  fs.writeFileSync(getTeamStatePath(projectRoot, state.teamId), JSON.stringify(state, null, 2), 'utf-8')
}

export function loadTeamState(projectRoot: string, teamId: string): TeamState {
  return JSON.parse(fs.readFileSync(getTeamStatePath(projectRoot, teamId), 'utf-8')) as TeamState
}

export function saveWorkflowState(projectRoot: string, state: WorkflowState): void {
  ensureStateDir(projectRoot)
  fs.writeFileSync(getWorkflowStatePath(projectRoot, state.workflowId), JSON.stringify(state, null, 2), 'utf-8')
}

export function loadWorkflowState(projectRoot: string, workflowId: string): WorkflowState {
  return JSON.parse(fs.readFileSync(getWorkflowStatePath(projectRoot, workflowId), 'utf-8')) as WorkflowState
}

export function saveExecutionPlan(projectRoot: string, plan: ExecutionPlan): void {
  ensureStateDir(projectRoot)
  fs.writeFileSync(getExecutionPlanPath(projectRoot, plan.id), JSON.stringify(plan, null, 2), 'utf-8')
}

export function loadExecutionPlan(projectRoot: string, planId: string): ExecutionPlan {
  return JSON.parse(fs.readFileSync(getExecutionPlanPath(projectRoot, planId), 'utf-8')) as ExecutionPlan
}

export function saveHandoff(projectRoot: string, handoff: HandoffDocument): string {
  ensureStateDir(projectRoot)
  const filePath = getHandoffPath(projectRoot, handoff.id)
  fs.writeFileSync(filePath, JSON.stringify(handoff, null, 2), 'utf-8')
  return filePath
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
