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

func (h *Host) Close() {
	for name, client := range h.clients {
		if err := client.Close(); err != nil {
			log.Error("Failed to close server", "name", name, "error", err)
		} else {
			log.Info("Server closed", "name", name)
		}
	}
}

func (h *Host) RunPrompt(ctx context.Context, prompt string, messages *[]history.HistoryMessage) (string, error) {
	message, err := h.runPromptNonInteractive(ctx, prompt, messages)
	return message, err
}

func (h *Host) runPromptNonInteractive(ctx context.Context, prompt string, messages *[]history.HistoryMessage) (string, error) {
	var message llm.Message
	var err error

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
		return "", err
	}

	var messageContent []history.ContentBlock

	if message.GetContent() != "" {
		return message.GetContent(), nil
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

		// Log usage statistics if available
		inputTokens, outputTokens := message.GetUsage()
		if inputTokens > 0 || outputTokens > 0 {
			log.Info("Usage statistics",
				"input_tokens", inputTokens,
				"output_tokens", outputTokens,
				"total_tokens", inputTokens+outputTokens)
		}

		parts := strings.Split(toolCall.GetName(), "__")
		if len(parts) != 2 {
			fmt.Printf(
				"Error: Invalid tool name format: %s\n",
				toolCall.GetName(),
			)
			continue
		}

		serverName, toolName := parts[0], parts[1]
		mcpClient, ok := h.clients[serverName]
		if !ok {
			fmt.Printf("Error: Server not found: %s\n", serverName)
			continue
		}

		var toolArgs map[string]interface{}
		if err := json.Unmarshal(input, &toolArgs); err != nil {
			fmt.Printf("Error parsing tool arguments: %v\n", err)
			continue
		}

		log.Info("Calling tool", "tool_name", toolName, "tool_args", toolArgs, "server", serverName)

		var toolResultPtr *mcp.CallToolResult
		req := mcp.CallToolRequest{}
		req.Params.Name = toolName
		req.Params.Arguments = toolArgs
		toolResultPtr, err = mcpClient.CallTool(
			context.Background(),
			req,
		)

		log.Info("Tool call completed", "tool_name", toolName, "tool_args", toolArgs, "server", serverName)

		if err != nil {
			errMsg := fmt.Sprintf(
				"Error calling tool %s: %v",
				toolName,
				err,
			)
			log.Errorf(errMsg)

			// Add error message as tool result
			toolResults = append(toolResults, history.ContentBlock{
				Type:      "tool_result",
				ToolUseID: toolCall.GetID(),
				Content: []history.ContentBlock{{
					Type: "text",
					Text: errMsg,
				}},
			})
			continue
		}

		toolResult := *toolResultPtr

		if toolResult.Content != nil {
			log.Debug("raw tool result content", "content", toolResult.Content)

			// Create the tool result block
			resultBlock := history.ContentBlock{
				Type:      "tool_result",
				ToolUseID: toolCall.GetID(),
				Content:   toolResult.Content,
			}

			// Extract text content
			var resultText string
			// Handle array content directly since we know it's []interface{}
			for _, item := range toolResult.Content {
				if contentMap, ok := item.(mcp.TextContent); ok {
					resultText += fmt.Sprintf("%v ", contentMap.Text)
				}
			}

			resultBlock.Text = strings.TrimSpace(resultText)
			log.Debug("created tool result block",
				"block", resultBlock,
				"tool_id", toolCall.GetID())

			toolResults = append(toolResults, resultBlock)
		}
	}

	*messages = append(*messages, history.HistoryMessage{
		Role:    message.GetRole(),
		Content: messageContent,
	})

	if len(toolResults) > 0 {
		for _, toolResult := range toolResults {
			*messages = append(*messages, history.HistoryMessage{
				Role:    "tool",
				Content: []history.ContentBlock{toolResult},
			})
		}
		// Make another call to get Claude's response to the tool results
		log.Info("Calling LLM to interpret tool result")
		return h.runPromptNonInteractive(ctx, "", messages)
	}
	return "", nil
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
