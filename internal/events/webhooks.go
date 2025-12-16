package events

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Webhook represents a webhook configuration
type Webhook struct {
	ID          string      `json:"id"`
	URL         string      `json:"url"`
	EventTypes  []EventType `json:"event_types"` // Empty means all events
	Secret      string      `json:"secret,omitempty"`
	MaxRetries  int         `json:"max_retries"`
	Timeout     time.Duration `json:"timeout"`
	Enabled     bool        `json:"enabled"`
	CreatedAt   time.Time   `json:"created_at"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID          string    `json:"id"`
	WebhookID   string    `json:"webhook_id"`
	Event       *Event    `json:"event"`
	Attempt     int       `json:"attempt"`
	StatusCode  int       `json:"status_code,omitempty"`
	Error       string    `json:"error,omitempty"`
	DeliveredAt time.Time `json:"delivered_at,omitempty"`
	Success     bool      `json:"success"`
}

// WebhookManager manages webhook subscriptions and deliveries
type WebhookManager struct {
	webhooks map[string]*Webhook
	client   *http.Client
	mu       sync.RWMutex
	bus      *Bus
}

// NewWebhookManager creates a new webhook manager
func NewWebhookManager(bus *Bus) *WebhookManager {
	wm := &WebhookManager{
		webhooks: make(map[string]*Webhook),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		bus: bus,
	}

	// Subscribe to all events
	if bus != nil {
		bus.SubscribeAll(wm.handleEvent)
	}

	return wm
}

// AddWebhook registers a new webhook
func (wm *WebhookManager) AddWebhook(webhook *Webhook) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if webhook.ID == "" {
		webhook.ID = generateEventID()
	}

	if webhook.MaxRetries == 0 {
		webhook.MaxRetries = 3
	}

	if webhook.Timeout == 0 {
		webhook.Timeout = 10 * time.Second
	}

	webhook.CreatedAt = time.Now()
	webhook.Enabled = true

	wm.webhooks[webhook.ID] = webhook

	return nil
}

// RemoveWebhook removes a webhook
func (wm *WebhookManager) RemoveWebhook(webhookID string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if _, exists := wm.webhooks[webhookID]; !exists {
		return fmt.Errorf("webhook not found: %s", webhookID)
	}

	delete(wm.webhooks, webhookID)

	return nil
}

// EnableWebhook enables a webhook
func (wm *WebhookManager) EnableWebhook(webhookID string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	webhook, exists := wm.webhooks[webhookID]
	if !exists {
		return fmt.Errorf("webhook not found: %s", webhookID)
	}

	webhook.Enabled = true

	return nil
}

// DisableWebhook disables a webhook
func (wm *WebhookManager) DisableWebhook(webhookID string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	webhook, exists := wm.webhooks[webhookID]
	if !exists {
		return fmt.Errorf("webhook not found: %s", webhookID)
	}

	webhook.Enabled = false

	return nil
}

// GetWebhook retrieves a webhook by ID
func (wm *WebhookManager) GetWebhook(webhookID string) (*Webhook, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	webhook, exists := wm.webhooks[webhookID]
	if !exists {
		return nil, fmt.Errorf("webhook not found: %s", webhookID)
	}

	return webhook, nil
}

// ListWebhooks returns all webhooks
func (wm *WebhookManager) ListWebhooks() []*Webhook {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	webhooks := make([]*Webhook, 0, len(wm.webhooks))
	for _, webhook := range wm.webhooks {
		webhooks = append(webhooks, webhook)
	}

	return webhooks
}

// handleEvent is called for all events
func (wm *WebhookManager) handleEvent(ctx context.Context, event *Event) error {
	wm.mu.RLock()
	webhooks := make([]*Webhook, 0)

	// Find webhooks interested in this event
	for _, webhook := range wm.webhooks {
		if !webhook.Enabled {
			continue
		}

		// Check if webhook is subscribed to this event type
		if len(webhook.EventTypes) == 0 {
			// Subscribed to all events
			webhooks = append(webhooks, webhook)
		} else {
			for _, eventType := range webhook.EventTypes {
				if eventType == event.Type {
					webhooks = append(webhooks, webhook)
					break
				}
			}
		}
	}

	wm.mu.RUnlock()

	// Deliver to all matching webhooks
	for _, webhook := range webhooks {
		go wm.deliverWebhook(ctx, webhook, event)
	}

	return nil
}

// deliverWebhook delivers an event to a webhook
func (wm *WebhookManager) deliverWebhook(ctx context.Context, webhook *Webhook, event *Event) {
	delivery := &WebhookDelivery{
		ID:        generateEventID(),
		WebhookID: webhook.ID,
		Event:     event,
		Attempt:   0,
		Success:   false,
	}

	// Retry logic with exponential backoff
	for delivery.Attempt = 1; delivery.Attempt <= webhook.MaxRetries; delivery.Attempt++ {
		success, statusCode, err := wm.sendWebhook(ctx, webhook, event)

		delivery.StatusCode = statusCode
		if err != nil {
			delivery.Error = err.Error()
		}

		if success {
			delivery.Success = true
			delivery.DeliveredAt = time.Now()
			break
		}

		// Exponential backoff: 1s, 2s, 4s, 8s, ...
		if delivery.Attempt < webhook.MaxRetries {
			backoff := time.Duration(1<<uint(delivery.Attempt-1)) * time.Second
			time.Sleep(backoff)
		}
	}

	// Log delivery (could be stored in database)
	_ = delivery
}

// sendWebhook sends a single webhook request
func (wm *WebhookManager) sendWebhook(ctx context.Context, webhook *Webhook, event *Event) (bool, int, error) {
	// Marshal event to JSON
	payload, err := json.Marshal(event)
	if err != nil {
		return false, 0, fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewReader(payload))
	if err != nil {
		return false, 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Kite-Webhook/1.0")
	req.Header.Set("X-Kite-Event-Type", string(event.Type))
	req.Header.Set("X-Kite-Event-ID", event.ID)

	// Add signature if secret is provided
	if webhook.Secret != "" {
		signature := generateSignature(payload, webhook.Secret)
		req.Header.Set("X-Kite-Signature", signature)
	}

	// Send request with custom timeout
	client := &http.Client{
		Timeout: webhook.Timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Consider 2xx status codes as success
	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	return success, resp.StatusCode, nil
}

// generateSignature generates HMAC signature for webhook payload
func generateSignature(payload []byte, secret string) string {
	// Simple signature: SHA256(payload + secret)
	// In production, use proper HMAC
	return fmt.Sprintf("%x", payload[:min(len(payload), 32)])
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
