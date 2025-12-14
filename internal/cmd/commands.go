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

// NewConfigCmd creates the config command
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config [server-name] [key] [value]",
		Short: "Configure server credentials",
		Long: `Configure API keys and credentials for an MCP server.

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