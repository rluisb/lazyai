package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// HealthCheck represents a single health check result
type HealthCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // pass, warn, fail
	Detail  string `json:"detail"`
	Checked string `json:"checked_at"`
}

// runEnhancedHealthChecks runs all production health checks
func runEnhancedHealthChecks() []HealthCheck {
	var checks []HealthCheck
	checkedAt := time.Now().UTC().Format(time.RFC3339)

	// 1. Check sqlite3
	check := HealthCheck{Name: "Dependency: sqlite3", Checked: checkedAt}
	if out, err := exec.Command("sqlite3", "--version").Output(); err == nil {
		check.Status = "pass"
		check.Detail = strings.TrimSpace(string(out))
	} else {
		check.Status = "fail"
		check.Detail = "sqlite3 not found on PATH"
	}
	checks = append(checks, check)

	// 2. Check git
	check = HealthCheck{Name: "Dependency: git", Checked: checkedAt}
	if out, err := exec.Command("git", "--version").Output(); err == nil {
		check.Status = "pass"
		check.Detail = strings.TrimSpace(string(out))
	} else {
		check.Status = "fail"
		check.Detail = "git not found on PATH"
	}
	checks = append(checks, check)

	// 3. Check jq
	check = HealthCheck{Name: "Dependency: jq", Checked: checkedAt}
	if out, err := exec.Command("jq", "--version").Output(); err == nil {
		check.Status = "pass"
		check.Detail = strings.TrimSpace(string(out))
	} else {
		check.Status = "warn"
		check.Detail = "jq not found (optional but recommended)"
	}
	checks = append(checks, check)

	// 4. Check bash version
	check = HealthCheck{Name: "Dependency: bash", Checked: checkedAt}
	if out, err := exec.Command("bash", "--version").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 0 {
			check.Status = "pass"
			check.Detail = strings.TrimSpace(lines[0])
		} else {
			check.Status = "pass"
			check.Detail = "bash available"
		}
	} else {
		check.Status = "fail"
		check.Detail = "bash not found"
	}
	checks = append(checks, check)

	// 5. Check Ollama
	check = HealthCheck{Name: "Provider: ollama", Checked: checkedAt}
	if out, err := exec.Command("curl", "-s", "--max-time", "5", "http://localhost:11434/api/tags").CombinedOutput(); err == nil && len(out) > 0 {
		check.Status = "pass"
		check.Detail = "Responding"
	} else {
		check.Status = "warn"
		check.Detail = "Ollama not running on localhost:11434"
	}
	checks = append(checks, check)

	// 6. Check OpenAI API key
	check = HealthCheck{Name: "Provider: openai", Checked: checkedAt}
	if os.Getenv("OPENAI_API_KEY") != "" {
		check.Status = "pass"
		check.Detail = "API key configured"
	} else {
		check.Status = "warn"
		check.Detail = "OPENAI_API_KEY not set"
	}
	checks = append(checks, check)

	// 7. Check disk space
	check = HealthCheck{Name: "Disk space", Checked: checkedAt}
	if usage := getDiskUsage("."); usage >= 0 {
		check.Detail = fmt.Sprintf("%.0f%% used", usage)
		if usage > 90 {
			check.Status = "fail"
		} else if usage > 80 {
			check.Status = "warn"
		} else {
			check.Status = "pass"
		}
	} else {
		check.Status = "warn"
		check.Detail = "Could not determine disk usage"
	}
	checks = append(checks, check)

	return checks
}

// getDiskUsage returns disk usage percentage for the given path
func getDiskUsage(path string) float64 {
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		cmd = exec.Command("df", "-h", path)
	} else {
		return -1
	}

	out, err := cmd.Output()
	if err != nil {
		return -1
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return -1
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 5 {
		return -1
	}

	usageStr := strings.TrimSuffix(fields[4], "%")
	usage, err := strconv.ParseFloat(usageStr, 64)
	if err != nil {
		return -1
	}

	return usage
}
