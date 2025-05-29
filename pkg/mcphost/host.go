package mcphost

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/thirdmartini/mcpgw/pkg/history"
	"github.com/thirdmartini/mcpgw/pkg/llm"
)

type Host struct {
	systemPrompt string
	provider     llm.Provider
	clients      map[string]mcpclient.MCPClient
	tools        []llm.Tool
}

type ChatResponse struct {
	Message string   `json:"message"`
	Images  []string `json:"images"`
}

func (h *Host) Close() {
	for name, client := range h.clients {
		if err := client.Close(); err != nil {
			log.Error("Failed to close server", "name", name, "error", err)
		} else {
			log.Info("Server closed", "name", name)
		}
	}
}

type ToolDescription struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *Host) ListTools() []ToolDescription {
	descriptions := []ToolDescription{}

	for _, tool := range h.tools {
		descriptions = append(descriptions, ToolDescription{
			Name:        tool.Name,
			Description: tool.Description,
		})
	}
	return descriptions
}

func (h *Host) RunPrompt(ctx context.Context, prompt string, messages *[]history.HistoryMessage) (*ChatResponse, error) {
	message, err := h.runPromptNonInteractive(ctx, prompt, messages)
	return message, err
}

func (h *Host) runPromptNonInteractive(ctx context.Context, prompt string, messages *[]history.HistoryMessage) (*ChatResponse, error) {
	var message llm.Message
	var err error
	var response ChatResponse

	// This appends the prompt to the history for next time
	if prompt != "" {
		*messages = append(
			*messages,
			history.HistoryMessage{
				Role: "user",
				Content: []history.ContentBlock{{
					Type: "text",
					Text: prompt,
				}},
			},
		)
	}

	// Convert MessageParam to llm.Message for provider
	// Messages already implement llm.Message interface
	llmMessages := make([]llm.Message, len(*messages))
	for i := range *messages {
		llmMessages[i] = &(*messages)[i]
	}

	message, err = h.provider.CreateMessage(
		ctx,
		prompt,
		llmMessages,
		h.tools,
	)

	if err != nil {
		return nil, err
	}

	var messageContent []history.ContentBlock

	log.Infof("runPrompt:%s::%s\n", prompt, message.GetContent())

	if message.GetContent() != "" {
		response.Message = message.GetContent()

		/*
			return &ChatResponse{
				Message: message.GetContent(),
			}, nil*/
	}

	toolResults := []history.ContentBlock{}
	messageContent = []history.ContentBlock{}

	// Add text content
	if message.GetContent() != "" {
		messageContent = append(messageContent, history.ContentBlock{
			Type: "text",
			Text: message.GetContent(),
		})
	}

	// Handle tool calls
	for _, toolCall := range message.GetToolCalls() {
		input, _ := json.Marshal(toolCall.GetArguments())
		messageContent = append(messageContent, history.ContentBlock{
			Type:  "tool_use",
			ID:    toolCall.GetID(),
			Name:  toolCall.GetName(),
			Input: input,
		})

		parts := strings.Split(toolCall.GetName(), "__")
		if len(parts) != 2 {
			log.Warnf("Error: Invalid tool name format: %s\n", toolCall.GetName())
			continue
		}

		serverName, toolName := parts[0], parts[1]
		mcpClient, ok := h.clients[serverName]
		if !ok {
			log.Warnf("Error: Server not found: %s\n", serverName)
			continue
		}

		var toolArgs map[string]interface{}
		if err := json.Unmarshal(input, &toolArgs); err != nil {
			log.Warnf("Error parsing tool arguments: %v\n", err)
			continue
		}

		log.Info("Calling tool", "tool_name", toolName, "tool_args", toolArgs, "server", serverName)

		req := mcp.CallToolRequest{}
		req.Params.Name = toolName
		req.Params.Arguments = toolArgs
		toolResult, err := mcpClient.CallTool(
			context.Background(),
			req,
		)

		if err != nil {
			log.Error("Tool call error", "tool_name", toolName, "tool_args", toolArgs, "server", serverName, "error", err)
			errMsg := fmt.Sprintf(
				"Error calling tool %s: %v",
				toolName,
				err,
			)

			// Add an error message as tool result
			toolResults = append(toolResults, history.ContentBlock{
				Type:      "tool_result",
				ToolUseID: toolCall.GetID(),
				Text:      errMsg,
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: errMsg,
					},
				},
			})
			continue
		}

		log.Info("Tool call success", "tool_name", toolName, "tool_args", toolArgs, "server", serverName, "result", toolResultToString(toolResult))
		if toolResult.Content != nil {
			//log.Debug("raw tool result content", "content", toolResult.Content)

			// Extract text content
			var resultText string
			// Handle array content directly since we know it's []interface{}
			for _, item := range toolResult.Content {
				switch v := item.(type) {
				case mcp.TextContent:
					resultText += fmt.Sprintf("%v ", v.Text)

				case mcp.ImageContent:
					response.Images = append(response.Images, v.Data)

				default:
					panic(fmt.Sprintf("Unknown content type: %T", item))
				}
			}
			toolResults = append(toolResults, history.ContentBlock{
				Type:      "tool_result",
				ToolUseID: toolCall.GetID(),
				Text:      strings.TrimSpace(resultText),
				Content:   toolResult.Content,
			})
		}
	}

	*messages = append(*messages, history.HistoryMessage{
		Role:    message.GetRole(),
		Content: messageContent,
	})

	fmt.Printf("runPrompt:%s::TR::%v\n", prompt, len(toolResults))

	if len(toolResults) > 0 {
		for _, toolResult := range toolResults {
			*messages = append(*messages, history.HistoryMessage{
				Role:    "tool",
				Content: []history.ContentBlock{toolResult},
			})
		}
		// Make another call to get Claude's response to the tool results
		log.Infof("Calling LLM to interpret tool result: m=%v\n", len(*messages))

		pr, err := h.runPromptNonInteractive(ctx, "", messages)
		if err != nil {
			return nil, err
		}

		response.Message = pr.Message
		response.Images = append(response.Images, pr.Images...)
		return &response, nil
	}
	return &response, nil
}

func (h *Host) WithServerConfig(configSrc string) error {
	mcpConfig, err := loadMCPConfig(configSrc)
	if err != nil {
		return err
	}

	h.clients, err = createMCPClients(mcpConfig)
	if err != nil {
		return fmt.Errorf("error creating MCP clients: %v", err)
	}

	for name := range h.clients {
		log.Info("Server connected", "name", name)
	}

	var allTools []llm.Tool
	for serverName, mcpClient := range h.clients {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		toolsResult, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
		cancel()

		if err != nil {
			log.Error(
				"Error fetching tools",
				"server",
				serverName,
				"error",
				err,
			)
			continue
		}

		serverTools := mcpToolsToAnthropicTools(serverName, toolsResult.Tools)
		allTools = append(allTools, serverTools...)
		log.Info(
			"Tools loaded",
			"server",
			serverName,
			"count",
			len(toolsResult.Tools),
		)
	}

	h.tools = allTools

	return nil
}

func NewHost(provider llm.Provider) *Host {
	return &Host{
		provider: provider,
	}
}
