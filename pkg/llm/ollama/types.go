package ollama

import (
	"fmt"
	"strings"
	"time"

	"github.com/thirdmartini/mcpgw/pkg/llm"

	"github.com/ollama/ollama/api"
)

// Message adapts Ollama's message format to our Message interface
type Message struct {
	metrics    api.Metrics
	message    api.Message
	ToolCallID string // Store tool call ID separately since Ollama API doesn't have this field
}

func (m *Message) GetRole() string {
	return m.message.Role
}

func (m *Message) GetContent() string {
	// For tool responses and regular messages, just return the content string
	return strings.TrimSpace(m.message.Content)
}

func (m *Message) GetToolCalls() []llm.ToolCall {
	var calls []llm.ToolCall
	for _, call := range m.message.ToolCalls {
		calls = append(calls, NewOllamaToolCall(call))
	}
	return calls
}

func (m *Message) GetMetrics() llm.Metrics {
	return llm.Metrics{
		InputTokenCount:  m.metrics.PromptEvalCount,
		InputEvalTime:    m.metrics.PromptEvalDuration,
		OutputTokenCount: m.metrics.EvalCount,
		OutputEvalTime:   m.metrics.EvalDuration,
	}
}

func (m *Message) IsToolResponse() bool {
	return m.message.Role == "tool"
}

func (m *Message) GetToolResponseID() string {
	return m.ToolCallID
}

// OllamaToolCall adapts Ollama's tool call format
type OllamaToolCall struct {
	call api.ToolCall
	id   string // Store a unique ID for the tool call
}

func NewOllamaToolCall(call api.ToolCall) *OllamaToolCall {
	return &OllamaToolCall{
		call: call,
		id: fmt.Sprintf(
			"tc_%s_%d",
			call.Function.Name,
			time.Now().UnixNano(),
		),
	}
}

func (t *OllamaToolCall) GetName() string {
	return t.call.Function.Name
}

func (t *OllamaToolCall) GetArguments() map[string]interface{} {
	return t.call.Function.Arguments
}

func (t *OllamaToolCall) GetID() string {
	return t.id
}
