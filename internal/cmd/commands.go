package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"github.com/yourusername/onemcp/internal/config"
	"github.com/yourusername/onemcp/internal/gateway"
	"github.com/yourusername/onemcp/internal/installer"
	"github.com/yourusername/onemcp/internal/mcp-server"
	"github.com/yourusername/onemcp/internal/web"
	"github.com/yourusername/onemcp/internal/storage"
)

var (
	mcpDir string
	store  *storage.FileStorage
)

// initConfig initializes the configuration and storage
func initConfig() error {
	if mcpDir == "" {
		var err error
		mcpDir, err = config.GetMCPDir()
		if err != nil {
			return fmt.Errorf("failed to get MCP directory: %w", err)
		}
	}

	var err error
	store, err = storage.NewFileStorage(mcpDir)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	return nil
}

// getPublicIP gets the public IP address of the server
func getPublicIP() string {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "your-server-ip"
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "your-server-ip"
	}

	return string(ip)
}

// NewInstallCmd creates the install command
func NewInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [server-name] [source]",
		Short: "Install an MCP server",
		Long: `Install an MCP server from various sources.

Examples:
  onemcp install github @modelcontextprotocol/server-github
  onemcp install postgres pip:mcp-server-postgres
  onemcp install my-server custom:git@github.com/user/repo.git
  onemcp install local-server custom:/path/to/local/server`,
		Args: cobra.ExactArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			source := args[1]

			fmt.Printf("Installing MCP server '%s' from %s...\n", name, source)

			// Check if server already exists
			if _, err := store.LoadServerConfig(name); err == nil {
				return fmt.Errorf("server '%s' is already installed", name)
			}

			// Create installer
			inst := installer.NewInstaller(store.GetCacheDir())

			var result *installer.InstallResult
			var err error

			// Determine installation method
			if strings.HasPrefix(source, "pip:") {
				// PIP installation
				packageName := strings.TrimPrefix(source, "pip:")
				result, err = inst.InstallFromPIP(packageName)
			} else if strings.HasPrefix(source, "custom:") {
				// Custom installation
				customSource := strings.TrimPrefix(source, "custom:")
				result, err = inst.InstallFromCustom(customSource)
			} else {
				// Default to NPM installation
				result, err = inst.InstallFromNPM(source)
			}

			if err != nil {
				return fmt.Errorf("installation failed: %w", err)
			}

			if !result.Success {
				return fmt.Errorf("installation failed: %s", result.Error)
			}

			// Create server configuration
			serverConfig := &storage.ServerConfig{
				Name:         name,
				Type:         result.Type,
				Package:      result.Package,
				Version:      result.Version,
				InstalledAt:  time.Now(),
				Status:       storage.StatusInstalled,
				Config:       make(map[string]interface{}),
				Path:         result.InstallPath,
				Dependencies: make(map[string]string),
			}

			// Add runtime dependencies
			switch result.Type {
			case storage.ServerTypeNPM:
				serverConfig.Dependencies["node"] = ">=18.0.0"
			case storage.ServerTypePIP:
				serverConfig.Dependencies["python"] = ">=3.8.0"
			}

			if err := store.SaveServerConfig(serverConfig); err != nil {
				return fmt.Errorf("failed to save server config: %w", err)
			}

			fmt.Printf("Successfully installed MCP server '%s' (version: %s)\n", name, result.Version)
			fmt.Printf("Installation path: %s\n", result.InstallPath)
			return nil
		},
	}

	return cmd
}

// NewListCmd creates the list command
func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed MCP servers",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			servers, err := store.ListServerConfigs()
			if err != nil {
				return fmt.Errorf("failed to list servers: %w", err)
			}

			if len(servers) == 0 {
				fmt.Println("No MCP servers installed")
				return nil
			}

			fmt.Println("Installed MCP servers:")
			fmt.Println("NAME\t\tTYPE\t\tSTATUS\t\tVERSION")
			fmt.Println("----\t\t----\t\t------\t\t-------")

			for _, server := range servers {
				fmt.Printf("%s\t\t%s\t\t%s\t\t%s\n",
					server.Name,
					server.Type,
					server.Status,
					server.Version)
			}

			return nil
		},
	}

	return cmd
}

