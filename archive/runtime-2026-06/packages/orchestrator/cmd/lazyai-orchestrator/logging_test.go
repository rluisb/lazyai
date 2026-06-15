package main

import (
	"os"
	"testing"

	charmlog "charm.land/log/v2"
	orchlog "github.com/rluisb/lazyai/packages/orchestrator/internal/log"
)

func TestApplyLoggingEnvSetsLogLevel(t *testing.T) {
	resetOrchestratorLog(t)
	t.Setenv("AI_SETUP_LOG_LEVEL", "")

	if err := applyLoggingEnv(loggingFlagConfig{LogLevel: "error", LogLevelExplicit: true}); err != nil {
		t.Fatalf("applyLoggingEnv returned error: %v", err)
	}

	if got := os.Getenv("AI_SETUP_LOG_LEVEL"); got != "error" {
		t.Fatalf("AI_SETUP_LOG_LEVEL = %q, want %q", got, "error")
	}
	if got := orchlog.Default().GetLevel(); got != charmlog.ErrorLevel {
		t.Fatalf("cached log level = %v, want %v", got, charmlog.ErrorLevel)
	}
}

func TestApplyLoggingEnvSetsLogFormat(t *testing.T) {
	resetOrchestratorLog(t)
	t.Setenv("AI_SETUP_LOG_FORMAT", "")

	if err := applyLoggingEnv(loggingFlagConfig{LogFormat: "logfmt", LogFormatExplicit: true}); err != nil {
		t.Fatalf("applyLoggingEnv returned error: %v", err)
	}

	if got := os.Getenv("AI_SETUP_LOG_FORMAT"); got != "logfmt" {
		t.Fatalf("AI_SETUP_LOG_FORMAT = %q, want %q", got, "logfmt")
	}
}

func resetOrchestratorLog(t *testing.T) {
	t.Helper()
	orchlog.Configure("", "")
	t.Cleanup(func() {
		orchlog.Configure("", "")
	})
}
