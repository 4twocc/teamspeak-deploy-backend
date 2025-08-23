package utils

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// IPRateLimiter tracks rate limits for each IP
// This is a simple in-memory rate limiter. In a production environment,
// consider using a distributed rate limiter like Redis for multi-instance deployments.
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit // requests per second
	b   int        // burst size
}

// NewIPRateLimiter creates a new rate limiter
// r: number of requests per second
// b: burst size (maximum number of requests that can be made in a single burst)
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

// AddIP creates a new rate limiter for an IP address and adds it to the map
func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.r, i.b)
	i.ips[ip] = limiter

	return limiter
}

// GetLimiter returns the rate limiter for the provided IP address if it exists.
// Otherwise, it calls AddIP to add the IP address to the map.
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	// 添加空指针检查
	if i == nil {
		return nil
	}

	i.mu.RLock()
	limiter, exists := i.ips[ip]
	i.mu.RUnlock()

	if !exists {
		// Upgrade to write lock and double-check
		i.mu.Lock()
		defer i.mu.Unlock()
		// Double-check in case limiter was added while acquiring lock
		if existingLimiter, found := i.ips[ip]; found {
			return existingLimiter
		}
		// Create new limiter if it doesn't exist
		limiter = rate.NewLimiter(i.r, i.b)
		// 添加空指针检查
		if limiter == nil {
			return nil
		}
		// 添加 limiter 到 ips 映射
		i.ips[ip] = limiter
		return limiter
	}
	return limiter
}

// RateLimitMiddleware creates a middleware that limits requests per IP (net/http)
func RateLimitMiddleware(r rate.Limit, b int) func(http.Handler) http.Handler {
	limiter := NewIPRateLimiter(r, b)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if !limiter.GetLimiter(ip).Allow() {
				Fail(w, http.StatusTooManyRequests, ErrTooManyRequests, "Too many requests")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
