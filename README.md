# SSL Certificate Monitor

A lightweight Go-based tool for monitoring SSL/TLS certificate expiry and sending notifications via multiple channels (Slack, Email, Webhook, Discord). Designed to run as a cron job or scheduled task.

## Features

- **Multi-domain monitoring**: Monitor multiple domains with custom ports
- **Configurable thresholds**: Set reminder days (e.g., 30, 14, 7, 1 days before expiry)
- **Multiple notification channels**:
  - Slack (via webhook)
  - Email (SMTP)
  - Generic webhook (HTTP POST)
  - Discord (via webhook)
- **State management**: Avoid duplicate notifications with configurable cooldown periods
- **Certificate verification**: Optional chain verification mode
- **Structured logging**: JSON or text output with configurable levels
- **Lightweight**: Single binary, no external dependencies beyond Go standard library

## Installation

### From Source

```bash
# Clone the repository
git clone <repository-url>
cd ssl-cert-monitor

# Build the binary
go build -o ssl-cert-monitor ./cmd/ssl-cert-monitor

# Move to PATH (optional)
sudo mv ssl-cert-monitor /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/hadi/ssl-cert-monitor/cmd/ssl-cert-monitor@latest
```

## Configuration

Create a `config.yaml` file (see [`config.example.yaml`](config.example.yaml) for full example):

```yaml
domains:
  - host: example.com
    port: 443
    name: "My Website"
  - host: api.example.com
    port: 443

reminder_days:
  - 30
  - 14
  - 7
  - 1

notifications:
  slack:
    enabled: true
    webhook_url: "https://hooks.slack.com/services/XXX/YYY/ZZZ"
    channel: "#alerts"
  email:
    enabled: false
    # ... SMTP configuration
  webhook:
    enabled: false
    # ... webhook configuration
  discord:
    enabled: false
    # ... Discord configuration

state:
  file: "/var/lib/ssl-monitor/state.json"
  cooldown_hours: 24

log:
  level: "info"
```

## Usage

### Basic Monitoring

```bash
./ssl-cert-monitor --config config.yaml
```

### Certificate Verification Only

```bash
./ssl-cert-monitor --config config.yaml --verify
```

### Show Version

```bash
./ssl-cert-monitor --version
```

### Command Line Options

```
-config string     Path to configuration file (default "config.yaml")
-verify            Only verify certificate chains, don't send notifications
-version           Show version information
```

## Deployment

### Cron Job (Linux/macOS)

Add to crontab to run every 6 hours:

```bash
0 */6 * * * /path/to/ssl-cert-monitor --config /etc/ssl-monitor/config.yaml
```

### Systemd Service (Linux)

Create `/etc/systemd/system/ssl-monitor.service`:

```ini
[Unit]
Description=SSL Certificate Monitor
After=network.target

[Service]
Type=oneshot
User=ssl-monitor
Group=ssl-monitor
ExecStart=/usr/local/bin/ssl-cert-monitor --config /etc/ssl-monitor/config.yaml
WorkingDirectory=/var/lib/ssl-monitor

[Install]
WantedBy=multi-user.target
```

Create a systemd timer to run periodically:

```ini
# /etc/systemd/system/ssl-monitor.timer
[Unit]
Description=Run SSL Certificate Monitor every 6 hours

[Timer]
OnCalendar=*-*-* 0/6:00:00
Persistent=true

[Install]
WantedBy=timers.target
```

### Docker

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o ssl-cert-monitor ./cmd/ssl-cert-monitor

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/ssl-cert-monitor /usr/local/bin/
COPY config.yaml /etc/ssl-monitor/config.yaml
VOLUME /var/lib/ssl-monitor
CMD ["ssl-cert-monitor", "--config", "/etc/ssl-monitor/config.yaml"]
```

## Notification Channels

### Slack
Requires a Slack webhook URL from [Slack Incoming Webhooks](https://api.slack.com/messaging/webhooks).

### Email
Configure SMTP settings for your email provider. For Gmail, use an App Password.

### Webhook
Send HTTP POST requests to any endpoint with customizable headers and body template.

### Discord
Requires a Discord webhook URL from Discord channel settings.

## State Management

The tool maintains a state file to track when notifications were last sent for each domain and threshold. This prevents duplicate notifications within the configured cooldown period.

State file location is configurable via `state.file` in the configuration.

## Logging

Logs are output in structured format with configurable levels:
- `debug`: Detailed information for troubleshooting
- `info`: General operational information (default)
- `warn`: Warning conditions
- `error`: Error conditions

Logs can be directed to a file or stdout.

## Development

### Project Structure

```
ssl-cert-monitor/
├── cmd/ssl-cert-monitor/     # CLI entry point
├── internal/
│   ├── config/              # Configuration loading and types
│   ├── checker/             # SSL certificate checking logic
│   ├── notifier/            # Notification channel implementations
│   ├── state/               # State persistence
│   └── engine/              # Core orchestration engine
├── config.example.yaml      # Example configuration
├── go.mod                   # Go module definition
└── README.md               # This file
```

### Building and Testing

```bash
# Run tests
go test ./...

# Build binary
go build -o ssl-cert-monitor ./cmd/ssl-cert-monitor

# Run with test configuration
./ssl-cert-monitor --config test-config.yaml
```

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Support

For issues and feature requests, please use the GitHub issue tracker.