package rpc

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const maxRequestBytes = 1 << 20

type rateLimiter struct {
	mu      sync.Mutex
	clients map[string]*clientBucket
	limit   int
	window  time.Duration
}

type clientBucket struct {
	count     int
	resetTime time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{clients: make(map[string]*clientBucket), limit: limit, window: window}
}

func (rl *rateLimiter) Allow(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	now := time.Now()
	rl.mu.Lock()
	defer rl.mu.Unlock()
	bucket := rl.clients[host]
	if bucket == nil || now.After(bucket.resetTime) {
		rl.clients[host] = &clientBucket{count: 1, resetTime: now.Add(rl.window)}
		return true
	}
	if bucket.count >= rl.limit {
		return false
	}
	bucket.count++
	return true
}

func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		s.metrics.IncRequest(r.URL.Path)
		if !s.rateLimiter.Allow(r.RemoteAddr) {
			s.metrics.IncRejected()
			writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
		logEvent("rpc.request", map[string]any{
			"method": r.Method,
			"path":   r.URL.Path,
			"remote": clientIP(r),
			"ms":     time.Since(start).Milliseconds(),
		})
	})
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
