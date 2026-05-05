package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/adapter"
)

func TestPrintMcpNextSteps(t *testing.T) {
	// Redirect stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create test data
	servers := map[string]adapter.McpServer{
		"server1": {
			Env:       map[string]string{"API_KEY": "test"},
			SetupHint: "Hint for server1",
		},
		"server2": {
			SetupHint: "Hint for server2",
		},
	}

	// Run function
	PrintMcpNextSteps(servers)

	// Restore stdout and get output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output
	if !strings.Contains(output, "Next Steps: MCP Server Configuration") {
		t.Errorf("Expected 'Next Steps: MCP Server Configuration', got %s", output)
	}
	if !strings.Contains(output, "Fill in any required environment variables") {
		t.Errorf("Expected 'Fill in any required environment variables', got %s", output)
	}
	if !strings.Contains(output, "Hint for server1") {
		t.Errorf("Expected 'Hint for server1', got %s", output)
	}
	if !strings.Contains(output, "Hint for server2") {
		t.Errorf("Expected 'Hint for server2', got %s", output)
	}
}

func TestPrintMcpNextSteps_NoHints(t *testing.T) {
	// Redirect stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create test data
	servers := map[string]adapter.McpServer{
		"server1": {
			Env: map[string]string{},
		},
	}

	// Run function
	PrintMcpNextSteps(servers)

	// Restore stdout and get output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output is empty
	if strings.TrimSpace(output) != "" {
		t.Errorf("Expected no output, got %s", output)
	}
}