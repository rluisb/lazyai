package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestGetConfigValueRejectsRetiredWorkflowDirectory(t *testing.T) {
	cfg := &LazyAIConfig{}
	_, err := getConfigValue(cfg, "workflows.directory")
	if err == nil || err.Error() != "unknown config key: workflows.directory" {
		t.Fatalf("getConfigValue error = %v, want unknown workflows.directory", err)
	}
}

func TestSetConfigValueRejectsRetiredWorkflowDirectory(t *testing.T) {
	cfg := &LazyAIConfig{}
	err := setConfigValue(cfg, "workflows.directory", ".opencode/workflows")
	if err == nil || err.Error() != "unknown config key: workflows.directory" {
		t.Fatalf("setConfigValue error = %v, want unknown workflows.directory", err)
	}
}

func TestSaveConfigOmitsRetiredWorkflowsSection(t *testing.T) {
	withTempDir(t)
	if err := os.MkdirAll(".opencode", 0o755); err != nil {
		t.Fatalf("mkdir .opencode: %v", err)
	}
	cfg := &LazyAIConfig{
		Agent:         AgentConfig{DefaultAgent: "primary-agent", DefaultModel: "gpt-4"},
		Database:      DatabaseConfig{Path: ".specify/lazyai.db"},
		Ledger:        LedgerConfig{Path: ".specify/ledger.jsonl"},
		Notifications: NotificationConfig{Enabled: false},
	}
	if err := saveConfig(cfg); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(".opencode", "config.yaml"))
	if err != nil {
		t.Fatalf("read config.yaml: %v", err)
	}
	contents := string(data)
	if strings.Contains(contents, "workflows:") || strings.Contains(contents, "directory:") {
		t.Fatalf("config.yaml should not contain retired workflows section:\n%s", contents)
	}
	if !strings.Contains(contents, "database:") || !strings.Contains(contents, "ledger:") {
		t.Fatalf("config.yaml lost active sections:\n%s", contents)
	}
}

func TestConfigListOmitsRetiredWorkflowDirectory(t *testing.T) {
	withTempDir(t)
	if err := os.MkdirAll(".opencode", 0o755); err != nil {
		t.Fatalf("mkdir .opencode: %v", err)
	}
	cfg := &LazyAIConfig{
		Agent:         AgentConfig{DefaultAgent: "primary-agent", DefaultModel: "gpt-4"},
		Database:      DatabaseConfig{Path: ".specify/lazyai.db"},
		Ledger:        LedgerConfig{Path: ".specify/ledger.jsonl"},
		Notifications: NotificationConfig{Enabled: false},
	}
	if err := saveConfig(cfg); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}
	out := captureStdout(t, func() {
		if err := configListCmd.RunE(&cobra.Command{}, []string{}); err != nil {
			t.Fatalf("config list: %v", err)
		}
	})
	if strings.Contains(out, "workflows.directory") {
		t.Fatalf("config list should not mention retired workflow config:\n%s", out)
	}
}
