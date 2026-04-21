// Package error provides structured error handling for the ai-setup CLI.
// Ported from src/errors/types.ts.
package error

import "fmt"

// ErrorCode represents a distinct failure scenario in the system.
type ErrorCode string

const (
	ErrFileNotFound       ErrorCode = "FILE_NOT_FOUND"
	ErrFilePermission     ErrorCode = "FILE_PERMISSION"
	ErrFileCorrupt        ErrorCode = "FILE_CORRUPT"
	ErrDirNotFound        ErrorCode = "DIR_NOT_FOUND"
	ErrManifestNotFound   ErrorCode = "MANIFEST_NOT_FOUND"
	ErrManifestCorrupt    ErrorCode = "MANIFEST_CORRUPT"
	ErrManifestVersion    ErrorCode = "MANIFEST_VERSION"
	ErrMigrationFailed    ErrorCode = "MIGRATION_FAILED"
	ErrConflictUnresolved ErrorCode = "CONFLICT_UNRESOLVED"
	ErrPartialWrite       ErrorCode = "PARTIAL_WRITE"
	ErrHashMismatch       ErrorCode = "HASH_MISMATCH"
	ErrUserCancelled      ErrorCode = "USER_CANCELLED"
	ErrInvalidInput       ErrorCode = "INVALID_INPUT"
	ErrMissingDependency  ErrorCode = "MISSING_DEPENDENCY"
	ErrUnknown            ErrorCode = "UNKNOWN"
)

// AiSetupError is the structured error type for all ai-setup errors.
// It carries an error code, context map, and optional cause for debugging.
type AiSetupError struct {
	Message string
	Code    ErrorCode
	Context map[string]any
	Cause   error
}

// Error implements the error interface.
func (e *AiSetupError) Error() string { return e.Message }

// Unwrap returns the underlying cause, enabling errors.Is/As traversal.
func (e *AiSetupError) Unwrap() error { return e.Cause }

// IsUserError returns true for user-facing errors that don't need stack traces.
// User errors: clear actionable messages, no stack.
// System errors: full stack trace when DEBUG enabled.
func (e *AiSetupError) IsUserError() bool {
	switch e.Code {
	case ErrUserCancelled, ErrInvalidInput, ErrConflictUnresolved:
		return true
	default:
		return false
	}
}

// ExitCode returns the suggested process exit code for this error.
// 0 = user cancellation (expected exit), 1 = error.
func (e *AiSetupError) ExitCode() int {
	if e.Code == ErrUserCancelled {
		return 0
	}
	return 1
}

// ---------------------------------------------------------------------------
// Factory functions — matching the TypeScript Errors object exactly
// ---------------------------------------------------------------------------

// FileNotFound creates a FILE_NOT_FOUND error.
func FileNotFound(path string) *AiSetupError {
	return &AiSetupError{
		Message: fmt.Sprintf("File not found: %s", path),
		Code:    ErrFileNotFound,
		Context: map[string]any{"path": path},
	}
}

// FilePermission creates a FILE_PERMISSION error.
func FilePermission(path, operation string) *AiSetupError {
	return &AiSetupError{
		Message: fmt.Sprintf("Permission denied reading %s (%s)", path, operation),
		Code:    ErrFilePermission,
		Context: map[string]any{"path": path, "operation": operation},
	}
}

// FileCorrupt creates a FILE_CORRUPT error.
func FileCorrupt(path string, cause error) *AiSetupError {
	return &AiSetupError{
		Message: fmt.Sprintf("File is corrupt or unreadable: %s", path),
		Code:    ErrFileCorrupt,
		Context: map[string]any{"path": path},
		Cause:   cause,
	}
}

// DirNotFound creates a DIR_NOT_FOUND error.
func DirNotFound(path string) *AiSetupError {
	return &AiSetupError{
		Message: fmt.Sprintf("Directory not found: %s", path),
		Code:    ErrDirNotFound,
		Context: map[string]any{"path": path},
	}
}

