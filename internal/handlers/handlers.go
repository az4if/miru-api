// Package handlers wires every JSON endpoint exposed by the API.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/miru/api/internal/cache"
	"github.com/miru/api/internal/client"
	"github.com/miru/api/internal/models"
	"github.com/miru/api/internal/proxy"
	"github.com/miru/api/pkg/hls"
	"github.com/gofiber/fiber/v2"
)

type H struct {
	Client       *client.Client
	Cache        *cache.Cache
	HLSProxyBase string
}

func New(c *client.Client, cc *cache.Cache, hlsBase string) *H {
	return &H{Client: c, Cache: cc, HLSProxyBase: strings.TrimRight(hlsBase, "/")}
}

// ----- helpers ---------------------------------------------------------------

func ok(c *fiber.Ctx, data any) error {
	return c.JSON(models.Envelope{Success: true, Data: data})
}

func cached(h *H, c *fiber.Ctx, key string, ttl time.Duration, fn func(ctx context.Context) (any, error)) error {
	v, err := h.Cache.GetOrFetch(c.Context(), key, ttl, fn)
	if err != nil {
		return mapUpstreamErr(err)
	}
	return ok(c, v)
}

func mapUpstreamErr(err error) error {
	if ue, okErr := err.(*client.UpstreamError); okErr {
		return fiber.NewError(http.StatusBadGateway, ue.Error())
	}
	return fiber.NewError(http.StatusBadGateway, err.Error())
}

func (h *H) fetchJSON(ctx context.Context, path string) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := h.Client.GetJSON(ctx, path, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// ----- meta -----------------------------------------------------------------

func (h *H) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"ok":      true,
		"name":    "miru-api",
		"version": "1.0.0",
		"now":     time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *H) Root(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"name":    "miru-api",
		"ok":      true,
		"version": "1.0.0",
		"docs":    "/docs",
		"demo":    "/demo",
		"openapi": "/openapi.json",
		"endpoints": []string{
			"GET /healthz",
			"GET /api/home",
			"GET /api/trending",
			"GET /api/recent?page=1",
			"GET /api/schedule",
			"GET /api/search?q=naruto",
			"GET /api/anime/:id",
			"GET /api/anime/:id/episodes",
			"GET /api/anime/:id/servers/:ep",
			"GET /api/anime/:id/watch/:ep?server=auto&source_type=sub",
			"GET /api/anime/:id/download/:ep?quality=1080p",
			"GET /api/anime/:id/downloads/:ep?quality=1080p&group=SubsPlease&type=sub",
			"GET /api/watch?id=:id&ep=:ep&server=auto&source_type=sub",
			"GET /api/proxy/hls?url=<absolute>",
		},
	})
}

// ----- discovery ------------------------------------------------------------

func (h *H) Home(c *fiber.Ctx) error {
	return cached(h, c, "home", 10*time.Minute, func(ctx context.Context) (any, error) {
		return h.fetchJSON(ctx, "/api/anime/home")
	})
}

func (h *H) rail(name string, _ time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		v, err := h.Cache.GetOrFetch(c.Context(), "home", 10*time.Minute, func(ctx context.Context) (any, error) {
			return h.fetchJSON(ctx, "/api/anime/home")
		})
		if err != nil {
			return mapUpstreamErr(err)
		}
		raw, _ := v.(json.RawMessage)
		var parsed map[string]json.RawMessage
		if err := json.Unmarshal(raw, &parsed); err != nil {
			return fiber.NewError(http.StatusBadGateway, "decode home: "+err.Error())
		}
		section, has := parsed[name]
		if !has {
			return fiber.NewError(http.StatusNotFound, "rail "+name+" not present")
		}
		return ok(c, section)
	}
}

func (h *H) Trending(c *fiber.Ctx) error { return h.rail("trending", 10*time.Minute)(c) }
func (h *H) Season(c *fiber.Ctx) error   { return h.rail("seasonal", 30*time.Minute)(c) }
func (h *H) Popular(c *fiber.Ctx) error  { return h.rail("popular", time.Hour)(c) }
func (h *H) TopRated(c *fiber.Ctx) error { return h.rail("top", time.Hour)(c) }
func (h *H) Upcoming(c *fiber.Ctx) error { return h.rail("upcoming", time.Hour)(c) }

