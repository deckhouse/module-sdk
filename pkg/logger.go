package pkg

import (
	"context"
	"io"
	"log/slog"

	"github.com/deckhouse/deckhouse/pkg/log"
)

type Logger interface {
	Debug(msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)

	Error(msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)

	Fatal(msg string, args ...any)

	Info(msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)

	Log(ctx context.Context, level slog.Level, msg string, args ...any)
	LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr)

	Warn(msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)

	Enabled(ctx context.Context, level slog.Level) bool
	With(args ...any) *log.Logger
	WithGroup(name string) *log.Logger
	Named(name string) *log.Logger
	SetLevel(level log.Level)
	SetOutput(w io.Writer)
	GetLevel() log.Level
	Handler() slog.Handler
}
