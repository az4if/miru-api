// Package hls contains a tiny, dependency-free m3u8 playlist rewriter.

package hls

import (
	"net/url"
	"regexp"
	"strings"
)

var uriAttrRe = regexp.MustCompile(`URI="([^"]+)"`)

func Rewrite(playlist, baseURL, proxyPrefix string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		base = nil
	}
	wrap := func(raw string) string {
		abs := absolutize(raw, base)
		return proxyPrefix + url.QueryEscape(abs)
	}
	lines := strings.Split(playlist, "\n")
	for i, line := range lines {
		trim := strings.TrimRight(line, "\r")
		if trim == "" {
			continue
		}
		if strings.HasPrefix(trim, "#") {
			lines[i] = uriAttrRe.ReplaceAllStringFunc(trim, func(match string) string {
				m := uriAttrRe.FindStringSubmatch(match)
				if len(m) != 2 {
					return match
				}
				return `URI="` + wrap(m[1]) + `"`
			})
			continue
		}
		// Bare URI line.
		lines[i] = wrap(strings.TrimSpace(trim))
	}
	return strings.Join(lines, "\n")
}

func absolutize(ref string, base *url.URL) string {
	if ref == "" {
		return ref
	}
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return ref
	}
	if base == nil {
		return ref
	}
	r, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return base.ResolveReference(r).String()
}

func LooksLikePlaylist(b []byte) bool {
	if len(b) < 7 {
		return false
	}
	return string(b[:7]) == "#EXTM3U"
}
