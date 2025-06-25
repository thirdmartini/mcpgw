package ollama

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ollama/ollama/api"

	"github.com/thirdmartini/mcpgw/pkg/history"
	"github.com/thirdmartini/mcpgw/pkg/llm"
)

func boolPtr(b bool) *bool {
	return &b
}

// Provider implements the Provider interface for Ollama
type Provider struct {
	client       *api.Client
	model        string
	systemPrompt string
}

// NewProvider creates a new Ollama provider
func NewProvider(host string, model string, systemPrompt string) (*Provider, error) {
	var client *api.Client
	var err error

	if host == "" {
		client, err = api.ClientFromEnvironment()
	} else {
		var u *url.URL
		if u, err = url.Parse(host); err == nil {
			client = api.NewClient(u, http.DefaultClient)
		}
	}
	if err != nil {
		return nil, err
	}
	return &Provider{
		client:       client,
		model:        model,
		systemPrompt: systemPrompt,
	}, nil
}

func (p *Provider) convertMessages(prompt string, messages []llm.Message) []api.Message {
	ollamaMessages := make([]api.Message, 0, len(messages)+1)

	if p.systemPrompt != "" {
		ollamaMessages = append(ollamaMessages, api.Message{
			Role:    "system",
			Content: p.systemPrompt,
		})
	}

	for _, msg := range messages {
		if msg.IsToolResponse() {
			var content string
			imageContent := make([]api.ImageData, 0)

			// Handle HistoryMessage format
			if historyMsg, ok := msg.(*history.HistoryMessage); ok {
				for _, block := range historyMsg.Content {
					if mcpContent, ok := block.Content.([]mcp.Content); ok {
						for _, c := range mcpContent {
							switch v := c.(type) {
							case mcp.ImageContent:
								imageDataRaw, err := base64.StdEncoding.DecodeString(v.Data)
								if err != nil {
									continue
								}
								imageContent = append(imageContent, api.ImageData(imageDataRaw))

							case mcp.TextContent:
								// We sometimes see a result of Text:"somedata",Image:"image data",Text:""
								content += v.Text + " "

							default:
								panic(fmt.Sprintf("unsupported content type: %T", v))
							}
						}
					} else {
						log.Fatalf("not mcp content %T\n", block.Content)
					}
				}
			}

			// If no content found yet, try standard content extraction
			if content == "" {
				content = msg.GetContent()
			}

			if content == "" {
				continue
			}

			ollamaMsg := api.Message{
				Role:    "tool",
				Content: content,
				Images:  imageContent,
			}
			ollamaMessages = append(ollamaMessages, ollamaMsg)
			continue
		}

		// Skip completely empty messages (no content and no tool calls)
		if msg.GetContent() == "" && len(msg.GetToolCalls()) == 0 {
			continue
		}

		ollamaMsg := api.Message{
			Role:    msg.GetRole(),
			Content: msg.GetContent(),
		}

		// Add tool calls for assistant messages
		if msg.GetRole() == "assistant" {
			for _, call := range msg.GetToolCalls() {
				if call.GetName() != "" {
					args := call.GetArguments()
					ollamaMsg.ToolCalls = append(
						ollamaMsg.ToolCalls,
						api.ToolCall{
							Function: api.ToolCallFunction{
								Name:      call.GetName(),
								Arguments: args,
							},
						},
					)
				}
			}
		}
		ollamaMessages = append(ollamaMessages, ollamaMsg)
	}

	return ollamaMessages
}

// Helper function to convert properties to Ollama's format
func convertProperties(props map[string]interface{}) map[string]struct {
	Type        api.PropertyType `json:"type"`
	Items       any              `json:"items,omitempty"`
	Description string           `json:"description"`
	Enum        []any            `json:"enum,omitempty"`
} {
	result := make(map[string]struct {
		Type        api.PropertyType `json:"type"`
		Items       any              `json:"items,omitempty"`
		Description string           `json:"description"`
		Enum        []any            `json:"enum,omitempty"`
	})

	for name, prop := range props {
		if propMap, ok := prop.(map[string]interface{}); ok {
			prop := struct {
				Type        api.PropertyType `json:"type"`
				Items       any              `json:"items,omitempty"`
				Description string           `json:"description"`
				Enum        []any            `json:"enum,omitempty"`
			}{
				Type:        api.PropertyType{getString(propMap, "type")},
				Description: getString(propMap, "description"),
			}

			// Handle enum if present
			if enumRaw, ok := propMap["enum"].([]interface{}); ok {
				for _, e := range enumRaw {
					if str, ok := e.(string); ok {
						prop.Enum = append(prop.Enum, str)
					}
				}
			}
			result[name] = prop
		}
	}
	return result
}

