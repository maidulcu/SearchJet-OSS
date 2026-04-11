// Package main is the entry point for SearchJet OSS server.
package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/uae-search-oss/uae-search-oss/config"
	"github.com/uae-search-oss/uae-search-oss/internal/api"
	"github.com/uae-search-oss/uae-search-oss/internal/engine/meilisearch"
	"github.com/uae-search-oss/uae-search-oss/pkg/pdpl"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		cfg = &config.Config{
			Server:      config.ServerConfig{Host: "0.0.0.0", Port: 8080},
			Meilisearch: config.MeilisearchConfig{Host: "http://localhost:7700", Index: "uae-search"},
			PDPL:        config.PDPLConfig{RetentionDays: 90, AnalyticsDays: 365},
		}
	}

	slog.Info("starting searchjet", "config", cfg)

	backend, err := meilisearch.New(
		cfg.Meilisearch.Host,
		cfg.Meilisearch.APIKey,
		cfg.Meilisearch.Index,
	)
	if err != nil {
		slog.Warn("meilisearch unavailable, running in demo mode", "error", err)
	}

	app := fiber.New(fiber.Config{
		AppName:      "SearchJet OSS",
		ServerHeader: "SearchJet",
		ReadTimeout:  10 * time.Second,
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format:     "${time} ${status} ${method} ${path} ${latency}\n",
		TimeFormat: time.RFC3339,
		TimeZone:   "Asia/Dubai",
		Output:     os.Stdout,
	}))

	handler := api.New(backend)
	handler.RegisterRoutes(app)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name":        "SearchJet OSS",
			"version":     "0.1.0",
			"description": "Bilingual search for UAE",
		})
	})

	go func() {
		addr := cfg.Server.Addr()
		slog.Info("server listening", "addr", addr)
		if err := app.Listen(addr); err != nil {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down")
	app.Shutdown()

	_ = pdpl.DefaultRetention
}
