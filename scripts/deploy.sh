#!/bin/bash
set -e

# Go Web Server Ubuntu Deployment Script
# This script sets up the Go web server on Ubuntu with systemd

echo "Go Web Server Ubuntu Deployment"
echo "==============================="

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo "This script must be run as root (use sudo)" 
   exit 1
fi

# Variables
APP_NAME="gowebserver"
APP_USER="gowebserver"
APP_GROUP="gowebserver"
APP_DIR="/opt/gowebserver"
LOG_DIR="/var/log/gowebserver"
SERVICE_FILE="/etc/systemd/system/gowebserver.service"

# Create application user if it doesn't exist
if ! id "$APP_USER" &>/dev/null; then
    echo "Creating user $APP_USER..."
    useradd --system --shell /bin/false --home "$APP_DIR" --create-home "$APP_USER"
fi

# Create application directory
echo "Setting up application directory..."
mkdir -p "$APP_DIR"/{bin,logs}
mkdir -p "$LOG_DIR"

# Set ownership and permissions
chown -R "$APP_USER:$APP_GROUP" "$APP_DIR"
chown -R "$APP_USER:$APP_GROUP" "$LOG_DIR"
chmod 755 "$APP_DIR"
chmod 755 "$LOG_DIR"

# Copy binary (assumes it's built and available)
if [ -f "./bin/server" ]; then
    echo "Copying server binary..."
    cp "./bin/server" "$APP_DIR/bin/"
    chown "$APP_USER:$APP_GROUP" "$APP_DIR/bin/server"
    chmod 755 "$APP_DIR/bin/server"
else
    echo "Warning: Server binary not found at ./bin/server"
    echo "Run 'mage build' first to create the binary"
fi

# Copy environment file if it exists
if [ -f ".env" ]; then
    echo "Copying environment configuration..."
    cp ".env" "$APP_DIR/"
    chown "$APP_USER:$APP_GROUP" "$APP_DIR/.env"
    chmod 600 "$APP_DIR/.env"
else
    echo "Warning: .env file not found"
    echo "Create .env file with database configuration"
fi

# Install systemd service
echo "Installing systemd service..."
cp "./scripts/gowebserver.service" "$SERVICE_FILE"
systemctl daemon-reload

# Enable and start service
echo "Enabling and starting service..."
systemctl enable gowebserver
systemctl restart gowebserver

# Show status
echo ""
echo "Deployment completed!"
echo ""
echo "Service status:"
systemctl status gowebserver --no-pager -l

echo ""
echo "Useful commands:"
echo "  sudo systemctl status gowebserver     # Check service status"
echo "  sudo systemctl restart gowebserver    # Restart service"
echo "  sudo systemctl stop gowebserver       # Stop service"
echo "  sudo systemctl start gowebserver      # Start service"
echo "  sudo journalctl -u gowebserver -f     # View logs"
echo "  sudo journalctl -u gowebserver --since today  # View today's logs"