func (h *H) Recent(c *fiber.Ctx) error {
	page := c.Query("page", "1")
	per := c.Query("per_page", "12")
	key := "recent:p=" + page + ":n=" + per
	return cached(h, c, key, 2*time.Minute, func(ctx context.Context) (any, error) {
		return h.fetchJSON(ctx, "/api/anime/recent?page="+page+"&per_page="+per)
	})
}

func (h *H) Schedule(c *fiber.Ctx) error {
	return cached(h, c, "schedule", 30*time.Minute, func(ctx context.Context) (any, error) {
		return h.fetchJSON(ctx, "/api/anime/schedule")
	})
}

func (h *H) Random(c *fiber.Ctx) error {
	v, err := h.fetchJSON(c.Context(), "/api/anime/random")
	if err != nil {
		return mapUpstreamErr(err)
	}
	return ok(c, v)
}

func (h *H) Search(c *fiber.Ctx) error {
	params := map[string]string{
		"query":  c.Query("q", ""),
		"page":   c.Query("page", "1"),
		"genres": c.Query("genres", ""),
		"format": c.Query("format", ""),
		"status": c.Query("status", ""),
		"season": c.Query("season", ""),
		"year":   c.Query("year", ""),
		"sort":   c.Query("sort", ""),
	}
	for k, v := range params {
		if v == "" && k != "query" {
			delete(params, k)
		}
	}
	qs := client.BuildQuery(params)
	return cached(h, c, "search:"+qs, 5*time.Minute, func(ctx context.Context) (any, error) {
		return h.fetchJSON(ctx, "/api/anime/search?"+qs)
	})
}

func (h *H) Info(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(http.StatusBadRequest, "id required")
	}
	resolved, err := h.resolveId(c.Context(), id)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, err.Error())
	}
	id = resolved
	return cached(h, c, "info:"+id, 30*time.Minute, func(ctx context.Context) (any, error) {
		return h.fetchJSON(ctx, "/api/anime/info/"+id)
	})
}

func (h *H) Episodes(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(http.StatusBadRequest, "id required")
	}
	resolved, err := h.resolveId(c.Context(), id)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, err.Error())
	}
	id = resolved
	return cached(h, c, "eps:"+id, 15*time.Minute, func(ctx context.Context) (any, error) {
		return h.fetchJSON(ctx, "/api/anime/eps/"+id)
	})
}

func (h *H) Views(c *fiber.Ctx) error {
	id := c.Params("id")
	ep := c.Params("ep")
	if id == "" || ep == "" {
		return fiber.NewError(http.StatusBadRequest, "id and ep required")
	}
	resolved, err := h.resolveId(c.Context(), id)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, err.Error())
	}
	id = resolved
	return cached(h, c, "views:"+id+":"+ep, 1*time.Minute, func(ctx context.Context) (any, error) {
		return h.fetchJSON(ctx, "/api/anime/views/"+id+"/"+ep)
	})
}

// ----- streaming ------------------------------------------------------------

// rawSource matches the upstream watch payload.
type rawSource struct {
	Quality   string `json:"quality"`
	URL       string `json:"url"`
	Type      string `json:"type"`
	OldHLS    bool   `json:"old_hls"`
	NeedProxy bool   `json:"need_proxy"`
}

type rawSub struct {
	URL  string `json:"url"`
	Lang string `json:"lang"`
}

type rawWatch struct {
	Sources   []rawSource        `json:"sources"`
	Skips     *models.WatchSkips `json:"skips"`
	Subtitles []rawSub           `json:"subs"`
	Server    string             `json:"server"`
	From      string             `json:"from"`
}

// fetchWatch fetches a watch payload for a specific server (no fallback).
func (h *H) fetchWatch(ctx context.Context, id, ep, server, srcType string) (*rawWatch, error) {
	path := "/api/anime/oppai/" + id + "/" + ep + "?server=" + server + "&source_type=" + srcType
	var rw rawWatch
	if err := h.Client.GetJSON(ctx, path, &rw); err != nil {
		return nil, err
	}
	return &rw, nil
}