// NewAddCmd creates the add command for quick server addition
func NewAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [server-name] [package-name]",
		Short: "Add an MCP server from npm",
		Long: `Quickly add an MCP server from npm registry.

Examples:
  onemcp add filesystem @modelcontextprotocol/server-filesystem
  onemcp add github @modelcontextprotocol/server-github
  onemcp add slack @modelcontextprotocol/server-slack`,
		Args: cobra.ExactArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			packageName := args[1]

			fmt.Printf("Adding MCP server '%s' from npm package '%s'...\n", name, packageName)

			// Check if server already exists
			if _, err := store.LoadServerConfig(name); err == nil {
				return fmt.Errorf("server '%s' is already added", name)
			}

			// Create installer
			inst := installer.NewInstaller(store.GetCacheDir())

			// Install from NPM
			result, err := inst.InstallFromNPM(packageName)
			if err != nil {
				return fmt.Errorf("installation failed: %w", err)
			}

			if !result.Success {
				return fmt.Errorf("installation failed: %s", result.Error)
			}

			// Create server configuration
			serverConfig := &storage.ServerConfig{
				Name:         name,
				Type:         result.Type,
				Package:      result.Package,
				Version:      result.Version,
				InstalledAt:  time.Now(),
				Status:       storage.StatusInstalled,
				Config:       make(map[string]interface{}),
				Path:         result.InstallPath,
				Dependencies: make(map[string]string),
			}

			// Add runtime dependencies
			serverConfig.Dependencies["node"] = ">=18.0.0"

			if err := store.SaveServerConfig(serverConfig); err != nil {
				return fmt.Errorf("failed to save server config: %w", err)
			}

			fmt.Printf("Successfully added MCP server '%s' (version: %s)\n", name, result.Version)
			fmt.Printf("Installation path: %s\n", result.InstallPath)
			fmt.Printf("\nTo configure API keys if needed:\n")
			fmt.Printf("  onemcp set-key %s [KEY_NAME] [KEY_VALUE]\n", name)
			return nil
		},
	}

	return cmd
}

// NewSetKeyCmd creates the set-key command for API key management
func NewSetKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-key [server-name] [key-name] [key-value]",
		Short: "Set API key for an MCP server",
		Long: `Set API keys and credentials required by MCP servers.

Examples:
  onemcp set-key github GITHUB_PERSONAL_ACCESS_TOKEN ghp_xxxxxxxxxx
  onemcp set-key slack SLACK_BOT_TOKEN xoxb-xxxxxxxxxx
  onemcp set-key tavily TAVILY_API_KEY tvly-xxxxxxxxxx

The keys will be securely stored and automatically provided to the server when it starts.`,
		Args: cobra.ExactArgs(3),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			serverName := args[0]
			keyName := args[1]
			keyValue := args[2]

			// Check if server exists
			if _, err := store.LoadServerConfig(serverName); err != nil {
				return fmt.Errorf("server '%s' not found. Add it first with 'onemcp add %s [package]'", serverName, serverName)
			}

			// Load existing credentials or create new
			creds, err := store.LoadCredentials(serverName)
			if err != nil {
				// Create new credentials if they don't exist
				creds = &storage.Credential{
					Data: make(map[string]string),
				}
			}

			// Set the credential
			creds.Data[keyName] = keyValue

			// Save credentials
			if err := store.SaveCredentials(serverName, creds); err != nil {
				return fmt.Errorf("failed to save API key: %w", err)
			}

			fmt.Printf("Successfully set API key '%s' for server '%s'\n", keyName, serverName)
			fmt.Printf("Key stored securely in %s\n", store.GetCredentialsPath(serverName))
			return nil
		},
	}

	return cmd
}

