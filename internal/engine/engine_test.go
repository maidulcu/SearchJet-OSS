package engine_test

import (
	"context"
	"testing"

	"github.com/uae-search-oss/uae-search-oss/internal/engine"
)

type mockBackend struct {
	docs []engine.Document
}

func (m *mockBackend) Search(ctx context.Context, q engine.Query) (engine.Result, error) {
	hits := make([]engine.Hit, 0)
	for _, doc := range m.docs {
		hits = append(hits, engine.Hit{Document: doc, Score: 1.0})
	}
	return engine.Result{
		Hits:  hits,
		Total: len(hits),
	}, nil
}

func (m *mockBackend) Index(ctx context.Context, doc engine.Document) error {
	m.docs = append(m.docs, doc)
	return nil
}

func (m *mockBackend) BulkIndex(ctx context.Context, docs []engine.Document) error {
	m.docs = append(m.docs, docs...)
	return nil
}

func (m *mockBackend) Delete(ctx context.Context, id string) error {
	result := m.docs[:0]
	for _, doc := range m.docs {
		if doc.ID != id {
			result = append(result, doc)
		}
	}
	m.docs = result
	return nil
}

func (m *mockBackend) Health(ctx context.Context) error {
	return nil
}

func TestQueryStructure(t *testing.T) {
	q := engine.Query{
		Text:    "مطعم دبي",
		Lang:    "ar",
		Emirate: "dubai",
		Page:    0,
		Limit:   20,
	}

	if q.Text == "" {
		t.Error("expected non-empty query text")
	}
	if q.Limit != 20 {
		t.Errorf("limit = %d, want 20", q.Limit)
	}
}

func TestDocumentStructure(t *testing.T) {
	doc := engine.Document{
		ID:      "doc-1",
		Title:   "Restaurant in Dubai",
		Body:    "Great food",
		Lang:    "en",
		Emirate: "dubai",
		Lat:     25.2,
		Lng:     55.3,
		Source:  "test",
	}

	if doc.ID == "" {
		t.Error("expected non-empty document ID")
	}
	if doc.Emirate != "dubai" {
		t.Errorf("emirate = %q, want dubai", doc.Emirate)
	}
}

func TestMockBackend(t *testing.T) {
	ctx := context.Background()
	backend := &mockBackend{}

	doc := engine.Document{
		ID:     "test-1",
		Title:  "Test Document",
		Body:   "Test body",
		Source: "test",
	}

	if err := backend.Index(ctx, doc); err != nil {
		t.Fatalf("index failed: %v", err)
	}

	result, err := backend.Search(ctx, engine.Query{Text: "test", Limit: 10})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("total = %d, want 1", result.Total)
	}

	if err := backend.Delete(ctx, "test-1"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	result, _ = backend.Search(ctx, engine.Query{Text: "test", Limit: 10})
	if result.Total != 0 {
		t.Errorf("after delete, total = %d, want 0", result.Total)
	}
}
