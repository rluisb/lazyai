package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/db"
)

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for '%s': %s", e.Field, e.Message)
}

// ValidateSessionID checks if a session ID is valid
func ValidateSessionID(id string) error {
	if id == "" {
		return ValidationError{Field: "session_id", Message: "cannot be empty"}
	}
	if !strings.HasPrefix(id, "ses_") {
		return ValidationError{Field: "session_id", Message: "must start with 'ses_'"}
	}
	return nil
}

// ValidateTaskID checks if a task ID is valid
func ValidateTaskID(id string) error {
	if id == "" {
		return ValidationError{Field: "task_id", Message: "cannot be empty"}
	}
	if !strings.HasPrefix(id, "task_") {
		return ValidationError{Field: "task_id", Message: "must start with 'task_'"}
	}
	return nil
}

// ValidateNotEmpty checks if a string is not empty
func ValidateNotEmpty(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return ValidationError{Field: fieldName, Message: "cannot be empty or whitespace only"}
	}
	return nil
}

// ValidateAgentName checks if an agent name is valid
func ValidateAgentName(name string) error {
	if err := ValidateNotEmpty(name, "agent"); err != nil {
		return err
	}
	validAgents := []string{"guide", "implementer", "researcher", "deployer", "responder", "planner", "reviewer", "evidence-verifier"}
	for _, agent := range validAgents {
		if agent == name {
			return nil
		}
	}
	return ValidationError{Field: "agent", Message: fmt.Sprintf("'%s' is not a recognized agent name", name)}
}

// ValidatePriority checks if a priority level is valid
func ValidatePriority(priority string) error {
	validPriorities := []string{"low", "normal", "high", "critical"}
	for _, p := range validPriorities {
		if p == priority {
			return nil
		}
	}
	return ValidationError{Field: "priority", Message: fmt.Sprintf("'%s' is not valid. Use: low, normal, high, critical", priority)}
}

// RetryWithBackoff retries a database operation with exponential backoff
func RetryWithBackoff(operation func() error, maxRetries int) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := operation(); err != nil {
			lastErr = err
			if strings.Contains(err.Error(), "database is locked") {
				// Exponential backoff: 100ms, 200ms, 400ms
				time.Sleep(time.Duration(100*(1<<i)) * time.Millisecond)
				continue
			}
			// Non-retryable error
			return err
		}
		return nil
	}
	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}

// EnsureDB checks if database exists and initializes if needed
func EnsureDB() (*db.DB, error) {
	database, err := getDB()
	if err != nil {
		if strings.Contains(err.Error(), "database not found") {
			return nil, fmt.Errorf("database not initialized. Run 'lazyai-cli init' first")
		}
		return nil, err
	}

	// Run migrations to ensure schema is up to date
	if err := db.RunMigrations(database); err != nil {
		database.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return database, nil
}

// SafeCloseDB safely closes a database connection
func SafeCloseDB(database *db.DB) {
	if database != nil {
		database.Close()
	}
}

// GetEnvOrDefault gets an environment variable or returns a default
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}
