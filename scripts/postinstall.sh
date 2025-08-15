#!/bin/bash
# Post-installation script for go-web-server

# Enable and start the systemd service
if command -v systemctl >/dev/null 2>&1; then
    systemctl daemon-reload
    systemctl enable gowebserver.service
fi

# Create configuration directory if it doesn't exist
mkdir -p /etc/gowebserver

echo "go-web-server installed successfully!"
echo "Configuration file: /etc/gowebserver/config.yaml"
echo "Systemd service: gowebserver.service"
echo ""
echo "To start the service:"
echo "  sudo systemctl start gowebserver"
echo ""
echo "To check the status:"
echo "  sudo systemctl status gowebserver"