// Servers lists upstream servers AND probes which actually deliver bytes.
// Cached 5 min so repeated picker hits are instant.
func (h *H) Servers(c *fiber.Ctx) error {
	id := c.Params("id")
	ep := c.Params("ep")
	if ep == "" {
		ep = c.Query("ep", "1")
	}
	if id == "" {
		return fiber.NewError(http.StatusBadRequest, "id required")
	}
	resolved, err := h.resolveId(c.Context(), id)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, err.Error())
	}
	id = resolved
	srcType := normalizeSourceType(c)
	key := "servers:" + id + ":" + ep + ":" + srcType
	return cached(h, c, key, 5*time.Minute, func(ctx context.Context) (any, error) {
		type result struct {
			ID        string `json:"id"`
			Default   bool   `json:"default"`
			Tip       string `json:"tip"`
			Working   bool   `json:"working"`
			Sources   int    `json:"sources"`
			LatencyMS int    `json:"latency_ms"`
		}

		var (
			animetsuOut []result
			animexOut   []result
			miruroOut   []result
			wgOuter     sync.WaitGroup
		)

		// 1. Animetsu fetch and probe
		wgOuter.Add(1)
		go func() {
			defer wgOuter.Done()
			var list []models.Server
			if err := h.Client.GetJSON(ctx, "/api/anime/servers/"+id+"/"+ep, &list); err != nil {
				return
			}
			resList := make([]result, len(list))
			var wg sync.WaitGroup
			for i, s := range list {
				i, s := i, s
				wg.Add(1)
				go func() {
					defer wg.Done()
					start := time.Now()
					pctx, cancel := context.WithTimeout(ctx, 8*time.Second)
					defer cancel()
					rw, err := h.fetchWatch(pctx, id, ep, s.ID, srcType)
					ms := int(time.Since(start).Milliseconds())
					if err != nil || rw == nil || len(rw.Sources) == 0 {
						resList[i] = result{ID: "animetsu:" + s.ID, Default: s.Default, Tip: s.Tip, Working: false, LatencyMS: ms}
						return
					}
					abs := absolutizeSource(rw.Sources[0], h.HLSProxyBase)
					code, _, perr := h.Client.HeadOrGet(pctx, abs)
					working := perr == nil && code < 400
					resList[i] = result{
						ID: "animetsu:" + s.ID, Default: s.Default, Tip: s.Tip,
						Working: working, Sources: len(rw.Sources),
						LatencyMS: int(time.Since(start).Milliseconds()),
					}
				}()
			}
			wg.Wait()
			animetsuOut = resList
		}()

		// 2. Animex fetch
		wgOuter.Add(1)
		go func() {
			defer wgOuter.Done()
			epVal, _ := strconv.Atoi(ep)
			if ax, err := h.animexServers(ctx, id, epVal, srcType); err == nil {
				for _, s := range ax {
					tip := "Soft sub, Multi quality"
					if srcType == "dub" {
						tip = "Hard sub, Multi quality"
					}
					animexOut = append(animexOut, result{
						ID:      "animex:" + s,
						Default: false,
						Tip:     tip,
						Working: true,
						Sources: 1,
					})
				}
			}
		}()

		// 3. Miruro fetch
		wgOuter.Add(1)
		go func() {
			defer wgOuter.Done()
			epVal, _ := strconv.Atoi(ep)
			if mr, err := h.miruroServers(ctx, id, epVal, srcType); err == nil {
				for _, s := range mr {
					tip := "Hard sub, Multi quality"
					if strings.ToLower(s) == "ally" {
						tip = "Soft sub, Multi quality"
					}
					miruroOut = append(miruroOut, result{
						ID:      "miruro:" + s,
						Default: false,
						Tip:     tip,
						Working: true,
						Sources: 1,
					})
				}
			}
		}()

		wgOuter.Wait()

		out := append(animetsuOut, animexOut...)
		out = append(out, miruroOut...)

		// Sort working first, then by latency.
		sort.SliceStable(out, func(i, j int) bool {
			if out[i].Working != out[j].Working {
				return out[i].Working
			}
			return out[i].LatencyMS < out[j].LatencyMS
		})
		return out, nil
	})
}

