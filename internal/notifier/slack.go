package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hadi/ssl-cert-monitor/internal/config"
)

// SlackNotifier sends notifications to Slack via webhook
type SlackNotifier struct {
	config config.SlackConfig
	client *http.Client
}

// NewSlackNotifier creates a new Slack notifier
func NewSlackNotifier(cfg config.SlackConfig) (*SlackNotifier, error) {
	if cfg.WebhookURL == "" {
		return nil, fmt.Errorf("webhook URL is required")
	}
	return &SlackNotifier{
		config: cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// slackMessage represents the payload sent to Slack
type slackMessage struct {
	Text      string `json:"text,omitempty"`
	Username  string `json:"username,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
	Channel   string `json:"channel,omitempty"`
}

// Send sends a notification to Slack
func (s *SlackNotifier) Send(ctx context.Context, n Notification) error {
	message := s.buildMessage(n)
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.WebhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
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
func (s *SlackNotifier) Name() string {
	return "Slack"
}

// buildMessage constructs the Slack message
func (s *SlackNotifier) buildMessage(n Notification) slackMessage {
	domainName := n.Domain.Name
	if domainName == "" {
		domainName = n.Domain.Host
	}

	message := fmt.Sprintf(
		"⚠️ SSL Certificate Expiry Alert\n"+
			"*Domain:* %s\n"+
			"*Days Remaining:* %.1f\n"+
			"*Expiry Date:* %s\n"+
			"*Threshold:* %d days\n"+
			"*Check Time:* %s",
		domainName,
		n.DaysRemaining,
		n.Expiry.Format("2006-01-02 15:04:05 MST"),
		n.Threshold,
		time.Now().Format("2006-01-02 15:04:05 MST"),
	)

	return slackMessage{
		Text:      message,
		Username:  s.config.Username,
		IconEmoji: s.config.IconEmoji,
		Channel:   s.config.Channel,
	}
}