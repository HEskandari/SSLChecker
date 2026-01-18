package notifier

import (
	"context"
	"fmt"
	"time"

	"github.com/hadi/ssl-cert-monitor/internal/config"
)

// Notification represents a notification to be sent
type Notification struct {
	Domain        config.DomainConfig
	DaysRemaining float64
	Expiry        time.Time
	Threshold     int
}

// Notifier defines the interface for sending notifications
type Notifier interface {
	Send(ctx context.Context, n Notification) error
	Name() string
}

// Manager coordinates multiple notifiers
type Manager struct {
	notifiers []Notifier
}

// NewManager creates a new notification manager
func NewManager(notifiers ...Notifier) *Manager {
	return &Manager{
		notifiers: notifiers,
	}
}

// Send sends a notification through all registered notifiers
func (m *Manager) Send(ctx context.Context, n Notification) error {
	var errs []error
	for _, notifier := range m.notifiers {
		if err := notifier.Send(ctx, n); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", notifier.Name(), err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to send notifications: %v", errs)
	}
	return nil
}

// BuildNotifiers creates notifiers based on configuration
func BuildNotifiers(cfg *config.Config) (*Manager, error) {
	var notifiers []Notifier

	if cfg.Notifications.Slack.Enabled {
		slack, err := NewSlackNotifier(cfg.Notifications.Slack)
		if err != nil {
			return nil, fmt.Errorf("failed to create Slack notifier: %w", err)
		}
		notifiers = append(notifiers, slack)
	}

	if cfg.Notifications.Email.Enabled {
		email, err := NewEmailNotifier(cfg.Notifications.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to create email notifier: %w", err)
		}
		notifiers = append(notifiers, email)
	}

	if cfg.Notifications.Webhook.Enabled {
		webhook, err := NewWebhookNotifier(cfg.Notifications.Webhook)
		if err != nil {
			return nil, fmt.Errorf("failed to create webhook notifier: %w", err)
		}
		notifiers = append(notifiers, webhook)
	}

	if cfg.Notifications.Discord.Enabled {
		discord, err := NewDiscordNotifier(cfg.Notifications.Discord)
		if err != nil {
			return nil, fmt.Errorf("failed to create Discord notifier: %w", err)
		}
		notifiers = append(notifiers, discord)
	}

	return NewManager(notifiers...), nil
}