func absolutizeSource(s rawSource, hlsBase string) string {
	// Already an absolute URL — use as-is.
	if strings.HasPrefix(s.URL, "http://") || strings.HasPrefix(s.URL, "https://") {
		return s.URL
	}
	// Relative path — must be turned into an absolute URL regardless of
	// need_proxy so the player (or our HLS proxy) can actually fetch it.
	// When need_proxy is true the path goes through hlsBase (the stream
	// proxy); otherwise we just prepend hlsBase as well since the cipher
	// path is always routed through the same proxy tier.
	rel := s.URL
	if !strings.HasPrefix(rel, "/") {
		rel = "/" + rel
	}
	return strings.TrimRight(hlsBase, "/") + rel
}

// Watch resolves playable sources for one episode, with optional automatic
// server fallback. Set ?server=auto (default) to try each server and return
// the first one that actually serves bytes.
func (h *H) Watch(c *fiber.Ctx) error {
	id := c.Query("id")
	ep := c.Query("ep", "1")
	if id == "" {
		return fiber.NewError(http.StatusBadRequest, "id required")
	}
	resolved, err := h.resolveId(c.Context(), id)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, err.Error())
	}
	id = resolved
	server := c.Query("server", "auto")
	srcType := normalizeSourceType(c)
	fallback := c.Query("fallback", "true") != "false"

	scheme := "https"
	if c.Protocol() == "http" {
		scheme = "http"
	}
	selfBase := scheme + "://" + c.Hostname()

	cacheKey := "watch:" + id + ":" + ep + ":" + server + ":" + srcType + ":fb=" + strconv.FormatBool(fallback)
	return cached(h, c, cacheKey, 5*time.Minute, func(ctx context.Context) (any, error) {
		if strings.HasPrefix(server, "animex:") {
			realServer := strings.TrimPrefix(server, "animex:")
			epVal, _ := strconv.Atoi(ep)
			return h.animexWatch(ctx, id, epVal, realServer, srcType, selfBase)
		}
		if strings.HasPrefix(server, "miruro:") {
			realServer := strings.TrimPrefix(server, "miruro:")
			epVal, _ := strconv.Atoi(ep)
			return h.miruroWatch(ctx, id, epVal, realServer, srcType, selfBase)
		}

		requestedServer := server
		if strings.HasPrefix(requestedServer, "animetsu:") {
			requestedServer = strings.TrimPrefix(requestedServer, "animetsu:")
		}
		order := buildServerOrder(h, ctx, id, ep, requestedServer, fallback)

		var (
			rw         *rawWatch
			usedServer string
			lastErr    error
		)
		for _, sv := range order {
			pctx, cancel := context.WithTimeout(ctx, 8*time.Second)
			r, err := h.fetchWatch(pctx, id, ep, sv, srcType)
			cancel()
			if err != nil || r == nil || len(r.Sources) == 0 {
				lastErr = err
				continue
			}
			// Quick probe of first source.
			abs := absolutizeSource(r.Sources[0], h.HLSProxyBase)
			pctx2, pcancel := context.WithTimeout(ctx, 5*time.Second)
			code, _, perr := h.Client.HeadOrGet(pctx2, abs)
			pcancel()
			if !fallback || (perr == nil && code < 400) {
				rw = r
				usedServer = sv
				if r.Server != "" {
					usedServer = r.Server
				}
				break
			}
			lastErr = perr
		}
		if rw == nil {
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, fiber.NewError(http.StatusBadGateway, "no working server returned playable sources")
		}

		out := models.WatchResponse{
			ID:         id,
			Server:     "animetsu:" + usedServer,
			SourceType: srcType,
			Skips:      rw.Skips,
			Subtitles:  buildSubtitles(rw.Subtitles, selfBase),
			From:       rw.From,
		}
		if n, err := strconv.Atoi(ep); err == nil {
			out.Episode = n
		}
		var expandedSources []rawSource
		for _, s := range rw.Sources {
			expandedSources = append(expandedSources, h.expandMasterPlaylist(ctx, s)...)
		}
		for _, s := range expandedSources {
			abs := absolutizeSource(s, h.HLSProxyBase)
			// Every source is HLS in this build; proxy it through /api/proxy/hls.
			out.Sources = append(out.Sources, models.WatchSource{
				Quality:   s.Quality,
				URL:       abs,
				Type:      s.Type,
				OldHLS:    s.OldHLS,
				NeedProxy: s.NeedProxy,
				ProxyURL:  selfBase + "/api/proxy/hls?url=" + queryEscape(abs),
			})
		}
		return out, nil
	})
}

