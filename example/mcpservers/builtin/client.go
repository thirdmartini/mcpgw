package builtin

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

type Builtin struct {
}

// Initialize sends the initial connection request to the server
func (b *Builtin) Initialize(ctx context.Context, request mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	fmt.Printf("Initalize: %v\n")
	result := &mcp.InitializeResult{
		Result: mcp.InitializeResult_OK,
	}
	return result, nil
}

// Ping checks if the server is alive
func (b *Builtin) Ping(ctx context.Context) error {
	fmt.Printf("Ping\n")
	return nil
}

// ListResourcesByPage manually list resources by page.
func (b *Builtin) ListResourcesByPage(
	ctx context.Context,
	request mcp.ListResourcesRequest,
) (*mcp.ListResourcesResult, error) {
	fmt.Printf("ListResourcesByPage: %+v\n", request)
	return nil, nil
}

// ListResources requests a list of available resources from the server
func (b *Builtin) ListResources(
	ctx context.Context,
	request mcp.ListResourcesRequest,
) (*mcp.ListResourcesResult, error) {
	fmt.Printf("ListResources: %+v\n", request)
	return nil, nil
}

// ListResourceTemplatesByPage manually list resource templates by page.
func (b *Builtin) ListResourceTemplatesByPage(
	ctx context.Context,
	request mcp.ListResourceTemplatesRequest,
) (*mcp.ListResourceTemplatesResult,
	error) {
	fmt.Printf("ListResourceTemplatesByPage: %+v\n", request)
	return nil, nil
}

// ListResourceTemplates requests a list of available resource templates from the server
func (b *Builtin) ListResourceTemplates(
	ctx context.Context,
	request mcp.ListResourceTemplatesRequest,
) (*mcp.ListResourceTemplatesResult,
	error) {
	fmt.Printf("ListResourceTemplates: %+v\n", request)
	return nil, nil

}

// ReadResource reads a specific resource from the server
func (b *Builtin) ReadResource(
	ctx context.Context,
	request mcp.ReadResourceRequest,
) (*mcp.ReadResourceResult, error) {
	fmt.Printf("ReadResource: %+v\n", request)
	return nil, nil

}

// Subscribe requests notifications for changes to a specific resource
func (b *Builtin) Subscribe(ctx context.Context, request mcp.SubscribeRequest) error {
	fmt.Printf("Subscribe: %+v\n", request)
	return nil

}

// Unsubscribe cancels notifications for a specific resource
func (b *Builtin) Unsubscribe(ctx context.Context, request mcp.UnsubscribeRequest) error {
	fmt.Printf("Unsubscribe: %+v\n", request)
	return nil

}

// ListPromptsByPage manually list prompts by page.
func (b *Builtin) ListPromptsByPage(
	ctx context.Context,
	request mcp.ListPromptsRequest,
) (*mcp.ListPromptsResult, error) {
	fmt.Printf("ListPromptsByPage: %+v\n", request)
	return nil, nil

}

// ListPrompts requests a list of available prompts from the server
func (b *Builtin) ListPrompts(
	ctx context.Context,
	request mcp.ListPromptsRequest,
) (*mcp.ListPromptsResult, error) {
	fmt.Printf("ListPrompts: %+v\n", request)
	return nil, nil

}

// GetPrompt retrieves a specific prompt from the server
func (b *Builtin) GetPrompt(
	ctx context.Context,
	request mcp.GetPromptRequest,
) (*mcp.GetPromptResult, error) {
	fmt.Printf("GetPrompt: %+v\n", request)
	return nil, nil

}

// ListToolsByPage manually list tools by page.
func (b *Builtin) ListToolsByPage(
	ctx context.Context,
	request mcp.ListToolsRequest,
) (*mcp.ListToolsResult, error) {
	fmt.Printf("ListToolsByPage: %+v\n", request)
	return nil, nil

}

// ListTools requests a list of available tools from the server
func (b *Builtin) ListTools(
	ctx context.Context,
	request mcp.ListToolsRequest,
) (*mcp.ListToolsResult, error) {
	fmt.Printf("ListTools: %+v\n", request)
	return nil, nil

}

// CallTool invokes a specific tool on the server
func (b *Builtin) CallTool(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	fmt.Printf("CallTool: %+v\n", request)
	return nil, nil

}

// SetLevel sets the logging level for the server
func (b *Builtin) SetLevel(ctx context.Context, request mcp.SetLevelRequest) error {
	fmt.Printf("SetLevel: %+v\n", request)
	return nil

}

func (b *Builtin) Complete(ctx context.Context, request mcp.CompleteRequest) (*mcp.CompleteResult, error) {
	fmt.Printf("Complete: %+v\n", request)
}

// Close client connection and cleanup resources
func (b *Builtin) Close() error {
	return nil
}

// OnNotification registers a handler for notifications
func (b *Builtin) OnNotification(handler func(notification mcp.JSONRPCNotification)) {

}
