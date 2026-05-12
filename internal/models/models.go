package models

import "encoding/json"

// Envelope is the standard success response wrapper.
type Envelope struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
	Cache   *CacheInfo `json:"cache,omitempty"`
}

// ErrorBody is the standard error response.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// CacheInfo describes how a payload was served.
type CacheInfo struct {
	Hit bool   `json:"hit"`
	Key string `json:"key,omitempty"`
}

// HomeRails
type HomeRails struct {
	Trending json.RawMessage `json:"trending"`
	Seasonal json.RawMessage `json:"seasonal"`
	Popular  json.RawMessage `json:"popular"`
	Top      json.RawMessage `json:"top"`
	Upcoming json.RawMessage `json:"upcoming"`
	From     string          `json:"from,omitempty"`
}

// WatchSource
type WatchSource struct {
	Quality   string `json:"quality"`
	URL       string `json:"url"`         // already-absolute, ready to play
	Type      string `json:"type"`        // e.g. "video/mpegurl", "video/mp4"
	OldHLS    bool   `json:"old_hls"`     // true => single-quality HLS
	NeedProxy bool   `json:"need_proxy"`  // true => URL passes through HLS_PROXY_BASE
	ProxyURL  string `json:"proxy_url"`   // /api/proxy/hls?url=<URL> convenience
}

// WatchSkips
type WatchSkips struct {
	Intro *Range `json:"intro,omitempty"`
	Outro *Range `json:"outro,omitempty"`
}

// Range is a [start, end) interval in seconds.
type Range struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// Subtitles
type Subtitle struct {
	URL     string `json:"url"`
	Lang    string `json:"lang"`
	Label   string `json:"label,omitempty"`
	Default bool   `json:"default,omitempty"`
}

// WatchResponse is the normalized /api/watch payload.
type WatchResponse struct {
	ID         string          `json:"id"`
	Episode    int             `json:"episode"`
	Server     string          `json:"server"`
	SourceType string          `json:"source_type"` // "sub" or "dub"
	Sources    []WatchSource   `json:"sources"`
	Skips      *WatchSkips     `json:"skips,omitempty"`
	Subtitles  []Subtitle      `json:"subtitles,omitempty"`
	From       string          `json:"from,omitempty"`
}

// Server lists a single playback server returned by /api/anime/:id/servers.
type Server struct {
	ID      string `json:"id"`
	Default bool   `json:"default"`
	Tip     string `json:"tip"`
}
