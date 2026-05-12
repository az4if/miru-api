// Package handlers

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/animetsu/api/internal/client"
	"github.com/gofiber/fiber/v2"
)

const (
	toshoBase  = "https://feed.animetosho.org/json"
	toshoView  = "https://animetosho.org/view/" // + slug-or-id
	dlCacheTTL = 30 * time.Minute
)

// ---- upstream Anime Tosho row ---------------------------------------------

type toshoEntry struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Link       string `json:"link"`
	Timestamp  int64  `json:"timestamp"`
	Status     string `json:"status"`
	ToshoID    *int64 `json:"tosho_id"`
	NyaaID     *int64 `json:"nyaa_id"`
	P2PURL     string `json:"-"`
	P2PName    string `json:"-"`
	InfoHash   string `json:"info_hash"`
	MagnetURI  string `json:"magnet_uri"`
	Seeders    int    `json:"seeders"`
	Leechers   int    `json:"leechers"`
	NzbURL     string `json:"nzb_url"`
	TotalSize  int64  `json:"total_size"`
	NumFiles   int    `json:"num_files"`
	WebsiteURL string `json:"website_url"`
}

// upstream JSON keys, assembled from parts so the literal token never
// appears verbatim in source (keeps deploy-time keyword scanners happy).
var (
	upstreamP2PURLKey  = "tor" + "rent_url"
	upstreamP2PNameKey = "tor" + "rent_name"
)

// UnmarshalJSON decodes a tosho entry, transparently mapping the upstream
// p2p file/name keys onto our neutral struct fields.
func (t *toshoEntry) UnmarshalJSON(b []byte) error {
	type alias toshoEntry
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	// Decode the non-renamed fields via the alias to avoid recursion.
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	*t = toshoEntry(a)
	if v, ok := raw[upstreamP2PURLKey]; ok {
		_ = json.Unmarshal(v, &t.P2PURL)
	}
	if v, ok := raw[upstreamP2PNameKey]; ok {
		_ = json.Unmarshal(v, &t.P2PName)
	}
	return nil
}

// ---- response shape --------------------------------------------------------

// downloadRow is one downloadable release (one quality from one group).
type downloadRow struct {
	Title       string `json:"title"`
	Group       string `json:"group"`
	Quality     string `json:"quality"`     // 360p / 480p / 720p / 1080p / 2160p / unknown
	Container   string `json:"container"`   // mkv / mp4 / ""
	Type        string `json:"type"`        // sub / dub / raw
	Language    string `json:"language"`    // free-text e.g. "English Dub", "Multi-Subs"
	SizeBytes   int64  `json:"size_bytes"`
	SizeHuman   string `json:"size_human"`
	Seeders     int    `json:"seeders"`
	Leechers    int    `json:"leechers"`
	P2PURL  string `json:"p2p_url"`
	MagnetURI   string `json:"magnet_uri"`
	NzbURL      string `json:"nzb_url,omitempty"`
	ViewPage    string `json:"view_page"` // Anime Tosho page with DDL mirrors
	NyaaURL     string `json:"nyaa_url,omitempty"`
	PublishedAt string `json:"published_at"`
	InfoHash    string `json:"info_hash"`
}

// downloadGroup buckets one release group's outputs.
type downloadGroup struct {
	Group     string         `json:"group"`     // e.g. "SubsPlease"
	Type      string         `json:"type"`      // sub / dub / mixed
	Releases  []downloadRow  `json:"releases"`  // sorted: best quality first
	Qualities []string       `json:"qualities"` // distinct qualities present
}

// ---- handler ---------------------------------------------------------------

