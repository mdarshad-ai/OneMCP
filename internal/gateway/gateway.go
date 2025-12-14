package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/mdarshad-ai/OneMCP/internal/config"
	"github.com/mdarshad-ai/OneMCP/internal/storage"
)

// Gateway represents the MCP gateway server
type Gateway struct {
	config     *config.Config
	storage    *storage.FileStorage
	servers    map[string]*ServerProcess
	serversMux sync.RWMutex
}

// ServerProcess represents a running MCP server process
type ServerProcess struct {
	Name       string
	Config     *storage.ServerConfig
	Cmd        *exec.Cmd
	Stdin      io.WriteCloser
	Stdout     io.ReadCloser
	Stderr     io.ReadCloser
	running    bool
	runningMux sync.RWMutex
}

// NewGateway creates a new MCP gateway
func NewGateway(cfg *config.Config, store *storage.FileStorage) *Gateway {
	gw := &Gateway{
		config:  cfg,
		storage: store,
		servers: make(map[string]*ServerProcess),
	}

	// Load all installed servers
	gw.loadServers()

	return gw
}

// loadServers loads all server configurations into the gateway
func (g *Gateway) loadServers() {
	log.Printf("DEBUG: Loading servers...")
	servers, err := g.storage.ListServerConfigs()
	if err != nil {
		log.Printf("Failed to load server configs: %v", err)
		return
	}

	log.Printf("DEBUG: Found %d server configs", len(servers))
	g.serversMux.Lock()
	defer g.serversMux.Unlock()

	for _, serverConfig := range servers {
		process := &ServerProcess{
			Name:   serverConfig.Name,
			Config: serverConfig,
		}
		g.servers[serverConfig.Name] = process
	}

	log.Printf("Loaded %d MCP servers", len(g.servers))
}

// startAllServers starts all installed MCP servers
func (g *Gateway) startAllServers() error {
	g.serversMux.RLock()
	serverNames := make([]string, 0, len(g.servers))
	for name := range g.servers {
		serverNames = append(serverNames, name)
	}
	g.serversMux.RUnlock()

	log.Printf("Found %d servers to start: %v", len(serverNames), serverNames)

	var errors []string
	for _, name := range serverNames {
		log.Printf("Attempting to start server: %s", name)
		if err := g.StartServer(name); err != nil {
			// Only log as error if it's not "already running"
			if !strings.Contains(err.Error(), "already running") {
				log.Printf("Failed to start server %s: %v", name, err)
				errors = append(errors, fmt.Sprintf("%s: %v", name, err))
			} else {
				log.Printf("Server %s already running", name)
			}
		} else {
			log.Printf("Successfully started server: %s", name)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to start servers: %s", strings.Join(errors, "; "))
	}
	return nil
}

// monitorServers monitors server health and restarts failed servers
func (g *Gateway) monitorServers(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			g.checkAndRestartServers()
		}
	}
}

// checkAndRestartServers checks all servers and restarts any that have stopped
func (g *Gateway) checkAndRestartServers() {
	g.serversMux.RLock()
	serverNames := make([]string, 0, len(g.servers))
	for name := range g.servers {
		serverNames = append(serverNames, name)
	}
	g.serversMux.RUnlock()

	for _, name := range serverNames {
		g.serversMux.RLock()
		process, exists := g.servers[name]
		g.serversMux.RUnlock()

		if !exists {
			continue
		}

		if !process.IsRunning() {
			log.Printf("Server %s is not running, attempting to restart...", name)
			if err := g.StartServer(name); err != nil {
				log.Printf("Failed to restart server %s: %v", name, err)
			} else {
				log.Printf("Successfully restarted server: %s", name)
			}
		}
	}
}

// Start starts the MCP gateway and all installed servers
func (g *Gateway) Start(ctx context.Context) error {
	log.Printf("Starting MCP Gateway on %s:%d", g.config.Gateway.Host, g.config.Gateway.Port)

	// Start all installed servers
	log.Printf("Starting all installed MCP servers...")
	if err := g.startAllServers(); err != nil {
		log.Printf("Warning: Failed to start some servers: %v", err)
	}

	// Start server health monitoring
	go g.monitorServers(ctx)

	// Keep the gateway running
	<-ctx.Done()
	return nil
}

