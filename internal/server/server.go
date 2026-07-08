// Package server wires Fiber, middleware, and routes.
package server

import (
	"github.com/miru/api/internal/cache"
	"github.com/miru/api/internal/client"
	"github.com/miru/api/internal/config"
	"github.com/miru/api/internal/docs"
	"github.com/miru/api/internal/handlers"
	mw "github.com/miru/api/internal/middleware"
	"github.com/miru/api/internal/models"
	"github.com/miru/api/internal/proxy"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog"
)

func New(cfg *config.Config, log zerolog.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:               "miru-api",
		DisableStartupMessage: true,
		ErrorHandler:          errorHandler,
		ReadTimeout:           cfg.UpstreamTimeout * 2,
		WriteTimeout:          0,
		IdleTimeout:           60_000_000_000,
		StreamRequestBody:     true,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:  "*",
		AllowMethods:  "GET,HEAD,OPTIONS",
		AllowHeaders:  "*",
		ExposeHeaders: "Content-Length,Content-Range,Content-Type,Accept-Ranges",
	}))
	app.Use(recover.New())
	app.Use(compress.New(compress.Config{Level: compress.LevelBestSpeed,
		Next: func(c *fiber.Ctx) bool {
			return c.Path() == "/api/proxy/hls"
		},
	}))
	app.Use(etag.New())
	app.Use(mw.RequestLog(log))
	app.Use(mw.RateLimit(cfg.RateLimitRPM))
	app.Use(mw.APIKey(cfg))

	hc := client.New(cfg)
	cc := cache.New(cfg.CacheDefaultTTL)
	h := handlers.New(hc, cc, cfg.HLSProxyBase)

	// Health
	app.Get("/", h.Root)
	app.Get("/health", h.Health)
	app.Get("/healthz", h.Health)
	app.Get("/livez", h.Health)
	app.Get("/readyz", h.Health)
	app.Get("/ping", func(c *fiber.Ctx) error { return c.SendString("pong") })

	// Docs + demo player + OpenAPI
	docs.Mount(app)

	// API
	api := app.Group("/api")
	api.Get("/", h.Root)
	api.Get("/home", h.Home)
	api.Get("/trending", h.Trending)
	api.Get("/season", h.Season)
	api.Get("/popular", h.Popular)
	api.Get("/top-rated", h.TopRated)
	api.Get("/upcoming", h.Upcoming)
	api.Get("/recent", h.Recent)
	api.Get("/schedule", h.Schedule)
	api.Get("/random", h.Random)
	api.Get("/search", h.Search)
	api.Get("/anime/:id", h.Info)
	api.Get("/anime/:id/episodes", h.Episodes)
	api.Get("/anime/:id/views/:ep", h.Views)
	api.Get("/anime/:id/servers", h.Servers)        // ?ep=
	api.Get("/anime/:id/servers/:ep", h.Servers)    // RESTful
	api.Get("/anime/:id/watch/:ep", h.WatchByPath)
	api.Get("/anime/:id/download/:ep", h.Download)
	api.Get("/anime/:id/downloads/:ep", h.Downloads) // real release files (animetosho)
	api.Get("/watch", h.Watch)

	// HLS proxy + CORS preflight (OPTIONS/HEAD)
	proxy.MountPreflight(api)
	api.Get("/proxy/hls", proxy.HLSHandler(hc))
	api.Get("/proxy/subtitle", proxy.SubtitleHandler(hc))

	// Filter enums
	api.Get("/genres", func(c *fiber.Ctx) error { return c.JSON(genres) })
	api.Get("/formats", func(c *fiber.Ctx) error { return c.JSON(formats) })
	api.Get("/statuses", func(c *fiber.Ctx) error { return c.JSON(statuses) })
	api.Get("/seasons", func(c *fiber.Ctx) error { return c.JSON(seasons) })
	api.Get("/sorts", func(c *fiber.Ctx) error { return c.JSON(sorts) })

	// 404
	app.Use(func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "no such route: "+c.OriginalURL())
	})
	return app
}

func errorHandler(c *fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	msg := err.Error()
	if fe, ok := err.(*fiber.Error); ok {
		status = fe.Code
		msg = fe.Message
	}
	return c.Status(status).JSON(models.Envelope{
		Success: false,
		Error: &models.ErrorBody{Code: httpCode(status), Message: msg, Status: status},
	})
}

func httpCode(s int) string {
	switch s {
	case 400:
		return "bad_request"
	case 401:
		return "unauthorized"
	case 404:
		return "not_found"
	case 429:
		return "rate_limited"
	case 502:
		return "upstream_error"
	default:
		return "internal_error"
	}
}

var (
	genres = []string{
		"Action", "Adventure", "Comedy", "Drama", "Ecchi", "Fantasy", "Horror",
		"Mahou Shoujo", "Mecha", "Music", "Mystery", "Psychological", "Romance",
		"Sci-Fi", "Slice of Life", "Sports", "Supernatural", "Thriller",
	}
	formats  = []string{"TV", "TV_SHORT", "MOVIE", "SPECIAL", "OVA", "ONA", "MUSIC"}
	statuses = []string{"RELEASING", "FINISHED", "NOT_YET_RELEASED", "CANCELLED", "HIATUS"}
	seasons  = []string{"WINTER", "SPRING", "SUMMER", "FALL"}
	sorts    = []string{
		"POPULARITY_DESC", "SCORE_DESC", "TRENDING_DESC", "UPDATED_AT_DESC",
		"START_DATE_DESC", "FAVOURITES_DESC", "TITLE_ROMAJI",
	}
)