func (p *Provider) convertTools(tools []llm.Tool) []api.Tool {
	ollamaTools := make([]api.Tool, len(tools))
	for i, tool := range tools {
		ollamaTools[i] = api.Tool{
			Type: "function",
			Function: api.ToolFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters: struct {
					Type       string   `json:"type"`
					Defs       any      `json:"$defs,omitempty"`
					Items      any      `json:"items,omitempty"`
					Required   []string `json:"required"`
					Properties map[string]struct {
						Type        api.PropertyType `json:"type"`
						Items       any              `json:"items,omitempty"`
						Description string           `json:"description"`
						Enum        []any            `json:"enum,omitempty"`
					} `json:"properties"`
				}{
					Type:       tool.InputSchema.Type,
					Required:   tool.InputSchema.Required,
					Properties: convertProperties(tool.InputSchema.Properties),
				},
			},
		}
	}
	return ollamaTools
}

func (p *Provider) CreateMessage(
	ctx context.Context,
	prompt string,
	messages []llm.Message,
	tools []llm.Tool,
) (llm.Message, error) {
	ollamaMessages := p.convertMessages(prompt, messages)
	ollamaTools := p.convertTools(tools)

	// Convert generic messages to Ollama format
	log.Debug("creating message",
		"prompt", prompt,
		"num_messages", len(messages),
		"num_tools", len(tools))

	for idx, m := range ollamaMessages {
		log.Infof("M[%d]: %+v:[%+v]", idx, m.Content, m.ToolCalls)
	}

	request := api.ChatRequest{
		Model: p.model,

		Messages: ollamaMessages,
		Tools:    ollamaTools,
		Stream:   boolPtr(false),
		Options: map[string]interface{}{
			"num_ctx": 120000,
		},
	}

	response := &Message{}

	//sending, err := json.MarshalIndent(&request, "", "  ")
	//log.Infof("%s\n", string(sending))

	err := p.client.Chat(ctx, &request, func(r api.ChatResponse) error {
		if r.Done {
			response.metrics = r.Metrics
			response.message = r.Message
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (p *Provider) SupportsTools() bool {
	// Check if model supports function calling
	resp, err := p.client.Show(context.Background(), &api.ShowRequest{
		Model: p.model,
	})
	if err != nil {
		return false
	}
	return strings.Contains(resp.Modelfile, "<tools>")
}

func (p *Provider) Name() string {
	return "ollama"
}

func (p *Provider) CreateToolResponse(
	toolCallID string,
	content interface{},
) (llm.Message, error) {
	log.Debug("creating tool response",
		"tool_call_id", toolCallID,
		"content_type", fmt.Sprintf("%T", content),
		"content", content)

	contentStr := ""
	switch v := content.(type) {
	case string:
		contentStr = v
		log.Debug("using string content directly")
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			log.Error("failed to marshal tool response",
				"error", err,
				"content", content)
			return nil, fmt.Errorf("error marshaling tool response: %w", err)
		}
		contentStr = string(bytes)
		log.Debug("marshaled content to JSON string",
			"result", contentStr)
	}

	// Create a message with an explicit tool role
	msg := &Message{
		message: api.Message{
			Role:    "tool", // Explicitly set role to "tool"
			Content: contentStr,
			// No need to set ToolCalls for a tool response
		},
		ToolCallID: toolCallID,
	}

	log.Debug("created tool response message",
		"role", msg.GetRole(),
		"content", msg.GetContent(),
		"tool_call_id", msg.GetToolResponseID(),
		"raw_content", contentStr)

	return msg, nil
}

// Helper function to safely get string values from map
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
