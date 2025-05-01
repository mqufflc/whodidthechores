package api

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

func loggingMiddleware(ctx context.Context, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logger.LogAttrs(ctx, slog.LevelError, "error while processing request", slog.Any("msg", err), slog.String("trace", string(debug.Stack())))
				}
			}()
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.LogAttrs(ctx, slog.LevelInfo, "incoming request", slog.String("method", r.Method), slog.String("path", r.URL.EscapedPath()), slog.Int64("duration_ms", time.Since(start).Milliseconds()))
		}
		return http.HandlerFunc(fn)
	}
}
