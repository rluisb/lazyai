/**
 * Error types and factories for ai-setup CLI
 * Provides structured error handling with codes, context, and recovery info
 */

/**
 * Enumeration of all error codes in the system
 * Each code represents a distinct failure scenario
 */
export const ErrorCodeEnum = {
  FILE_NOT_FOUND: 'FILE_NOT_FOUND',
  FILE_PERMISSION: 'FILE_PERMISSION',
  FILE_CORRUPT: 'FILE_CORRUPT',
  DIR_NOT_FOUND: 'DIR_NOT_FOUND',
  MANIFEST_NOT_FOUND: 'MANIFEST_NOT_FOUND',
  MANIFEST_CORRUPT: 'MANIFEST_CORRUPT',
  MANIFEST_VERSION: 'MANIFEST_VERSION',
  MIGRATION_FAILED: 'MIGRATION_FAILED',
  CONFLICT_UNRESOLVED: 'CONFLICT_UNRESOLVED',
  PARTIAL_WRITE: 'PARTIAL_WRITE',
  HASH_MISMATCH: 'HASH_MISMATCH',
  USER_CANCELLED: 'USER_CANCELLED',
  INVALID_INPUT: 'INVALID_INPUT',
  MISSING_DEPENDENCY: 'MISSING_DEPENDENCY',
  UNKNOWN: 'UNKNOWN',
} as const

export type ErrorCode = (typeof ErrorCodeEnum)[keyof typeof ErrorCodeEnum]

/**
 * Structured error class for all ai-setup errors
 * Carries error code, context, and optional cause for debugging
 */
export class AiSetupError extends Error {
  readonly code: ErrorCode
  readonly context: Record<string, unknown>
  override readonly cause: Error | undefined

  constructor(
    message: string,
    code: ErrorCode = ErrorCodeEnum.UNKNOWN,
    context?: Record<string, unknown>,
    cause?: Error,
  ) {
    super(message)
    this.name = 'AiSetupError'
    this.code = code
    this.context = context || {}
    this.cause = cause

    // Maintain proper prototype chain for instanceof checks
    Object.setPrototypeOf(this, AiSetupError.prototype)
  }

  /**
   * Determine if this error should be shown to user vs full stack trace
   * User errors: clear actionable messages, no stack
   * System errors: full stack trace when DEBUG enabled
   */
  get isUserError(): boolean {
    const userErrorCodes: ErrorCode[] = [
      ErrorCodeEnum.USER_CANCELLED,
      ErrorCodeEnum.INVALID_INPUT,
      ErrorCodeEnum.CONFLICT_UNRESOLVED,
    ]
    return userErrorCodes.includes(this.code)
  }

  /**
   * Suggested process exit code for this error
   * 0 = success / user cancellation
   * 1 = error
   */
  get exitCode(): number {
    return this.code === ErrorCodeEnum.USER_CANCELLED ? 0 : 1
  }
}

/**
 * Convenience factory for creating typed errors
 * Usage: throw Errors.fileNotFound(path)
 */
export const Errors = {
  fileNotFound(path: string): AiSetupError {
    return new AiSetupError(
      `File not found: ${path}`,
      ErrorCodeEnum.FILE_NOT_FOUND,
      { path },
    )
  },

  filePermission(path: string, operation: string): AiSetupError {
    return new AiSetupError(
      `Permission denied reading ${path} (${operation})`,
      ErrorCodeEnum.FILE_PERMISSION,
      { path, operation },
    )
  },

  fileCorrupt(path: string, cause?: Error): AiSetupError {
    return new AiSetupError(
      `File is corrupt or unreadable: ${path}`,
      ErrorCodeEnum.FILE_CORRUPT,
      { path },
      cause,
    )
  },

  dirNotFound(path: string): AiSetupError {
    return new AiSetupError(
      `Directory not found: ${path}`,
      ErrorCodeEnum.DIR_NOT_FOUND,
      { path },
    )
  },

  manifestNotFound(dir: string): AiSetupError {
    return new AiSetupError(
      `Setup manifest not found in ${dir}. Run 'ai-setup init' first.`,
      ErrorCodeEnum.MANIFEST_NOT_FOUND,
      { dir },
    )
  },

  manifestCorrupt(dir: string, cause?: Error): AiSetupError {
    return new AiSetupError(
      `Setup manifest is corrupt: ${dir}/.ai-setup.json`,
      ErrorCodeEnum.MANIFEST_CORRUPT,
      { dir },
      cause,
    )
  },

  manifestVersion(version: string): AiSetupError {
    return new AiSetupError(
      `Unsupported manifest schema version: ${version}. Please update ai-setup.`,
      ErrorCodeEnum.MANIFEST_VERSION,
      { version },
    )
  },

  migrationFailed(from: string, to: string, cause?: Error): AiSetupError {
    return new AiSetupError(
      `Failed to migrate manifest from v${from} to v${to}`,
      ErrorCodeEnum.MIGRATION_FAILED,
      { from, to },
      cause,
    )
  },

  conflictUnresolved(path: string, strategy: string): AiSetupError {
    return new AiSetupError(
      `Could not resolve conflict for ${path} using strategy '${strategy}'`,
      ErrorCodeEnum.CONFLICT_UNRESOLVED,
      { path, strategy },
    )
  },

  partialWrite(succeeded: string[], failed: string[]): AiSetupError {
    return new AiSetupError(
      `Partial failure: wrote ${succeeded.length} files, failed on ${failed.length}`,
      ErrorCodeEnum.PARTIAL_WRITE,
      { succeeded, failed, count: failed.length },
    )
  },

  hashMismatch(path: string, expected: string, actual: string): AiSetupError {
    return new AiSetupError(
      `File modified after install: ${path}`,
      ErrorCodeEnum.HASH_MISMATCH,
      { path, expected, actual },
    )
  },

  userCancelled(): AiSetupError {
    return new AiSetupError(
      'Operation cancelled by user',
      ErrorCodeEnum.USER_CANCELLED,
    )
  },

  invalidInput(message: string, context?: Record<string, unknown>): AiSetupError {
    return new AiSetupError(
      `Invalid input: ${message}`,
      ErrorCodeEnum.INVALID_INPUT,
      context,
    )
  },

  missingDependency(pkg: string): AiSetupError {
    return new AiSetupError(
      `Required dependency not found: ${pkg}. Run 'npm install' to fix.`,
      ErrorCodeEnum.MISSING_DEPENDENCY,
      { package: pkg },
    )
  },

  unknown(message: string, cause?: Error): AiSetupError {
    return new AiSetupError(
      message,
      ErrorCodeEnum.UNKNOWN,
      {},
      cause,
    )
  },
}
