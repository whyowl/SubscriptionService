package middleware

import (
	"context"
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type loggerKey struct{}

func FromContext(ctx context.Context) *zap.Logger {
	if v := ctx.Value(loggerKey{}); v != nil {
		if l, ok := v.(*zap.Logger); ok && l != nil {
			return l
		}
	}
	return zap.L()
}

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

func WithLogger(l *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sw := &statusWriter{ResponseWriter: w}
			start := time.Now()

			reqID := chimw.GetReqID(r.Context())
			reqLogger := l.With(
				zap.String("request_id", reqID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
			)

			ctx := context.WithValue(r.Context(), loggerKey{}, reqLogger)
			next.ServeHTTP(sw, r.WithContext(ctx))

			reqLogger.Info("request",
				zap.Int("status", sw.status),
				zap.Int("bytes", sw.bytes),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}
