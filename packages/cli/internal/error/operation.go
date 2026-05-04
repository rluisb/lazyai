// Operation tracking for install/update/remove operations.
// Ported from src/errors/operation.ts.
package error

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// failedEntry records a file path and its error message for a failed operation.
type failedEntry struct {
	Path  string
	Error string
}

// OperationTracker tracks succeeded/failed files and backup paths for an operation.
type OperationTracker struct {
	succeeded     []string
	failed        []failedEntry
	backups       map[string]string
	operationType string
	timestamp     string
}

// NewOperationTracker creates a new tracker for the given operation type.
func NewOperationTracker(operationType string) *OperationTracker {
	return &OperationTracker{
		succeeded:     []string{},
		failed:        []failedEntry{},
		backups:       make(map[string]string),
		operationType: operationType,
		timestamp:     time.Now().UTC().Format(time.RFC3339),
	}
}

// TrackSuccess records a successfully processed file.
func (t *OperationTracker) TrackSuccess(filePath string) {
	t.succeeded = append(t.succeeded, filePath)
}

// TrackFailure records a failed file along with its error message.
func (t *OperationTracker) TrackFailure(filePath string, errMsg string) {
	t.failed = append(t.failed, failedEntry{Path: filePath, Error: errMsg})
}

// RegisterBackup records a backup mapping from source path to backup path.
func (t *OperationTracker) RegisterBackup(sourcePath, backupPath string) {
	t.backups[sourcePath] = backupPath
}

// SucceededCount returns the number of successfully processed files.
func (t *OperationTracker) SucceededCount() int {
	return len(t.succeeded)
}

// FailedCount returns the number of failed files.
func (t *OperationTracker) FailedCount() int {
	return len(t.failed)
}

// Result returns the overall operation result based on success/failure counts.
func (t *OperationTracker) Result() types.OperationResult {
	if len(t.failed) > 0 && len(t.succeeded) > 0 {
		return types.OperationResultPartial
	}
	if len(t.failed) > 0 {
		return types.OperationResultFailure
	}
	return types.OperationResultSuccess
}

// ToOperation converts the tracker state into a types.Operation record.
// Generates a unique ID in the format: op_{timestamp}_{random 6 chars}.
func (t *OperationTracker) ToOperation() types.Operation {
	id := fmt.Sprintf("op_%d_%s", time.Now().UnixMilli(), randomAlphaNum(6))

	filesAffected := make([]string, 0, len(t.succeeded)+len(t.failed))
	filesAffected = append(filesAffected, t.succeeded...)
	for _, entry := range t.failed {
		filesAffected = append(filesAffected, entry.Path)
	}

	backupPaths := make([]string, 0, len(t.backups))
	for _, bp := range t.backups {
		backupPaths = append(backupPaths, bp)
	}

	op := types.Operation{
		ID:            id,
		Type:          t.operationType,
		Timestamp:     t.timestamp,
		FilesAffected: filesAffected,
		Result:        t.Result(),
		BackupPaths:   backupPaths,
	}

	if len(t.failed) > 0 {
		errParts := make([]string, 0, len(t.failed))
		for _, entry := range t.failed {
			errParts = append(errParts, fmt.Sprintf("%s: %s", entry.Path, entry.Error))
		}
		op.Error = fmt.Sprintf("%v", errParts)
		// Join error parts with semicolons to match TS behavior
		joined := ""
		for i, part := range errParts {
			if i > 0 {
				joined += "; "
			}
			joined += part
		}
		op.Error = joined
	}

	return op
}

// randomAlphaNum generates a random alphanumeric string of the given length.
func randomAlphaNum(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
