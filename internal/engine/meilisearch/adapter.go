// Package meilisearch provides a SearchBackend adapter for Meilisearch.
package meilisearch

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"github.com/uae-search-oss/uae-search-oss/internal/engine"
)

// Adapter implements engine.SearchBackend backed by Meilisearch.
type Adapter struct {
	client    meilisearch.ServiceManager
	indexName string
}

// New creates a Meilisearch adapter.
// host example: "http://localhost:7700"
func New(host, apiKey, indexName string) (*Adapter, error) {
	client := meilisearch.New(host, meilisearch.WithAPIKey(apiKey))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = ctx

	if _, err := client.Health(); err != nil {
		return nil, fmt.Errorf("meilisearch unreachable at %s: %w", host, err)
	}

	slog.Info("meilisearch connected", "host", host, "index", indexName)
	return &Adapter{client: client, indexName: indexName}, nil
}

func (a *Adapter) index() meilisearch.IndexManager {
	return a.client.Index(a.indexName)
}

// Search executes a query against Meilisearch.
func (a *Adapter) Search(ctx context.Context, q engine.Query) (engine.Result, error) {
	start := time.Now()

	limit := int64(q.Limit)
	if limit == 0 {
		limit = 20
	}
	offset := int64(q.Page * int(limit))

	req := &meilisearch.SearchRequest{
		Limit:  limit,
		Offset: offset,
	}

	res, err := a.index().SearchWithContext(ctx, q.Text, req)
	if err != nil {
		return engine.Result{}, fmt.Errorf("meilisearch search failed: %w", err)
	}

	hits := make([]engine.Hit, 0, len(res.Hits))
	for _, raw := range res.Hits {
		m := make(map[string]any)
		for k, v := range raw {
			m[k] = v
		}
		hits = append(hits, mapToHit(m))
	}

	return engine.Result{
		Hits:       hits,
		Total:      int(res.EstimatedTotalHits),
		Page:       q.Page,
		Limit:      int(limit),
		Processing: time.Since(start).Milliseconds(),
	}, nil
}

// Index adds or updates a single document.
func (a *Adapter) Index(ctx context.Context, doc engine.Document) error {
	_, err := a.index().AddDocuments([]engine.Document{doc}, nil)
	if err != nil {
		return fmt.Errorf("meilisearch index failed for doc %s: %w", doc.ID, err)
	}
	return nil
}

// BulkIndex adds or updates multiple documents efficiently.
func (a *Adapter) BulkIndex(ctx context.Context, docs []engine.Document) error {
	_, err := a.index().AddDocuments(docs, nil)
	if err != nil {
		return fmt.Errorf("meilisearch bulk index failed (%d docs): %w", len(docs), err)
	}
	slog.Info("bulk indexed", "count", len(docs), "index", a.indexName)
	return nil
}

// Delete removes a document by ID.
func (a *Adapter) Delete(ctx context.Context, id string) error {
	_, err := a.index().DeleteDocument(id, nil)
	if err != nil {
		return fmt.Errorf("meilisearch delete failed for id %s: %w", id, err)
	}
	return nil
}

// Health checks Meilisearch availability.
func (a *Adapter) Health(_ context.Context) error {
	if _, err := a.client.Health(); err != nil {
		return fmt.Errorf("meilisearch health check failed: %w", err)
	}
	return nil
}

func mapToHit(m map[string]any) engine.Hit {
	getString := func(key string) string {
		if v, ok := m[key].(string); ok {
			return v
		}
		return ""
	}
	getFloat := func(key string) float64 {
		if v, ok := m[key].(float64); ok {
			return v
		}
		return 0
	}

	return engine.Hit{
		Document: engine.Document{
			ID:      getString("id"),
			Title:   getString("title"),
			Body:    getString("body"),
			Lang:    getString("lang"),
			Emirate: getString("emirate"),
			Source:  getString("source"),
			Lat:     getFloat("lat"),
			Lng:     getFloat("lng"),
		},
	}
}
