// Package server wires the MCP Streamable-HTTP handler behind bearer auth,
// per-source rate limiting, and an unauthenticated health check.
package server

import (
	"crypto/subtle"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Options configure the HTTP surface.
type Options struct {
	// BearerToken is compared, in constant time, against the Authorization header.
	BearerToken string
	// RatePerMin caps requests per source IP per minute. <= 0 disables limiting.
	RatePerMin int
	// MCPHandler is the MCP Streamable-HTTP handler, mounted (auth-gated) at /mcp.
	MCPHandler http.Handler
}

// New returns the top-level HTTP handler: /healthz open, /mcp gated by auth and
// rate limiting.
func New(opts Options) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	lim := newLimiter(opts.RatePerMin)
	gated := authMiddleware(opts.BearerToken, lim.middleware(opts.MCPHandler))
	mux.Handle("/mcp", gated)
	mux.Handle("/mcp/", gated)
	return mux
}

func authMiddleware(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		got := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
		if subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, req)
	})
}

// limiter is a per-source-IP token bucket. Bucket size and refill rate both
// derive from perMin, so a client may burst up to perMin requests then refills
// at perMin/60 per second.
type limiter struct {
	perMin int
	mu     sync.Mutex
	seen   map[string]*bucket
}

type bucket struct {
	tokens float64
	last   time.Time
}

func newLimiter(perMin int) *limiter {
	return &limiter{perMin: perMin, seen: make(map[string]*bucket)}
}

func (l *limiter) allow(ip string) bool {
	if l.perMin <= 0 {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	rate := float64(l.perMin) / 60.0
	burst := float64(l.perMin)

	b := l.seen[ip]
	if b == nil {
		b = &bucket{tokens: burst, last: now}
		l.seen[ip] = b
	}
	b.tokens += now.Sub(b.last).Seconds() * rate
	if b.tokens > burst {
		b.tokens = burst
	}
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (l *limiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !l.allow(clientIP(req)) {
			http.Error(w, "rate limited", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, req)
	})
}

// clientIP prefers the first X-Forwarded-For hop (the server runs behind Caddy)
// and falls back to the transport remote address.
func clientIP(req *http.Request) string {
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return host
}
