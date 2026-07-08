package proxy

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/miru/api/internal/client"
	"github.com/miru/api/pkg/hls"
	"github.com/gofiber/fiber/v2"
)

func MountPreflight(app fiber.Router) {
	preflight := func(ctx *fiber.Ctx) error {
		setCORS(ctx)
		ctx.Set("Access-Control-Max-Age", "86400")
		return ctx.SendStatus(fiber.StatusNoContent)
	}
	app.Options("/proxy/hls", preflight)
	app.Head("/proxy/hls", preflight)
	app.Options("/proxy/subtitle", preflight)
	app.Head("/proxy/subtitle", preflight)
}

// SubtitleHandler
func SubtitleHandler(c *client.Client) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		target := ctx.Query("url")
		if target == "" {
			return fiber.NewError(fiber.StatusBadRequest, "missing url query parameter")
		}
		if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
			return fiber.NewError(fiber.StatusBadRequest, "url must be absolute http(s)")
		}
		referer := ctx.Query("referer")
		refVal, originVal := DetermineReferer(target, referer)

		extra := http.Header{}
		extra.Set("Accept-Encoding", "identity")
		if refVal != "" {
			extra.Set("Referer", refVal)
		} else {
			extra.Set("Referer", "")
		}
		if originVal != "" {
			extra.Set("Origin", originVal)
		} else {
			extra.Set("Origin", "")
		}
		resp, err := c.GetStream(ctx.Context(), target, extra)
		if err != nil {
			return fiber.NewError(fiber.StatusBadGateway, "upstream fetch failed: "+err.Error())
		}
		defer resp.Body.Close()
		body, rerr := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		if rerr != nil {
			return fiber.NewError(fiber.StatusBadGateway, "read upstream: "+rerr.Error())
		}
		setCORS(ctx)
		lower := strings.ToLower(target)
		ct := "text/vtt; charset=utf-8"
		if strings.HasSuffix(lower, ".srt") || strings.Contains(lower, ".srt?") {
			ct = "application/x-subrip; charset=utf-8"
		}
		ctx.Set("Content-Type", ct)
		ctx.Set("Cache-Control", "public, max-age=3600")
		return ctx.Status(resp.StatusCode).Send(body)
	}
}


// HLSHandler returns the Fiber handler for /api/proxy/hls.
func HLSHandler(c *client.Client) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		target := ctx.Query("url")
		if target == "" {
			return fiber.NewError(fiber.StatusBadRequest, "missing url query parameter")
		}
		if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
			return fiber.NewError(fiber.StatusBadRequest, "url must be absolute http(s)")
		}

		referer := ctx.Query("referer")
		refVal, originVal := DetermineReferer(target, referer)

		extra := http.Header{}
		extra.Set("Accept-Encoding", "identity")
		if refVal != "" {
			extra.Set("Referer", refVal)
		} else {
			extra.Set("Referer", "")
		}
		if originVal != "" {
			extra.Set("Origin", originVal)
		} else {
			extra.Set("Origin", "")
		}
		clientRange := ctx.Get("Range")
		if clientRange != "" {
			extra.Set("Range", clientRange)
		}

		resp, err := c.GetStream(ctx.Context(), target, extra)
		if err != nil {
			return fiber.NewError(fiber.StatusBadGateway, "upstream fetch failed: "+err.Error())
		}

		ctype := resp.Header.Get("Content-Type")
		lowerTarget := strings.ToLower(target)
		isPlaylist := strings.Contains(strings.ToLower(ctype), "mpegurl") ||
			strings.HasSuffix(lowerTarget, ".m3u8") ||
			strings.Contains(lowerTarget, ".m3u8?")

		setCORS(ctx)

		if isPlaylist {
			defer resp.Body.Close()
			body, rerr := io.ReadAll(io.LimitReader(resp.Body, 8<<20)) // 8 MiB cap
			if rerr != nil {
				return fiber.NewError(fiber.StatusBadGateway, "read upstream: "+rerr.Error())
			}
			if !hls.LooksLikePlaylist(body) {
				// Subtitle sniff: many soft-sub URLs (kite/zoro) don't end
				// in .vtt — detect by the WEBVTT magic header so the
				// browser will render them as <track> cues.
				if len(body) >= 6 && string(body[:6]) == "WEBVTT" {
					ctx.Set("Content-Type", "text/vtt; charset=utf-8")
					ctx.Set("Cache-Control", "public, max-age=3600")
				} else {
					ctx.Set("Content-Type", ctype)
				}
				return ctx.Status(resp.StatusCode).Send(body)
			}
			scheme := "https"
			if ctx.Protocol() == "http" {
				scheme = "http"
			}
			prefix := scheme + "://" + ctx.Hostname() + "/api/proxy/hls?"
			if refVal != "" {
				prefix += "referer=" + url.QueryEscape(refVal) + "&"
			}
			prefix += "url="
			rewritten := hls.Rewrite(string(body), target, prefix)
			ctx.Set("Content-Type", "application/vnd.apple.mpegurl")
			ctx.Set("Cache-Control", "public, max-age=15")
			return ctx.Status(resp.StatusCode).SendString(rewritten)
		}

		// Segment / AES key / init-segment: stream through.
		// Force the right Content-Type for HLS segments and keys.
		isSegment := false
		if strings.HasSuffix(lowerTarget, ".ts") || strings.Contains(lowerTarget, ".ts?") {
			ctx.Set("Content-Type", "video/mp2t")
			isSegment = true
		} else if strings.HasSuffix(lowerTarget, ".m4s") || strings.Contains(lowerTarget, ".m4s?") {
			ctx.Set("Content-Type", "video/iso.segment")
			isSegment = true
		} else if strings.HasSuffix(lowerTarget, ".mp4") || strings.Contains(lowerTarget, ".mp4?") {
			ctx.Set("Content-Type", "video/mp4")
			isSegment = true
		} else if strings.HasSuffix(lowerTarget, ".key") || strings.Contains(lowerTarget, ".key?") {
			ctx.Set("Content-Type", "application/octet-stream")
		} else if strings.HasSuffix(lowerTarget, ".vtt") || strings.Contains(lowerTarget, ".vtt?") {
			ctx.Set("Content-Type", "text/vtt; charset=utf-8")
			ctx.Set("Cache-Control", "public, max-age=3600")
		} else if strings.HasSuffix(lowerTarget, ".srt") || strings.Contains(lowerTarget, ".srt?") {
			ctx.Set("Content-Type", "application/x-subrip; charset=utf-8")
			ctx.Set("Cache-Control", "public, max-age=3600")
		}
		if isSegment {
			ctx.Set("Cache-Control", "public, max-age=3600, immutable")
		}
		return streamThrough(ctx, resp, clientRange != "")
	}
}

