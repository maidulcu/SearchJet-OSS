// Package crawler provides a configurable web crawler for indexing content.
package crawler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/uae-search-oss/uae-search-oss/internal/engine"
)

type Config struct {
	AllowedDomains []string
	URLPatterns    []string
	MaxDepth       int
	Concurrency    int
	Delay          time.Duration
}

type Result struct {
	URL    string
	Title  string
	Body   string
	Source string
	Lat    float64
	Lng    float64
	Links  []string
}

func (c *Config) defaults() {
	if len(c.AllowedDomains) == 0 {
		c.AllowedDomains = []string{"*.ae"}
	}
	if c.MaxDepth == 0 {
		c.MaxDepth = 3
	}
	if c.Concurrency == 0 {
		c.Concurrency = 8
	}
	if c.Delay == 0 {
		c.Delay = 100 * time.Millisecond
	}
}

type Crawler struct {
	colly *colly.Collector
	cfg   Config
}

func New(cfg Config) *Crawler {
	cfg.defaults()
	c := &Crawler{cfg: cfg}

	c.colly = colly.NewCollector(
		colly.AllowedDomains(cfg.AllowedDomains...),
		colly.MaxDepth(cfg.MaxDepth),
		colly.Async(true),
	)

	c.colly.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       cfg.Delay,
		Parallelism: cfg.Concurrency,
	})

	return c
}

func (cr *Crawler) Run(ctx context.Context, urls []string, onResult func(Result)) error {
	cr.colly.OnHTML("html", func(e *colly.HTMLElement) {
		result := Result{
			URL:    e.Request.URL.String(),
			Title:  strings.TrimSpace(e.ChildText("title")),
			Body:   strings.TrimSpace(e.ChildText("body")),
			Source: "crawler",
		}

		e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
			if href := el.Attr("href"); href != "" {
				result.Links = append(result.Links, href)
			}
		})

		onResult(result)
	})

	cr.colly.OnError(func(r *colly.Response, err error) {
		slog.Error("crawl error",
			"url", r.Request.URL,
			"error", err,
		)
	})

	for _, url := range urls {
		if err := cr.colly.Visit(url); err != nil {
			return fmt.Errorf("visit failed for %s: %w", url, err)
		}
	}

	cr.colly.Wait()
	return nil
}

func (cr *Crawler) RunToIndex(ctx context.Context, backend engine.SearchBackend, urls []string) error {
	var indexed int
	err := cr.Run(ctx, urls, func(res Result) {
		if res.Title == "" || res.Body == "" {
			return
		}

		doc := engine.Document{
			ID:     strings.TrimPrefix(res.URL, "https://"),
			Title:  res.Title,
			Body:   res.Body[:min(len(res.Body), 5000)],
			Lang:   "en",
			Source: res.Source,
		}

		if err := backend.Index(ctx, doc); err != nil {
			slog.Warn("index failed", "url", res.URL, "error", err)
			return
		}
		indexed++
	})

	slog.Info("crawl complete", "indexed", indexed)
	return err
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
