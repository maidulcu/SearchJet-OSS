// Package uaeapimarketplace provides a connector to UAE API Marketplace.
// This enables fetching business APIs, government services, and third-party integrations.
package uaeapimarketplace

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Config struct {
	BaseURL string
	Timeout time.Duration
	APIKey  string
}

type APIListing struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Provider    string `json:"provider"`
	Endpoint    string `json:"endpoint"`
	AuthType    string `json:"auth_type"`
	Pricing     string `json:"pricing"`
	Tier        string `json:"tier"`
}

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

func New(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://uaeapi.gov.ae/api/v1"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &Client{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		baseURL:    cfg.BaseURL,
		apiKey:     cfg.APIKey,
	}
}

func (c *Client) fetch(ctx context.Context, endpoint string) ([]APIListing, error) {
	url := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Success bool         `json:"success"`
		Data    []APIListing `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode: %w", err)
	}

	return result.Data, nil
}

func (c *Client) ListAPIs(ctx context.Context, category string) ([]APIListing, error) {
	endpoint := "/listings"
	if category != "" {
		endpoint += "?category=" + category
	}

	listings, err := c.fetch(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	slog.Info("fetched API listings", "count", len(listings))
	return listings, nil
}

func (c *Client) GetAPI(ctx context.Context, id string) (*APIListing, error) {
	url := fmt.Sprintf("/listings/%s", id)
	listings, err := c.fetch(ctx, url)
	if err != nil {
		return nil, err
	}
	if len(listings) == 0 {
		return nil, fmt.Errorf("API not found: %s", id)
	}
	return &listings[0], nil
}

func (c *Client) SearchAPIs(ctx context.Context, query string) ([]APIListing, error) {
	endpoint := fmt.Sprintf("/search?q=%s", query)
	return c.fetch(ctx, endpoint)
}

type Category struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func (c *Client) ListCategories(ctx context.Context) ([]Category, error) {
	listings, err := c.fetch(ctx, "/categories")
	if err != nil {
		return nil, err
	}

	catMap := make(map[string]Category)
	for _, l := range listings {
		if c, ok := catMap[l.Category]; ok {
			c.Count++
			catMap[l.Category] = c
		} else {
			catMap[l.Category] = Category{
				ID:    l.Category,
				Name:  l.Category,
				Count: 1,
			}
		}
	}

	categories := make([]Category, 0, len(catMap))
	for _, c := range catMap {
		categories = append(categories, c)
	}
	return categories, nil
}

func (c *Client) GetFeatured() []string {
	return []string{
		"uae-pass",
		"dubai-police",
		"moei",
		"adgm",
		"sharjah-chamber",
	}
}
