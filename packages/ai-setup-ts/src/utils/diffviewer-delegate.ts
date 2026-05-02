import { spawn } from 'node:child_process'
import { existsSync, statSync } from 'node:fs'
import { dirname, delimiter, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { computeLineDiff } from './diff.js'

const REVIEW_CONTRACT_VERSION = 1
const LARGE_DIFF_CHANGED_LINES = 20
const DIFFVIEWER_BINARY_ENV = 'AI_SETUP_DIFFVIEWER_BINARY'

export interface ReviewRequest {
  version: 1
  title?: string
  files: ReviewFile[]
}

export interface ReviewFile {
  path: string
  currentContent: string
  newContent: string
}

export interface DiffFile extends ReviewFile {
  changedLines?: number
}

export type ReviewStatus = 'confirmed' | 'cancelled'
export type ResolutionAction = 'accept' | 'deny' | 'skip'

export interface Resolution {
  path: string
  action: ResolutionAction
}

export interface ReviewResponse {
  version: 1
  status: ReviewStatus
  resolutions: Resolution[]
  message?: string
}

export type DelegateResult = {
  mode: 'delegated' | 'fallback'
  status?: ReviewStatus
  resolutions?: Resolution[]
  error?: string
}

export interface DiffViewerDelegateOptions {
  resolveBinary?: () => string | null
}

export class DiffViewerDelegate {
  private readonly binaryResolver: (() => string | null) | undefined

  constructor(options: DiffViewerDelegateOptions = {}) {
    this.binaryResolver = options.resolveBinary
  }

  shouldDelegateReview(files: DiffFile[]): boolean {
    return shouldDelegateReview(files)
  }

  resolveDiffViewerBinary(): string | null {
    return this.binaryResolver ? this.binaryResolver() : resolveDiffViewerBinary()
  }

  async runDiffReview(request: ReviewRequest): Promise<DelegateResult> {
    const binary = this.resolveDiffViewerBinary()
    if (!binary) {
      return fallback('diffviewer binary not found')
    }

    return executeDiffViewer(binary, request)
  }
}

export function shouldDelegateReview(files: DiffFile[]): boolean {
  if (files.length > 1) return true

  return files.some((file) => changedLineCount(file) >= LARGE_DIFF_CHANGED_LINES)
}

export function resolveDiffViewerBinary(): string | null {
  const envBinary = process.env[DIFFVIEWER_BINARY_ENV]?.trim()
  if (envBinary) {
    const resolvedEnvBinary = findExecutableCandidate(envBinary)
    if (resolvedEnvBinary) return resolvedEnvBinary
  }

  for (const candidate of diffViewerBinaryCandidates()) {
    if (isExecutableFile(candidate)) return candidate
  }

  return null
}

export async function runDiffReview(request: ReviewRequest): Promise<DelegateResult> {
  return new DiffViewerDelegate().runDiffReview(request)
}

function changedLineCount(file: DiffFile): number {
  if (typeof file.changedLines === 'number' && Number.isFinite(file.changedLines)) {
    return Math.max(0, Math.floor(file.changedLines))
  }

  const diff = computeLineDiff(file.currentContent, file.newContent)
  return diff.stats.additions + diff.stats.deletions
}

function diffViewerBinaryCandidates(): string[] {
  const executableName = diffViewerExecutableName()
  const moduleDir = dirname(fileURLToPath(import.meta.url))
  const repoRootFromModule = resolve(moduleDir, '../../../..')
  const cwd = process.cwd()

  return uniqueStrings([
    resolve(cwd, 'packages/diffviewer/bin', executableName),
    resolve(cwd, 'packages/diffviewer', executableName),
    resolve(cwd, 'packages/diffviewer/cmd/diffviewer', executableName),
    resolve(repoRootFromModule, 'packages/diffviewer/bin', executableName),
    resolve(repoRootFromModule, 'packages/diffviewer', executableName),
    resolve(repoRootFromModule, 'packages/diffviewer/cmd/diffviewer', executableName),
    ...pathExecutableCandidates(executableName),
  ])
}

function diffViewerExecutableName(): string {
  return process.platform === 'win32' ? 'diffviewer.exe' : 'diffviewer'
}

function pathExecutableCandidates(executableName: string): string[] {
  return (process.env.PATH ?? '')
    .split(delimiter)
    .filter((entry) => entry.length > 0)
    .map((entry) => resolve(entry, executableName))
}

function findExecutableCandidate(candidate: string): string | null {
  if (isExecutableFile(candidate)) return candidate

  if (!candidate.includes('/') && !candidate.includes('\\')) {
    for (const pathCandidate of pathExecutableCandidates(candidate)) {
      if (isExecutableFile(pathCandidate)) return pathCandidate
    }
  }

  return null
}

function isExecutableFile(candidate: string): boolean {
  if (!existsSync(candidate)) return false

  try {
    return statSync(candidate).isFile()
  } catch {
    return false
  }
}

function uniqueStrings(values: string[]): string[] {
  return [...new Set(values)]
}

function executeDiffViewer(binary: string, request: ReviewRequest): Promise<DelegateResult> {
  return new Promise((resolveResult) => {
    let stdout = ''
    let stderr = ''
    let settled = false

    const finish = (result: DelegateResult) => {
      if (settled) return
      settled = true
      resolveResult(result)
    }

    let child: ReturnType<typeof spawn>
    try {
      child = spawn(binary, [], { stdio: ['pipe', 'pipe', 'pipe'] })
    } catch (error) {
      finish(fallback(`diffviewer launch failed: ${errorMessage(error)}`))
      return
    }

    if (!child.stdin || !child.stdout || !child.stderr) {
      finish(fallback('diffviewer launch failed: stdio pipes unavailable'))
      return
    }

    child.stdout.on('data', (chunk: Buffer | string) => {
      stdout += chunk.toString()
    })

    child.stderr.on('data', (chunk: Buffer | string) => {
      stderr += chunk.toString()
    })

    child.on('error', (error) => {
      finish(fallback(`diffviewer execution failed: ${errorMessage(error)}`))
    })

    child.on('close', (code) => {
      if (code !== 0) {
        finish(fallback(formatProcessFailure(code, stderr)))
        return
      }

      finish(parseDelegateResponse(stdout))
    })

    try {
      child.stdin.end(`${JSON.stringify(request)}\n`)
    } catch (error) {
      finish(fallback(`diffviewer request write failed: ${errorMessage(error)}`))
    }
  })
}

function parseDelegateResponse(stdout: string): DelegateResult {
  let parsed: unknown

  try {
    parsed = JSON.parse(stdout.trim())
  } catch (error) {
    return fallback(`invalid diffviewer response: ${errorMessage(error)}`)
  }

  const response = validateReviewResponse(parsed)
  if (!response.valid) {
    return fallback(`invalid diffviewer response: ${response.error}`)
  }

  return {
    mode: 'delegated',
    status: response.value.status,
    resolutions: response.value.resolutions,
  }
}

function validateReviewResponse(value: unknown): { valid: true; value: ReviewResponse } | { valid: false; error: string } {
  if (!isRecord(value)) return { valid: false, error: 'response must be an object' }
  if (!hasOnlyKeys(value, ['version', 'status', 'resolutions', 'message'])) {
    return { valid: false, error: 'response contains unsupported fields' }
  }
  if (value.version !== REVIEW_CONTRACT_VERSION) {
    return { valid: false, error: `version must be ${REVIEW_CONTRACT_VERSION}` }
  }
  if (value.status !== 'confirmed' && value.status !== 'cancelled') {
    return { valid: false, error: 'status must be confirmed or cancelled' }
  }
  const status = value.status
  if (!Array.isArray(value.resolutions)) {
    return { valid: false, error: 'resolutions must be an array' }
  }
  if ('message' in value && (typeof value.message !== 'string' || value.message.length === 0)) {
    return { valid: false, error: 'message must be a non-empty string when provided' }
  }
  const message = typeof value.message === 'string' ? value.message : undefined

  const resolutions: Resolution[] = []
  for (let i = 0; i < value.resolutions.length; i++) {
    const resolution = validateResolution(value.resolutions[i], i)
    if (!resolution.valid) return resolution
    resolutions.push(resolution.value)
  }

  return {
    valid: true,
    value: {
      version: REVIEW_CONTRACT_VERSION,
      status,
      resolutions,
      ...(message ? { message } : {}),
    },
  }
}

function validateResolution(value: unknown, index: number): { valid: true; value: Resolution } | { valid: false; error: string } {
  if (!isRecord(value)) return { valid: false, error: `resolutions[${index}] must be an object` }
  if (!hasOnlyKeys(value, ['path', 'action'])) {
    return { valid: false, error: `resolutions[${index}] contains unsupported fields` }
  }
  if (typeof value.path !== 'string' || value.path.length === 0) {
    return { valid: false, error: `resolutions[${index}].path must be a non-empty string` }
  }
  if (value.action !== 'accept' && value.action !== 'deny' && value.action !== 'skip') {
    return { valid: false, error: `resolutions[${index}].action must be accept, deny, or skip` }
  }

  return { valid: true, value: { path: value.path, action: value.action } }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function hasOnlyKeys(value: Record<string, unknown>, allowedKeys: string[]): boolean {
  const allowed = new Set(allowedKeys)
  return Object.keys(value).every((key) => allowed.has(key))
}

function formatProcessFailure(code: number | null, stderr: string): string {
  const details = stderr.trim()
  const prefix = code === null ? 'diffviewer terminated before exit code' : `diffviewer exited with code ${code}`
  return details ? `${prefix}: ${details}` : prefix
}

function fallback(error: string): DelegateResult {
  return { mode: 'fallback', error }
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error)
}
