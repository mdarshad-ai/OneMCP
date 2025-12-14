# OneMCP

OneMCP is a cross-platform tool that provides centralized installation and management of MCP (Model Context Protocol) servers. It offers a single gateway endpoint that aggregates multiple MCP servers, allowing users to configure just one MCP connection in their tools (Claude, Cursor, opencode) while accessing all installed MCP servers.

## Features

- **Centralized MCP Server Management**: Install and manage multiple MCP servers from npm, pip, or custom repositories
- **Unified Gateway**: Single MCP endpoint that routes requests to individual servers
- **Web Interface**: Modern web UI for server configuration and monitoring
- **Cross-Platform**: Native support for Windows, macOS, and Linux
- **Simple Configuration**: One MCP config connects to gateway, all servers accessible through it

## Installation

### From Source

```bash
git clone https://github.com/yourusername/mcp-manager.git
cd mcp-manager
make setup
make build
```

### Pre-built Binaries

Download the latest release from the [releases page](https://github.com/yourusername/mcp-manager/releases).

## Quick Start

1. **Install an MCP server**:
   ```bash
   mcp-manager install github @modelcontextprotocol/server-github
   ```

2. **Configure credentials**:
   ```bash
   mcp-manager config github GITHUB_PERSONAL_ACCESS_TOKEN ghp_your_token_here
   ```

3. **Start the gateway**:
   ```bash
   mcp-manager start
   ```

4. **Configure your MCP client** (VS Code, Claude, etc.) to connect to the gateway.

## Usage

### Install Servers

```bash
# Install from npm
mcp-manager install github @modelcontextprotocol/server-github

# Install from pip
mcp-manager install postgres pip:mcp-server-postgres

# Install from custom source
mcp-manager install my-server custom:/path/to/server
```

### Manage Servers

```bash
# List installed servers
mcp-manager list

# Configure server credentials
mcp-manager config github GITHUB_PERSONAL_ACCESS_TOKEN ghp_xxx

# Start gateway
mcp-manager start

# Stop gateway
mcp-manager stop
```

### Web Interface

Access the web interface at `http://localhost:8080` to:
- Install new servers
- Configure credentials
- Monitor server status
- View logs

## Architecture

MCP Manager uses a filesystem-based storage system located in `~/.mcp/`:

```
~/.mcp/
├── config.json          # Global configuration
├── servers/             # Installed server metadata
│   ├── github.json
│   └── filesystem.json
├── credentials/         # API keys and credentials
│   ├── github.key
│   └── slack.key
├── logs/               # Server logs
└── cache/              # Downloaded packages
```

## Supported MCP Clients

- **Claude Desktop**: Configure with stdio transport
- **Cursor**: Use the MCP configuration
- **VS Code**: Add to `.vscode/mcp.json`
- **opencode**: Configure gateway endpoint

## Development

### Prerequisites

- Go 1.21+
- Node.js 18+ (for web UI)
- Make

### Setup

```bash
make setup    # Install dependencies
make build    # Build the binary
make test     # Run tests
make run      # Run in development mode
```

### Project Structure

```
cmd/mcp-manager/        # CLI entry point
internal/
├── cmd/               # CLI commands
├── config/            # Configuration management
├── storage/           # Filesystem storage
├── gateway/           # MCP gateway logic
└── installer/         # Server installation logic
pkg/                   # Public packages
web/                   # Web interface
scripts/              # Build and deployment scripts
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Roadmap

- [x] MCP server auto-discovery
- [x] Web interface for server management
- [x] Internet deployment & hosting
- [ ] Cross-platform installer packages
- [ ] Plugin system for custom server types
- [ ] Server health monitoring
- [ ] Backup/restore functionality

---

**Status**: Early development - Phase 1 (Core Infrastructure)</content>
<parameter name="filePath">mcp-manager/README.md