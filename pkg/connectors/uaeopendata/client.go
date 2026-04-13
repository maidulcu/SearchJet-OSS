// Package uaeopendata provides a connector to UAE Open Data portals.
package uaeopendata

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/uae-search-oss/uae-search-oss/internal/engine"
)

type Config struct {
	DataGovAEURL string
	Timeout      time.Duration
}

type OpenDataRecord struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Emirate     string                 `json:"emirate"`
	SourceURL   string                 `json:"source_url"`
	LastUpdated time.Time              `json:"last_updated"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func New(cfg Config) *Client {
	if cfg.DataGovAEURL == "" {
		cfg.DataGovAEURL = "https://data.gov.ae/api/3/action"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &Client{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		baseURL:    cfg.DataGovAEURL,
	}
}

func (c *Client) FetchDatasets(ctx context.Context) ([]OpenDataRecord, error) {
	url := c.baseURL + "/package_list"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch datasets: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Success bool     `json:"success"`
		Result  []string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	records := make([]OpenDataRecord, 0, len(result.Result))
	for _, name := range result.Result {
		records = append(records, OpenDataRecord{
			ID:        name,
			Title:     name,
			SourceURL: c.baseURL + "/package_show?id=" + name,
		})
	}

	slog.Info("fetched datasets", "count", len(records))
	return records, nil
}

func (c *Client) FetchDataset(ctx context.Context, id string) (*OpenDataRecord, error) {
	url := c.baseURL + "/package_show?id=" + id
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch dataset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Success bool `json:"success"`
		Result  struct {
			ID        string                  `json:"id"`
			Title     string                  `json:"title"`
			Notes     string                  `json:"notes"`
			Tags      []struct{ Name string } `json:"tags"`
			Resources []struct {
				Name   string `json:"name"`
				Format string `json:"format"`
				URL    string `json:"url"`
			} `json:"resources"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	record := &OpenDataRecord{
		ID:          result.Result.ID,
		Title:       result.Result.Title,
		Description: result.Result.Notes,
	}

	if len(result.Result.Tags) > 0 {
		record.Category = result.Result.Tags[0].Name
	}

	if len(result.Result.Resources) > 0 {
		record.SourceURL = result.Result.Resources[0].URL
	}

	return record, nil
}

func (c *Client) IndexAll(ctx context.Context, backend engine.SearchBackend) error {
	datasets, err := c.FetchDatasets(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch datasets: %w", err)
	}

	docs := make([]engine.Document, 0, len(datasets))
	for _, ds := range datasets {
		docs = append(docs, engine.Document{
			ID:      "uae-opendata-" + ds.ID,
			Title:   ds.Title,
			Body:    ds.Description,
			Lang:    "en",
			Emirate: ds.Emirate,
			Source:  "uae-opendata",
		})
	}

	if err := backend.BulkIndex(ctx, docs); err != nil {
		return fmt.Errorf("failed to bulk index: %w", err)
	}

	slog.Info("indexed uae opendata", "count", len(docs))
	return nil
}
