package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/onemcp/internal/cmd"
)

var rootCmd = &cobra.Command{
	Use:   "onemcp",
	Short: "MCP Manager - Centralized MCP server management",
	Long: `MCP Manager provides centralized installation and management of MCP (Model Context Protocol) servers.
It offers a single gateway endpoint that aggregates multiple MCP servers, allowing users to configure
just one MCP connection in their tools (Claude, Cursor, opencode) while accessing all installed MCP servers.`,
}

func init() {
	rootCmd.AddCommand(cmd.NewInstallCmd())
	rootCmd.AddCommand(cmd.NewAddCmd())
	rootCmd.AddCommand(cmd.NewListCmd())
	rootCmd.AddCommand(cmd.NewSetKeyCmd())
	rootCmd.AddCommand(cmd.NewGetKeysCmd())
	rootCmd.AddCommand(cmd.NewRemoveKeyCmd())
	rootCmd.AddCommand(cmd.NewConfigCmd())
	rootCmd.AddCommand(cmd.NewStartCmd())
	rootCmd.AddCommand(cmd.NewStartServerCmd())
	rootCmd.AddCommand(cmd.NewStopServerCmd())
	rootCmd.AddCommand(cmd.NewStatusCmd())
	rootCmd.AddCommand(cmd.NewWebCmd())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}