// ManifestNotFound creates a MANIFEST_NOT_FOUND error.
func ManifestNotFound(dir string) *AiSetupError {
	return &AiSetupError{
		Message: fmt.Sprintf("Setup manifest not found in %s. Run 'ai-setup init' first.", dir),
		Code:    ErrManifestNotFound,
		Context: map[string]any{"dir": dir},
	}
}

// ManifestCorrupt creates a MANIFEST_CORRUPT error.
func ManifestCorrupt(dir string, cause error) *AiSetupError {
	return &AiSetupError{
		Message: fmt.Sprintf("Setup manifest is corrupt: %s/.ai-setup.json", dir),
		Code:    ErrManifestCorrupt,
		Context: map[string]any{"dir": dir},
		Cause:   cause,
	}
}

// ManifestVersion creates a MANIFEST_VERSION error.
func ManifestVersion(version string) *AiSetupError {
	return &AiSetupError{
		Message: fmt.Sprintf("Unsupported manifest schema version: %s. Please update ai-setup.", version),
		Code:    ErrManifestVersion,
		Context: map[string]any{"version": version},
	}
}

// MigrationFailed creates a MIGRATION_FAILED error.
func MigrationFailed(from, to string, cause error) *AiSetupError {
	return &AiSetupError{
		Message: fmt.Sprintf("Failed to migrate manifest from v%s to v%s", from, to),
		Code:    ErrMigrationFailed,
		Context: map[string]any{"from": from, "to": to},
		Cause:   cause,
	}
}

// ConflictUnresolved creates a CONFLICT_UNRESOLVED error.
func ConflictUnresolved(path, strategy string) *AiSetupError {
	return &AiSetupError{
		Message: fmt.Sprintf("Could not resolve conflict for %s using strategy '%s'", path, strategy),
		Code:    ErrConflictUnresolved,
		Context: map[string]any{"path": path, "strategy": strategy},
	}
}

// PartialWrite creates a PARTIAL_WRITE error.
func PartialWrite(succeeded, failed []string) *AiSetupError {
	return &AiSetupError{
		Message: fmt.Sprintf("Partial failure: wrote %d files, failed on %d", len(succeeded), len(failed)),
		Code:    ErrPartialWrite,
		Context: map[string]any{"succeeded": succeeded, "failed": failed, "count": len(failed)},
	}
}

// HashMismatch creates a HASH_MISMATCH error.
func HashMismatch(path, expected, actual string) *AiSetupError {
	return &AiSetupError{
		Message: fmt.Sprintf("File modified after install: %s", path),
		Code:    ErrHashMismatch,
		Context: map[string]any{"path": path, "expected": expected, "actual": actual},
	}
}

// UserCancelled creates a USER_CANCELLED error.
func UserCancelled() *AiSetupError {
	return &AiSetupError{
		Message: "Operation cancelled by user",
		Code:    ErrUserCancelled,
		Context: map[string]any{},
	}
}

// InvalidInput creates an INVALID_INPUT error.
func InvalidInput(message string, context map[string]any) *AiSetupError {
	return &AiSetupError{
		Message: fmt.Sprintf("Invalid input: %s", message),
		Code:    ErrInvalidInput,
		Context: context,
	}
}

// MissingDependency creates a MISSING_DEPENDENCY error.
func MissingDependency(pkg string) *AiSetupError {
	return &AiSetupError{
		Message: fmt.Sprintf("Required dependency not found: %s. Run 'npm install' to fix.", pkg),
		Code:    ErrMissingDependency,
		Context: map[string]any{"package": pkg},
	}
}

// Unknown creates an UNKNOWN error with the given message and optional cause.
func Unknown(message string, cause error) *AiSetupError {
	return &AiSetupError{
		Message: message,
		Code:    ErrUnknown,
		Context: map[string]any{},
		Cause:   cause,
	}
}
