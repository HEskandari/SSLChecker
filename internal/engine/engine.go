package engine

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hadi/ssl-cert-monitor/internal/checker"
	"github.com/hadi/ssl-cert-monitor/internal/config"
	"github.com/hadi/ssl-cert-monitor/internal/notifier"
	"github.com/hadi/ssl-cert-monitor/internal/state"
)

// Engine orchestrates the certificate checking and notification process
type Engine struct {
	config   *config.Config
	checker  *checker.Checker
	notifier *notifier.Manager
	state    *state.Manager
	logger   *slog.Logger
}

// NewEngine creates a new engine instance
func NewEngine(cfg *config.Config, logger *slog.Logger) (*Engine, error) {
	// Create state manager
	stateManager, err := state.NewManager(cfg.State.File, cfg.State.CooldownHours)
	if err != nil {
		return nil, fmt.Errorf("failed to create state manager: %w", err)
	}

	// Create notifier manager
	notifierManager, err := notifier.BuildNotifiers(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build notifiers: %w", err)
	}

	return &Engine{
		config:   cfg,
		checker:  checker.NewChecker(),
		notifier: notifierManager,
		state:    stateManager,
		logger:   logger,
	}, nil
}

// Run executes the certificate checking and notification process
func (e *Engine) Run(ctx context.Context) error {
	e.logger.Info("Starting SSL certificate monitoring", "domains", len(e.config.Domains))

	var totalChecked, totalErrors, totalNotifications int

	for _, domain := range e.config.Domains {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		domainName := domain.Name
		if domainName == "" {
			domainName = domain.Host
		}

		e.logger.Debug("Checking domain", "domain", domainName, "host", domain.Host, "port", domain.Port)
		result := e.checker.CheckDomain(domain)
		totalChecked++

		if !result.Success {
			e.logger.Warn("Failed to check domain", "domain", domainName, "error", result.Error)
			totalErrors++
			continue
		}

		e.logger.Info("Certificate check successful",
			"domain", domainName,
			"days_remaining", result.DaysRemaining,
			"expiry", result.Expiry.Format("2006-01-02"),
		)

		// Check thresholds and send notifications
		notificationsSent := e.checkThresholds(domain, result)
		totalNotifications += notificationsSent
	}

	e.logger.Info("Monitoring completed",
		"domains_checked", totalChecked,
		"errors", totalErrors,
		"notifications_sent", totalNotifications,
	)

	return nil
}

// checkThresholds evaluates certificate expiry against configured thresholds
func (e *Engine) checkThresholds(domain config.DomainConfig, result config.CheckResult) int {
	domainName := domain.Name
	if domainName == "" {
		domainName = domain.Host
	}

	notificationsSent := 0

	for _, threshold := range e.config.ReminderDays {
		// Check if days remaining is less than or equal to threshold
		if result.DaysRemaining <= float64(threshold) && result.DaysRemaining > 0 {
			// Check if we should send notification based on cooldown
			if e.state.ShouldSend(domain.Host, threshold) {
				e.logger.Info("Sending notification",
					"domain", domainName,
					"days_remaining", result.DaysRemaining,
					"threshold", threshold,
				)

				notification := notifier.Notification{
					Domain:        domain,
					DaysRemaining: result.DaysRemaining,
					Expiry:        result.Expiry,
					Threshold:     threshold,
				}

				if err := e.notifier.Send(context.Background(), notification); err != nil {
					e.logger.Error("Failed to send notification",
						"domain", domainName,
						"threshold", threshold,
						"error", err,
					)
				} else {
					// Mark as sent in state
					if err := e.state.MarkSent(domain.Host, threshold); err != nil {
						e.logger.Error("Failed to update state",
							"domain", domainName,
							"error", err,
						)
					} else {
						notificationsSent++
					}
				}
			} else {
				lastSent, _ := e.state.GetLastSent(domain.Host, threshold)
				e.logger.Debug("Skipping notification due to cooldown",
					"domain", domainName,
					"threshold", threshold,
					"last_sent", lastSent.Format(time.RFC3339),
				)
			}
		}
	}

	// Check for expired certificates (days remaining <= 0)
	if result.DaysRemaining <= 0 {
		e.logger.Error("Certificate has expired!",
			"domain", domainName,
			"days_remaining", result.DaysRemaining,
			"expiry", result.Expiry.Format("2006-01-02"),
		)
		// Could send an urgent notification here if desired
	}

	return notificationsSent
}

// VerifyAll attempts to verify certificate chains for all domains
func (e *Engine) VerifyAll() error {
	e.logger.Info("Verifying certificate chains")

	for _, domain := range e.config.Domains {
		domainName := domain.Name
		if domainName == "" {
			domainName = domain.Host
		}

		e.logger.Debug("Verifying domain", "domain", domainName)
		if err := e.checker.VerifyCertificateChain(domain); err != nil {
			e.logger.Warn("Certificate verification failed",
				"domain", domainName,
				"error", err,
			)
		} else {
			e.logger.Info("Certificate verification successful", "domain", domainName)
		}
	}

	return nil
}