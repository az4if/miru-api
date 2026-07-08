// Package config loads runtime configuration from environment variables with
// sensible defaults so the server runs out of the box on any free host.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all tunables for the API server.
type Config struct {
	// Port the HTTP server binds to. Defaults to 8080. Most PaaS injects PORT.
	Port string

	// Optional API key. When non-empty, every request must carry an
	// X-API-Key header that matches. /health and /api/proxy/hls remain open.
	APIKey string

	// Per-IP rate limit, requests per minute. Set to 0 to disable.
	RateLimitRPM int

	// Default cache TTL applied when a handler does not override it.
	CacheDefaultTTL time.Duration

	// Upstream HTTP request timeout.
	UpstreamTimeout time.Duration

	// Upstream HTTP host (Miru API root). Override only if Miru
	// changes hostnames.
	UpstreamBase string

	// HLS proxy host that Miru uses to serve playlists/segments.
	HLSProxyBase string

	// Forwarded Referer / User-Agent for upstream calls.
	UpstreamReferer   string
	UpstreamUserAgent string

	// Log level: trace, debug, info, warn, error.
	LogLevel string
}

// Load reads env vars and returns a populated Config.
func Load() *Config {
	return &Config{
		Port:              env("PORT", "8080"),
		APIKey:            env("API_KEY", ""),
		RateLimitRPM:      envInt("RATE_LIMIT_RPM", 60),
		CacheDefaultTTL:   envDuration("CACHE_DEFAULT_TTL", 10*time.Minute),
		UpstreamTimeout:   envDuration("UPSTREAM_TIMEOUT", 15*time.Second),
		UpstreamBase:      env("UPSTREAM_BASE", "https://miru.live/v2"),
		HLSProxyBase:      env("HLS_PROXY_BASE", "https://swiftstream.top/proxy"),
		UpstreamReferer:   env("UPSTREAM_REFERER", "https://miru.live/"),
		UpstreamUserAgent: env("UPSTREAM_USER_AGENT", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"),
		LogLevel:          env("LOG_LEVEL", "info"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
