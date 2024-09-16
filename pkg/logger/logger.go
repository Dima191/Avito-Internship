package logger

import (
	"context"
	"log/slog"
	"os"
)

const (
	LogIDContextKey = "log_id"
	LogIDFieldName  = "LOG_ID"
)

func New() *slog.Logger {
	logOptions := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelWarn,
	}

	return slog.New(slog.NewTextHandler(os.Stdout, logOptions))
}

func EndToEndLogging(ctx context.Context, l *slog.Logger) *slog.Logger {
	logID, ok := ctx.Value(LogIDContextKey).(uint32)
	if !ok {
		l.Error("Failed to get log id")
	}

	return l.With(slog.Group("end-to-end",
		slog.Any(LogIDFieldName, logID),
	))
}