// NewGetKeysCmd creates the get-keys command to view configured keys
func NewGetKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-keys [server-name]",
		Short: "View configured API keys for a server",
		Long: `View all configured API keys for a specific MCP server.

Example:
  onemcp get-keys github`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			serverName := args[0]

			// Load credentials
			creds, err := store.LoadCredentials(serverName)
			if err != nil {
				fmt.Printf("No API keys configured for server '%s'\n", serverName)
				return nil
			}

			fmt.Printf("API keys for server '%s':\n", serverName)
			for key := range creds.Data {
				fmt.Printf("  - %s: [CONFIGURED]\n", key)
			}

			return nil
		},
	}

	return cmd
}

// NewRemoveKeyCmd creates the remove-key command
func NewRemoveKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-key [server-name] [key-name]",
		Short: "Remove an API key for a server",
		Long: `Remove a specific API key for an MCP server.

Example:
  onemcp remove-key github GITHUB_PERSONAL_ACCESS_TOKEN`,
		Args: cobra.ExactArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			serverName := args[0]
			keyName := args[1]

			// Load existing credentials
			creds, err := store.LoadCredentials(serverName)
			if err != nil {
				return fmt.Errorf("no credentials found for server '%s'", serverName)
			}

			// Check if key exists
			if _, exists := creds.Data[keyName]; !exists {
				return fmt.Errorf("API key '%s' not found for server '%s'", keyName, serverName)
			}

			// Remove the key
			delete(creds.Data, keyName)

			// Save updated credentials
			if err := store.SaveCredentials(serverName, creds); err != nil {
				return fmt.Errorf("failed to update credentials: %w", err)
			}

			fmt.Printf("Successfully removed API key '%s' for server '%s'\n", keyName, serverName)
			return nil
		},
	}

	return cmd
}

// NewConfigCmd creates the config command (legacy, kept for compatibility)
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config [server-name] [key] [value]",
		Short: "Configure server credentials (legacy)",
		Long: `Configure API keys and credentials for an MCP server.

Note: Use 'set-key' command instead for better API key management.

Example:
  onemcp config github GITHUB_PERSONAL_ACCESS_TOKEN ghp_xxxxxxxxxx`,
		Args: cobra.ExactArgs(3),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			key := args[1]
			value := args[2]

			// Load existing credentials or create new
			creds, err := store.LoadCredentials(name)
			if err != nil {
				// Create new credentials if they don't exist
				creds = &storage.Credential{
					Data: make(map[string]string),
				}
			}

			// Set the credential
			creds.Data[key] = value

			// Save credentials
			if err := store.SaveCredentials(name, creds); err != nil {
				return fmt.Errorf("failed to save credentials: %w", err)
			}

			fmt.Printf("Successfully configured credential for server '%s'\n", name)
			fmt.Printf("Note: Consider using 'onemcp set-key' for better API key management.\n")
			return nil
		},
	}

	return cmd
}

// NewStartCmd creates the start command
func NewStartCmd() *cobra.Command {
	var cfg *config.Config
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the MCP gateway",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := initConfig(); err != nil {
				return err
			}
			var err error
			cfg, err = config.LoadConfig(mcpDir)
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Starting MCP gateway...")
			fmt.Println("DEBUG: initConfig completed")

			// Create gateway for server management
			fmt.Println("DEBUG: Creating gateway...")
			gw := gateway.NewGateway(cfg, store)
			fmt.Println("DEBUG: Gateway created")

			// Start the gateway (which starts all servers)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go func() {
				if err := gw.Start(ctx); err != nil {
					log.Printf("Gateway error: %v", err)
				}
			}()

			// Give gateway time to start servers
			time.Sleep(2 * time.Second)

			// Create MCP server
			fmt.Println("DEBUG: Creating MCP server service...")
			mcpSrv := mcp_server.NewServer(cfg, gw)
			mcpServer := mcpSrv.CreateMCPServer()
			if mcpServer == nil {
				return fmt.Errorf("failed to create MCP server")
			}
			fmt.Println("DEBUG: MCP server created")

			// Create web server
			fmt.Println("DEBUG: Creating web server...")
			webSrv := web.NewServer(cfg, gw, store)
			fmt.Println("DEBUG: Web server created")

			// Start web server in a goroutine
			go func() {
				log.Printf("Starting web server on port %d", cfg.Web.Port)
				if err := webSrv.Start(); err != nil {
					log.Printf("Web server error: %v", err)
				}
			}()

			// Start MCP server with stdio transport
			fmt.Println("MCP gateway started successfully")
			if cfg.Web.Host == "0.0.0.0" {
				fmt.Printf("Web interface: http://%s:%d\n", getPublicIP(), cfg.Web.Port)
			} else {
				fmt.Printf("Web interface: http://%s:%d\n", cfg.Web.Host, cfg.Web.Port)
			}
			fmt.Printf("Ready to accept MCP connections on stdio\n")
			fmt.Println("Press Ctrl+C to stop")

			// Run the MCP server
			transport := &mcpsdk.StdioTransport{}
			return mcpServer.Run(context.Background(), transport)
		},
	}

	return cmd
}

