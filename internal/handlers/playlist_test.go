package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/miru/api/internal/cache"
	"github.com/miru/api/internal/client"
	"github.com/miru/api/internal/config"
)

func TestPrioritizeServers(t *testing.T) {
	input := []string{"pahe", "meg", "kite", "fsoft", "zoro"}
	expected := []string{"kite", "zoro", "pahe", "meg", "fsoft"}

	output := prioritizeServers(input)
	if len(output) != len(expected) {
		t.Fatalf("expected length %d, got %d", len(expected), len(output))
	}
	for i, v := range expected {
		if output[i] != v {
			t.Errorf("at index %d: expected %s, got %s", i, v, output[i])
		}
	}
}

func TestExpandMasterPlaylist(t *testing.T) {
	// Setup mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "master.m3u8") {
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:BANDWIDTH=800000,RESOLUTION=640x360,NAME="360p"
319579360.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=2800000,RESOLUTION=1280x720,NAME="720p"
319579720.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=5500000,RESOLUTION=1920x1080,NAME="1080p"
3195791080.m3u8`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	cfg := &config.Config{
		UpstreamTimeout: 5 * time.Second,
		UpstreamBase:    mockServer.URL,
		HLSProxyBase:    mockServer.URL,
	}

	hc := client.New(cfg)
	cc := cache.New(10 * time.Minute)
	h := New(hc, cc, mockServer.URL)

	src := rawSource{
		Quality:   "auto",
		URL:       mockServer.URL + "/media/master.m3u8",
		Type:      "video/mpegurl",
		OldHLS:    false,
		NeedProxy: true,
	}

	variants := h.expandMasterPlaylist(context.Background(), src)
	if len(variants) != 3 {
		t.Fatalf("expected 3 variants, got %d", len(variants))
	}

	expectedQualities := []string{"360p", "720p", "1080p"}
	for i, q := range expectedQualities {
		if variants[i].Quality != q {
			t.Errorf("expected quality %s, got %s", q, variants[i].Quality)
		}
		expectedSuffix := ""
		switch q {
		case "360p":
			expectedSuffix = "319579360.m3u8"
		case "720p":
			expectedSuffix = "319579720.m3u8"
		case "1080p":
			expectedSuffix = "3195791080.m3u8"
		}
		if !strings.HasSuffix(variants[i].URL, expectedSuffix) {
			t.Errorf("expected URL ending with %s, got %s", expectedSuffix, variants[i].URL)
		}
		if !variants[i].OldHLS {
			t.Errorf("expected variant %s to have OldHLS=true", q)
		}
	}
}
