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

	"github.com/gofiber/fiber/v2"
)

const (
	paheUpstreamBase = "https://miru.live/v2/api/anime/dl"
	paheCacheTTL     = 30 * time.Minute
	paheTimeout      = 12 * time.Second
	paheUA           = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"
)

// paheUpstreamRow mirrors the JSON returned by miru.live's dl endpoint.
type paheUpstreamRow struct {
	Link string `json:"link"`
	Name string `json:"name"`
}

func (h *H) fetchPaheReleases(
	ctx context.Context,
	id, ep string, epNum int,
	title string,
	f dlFilter,
) (fiber.Map, error) {
	if id == "" || ep == "" {
		return nil, fmt.Errorf("pahe: id and ep required")
	}

	cacheKey := "pahe:dl:" + id + ":" + ep
	v, err := h.Cache.GetOrFetch(ctx, cacheKey, paheCacheTTL, func(ctx context.Context) (any, error) {
		return h.fetchPaheUpstream(ctx, id, ep)
	})
	if err != nil {
		return nil, err
	}
	upstream, _ := v.([]paheUpstreamRow)
	if len(upstream) == 0 {
		return nil, fmt.Errorf("pahe: no downloads returned for episode %s", ep)
	}

	rows := make([]downloadRow, 0, len(upstream))
	for _, u := range upstream {
		link := strings.TrimSpace(u.Link)
		if link == "" {
			continue
		}
		row := paheRowFromLabel(u.Name, link)
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("pahe: upstream returned only empty rows")
	}

	// Apply shared filters (quality / group / type).
	filtered := make([]downloadRow, 0, len(rows))
	for _, r := range rows {
		if f.quality != "" && !strings.Contains(strings.ToLower(r.Quality), f.quality) {
			continue
		}
		if f.group != "" && !strings.Contains(strings.ToLower(r.Group), f.group) {
			continue
		}
		if f.typ != "" && f.typ != "all" && !strings.EqualFold(r.Type, f.typ) {
			continue
		}
		filtered = append(filtered, r)
	}
	if len(filtered) == 0 {
		filtered = rows // never filter into an empty response
	}

	// Best quality first.
	sort.SliceStable(filtered, func(i, j int) bool {
		return qualityScore(filtered[i].Quality) > qualityScore(filtered[j].Quality)
	})

	// Bucket by release group.
	groupIdx := map[string]int{}
	groups := make([]downloadGroup, 0, 4)
	for _, r := range filtered {
		key := strings.ToLower(r.Group)
		if _, ok := groupIdx[key]; !ok {
			groupIdx[key] = len(groups)
			groups = append(groups, downloadGroup{Group: r.Group, Type: r.Type})
		}
		idx := groupIdx[key]
		if f.limit > 0 && len(groups[idx].Releases) >= f.limit {
			continue
		}
		groups[idx].Releases = append(groups[idx].Releases, r)
		if groups[idx].Type != r.Type && groups[idx].Type != "mixed" {
			groups[idx].Type = "mixed"
		}
	}
	for i := range groups {
		seen := map[string]struct{}{}
		for _, r := range groups[i].Releases {
			if _, ok := seen[r.Quality]; ok {
				continue
			}
			seen[r.Quality] = struct{}{}
			groups[i].Qualities = append(groups[i].Qualities, r.Quality)
		}
	}
	sort.SliceStable(groups, func(i, j int) bool {
		return groupPriority(groups[i].Group) > groupPriority(groups[j].Group)
	})

	return fiber.Map{
		"id":      id,
		"episode": ep,
		"title":   title,
		"groups":  groups,
		"flat":    filtered,
		"source":  "pahe",
		"query":   title,
		"upstream": fiber.Map{
			"endpoint": paheUpstreamBase + "/" + id + "/" + ep,
			"provider": "miru.live",
			"note":     "links are pahe.win redirects → continue to kwik to download the MP4",
		},
	}, nil
}

