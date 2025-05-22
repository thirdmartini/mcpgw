package mcphost

/*
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/log"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"

	"github.com/thirdmartini/mcpgw/pkg/history"
	"github.com/thirdmartini/mcpgw/pkg/llm"
	"github.com/thirdmartini/mcpgw/pkg/llm/anthropic"
	"github.com/thirdmartini/mcpgw/pkg/llm/google"
	"github.com/thirdmartini/mcpgw/pkg/llm/ollama"
	"github.com/thirdmartini/mcpgw/pkg/llm/openai"
)

// loadSystemPrompt loads the system prompt from a JSON file
func loadSystemPrompt(filePath string) (string, error) {
	if filePath == "" {
		return "", nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading config file: %v", err)
	}

	// Parse only the systemPrompt field
	var config struct {
		SystemPrompt string `json:"systemPrompt"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("error parsing config file: %v", err)
	}

	return config.SystemPrompt, nil
}
*/
