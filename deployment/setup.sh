#!/bin/bash
# SSL Certificate Monitor Setup Script
# Run as root or with sudo

set -e

# Configuration
BINARY_PATH="/usr/local/bin/ssl-cert-monitor"
CONFIG_DIR="/etc/ssl-monitor"
STATE_DIR="/var/lib/ssl-monitor"
USER="ssl-monitor"
GROUP="ssl-monitor"

echo "Setting up SSL Certificate Monitor..."

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root or with sudo"
    exit 1
fi

# Create user and group if they don't exist
if ! id "$USER" &>/dev/null; then
    echo "Creating user $USER..."
    useradd -r -s /bin/false -d "$STATE_DIR" "$USER"
fi

# Create directories
echo "Creating directories..."
mkdir -p "$CONFIG_DIR"
mkdir -p "$STATE_DIR"

# Set permissions
echo "Setting permissions..."
chown -R "$USER:$GROUP" "$STATE_DIR"
chmod 750 "$STATE_DIR"

# Copy binary if it exists in current directory
if [ -f "./ssl-cert-monitor" ]; then
    echo "Copying binary to $BINARY_PATH..."
    cp "./ssl-cert-monitor" "$BINARY_PATH"
    chmod 755 "$BINARY_PATH"
    chown root:root "$BINARY_PATH"
else
    echo "Warning: ssl-cert-monitor binary not found in current directory"
    echo "Please build it first: go build -o ssl-cert-monitor ./cmd/ssl-cert-monitor"
fi

# Copy example configuration if config doesn't exist
if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
    if [ -f "./config.example.yaml" ]; then
        echo "Copying example configuration..."
        cp "./config.example.yaml" "$CONFIG_DIR/config.yaml"
        chmod 640 "$CONFIG_DIR/config.yaml"
        chown root:"$GROUP" "$CONFIG_DIR/config.yaml"
        echo "Please edit $CONFIG_DIR/config.yaml with your settings"
    else
        echo "Creating empty configuration file..."
        cat > "$CONFIG_DIR/config.yaml" << EOF
# SSL Certificate Monitor Configuration
# Edit with your domains and notification settings

domains:
  - host: example.com
    port: 443
    name: "Example Website"

reminder_days:
  - 30
  - 14
  - 7
  - 1

notifications:
  slack:
    enabled: false
    webhook_url: ""

state:
  file: "$STATE_DIR/state.json"
  cooldown_hours: 24

log:
  level: "info"
EOF
        chmod 640 "$CONFIG_DIR/config.yaml"
        chown root:"$GROUP" "$CONFIG_DIR/config.yaml"
    fi
fi

# Systemd setup
if [ -d "/etc/systemd/system" ]; then
    echo "Setting up systemd service..."
    
    # Copy systemd files if they exist
    if [ -f "./deployment/systemd/ssl-monitor.service" ]; then
        cp "./deployment/systemd/ssl-monitor.service" "/etc/systemd/system/"
    else
        # Create service file
        cat > "/etc/systemd/system/ssl-monitor.service" << EOF
[Unit]
Description=SSL Certificate Monitor
After=network.target

[Service]
Type=oneshot
User=$USER
Group=$GROUP
ExecStart=$BINARY_PATH --config $CONFIG_DIR/config.yaml
WorkingDirectory=$STATE_DIR
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
    fi
    
    if [ -f "./deployment/systemd/ssl-monitor.timer" ]; then
        cp "./deployment/systemd/ssl-monitor.timer" "/etc/systemd/system/"
    else
        # Create timer file
        cat > "/etc/systemd/system/ssl-monitor.timer" << EOF
[Unit]
Description=Run SSL Certificate Monitor every 6 hours
Requires=ssl-monitor.service

[Timer]
OnCalendar=*-*-* 0/6:00:00
Persistent=true

[Install]
WantedBy=timers.target
EOF
    fi
    
    echo "Reloading systemd..."
    systemctl daemon-reload
    systemctl enable ssl-monitor.timer
    systemctl start ssl-monitor.timer
    
    echo "Systemd timer enabled and started"
    echo "Check status with: systemctl status ssl-monitor.timer"
else
    echo "Systemd not found, setting up cron job..."
    
    # Add cron job
    CRON_JOB="0 */6 * * * $USER $BINARY_PATH --config $CONFIG_DIR/config.yaml"
    
    if command -v crontab &> /dev/null; then
        # Add to user's crontab
        (crontab -u "$USER" -l 2>/dev/null || true; echo "$CRON_JOB") | crontab -u "$USER" -
        echo "Cron job added for user $USER"
    else
        echo "Crontab not found. Please manually add to crontab:"
        echo "$CRON_JOB"
    fi
fi

echo ""
echo "Setup complete!"
echo ""
echo "Next steps:"
echo "1. Edit configuration: $CONFIG_DIR/config.yaml"
echo "2. Add your domains and enable notification channels"
echo "3. Test with: $BINARY_PATH --config $CONFIG_DIR/config.yaml"
echo ""
echo "For systemd systems:"
echo "  Check logs: journalctl -u ssl-monitor.service"
echo "  Manually run: systemctl start ssl-monitor.service"
echo ""
echo "For cron systems:"
echo "  Check logs in system logs or configure logging in config.yaml"