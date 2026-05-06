package log

import (
	"os"
	"strings"
	"sync"

	charmlog "charm.land/log/v2"
)

const prefix = "cli"

var (
	defaultMu     sync.RWMutex
	defaultLogger *charmlog.Logger
	defaultOutput *os.File
)

func New(level, format string) *charmlog.Logger {
	if strings.TrimSpace(level) == "" {
		level = os.Getenv("AI_SETUP_LOG_LEVEL")
	}
	if strings.TrimSpace(format) == "" {
		format = os.Getenv("AI_SETUP_LOG_FORMAT")
	}

	return charmlog.NewWithOptions(os.Stderr, charmlog.Options{
		Level:           parseLevel(level),
		Formatter:       parseFormat(format),
		Prefix:          prefix,
		ReportTimestamp: false,
	})
}

func Default() *charmlog.Logger {
	defaultMu.RLock()
	logger := defaultLogger
	output := defaultOutput
	defaultMu.RUnlock()
	if logger != nil && output == os.Stderr {
		return logger
	}

	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultLogger == nil || defaultOutput != os.Stderr {
		defaultLogger = New("", "")
		defaultOutput = os.Stderr
	}
	return defaultLogger
}

func Configure(level, format string) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultLogger = New(level, format)
	defaultOutput = os.Stderr
}

func parseLevel(level string) charmlog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return charmlog.DebugLevel
	case "warn":
		return charmlog.WarnLevel
	case "error":
		return charmlog.ErrorLevel
	default:
		return charmlog.InfoLevel
	}
}

func parseFormat(format string) charmlog.Formatter {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return charmlog.JSONFormatter
	case "logfmt":
		return charmlog.LogfmtFormatter
	case "text":
		return charmlog.TextFormatter
	default:
		return charmlog.TextFormatter
	}
}
