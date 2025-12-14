# OneMCP - MCP Server Manager

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/mdarshad-ai/OneMCP)](https://github.com/mdarshad-ai/OneMCP/releases)

OneMCP is a cross-platform tool that provides **centralized installation and management** of MCP (Model Context Protocol) servers. It offers a **single gateway endpoint** that aggregates multiple MCP servers, allowing users to configure just **one MCP connection** in their tools (Claude, Cursor, opencode) while accessing all installed MCP servers.

## ✨ Features

- **🚀 Quick Server Installation**: Add MCP servers with single commands
- **🔐 Secure API Key Management**: Encrypted storage with environment variable injection
- **🌐 Web Interface**: Modern UI for server management and monitoring
- **⚡ Continuous Server Management**: Auto-start, health monitoring, and restart
- **🔄 Cross-Platform**: Native binaries for Windows, macOS, and Linux
- **📦 Multiple Installation Sources**: npm, pip, and custom repositories
- **🛡️ Secure Credential Storage**: Filesystem-based with restricted permissions

## 📦 Installation

### Quick Install (Recommended)

```bash
# One-command install for all platforms
curl -fsSL https://raw.githubusercontent.com/mdarshad-ai/OneMCP/main/install.sh | bash
```

### Manual Installation

1. **Download** the appropriate binary for your platform from [Releases](https://github.com/mdarshad-ai/OneMCP/releases)
2. **Make executable**: `chmod +x onemcp`
3. **Move to PATH**: `sudo mv onemcp /usr/local/bin/` (Linux/macOS) or add to PATH (Windows)

### From Source

```bash
git clone https://github.com/mdarshad-ai/OneMCP.git
cd OneMCP
make setup
make build
```

## 🚀 Quick Start

1. **Add your first MCP server**:
   ```bash
   onemcp add filesystem @modelcontextprotocol/server-filesystem
   ```

2. **Configure API keys** (if needed):
   ```bash
   onemcp set-key filesystem ALLOWED_DIRECTORIES "/tmp:/home/user"
   ```

3. **Start the MCP gateway**:
   ```bash
   onemcp start
   ```

4. **Configure your MCP client** to connect to OneMCP gateway.

## 📖 Usage

### Server Management

```bash
# Add servers from npm
onemcp add github @modelcontextprotocol/server-github
onemcp add brave-search @modelcontextprotocol/server-brave-search
onemcp add slack @modelcontextprotocol/server-slack

# List installed servers
onemcp list

# Check server status
onemcp status
```

### API Key Management

```bash
# Set API keys for servers
onemcp set-key github GITHUB_PERSONAL_ACCESS_TOKEN ghp_xxxxxxxxxx
onemcp set-key brave-search BRAVE_API_KEY your-brave-api-key
onemcp set-key slack SLACK_BOT_TOKEN xoxb-your-token

# View configured keys
onemcp get-keys github

# Remove keys
onemcp remove-key github GITHUB_PERSONAL_ACCESS_TOKEN
```

### Gateway Control

```bash
# Start MCP gateway (starts all servers automatically)
onemcp start

# Start/stop individual servers
onemcp start-server filesystem
onemcp stop-server filesystem

# Access web interface
onemcp web  # Opens at http://localhost:8080
```

## 🏗️ Architecture

OneMCP uses a filesystem-based storage system in `~/.mcp/`:

```
~/.mcp/
├── config.json          # Global configuration
├── servers/             # Server metadata & configs
│   ├── filesystem.json
│   ├── github.json
│   └── brave-search.json
├── credentials/         # Encrypted API keys (0600 perms)
│   ├── filesystem.key
│   ├── github.key
│   └── brave-search.key
├── cache/               # Downloaded packages
└── logs/               # Server logs
```

### Security Features

- ✅ **Encrypted JSON storage** for API keys
- ✅ **File permission restrictions** (0600 - owner only)
- ✅ **Environment variable injection** (not command args)
- ✅ **Runtime-only key exposure**
- ✅ **Process isolation** per server

## 🖥️ Supported MCP Clients

### opencode
```json
// ~/.opencode/mcp-servers.json
{
  "mcpServers": {
    "onemcp": {
      "command": "/path/to/onemcp",
      "args": ["start"]
    }
  }
}
```

### Claude Desktop
```json
// ~/Library/Application Support/Claude/claude_desktop_config.json
{
  "mcpServers": {
    "onemcp": {
      "command": "/path/to/onemcp",
      "args": ["start"]
    }
  }
}
```

### Cursor
```json
// .cursor/mcp.json or .vscode/mcp.json
{
  "mcpServers": {
    "onemcp": {
      "command": "/path/to/onemcp",
      "args": ["start"]
    }
  }
}
```

## 🛠️ Development

### Prerequisites

- **Go 1.23+**
- **Make**
- **Git**

### Setup

```bash
# Clone repository
git clone https://github.com/mdarshad-ai/OneMCP.git
cd OneMCP

# Setup development environment
make setup

# Build for development
make build

# Run tests
make test

# Cross-platform builds
make build-all  # Creates binaries for Linux, macOS, Windows
```

### Project Structure

```
onemcp/
├── cmd/onemcp/              # CLI entry point
├── internal/
│   ├── cmd/                 # CLI command implementations
│   ├── config/              # Configuration management
│   ├── storage/             # Filesystem storage layer
│   ├── gateway/             # MCP gateway & server management
│   ├── installer/           # Package installation logic
│   ├── mcp-server/          # MCP protocol server
│   └── web/                 # Web interface
├── pkg/                     # Public packages (future)
├── Makefile                 # Build automation
├── install.sh              # Installation script
├── RELEASE_NOTES.md        # Release notes
└── README.md               # This file
```

## 🤝 Contributing

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feature/amazing-feature`
3. **Make** your changes and add tests
4. **Commit**: `git commit -m 'Add amazing feature'`
5. **Push**: `git push origin feature/amazing-feature`
6. **Open** a Pull Request

### Development Guidelines

- Follow Go conventions and best practices
- Add tests for new features
- Update documentation
- Ensure cross-platform compatibility

## 📋 Roadmap

- [x] **Core Infrastructure** - CLI and basic server management
- [x] **Web Interface** - Modern UI for server management
- [x] **API Key Management** - Secure credential storage
- [x] **Continuous Server Management** - Auto-start and health monitoring
- [x] **Cross-Platform Builds** - Native binaries for all platforms
- [ ] **Package Manager Integration** - Homebrew, APT, Chocolatey
- [ ] **Plugin System** - Custom server types and extensions
- [ ] **Advanced Monitoring** - Metrics and performance tracking
- [ ] **Backup/Restore** - Configuration backup and recovery
- [ ] **Docker Support** - Containerized deployments

## 📄 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Model Context Protocol](https://modelcontextprotocol.io/) - The protocol that makes this possible
- [Official Go SDK](https://github.com/modelcontextprotocol/go-sdk) - MCP implementation
- [Cobra](https://github.com/spf13/cobra) - CLI framework

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/mdarshad-ai/OneMCP/issues)
- **Discussions**: [GitHub Discussions](https://github.com/mdarshad-ai/OneMCP/discussions)
- **Documentation**: See [INSTALL.md](INSTALL.md) and [RELEASE_NOTES.md](RELEASE_NOTES.md)

---

**OneMCP** - *Simplifying MCP server management for everyone!* 🎉</content>
<parameter name="filePath">mcp-manager/README.md