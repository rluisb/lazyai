import { EventEmitter } from 'node:events'
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('node:child_process', () => ({
  spawn: vi.fn(),
}))

import { spawn } from 'node:child_process'
import {
  DiffViewerDelegate,
  type ReviewRequest,
  shouldDelegateReview,
} from '../utils/diffviewer-delegate.js'

function makeRequest(): ReviewRequest {
  return {
    version: 1,
    files: [
      {
        path: 'AGENTS.md',
        currentContent: 'old',
        newContent: 'new',
      },
    ],
  }
}

function mockSpawnResponse(stdout: string, code = 0) {
  let stdinPayload = ''
  const child = new EventEmitter() as EventEmitter & {
    stdout: EventEmitter
    stderr: EventEmitter
    stdin: { end: (payload: string) => void }
  }

  child.stdout = new EventEmitter()
  child.stderr = new EventEmitter()
  child.stdin = {
    end: vi.fn((payload: string) => {
      stdinPayload = payload
      queueMicrotask(() => {
        child.stdout.emit('data', stdout)
        child.emit('close', code)
      })
    }),
  }

  vi.mocked(spawn).mockReturnValue(child as never)
  return { getStdinPayload: () => stdinPayload }
}

describe('diffviewer delegation', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('does not delegate for one file with fewer than 20 changed lines', () => {
    expect(
      shouldDelegateReview([
        {
          path: 'one.md',
          currentContent: 'line 1\nline 2',
          newContent: 'line 1\nchanged line 2',
        },
      ]),
    ).toBe(false)
  })

  it('delegates for multiple files or one file with at least 20 changed lines', () => {
    expect(
      shouldDelegateReview([
        { path: 'one.md', currentContent: 'old', newContent: 'new' },
        { path: 'two.md', currentContent: 'old', newContent: 'new' },
      ]),
    ).toBe(true)

    expect(
      shouldDelegateReview([
        {
          path: 'large.md',
          currentContent: Array.from({ length: 20 }, (_, i) => `old ${i}`).join('\n'),
          newContent: Array.from({ length: 20 }, (_, i) => `new ${i}`).join('\n'),
        },
      ]),
    ).toBe(true)
  })

  it('falls back when the binary is not found', async () => {
    const delegate = new DiffViewerDelegate({ resolveBinary: () => null })

    await expect(delegate.runDiffReview(makeRequest())).resolves.toEqual({
      mode: 'fallback',
      error: 'diffviewer binary not found',
    })
    expect(spawn).not.toHaveBeenCalled()
  })

  it('returns delegated cancelled status for a valid cancelled response', async () => {
    const spawned = mockSpawnResponse(JSON.stringify({ version: 1, status: 'cancelled', resolutions: [] }))
    const delegate = new DiffViewerDelegate({ resolveBinary: () => '/mock/diffviewer' })

    await expect(delegate.runDiffReview(makeRequest())).resolves.toEqual({
      mode: 'delegated',
      status: 'cancelled',
      resolutions: [],
    })
    expect(spawn).toHaveBeenCalledWith('/mock/diffviewer', [], { stdio: ['pipe', 'pipe', 'pipe'] })
    expect(JSON.parse(spawned.getStdinPayload())).toEqual(makeRequest())
  })

  it('falls back when the response is malformed', async () => {
    mockSpawnResponse('{not-json')
    const delegate = new DiffViewerDelegate({ resolveBinary: () => '/mock/diffviewer' })

    const result = await delegate.runDiffReview(makeRequest())

    expect(result.mode).toBe('fallback')
    expect(result.error).toContain('invalid diffviewer response')
  })
})
