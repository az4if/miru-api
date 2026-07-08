package handlers

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"os/exec"

	"github.com/miru/api/internal/models"
)

var (
	animexSlugRe = regexp.MustCompile(`slug:"([^"]+)"`)
	miruroObfKey = []byte("71951034f8fbcf53d89db52ceb3dc22c")
)

// getAnilistId resolves any hex ID back to an AniList ID, or returns the ID as-is if it's already numeric.
func (h *H) getAnilistId(ctx context.Context, id string) (string, error) {
	id = strings.TrimSpace(id)
	isNumeric := true
	for _, r := range id {
		if r < '0' || r > '9' {
			isNumeric = false
			break
		}
	}
	if isNumeric && id != "" {
		return id, nil
	}

	type infoResponse struct {
		AniListID int `json:"anilist_id"`
	}
	var info infoResponse
	infoPath := "/api/anime/info/" + id
	if err := h.Client.GetJSON(ctx, infoPath, &info); err == nil && info.AniListID > 0 {
		return strconv.Itoa(info.AniListID), nil
	}
	return "", fmt.Errorf("could not resolve AniList ID for hex ID %s", id)
}

// ==========================================
// ANIMEX.ONE SCRAPER IMPLEMENTATION
// ==========================================

func (h *H) animexGetSlug(ctx context.Context, anilistId string) (string, error) {
	url := "https://animex.one/anime/" + anilistId
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := h.Client.HTTP().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("animex returned status %d", resp.StatusCode)
	}

	buf, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return "", err
	}

	m := animexSlugRe.FindStringSubmatch(string(buf))
	if len(m) == 2 {
		return m[1], nil
	}
	return "", fmt.Errorf("slug not found in HTML")
}