// WatchByPath is the REST-style alias.
func (h *H) WatchByPath(c *fiber.Ctx) error {
	id := c.Params("id")
	ep := c.Params("ep")
	if id == "" || ep == "" {
		return fiber.NewError(http.StatusBadRequest, "id and ep required")
	}
	c.Request().URI().QueryArgs().Set("id", id)
	c.Request().URI().QueryArgs().Set("ep", ep)
	return h.Watch(c)
}

// Download returns the highest-quality direct/proxied source URL plus a
// suggested filename. Players (and curl/yt-dlp/ffmpeg) can grab this.
//
//	GET /api/anime/:id/download/:ep?quality=1080p&server=auto&source_type=sub
func (h *H) Download(c *fiber.Ctx) error {
	id := c.Params("id")
	ep := c.Params("ep")
	if id == "" || ep == "" {
		return fiber.NewError(http.StatusBadRequest, "id and ep required")
	}
	resolved, err := h.resolveId(c.Context(), id)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, err.Error())
	}
	id = resolved
	wantQ := strings.ToLower(c.Query("quality", ""))
	c.Request().URI().QueryArgs().Set("id", id)
	c.Request().URI().QueryArgs().Set("ep", ep)
	server := c.Query("server", "auto")
	srcType := normalizeSourceType(c)
	epVal, _ := strconv.Atoi(ep)

	scheme := "https"
	if c.Protocol() == "http" {
		scheme = "http"
	}
	selfBase := scheme + "://" + c.Hostname()

	var (
		watchResp models.WatchResponse
	)

	if strings.HasPrefix(server, "animex:") {
		realServer := strings.TrimPrefix(server, "animex:")
		res, err := h.animexWatch(c.Context(), id, epVal, realServer, srcType, selfBase)
		if err != nil {
			return fiber.NewError(http.StatusBadGateway, err.Error())
		}
		watchResp = res.(models.WatchResponse)
	} else if strings.HasPrefix(server, "miruro:") {
		realServer := strings.TrimPrefix(server, "miruro:")
		res, err := h.miruroWatch(c.Context(), id, epVal, realServer, srcType, selfBase)
		if err != nil {
			return fiber.NewError(http.StatusBadGateway, err.Error())
		}
		watchResp = res.(models.WatchResponse)
	} else {
		requestedServer := server
		if strings.HasPrefix(requestedServer, "animetsu:") {
			requestedServer = strings.TrimPrefix(requestedServer, "animetsu:")
		}
		order := buildServerOrder(h, c.Context(), id, ep, requestedServer, true)
		var rw *rawWatch
		var usedServer string
		for _, sv := range order {
			ctx, cancel := context.WithTimeout(c.Context(), 8*time.Second)
			r, err := h.fetchWatch(ctx, id, ep, sv, srcType)
			cancel()
			if err != nil || r == nil || len(r.Sources) == 0 {
				continue
			}
			abs := absolutizeSource(r.Sources[0], h.HLSProxyBase)
			pctx, pcancel := context.WithTimeout(c.Context(), 5*time.Second)
			code, _, perr := h.Client.HeadOrGet(pctx, abs)
			pcancel()
			if perr == nil && code < 400 {
				rw = r
				usedServer = sv
				if r.Server != "" {
					usedServer = r.Server
				}
				break
			}
		}
		if rw == nil {
			return fiber.NewError(http.StatusBadGateway, "no working server")
		}

		var expandedSources []rawSource
		for _, s := range rw.Sources {
			expandedSources = append(expandedSources, h.expandMasterPlaylist(c.Context(), s)...)
		}

		var watchSources []models.WatchSource
		for _, s := range expandedSources {
			abs := absolutizeSource(s, h.HLSProxyBase)
			watchSources = append(watchSources, models.WatchSource{
				Quality:   s.Quality,
				URL:       abs,
				Type:      s.Type,
				OldHLS:    s.OldHLS,
				NeedProxy: s.NeedProxy,
				ProxyURL:  selfBase + "/api/proxy/hls?url=" + queryEscape(abs),
			})
		}

		watchResp = models.WatchResponse{
			ID:         id,
			Episode:    epVal,
			Server:     "animetsu:" + usedServer,
			SourceType: srcType,
			Sources:    watchSources,
			Subtitles:  buildSubtitles(rw.Subtitles, selfBase),
		}
	}

	if len(watchResp.Sources) == 0 {
		return fiber.NewError(http.StatusBadGateway, "no sources returned")
	}

	scoreOf := func(q string) int {
		q = strings.ToLower(q)
		switch {
		case strings.Contains(q, "1080"):
			return 1080
		case strings.Contains(q, "720"):
			return 720
		case strings.Contains(q, "480"):
			return 480
		case strings.Contains(q, "360"):
			return 360
		case q == "master":
			return 9999
		}
		return 0
	}

	var best models.WatchSource
	if wantQ != "" {
		for _, s := range watchResp.Sources {
			if strings.Contains(strings.ToLower(s.Quality), wantQ) {
				best = s
				break
			}
		}
	}
	if best.URL == "" {
		best = watchResp.Sources[0]
		for _, s := range watchResp.Sources[1:] {
			if scoreOf(s.Quality) > scoreOf(best.Quality) {
				best = s
			}
		}
	}

	filename := id + "_ep" + ep + "_" + strings.ToLower(best.Quality) + "_" + srcType + ".mp4"
	hint := "HLS stream — mux to MP4 with: ffmpeg -i \"" + best.ProxyURL + "\" -c copy " + filename
	return ok(c, fiber.Map{
		"id":        id,
		"episode":   ep,
		"server":    watchResp.Server,
		"quality":   best.Quality,
		"type":      best.Type,
		"url":       best.URL,
		"proxy_url": best.ProxyURL,
		"filename":  filename,
		"subtitles": watchResp.Subtitles,
		"hint":      hint,
	})
}

