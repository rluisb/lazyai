package log

import (
	"os"
	"strings"
	"sync"

	"charm.land/lipgloss/v2"
	charmlog "charm.land/log/v2"

	"github.com/rluisb/lazyai/packages/cli/internal/theme"
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

	logger := charmlog.NewWithOptions(os.Stderr, charmlog.Options{
		Level:           parseLevel(level),
		Formatter:       parseFormat(format),
		Prefix:          prefix,
		ReportTimestamp: false,
	})
	logger.SetStyles(themedStyles())
	return logger
}

// themedStyles returns a charmlog Styles configuration that maps each log
// level to its canonical design-system glyph + color (FR-008).
//
// charmlog renders the level via `Style.String()` (see text formatter at
// charm.land/log/v2/text.go:191), which honors `SetString` content. Setting
// the per-level style with `.SetString(glyph).Foreground(token)` replaces
// the default `info`/`warn`/`error`/`debug` keyword with the canonical
// design-system glyph + color.
//
// Mapping:
//
//	debug → `○` in dimmed   (theme.GlyphPending — passive / suppressed by default)
//	info  → `•` in secondary (theme.GlyphBullet)
//	warn  → `⚠` in warning   (theme.GlyphWarn)
//	error → `✗` in error     (theme.GlyphError)
//	fatal → `✗` in error, bold (escalated)
//
// Output styling auto-strips ANSI when the writer is not a TTY (lipgloss's
// renderer applies the writer's color profile through charmlog's
// TextFormatter).
func themedStyles() *charmlog.Styles {
	s := charmlog.DefaultStyles()
	s.Levels = map[charmlog.Level]lipgloss.Style{
		charmlog.DebugLevel: lipgloss.NewStyle().
			SetString(theme.GlyphPending).
			Foreground(theme.Dimmed),
		charmlog.InfoLevel: lipgloss.NewStyle().
			SetString(theme.GlyphBullet).
			Foreground(theme.Secondary),
		charmlog.WarnLevel: lipgloss.NewStyle().
			SetString(theme.GlyphWarn).
			Foreground(theme.Warning),
		charmlog.ErrorLevel: lipgloss.NewStyle().
			SetString(theme.GlyphError).
			Foreground(theme.Error),
		charmlog.FatalLevel: lipgloss.NewStyle().
			SetString(theme.GlyphError).
			Foreground(theme.Error).
			Bold(true),
	}
	s.Prefix = lipgloss.NewStyle().Foreground(theme.Primary).Bold(true)
	s.Message = lipgloss.NewStyle().Foreground(theme.Text)
	s.Key = lipgloss.NewStyle().Foreground(theme.Highlight)
	s.Value = lipgloss.NewStyle().Foreground(theme.Text)
	s.Separator = lipgloss.NewStyle().Foreground(theme.Dimmed)
	return s
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