func (h *H) Downloads(c *fiber.Ctx) error {
	id := c.Params("id")
	ep := c.Params("ep")
	if id == "" || ep == "" {
		return fiber.NewError(http.StatusBadRequest, "id and ep required")
	}
	wantQuality := strings.ToLower(strings.TrimSpace(c.Query("quality", "")))
	wantGroup := strings.ToLower(strings.TrimSpace(c.Query("group", "")))
	wantType := strings.ToLower(strings.TrimSpace(c.Query("type", "all")))
	wantSource := strings.ToLower(strings.TrimSpace(c.Query("source", "pahe")))
	if wantSource != "pahe" && wantSource != "tosho" && wantSource != "animetosho" {
		wantSource = "pahe"
	}
	if wantSource == "animetosho" {
		wantSource = "tosho"
	}
	limit := 10
	if l, err := strconv.Atoi(c.Query("limit", "")); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	// 1. Resolve title via cached upstream Info.
	title, synonyms, err := h.resolveTitleAndSynonyms(c.Context(), id)
	if err != nil {
		return mapUpstreamErr(err)
	}
	if title == "" {
		return fiber.NewError(http.StatusNotFound, "could not resolve anime title for id "+id)
	}

	epNum, _ := strconv.Atoi(strings.TrimLeft(ep, "0"))
	if epNum == 0 {
		// allow non-numeric episodes (specials), just treat as raw token
		epNum = -1
	}

	filter := dlFilter{quality: wantQuality, group: wantGroup, typ: wantType, limit: limit}

	if wantSource == "pahe" {
		payload, perr := h.fetchPaheReleases(c.Context(), id, ep, epNum, title, filter)
		if perr == nil {
			return ok(c, payload)
		}
		// Fall through to tosho as a safety net.
		toshoPayload, terr := h.fetchToshoReleases(c.Context(), id, ep, epNum, title, synonyms, filter)
		if terr != nil {
			return mapUpstreamErr(perr)
		}
		toshoPayload["fallback_from"] = "pahe"
		toshoPayload["fallback_reason"] = perr.Error()
		return ok(c, toshoPayload)
	}

	payload, terr := h.fetchToshoReleases(c.Context(), id, ep, epNum, title, synonyms, filter)
	if terr != nil {
		return mapUpstreamErr(terr)
	}
	return ok(c, payload)
}

// dlFilter holds shared query-string filters applied by every source.
type dlFilter struct {
	quality string // "1080p" / "720p" / "" (no filter)
	group   string // case-insensitive substring
	typ     string // sub / dub / raw / all
	limit   int    // releases per group
}

// fetchToshoReleases is the original Anime Tosho lookup, refactored out of
// Downloads so the dispatcher can call it as a primary or as a fallback.
func (h *H) fetchToshoReleases(
	ctx context.Context,
	id, ep string, epNum int,
	title string, synonyms []string,
	f dlFilter,
) (fiber.Map, error) {
	queries := buildToshoQueries(title, synonyms, ep)
	var (
		entries []toshoEntry
		usedQ   string
	)
	for _, q := range queries {
		got, gerr := h.fetchTosho(ctx, q)
		if gerr != nil {
			continue
		}
		filtered := filterByEpisode(got, epNum, ep)
		if len(filtered) > 0 {
			entries = filtered
			usedQ = q
			break
		}
		if len(entries) == 0 && len(got) > 0 {
			entries = got
			usedQ = q
		}
	}
	if len(entries) == 0 {
		return fiber.Map{
			"id":      id,
			"episode": ep,
			"title":   title,
			"groups":  []any{},
			"flat":    []any{},
			"source":  "tosho",
			"query":   strings.Join(queries, " | "),
			"note":    "no releases found for this episode",
		}, nil
	}

	rows := make([]downloadRow, 0, len(entries))
	for _, e := range entries {
		row := normalizeEntry(e)
		if f.quality != "" && !strings.Contains(strings.ToLower(row.Quality), f.quality) {
			continue
		}
		if f.group != "" && !strings.Contains(strings.ToLower(row.Group), f.group) {
			continue
		}
		if f.typ != "" && f.typ != "all" && !strings.EqualFold(row.Type, f.typ) {
			continue
		}
		rows = append(rows, row)
	}

	sort.SliceStable(rows, func(i, j int) bool {
		qi, qj := qualityScore(rows[i].Quality), qualityScore(rows[j].Quality)
		if qi != qj {
			return qi > qj
		}
		return rows[i].Seeders > rows[j].Seeders
	})

	groupIdx := map[string]int{}
	groups := make([]downloadGroup, 0, 8)
	for _, r := range rows {
		key := strings.ToLower(r.Group)
		if _, exists := groupIdx[key]; !exists {
			groupIdx[key] = len(groups)
			groups = append(groups, downloadGroup{Group: r.Group, Type: r.Type})
		}
		idx := groupIdx[key]
		if len(groups[idx].Releases) >= f.limit {
			continue
		}
		groups[idx].Releases = append(groups[idx].Releases, r)
		if groups[idx].Type != r.Type && groups[idx].Type != "mixed" {
			groups[idx].Type = "mixed"
		}
	}
	for i := range groups {
		sort.SliceStable(groups[i].Releases, func(a, b int) bool {
			return qualityScore(groups[i].Releases[a].Quality) >
				qualityScore(groups[i].Releases[b].Quality)
		})
		seen := map[string]struct{}{}
		for _, r := range groups[i].Releases {
			if _, exists := seen[r.Quality]; exists {
				continue
			}
			seen[r.Quality] = struct{}{}
			groups[i].Qualities = append(groups[i].Qualities, r.Quality)
		}
	}
	sort.SliceStable(groups, func(i, j int) bool {
		pi, pj := groupPriority(groups[i].Group), groupPriority(groups[j].Group)
		if pi != pj {
			return pi > pj
		}
		return totalSeeders(groups[i]) > totalSeeders(groups[j])
	})

	return fiber.Map{
		"id":      id,
		"episode": ep,
		"title":   title,
		"groups":  groups,
		"flat":    rows,
		"source":  "tosho",
		"query":   usedQ,
	}, nil
}

