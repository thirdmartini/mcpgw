package mcphost

import (
	"github.com/thirdmartini/mcpgw/pkg/history"
)

type Conversation struct {
	Id       string
	Messages []history.HistoryMessage
	Window   int
}

func (s *Conversation) Prune() {
	s.Messages = s.pruneMessages(s.Messages)
}

func (s *Conversation) pruneMessages(messages []history.HistoryMessage) []history.HistoryMessage {
	if len(messages) <= s.Window {
		return messages
	}

	// Keep only the most recent Messages based on Window size
	messages = messages[len(messages)-s.Window:]

	// Handle Messages
	toolUseIds := make(map[string]bool)
	toolResultIds := make(map[string]bool)

	// First pass: collect all tool use and result IDs
	for _, msg := range messages {
		for _, block := range msg.Content {
			if block.Type == "tool_use" {
				toolUseIds[block.ID] = true
			} else if block.Type == "tool_result" {
				toolResultIds[block.ToolUseID] = true
			}
		}
	}

	// Second pass: filter out orphaned tool calls/results
	var prunedMessages []history.HistoryMessage
	for _, msg := range messages {
		var prunedBlocks []history.ContentBlock
		for _, block := range msg.Content {
			keep := true
			if block.Type == "tool_use" {
				keep = toolResultIds[block.ID]
			} else if block.Type == "tool_result" {
				keep = toolUseIds[block.ToolUseID]
			}
			if keep {
				prunedBlocks = append(prunedBlocks, block)
			}
		}
		// Only include Messages that have content or are not assistant Messages
		if (len(prunedBlocks) > 0 && msg.Role == "assistant") ||
			msg.Role != "assistant" {
			hasTextBlock := false
			for _, block := range msg.Content {
				if block.Type == "text" {
					hasTextBlock = true
					break
				}
			}
			if len(prunedBlocks) > 0 || hasTextBlock {
				msg.Content = prunedBlocks
				prunedMessages = append(prunedMessages, msg)
			}
		}
	}
	return prunedMessages
}

func (s *Conversation) Append(message history.HistoryMessage) {
	s.Messages = append(s.Messages, message)
}

func (s *Conversation) LastReply() *history.HistoryMessage {
	return &s.Messages[len(s.Messages)-1]
}

func (s *Conversation) LastResponse() *ChatResponse {
	history := s.Messages[len(s.Messages)-1]

	response := &ChatResponse{
		Message: history.GetContent(),
		Images:  history.GetImages(),
	}
	// check to see if the previous previous history has an image content block

	if len(response.Images) == 0 {
		// check history 1 back
		if len(s.Messages)-2 >= 0 {
			history = s.Messages[len(s.Messages)-2]
			if history.IsToolResponse() {
				response.Images = history.GetImages()
			}
		}
	}

	return response
}
