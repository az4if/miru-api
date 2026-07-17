# miru-api

> [!WARNING]
> **Project Status: No Longer Maintained.** The upstream source, [Animetsu](https://animetsu.live/), is dead/scrapped. As a result, this repository will receive **no further updates**. Endpoints relying on `UPSTREAM_BASE` may no longer function correctly.

[![Go Version](https://img.shields.io/github/go-mod/go-version/miru/api?label=Go)](https://go.dev)
[![License](https://img.shields.io/github/license/miru/api?label=License)](LICENSE)

A high-performance, lightweight, and robust anime metadata and stream provider aggregator written in Go (Fiber v2). Similar in spirit to `consumet-api`, it provides unified REST endpoints for searching anime, fetching metadata, resolving stream sources, and proxying HLS streams.

> [!IMPORTANT]
> **Legal Disclaimer:** `miru-api` does not host, store, or distribute any video or media files. It purely acts as a search index and link aggregator, proxying requests to third-party public media providers on behalf of clients to bypass browser restrictions (such as CORS, Origin, and Referer validation).

---

## Key Features

- **Multi-Source Aggregation:** Combines streams and metadata from multiple providers (Zoro/Aniwatch/HiAnime via the `kite`/`zoro` server IDs, AnimePahe/kwik via the `pahe` server ID, and dynamic support for Miruro, Animex, and others).
- **Zoro/Aniwatch Prioritization:** Automatically prioritizes Zoro/Aniwatch/HiAnime servers at the front of fallback chains to ensure fast and reliable stream starts.
- **Dynamic Referer & Origin Handling:** Seamlessly resolves and applies correct hotlink bypass headers (`Referer` and `Origin`) dynamically depending on the upstream streaming host (e.g., `ultracloud.cc`, `aniwatchtv.site`, etc.).
- **Master Playlist Variant Splitting:** Automatically parses and splits multi-quality HLS master playlists (`master.m3u8`) into individual quality sources (`1080p`, `720p`, `360p`, etc.) returned in the watch payload, enabling clients to control stream quality.
- **High-Performance HLS Proxy:** A built-in streaming proxy `/api/proxy/hls` that rewrites playlist URIs, handles AES-128 key fetching, and streams video segments on-the-fly without buffering them in memory, keeping resource usage extremely low.
- **Fast and Cached:** Uses a smart in-memory SWR (stale-while-revalidate) cache to deliver sub-millisecond response times for cached pages and data.

---

## File Tree

```
miru-api/
├── cmd/server/main.go              ← Server entrypoint: load config, start Fiber
├── internal/
│   ├── config/config.go            ← Configuration loading from environment variables
│   ├── logger/logger.go            ← Structured logging (zerolog)
│   ├── client/client.go            ← Upstream HTTP client with connection pooling
│   ├── cache/cache.go              ← In-memory TTL cache + single-flight
│   ├── middleware/middleware.go    ← API key auth, rate limiter, logger
│   ├── models/models.go            ← Response DTOs and database models
│   ├── handlers/handlers.go        ← REST endpoint handlers & playlist parsing
│   ├── proxy/hls_proxy.go          ← CORS-permissive HLS stream & subtitle proxy
│   ├── server/server.go            ← Fiber server configuration and routing
│   └── docs/
│       ├── docs.go                 ← Swagger UI registration
│       └── openapi.go              ← OpenAPI 3 specification
├── pkg/hls/hls.go                  ← HLS playlist URI rewriter
├── Dockerfile                      ← Lightweight distroless container (~13 MB)
├── .env.example                    ← Configuration environment templates
├── go.mod / go.sum
└── README.md
```

---

## Quick Start

### Run from Source
Requires Go 1.22+ installed:
```bash
git clone <your-fork-url>
cd miru-api
go run ./cmd/server
```
The server will start listening on port `8080`. Open `http://localhost:8080/docs` to access the interactive Swagger documentation.

### Run with Docker
```bash
docker build -t miru-api .
docker run --rm -p 8080:8080 miru-api
```

---

## Endpoints

All JSON responses are wrapped in a standard envelope:
```json
{ "success": true, "data": ... }
```

### Discovery Endpoints
- `GET /api/home` — Get all home rails (seasonal, trending, popular, etc.).
- `GET /api/trending` — Trending anime listing.
- `GET /api/recent?page=1` — Recently updated episodes.
- `GET /api/search?q=frieren` — Search anime with rich filters.
- `GET /api/anime/:id` — Full metadata, trailer, and schedule.
- `GET /api/anime/:id/episodes` — Episode lists with thumbnails.

### Streaming Endpoints
- `GET /api/anime/:id/servers/:ep` — Lists available playback servers and tests latency.
- `GET /api/anime/:id/watch/:ep?server=auto&source_type=sub` — Resolves playable sources. Multi-quality master playlists are automatically parsed and split into individual sources (e.g., `1080p`, `720p`, `360p`) so your player can offer quality selection.
- `GET /api/proxy/hls?url=<absolute_url>&referer=<referer_url>` — Proxy HLS manifest, key, or video segments. Rewrites relative links inside the playlist, propagates referer headers, and enforces CORS headers.

---

## Configuration Options

Tunable via environment variables or a `.env` file:

| Environment Variable | Default Value | Description |
|---|---|---|
| `PORT` | `8080` | TCP port the server binds to. |
| `API_KEY` | _empty_ | If set, every request must include an `X-API-Key` header. |
| `RATE_LIMIT_RPM` | `120` | Per-IP request limit per minute. Set to `0` to disable. |
| `CACHE_DEFAULT_TTL` | `10m` | Default cache duration. |
| `UPSTREAM_TIMEOUT` | `15s` | Timeout for fetching json data from upstream. |
| `UPSTREAM_BASE` | `https://animetsu.live/v2` | Base URL of the upstream server. |
| `HLS_PROXY_BASE` | `https://swiftstream.top/proxy` | Upstream HLS segment proxy base. |
| `LOG_LEVEL` | `info` | Logging verbosity (`trace`, `debug`, `info`, `warn`, `error`). |

---

## License

This project is open-source software licensed under the [MIT License](LICENSE).
