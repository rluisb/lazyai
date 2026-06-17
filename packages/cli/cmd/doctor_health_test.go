package cmd

import (
	"testing"
)

func TestGetDiskUsage(t *testing.T) {
	// Test with current directory
	usage := getDiskUsage(".")

	// Should return a value between 0 and 100, or -1 on error
	if usage >= 0 {
		if usage < 0 || usage > 100 {
			t.Errorf("Disk usage should be between 0 and 100, got %.0f", usage)
		}
		t.Logf("Disk usage: %.0f%%", usage)
	} else {
		t.Log("Could not determine disk usage (expected on some systems)")
	}
}

func TestRunEnhancedHealthChecks(t *testing.T) {
	checks := runEnhancedHealthChecks()

	// Should return at least 6 checks.
	if len(checks) < 6 {
		t.Errorf("Expected at least 6 health checks, got %d", len(checks))
	}

	// Verify all checks have required fields
	for i, check := range checks {
		if check.Name == "" {
			t.Errorf("Check %d has no name", i)
		}
		if check.Status == "" {
			t.Errorf("Check %d has no status", i)
		}
		if check.Detail == "" {
			t.Errorf("Check %d has no detail", i)
		}
		if check.Checked == "" {
			t.Errorf("Check %d has no checked timestamp", i)
		}

		// Status should be one of: pass, warn, fail
		if check.Status != "pass" && check.Status != "warn" && check.Status != "fail" {
			t.Errorf("Check %d has invalid status: %s", i, check.Status)
		}
	}
	for _, check := range checks {
		if check.Name == "legacy runtime binary" {
			t.Fatal("legacy runtime binary check should be removed")
		}
	}

	// Count results
	pass, warn, fail := 0, 0, 0
	for _, check := range checks {
		switch check.Status {
		case "pass":
			pass++
		case "warn":
			warn++
		case "fail":
			fail++
		}
	}

	t.Logf("Health checks: %d pass, %d warn, %d fail (of %d total)", pass, warn, fail, len(checks))

	// At least some checks should pass (sqlite3 and git should be available)
	if pass == 0 {
		t.Error("No health checks passed")
	}
}

func TestHealthCheckStruct(t *testing.T) {
	check := HealthCheck{
		Name:    "Test Check",
		Status:  "pass",
		Detail:  "Test detail",
		Checked: "2026-05-24T00:00:00Z",
	}

	if check.Name != "Test Check" {
		t.Errorf("Expected name 'Test Check', got '%s'", check.Name)
	}

	if check.Status != "pass" {
		t.Errorf("Expected status 'pass', got '%s'", check.Status)
	}

	if check.Detail != "Test detail" {
		t.Errorf("Expected detail 'Test detail', got '%s'", check.Detail)
	}
}
