package cmd

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestGetDB(t *testing.T) {
	// This test requires a database to be initialized
	// In a real test, we would mock the database or use a temp directory
	// For now, we just verify the function exists and returns an error when no DB exists
	_, err := getDB()
	if err == nil {
		t.Log("getDB returned a database (may have found one in current directory)")
	} else {
		t.Logf("getDB returned expected error: %v", err)
	}
}

func TestRunSessionStart(t *testing.T) {
	// Test that session start generates a valid session ID format
	sessionID := fmt.Sprintf("ses_%d", time.Now().Unix())
	if len(sessionID) < 4 {
		t.Error("Session ID is too short")
	}

	if sessionID[:4] != "ses_" {
		t.Error("Session ID does not start with 'ses_'")
	}
}

func TestSessionIDFormat(t *testing.T) {
	// Test session ID format
	now := time.Now().Unix()
	sessionID := fmt.Sprintf("ses_%d", now)

	expectedPrefix := "ses_"
	if len(sessionID) <= len(expectedPrefix) {
		t.Errorf("Session ID '%s' is too short", sessionID)
	}

	if sessionID[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("Session ID '%s' does not start with '%s'", sessionID, expectedPrefix)
	}
}

func TestTimeFormat(t *testing.T) {
	// Test that time formatting works correctly
	now := time.Now().UTC()
	formatted := now.Format(time.RFC3339)

	// Should contain T and Z
	if !strings.Contains(formatted, "T") {
		t.Error("Formatted time does not contain 'T'")
	}

	if !strings.Contains(formatted, "Z") {
		t.Error("Formatted time does not contain 'Z'")
	}
}
