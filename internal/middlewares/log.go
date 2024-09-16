package middlewares

import (
	"avito_intership/pkg/logger"
	"context"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

func Log(l *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logID := uuid.New().ID()
			ctx := context.WithValue(r.Context(), logger.LogIDContextKey, logID)
			r = r.WithContext(ctx)

			l.Info("REQUEST",
				slog.Any("ID", logID),
				slog.String("Method", r.Method),
				slog.String("Host", r.Host),
				slog.String("Request URI", r.RequestURI),
				slog.String("Requested From", r.RemoteAddr))

			next.ServeHTTP(w, r)
		})
	}
}
