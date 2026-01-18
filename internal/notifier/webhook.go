package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/hadi/ssl-cert-monitor/internal/config"
)

// WebhookNotifier sends notifications to a generic webhook endpoint
type WebhookNotifier struct {
	config   config.WebhookConfig
	client   *http.Client
	template *template.Template
}

// NewWebhookNotifier creates a new webhook notifier
func NewWebhookNotifier(cfg config.WebhookConfig) (*WebhookNotifier, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("webhook URL is required")
	}
	if cfg.Method == "" {
		cfg.Method = "POST"
	}

	var tmpl *template.Template
	if cfg.BodyTemplate != "" {
		var err error
		tmpl, err = template.New("webhook").Parse(cfg.BodyTemplate)
		if err != nil {
			return nil, fmt.Errorf("failed to parse body template: %w", err)
		}
	}

	return &WebhookNotifier{
		config:   cfg,
		client:   &http.Client{Timeout: 10 * time.Second},
		template: tmpl,
	}, nil
}

// webhookData represents the data available in the template
type webhookData struct {
	Domain        string
	Host          string
	Port          int
	Name          string
	DaysRemaining float64
	Expiry        time.Time
	Threshold     int
	CheckTime     time.Time
}

// Send sends a notification to the webhook endpoint
func (w *WebhookNotifier) Send(ctx context.Context, n Notification) error {
	var body []byte
	var err error

	if w.template != nil {
		data := webhookData{
			Domain:        n.Domain.Host,
			Host:          n.Domain.Host,
			Port:          n.Domain.Port,
			Name:          n.Domain.Name,
			DaysRemaining: n.DaysRemaining,
			Expiry:        n.Expiry,
			Threshold:     n.Threshold,
			CheckTime:     time.Now(),
		}

		var buf bytes.Buffer
		if err := w.template.Execute(&buf, data); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}
		body = buf.Bytes()
	} else {
		// Default JSON body
		defaultBody := map[string]interface{}{
			"domain":         n.Domain.Host,
			"name":           n.Domain.Name,
			"days_remaining": n.DaysRemaining,
			"expiry":         n.Expiry.Format(time.RFC3339),
			"threshold":      n.Threshold,
			"check_time":     time.Now().Format(time.RFC3339),
			"message":        fmt.Sprintf("SSL certificate for %s expires in %.1f days", n.Domain.Host, n.DaysRemaining),
		}
		body, err = json.Marshal(defaultBody)
		if err != nil {
			return fmt.Errorf("failed to marshal default body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, w.config.Method, w.config.URL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range w.config.Headers {
		req.Header.Set(key, value)
	}

	// If no Content-Type header is set and we have a body, set a default
	if req.Header.Get("Content-Type") == "" && len(body) > 0 {
		if strings.Contains(string(body), "{") {
			req.Header.Set("Content-Type", "application/json")
		} else {
			req.Header.Set("Content-Type", "text/plain")
		}
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// Name returns the name of the notifier
func (w *WebhookNotifier) Name() string {
	return "Webhook"
}