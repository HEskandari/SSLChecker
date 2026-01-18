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

// DiscordNotifier sends notifications to Discord via webhook
type DiscordNotifier struct {
	config config.DiscordConfig
	client *http.Client
}

// NewDiscordNotifier creates a new Discord notifier
func NewDiscordNotifier(cfg config.DiscordConfig) (*DiscordNotifier, error) {
	if cfg.WebhookURL == "" {
		return nil, fmt.Errorf("webhook URL is required")
	}
	return &DiscordNotifier{
		config: cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// discordMessage represents the payload sent to Discord
type discordMessage struct {
	Content   string         `json:"content,omitempty"`
	Username  string         `json:"username,omitempty"`
	AvatarURL string         `json:"avatar_url,omitempty"`
	Embeds    []discordEmbed `json:"embeds,omitempty"`
}

// discordEmbed represents a Discord embed
type discordEmbed struct {
	Title       string               `json:"title,omitempty"`
	Description string               `json:"description,omitempty"`
	Color       int                  `json:"color,omitempty"`
	Fields      []discordEmbedField  `json:"fields,omitempty"`
	Timestamp   string               `json:"timestamp,omitempty"`
	Footer      *discordEmbedFooter  `json:"footer,omitempty"`
}

// discordEmbedField represents a field within a Discord embed
type discordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// discordEmbedFooter represents the footer of a Discord embed
type discordEmbedFooter struct {
	Text string `json:"text"`
}

// Send sends a notification to Discord
func (d *DiscordNotifier) Send(ctx context.Context, n Notification) error {
	message := d.buildMessage(n)
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", d.config.WebhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
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
func (d *DiscordNotifier) Name() string {
	return "Discord"
}

// buildMessage constructs the Discord message
func (d *DiscordNotifier) buildMessage(n Notification) discordMessage {
	domainName := n.Domain.Name
	if domainName == "" {
		domainName = n.Domain.Host
	}

	// Determine color based on urgency
	color := 0x00FF00 // Green
	if n.DaysRemaining <= 7 {
		color = 0xFF0000 // Red
	} else if n.DaysRemaining <= 30 {
		color = 0xFFA500 // Orange
	}

	embed := discordEmbed{
		Title:       "⚠️ SSL Certificate Expiry Alert",
		Description: fmt.Sprintf("Certificate for **%s** is expiring soon!", domainName),
		Color:       color,
		Fields: []discordEmbedField{
			{
				Name:   "Domain",
				Value:  domainName,
				Inline: true,
			},
			{
				Name:   "Host",
				Value:  fmt.Sprintf("%s:%d", n.Domain.Host, n.Domain.Port),
				Inline: true,
			},
			{
				Name:   "Days Remaining",
				Value:  fmt.Sprintf("%.1f", n.DaysRemaining),
				Inline: true,
			},
			{
				Name:   "Expiry Date",
				Value:  n.Expiry.Format("2006-01-02 15:04:05 MST"),
				Inline: true,
			},
			{
				Name:   "Threshold",
				Value:  fmt.Sprintf("%d days", n.Threshold),
				Inline: true,
			},
			{
				Name:   "Check Time",
				Value:  time.Now().Format("2006-01-02 15:04:05 MST"),
				Inline: true,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordEmbedFooter{
			Text: "SSL Certificate Monitor",
		},
	}

	return discordMessage{
		Username:  d.config.Username,
		AvatarURL: d.config.AvatarURL,
		Embeds:    []discordEmbed{embed},
	}
}