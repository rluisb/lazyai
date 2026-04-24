/**
 * Error boundary tests
 *
 * Tests the central handleError function and AiSetupError class behavior.
 * handleError calls process.exit — we spy on it to avoid actually exiting.
 */
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { handleError } from '../errors/boundary.js'
import { AiSetupError, ErrorCodeEnum, Errors } from '../errors/index.js'

// We must mock process.exit to prevent test process from exiting
vi.spyOn(process, 'exit').mockImplementation((_code?: number | string | null) => {
  throw new Error(`process.exit(${_code})`)
})

describe('AiSetupError', () => {
  it('creates error with correct code', () => {
    const err = Errors.fileNotFound('/path/to/file')
    expect(err).toBeInstanceOf(AiSetupError)
    expect(err.code).toBe(ErrorCodeEnum.FILE_NOT_FOUND)
    expect(err.message).toContain('/path/to/file')
  })

  it('marks USER_CANCELLED and INVALID_INPUT as user errors', () => {
    expect(Errors.userCancelled().isUserError).toBe(true)
    expect(Errors.invalidInput('bad value').isUserError).toBe(true)
  })

  it('marks system errors as non-user errors', () => {
    expect(Errors.fileNotFound('/x').isUserError).toBe(false)
    expect(Errors.manifestCorrupt('/dir').isUserError).toBe(false)
    expect(Errors.unknown('oops').isUserError).toBe(false)
  })

  it('USER_CANCELLED has exitCode 0, others have exitCode 1', () => {
    expect(Errors.userCancelled().exitCode).toBe(0)
    expect(Errors.fileNotFound('/x').exitCode).toBe(1)
    expect(Errors.manifestCorrupt('/dir').exitCode).toBe(1)
  })

  it('preserves cause chain', () => {
    const cause = new Error('disk full')
    const err = Errors.migrationFailed('0', '1', cause)
    expect(err.cause).toBe(cause)
    expect(err.cause?.message).toBe('disk full')
  })

  it('error context is accessible', () => {
    const err = Errors.hashMismatch('/path/file.md', 'expected', 'actual')
    expect(err.context.path).toBe('/path/file.md')
    expect(err.context.expected).toBe('expected')
    expect(err.context.actual).toBe('actual')
  })
})

describe('handleError', () => {
  let consoleSpy: ReturnType<typeof vi.spyOn>

  beforeEach(() => {
    consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined)
    delete process.env.AI_SETUP_DEBUG
  })

  afterEach(() => {
    consoleSpy.mockRestore()
    delete process.env.AI_SETUP_DEBUG
  })

  it('exits 0 for USER_CANCELLED', () => {
    const err = Errors.userCancelled()
    expect(() => handleError(err)).toThrow('process.exit(0)')
  })

  it('exits 1 for regular AiSetupError', () => {
    const err = Errors.fileNotFound('/missing')
    expect(() => handleError(err)).toThrow('process.exit(1)')
  })

  it('exits 1 for plain Error', () => {
    const err = new Error('something went wrong')
    expect(() => handleError(err)).toThrow('process.exit(1)')
  })

  it('exits 1 for unknown value (string)', () => {
    expect(() => handleError('unexpected string error')).toThrow('process.exit(1)')
  })

  it('exits 0 for symbol (clack prompts cancel)', () => {
    const cancelSymbol = Symbol('cancel')
    expect(() => handleError(cancelSymbol)).toThrow('process.exit(0)')
  })

  it('shows error message to stderr', () => {
    const err = Errors.fileNotFound('/test-path')
    try {
      handleError(err)
    } catch {
      // swallow mocked exit throw
    }
    expect(consoleSpy).toHaveBeenCalledWith(
      expect.stringContaining('/test-path')
    )
  })

  it('shows debug context when AI_SETUP_DEBUG=1', () => {
    process.env.AI_SETUP_DEBUG = '1'
    const cause = new Error('root cause')
    const err = Errors.migrationFailed('0', '1', cause)
    try {
      handleError(err)
    } catch {
      // swallow mocked exit throw
    }
    // Debug mode should print error code + context + cause
    const calls = consoleSpy.mock.calls.flat().join(' ')
    expect(calls).toContain('MIGRATION_FAILED')
  })

  it('does NOT show stack in non-debug mode for user errors', () => {
    const err = Errors.invalidInput('bad arg')
    try {
      handleError(err)
    } catch {
      // swallow
    }
    const calls = consoleSpy.mock.calls.flat().join(' ')
    expect(calls).not.toContain('at ')
  })
})
