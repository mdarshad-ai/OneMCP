# MCP Manager - Web Interface & Internet Deployment

## Phase 3: Web Interface ✅

The MCP Manager now includes a modern web interface for managing MCP servers with the following features:

- **Server Management**: Add, remove, start, and stop MCP servers
- **Real-time Status**: View server status and health
- **Configuration**: Set API keys and credentials through the web UI
- **Responsive Design**: Works on desktop and mobile devices

## Internet Access Setup

### Quick Start (Development)

For testing on your local machine:

```bash
# Build the application
make build

# Start with web interface
./mcp-manager start
```

The web interface will be available at `http://localhost:80`

### Production Deployment (Digital Ocean Ubuntu)

#### Option 1: Automated Deployment

```bash
# Run the deployment script
./deploy.sh
```

This will:
- Install system dependencies
- Configure firewall
- Set up systemd service
- Make the service accessible on port 80

#### Option 2: Manual Deployment

1. **Install Dependencies**
```bash
sudo apt update
sudo apt install -y curl wget git build-essential
```

2. **Install Go** (if not present)
```bash
wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
```

3. **Install Node.js** (if not present)
```bash
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs
```

4. **Build and Install**
```bash
make build
sudo cp mcp-manager.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable mcp-manager
```

5. **Configure Firewall**
```bash
sudo ufw allow 80
sudo ufw allow 22
sudo ufw --force enable
```

6. **Start Service**
```bash
sudo systemctl start mcp-manager
```

### Accessing from Internet

1. **Get your public IP**:
```bash
curl -s https://api.ipify.org
```

2. **Access the web interface**:
```
http://YOUR_PUBLIC_IP
```

3. **Optional: Set up a domain**
   - Point your domain DNS to your Digital Ocean droplet IP
   - The web interface will be accessible at `http://yourdomain.com`

### HTTPS Setup (Recommended)

For production use, set up HTTPS with Let's Encrypt:

```bash
# Install certbot
sudo apt install -y certbot

# Get SSL certificate (replace yourdomain.com)
sudo certbot certonly --standalone -d yourdomain.com

# The certificates will be in /etc/letsencrypt/live/yourdomain.com/
```

## Web Interface Features

### Server Management
- **Add Server**: Install new MCP servers from npm, pip, or custom sources
- **Start/Stop**: Control server lifecycle
- **Remove**: Uninstall servers
- **Status**: Real-time server status monitoring

### API Endpoints

```
GET    /api/servers          # List all servers
POST   /api/servers          # Add new server
DELETE /api/servers/:name    # Remove server
POST   /api/servers/:name/start   # Start server
POST   /api/servers/:name/stop    # Stop server
GET    /api/config           # Get configuration
PUT    /api/config           # Update configuration
```

### Security Considerations

- **Local Access Only**: By default, the web interface only accepts connections from localhost
- **API Keys**: Credentials are stored securely in the filesystem
- **Firewall**: Only port 80 is open to the internet
- **HTTPS**: Consider setting up SSL for production use

## Troubleshooting

### Port 80 Already in Use

If port 80 is already used by another service:

```bash
# Check what's using port 80
sudo netstat -tulpn | grep :80

# Stop conflicting service (example: apache2)
sudo systemctl stop apache2
sudo systemctl disable apache2
```

### Permission Issues

If you get permission errors:

```bash
# Run as root (not recommended for production)
sudo ./mcp-manager start

# Or configure to run on port 8080 and use reverse proxy
```

### Firewall Issues

```bash
# Check firewall status
sudo ufw status

# Allow port 80
sudo ufw allow 80
```

## Monitoring

### Service Logs
```bash
# View service logs
sudo journalctl -u mcp-manager -f

# View recent logs
sudo journalctl -u mcp-manager --since "1 hour ago"
```

### System Monitoring
```bash
# Check service status
sudo systemctl status mcp-manager

# Restart service
sudo systemctl restart mcp-manager
```

## Backup & Recovery

### Configuration Backup
```bash
# Backup MCP directory
tar -czf mcp-backup.tar.gz ~/.mcp/
```

### Service Management
```bash
# Stop service
sudo systemctl stop mcp-manager

# Start service
sudo systemctl start mcp-manager

# Restart service
sudo systemctl restart mcp-manager
```

---

**🎉 Your MCP Manager is now accessible from the internet!**

Visit `http://YOUR_PUBLIC_IP` to access the web interface and manage your MCP servers.