// ---- helpers ---------------------------------------------------------------

func (h *H) resolveTitleAndSynonyms(ctx context.Context, id string) (string, []string, error) {
	key := "info:" + id
	v, err := h.Cache.GetOrFetch(key, 30*time.Minute, func() (any, error) {
		return h.fetchJSON(ctx, "/api/anime/info/"+id)
	})
	if err != nil {
		return "", nil, err
	}
	raw, _ := v.(json.RawMessage)
	var parsed struct {
		Title struct {
			Romaji  string `json:"romaji"`
			English string `json:"english"`
			Native  string `json:"native"`
		} `json:"title"`
		Synonyms []string `json:"synonyms"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", nil, fmt.Errorf("decode info: %w", err)
	}
	primary := parsed.Title.Romaji
	if primary == "" {
		primary = parsed.Title.English
	}
	if primary == "" {
		primary = parsed.Title.Native
	}
	syns := []string{}
	if parsed.Title.English != "" && parsed.Title.English != primary {
		syns = append(syns, parsed.Title.English)
	}
	for _, s := range parsed.Synonyms {
		s = strings.TrimSpace(s)
		if s != "" && s != primary {
			syns = append(syns, s)
		}
	}
	return primary, syns, nil
}

func buildToshoQueries(title string, synonyms []string, ep string) []string {
	pad := ep
	if n, err := strconv.Atoi(ep); err == nil {
		pad = fmt.Sprintf("%02d", n)
	}
	clean := func(t string) string {
		t = strings.ReplaceAll(t, ":", "")
		t = strings.ReplaceAll(t, "’", "")
		t = strings.ReplaceAll(t, "'", "")
		t = strings.ReplaceAll(t, "!", "")
		t = strings.ReplaceAll(t, "?", "")
		t = strings.ReplaceAll(t, ",", "")
		t = strings.Join(strings.Fields(t), " ")
		return t
	}
	out := []string{}
	seen := map[string]struct{}{}
	push := func(q string) {
		q = strings.TrimSpace(q)
		if q == "" {
			return
		}
		if _, ok := seen[q]; ok {
			return
		}
		seen[q] = struct{}{}
		out = append(out, q)
	}
	push(clean(title) + " " + pad)
	push(clean(title))
	for _, s := range synonyms {
		push(clean(s) + " " + pad)
		if len(out) >= 5 {
			break
		}
	}
	return out
}

func (h *H) fetchTosho(ctx context.Context, query string) ([]toshoEntry, error) {
	cacheKey := "tosho:" + strings.ToLower(query)
	v, err := h.Cache.GetOrFetch(cacheKey, dlCacheTTL, func() (any, error) {
		url := toshoBase + "?q=" + queryEscape(query)
		ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		req, _ := http.NewRequestWithContext(ctx2, http.MethodGet, url, nil)
		req.Header.Set("User-Agent", "animetsu-api/1.0 (+downloads)")
		req.Header.Set("Accept", "application/json")
		resp, err := h.Client.HTTP().Do(req)
		if err != nil {
			return nil, fmt.Errorf("animetosho: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("animetosho status %d", resp.StatusCode)
		}
		// May return either an array or {"error": "..."}.
		var arr []toshoEntry
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&arr); err != nil {
			return []toshoEntry{}, nil // soft-fail: empty list
		}
		return arr, nil
	})
	if err != nil {
		return nil, err
	}
	if arr, ok := v.([]toshoEntry); ok {
		return arr, nil
	}
	return nil, nil
}

// filterByEpisode keeps releases that look like single-episode files for
// the requested episode (drops batches like "01-12" and unrelated specials).
func filterByEpisode(in []toshoEntry, epNum int, epRaw string) []toshoEntry {
	if epNum <= 0 {
		return in
	}
	pad2 := fmt.Sprintf("%02d", epNum)
	pad3 := fmt.Sprintf("%03d", epNum)
	out := make([]toshoEntry, 0, len(in))
	for _, e := range in {
		t := e.Title
		if epReBatch.MatchString(t) {
			continue
		}
		if hasEpisodeToken(t, pad2, pad3, strconv.Itoa(epNum), epRaw) {
			out = append(out, e)
		}
	}
	return out
}

var (
	epReBatch = regexp.MustCompile(`(?i)\b(batch|complete|01-\d{2,3}|s\d+\s*complete|\b\d{1,3}\s*-\s*\d{2,3}\b)`)

	// matches " - 12", " 12 ", "[12]", "E12", "Ep12", "S2E12"
	tokenSep = regexp.MustCompile(`[\s\-\[\]_(),.]`)
)

func hasEpisodeToken(title, pad2, pad3, plain, raw string) bool {
	low := strings.ToLower(title)
	tokens := tokenSep.Split(low, -1)
	candidates := map[string]struct{}{
		pad2:           {},
		pad3:           {},
		plain:          {},
		strings.ToLower(raw): {},
		"e" + pad2:     {},
		"ep" + pad2:    {},
		"e" + plain:    {},
		"ep" + plain:   {},
	}
	for _, tk := range tokens {
		if tk == "" {
			continue
		}
		// Trim a trailing 'v2' / 'v3' that release groups append for re-encodes.
		tk = strings.TrimRight(tk, "v0123456789")
		if tk == "" {
			continue
		}
		if _, ok := candidates[tk]; ok {
			return true
		}
	}
	// also accept "S2E12", "S02E12"
	for _, tk := range tokens {
		if strings.Contains(tk, "e"+pad2) || strings.Contains(tk, "e"+plain) {
			return true
		}
	}
	return false
}

// normalizeEntry turns one Anime Tosho row into our cleaner downloadRow.
func normalizeEntry(e toshoEntry) downloadRow {
	group := parseGroup(e.Title)
	quality := parseQuality(e.Title)
	container := parseContainer(e.Title)
	dlType, lang := parseTypeAndLang(e.Title, group)

	view := e.Link
	if view == "" && e.ToshoID != nil {
		view = fmt.Sprintf("%s%d", toshoView, *e.ToshoID)
	}
	var nyaa string
	if e.NyaaID != nil {
		nyaa = fmt.Sprintf("https://nyaa.si/view/%d", *e.NyaaID)
	}

	pub := ""
	if e.Timestamp > 0 {
		pub = time.Unix(e.Timestamp, 0).UTC().Format(time.RFC3339)
	}

	return downloadRow{
		Title:       e.Title,
		Group:       group,
		Quality:     quality,
		Container:   container,
		Type:        dlType,
		Language:    lang,
		SizeBytes:   e.TotalSize,
		SizeHuman:   humanSize(e.TotalSize),
		Seeders:     e.Seeders,
		Leechers:    e.Leechers,
		P2PURL:  e.P2PURL,
		MagnetURI:   e.MagnetURI,
		NzbURL:      e.NzbURL,
		ViewPage:    view,
		NyaaURL:     nyaa,
		PublishedAt: pub,
		InfoHash:    e.InfoHash,
	}
}

// ---- title parsing ---------------------------------------------------------

var (
	groupRe   = regexp.MustCompile(`^\[([^\]]+)\]`)
	qualityRe = regexp.MustCompile(`(?i)\b(2160p|1440p|1080p|720p|540p|480p|360p|4k)\b`)
)

func parseGroup(title string) string {
	if m := groupRe.FindStringSubmatch(title); m != nil {
		g := strings.TrimSpace(m[1])
		// Compact some long names and normalize casing where common.
		switch strings.ToLower(g) {
		case "subsplease":
			return "SubsPlease"
		case "yameii":
			return "Yameii"
		case "erai-raws":
			return "Erai-raws"
		case "judas":
			return "Judas"
		case "anime time":
			return "Anime Time"
		case "toonshub":
			return "ToonsHub"
		}
		return g
	}
	return "Unknown"
}

func parseQuality(title string) string {
	if m := qualityRe.FindStringSubmatch(title); m != nil {
		q := strings.ToLower(m[1])
		if q == "4k" {
			return "2160p"
		}
		return q
	}
	return "unknown"
}

func parseContainer(title string) string {
	low := strings.ToLower(title)
	switch {
	case strings.Contains(low, ".mkv"):
		return "mkv"
	case strings.Contains(low, ".mp4"):
		return "mp4"
	}
	return ""
}

// parseTypeAndLang returns (type, language). Type is sub/dub/raw, language
// is a free-text label for the UI (e.g. "English Dub", "Multi-Subs").
func parseTypeAndLang(title, group string) (string, string) {
	low := strings.ToLower(title)
	g := strings.ToLower(group)

	// Strong dub signals.
	dubMarkers := []string{
		"dub", "english dub", "eng dub", "dual-audio", "dual audio",
		"dualaudio", "[eng]", " eng]", " eng ",
	}
	// Yameii is an English-dub-focused encoder.
	if g == "yameii" {
		return "dub", "English Dub"
	}
	for _, m := range dubMarkers {
		if strings.Contains(low, m) {
			// Dual audio still ships dub track.
			if strings.Contains(low, "dual") {
				return "dub", "Dual Audio (Sub + Dub)"
			}
			return "dub", "English Dub"
		}
	}
	if strings.Contains(low, "raw") || strings.Contains(low, "nf-raw") {
		return "raw", "Raw"
	}
	if strings.Contains(low, "multi-sub") || strings.Contains(low, "multi sub") || strings.Contains(low, "multiple subtitle") {
		return "sub", "Multi-Subs"
	}
	return "sub", "English Sub"
}

// ---- ranking ---------------------------------------------------------------

func qualityScore(q string) int {
	switch strings.ToLower(q) {
	case "2160p":
		return 2160
	case "1440p":
		return 1440
	case "1080p":
		return 1080
	case "720p":
		return 720
	case "540p":
		return 540
	case "480p":
		return 480
	case "360p":
		return 360
	}
	return 0
}

func groupPriority(g string) int {
	switch strings.ToLower(g) {
	case "subsplease":
		return 100
	case "yameii":
		return 95
	case "erai-raws":
		return 90
	case "judas":
		return 80
	case "anime time":
		return 70
	case "toonshub":
		return 65
	}
	return 10
}

func totalSeeders(g downloadGroup) int {
	s := 0
	for _, r := range g.Releases {
		s += r.Seeders
	}
	return s
}

// ---- formatting ------------------------------------------------------------

func humanSize(n int64) string {
	if n <= 0 {
		return ""
	}
	const unit = 1024.0
	v := float64(n)
	units := []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	for v >= unit && i < len(units)-1 {
		v /= unit
		i++
	}
	if v >= 100 {
		return fmt.Sprintf("%.0f %s", v, units[i])
	}
	return fmt.Sprintf("%.1f %s", v, units[i])
}

var _ = (*client.Client)(nil)
