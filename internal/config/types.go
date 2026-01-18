package config

import (
	"time"
)

// DomainConfig represents a single domain to monitor
type DomainConfig struct {
	Host                 string `yaml:"host"`
	Port                 int    `yaml:"port"`
	Name                 string `yaml:"name,omitempty"`
	InsecureSkipVerify   bool   `yaml:"insecure_skip_verify,omitempty"`
}

// SlackConfig holds Slack webhook configuration
type SlackConfig struct {
	Enabled    bool   `yaml:"enabled"`
	WebhookURL string `yaml:"webhook_url"`
	Channel    string `yaml:"channel,omitempty"`
	Username   string `yaml:"username,omitempty"`
	IconEmoji  string `yaml:"icon_emoji,omitempty"`
}

// EmailConfig holds SMTP email configuration
type EmailConfig struct {
	Enabled   bool   `yaml:"enabled"`
	SMTPHost  string `yaml:"smtp_host"`
	SMTPPort  int    `yaml:"smtp_port"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	From      string `yaml:"from"`
	To        string `yaml:"to"`
	UseTLS    bool   `yaml:"use_tls"`
}

// WebhookConfig holds generic webhook configuration
type WebhookConfig struct {
	Enabled  bool              `yaml:"enabled"`
	URL      string            `yaml:"url"`
	Method   string            `yaml:"method"`
	Headers  map[string]string `yaml:"headers"`
	BodyTemplate string        `yaml:"body_template"`
}

// DiscordConfig holds Discord webhook configuration
type DiscordConfig struct {
	Enabled    bool   `yaml:"enabled"`
	WebhookURL string `yaml:"webhook_url"`
	Username   string `yaml:"username,omitempty"`
	AvatarURL  string `yaml:"avatar_url,omitempty"`
}

// NotificationsConfig holds all notification channel configurations
type NotificationsConfig struct {
	Slack   SlackConfig   `yaml:"slack"`
	Email   EmailConfig   `yaml:"email"`
	Webhook WebhookConfig `yaml:"webhook"`
	Discord DiscordConfig `yaml:"discord"`
}

// StateConfig holds state persistence configuration
type StateConfig struct {
	File          string `yaml:"file"`
	CooldownHours int    `yaml:"cooldown_hours"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file,omitempty"`
}

// Config is the root configuration structure
type Config struct {
	Domains       []DomainConfig      `yaml:"domains"`
	ReminderDays  []int               `yaml:"reminder_days"`
	Notifications NotificationsConfig `yaml:"notifications"`
	State         StateConfig         `yaml:"state"`
	Log           LogConfig           `yaml:"log"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		ReminderDays: []int{30, 14, 7, 1},
		State: StateConfig{
			CooldownHours: 24,
		},
		Log: LogConfig{
			Level: "info",
		},
	}
}

// CheckResult holds the result of a certificate check
type CheckResult struct {
	Domain        DomainConfig
	Success       bool
	Error         error
	Expiry        time.Time
	DaysRemaining float64
}