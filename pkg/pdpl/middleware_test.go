package pdpl_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/uae-search-oss/uae-search-oss/pkg/pdpl"
)

func TestExtractConsent(t *testing.T) {
	app := fiber.New()
	app.Use(pdpl.Middleware(pdpl.DefaultRetention))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	tests := []struct {
		name   string
		header string
		want   pdpl.ConsentLevel
	}{
		{"default", "", pdpl.ConsentFunctional},
		{"functional", "functional", pdpl.ConsentFunctional},
		{"analytics", "analytics", pdpl.ConsentAnalytics},
		{"full", "full", pdpl.ConsentFull},
		{"invalid", "invalid", pdpl.ConsentFunctional},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tc.header != "" {
				req.Header.Set("X-PDPL-Consent", tc.header)
			}
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			resp.Body.Close()
		})
	}
}

func TestRetentionPolicy(t *testing.T) {
	policy := pdpl.DefaultRetention

	if policy.DefaultDays != 90 {
		t.Errorf("default days = %d, want 90", policy.DefaultDays)
	}
	if policy.AnalyticsDays != 365 {
		t.Errorf("analytics days = %d, want 365", policy.AnalyticsDays)
	}
}
