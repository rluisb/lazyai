package log

import (
	"testing"

	charmlog "charm.land/log/v2"
)

func TestNew_LevelParsing(t *testing.T) {
	tests := []struct {
		level string
		want  charmlog.Level
	}{
		{"debug", charmlog.DebugLevel},
		{"info", charmlog.InfoLevel},
		{"warn", charmlog.WarnLevel},
		{"error", charmlog.ErrorLevel},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			logger := New(tt.level, "text")
			assertLogger(t, logger)
			if got := logger.GetLevel(); got != tt.want {
				t.Fatalf("expected level %v, got %v", tt.want, got)
			}
		})
	}
}

func TestNew_FormatParsing(t *testing.T) {
	for _, format := range []string{"text", "json", "logfmt"} {
		t.Run(format, func(t *testing.T) {
			assertLogger(t, New("info", format))
		})
	}
}

func TestDefault(t *testing.T) {
	resetDefaultLogger(t)
	t.Setenv("AI_SETUP_LOG_LEVEL", "")
	t.Setenv("AI_SETUP_LOG_FORMAT", "")
	Configure("", "")
	logger := Default()
	assertLogger(t, logger)
	if got := logger.GetLevel(); got != charmlog.InfoLevel {
		t.Fatalf("expected default level %v, got %v", charmlog.InfoLevel, got)
	}
}

func TestDefaultReturnsCachedLogger(t *testing.T) {
	resetDefaultLogger(t)

	first := Default()
	second := Default()

	if first != second {
		t.Fatal("expected Default to return cached logger")
	}
}

func TestConfigureReplacesDefaultLogger(t *testing.T) {
	resetDefaultLogger(t)

	before := Default()
	Configure("debug", "json")
	after := Default()

	if before == after {
		t.Fatal("expected Configure to replace cached logger")
	}
	if got := after.GetLevel(); got != charmlog.DebugLevel {
		t.Fatalf("expected configured level %v, got %v", charmlog.DebugLevel, got)
	}
}

func TestNew_ExplicitArgsOverrideEnv(t *testing.T) {
	t.Setenv("AI_SETUP_LOG_LEVEL", "debug")
	t.Setenv("AI_SETUP_LOG_FORMAT", "logfmt")
	logger := New("warn", "json")
	assertLogger(t, logger)
	if got := logger.GetLevel(); got != charmlog.WarnLevel {
		t.Fatalf("expected explicit level %v, got %v", charmlog.WarnLevel, got)
	}
}

func TestNew_InvalidLevel(t *testing.T) {
	logger := New("invalid", "text")
	assertLogger(t, logger)
	if got := logger.GetLevel(); got != charmlog.InfoLevel {
		t.Fatalf("expected invalid level to default to %v, got %v", charmlog.InfoLevel, got)
	}
}

func TestNew_InvalidFormat(t *testing.T) {
	assertLogger(t, New("info", "invalid"))
}

func assertLogger(t *testing.T, logger *charmlog.Logger) {
	t.Helper()
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
	if got := logger.GetPrefix(); got != prefix {
		t.Fatalf("expected prefix %q, got %q", prefix, got)
	}
}

func resetDefaultLogger(t *testing.T) {
	t.Helper()
	Configure("", "")
	t.Cleanup(func() {
		Configure("", "")
	})
}
