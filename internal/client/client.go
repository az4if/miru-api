package client

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/animetsu/api/internal/config"
)

type Client struct {
	hc     *http.Client // for JSON calls (short total timeout)
	stream *http.Client // for proxy/streaming (header timeout only)
	cfg    *config.Config
	base   string
}

func New(cfg *config.Config) *Client {
	dial := (&net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 60 * time.Second,
	}).DialContext

	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dial,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          512,
		MaxIdleConnsPerHost:   128,
		IdleConnTimeout:       120 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	}
	streamTr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dial,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          512,
		MaxIdleConnsPerHost:   128,
		IdleConnTimeout:       120 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		DisableCompression: true,
	}

	return &Client{
		hc:     &http.Client{Transport: tr, Timeout: cfg.UpstreamTimeout},
		stream: &http.Client{Transport: streamTr /* no Timeout; body may stream forever */},
		cfg:    cfg,
		base:   strings.TrimRight(cfg.UpstreamBase, "/"),
	}
}

func (c *Client) HTTP() *http.Client  { return c.hc }
func (c *Client) Cfg() *config.Config { return c.cfg }

func (c *Client) GetJSON(ctx context.Context, path string, out any) error {
	full := c.base + ensureLeadingSlash(path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, full, nil)
	if err != nil {
		return err
	}
	c.applyHeaders(req)
	resp, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("upstream: %w", err)
	}
	defer resp.Body.Close()
	var reader io.Reader = resp.Body
	if strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") {
		gr, gerr := gzip.NewReader(resp.Body)
		if gerr != nil {
			return fmt.Errorf("gzip: %w", gerr)
		}
		defer gr.Close()
		reader = gr
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(reader, 512))
		return &UpstreamError{Status: resp.StatusCode, URL: full, Body: string(body)}
	}
	dec := json.NewDecoder(reader)
	dec.UseNumber()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("decode upstream JSON: %w", err)
	}
	return nil
}

// GetRaw issues a one-shot upstream GET using the short-timeout client.
// Used for HLS playlist fetches (small bodies).
func (c *Client) GetRaw(ctx context.Context, target string, extraHeaders http.Header) (*http.Response, error) {
	return c.doRaw(ctx, c.hc, target, extraHeaders)
}

// GetStream issues an upstream GET using the streaming client (no body
// timeout). Used by /api/proxy/hls for HLS segment streaming.
func (c *Client) GetStream(ctx context.Context, target string, extraHeaders http.Header) (*http.Response, error) {
	return c.doRaw(ctx, c.stream, target, extraHeaders)
}

func (c *Client) doRaw(ctx context.Context, hc *http.Client, target string, extraHeaders http.Header) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	c.applyHeaders(req)
	for k, vs := range extraHeaders {
		for _, v := range vs {
			req.Header.Set(k, v)
		}
	}
	return hc.Do(req)
}

func (c *Client) HeadOrGet(ctx context.Context, target string) (int, string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodHead, target, nil)
	c.applyHeaders(req)
	resp, err := c.hc.Do(req)
	if err == nil && resp.StatusCode < 400 {
		ct := resp.Header.Get("Content-Type")
		resp.Body.Close()
		return resp.StatusCode, ct, nil
	}
	if resp != nil {
		resp.Body.Close()
	}
	req2, _ := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	c.applyHeaders(req2)
	req2.Header.Set("Range", "bytes=0-0")
	resp2, err := c.hc.Do(req2)
	if err != nil {
		return 0, "", err
	}
	defer resp2.Body.Close()
	io.Copy(io.Discard, io.LimitReader(resp2.Body, 256))
	return resp2.StatusCode, resp2.Header.Get("Content-Type"), nil
}

func (c *Client) applyHeaders(req *http.Request) {
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "*/*")
	}
	req.Header.Set("User-Agent", c.cfg.UpstreamUserAgent)
	req.Header.Set("Referer", c.cfg.UpstreamReferer)
	req.Header.Set("Origin", strings.TrimSuffix(c.cfg.UpstreamReferer, "/"))
}

func BuildQuery(params map[string]string) string {
	v := url.Values{}
	for k, val := range params {
		v.Set(k, val)
	}
	return v.Encode()
}

func ensureLeadingSlash(s string) string {
	if strings.HasPrefix(s, "/") {
		return s
	}
	return "/" + s
}

type UpstreamError struct {
	Status int
	URL    string
	Body   string
}

func (e *UpstreamError) Error() string {
	return fmt.Sprintf("upstream %d: %s", e.Status, e.URL)
}
