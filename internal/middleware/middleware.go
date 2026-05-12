// Package middleware bundles the small custom middlewares used by the server.
package middleware

import (
	"strings"
	"sync"
	"time"

	"github.com/animetsu/api/internal/config"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

func isOpen(path string) bool {
	switch path {
	case "/", "/health", "/healthz", "/livez", "/readyz", "/ping",
		"/docs", "/demo", "/openapi.json":
		return true
	}
	return strings.HasPrefix(path, "/api/proxy/")
}

func APIKey(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if cfg.APIKey == "" {
			return c.Next()
		}
		if isOpen(c.Path()) {
			return c.Next()
		}
		key := c.Get("X-API-Key")
		if key == "" {
			key = c.Query("key")
		}
		if key != cfg.APIKey {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or missing api key")
		}
		return c.Next()
	}
}

func RequestLog(log zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		log.Info().
			Str("method", c.Method()).
			Str("path", c.OriginalURL()).
			Int("status", c.Response().StatusCode()).
			Str("ip", c.IP()).
			Dur("latency", time.Since(start)).
			Msg("req")
		return err
	}
}

func RateLimit(n int) fiber.Handler {
	if n <= 0 {
		return func(c *fiber.Ctx) error { return c.Next() }
	}
	type bucket struct {
		mu     sync.Mutex
		tokens float64
		last   time.Time
	}
	var buckets sync.Map
	per := float64(n)
	return func(c *fiber.Ctx) error {
		if isOpen(c.Path()) {
			return c.Next()
		}
		ip := c.IP()
		actual, _ := buckets.LoadOrStore(ip, &bucket{tokens: per, last: time.Now()})
		b := actual.(*bucket)
		b.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(b.last).Minutes()
		b.tokens = mins(per, b.tokens+elapsed*per)
		b.last = now
		if b.tokens < 1 {
			b.mu.Unlock()
			return fiber.NewError(fiber.StatusTooManyRequests, "rate limit exceeded")
		}
		b.tokens--
		b.mu.Unlock()
		return c.Next()
	}
}

func mins(a, b float64) float64 { if a < b { return a }; return b }
