package mcp

import (
	charmlog "charm.land/log/v2"
	orchlog "github.com/rluisb/lazyai/packages/orchestrator/internal/log"
)

type componentLogger struct {
	component string
}

func (l componentLogger) logger() *charmlog.Logger {
	return orchlog.Default().With("component", l.component)
}

func (l componentLogger) Info(msg string, keyvals ...any) {
	l.logger().Info(msg, keyvals...)
}

func (l componentLogger) Warn(msg string, keyvals ...any) {
	l.logger().Warn(msg, keyvals...)
}

func (l componentLogger) Error(msg string, keyvals ...any) {
	l.logger().Error(msg, keyvals...)
}

var toolLog = componentLogger{component: "mcp"}
