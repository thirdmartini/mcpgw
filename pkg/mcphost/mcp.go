package mcphost

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/thirdmartini/mcpgw/pkg/llm"
)

const (
	transportStdio = "stdio"
	transportSSE   = "sse"
)

var (
	// Tokyo Night theme colors
	tokyoPurple = lipgloss.Color("99")  // #9d7cd8
	tokyoCyan   = lipgloss.Color("73")  // #7dcfff
	tokyoBlue   = lipgloss.Color("111") // #7aa2f7
	tokyoGreen  = lipgloss.Color("120") // #73daca
	tokyoRed    = lipgloss.Color("203") // #f7768e
	tokyoOrange = lipgloss.Color("215") // #ff9e64
	tokyoFg     = lipgloss.Color("189") // #c0caf5
	tokyoGray   = lipgloss.Color("237") // #3b4261
	tokyoBg     = lipgloss.Color("234") // #1a1b26

	promptStyle = lipgloss.NewStyle().
			Foreground(tokyoBlue).
			PaddingLeft(2)

	responseStyle = lipgloss.NewStyle().
			Foreground(tokyoFg).
			PaddingLeft(2)

	errorStyle = lipgloss.NewStyle().
			Foreground(tokyoRed).
			Bold(true)

	toolNameStyle = lipgloss.NewStyle().
			Foreground(tokyoCyan).
			Bold(true)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(tokyoFg).
				PaddingBottom(1)

	contentStyle = lipgloss.NewStyle().
			Background(tokyoBg).
			PaddingLeft(4).
			PaddingRight(4)
)

type MCPConfig struct {
	MCPServers map[string]ServerConfigWrapper `json:"mcpServers"`
}

type ServerConfig interface {
	GetType() string
}

type STDIOServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

func (s STDIOServerConfig) GetType() string {
	return transportStdio
}

type SSEServerConfig struct {
	Url     string   `json:"url"`
	Headers []string `json:"headers,omitempty"`
}

func (s SSEServerConfig) GetType() string {
	return transportSSE
}

type ServerConfigWrapper struct {
	Config ServerConfig
}

func (w *ServerConfigWrapper) UnmarshalJSON(data []byte) error {
	var typeField struct {
		Url string `json:"url"`
	}

	if err := json.Unmarshal(data, &typeField); err != nil {
		return err
	}
	if typeField.Url != "" {
		// If the URL field is present, treat it as an SSE server
		var sse SSEServerConfig
		if err := json.Unmarshal(data, &sse); err != nil {
			return err
		}
		w.Config = sse
	} else {
		// Otherwise, treat it as a STDIOServerConfig
		var stdio STDIOServerConfig
		if err := json.Unmarshal(data, &stdio); err != nil {
			return err
		}
		w.Config = stdio
	}

	return nil
}
func (w ServerConfigWrapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.Config)
}

func mcpToolsToAnthropicTools(
	serverName string,
	mcpTools []mcp.Tool,
) []llm.Tool {
	anthropicTools := make([]llm.Tool, len(mcpTools))

	for i, tool := range mcpTools {
		namespacedName := fmt.Sprintf("%s__%s", serverName, tool.Name)

		anthropicTools[i] = llm.Tool{
			Name:        namespacedName,
			Description: tool.Description,
			InputSchema: llm.Schema{
				Type:       tool.InputSchema.Type,
				Properties: tool.InputSchema.Properties,
				Required:   tool.InputSchema.Required,
			},
		}
	}

	return anthropicTools
}

func loadMCPConfig(configFile string) (*MCPConfig, error) {
	var configPath string
	if configFile != "" {
		configPath = configFile
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("error getting home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, ".mcp.json")
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		defaultConfig := MCPConfig{
			MCPServers: make(map[string]ServerConfigWrapper),
		}

		// Create the file with default config
		configData, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("error creating default config: %w", err)
		}

		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			return nil, fmt.Errorf("error writing default config file: %w", err)
		}

		log.Info("Created default config file", "path", configPath)
		return &defaultConfig, nil
	}

	// Read existing config
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf(
			"error reading config file %s: %w",
			configPath,
			err,
		)
	}

	var config MCPConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return &config, nil
}

func createMCPClients(
	config *MCPConfig,
) (map[string]mcpclient.MCPClient, error) {
	clients := make(map[string]mcpclient.MCPClient)

	for name, server := range config.MCPServers {
		var client mcpclient.MCPClient
		var err error

		if server.Config.GetType() == transportSSE {
			sseConfig := server.Config.(SSEServerConfig)

			options := []mcpclient.ClientOption{}

			if sseConfig.Headers != nil {
				// Parse headers from the config
				headers := make(map[string]string)
				for _, header := range sseConfig.Headers {
					parts := strings.SplitN(header, ":", 2)
					if len(parts) == 2 {
						key := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(parts[1])
						headers[key] = value
					}
				}
				options = append(options, mcpclient.WithHeaders(headers))
			}

			client, err = mcpclient.NewSSEMCPClient(
				sseConfig.Url,
				options...,
			)
			if err == nil {
				err = client.(*mcpclient.SSEMCPClient).Start(context.Background())
			}
		} else {
			stdioConfig := server.Config.(STDIOServerConfig)
			var env []string
			for k, v := range stdioConfig.Env {
				env = append(env, fmt.Sprintf("%s=%s", k, v))
			}
			client, err = mcpclient.NewStdioMCPClient(
				stdioConfig.Command,
				env,
				stdioConfig.Args...)
		}
		if err != nil {
			for _, c := range clients {
				c.Close()
			}
			return nil, fmt.Errorf(
				"failed to create MCP client for %s: %w",
				name,
				err,
			)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		log.Info("Initializing server...", "name", name)
		initRequest := mcp.InitializeRequest{}
		initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
		initRequest.Params.ClientInfo = mcp.Implementation{
			Name:    "mcphost",
			Version: "0.1.0",
		}
		initRequest.Params.Capabilities = mcp.ClientCapabilities{}

		_, err = client.Initialize(ctx, initRequest)
		if err != nil {
			client.Close()
			for _, c := range clients {
				c.Close()
			}
			return nil, fmt.Errorf(
				"failed to initialize MCP client for %s: %w",
				name,
				err,
			)
		}

		clients[name] = client
	}

	return clients, nil
}

func toolResultToString(toolResult *mcp.CallToolResult) string {
	toolResult.Content = append(toolResult.Content, mcp.TextContent{})

	content := ""
	for _, mcpContent := range toolResult.Content {
		switch v := mcpContent.(type) {
		case mcp.TextContent:
			content += "[ Text:" + v.Text + " ]"
		case mcp.ImageContent:
			content += "[(Image Data)]"
		}
	}
	return content
}