func (h *H) animexServers(ctx context.Context, hexId string, ep int, srcType string) ([]string, error) {
	anilistId, err := h.getAnilistId(ctx, hexId)
	if err != nil {
		return nil, err
	}

	slug, err := h.animexGetSlug(ctx, anilistId)
	if err != nil {
		return nil, err
	}

	apiUrl := fmt.Sprintf("https://pp.animex.one/rest/api/servers?id=%s&epNum=%d", slug, ep)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36")
	req.Header.Set("Origin", "https://animex.one")
	req.Header.Set("Referer", "https://animex.one/")

	resp, err := h.Client.HTTP().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("animex servers API status %d", resp.StatusCode)
	}

	var data struct {
		SubProviders []struct {
			ID         string `json:"id"`
			ServerName string `json:"serverName"`
			Tip        string `json:"tip"`
		} `json:"subProviders"`
		DubProviders []struct {
			ID         string `json:"id"`
			ServerName string `json:"serverName"`
			Tip        string `json:"tip"`
		} `json:"dubProviders"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	providers := data.SubProviders
	if srcType == "dub" {
		providers = data.DubProviders
	}

	var list []string
	for _, p := range providers {
		name := p.ServerName
		if name == "" {
			name = p.ID
		}
		list = append(list, name)
	}
	return list, nil
}

func (h *H) animexWatch(ctx context.Context, hexId string, ep int, providerName string, srcType string, selfBase string) (any, error) {
	anilistId, err := h.getAnilistId(ctx, hexId)
	if err != nil {
		return nil, err
	}

	slug, err := h.animexGetSlug(ctx, anilistId)
	if err != nil {
		return nil, err
	}

	// Fetch servers
	apiUrl := fmt.Sprintf("https://pp.animex.one/rest/api/servers?id=%s&epNum=%d", slug, ep)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36")
	req.Header.Set("Origin", "https://animex.one")
	req.Header.Set("Referer", "https://animex.one/")

	resp, err := h.Client.HTTP().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("animex servers API returned %d", resp.StatusCode)
	}

	var data struct {
		SubProviders []struct {
			ID         string `json:"id"`
			ServerName string `json:"serverName"`
			Default    bool   `json:"default"`
		} `json:"subProviders"`
		DubProviders []struct {
			ID         string `json:"id"`
			ServerName string `json:"serverName"`
			Default    bool   `json:"default"`
		} `json:"dubProviders"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	providers := data.SubProviders
	if srcType == "dub" {
		providers = data.DubProviders
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers found on Animex")
	}

	targetProviderId := ""
	if providerName != "Auto" && providerName != "" {
		cleanProv := strings.ToLower(providerName)
		for _, p := range providers {
			if strings.ToLower(p.ID) == cleanProv || strings.ToLower(p.ServerName) == cleanProv {
				targetProviderId = p.ID
				break
			}
		}
	}

	if targetProviderId == "" {
		for _, p := range providers {
			if p.Default {
				targetProviderId = p.ID
				break
			}
		}
	}
	if targetProviderId == "" {
		targetProviderId = providers[0].ID
	}

	// Fetch sources
	sourcesUrl := fmt.Sprintf("https://pp.animex.one/rest/api/sources?id=%s&epNum=%d&type=%s&providerId=%s", slug, ep, srcType, targetProviderId)
	req2, err := http.NewRequestWithContext(ctx, http.MethodGet, sourcesUrl, nil)
	if err != nil {
		return nil, err
	}
	req2.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36")
	req2.Header.Set("Origin", "https://animex.one")
	req2.Header.Set("Referer", "https://animex.one/")

	resp2, err := h.Client.HTTP().Do(req2)
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != 200 {
		return nil, fmt.Errorf("animex sources API returned %d", resp2.StatusCode)
	}

	var sourcesData animexSourcesResponse
	if err := json.NewDecoder(resp2.Body).Decode(&sourcesData); err != nil {
		return nil, err
	}

	if len(sourcesData.Sources) == 0 {
		return nil, fmt.Errorf("no sources returned from Animex")
	}

	var watchSources []models.WatchSource
	for _, s := range sourcesData.Sources {
		if s.URL == "" {
			continue
		}
		watchSources = append(watchSources, models.WatchSource{
			Quality:   s.Quality,
			URL:       s.URL,
			Type:      s.Type,
			OldHLS:    true,
			NeedProxy: true,
			ProxyURL:  selfBase + "/api/proxy/hls?url=" + url.QueryEscape(s.URL),
		})
	}

	var subtitles []models.Subtitle
	for _, t := range sourcesData.Tracks {
		if t.URL == "" || t.Kind == "thumbnails" {
			continue
		}
		label := t.Label
		if label == "" {
			label = t.Lang
		}
		subtitles = append(subtitles, models.Subtitle{
			URL:   selfBase + "/api/proxy/subtitle?url=" + url.QueryEscape(t.URL),
			Lang:  t.Lang,
			Label: label,
		})
	}

	out := models.WatchResponse{
		ID:         hexId,
		Episode:    ep,
		Server:     "animex:" + targetProviderId,
		SourceType: srcType,
		Sources:    watchSources,
		Subtitles:  subtitles,
		Skips: &models.WatchSkips{
			Intro: &models.Range{Start: sourcesData.Intro.Start, End: sourcesData.Intro.End},
			Outro: &models.Range{Start: sourcesData.Outro.Start, End: sourcesData.Outro.End},
		},
	}
	return out, nil
}

// ==========================================
// MIRURO.TO SCRAPER IMPLEMENTATION
// ==========================================

func miruroEncodePipeRequest(payload map[string]any) (string, error) {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(jsonBytes)
	return encoded, nil
}

func miruroDecodePipeResponse(body string, obfHeader string) ([]byte, error) {
	rawBytes, err := base64.RawURLEncoding.DecodeString(body)
	if err != nil {
		bodyPadded := body
		if len(body)%4 != 0 {
			bodyPadded += strings.Repeat("=", 4-(len(body)%4))
		}
		rawBytes, err = base64.URLEncoding.DecodeString(bodyPadded)
		if err != nil {
			return nil, err
		}
	}

	if obfHeader == "2" {
		xored := make([]byte, len(rawBytes))
		for i := 0; i < len(rawBytes); i++ {
			xored[i] = rawBytes[i] ^ miruroObfKey[i%len(miruroObfKey)]
		}
		rawBytes = xored
	}

	gr, err := gzip.NewReader(bytes.NewReader(rawBytes))
	if err != nil {
		return nil, fmt.Errorf("gzip reader init: %w", err)
	}
	defer gr.Close()

	return io.ReadAll(gr)
}