// fetchPaheUpstream pulls the JSON list of pahe.win links for a given AniList
// id + episode from miru.live's public download endpoint.
func (h *H) fetchPaheUpstream(ctx context.Context, id, ep string) ([]paheUpstreamRow, error) {
	ctx2, cancel := context.WithTimeout(ctx, paheTimeout)
	defer cancel()

	u := paheUpstreamBase + "/" + id + "/" + ep
	req, err := http.NewRequestWithContext(ctx2, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("pahe: build request: %w", err)
	}
	req.Header.Set("User-Agent", paheUA)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://miru.live/")
	req.Header.Set("Origin", "https://miru.live")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	resp, err := h.Client.HTTP().Do(req)
	if err != nil {
		return nil, fmt.Errorf("pahe upstream transport: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("pahe upstream 404: no downloads for ep %s", ep)
	}
	if resp.StatusCode == http.StatusBadGateway || resp.StatusCode == http.StatusGatewayTimeout {
		return nil, fmt.Errorf("pahe upstream %d: episode not yet indexed", resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("pahe upstream status %d", resp.StatusCode)
	}

	const maxBody = 1 << 20 // 1 MB ceiling: the JSON list is < 5 KB
	buf := make([]byte, 0, 16<<10)
	tmp := make([]byte, 32<<10)
	for {
		n, rerr := resp.Body.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
			if len(buf) > maxBody {
				return nil, fmt.Errorf("pahe upstream body too large")
			}
		}
		if rerr != nil {
			break
		}
	}

	// Some edge errors come back as HTML (Cloudflare interstitials, 502
	// pages). Reject anything that isn't JSON so we fall back cleanly.
	trimmed := strings.TrimSpace(string(buf))
	if trimmed == "" {
		return nil, fmt.Errorf("pahe upstream: empty body")
	}
	if trimmed[0] != '[' && trimmed[0] != '{' {
		excerpt := trimmed
		if len(excerpt) > 120 {
			excerpt = excerpt[:120] + "…"
		}
		return nil, fmt.Errorf("pahe upstream: non-JSON response (%s)", excerpt)
	}

	var rows []paheUpstreamRow
	if err := json.Unmarshal(buf, &rows); err != nil {
		// Some episodes return an object with an "error" field instead of
		// a list; treat that as "no downloads".
		return nil, fmt.Errorf("pahe upstream decode: %w", err)
	}
	return rows, nil
}

// ---- label parsing ---------------------------------------------------------

var (
	paheLabelRE = regexp.MustCompile(
		`^\s*(?P<group>[^·•|]+?)\s*[·•|]\s*(?P<res>\d{3,4})p\s*\(\s*(?P<size>[\d.]+)\s*(?P<unit>[KMGT]B)\s*\)\s*$`,
	)
	paheResOnlyRE = regexp.MustCompile(`(?i)(2160|1440|1080|720|540|480|360)p`)
	paheSizeRE    = regexp.MustCompile(`(?i)\(\s*([\d.]+)\s*([KMGT]B)\s*\)`)
)

// paheRowFromLabel turns a raw upstream row into a fully populated downloadRow.
func paheRowFromLabel(label, link string) downloadRow {
	label = strings.TrimSpace(label)
	row := downloadRow{
		Title:     label,
		Group:     "Unknown",
		Quality:   "unknown",
		Container: "mp4",
		Type:      "sub",
		Language:  "English Sub",
		ViewPage:  link, // pahe.win continue page → kwik.cx player
	}

	if m := paheLabelRE.FindStringSubmatch(label); len(m) == 5 {
		row.Group = strings.TrimSpace(m[1])
		row.Quality = m[2] + "p"
		row.SizeBytes = humanToBytes(m[3], m[4])
		row.SizeHuman = strings.TrimSpace(m[3]) + " " + strings.ToUpper(m[4])
	} else {
		// Best-effort: still pull whatever we can from a non-standard label.
		if g := parseFansubFromLabel(label); g != "" {
			row.Group = g
		}
		if mm := paheResOnlyRE.FindStringSubmatch(label); len(mm) == 2 {
			row.Quality = strings.ToLower(mm[1]) + "p"
		}
		if mm := paheSizeRE.FindStringSubmatch(label); len(mm) == 3 {
			row.SizeBytes = humanToBytes(mm[1], mm[2])
			row.SizeHuman = strings.TrimSpace(mm[1]) + " " + strings.ToUpper(mm[2])
		}
	}

	// Heuristic: a "Dub" tag in the label flips the type.
	low := strings.ToLower(label)
	if strings.Contains(low, "dub") {
		row.Type = "dub"
		row.Language = "English Dub"
	}

	return row
}

// humanToBytes converts ("154", "MB") → 154 * 1024 * 1024.
func humanToBytes(num, unit string) int64 {
	n, err := strconv.ParseFloat(strings.TrimSpace(num), 64)
	if err != nil {
		return 0
	}
	mul := int64(1)
	switch strings.ToUpper(strings.TrimSpace(unit)) {
	case "KB":
		mul = 1 << 10
	case "MB":
		mul = 1 << 20
	case "GB":
		mul = 1 << 30
	case "TB":
		mul = 1 << 40
	}
	return int64(n * float64(mul))
}

// parseFansubFromLabel pulls the group from the human label as a fallback.
// prefix is the part before "·" (e.g. "SubsPlease")
func parseFansubFromLabel(label string) string {
	if i := strings.IndexAny(label, "·•|"); i > 0 {
		return strings.TrimSpace(label[:i])
	}
	if fields := strings.Fields(label); len(fields) > 0 {
		return fields[0]
	}
	return "Unknown"
}