// ----- utilities ------------------------------------------------------------

func normalizeSourceType(c *fiber.Ctx) string {
	v := c.Query("dub", "")
	if v == "1" || strings.EqualFold(v, "true") {
		return "dub"
	}
	st := c.Query("source_type", "sub")
	if strings.EqualFold(st, "dub") {
		return "dub"
	}
	return "sub"
}

func removeStr(slice []string, s string) []string {
	out := make([]string, 0, len(slice))
	for _, v := range slice {
		if v != s {
			out = append(out, v)
		}
	}
	return out
}

func buildServerOrder(h *H, ctx context.Context, id, ep, requested string, fallback bool) []string {
	if requested != "" && requested != "auto" && !fallback {
		return []string{requested}
	}
	servers := h.fetchServerList(ctx, id, ep)
	if len(servers) == 0 {
		servers = []string{"kite", "zoro", "pahe", "meg", "kiss", "fsoft"}
	} else {
		servers = prioritizeServers(servers)
	}
	if requested != "" && requested != "auto" {
		return append([]string{requested}, removeStr(servers, requested)...)
	}
	return servers
}

// fetchServerList returns server IDs from the upstream, default-first.
func (h *H) fetchServerList(ctx context.Context, id, ep string) []string {
	key := "serverlist:" + id + ":" + ep
	v, err := h.Cache.GetOrFetch(ctx, key, 5*time.Minute, func(ctx context.Context) (any, error) {
		var list []models.Server
		if err := h.Client.GetJSON(ctx, "/api/anime/servers/"+id+"/"+ep, &list); err != nil {
			return nil, err
		}
		// Don't cache empty results (upstream may be cold/transient. Returning
		// an error here makes GetOrFetch skip caching so the next call retries.)
		if len(list) == 0 {
			return nil, fmt.Errorf("upstream returned empty server list for %s/%s", id, ep)
		}
		out := make([]string, 0, len(list))
		for _, s := range list {
			if s.Default {
				out = append(out, s.ID)
			}
		}
		for _, s := range list {
			if !s.Default {
				out = append(out, s.ID)
			}
		}
		return prioritizeServers(out), nil
	})
	if err != nil {
		return nil
	}
	if ids, ok := v.([]string); ok {
		return ids
	}
	return nil
}

func queryEscape(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '-' || r == '_' || r == '.' || r == '~':
			b.WriteRune(r)
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
		default:
			for _, bb := range []byte(string(r)) {
				b.WriteString("%")
				const hex = "0123456789ABCDEF"
				b.WriteByte(hex[bb>>4])
				b.WriteByte(hex[bb&0x0F])
			}
		}
	}
	return b.String()
}