// NewStartServerCmd creates the start-server command
func NewStartServerCmd() *cobra.Command {
	var cfg *config.Config
	cmd := &cobra.Command{
		Use:   "start-server [server-name]",
		Short: "Start a specific MCP server",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := initConfig(); err != nil {
				return err
			}
			var err error
			cfg, err = config.LoadConfig(mcpDir)
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			serverName := args[0]

			gw := gateway.NewGateway(cfg, store)
			if err := gw.StartServer(serverName); err != nil {
				return fmt.Errorf("failed to start server: %w", err)
			}

			fmt.Printf("Started MCP server: %s\n", serverName)
			return nil
		},
	}

	return cmd
}

// NewStopServerCmd creates the stop-server command
func NewStopServerCmd() *cobra.Command {
	var cfg *config.Config
	cmd := &cobra.Command{
		Use:   "stop-server [server-name]",
		Short: "Stop a specific MCP server",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := initConfig(); err != nil {
				return err
			}
			var err error
			cfg, err = config.LoadConfig(mcpDir)
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			serverName := args[0]

			gw := gateway.NewGateway(cfg, store)
			if err := gw.StopServer(serverName); err != nil {
				return fmt.Errorf("failed to stop server: %w", err)
			}

			fmt.Printf("Stopped MCP server: %s\n", serverName)
			return nil
		},
	}

	return cmd
}

// NewStatusCmd creates the status command
func NewStatusCmd() *cobra.Command {
	var cfg *config.Config
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show status of all MCP servers",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := initConfig(); err != nil {
				return err
			}
			var err error
			cfg, err = config.LoadConfig(mcpDir)
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			gw := gateway.NewGateway(cfg, store)
			servers := gw.ListServers()

			if len(servers) == 0 {
				fmt.Println("No MCP servers installed")
				return nil
			}

			fmt.Println("MCP Server Status:")
			fmt.Println("NAME\t\tTYPE\t\tSTATUS\t\tVERSION")
			fmt.Println("----\t\t----\t\t------\t\t-------")

			for _, server := range servers {
				fmt.Printf("%s\t\t%s\t\t%s\t\t%s\n",
					server.Name,
					server.Type,
					server.Status,
					server.Version)
			}

			return nil
		},
	}

	return cmd
}

// NewWebCmd creates the web command
func NewWebCmd() *cobra.Command {
	var cfg *config.Config
	cmd := &cobra.Command{
		Use:   "web",
		Short: "Start the web interface only",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := initConfig(); err != nil {
				return err
			}
			var err error
			cfg, err = config.LoadConfig(mcpDir)
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Starting MCP Manager Web Interface...")

			// Create gateway for server management
			gw := gateway.NewGateway(cfg, store)

			// Create web server
			webSrv := web.NewServer(cfg, gw, store)

			fmt.Printf("Web interface starting on %s:%d\n", cfg.Web.Host, cfg.Web.Port)
			fmt.Println("Press Ctrl+C to stop")

			// Start web server
			return webSrv.Start()
		},
	}

	return cmd
}