package mcp_server

import (
	"context"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yourusername/onemcp/internal/config"
	"github.com/yourusername/onemcp/internal/gateway"
)

// Server represents the MCP server that aggregates tools from all installed servers
type Server struct {
	config *config.Config
	gw     *gateway.Gateway
}

// NewServer creates a new MCP server
func NewServer(cfg *config.Config, gw *gateway.Gateway) *Server {
	fmt.Printf("DEBUG: NewServer called\n")
	return &Server{
		config: cfg,
		gw:     gw,
	}
}

// CreateMCPServer creates and configures the MCP server
func (s *Server) CreateMCPServer() *mcpsdk.Server {
	fmt.Printf("DEBUG: Creating MCP server...\n")
	server := mcpsdk.NewServer(&mcpsdk.Implementation{
		Name:    "onemcp",
		Version: "0.1.0",
	}, nil)

	fmt.Printf("DEBUG: Adding list_servers tool...\n")
	// Add a tool to list all servers
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "list_servers",
		Description: "List all installed MCP servers and their status",
	}, s.ListServers)

	fmt.Printf("DEBUG: MCP server created successfully\n")
	return server
}

// ListServersArgs represents arguments for the list_servers tool
type ListServersArgs struct{}

// ListServers lists all installed servers
func (s *Server) ListServers(ctx context.Context, req *mcpsdk.CallToolRequest, args ListServersArgs) (*mcpsdk.CallToolResult, any, error) {
	fmt.Printf("DEBUG: ListServers tool called\n")
	servers := s.gw.ListServers()
	fmt.Printf("DEBUG: Found %d servers\n", len(servers))

	if len(servers) == 0 {
		return &mcpsdk.CallToolResult{
			Content: []mcpsdk.Content{
				&mcpsdk.TextContent{Text: "No MCP servers installed"},
			},
		}, nil, nil
	}

	result := "Installed MCP servers:\n"
	for _, server := range servers {
		result += fmt.Sprintf("- %s (%s): %s\n", server.Name, server.Type, server.Status)
	}

	fmt.Printf("DEBUG: Returning result: %s\n", result)
	return &mcpsdk.CallToolResult{
		Content: []mcpsdk.Content{
			&mcpsdk.TextContent{Text: result},
		},
	}, nil, nil
}