// buildSubtitles normalizes upstream "subs" entries into typed
// Subtitle structs. Each URL is wrapped in /api/proxy/hls so the browser
// gets permissive CORS even for .vtt files (the proxy passes non-playlist
// bytes through unchanged). The first English track (or the first track,
// if there is no English) is marked as default.
func buildSubtitles(in []rawSub, selfBase string) []models.Subtitle {
	if len(in) == 0 {
		return nil
	}
	out := make([]models.Subtitle, 0, len(in))
	defaultIdx := -1
	for i, s := range in {
		if s.URL == "" {
			continue
		}
		abs := sanitizeSubURL(s.URL)
		if !strings.HasPrefix(abs, "http://") && !strings.HasPrefix(abs, "https://") {
			continue
		}
		label := s.Lang
		if label == "" {
			label = "Subtitle"
		}
		if defaultIdx == -1 && strings.Contains(strings.ToLower(s.Lang), "english") {
			defaultIdx = i
		}
		out = append(out, models.Subtitle{
			URL:   selfBase + "/api/proxy/subtitle?url=" + queryEscape(abs),
			Lang:  s.Lang,
			Label: label,
		})
	}
	if len(out) == 0 {
		return nil
	}
	if defaultIdx < 0 || defaultIdx >= len(out) {
		defaultIdx = 0
	}
	out[defaultIdx].Default = true
	return out
}

func sanitizeSubURL(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return ""
	}
	for _, p := range []string{"https://", "http://"} {
		if strings.HasPrefix(u, p) {
			rest := strings.TrimLeft(u[len(p):], "/")
			return p + rest
		}
	}
	return u
}

var (
	hlsNameAttrRe = regexp.MustCompile(`NAME="([^"]+)"`)
	hlsResAttrRe  = regexp.MustCompile(`RESOLUTION=([0-9xX]+)`)
)

func prioritizeServers(servers []string) []string {
	preferred := []string{"kite", "zoro"}
	out := make([]string, 0, len(servers))
	for _, p := range preferred {
		for _, s := range servers {
			if s == p {
				out = append(out, s)
				break
			}
		}
	}
	for _, s := range servers {
		isPreferred := false
		for _, p := range preferred {
			if s == p {
				isPreferred = true
				break
			}
		}
		if !isPreferred {
			out = append(out, s)
		}
	}
	return out
}

func (h *H) expandMasterPlaylist(ctx context.Context, src rawSource) []rawSource {
	abs := absolutizeSource(src, h.HLSProxyBase)

	lower := strings.ToLower(abs)
	if !strings.HasSuffix(lower, ".m3u8") && !strings.Contains(lower, ".m3u8?") {
		return []rawSource{src}
	}

	referer := ""
	refVal, originVal := proxy.DetermineReferer(abs, referer)

	extra := http.Header{}
	extra.Set("Accept-Encoding", "identity")
	if refVal != "" {
		extra.Set("Referer", refVal)
	}
	if originVal != "" {
		extra.Set("Origin", originVal)
	}

	resp, err := h.Client.GetRaw(ctx, abs, extra)
	if err != nil {
		return []rawSource{src}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return []rawSource{src}
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MiB cap
	if err != nil {
		return []rawSource{src}
	}

	playlistStr := string(body)
	if !strings.Contains(playlistStr, "#EXTM3U") {
		return []rawSource{src}
	}

	if !strings.Contains(playlistStr, "#EXT-X-STREAM-INF") {
		return []rawSource{src}
	}

	var variants []rawSource
	lines := strings.Split(playlistStr, "\n")

	base, err := url.Parse(abs)
	if err != nil {
		base = nil
	}

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			continue
		}

		quality := "auto"
		if nameMatch := hlsNameAttrRe.FindStringSubmatch(line); len(nameMatch) == 2 {
			quality = strings.Trim(nameMatch[1], `"`+` `)
		} else if resMatch := hlsResAttrRe.FindStringSubmatch(line); len(resMatch) == 2 {
			resParts := strings.Split(resMatch[1], "x")
			if len(resParts) == 2 {
				quality = resParts[1] + "p"
			}
		}

		var streamURL string
		for j := i + 1; j < len(lines); j++ {
			nextLine := strings.TrimSpace(lines[j])
			if nextLine == "" {
				continue
			}
			if strings.HasPrefix(nextLine, "#") {
				break
			}
			streamURL = nextLine
			i = j
			break
		}

		if streamURL != "" {
			variantURL := hls.Absolutize(streamURL, base)
			variants = append(variants, rawSource{
				Quality:   quality,
				URL:       variantURL,
				Type:      src.Type,
				OldHLS:    true,
				NeedProxy: src.NeedProxy,
			})
		}
	}

	if len(variants) > 0 {
		return variants
	}

	return []rawSource{src}
}

