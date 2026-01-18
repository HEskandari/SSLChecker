package notifier

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/hadi/ssl-cert-monitor/internal/config"
)

// EmailNotifier sends notifications via SMTP email
type EmailNotifier struct {
	config config.EmailConfig
}

// NewEmailNotifier creates a new email notifier
func NewEmailNotifier(cfg config.EmailConfig) (*EmailNotifier, error) {
	if cfg.SMTPHost == "" {
		return nil, fmt.Errorf("SMTP host is required")
	}
	if cfg.From == "" {
		return nil, fmt.Errorf("from address is required")
	}
	if cfg.To == "" {
		return nil, fmt.Errorf("to address is required")
	}
	return &EmailNotifier{
		config: cfg,
	}, nil
}

// Send sends an email notification
func (e *EmailNotifier) Send(ctx context.Context, n Notification) error {
	domainName := n.Domain.Name
	if domainName == "" {
		domainName = n.Domain.Host
	}

	subject := fmt.Sprintf("SSL Certificate Expiry Alert: %s (%.1f days remaining)", domainName, n.DaysRemaining)
	body := e.buildEmailBody(n)

	message := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"Content-Type: text/plain; charset=UTF-8\r\n"+
			"\r\n"+
			"%s",
		e.config.From,
		e.config.To,
		subject,
		body,
	))

	auth := smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.SMTPHost)
	addr := fmt.Sprintf("%s:%d", e.config.SMTPHost, e.config.SMTPPort)

	var err error
	if e.config.UseTLS {
		err = smtp.SendMail(addr, auth, e.config.From, []string{e.config.To}, message)
	} else {
		// For non-TLS connections (not recommended)
		err = smtp.SendMail(addr, auth, e.config.From, []string{e.config.To}, message)
	}

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

// Name returns the name of the notifier
func (e *EmailNotifier) Name() string {
	return "Email"
}

// buildEmailBody constructs the email body
func (e *EmailNotifier) buildEmailBody(n Notification) string {
	var sb strings.Builder

	sb.WriteString("SSL Certificate Expiry Alert\n")
	sb.WriteString("============================\n\n")
	sb.WriteString(fmt.Sprintf("Domain: %s\n", n.Domain.Name))
	sb.WriteString(fmt.Sprintf("Host: %s:%d\n", n.Domain.Host, n.Domain.Port))
	sb.WriteString(fmt.Sprintf("Days Remaining: %.1f\n", n.DaysRemaining))
	sb.WriteString(fmt.Sprintf("Expiry Date: %s\n", n.Expiry.Format("2006-01-02 15:04:05 MST")))
	sb.WriteString(fmt.Sprintf("Threshold: %d days\n", n.Threshold))
	sb.WriteString(fmt.Sprintf("Check Time: %s\n", time.Now().Format("2006-01-02 15:04:05 MST")))
	sb.WriteString("\n")
	sb.WriteString("Action Required:\n")
	if n.DaysRemaining <= 7 {
		sb.WriteString("  ⚠️  Certificate expires soon! Please renew immediately.\n")
	} else if n.DaysRemaining <= 30 {
		sb.WriteString("  ⚠️  Certificate expires within 30 days. Plan for renewal.\n")
	} else {
		sb.WriteString("  ℹ️  Certificate expiry approaching. Monitor regularly.\n")
	}
	sb.WriteString("\n")
	sb.WriteString("This is an automated notification from SSL Certificate Monitor.\n")

	return sb.String()
}