// StartServer starts a specific MCP server
func (g *Gateway) StartServer(serverName string) error {
	g.serversMux.Lock()
	defer g.serversMux.Unlock()

	process, exists := g.servers[serverName]
	if !exists {
		return fmt.Errorf("server %s not found", serverName)
	}

	if process.IsRunning() {
		log.Printf("Server %s is already running", serverName)
		return nil // Not an error, just already running
	}

	// Build the command based on server type and configuration
	var cmd *exec.Cmd
	var err error

	switch process.Config.Type {
	case storage.ServerTypeNPM:
		cmd, err = g.buildNPMCommand(process)
	case storage.ServerTypePIP:
		cmd, err = g.buildPIPCommand(process)
	case storage.ServerTypeCustom:
		cmd, err = g.buildCustomCommand(process)
	default:
		return fmt.Errorf("unsupported server type: %s", process.Config.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to build command: %w", err)
	}

	// Set up environment variables
	cmd.Env = os.Environ()

	// Load credentials and add them as environment variables
	if creds, err := g.storage.LoadCredentials(serverName); err == nil {
		for key, value := range creds.Data {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Create pipes for stdin/stdout/stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	process.Cmd = cmd
	process.Stdin = stdin
	process.Stdout = stdout
	process.Stderr = stderr

	// Start the process
	log.Printf("Starting server %s with command: %s %v", serverName, cmd.Path, cmd.Args[1:])
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server %s: %w", serverName, err)
	}

	process.runningMux.Lock()
	process.running = true
	process.runningMux.Unlock()

	log.Printf("Started MCP server: %s (PID: %d)", serverName, cmd.Process.Pid)

	// Start goroutine to monitor the process
	go func() {
		// Read stderr to capture any error output
		stderrBytes, _ := io.ReadAll(process.Stderr)
		if len(stderrBytes) > 0 {
			log.Printf("MCP server %s stderr: %s", serverName, string(stderrBytes))
		}

		err := cmd.Wait()
		process.runningMux.Lock()
		process.running = false
		process.runningMux.Unlock()

		if err != nil {
			log.Printf("MCP server %s exited with error: %v", serverName, err)
		} else {
			log.Printf("MCP server %s exited normally", serverName)
		}
	}()

	return nil
}

// buildNPMCommand builds the command for npm-based servers
func (g *Gateway) buildNPMCommand(process *ServerProcess) (*exec.Cmd, error) {
	cmdParts := strings.Fields(process.Config.Path)
	if len(cmdParts) < 2 {
		return nil, fmt.Errorf("invalid server path: %s", process.Config.Path)
	}

	// For filesystem server, we need to pass allowed directories
	if strings.Contains(process.Config.Package, "filesystem") {
		// Get allowed directories from credentials
		allowedDirs := "/tmp" // default
		if creds, err := g.storage.LoadCredentials(process.Name); err == nil {
			if dirs, ok := creds.Data["ALLOWED_DIRECTORIES"]; ok && dirs != "" {
				allowedDirs = dirs
			}
		}

		// Add the allowed directory as an argument
		args := append(cmdParts[1:], allowedDirs)
		return exec.Command(cmdParts[0], args...), nil
	}

	// For other npm servers, just run the command
	return exec.Command(cmdParts[0], cmdParts[1:]...), nil
}

// buildPIPCommand builds the command for pip-based servers
func (g *Gateway) buildPIPCommand(process *ServerProcess) (*exec.Cmd, error) {
	// For pip servers, the path should be the python command
	return exec.Command("python3", process.Config.Path), nil
}

// buildCustomCommand builds the command for custom servers
func (g *Gateway) buildCustomCommand(process *ServerProcess) (*exec.Cmd, error) {
	// For custom servers, parse the path as command + args
	cmdParts := strings.Fields(process.Config.Path)
	if len(cmdParts) == 0 {
		return nil, fmt.Errorf("invalid custom server path: %s", process.Config.Path)
	}
	return exec.Command(cmdParts[0], cmdParts[1:]...), nil
}

// StopServer stops a specific MCP server
func (g *Gateway) StopServer(serverName string) error {
	g.serversMux.Lock()
	defer g.serversMux.Unlock()

	process, exists := g.servers[serverName]
	if !exists {
		return fmt.Errorf("server %s not found", serverName)
	}

	if !process.IsRunning() {
		return fmt.Errorf("server %s is not running", serverName)
	}

	// Send interrupt signal
	if err := process.Cmd.Process.Signal(os.Interrupt); err != nil {
		// If interrupt fails, try to kill the process
		if killErr := process.Cmd.Process.Kill(); killErr != nil {
			return fmt.Errorf("failed to stop server %s: interrupt failed (%v), kill failed (%v)", serverName, err, killErr)
		}
	}

	process.runningMux.Lock()
	process.running = false
	process.runningMux.Unlock()

	log.Printf("Stopped MCP server: %s", serverName)
	return nil
}

// IsServerRunning checks if a server is running
func (g *Gateway) IsServerRunning(serverName string) bool {
	g.serversMux.RLock()
	defer g.serversMux.RUnlock()

	process, exists := g.servers[serverName]
	if !exists {
		return false
	}

	return process.IsRunning()
}

// ListServers returns information about all servers
func (g *Gateway) ListServers() map[string]*ServerInfo {
	g.serversMux.RLock()
	defer g.serversMux.RUnlock()

	result := make(map[string]*ServerInfo)
	for name, process := range g.servers {
		result[name] = &ServerInfo{
			Name:    name,
			Type:    string(process.Config.Type),
			Version: process.Config.Version,
			Status:  g.getServerStatus(name),
			Path:    process.Config.Path,
		}
	}

	return result
}

// ServerInfo represents information about a server
type ServerInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Status  string `json:"status"`
	Path    string `json:"path"`
}

// getServerStatus returns the current status of a server
func (g *Gateway) getServerStatus(serverName string) string {
	if g.IsServerRunning(serverName) {
		return "running"
	}
	return "stopped"
}

// IsRunning returns true if the process is running
func (p *ServerProcess) IsRunning() bool {
	p.runningMux.RLock()
	defer p.runningMux.RUnlock()

	if !p.running || p.Cmd == nil || p.Cmd.Process == nil {
		return false
	}

	// Check if the process is actually still running
	return p.Cmd.Process.Signal(syscall.Signal(0)) == nil
}

// SendMessage sends a message to the server process
func (p *ServerProcess) SendMessage(message interface{}) error {
	if !p.IsRunning() {
		return fmt.Errorf("server %s is not running", p.Name)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = p.Stdin.Write(append(data, '\n'))
	return err
}

// ReceiveMessage receives a message from the server process
func (p *ServerProcess) ReceiveMessage() ([]byte, error) {
	if !p.IsRunning() {
		return nil, fmt.Errorf("server %s is not running", p.Name)
	}

	// Read line from stdout
	buf := make([]byte, 1024)
	n, err := p.Stdout.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf[:n], nil
}