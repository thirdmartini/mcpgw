package mcphost

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

// ToolInterface provides a minimal footprint for what our agent needs to interact with extranal tools,
// this allows us to abstract away the MCP complexity away from the agent code and implement other schemes for
// tool request interactions
type ToolInterface interface {
	Initialize(ctx context.Context, request mcp.InitializeRequest) (*mcp.InitializeResult, error)
	Close() error
	ListTools(ctx context.Context, request mcp.ListToolsRequest) (*mcp.ListToolsResult, error)
	CallTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
}
