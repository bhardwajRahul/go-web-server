#!/bin/bash
# Pre-removal script for go-web-server

# Stop and disable the systemd service
if command -v systemctl >/dev/null 2>&1; then
    systemctl stop gowebserver.service 2>/dev/null || true
    systemctl disable gowebserver.service 2>/dev/null || true
fi

echo "go-web-server service stopped and disabled"