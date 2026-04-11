// Package engine defines the core SearchBackend interface and shared types.
// All search adapters (Meilisearch, ZincSearch, etc.) implement this interface.
package engine

import "context"

// Query represents a search request.
type Query struct {
	Text    string            // Raw query text (Arabic, English, or mixed)
	Lang    string            // "ar", "en", or "auto"
	Emirate string            // Optional: "dubai", "abudhabi", "sharjah", etc.
	Lat     float64           // Optional: latitude for geo ranking
	Lng     float64           // Optional: longitude for geo ranking
	Filters map[string]string // Facet filters (e.g. "category": "restaurant")
	Page    int
	Limit   int
}

// Document represents an indexable item.
type Document struct {
	ID       string            `json:"id"`
	Title    string            `json:"title"`
	Body     string            `json:"body"`
	Lang     string            `json:"lang"`
	Emirate  string            `json:"emirate,omitempty"`
	Lat      float64           `json:"lat,omitempty"`
	Lng      float64           `json:"lng,omitempty"`
	Source   string            `json:"source"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Hit is a single search result.
type Hit struct {
	Document
	Score float64 `json:"score"`
}

// Result is the response from a search query.
type Result struct {
	Hits       []Hit  `json:"hits"`
	Total      int    `json:"total"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	Processing int64  `json:"processing_ms"`
	QueryID    string `json:"query_id"`
}

// SearchBackend is the pluggable search engine interface.
// Implement this to add a new backend (Meilisearch, ZincSearch, Bleve, etc.)
type SearchBackend interface {
	Search(ctx context.Context, q Query) (Result, error)
	Index(ctx context.Context, doc Document) error
	BulkIndex(ctx context.Context, docs []Document) error
	Delete(ctx context.Context, id string) error
	Health(ctx context.Context) error
}
