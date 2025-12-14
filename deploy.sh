#!/bin/bash

# MCP Manager Deployment Script for Digital Ocean Ubuntu
# This script sets up MCP Manager to be accessible from the internet

set -e

echo "🚀 Setting up MCP Manager for internet access..."

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   echo "❌ This script should not be run as root"
   exit 1
fi

# Update system
echo "📦 Updating system packages..."
sudo apt update && sudo apt upgrade -y

# Install required dependencies
echo "📦 Installing dependencies..."
sudo apt install -y curl wget git build-essential

# Install Go if not present
if ! command -v go &> /dev/null; then
    echo "📦 Installing Go..."
    wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    export PATH=$PATH:/usr/local/go/bin
fi

# Install Node.js if not present
if ! command -v node &> /dev/null; then
    echo "📦 Installing Node.js..."
    curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
    sudo apt-get install -y nodejs
fi

# Build MCP Manager
echo "🔨 Building MCP Manager..."
cd /home/arshad/mcp-manager
export PATH=$HOME/go/bin:$PATH
make build

# Configure firewall
echo "🔥 Configuring firewall..."
sudo ufw allow 80
sudo ufw allow 22
sudo ufw --force enable

# Install systemd service
echo "⚙️ Installing systemd service..."
sudo cp mcp-manager.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable mcp-manager

# Get public IP
PUBLIC_IP=$(curl -s https://api.ipify.org)

echo "✅ Setup complete!"
echo ""
echo "🌐 Your MCP Manager will be accessible at:"
echo "   http://$PUBLIC_IP"
echo ""
echo "🚀 To start the service:"
echo "   sudo systemctl start mcp-manager"
echo ""
echo "📊 To check status:"
echo "   sudo systemctl status mcp-manager"
echo ""
echo "📝 To view logs:"
echo "   sudo journalctl -u mcp-manager -f"
echo ""
echo "⚠️  Note: The web interface runs on port 80, so you may need to:"
echo "   - Stop any other web servers running on port 80"
echo "   - Configure your domain DNS to point to this server"
echo "   - Set up SSL/HTTPS with certbot if needed"