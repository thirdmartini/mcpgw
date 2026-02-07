package mcphost

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/thirdmartini/mcpgw/pkg/kvlog"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/thirdmartini/mcpgw/pkg/history"
	"github.com/thirdmartini/mcpgw/pkg/llm"
)

var log = kvlog.NewLogger("host")

type Host struct {
	systemPrompt string
	provider     llm.Provider
	clients      map[string]ToolInterface
	tools        []llm.Tool
}

type ChatResponse struct {
	Message string      `json:"message"`
	Images  []string    `json:"images"`
	Metrics llm.Metrics `json:"metrics"`
}

func (h *Host) Close() {
	for name, client := range h.clients {
		if err := client.Close(); err != nil {
			log.WithKVs(kvlog.KVs{
				"name":  name,
				"error": err,
			}).Errorf("Failed to close server")
		} else {
			log.Infof("Server %s closed", name)
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

func (h *Host) RunPrompt(ctx context.Context, prompt string, conversation *Conversation) error {
	return h.runPromptNonInteractive(ctx, prompt, conversation)
}

func (h *Host) runPromptNonInteractive(ctx context.Context, prompt string, conversation *Conversation) error {
	var message llm.Message
	var err error

	// This appends the prompt to the history for next time
	if prompt != "" {
		log.Infof("User prompt").MultiLine(prompt)

		conversation.Append(history.HistoryMessage{
			Role: "user",
			Content: []history.ContentBlock{{
				Type: "text",
				Text: prompt,
			}},
		})
	}

	// FIXME: this is such a hideous hack

	// Convert MessageParam to llm.message for provider
	// Messages already implement llm.message interface
	llmMessages := make([]llm.Message, len(conversation.Messages))
	for i := range conversation.Messages {
		llmMessages[i] = &(conversation.Messages)[i]
	}

	// SEB: notice, prompt is pointless as we are sending the entire conversation down including the prompt as the last llmMessage
	message, err = h.provider.CreateMessage(
		ctx,
		prompt,
		llmMessages,
		h.tools,
	)

	if err != nil {
		log.Errorf("Failed to create a message %v", err)
		return err
	}

	// If we didn't get any tool calls, then we are done, respond to the user
	if !llm.HasToolCalls(message) {
		log.Infof("Agent Response").MultiLine(message.GetContent())

		conversation.Append(history.HistoryMessage{
			Role:    message.GetRole(),
			Metrics: message.GetMetrics(),
			Content: []history.ContentBlock{
				{
					Type: "text",
					Text: message.GetContent(),
				},
			},
		})
		return nil
	}

	//log.Infof("ToolCalls And Message: [%s]", message.GetContent())

	var messageContent []history.ContentBlock
	var toolResults []history.ContentBlock

	// SEB: sometimes we get some commentary from the LLM , in which case it may be worthwhile sending this "midaction" update to the UI
	if message.GetContent() != "" {
		log.Infof("Mid-Action Response").MultiLine(message.GetContent())

		messageContent = append(messageContent, history.ContentBlock{
			Type: "text",
			Text: message.GetContent(),
		})
	}

	for _, toolCall := range message.GetToolCalls() {
		input, _ := json.Marshal(toolCall.GetArguments())
		messageContent = append(messageContent, history.ContentBlock{
			Type:  "tool_use",
			ID:    toolCall.GetID(),
			Name:  toolCall.GetName(),
			Input: input,
		})
	}
	conversation.Append(history.HistoryMessage{
		Role:    message.GetRole(),
		Metrics: message.GetMetrics(),
		Content: messageContent,
	})

	// handle toolcalls requested by llm
	//toolResults := h.runToolCalls(ctx, message)

	for _, toolCall := range message.GetToolCalls() {
		input, _ := json.Marshal(toolCall.GetArguments())

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

		log.WithKVs(kvlog.KVs{"tool_name": toolName, "tool_args": toolArgs, "server": serverName}).Infof("Agent requests tool call")

		req := mcp.CallToolRequest{}
		req.Params.Name = toolName
		req.Params.Arguments = toolArgs
		toolResult, err := mcpClient.CallTool(
			context.Background(),
			req,
		)

		if err != nil {
			log.WithKVs(kvlog.KVs{"tool_name": toolName, "tool_args": toolArgs, "server": serverName, "err": err}).Errorf("Tool call error")
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

		//log.Info("Tool call success", "tool_name", toolName, "tool_args", toolArgs, "server", serverName, "result", toolResultToString(toolResult))
		if toolResult.Content != nil {
			//log.Debug("raw tool result content", "content", toolResult.Content)

			// Extract text content
			var resultText string
			var resultImages []string
			// Handle array content directly since we know it's []interface{}
			for _, item := range toolResult.Content {
				switch v := item.(type) {
				case mcp.TextContent:
					resultText += fmt.Sprintf("%v ", v.Text)

				case mcp.ImageContent:
					resultImages = append(resultImages, v.Data)

				default:
					panic(fmt.Sprintf("Unknown content type: %T", item))
				}
			}
			toolResults = append(toolResults, history.ContentBlock{
				Type:      "tool_result",
				ToolUseID: toolCall.GetID(),
				Text:      strings.TrimSpace(resultText),
				Content:   toolResult.Content,
				Images:    resultImages,
			})
		}
	}

	log.Infof("Tool Results")

	for _, toolResult := range toolResults {
		log.KVs(kvlog.KVs{
			"tool": toolResult.Name,
			"type": toolResult.Type,
		}).Infof("Tool Response").MultiLine(toolResult.Text)

		conversation.Append(history.HistoryMessage{
			Role:    "tool",
			Metrics: message.GetMetrics(),
			Content: []history.ContentBlock{toolResult},
		})
	}

	log.Infof("Calling LLM to interpret tool results")
	return h.runPromptNonInteractive(ctx, "", conversation)
}

// AddBuiltinTool adds a builtin tool to the host (compiled in)
func (h *Host) AddBuiltinTool(name string, client ToolInterface) error {
	log.WithKVs(kvlog.KVs{"name": name}).Infof("Builtin added")
	h.clients[name] = client
	return nil
}

// WithConfig builds a list of mcp clients from a MCPConfig specification
func (h *Host) WithConfig(mcpConfig *MCPConfig) error {
	var err error

	h.clients, err = createMCPClients(mcpConfig)
	if err != nil {
		return fmt.Errorf("error creating MCP clients: %v", err)
	}

	for name := range h.clients {
		log.WithKVs(kvlog.KVs{"name": name}).Infof("Server connected")
	}

	return nil
}

// WithConfigFile loads a MCPConfig from a file and initializes the host
func (h *Host) WithConfigFile(configSrc string) error {
	mcpConfig, err := loadMCPConfig(configSrc)
	if err != nil {
		return err
	}

	return h.WithConfig(mcpConfig)
}

// Start initializes all mcp tools/servers
func (h *Host) Start() error {
	var allTools []llm.Tool
	for serverName, mcpClient := range h.clients {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		toolsResult, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
		cancel()

		if err != nil {
			log.Errorf("%s: Error fetching tools", serverName)
			continue
		}

		serverTools := mcpToolsToAnthropicTools(serverName, toolsResult.Tools)
		allTools = append(allTools, serverTools...)

		log.Eventf("%s: Tools loaded", serverName)
	}
	h.tools = allTools
	return nil
}

// NewHost creates a new Host instance
func NewHost(provider llm.Provider) *Host {
	return &Host{
		provider: provider,
	}
}
