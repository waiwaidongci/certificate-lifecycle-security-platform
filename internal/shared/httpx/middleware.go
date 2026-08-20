package httpx

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/acme/certpilot/internal/shared/apperror"
	"github.com/acme/certpilot/internal/shared/logger"
	"github.com/acme/certpilot/internal/shared/metrics"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", requestID)
		ctx := WithRequestID(r.Context(), requestID)
		ctx = context.WithValue(ctx, "request_start", time.Now())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AccessLog(log *logger.Logger, m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder := NewStatusRecorder(w)
			started := time.Now()
			next.ServeHTTP(recorder, r)
			latency := time.Since(started)
			if m != nil {
				m.Inc("http_requests_total")
				m.Observe("http_request_duration_seconds", latency.Seconds())
			}
			log.Info(r.Context(), "http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", recorder.Status(),
				"bytes", recorder.Bytes(),
				"duration_ms", latency.Milliseconds(),
				"request_id", RequestIDFromContext(r.Context()),
			)
		})
	}
}

func Timeout(duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), duration)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func PanicRecovery(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					log.Error(r.Context(), "panic recovered", "panic", recovered, "stack", string(debug.Stack()))
					WriteError(r.Context(), w, apperror.Internal(fmt.Errorf("%v", recovered)))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

type RateLimiter struct {
	mu     sync.Mutex
	window time.Time
	count  int
	rps    int
	burst  int
}

func NewRateLimiter(rps, burst int) *RateLimiter {
	return &RateLimiter{rps: rps, burst: burst, window: time.Now()}
}

func (l *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.allow() {
			WriteError(r.Context(), w, apperror.New(apperror.CodeRateLimited, "rate limit exceeded"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (l *RateLimiter) allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	if now.Sub(l.window) >= time.Second {
		l.window = now
		l.count = 0
	}
	if l.count >= l.burst {
		return false
	}
	if l.count >= l.rps && now.Sub(l.window) < time.Second {
		return false
	}
	l.count++
	return true
}

func AuthPlaceholder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The real deployment replaces this with an identity provider check.
		ctx := context.WithValue(r.Context(), "auth_actor", r.Header.Get("X-Actor"))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ActorFromContext(ctx context.Context) string {
	if value, ok := ctx.Value("auth_actor").(string); ok {
		return value
	}
	return "system"
}
