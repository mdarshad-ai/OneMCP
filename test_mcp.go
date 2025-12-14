package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// Start the MCP gateway
	cmd := exec.Command("./onemcp", "start")
	cmd.Dir = "/home/arshad/onemcp"

	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Printf("Error creating stdin pipe: %v\n", err)
		os.Exit(1)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Error creating stdout pipe: %v\n", err)
		os.Exit(1)
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting MCP gateway: %v\n", err)
		os.Exit(1)
	}

	// Give it a moment to start
	// time.Sleep(2 * time.Second)

	// Send initialize request
	initRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	data, _ := json.Marshal(initRequest)
	fmt.Printf("Sending: %s\n", string(data))

	_, err = stdin.Write(append(data, '\n'))
	if err != nil {
		fmt.Printf("Error writing to stdin: %v\n", err)
		os.Exit(1)
	}

	// Read response
	scanner := bufio.NewScanner(stdout)
	if scanner.Scan() {
		response := scanner.Text()
		fmt.Printf("Received: %s\n", response)
	}

	// Send tools/list request
	toolsRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	}

	data, _ = json.Marshal(toolsRequest)
	fmt.Printf("Sending: %s\n", string(data))

	_, err = stdin.Write(append(data, '\n'))
	if err != nil {
		fmt.Printf("Error writing tools request: %v\n", err)
		os.Exit(1)
	}

	// Read response
	if scanner.Scan() {
		response := scanner.Text()
		fmt.Printf("Received: %s\n", response)
	}

	// Clean up
	cmd.Process.Kill()
	cmd.Wait()
}