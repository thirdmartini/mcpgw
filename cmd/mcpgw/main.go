package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

var (
	configFile string
	debugMode  bool
)

var rootCmd = &cobra.Command{
	Use:   "mcpGW",
	Short: "Serve a chat gw and custom ui for MCP servers",
	Long: `mcpGW is a CLI tool that allows you to interact with various AI models
through a unified interface. It supports various tools through MCP servers.

Example:
  mcpGW --config ./config.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServer(context.Background())
	},
}

func main() {
	rootCmd.PersistentFlags().
		StringVar(&configFile, "config", "", "config file (default is $HOME/.mcp.json)")
	rootCmd.PersistentFlags().
		BoolVar(&debugMode, "debug", false, "enable debug logging")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
