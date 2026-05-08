// Error boundary — single exit point for all CLI errors.
// Ported from src/errors/boundary.ts.
package error

import (
	"fmt"
	"os"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/theme"
)

// IsDebugEnabled returns true if debug/verbose mode is active.
// Checks the AI_SETUP_DEBUG environment variable and the --verbose CLI flag.
func IsDebugEnabled() bool {
	if os.Getenv("AI_SETUP_DEBUG") == "1" {
		return true
	}
	for _, arg := range os.Args {
		if arg == "--verbose" {
			return true
		}
	}
	return false
}

// HandleError is the single exit point for all errors.
// It prints styled error messages and calls os.Exit with the appropriate code.
//
// Exit codes:
//   - 0: user cancellation (expected exit)
//   - 1: errors (file not found, permissions, validation, etc.)
func HandleError(err error) {
	debug := IsDebugEnabled()

	// Try to extract an AiSetupError
	var aiErr *AiSetupError
	isAiSetupError := AsAiSetupError(err, &aiErr)

	// 1. USER_CANCELLED → print cancel message, exit 0
	if isAiSetupError && aiErr.Code == ErrUserCancelled {
		fmt.Fprintf(os.Stderr, "\n  %s\n\n", aiErr.Message)
		os.Exit(0)
	}

	// Extract message
	message := err.Error()

	// 2. User-facing errors → show message (no stack), exit 1.
	// Renders `✗ <msg>` in the design-system error color via theme.Errorf
	// (FR-007). Pre-refactor used the ❌ emoji which violated the no-emoji
	// rule of the lazyai-design-system skill.
	if isAiSetupError && aiErr.IsUserError() {
		fmt.Fprintln(os.Stderr)
		theme.Errorf(os.Stderr, "%s", message)
		fmt.Fprintln(os.Stderr)
		if debug {
			fmt.Fprintf(os.Stderr, "Debug context: %v\n", aiErr.Context)
		}
		os.Exit(1)
	}

	// 3. AiSetupError system errors → show message, debug adds code + context + cause.
	fmt.Fprintln(os.Stderr)
	theme.Errorf(os.Stderr, "%s", message)
	fmt.Fprintln(os.Stderr)
	if debug {
		if isAiSetupError {
			fmt.Fprintf(os.Stderr, "Error code: %s\n", aiErr.Code)
			fmt.Fprintf(os.Stderr, "Context: %v\n", aiErr.Context)
			if aiErr.Cause != nil {
				fmt.Fprintf(os.Stderr, "Caused by: %s\n", aiErr.Cause.Error())
				// Show cause stack if it has one (via DebugStack)
				if stack := debugStack(aiErr.Cause); stack != "" {
					fmt.Fprintf(os.Stderr, "Stack: %s\n", stack)
				}
			}
		}
		// 4. Generic error → show stack if available
		if stack := debugStack(err); stack != "" {
			fmt.Fprintf(os.Stderr, "Stack: %s\n", stack)
		}
	}

	os.Exit(1)
}

// IsUserCancelled checks if the error is a USER_CANCELLED AiSetupError.
func IsUserCancelled(err error) bool {
	var aiErr *AiSetupError
	if AsAiSetupError(err, &aiErr) {
		return aiErr.Code == ErrUserCancelled
	}
	return false
}

// AsAiSetupError attempts to extract an *AiSetupError from the error chain.
// Returns true if found, and sets target to the found value.
func AsAiSetupError(err error, target **AiSetupError) bool {
	if err == nil {
		return false
	}
	// Direct type assertion
	if ai, ok := err.(*AiSetupError); ok {
		*target = ai
		return true
	}
	// Try wrapping interface
	type wrapper interface{ Unwrap() error }
	if w, ok := err.(wrapper); ok {
		return AsAiSetupError(w.Unwrap(), target)
	}
	return false
}

// debugStack returns a stack trace string if the error supports it,
// otherwise returns an empty string. Go doesn't have built-in stack traces
// on errors, so we check for a Stack() method.
func debugStack(err error) string {
	type stacker interface{ Stack() string }
	if s, ok := err.(stacker); ok {
		return strings.TrimSpace(s.Stack())
	}
	return ""
}
