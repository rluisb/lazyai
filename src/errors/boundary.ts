/**
 * Single error handling boundary for the CLI
 * All errors flow through handleError() for consistent user output and exit codes
 */

import { cancel } from '@clack/prompts'
import { AiSetupError, ErrorCodeEnum } from './types.js'

function isDebugEnabled(): boolean {
  return process.env.AI_SETUP_DEBUG === '1' || process.argv.includes('--verbose')
}

/**
 * Central error handler - this is the ONLY place where process.exit() is called
 * All errors should throw and propagate to this boundary
 *
 * Exit codes:
 * - 0: user cancellation (expected exit)
 * - 1: errors (file not found, permissions, validation, etc)
 */
export function handleError(err: unknown): never {
  const debug = isDebugEnabled()

  // Handle @clack/prompts cancel symbol
  if (typeof err === 'symbol') {
    cancel('Operation cancelled')
    process.exit(0)
  }

  // Extract error details
  const isAiSetupError = err instanceof AiSetupError
  const message = err instanceof Error ? err.message : String(err)
  const errorCode = isAiSetupError ? err.code : ErrorCodeEnum.UNKNOWN

  // Show message to user
  if (errorCode === ErrorCodeEnum.USER_CANCELLED) {
    cancel(message)
    process.exit(0)
  }

  // User-facing errors: show message, no stack
  if (isAiSetupError && err.isUserError) {
    console.error(`\n❌  ${message}\n`)
    if (debug) {
      console.error('Debug context:', err.context)
    }
    process.exit(1)
  }

  // System errors: show message and context
  console.error(`\n❌  ${message}\n`)

  // Debug mode: show full context and cause chain
  if (debug) {
    if (isAiSetupError) {
      console.error('Error code:', errorCode)
      console.error('Context:', err.context)
      if (err.cause) {
        console.error('Caused by:', err.cause.message)
        console.error('Stack:', err.cause.stack)
      }
    }
    if (err instanceof Error) {
      console.error('Stack:', err.stack)
    }
  }

  process.exit(1)
}
