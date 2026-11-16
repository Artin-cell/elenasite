package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	count int
	reset time.Time
}

type RateLimiter struct {
	mu         sync.Mutex
	visitors   map[string]visitor
	limit      int
	window     time.Duration
	maxVisitors int // MEDIUM #6 FIX: ограничение на размер map для защиты от утечки памяти
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	r := &RateLimiter{
		visitors:    make(map[string]visitor),
		limit:       limit,
		window:      window,
		maxVisitors: 10000, // MEDIUM #6 FIX: максимум 10к уникальных IP в памяти
	}
	go r.cleanupLoop()
	return r
}

func (r *RateLimiter) cleanupLoop() {
	interval := r.window
	if interval <= 0 {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		r.mu.Lock()
		for key, v := range r.visitors {
			if now.After(v.reset) {
				delete(r.visitors, key)
			}
		}
		r.mu.Unlock()
	}
}

func (r *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if r.limit <= 0 {
			c.Next()
			return
		}
		now := time.Now()
		key := c.ClientIP()

		r.mu.Lock()
		v, ok := r.visitors[key]
		if !ok || now.After(v.reset) {
			// MEDIUM #6 FIX: если map переполнен, удаляем 10% самых старых записей
			if len(r.visitors) >= r.maxVisitors {
				var oldestKeys []string
				cutoff := now.Add(-r.window * 2)
				for k, vis := range r.visitors {
					if vis.reset.Before(cutoff) {
						oldestKeys = append(oldestKeys, k)
					}
				}
				for _, k := range oldestKeys {
					delete(r.visitors, k)
				}
			}
			v = visitor{count: 0, reset: now.Add(r.window)}
		}
		v.count++
		r.visitors[key] = v
		allowed := v.count <= r.limit
		retryAfter := v.reset.Sub(now)
		r.mu.Unlock()

		if !allowed {
			c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())+1))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "слишком много запросов"})
			return
		}
		c.Next()
	}
}
