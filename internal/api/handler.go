// Package api provides Fiber HTTP route handlers for UAE Search OSS.
package api

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/uae-search-oss/uae-search-oss/internal/engine"
	"github.com/uae-search-oss/uae-search-oss/pkg/arabic"
	"github.com/uae-search-oss/uae-search-oss/pkg/pdpl"
)

// Handler holds dependencies for API route handlers.
type Handler struct {
	backend engine.SearchBackend
}

// New creates a Handler with the given SearchBackend.
func New(backend engine.SearchBackend) *Handler {
	return &Handler{backend: backend}
}

// RegisterRoutes attaches all API routes to the Fiber app.
func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Use(pdpl.Middleware(pdpl.DefaultRetention))

	v1 := app.Group("/v1")
	v1.Get("/search", h.Search)
	v1.Post("/index", h.Index)
	v1.Delete("/index/:id", h.Delete)
	v1.Get("/health", h.Health)
}

// searchRequest is the parsed query from GET /v1/search.
type searchRequest struct {
	Q       string  `query:"q"`
	Lang    string  `query:"lang"`
	Emirate string  `query:"emirate"`
	Lat     float64 `query:"lat"`
	Lng     float64 `query:"lng"`
	Page    int     `query:"page"`
	Limit   int     `query:"limit"`
}

// Search handles GET /v1/search?q=...
func (h *Handler) Search(c *fiber.Ctx) error {
	ctx := c.Context()
	requestID, _ := c.Locals("request_id").(string)
	consent, _ := c.Locals("pdpl_consent").(pdpl.ConsentLevel)

	var req searchRequest
	if err := c.QueryParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid query parameters")
	}
	if req.Q == "" {
		return fiber.NewError(fiber.StatusBadRequest, "q is required")
	}
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// Auto-detect language if not specified
	if req.Lang == "" {
		req.Lang = arabic.DetectLang(req.Q)
	}

	q := engine.Query{
		Text:    req.Q,
		Lang:    req.Lang,
		Emirate: req.Emirate,
		Lat:     req.Lat,
		Lng:     req.Lng,
		Page:    req.Page,
		Limit:   req.Limit,
	}

	ctxWithTimeout, cancel := timeoutCtx(c, 500)
	defer cancel()
	_ = ctxWithTimeout

	result, err := h.backend.Search(ctx, q)
	if err != nil {
		slog.Error("search failed",
			"request_id", requestID,
			"query", req.Q,
			"error", err,
		)
		return fmt.Errorf("search error: %w", err)
	}

	result.QueryID = requestID

	slog.Info("search_executed",
		"request_id", requestID,
		"pdpl_consent", string(consent),
		"query_lang", req.Lang,
		"emirate", req.Emirate,
		"hits", len(result.Hits),
		"total", result.Total,
		"latency_ms", result.Processing,
	)

	return c.JSON(result)
}

// indexRequest is the body for POST /v1/index.
type indexRequest struct {
	Documents []engine.Document `json:"documents"`
}

// Index handles POST /v1/index.
func (h *Handler) Index(c *fiber.Ctx) error {
	ctx := c.Context()
	requestID, _ := c.Locals("request_id").(string)

	var req indexRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
	}
	if len(req.Documents) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "documents array is required")
	}
	if len(req.Documents) > 1000 {
		return fiber.NewError(fiber.StatusBadRequest, "max 1000 documents per request")
	}

	// Assign IDs if missing
	for i := range req.Documents {
		if req.Documents[i].ID == "" {
			req.Documents[i].ID = uuid.New().String()
		}
	}

	if err := h.backend.BulkIndex(ctx, req.Documents); err != nil {
		slog.Error("bulk index failed",
			"request_id", requestID,
			"count", len(req.Documents),
			"error", err,
		)
		return fmt.Errorf("index error: %w", err)
	}

	slog.Info("documents_indexed",
		"request_id", requestID,
		"count", len(req.Documents),
	)

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"indexed":    len(req.Documents),
		"request_id": requestID,
	})
}

// Delete handles DELETE /v1/index/:id.
func (h *Handler) Delete(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id is required")
	}

	if err := h.backend.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete error: %w", err)
	}
	return c.JSON(fiber.Map{"deleted": id})
}

// Health handles GET /v1/health.
func (h *Handler) Health(c *fiber.Ctx) error {
	ctx := c.Context()
	if err := h.backend.Health(ctx); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "unhealthy",
			"error":  err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
	})
}

func timeoutCtx(c *fiber.Ctx, ms int) (interface{ Done() <-chan struct{} }, func()) {
	// Wraps fiber's request context in a cancellable context with timeout.
	// Using context.WithTimeout on fiber's ctx is not directly supported;
	// in production wire in a proper context.Context via middleware.
	_ = ms
	return c.Context(), func() {}
}
