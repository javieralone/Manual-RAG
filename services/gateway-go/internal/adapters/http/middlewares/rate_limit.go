package middlewares

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"api-go/internal/adapters/http/response"
	"api-go/internal/adapters/observability"
)

type rateLimitEntry struct {
	started time.Time
	count   int
}

type rateLimiter struct {
	mu      sync.Mutex
	entries map[string]rateLimitEntry
	limit   int
	window  time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{entries: make(map[string]rateLimitEntry), limit: limit, window: window}
}

func (l *rateLimiter) allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.entries) >= 1024 {
		for entryKey, entry := range l.entries {
			if now.Sub(entry.started) >= l.window {
				delete(l.entries, entryKey)
			}
		}
	}

	entry, ok := l.entries[key]
	if !ok || now.Sub(entry.started) >= l.window {
		l.entries[key] = rateLimitEntry{started: now, count: 1}
		return true
	}
	if entry.count >= l.limit {
		return false
	}
	entry.count++
	l.entries[key] = entry
	return true
}

// RateLimitMiddleware applies independent fixed-window limits to the client IP and identity.
func RateLimitMiddleware(limit int, window time.Duration, metrics *observability.Metrics, logger *slog.Logger) func(http.Handler) http.Handler {
	ipLimiter := newRateLimiter(limit, window)
	userLimiter := newRateLimiter(limit, window)
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route := r.URL.Path
			ip := clientIP(r)
			if !ipLimiter.allow(ip, time.Now()) {
				rejectRateLimit(w, logger, metrics, "ip", route, window)
				return
			}

			if identity, ok := IdentityFromContext(r.Context()); ok {
				if !userLimiter.allow(identity.Username, time.Now()) {
					rejectRateLimit(w, logger, metrics, "user", route, window)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func rejectRateLimit(w http.ResponseWriter, logger *slog.Logger, metrics *observability.Metrics, scope, route string, window time.Duration) {
	if metrics != nil {
		metrics.RateLimitRejections.WithLabelValues(scope, route).Inc()
	}
	logger.Warn("rate_limit_rejected", "scope", scope, "route", route)
	seconds := int(window.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	w.Header().Set("Retry-After", fmt.Sprintf("%d", seconds))
	response.WriteError(w, http.StatusTooManyRequests, "límite de peticiones excedido")
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host
	}
	if strings.TrimSpace(r.RemoteAddr) != "" {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return "unknown"
}