func setCORS(ctx *fiber.Ctx) {
	ctx.Set("Access-Control-Allow-Origin", "*")
	ctx.Set("Access-Control-Allow-Methods", "GET,HEAD,OPTIONS")
	ctx.Set("Access-Control-Allow-Headers", "*")
	ctx.Set("Access-Control-Expose-Headers", "Content-Length,Content-Range,Content-Type,Accept-Ranges")
}

func streamThrough(ctx *fiber.Ctx, resp *http.Response, clientRequestedRange bool) error {
	for _, h := range []string{
		"Content-Type", "Accept-Ranges", "Last-Modified", "ETag",
	} {
		if v := resp.Header.Get(h); v != "" {
			ctx.Set(h, v)
		}
	}
	// Forward Content-Range only when the client explicitly asked for a range.
	if clientRequestedRange {
		if v := resp.Header.Get("Content-Range"); v != "" {
			ctx.Set("Content-Range", v)
		}
	}

	// Advertise Range support for downstream players even if upstream forgot.
	if resp.Header.Get("Accept-Ranges") == "" {
		ctx.Set("Accept-Ranges", "bytes")
	}

	status := resp.StatusCode
	if !clientRequestedRange && status == http.StatusPartialContent {
		status = http.StatusOK
	}
	ctx.Status(status)

	contentLength := int(resp.ContentLength)
	if contentLength > 0 && (clientRequestedRange || resp.StatusCode != http.StatusPartialContent) {
		return ctx.SendStream(resp.Body, contentLength)
	}
	return ctx.SendStream(resp.Body)
}

func DetermineReferer(targetURL, requestedReferer string) (string, string) {
	if requestedReferer != "" {
		refURL, err := url.Parse(requestedReferer)
		if err == nil && refURL.Host != "" {
			origin := refURL.Scheme + "://" + refURL.Host
			return requestedReferer, origin
		}
		return requestedReferer, ""
	}

	u, err := url.Parse(targetURL)
	if err != nil {
		return "", ""
	}
	host := strings.ToLower(u.Host)

	// Domain overrides.
	if strings.Contains(host, "ultracloud.cc") || strings.Contains(host, "megacloud") {
		return "https://megacloud.tv/", ""
	}
	if strings.Contains(host, "rapid-cloud") {
		return "https://rapid-cloud.co/", ""
	}
	if strings.Contains(host, "rabbitstream") {
		return "https://rabbitstream.net/", ""
	}
	if strings.Contains(host, "aniwatchtv") {
		return "https://aniwatchtv.to/", ""
	}
	if strings.Contains(host, "hianime") {
		return "https://hianime.to/", ""
	}
	if strings.Contains(host, "miruro") {
		return "https://miruro.to/", ""
	}
	if strings.Contains(host, "animex") {
		return "https://animex.one/", ""
	}

	// General fallback: use the target's own origin as referer.
	if u.Scheme != "" && u.Host != "" {
		origin := u.Scheme + "://" + u.Host
		return origin + "/", origin
	}
	return "", ""
}
