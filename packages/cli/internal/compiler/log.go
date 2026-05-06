package compiler

import ailog "github.com/rluisb/lazyai/packages/cli/internal/log"

var compilerLog = componentLogger{component: "compiler"}

type componentLogger struct {
	component string
}

func (l componentLogger) Info(msg string, keyvals ...any) {
	ailog.Default().With("component", l.component).Info(msg, keyvals...)
}

func (l componentLogger) Warn(msg string, keyvals ...any) {
	ailog.Default().With("component", l.component).Warn(msg, keyvals...)
}

func (l componentLogger) Error(msg string, keyvals ...any) {
	ailog.Default().With("component", l.component).Error(msg, keyvals...)
}
