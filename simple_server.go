package main

import (
	"context"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// SimpleServer represents a simple MCP server for testing
type SimpleServer struct{}

// ListServersArgs represents arguments for the list_servers tool
type ListServersArgs struct{}

// ListServers lists all installed servers
func (s *SimpleServer) ListServers(ctx context.Context, req *mcpsdk.CallToolRequest, args ListServersArgs) (*mcpsdk.CallToolResult, any, error) {
	fmt.Printf("DEBUG: ListServers tool called\n")

	result := "Test MCP servers:\n- test-server (npm): installed\n- filesystem (npm): installed"

	return &mcpsdk.CallToolResult{
		Content: []mcpsdk.Content{
			&mcpsdk.TextContent{Text: result},
		},
	}, nil, nil
}

func main() {
	fmt.Printf("DEBUG: Creating simple MCP server...\n")
	server := mcpsdk.NewServer(&mcpsdk.Implementation{
		Name:    "test-mcp-server",
		Version: "0.1.0",
	}, nil)

	fmt.Printf("DEBUG: Adding tool...\n")
	simpleServer := &SimpleServer{}
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "list_servers",
		Description: "List all installed MCP servers and their status",
	}, simpleServer.ListServers)

	fmt.Printf("DEBUG: Starting server...\n")
	transport := &mcpsdk.StdioTransport{}
	if err := server.Run(context.Background(), transport); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}