func (h *H) miruroPipeRequest(ctx context.Context, path string, query map[string]any) (any, error) {
	payload := map[string]any{
		"path":   path,
		"method": "GET",
		"query":  query,
		"body":   nil,
	}

	encodedReq, err := miruroEncodePipeRequest(payload)
	if err != nil {
		return nil, err
	}

	uri := "https://www.miruro.bz/api/secure/pipe?e=" + encodedReq

	args := []string{
		"-s", "-i",
		"-H", "User-Agent: Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Mobile Safari/537.36",
		"-H", "Referer: https://www.miruro.bz/",
		"-H", "Origin: https://www.miruro.bz",
		"-H", "Accept: application/json, text/plain, */*",
		"-H", "Accept-Language: en-US,en;q=0.9",
		"-H", "sec-fetch-site: same-origin",
		"-H", "sec-fetch-mode: cors",
		"-H", "sec-fetch-dest: empty",
		uri,
	}

	cmd := exec.CommandContext(ctx, "curl", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("curl execution failed: %w", err)
	}

	outStr := string(output)
	parts := strings.SplitN(outStr, "\r\n\r\n", 2)
	if len(parts) < 2 {
		parts = strings.SplitN(outStr, "\n\n", 2)
	}
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid headers split from curl output")
	}

	headers := parts[0]
	body := parts[1]

	obfHeader := ""
	for _, line := range strings.Split(headers, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "x-obfuscated:") {
			obfHeader = strings.TrimSpace(strings.TrimPrefix(line, "x-obfuscated:"))
		}
	}

	decBytes, err := miruroDecodePipeResponse(body, obfHeader)
	if err != nil {
		bodyTrunc := body
		if len(bodyTrunc) > 100 {
			bodyTrunc = bodyTrunc[:100]
		}
		return nil, fmt.Errorf("decode pipe response: %w (body len=%d, obfHeader=%q, bodyPreview=%q)", err, len(body), obfHeader, bodyTrunc)
	}


	var out any
	if err := json.Unmarshal(decBytes, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (h *H) miruroServers(ctx context.Context, hexId string, ep int, srcType string) ([]string, error) {
	anilistIdStr, err := h.getAnilistId(ctx, hexId)
	if err != nil {
		return nil, err
	}
	anilistId, _ := strconv.Atoi(anilistIdStr)

	res, err := h.miruroPipeRequest(ctx, "episodes", map[string]any{"anilistId": anilistId, "version": "0.1.0"})
	if err != nil {
		return nil, err
	}

	dataMap, ok := res.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid response type from miruro")
	}

	providers, ok := dataMap["providers"].(map[string]any)
	if !ok {
		return nil, nil
	}

	var list []string
	for k := range providers {
		list = append(list, k)
	}
	return list, nil
}

