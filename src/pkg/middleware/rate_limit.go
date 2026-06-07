package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/timoruohomaki/open311-to-Go/pkg/httputil"
)

// RateLimitMiddleware limits each client to requestsPerMinute requests using a
// fixed window. A non-positive limit disables rate limiting. /health is exempt.
// On exceed it responds 429 with a Retry-After header.
func RateLimitMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
	if requestsPerMinute <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}

	limiter := newFixedWindowLimiter(requestsPerMinute, time.Minute)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			allowed, retryAfter := limiter.allow(clientIP(r), time.Now())
			if !allowed {
				secs := int(retryAfter.Seconds())
				if secs < 1 {
					secs = 1
				}
				w.Header().Set("Retry-After", strconv.Itoa(secs))
				_ = httputil.SendError(w, r, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// clientIP extracts the client address, preferring the first X-Forwarded-For
// hop (set by the fronting proxy/Nginx) and falling back to RemoteAddr.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

type windowCounter struct {
	count   int
	resetAt time.Time
}

type fixedWindowLimiter struct {
	mu        sync.Mutex
	limit     int
	window    time.Duration
	clients   map[string]*windowCounter
	lastSweep time.Time
}

func newFixedWindowLimiter(limit int, window time.Duration) *fixedWindowLimiter {
	return &fixedWindowLimiter{
		limit:   limit,
		window:  window,
		clients: make(map[string]*windowCounter),
	}
}

// allow reports whether a request from key is permitted at time now, and if not,
// how long until the current window resets.
func (l *fixedWindowLimiter) allow(key string, now time.Time) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.sweep(now)

	c, ok := l.clients[key]
	if !ok || now.After(c.resetAt) {
		l.clients[key] = &windowCounter{count: 1, resetAt: now.Add(l.window)}
		return true, 0
	}
	if c.count >= l.limit {
		return false, c.resetAt.Sub(now)
	}
	c.count++
	return true, 0
}

// sweep evicts expired client entries at most once per window to bound memory.
func (l *fixedWindowLimiter) sweep(now time.Time) {
	if now.Sub(l.lastSweep) < l.window {
		return
	}
	l.lastSweep = now
	for k, c := range l.clients {
		if now.After(c.resetAt) {
			delete(l.clients, k)
		}
	}
}