func (h *H) resolveId(ctx context.Context, id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("empty id")
	}

	isHex := len(id) == 24
	if isHex {
		for _, r := range id {
			if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
				isHex = false
				break
			}
		}
	}
	if isHex {
		return id, nil
	}

	cacheKey := "mapping:" + id
	if val, ok := h.Cache.Get(cacheKey); ok {
		if mapped, okStr := val.(string); okStr {
			return mapped, nil
		}
	}

	anilistID, err := strconv.Atoi(id)
	if err != nil {
		return "", fmt.Errorf("invalid id format: %s", id)
	}

	type alResponse struct {
		Data struct {
			Media struct {
				Title struct {
					Romaji  string `json:"romaji"`
					English string `json:"english"`
				} `json:"title"`
			} `json:"media"`
		} `json:"data"`
	}

	var al alResponse
	payload := map[string]any{
		"query":     "query ($id: Int) { Media (id: $id, type: ANIME) { title { romaji english } } }",
		"variables": map[string]any{"id": anilistID},
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://graphql.anilist.co", strings.NewReader(string(reqBody)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := h.Client.HTTP().Do(req)
	if err != nil {
		return "", fmt.Errorf("anilist api call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("anilist api returned status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&al); err != nil {
		return "", fmt.Errorf("decode anilist response: %w", err)
	}

	titleRomaji := strings.TrimSpace(al.Data.Media.Title.Romaji)
	titleEnglish := strings.TrimSpace(al.Data.Media.Title.English)

	if titleRomaji == "" && titleEnglish == "" {
		return "", fmt.Errorf("no titles found on AniList for ID %d", anilistID)
	}

	titlesToSearch := []string{}
	if titleRomaji != "" {
		titlesToSearch = append(titlesToSearch, titleRomaji)
	}
	if titleEnglish != "" && titleEnglish != titleRomaji {
		titlesToSearch = append(titlesToSearch, titleEnglish)
	}

	type searchResult struct {
		ID         string `json:"id"`
		Title      struct {
			Romaji  string `json:"romaji"`
			English string `json:"english"`
		} `json:"title"`
		CoverImage struct {
			Large  string `json:"large"`
			Medium string `json:"medium"`
			Small  string `json:"small"`
		} `json:"cover_image"`
		Banner string `json:"banner"`
	}

	type searchResponse struct {
		Results []searchResult `json:"results"`
	}

	coverRegex := regexp.MustCompile(`(?:bx|b|banner/|/)([0-9]+)`)

	for _, title := range titlesToSearch {
		var sResp searchResponse
		searchPath := "/api/anime/search?query=" + url.QueryEscape(title)
		if err := h.Client.GetJSON(ctx, searchPath, &sResp); err != nil {
			continue
		}

		for _, r := range sResp.Results {
			var candidateID int
			for _, img := range []string{r.CoverImage.Large, r.CoverImage.Medium, r.CoverImage.Small, r.Banner} {
				if img == "" {
					continue
				}
				if m := coverRegex.FindStringSubmatch(img); len(m) == 2 {
					if cid, _ := strconv.Atoi(m[1]); cid > 0 {
						candidateID = cid
						break
					}
				}
			}

			if candidateID == anilistID {
				h.Cache.Set(cacheKey, r.ID, 48 * time.Hour)
				return r.ID, nil
			}
		}

		for _, r := range sResp.Results {
			type infoResponse struct {
				AniListID int `json:"anilist_id"`
			}
			var info infoResponse
			infoPath := "/api/anime/info/" + r.ID
			if err := h.Client.GetJSON(ctx, infoPath, &info); err == nil {
				if info.AniListID == anilistID {
					h.Cache.Set(cacheKey, r.ID, 48 * time.Hour)
					return r.ID, nil
				}
			}
		}
	}

	return "", fmt.Errorf("could not resolve AniList ID %d to upstream anime ID", anilistID)
}
