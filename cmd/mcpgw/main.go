package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

var (
	configFile       string
	systemPromptFile string
	messageWindow    int
	modelFlag        string // New flag for model selection
	serverAddress    string
	serverRoot       string
	whisperAddress   string

	openaiBaseURL    string // Base URL for OpenAI API
	anthropicBaseURL string // Base URL for Anthropic API
	openaiAPIKey     string
	anthropicAPIKey  string
	googleAPIKey     string

	debugMode bool
)

var rootCmd = &cobra.Command{
	Use:   "mcphost",
	Short: "Chat with AI models through a unified interface",
	Long: `MCPHost is a CLI tool that allows you to interact with various AI models
through a unified interface. It supports various tools through MCP servers
and provides streaming responses.

Available models can be specified using the --model flag:
- Anthropic Claude (default): anthropic:claude-3-5-sonnet-latest
- OpenAI: openai:gpt-4
- Ollama models: ollama:modelname
- Google: google:modelname

Example:
  mcphost -m ollama:qwen2.5:3b
  mcphost -m openai:gpt-4
  mcphost -m google:gemini-2.0-flash`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServer(context.Background())
	},
}

func main() {
	rootCmd.PersistentFlags().
		StringVar(&configFile, "config", "", "config file (default is $HOME/.mcp.json)")
	rootCmd.PersistentFlags().
		StringVar(&systemPromptFile, "system-prompt", "", "system prompt json file")
	rootCmd.PersistentFlags().
		IntVar(&messageWindow, "message-window", 10, "number of messages to keep in context")
	rootCmd.PersistentFlags().
		StringVarP(&modelFlag, "model", "m", "anthropic:claude-3-5-sonnet-latest",
			"model to use (format: provider:model, e.g. anthropic:claude-3-5-sonnet-latest or ollama:qwen2.5:3b)")

	rootCmd.PersistentFlags().
		StringVarP(&serverAddress, "server", "s", ":8080",
			"The server address to listen on. The default is ':8080'.")
	rootCmd.PersistentFlags().
		StringVarP(&serverRoot, "server-root", "", "example/ui",
			"The server root to use for your UI experience")
	rootCmd.PersistentFlags().
		StringVarP(&whisperAddress, "whisper", "w", "http://localhost:8081",
			"The whisper server address to use for transcription")

	// Add debug flag
	rootCmd.PersistentFlags().
		BoolVar(&debugMode, "debug", false, "enable debug logging")

	flags := rootCmd.PersistentFlags()
	flags.StringVar(&openaiBaseURL, "openai-url", "", "base URL for OpenAI API (defaults to api.openai.com)")
	flags.StringVar(&anthropicBaseURL, "anthropic-url", "", "base URL for Anthropic API (defaults to api.anthropic.com)")
	flags.StringVar(&openaiAPIKey, "openai-api-key", "", "OpenAI API key")
	flags.StringVar(&anthropicAPIKey, "anthropic-api-key", "", "Anthropic API key")
	flags.StringVar(&googleAPIKey, "google-api-key", "", "Google (Gemini) API key")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
