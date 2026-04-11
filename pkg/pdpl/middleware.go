// Package pdpl provides UAE Personal Data Protection Law (PDPL) compliance helpers.
// This includes consent logging, audit trail, and data minimization middleware.
package pdpl

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ConsentLevel represents the user's data processing consent.
type ConsentLevel string

const (
	ConsentNone      ConsentLevel = "none"
	ConsentFunctional ConsentLevel = "functional"
	ConsentAnalytics ConsentLevel = "analytics"
	ConsentFull      ConsentLevel = "full"
)

// AuditEvent is a structured PDPL audit log entry.
type AuditEvent struct {
	EventID    string       `json:"event_id"`
	Timestamp  time.Time    `json:"timestamp"`
	RequestID  string       `json:"request_id"`
	Action     string       `json:"action"`
	Consent    ConsentLevel `json:"pdpl_consent"`
	IPHash     string       `json:"ip_hash,omitempty"` // hashed, not raw IP
	UserAgent  string       `json:"user_agent,omitempty"`
	DataFields []string     `json:"data_fields,omitempty"` // fields accessed
	RetainUntil time.Time   `json:"retain_until"`
}

// RetentionPolicy defines how long audit logs are kept.
type RetentionPolicy struct {
	DefaultDays int
	AnalyticsDays int
}

var DefaultRetention = RetentionPolicy{
	DefaultDays:   90,
	AnalyticsDays: 365,
}

// Middleware returns a Fiber middleware that attaches a request ID and logs
// PDPL-compliant audit events for every search/index operation.
func Middleware(retention RetentionPolicy) fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := uuid.New().String()
		c.Locals("request_id", requestID)

		consent := extractConsent(c)
		c.Locals("pdpl_consent", consent)

		start := time.Now()
		err := c.Next()
		elapsed := time.Since(start)

		// Log audit event
		event := AuditEvent{
			EventID:     uuid.New().String(),
			Timestamp:   time.Now().UTC(),
			RequestID:   requestID,
			Action:      c.Method() + " " + c.Path(),
			Consent:     consent,
			RetainUntil: retainUntil(consent, retention),
		}

		if consent >= ConsentAnalytics {
			event.UserAgent = c.Get("User-Agent")
		}

		slog.Info("pdpl_audit",
			"event_id", event.EventID,
			"request_id", event.RequestID,
			"action", event.Action,
			"pdpl_consent", string(event.Consent),
			"latency_ms", elapsed.Milliseconds(),
			"retain_until", event.RetainUntil.Format(time.DateOnly),
			"status", c.Response().StatusCode(),
		)

		return err
	}
}

// extractConsent reads consent level from the X-PDPL-Consent header or cookie.
func extractConsent(c *fiber.Ctx) ConsentLevel {
	header := c.Get("X-PDPL-Consent")
	switch ConsentLevel(header) {
	case ConsentFunctional, ConsentAnalytics, ConsentFull:
		return ConsentLevel(header)
	}
	// Default to functional-only (minimum required for search)
	return ConsentFunctional
}

func retainUntil(consent ConsentLevel, policy RetentionPolicy) time.Time {
	days := policy.DefaultDays
	if consent == ConsentAnalytics || consent == ConsentFull {
		days = policy.AnalyticsDays
	}
	return time.Now().UTC().AddDate(0, 0, days)
}