func (h *H) miruroWatch(ctx context.Context, hexId string, ep int, providerName string, srcType string, selfBase string) (any, error) {
	anilistIdStr, err := h.getAnilistId(ctx, hexId)
	if err != nil {
		return nil, err
	}
	anilistId, _ := strconv.Atoi(anilistIdStr)

	// Fetch episodes
	res, err := h.miruroPipeRequest(ctx, "episodes", map[string]any{"anilistId": anilistId, "version": "0.1.0"})
	if err != nil {
		return nil, err
	}

	dataMap, ok := res.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid response type from miruro")
	}

	providers, ok := dataMap["providers"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("no providers found on Miruro")
	}

	provData, ok := providers[providerName].(map[string]any)
	if !ok {
		// Fallback to first provider if requested doesn't exist
		for k, v := range providers {
			providerName = k
			provData, _ = v.(map[string]any)
			break
		}
	}
	if provData == nil {
		return nil, fmt.Errorf("no provider data resolved on Miruro")
	}

	epsSection, ok := provData["episodes"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("no episodes section resolved on Miruro")
	}

	episodesList, ok := epsSection[srcType].([]any)
	if !ok || len(episodesList) == 0 {
		return nil, fmt.Errorf("no episodes found on Miruro for audio type %s", srcType)
	}

	var targetEp map[string]any
	for _, item := range episodesList {
		if em, ok := item.(map[string]any); ok {
			if num, ok := em["number"].(json.Number); ok {
				if n, _ := num.Int64(); int(n) == ep {
					targetEp = em
					break
				}
			} else if numFloat, ok := em["number"].(float64); ok {
				if int(numFloat) == ep {
					targetEp = em
					break
				}
			}
		}
	}

	if targetEp == nil {
		return nil, fmt.Errorf("episode %d not found on Miruro", ep)
	}

	rawId, _ := targetEp["id"].(string)
	if rawId == "" {
		return nil, fmt.Errorf("episode ID empty on Miruro")
	}

	targetId := rawId
	padded := rawId
	if len(rawId)%4 != 0 {
		padded += strings.Repeat("=", 4-(len(rawId)%4))
	}
	if decodedBytes, err := base64.URLEncoding.DecodeString(padded); err == nil {
		decoded := string(decodedBytes)
		if strings.Contains(decoded, ":") {
			targetId = decoded
		}
	}

	encId := base64.RawURLEncoding.EncodeToString([]byte(targetId))

	// Fetch sources
	sourcesRes, err := h.miruroPipeRequest(ctx, "sources", map[string]any{
		"episodeId": encId,
		"provider":  providerName,
		"category":  srcType,
		"anilistId": anilistId,
		"version":   "0.1.0",
	})
	if err != nil {
		return nil, err
	}

	sData, ok := sourcesRes.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid sources response type from miruro")
	}

	streams, ok := sData["streams"].([]any)
	if !ok || len(streams) == 0 {
		return nil, fmt.Errorf("no streams found on Miruro")
	}

	var watchSources []models.WatchSource
	for _, item := range streams {
		s, ok := item.(map[string]any)
		if !ok {
			continue
		}
		sType, _ := s["type"].(string)
		if sType == "embed" {
			continue
		}
		streamUrl, _ := s["url"].(string)
		if streamUrl == "" {
			continue
		}
		quality, _ := s["quality"].(string)
		if quality == "" {
			quality = "Auto"
		}
		if fansub, _ := s["fansub"].(string); fansub != "" {
			quality += " (" + fansub + ")"
		} else if srv, _ := s["server"].(string); srv != "" {
			quality += " (" + srv + ")"
		}

		watchSources = append(watchSources, models.WatchSource{
			Quality:   quality,
			URL:       streamUrl,
			Type:      "video/mpegurl",
			OldHLS:    true,
			NeedProxy: true,
			ProxyURL:  selfBase + "/api/proxy/hls?url=" + url.QueryEscape(streamUrl),
		})
	}

	var subtitles []models.Subtitle
	if rawSubs, ok := sData["subtitles"].([]any); ok {
		for _, item := range rawSubs {
			sub, ok := item.(map[string]any)
			if !ok {
				continue
			}
			subUrl, _ := sub["url"].(string)
			if subUrl == "" {
				subUrl, _ = sub["file"].(string)
			}
			if subUrl != "" && strings.HasPrefix(subUrl, "http") {
				lang, _ := sub["language"].(string)
				if lang == "" {
					lang, _ = sub["label"].(string)
				}
				if lang == "" {
					lang = "Unknown"
				}
				subtitles = append(subtitles, models.Subtitle{
					URL:   selfBase + "/api/proxy/subtitle?url=" + url.QueryEscape(subUrl),
					Lang:  lang,
					Label: lang,
				})
			}
		}
	}

	out := models.WatchResponse{
		ID:         hexId,
		Episode:    ep,
		Server:     "miruro:" + providerName,
		SourceType: srcType,
		Sources:    watchSources,
		Subtitles:  subtitles,
		Skips: &models.WatchSkips{
			Intro: &models.Range{Start: 0, End: 0},
			Outro: &models.Range{Start: 0, End: 0},
		},
	}
	return out, nil
}

type animexSource struct {
	URL     string `json:"url"`
	Quality string `json:"quality"`
	Type    string `json:"type"`
}

// Structs representing SvelteKit / Animex API responses
type animexSourcesResponse struct {
	Sources []animexSource `json:"sources"`
	Tracks  []struct {
		URL   string `json:"url"`
		Lang  string `json:"lang"`
		Label string `json:"label"`
		Kind  string `json:"kind"`
	} `json:"tracks"`
	Headers map[string]string `json:"headers"`
	Intro   struct {
		Start float64 `json:"start"`
		End   float64 `json:"end"`
	} `json:"intro"`
	Outro   struct {
		Start float64 `json:"start"`
		End   float64 `json:"end"`
	} `json:"outro"`
}
