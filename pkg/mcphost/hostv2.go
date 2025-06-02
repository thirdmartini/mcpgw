package mcphost

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/thirdmartini/mcpgw/pkg/history"
	"github.com/thirdmartini/mcpgw/pkg/llm"
)

func (h *Host) runToolUse(ctx context.Context, message llm.Message, messages *[]history.HistoryMessage) (*ChatResponse, error) {
	var messageContent []history.ContentBlock
	var toolResults []history.ContentBlock

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

		log.Info("LLM Requests Tool Call", "tool_name", toolName, "tool_args", toolArgs, "server", serverName)

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

	if len(toolResults) > 0 {
		for _, toolResult := range toolResults {
			*messages = append(*messages, history.HistoryMessage{
				Role:    "tool",
				Content: []history.ContentBlock{toolResult},
			})
		}
		// Make another call to get Claude's response to the tool results
		log.Infof("Calling LLM to interpret tool results")

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

func (h *Host) runPromptNonInteractive2(ctx context.Context, prompt string, messages *[]history.HistoryMessage) (*ChatResponse, error) {
	var message llm.Message
	var err error
	var response ChatResponse

	// This appends the prompt to the history for next time
	if prompt != "" {
		log.Infof("Pompt: %s\n", prompt)

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

	if message.GetContent() != "" {
		response.Message = message.GetContent()
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

	// No tool calls, the LLM is sending us a direct prompt
	if len(message.GetToolCalls()) == 0 {
		response.Message = message.GetContent()
		return &response, nil
	}

	//response, err := h.runToolUse(ctx, message, messages)
}
