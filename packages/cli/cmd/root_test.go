package cmd

import (
	"os"
	"testing"

	charmlog "charm.land/log/v2"
	clilog "github.com/rluisb/lazyai/packages/cli/internal/log"
	buildversion "github.com/rluisb/lazyai/packages/cli/internal/version"
)

func TestSyncBuildVersionPropagatesCmdVersion(t *testing.T) {
	originalCmdVersion := Version
	originalBuildVersion := buildversion.Version
	t.Cleanup(func() {
		Version = originalCmdVersion
		buildversion.Version = originalBuildVersion
	})

	Version = "1.2.3"
	buildversion.Version = buildversion.DevVersion

	syncBuildVersion()

	if buildversion.Version != "1.2.3" {
		t.Fatalf("buildversion.Version = %q, want %q", buildversion.Version, "1.2.3")
	}
}

func TestSyncBuildVersionUsesInternalVersionWhenCmdVersionIsDev(t *testing.T) {
	originalCmdVersion := Version
	originalBuildVersion := buildversion.Version
	t.Cleanup(func() {
		Version = originalCmdVersion
		buildversion.Version = originalBuildVersion
	})

	Version = buildversion.DevVersion
	buildversion.Version = "2.3.4"

	syncBuildVersion()

	if Version != "2.3.4" {
		t.Fatalf("Version = %q, want %q", Version, "2.3.4")
	}
}

func TestApplyLoggingEnvMapsVerboseToDebugWhenLogLevelUnset(t *testing.T) {
	resetCLILog(t)
	t.Setenv("AI_SETUP_LOG_LEVEL", "")

	if err := applyLoggingEnv(loggingFlagConfig{Verbose: true}); err != nil {
		t.Fatalf("applyLoggingEnv returned error: %v", err)
	}

	if got := os.Getenv("AI_SETUP_LOG_LEVEL"); got != "debug" {
		t.Fatalf("AI_SETUP_LOG_LEVEL = %q, want %q", got, "debug")
	}
	if got := clilog.Default().GetLevel(); got != charmlog.DebugLevel {
		t.Fatalf("cached log level = %v, want %v", got, charmlog.DebugLevel)
	}
}

func TestApplyLoggingEnvExplicitLogLevelOverridesVerbose(t *testing.T) {
	resetCLILog(t)
	t.Setenv("AI_SETUP_LOG_LEVEL", "")

	if err := applyLoggingEnv(loggingFlagConfig{Verbose: true, LogLevel: "warn", LogLevelExplicit: true}); err != nil {
		t.Fatalf("applyLoggingEnv returned error: %v", err)
	}

	if got := os.Getenv("AI_SETUP_LOG_LEVEL"); got != "warn" {
		t.Fatalf("AI_SETUP_LOG_LEVEL = %q, want %q", got, "warn")
	}
	if got := clilog.Default().GetLevel(); got != charmlog.WarnLevel {
		t.Fatalf("cached log level = %v, want %v", got, charmlog.WarnLevel)
	}
}

func TestApplyLoggingEnvSetsLogFormat(t *testing.T) {
	resetCLILog(t)
	t.Setenv("AI_SETUP_LOG_FORMAT", "")

	if err := applyLoggingEnv(loggingFlagConfig{LogFormat: "json", LogFormatExplicit: true}); err != nil {
		t.Fatalf("applyLoggingEnv returned error: %v", err)
	}

	if got := os.Getenv("AI_SETUP_LOG_FORMAT"); got != "json" {
		t.Fatalf("AI_SETUP_LOG_FORMAT = %q, want %q", got, "json")
	}
}

func resetCLILog(t *testing.T) {
	t.Helper()
	clilog.Configure("", "")
	t.Cleanup(func() {
		clilog.Configure("", "")
	})
}
